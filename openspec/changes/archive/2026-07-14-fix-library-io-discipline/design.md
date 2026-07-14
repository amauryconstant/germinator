## Context

The library package is consumed by 9 `cmd/` commands and orchestrates filesystem I/O for `library init`, `library add`, `library refresh`, `library remove`, and `library validate`. The package's I/O discipline has eroded in 4 ways:

1. **Direct `os.Stdout` writes** at `adder.go:92` and `creator.go:38` for dry-run output. The project's forbidigo pattern at `.golangci.yml:82` only catches `cmd/**` writes; the library package is outside the scope. The result is that piped consumers (e.g., `germinator library add --dry-run | grep foo`) see the dry-run output mixed with their expected stdout stream.

2. **`os.Rename` without `EXDEV` fallback** at `adder.go:333`, `remover.go:193`, and `remover.go:226`. The atomic-write pattern (write-temp-then-rename) fails with `cross-device link` (`syscall.EXDEV`) when the temp file is on a different filesystem from the target. This commonly happens when `TMPDIR=/tmp` and the library is on the user's home filesystem; the `library.yaml.tmp` file lands on `/tmp` and the rename to `library.yaml` (on `/home`) fails.

   **Separately**, `SaveLibrary` (`saver.go:33`) uses direct `os.WriteFile` — not the temp+rename pattern. This is **not** EXDEV-vulnerable (no rename = no EXDEV) but is **non-atomic**: a process kill mid-write leaves a torn `library.yaml`. This change folds `SaveLibrary` into the same helper, but the rationale is atomicity (not EXDEV). Both concerns share one helper because the EXDEV fallback path is a strict superset of the atomic-write path.

3. **Per-directory cancellation missing** in `DiscoverOrphans` (adder.go:785). The function checks `ctx.Err()` only at the top of the directory loop; if the scan encounters a large directory, the cancellation signal is not respected until the current directory finishes. For libraries with thousands of resource files, this can be a multi-second delay.

4. **Unix-only permission bits** at `adder.go:105` (`0o750`, `0o644`). On Windows, these bits are ignored (per `os.Chmod` semantics); the project does not claim Windows support, but the behavior is undocumented and surprising.

The review's Top-10 fixes #7 and #8 call out the first two as priorities. The third and fourth are smaller correctness/observability wins.

### Constraints

1. **No public API break** for external consumers. The library package is `internal/`, so signature changes are acceptable but should be atomic.
2. **The `*Request` dual-form direction is already partially landed** for `Init` (`internal/library/requests.go:19`), `Refresh`, `RemoveResource`, `RemovePreset`, `Validate`, `Fix` — each has a `*Request` struct consumed by a `(*Library) X` method, with the package-level function preserving its existing signature and delegating internally. The writer change in this PR extends `InitRequest`, `AddRequest`, `BatchAddOptions`, and `CreateOptions` with `Stdout io.Writer` so the writer plumbing lives on both legacy structs and the dual-form types. When `extract-io-adapters` Stage 2 introduces `(*Library).Init(ctx, *InitRequest)` etc., the field flows through unchanged. **Note**: `Init` is documented in `internal/library/AGENTS.md:402` as a **package function** (not a `*Library` method) precisely because there is no pre-existing `*Library` to receive a method. When Stage 2 introduces the dual-form methods, `Init` may need a static-instance pattern (`library.Global()` or a singleton accessor) rather than a simple method rename. The `Stdout` field flows through either way.
3. **Single-handling rule** (per `cmd/AGENTS.md`): "errors are either logged OR returned, NEVER both." The library package must not write user-facing output to `os.Stdout` (which is the cmd layer's job). The review's call-out is consistent with this rule.
4. **EXDEV fallback** must remain atomic-or-fail: the cross-filesystem case should produce a state where the new file is fully written before the old one is replaced. A copy+remove approach (rather than a retry-with-different-tmpdir) is simpler and matches the project's atomic-write contract.

## Goals / Non-Goals

**Goals:**

- Remove all `fmt.Fprintln(os.Stdout, ...)` calls from `internal/library/`.
- Add `Stdout io.Writer` field to `AddRequest` / `BatchAddOptions` / `CreateOptions` / `InitRequest` for dry-run output.
- Add `EXDEV` fallback to `os.Rename` at `adder.go:333`, `remover.go:193`, and `remover.go:226` (via the new `atomicWriteFile` helper); convert `SaveLibrary` (`saver.go:33`) from non-atomic direct `os.WriteFile` to the same atomic helper.
- Refactor `scanDirectory` (`adder.go:845`) into an `errgroup.SetLimit(scanConcurrencyLimit)`-bounded parallel walk for sibling-subtree descent; the outer `DiscoverOrphans` 4-directory loop is **not** wrapped.
- Document the Windows permission-bits limitation at `adder.go:105`, `creator.go:57`, `saver.go:21` — with the existing pre-change mode bits (`0o750` dirs, `0o600` for `library.yaml` via `SaveLibrary`, `0o644` for `library.yaml` via `CreateLibrary` and resource files).

**Non-Goals:**

- Adding Windows-specific ACLs (the project does not claim Windows support; documenting the limitation is sufficient).
- Refactoring the dry-run output format (the textual content is preserved; only the destination changes).
- Changing the package-level function signatures (the writer lives on the request struct fields, not as positional parameters — this keeps `extract-io-adapters` Stage 2's mechanical rename low-risk).
- Unifying the `SaveLibrary` (`0o600`) vs `CreateLibrary` (`0o644`) modes for `library.yaml` — out of scope; documented as pre-existing behavior in `library-library-persistence/spec.md`.
- Adding `//go:build !windows` build tags (per the project convention of supporting Unix/macOS as the primary platforms; Windows is documented as out-of-scope).

## Decisions

### 1. Windows permission-bits: document, do not fix

**Choice**: Add a comment at `internal/library/adder.go:105` documenting that `0o750` and `0o644` are ignored on Windows. Do not add `runtime.GOOS` switch or per-platform permission code.

**Rationale**: The project does not claim Windows support. Adding a `runtime.GOOS` switch adds 15-20 LOC of Windows-specific code for a platform the project does not target. The comment makes the limitation explicit so a future contributor doesn't add a Windows fix without considering whether to claim Windows support.

**Alternatives considered**:

- *Active fix via `runtime.GOOS` switch*: rejected; out of project scope.
- *Defer entirely (no comment)*: rejected; the limitation is surprising and the comment is cheap.

### 2. Writer placement: as a `Stdout io.Writer` field on each request struct

**Choice**: Add a `Stdout io.Writer` field to each request/option struct that participates in dry-run output:

```go
// internal/library/adder.go:25
type AddRequest struct {
    // ...existing fields...
    // Stdout receives dry-run output (the cmd layer passes opts.IO.Out).
    // Optional: nil means "no dry-run output" (the default for tests).
    Stdout io.Writer
}

// internal/library/adder.go:540
type BatchAddOptions struct {
    // ...existing fields...
    // Stdout receives future batch dry-run output (cmd layer passes opts.IO.Out).
    Stdout io.Writer
}

// internal/library/creator.go:16
type CreateOptions struct {
    // ...existing fields...
    // Stdout receives dry-run output (typically opts.IO.Out from cmd layer).
    Stdout io.Writer
}

// internal/library/requests.go:19
type InitRequest struct {
    // ...existing fields...
    // Stdout receives dry-run output (typically opts.IO.Out from cmd layer).
    Stdout io.Writer
}
```

The signatures stay positional-clean:

```go
func AddResource(ctx context.Context, opts AddRequest) error
func BatchAddResources(ctx context.Context, opts BatchAddOptions) (*BatchAddResult, error)
func CreateLibrary(opts CreateOptions) error
func Init(ctx context.Context, req *InitRequest) error
```

**Rationale**: The `*Request` dual-form direction is already established in `internal/library/requests.go` (`InitRequest`, `RefreshRequest`, `RemoveResourceRequest`, `RemovePresetRequest`, `ValidateRequest`, `FixRequest`). Adding `Stdout io.Writer` as a struct field follows the existing convention and avoids positional-parameter sprawl on every layer (cmd → Init → CreateLibrary). It also keeps the dual-form direction consistent: when `extract-io-adapters` Stage 2 introduces `(*Library).Init(ctx, *InitRequest)`, the field flows through unchanged.

The cmd layer is the only writer in production. Tests may pass `nil` (dry-run block is a no-op via the `if opts.Stdout != nil` guard) or `bytes.Buffer` (capture dry-run output for assertion). Existing `adder_test.go` keyed struct literals continue to compile with zero-value `Stdout: nil`. All callers (currently `cmd/library_init.go:161-167` and `cmd/library_add.go:352/525/654`) populate `Stdout: opts.IO.Out` on the request struct.

**Alternatives considered**:

- *Positional `io.Writer` parameter on each function*: rejected; disrupts the `*Request` dual-form direction; signature churn on every layer; contradicts the codebase's existing `InitRequest`/`AddRequest` shape.
- *Functional options (`WithStdout(io.Writer)`)*: rejected; the package-level functions are not constructors — they're commands. Functional options fit `NewServer(...)` style constructors, not `(ctx, opts, error)` returns.
- *Global `package var stdout io.Writer = os.Stdout`*: rejected; package-level mutable globals are forbidden by `cmd/AGENTS.md:46`.
- *Render target on a typed `RenderTarget` struct with `Out`/`ErrOut`/`Logger` fields*: rejected; the cmd layer only has one writer (`opts.IO.Out`) for dry-run output. A multi-writer struct adds API surface for one consumer.

### 3. `errgroup` placement: into `scanDirectory`'s recursive walk

**Choice**: Refactor `scanDirectory` (`adder.go:845`) from a sequential `filepath.WalkDir` recursion into an `errgroup.SetLimit(scanConcurrencyLimit)`-bounded parallel walk over sibling subtrees (where `scanConcurrencyLimit` is a file-scope `const = 8` in `internal/library/adder.go`). The outer `DiscoverOrphans` 4-directory loop is **not** wrapped — N=4 fails the errgroup trigger and the outer loop's per-iteration ctx check is sufficient.

**Rationale**: A typical library has 4-8 top-level resource directories (`adder.go:795-800` shows the fixed 4-dir map); an 8-worker cap (`scanConcurrencyLimit`) on the outer loop is a no-op. The actual latency hotspot is the **recursive** `filepath.WalkDir` inside `scanDirectory`: each subtree is walked sequentially, so a deeply nested library (e.g., `skills/sub1/sub2/sub3/...`) holds up cancellation. Parallelizing sibling-subtree descent via `errgroup.SetLimit(scanConcurrencyLimit)` bounds memory while honoring ctx cancellation as soon as the next goroutine observes `ctx.Err()`.

**Why the errgroup lives INSIDE `scanDirectory`, not at the entry of `DiscoverOrphans`**: `scanDirectory` calls itself recursively (via the `DiscoverOrphans` 4-dir loop at `adder.go:802-810`, which calls `scanDirectory`, which recursively walks each subtree). Parallelizing at the entry point would only parallelize the top-level 4-dir map — the skill's `N>10` trigger keys on an **unbounded** number of siblings at any depth, not the fixed 4 top-level dirs. The errgroup with `SetLimit(scanConcurrencyLimit)` (= 8) lives inside `scanDirectory` so each level's sibling-subtree descent races.

**Why `errgroup.SetLimit(scanConcurrencyLimit)` (= 8) and not a hand-rolled semaphore**: Per `golang-cli-architecture`, `SetLimit(n)` is the built-in worker pool — single primitive covers both simple-concurrent and bounded fan-out. The hand-rolled `sem := make(chan struct{}, 8); sem <- {}; defer <-sem` pattern is explicitly noted as a non-idiomatic alternative in the skill's concurrency reference. The cap is hoisted to a file-scope `const scanConcurrencyLimit = 8` so it is grep-able, named in the changelog, and easy to tune in a follow-up change.

**Why not wrap the outer `DiscoverOrphans` 4-directory loop**: The skill's decision triggers are "Sequential I/O >500ms with independent calls → errgroup" and "N>10 items with independent I/O each → `errgroup.SetLimit(n)`". The outer loop has N=4, fails the second trigger, and `dirs/Library.yaml` reads are cheap (typically <50ms per dir on local filesystems). Wrapping it with `SetLimit(scanConcurrencyLimit)` (= 8) would be cargo-cult concurrency.

**Mutex on shared `*DiscoverResult`**: Parallel `scanDirectory` goroutines write to `result.Orphans` / `result.Conflicts` (slice appends) AND increment `result.Summary.TotalScanned` (integer write at `adder.go:868`). All three writes SHALL be covered by a single `sync.Mutex`. The mutex is the cheapest option and matches the codebase's existing `sync.Once` (`internal/cmdutil/factory.go`) precedent; `atomic.AddInt64` would also work but adds a second primitive for one counter. Per `golang-safety`, concurrent writes to shared slice backing arrays and integer fields are unsafe — the mutex enforces serial access for both.

**Alternatives considered**:

- *Wrap outer 4-directory loop in errgroup*: rejected; N=4 fails the skill's N>10 trigger; no measurable benefit on the actual hotspot.
- *Hand-rolled semaphore channel*: rejected; `errgroup.SetLimit` is the idiomatic primitive per the skill.
- *Per-worker `*DiscoverResult` slice merged after `g.Wait()`*: rejected as overly elaborate for the append volume.
- *No concurrency at all (sequential only)*: rejected; the recursive walk's latency on deep trees is exactly the cited C-018 review finding.

### 4. EXDEV fallback: copy + remove

**Choice**: On `syscall.EXDEV`, fall back to `io.Copy(tmpPath, yamlPath)` + `os.Remove(tmpPath)`. The new `yamlPath` is fully written before the old temp is removed.

**Rationale**: This is the standard cross-filesystem rename fallback. It preserves the atomic-write contract at the user-observable level: the new `library.yaml` is either fully present or fully absent (modulo a small window where both exist; see Risks).

**Alternatives considered**:

- *Retry the rename with a tmpdir on the same filesystem as `yamlPath`*: rejected; requires runtime detection of the target filesystem, adds complexity.
- *Document the limitation and require same-filesystem tmpdir*: rejected; the fallback is cheap and improves UX.

### 5. EXDEV fallback factoring: `atomicWriteFile` helper in `internal/library/saver.go`

**Choice**: Extract a single `library.atomicWriteFile(path string, data []byte, perm os.FileMode) error` helper in `internal/library/saver.go` (next to `SaveLibrary` at `saver.go:15`, which already imports `os` and `gerrors`). The helper:

1. Writes `data` to `path+".tmp"` with `perm`
2. `Sync()`s the temp
3. Calls `os.Rename(tmp, path)`
4. On `errors.Is(err, syscall.EXDEV)`: opens `tmp` for read, opens `path` with `O_WRONLY|O_CREATE|O_TRUNC` and the same `perm`, `io.Copy`s, `Sync()`s, then `os.Remove(tmp)`

Four library.yaml write sites delegate to `atomicWriteFile`:

- **Three EXDEV-vulnerable sites** (gain EXDEV fallback via the helper): `adder.go:333`, `remover.go:193`, `remover.go:226`. Each calls `atomicWriteFile(yamlPath, output, 0o600)`.
- **One non-atomic site** (gains atomicity via the helper): `saver.go:33` (`SaveLibrary` currently uses direct `os.WriteFile`, no temp+rename). Calls `atomicWriteFile(yamlPath, data, 0o600)`.

**Rationale**: The helper unifies two related concerns. For the 3 EXDEV-vulnerable sites: it centralizes the `errors.Is(err, syscall.EXDEV)` sentinel-detection pattern (per `golang-error-handling`) and the copy+remove fallback. For `SaveLibrary`: it converts a 1-line non-atomic direct WriteFile into the same atomic temp+rename pattern, eliminating the torn-write failure mode. The single helper is the canonical "how to write library.yaml" implementation. Per the Functional Core / Imperative Shell model (`golang-cli-architecture`), the helper lives in `internal/library/` (Imperative Shell) — **not** in `internal/core/`, which forbids I/O via `depguard`. Per `golang-error-handling`, `errors.Is(err, syscall.EXDEV)` is the canonical sentinel-matching pattern; `syscall.EXDEV` is the standard Go cross-platform constant on Unix/macOS (the package's supported platforms).

**Alternatives considered**:

- *Inline repetition at each site*: rejected; DRY violation across 4 sites.
- *Retry with same-filesystem tmpdir (`os.CreateTemp(filepath.Dir(yamlPath), ".tmp-*")`)*: rejected for this change; deferred (requires runtime detection of target filesystem, more complex than copy+remove; the fallback is cheaper and matches the project's atomic-write contract).
- *Helper in `internal/core/`*: rejected; `core/` policy forbids I/O via `depguard`.

### 6. `fmt.Fprintln(os.Stdout, ...)` removal: per-call-site, gated on `opts.Stdout != nil`

**Choice**: Each of the 2 sites (`adder.go:92-96` and `creator.go:38-43`) is updated in its own commit. Each `fmt.Fprintln(os.Stdout, ...)` is replaced with `if opts.Stdout != nil { fmt.Fprintln(opts.Stdout, ...) }`. The `os` import in `adder.go` and `creator.go` is removed if no other uses remain (per `goimports`).

**Rationale**: Atomic per-site commits avoid the intermediate state where the request struct's `Stdout` field has been added but the body still references `os.Stdout` directly (which would compile but contradict the migration contract). The `if opts.Stdout != nil` guard makes test injection trivial: tests pass `nil` and the dry-run block is a no-op.

**Alternatives considered**:

- *Single mega-commit*: rejected; harder to review and bisect; harder to bisect EXDEV vs dry-run vs errgroup regressions.
- *Always write to a non-nil writer (no nil-guard)*: rejected; the codebase's test convention is to construct real disk state and exercise the happy path; injecting `&bytes.Buffer{}` everywhere would be repetitive, and a required non-nil writer would mean mandatory test setup, which is hostile to the existing test inventory.

## Risks / Trade-offs

- **EXDEV fallback has a small window** where `yamlPath` exists with new content but `tmpPath` is not yet removed. A crash in this window leaves the library readable (next `LoadLibrary` sees the new file) but with a stale `tmp` file (which the next `Save` overwrites). **Mitigation**: document the window in the spec; the next `LoadLibrary` call does not see the tmp file (it looks for `library.yaml`, not `library.yaml.tmp`).
- **`errgroup.SetLimit(scanConcurrencyLimit)` (= 8) adds up to 8 sibling-subtree goroutines.** For deeply nested libraries, the cap bounds memory; for shallow libraries the cap has no effect. **Mitigation**: design Decision 3 evaluates the placement; a follow-up change can introduce a configurable cap if real workloads require it (e.g., a Factory field, deferred to keep API surface small).
- **`scanDirectory` refactor changes the recursive traversal shape.** Pre-change: `filepath.WalkDir` was sequential within each subtree. Post-change: parallel sibling-subtree descent via errgroup. The order of `result.Orphans` is **not** preserved across parallel goroutines; downstream consumers that sorted/asserted on order may need adjustment. **Mitigation**: `DiscoverOrphans`'s contract returns an unordered list of orphans (`*DiscoverResult.Orphans`); no consumer orders them by directory or path. The summary statistics (`TotalScanned`, `TotalOrphans`, etc.) are unaffected.
- **Adding `Stdout io.Writer` to a struct is a public surface change.** The structs (`AddRequest`, `BatchAddOptions`, `CreateOptions`, `InitRequest`) are exported, so adding a field is source-compatible but technically a struct-shape change. **Mitigation**: the package is `internal/`, so external consumers see no API break; tests that pass the struct by value are unaffected (a new field is fine); tests that use positional struct literals will see a compile error and need to switch to keyed literals (already the codebase's idiomatic style).
- **`Stdout` is `nil` for non-cmd callers (e.g., tests that don't care about output).** **Mitigation**: every dry-run block is gated on `if opts.Stdout != nil`; passing `nil` is a no-op.
- **`errors` package import** — `adder.go` already imports `errors` (used at line 22 for `ErrNameConflict`); `saver.go` requires new `errors`, `syscall`, and `io` imports for the helper. **Mitigation**: task 3.1 adds all three to `saver.go` in one diff hunk; `adder.go` and `remover.go` need no new imports because they call `atomicWriteFile` (a package-internal helper that encapsulates `syscall`/`errors`).

## Migration Plan

The change ships in **one PR with 4 atomic phases** (each commit is independently testable):

1. **Phase 1 — Add `Stdout io.Writer` field to `CreateOptions` + `InitRequest`, forward through `Init` → `CreateLibrary`, replace `os.Stdout` writes** (tasks 1.1-1.9). Update `cmd/library_init.go:161-167` to include `Stdout: opts.IO.Out` in the `InitRequest` literal. Update `cmd/library_init_test.go` (existing `runF` injection). Verify `mise run test` and `mise run lint`.
2. **Phase 2 — Add `Stdout io.Writer` field to `AddRequest` + `BatchAddOptions`, replace `os.Stdout` writes** (tasks 2.1-2.10). Update `cmd/library_add.go:352/525/654` to populate `Stdout: opts.IO.Out` in each request struct literal. Update `libraryAdapter` pass-throughs. Verify `mise run test` and `mise run test:e2e`.
3. **Phase 3 — Add `atomicWriteFile` helper + delegate 4 call sites** (tasks 3.1-3.7). Add `errors`, `syscall`, `io` imports to `saver.go`. Add a test in `saver_test.go` that uses a `TMPDIR` on a different filesystem from the library path (or stub the rename via test-only seam if cross-filesystem setup is unreliable in CI). Verify `mise run test`.
4. **Phase 4 — `scanDirectory` refactor with `errgroup.SetLimit(scanConcurrencyLimit)` + Windows doc comments** (tasks 4.1-4.8). Add `golang.org/x/sync/errgroup` to `go.mod`. Add `TestDiscoverOrphans_CtxCancelled` to `refresher_test.go` (where existing `TestDiscoverOrphans` lives). Verify `mise run test`, `mise run test:race`, and the `rg` guardrails.

**Rollback strategy**: revert each phase commit independently. Phases 1-2 are struct-field additions (revert restores the prior struct shapes and re-adds the `os.Stdout` writes; no signature change, so the rollback is clean). Phase 3 is a small correctness fix (revert restores the failing-rename behavior). Phase 4 is a performance/correctness change (revert restores the sequential `filepath.WalkDir`).

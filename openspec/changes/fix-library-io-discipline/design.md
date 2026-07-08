## Context

The library package is consumed by 9 `cmd/` commands and orchestrates filesystem I/O for `library init`, `library add`, `library refresh`, `library remove`, and `library validate`. The package's I/O discipline has eroded in 4 ways:

1. **Direct `os.Stdout` writes** at `adder.go:92` and `creator.go:38` for dry-run output. The project's forbidigo pattern at `.golangci.yml:82` only catches `cmd/**` writes; the library package is outside the scope. The result is that piped consumers (e.g., `germinator library add --dry-run | grep foo`) see the dry-run output mixed with their expected stdout stream.

2. **`os.Rename` without `EXDEV` fallback** at `adder.go:333`. The atomic-write pattern (write-temp-then-rename) fails with `cross-device link` (`syscall.EXDEV`) when the temp file is on a different filesystem from the target. This commonly happens when `TMPDIR=/tmp` and the library is on the user's home filesystem; the `library.yaml.tmp` file lands on `/tmp` and the rename to `library.yaml` (on `/home`) fails.

3. **Per-directory cancellation missing** in `DiscoverOrphans` (adder.go:785). The function checks `ctx.Err()` only at the top of the directory loop; if the scan encounters a large directory, the cancellation signal is not respected until the current directory finishes. For libraries with thousands of resource files, this can be a multi-second delay.

4. **Unix-only permission bits** at `adder.go:105` (`0o750`, `0o644`). On Windows, these bits are ignored (per `os.Chmod` semantics); the project does not claim Windows support, but the behavior is undocumented and surprising.

The review's Top-10 fixes #7 and #8 call out the first two as priorities. The third and fourth are smaller correctness/observability wins.

### Constraints

1. **No public API break** for external consumers. The library package is `internal/`, so signature changes are acceptable but should be atomic.
2. **`extract-io-adapters` Stage 2** is in flight: it converts `AddResource` / `BatchAddResources` / `DiscoverOrphans` from package-level functions to methods on `*library.Library`. The `io.Writer` parameter added in this change must be present in both the package-level function and the `*Library` method, to keep the migration atomic.
3. **Single-handling rule** (per `cmd/AGENTS.md`): "errors are either logged OR returned, NEVER both." The library package must not write user-facing output to `os.Stdout` (which is the cmd layer's job). The review's call-out is consistent with this rule.
4. **EXDEV fallback** must remain atomic-or-fail: the cross-filesystem case should produce a state where the new file is fully written before the old one is replaced. A copy+remove approach (rather than a retry-with-different-tmpdir) is simpler and matches the project's atomic-write contract.

## Goals / Non-Goals

**Goals:**

- Remove all `fmt.Fprintln(os.Stdout, ...)` calls from `internal/library/`.
- Add `io.Writer` parameter to `AddResource` / `BatchAddResources` / `CreateLibrary` for dry-run output.
- Add `EXDEV` fallback to `os.Rename` in `adder.go:333`.
- Wrap `DiscoverOrphans` directory scan in `errgroup` for per-directory cancellation.
- Document the Windows permission-bits limitation at `adder.go:105`.

**Non-Goals:**

- Adding Windows-specific ACLs (the project does not claim Windows support; documenting the limitation is sufficient).
- Refactoring the dry-run output format (the textual content is preserved; only the destination changes).
- Changing the package-level function signatures *other than* the `io.Writer` parameter (the `extract-io-adapters` Stage 2 migration is the right place for broader API changes).
- Adding `//go:build !windows` build tags (per the project convention of supporting Unix/macOS as the primary platforms; Windows is documented as out-of-scope).

## Decisions

### 1. Windows permission-bits: document, do not fix

**Choice**: Add a comment at `internal/library/adder.go:105` documenting that `0o750` and `0o644` are ignored on Windows. Do not add `runtime.GOOS` switch or per-platform permission code.

**Rationale**: The project does not claim Windows support. Adding a `runtime.GOOS` switch adds 15-20 LOC of Windows-specific code for a platform the project does not target. The comment makes the limitation explicit so a future contributor doesn't add a Windows fix without considering whether to claim Windows support.

**Alternatives considered**:

- *Active fix via `runtime.GOOS` switch*: rejected; out of project scope.
- *Defer entirely (no comment)*: rejected; the limitation is surprising and the comment is cheap.

### 2. `io.Writer` parameter placement: as last parameter, before `error`

**Choice**: The `io.Writer` is added as a new parameter to `AddResource`, `BatchAddResources`, and `CreateLibrary`. Specifically:

```go
func AddResource(ctx context.Context, opts AddOptions, stdout io.Writer) error
func BatchAddResources(ctx context.Context, opts *BatchAddOptions, stdout io.Writer) (*BatchAddResult, error)
func CreateLibrary(opts CreateOptions, stdout io.Writer) error
```

**Rationale**: The `stdout` parameter is a render target, not a primary input. Placing it as the last parameter (before `error` in the return signature) is consistent with Go conventions for "context, primary input, render target." The cmd layer is the only caller and will pass `opts.IO.Out` at the call site.

**Alternatives considered**:

- *Add `stdout` to a request struct (`AddRequest`, `BatchAddRequest`, `CreateRequest`)*: rejected; the structs are already large; adding a writer to them blurs the read/write boundary.
- *Use a global `package var stdout io.Writer = os.Stdout` for tests*: rejected; package-level mutable globals are forbidden by `cmd/AGENTS.md:46`.

### 3. `errgroup` concurrency cap: 8

**Choice**: Wrap the `DiscoverOrphans` directory scan in `errgroup.WithContext` with a semaphore channel that caps concurrency at 8.

**Rationale**: A typical library has 4-8 resource subdirectories; an 8-worker cap handles the common case without goroutine sprawl. For libraries with thousands of nested directories, the cap bounds memory usage while still parallelizing.

**Why errgroup over sequential-first (skill threshold justification)**: The `golang-cli-architecture` skill prescribes sequential-first for CLIs (`SKILL.md:892-893`) with two errgroup decision triggers: "Sequential I/O >500ms with independent calls → errgroup" and "N>10 items processed with independent I/O each → errgroup with SetLimit(n)". For the typical 4-8 subdirectory case, sequential would satisfy both thresholds (N<10, I/O latency <500ms). The choice to use errgroup uniformly is justified by the **outlier case**: a 10k+ subdirectory library's sequential scan exceeds the 500ms threshold (per-directory I/O + YAML parse × 10k = measurable latency). Branching on dir count would add complexity without real benefit. Cap=8 bounds memory on the outlier case while staying negligible for the typical case.

**Why cap=8 and not `SetLimit(N)`**: The skill's recommendation is `SetLimit(n)` where n is the workload size. Cap=8 is a fixed upper bound because:
- 8 workers process 10k items at ~1250 items/worker — still I/O-bound (each worker is mostly waiting on `os.ReadDir` + `yaml.Unmarshal`).
- 8 is the documented maximum for outliers; for the typical case (4-8 dirs), the cap has no effect.
- Making it configurable adds API surface (a Factory field or function param) for one internal function — rejected.

**Alternatives considered**:

- *Sequential for typical, errgroup for outliers (branching on dir count)*: rejected; conditional logic adds complexity; the errgroup overhead is negligible at N<10.
- *No cap (unlimited goroutines)*: rejected; risk of OOM on libraries with 10,000+ directories.
- *Configurable cap via Factory*: rejected; over-engineered for a single internal function.

### 4. EXDEV fallback: copy + remove

**Choice**: On `syscall.EXDEV`, fall back to `io.Copy(tmpPath, yamlPath)` + `os.Remove(tmpPath)`. The new `yamlPath` is fully written before the old temp is removed.

**Rationale**: This is the standard cross-filesystem rename fallback. It preserves the atomic-write contract at the user-observable level: the new `library.yaml` is either fully present or fully absent (modulo a small window where both exist; see Risks).

**Alternatives considered**:

- *Retry the rename with a tmpdir on the same filesystem as `yamlPath`*: rejected; requires runtime detection of the target filesystem, adds complexity.
- *Document the limitation and require same-filesystem tmpdir*: rejected; the fallback is cheap and improves UX.

### 5. `fmt.Fprintln(os.Stdout, ...)` removal: per-call-site

**Choice**: Each of the 2 sites (`adder.go:92`, `creator.go:38`) is updated in its own commit. The `os` import in `adder.go` and `creator.go` is removed if no other uses remain (per `goimports`).

**Rationale**: Atomic per-site commits avoid the intermediate state where the function signature is updated but the body still references `os.Stdout` (which would fail to compile).

**Alternatives considered**:

- *Single mega-commit*: rejected; harder to review and bisect.

## Risks / Trade-offs

- **EXDEV fallback has a small window** where `yamlPath` exists with new content but `tmpPath` is not yet removed. A crash in this window leaves the library readable (next `LoadLibrary` sees the new file) but with a stale `tmp` file (which the next `Save` overwrites). **Mitigation**: document the window in the spec; the next `LoadLibrary` call does not see the tmp file (it looks for `library.yaml`, not `library.yaml.tmp`).
- **`errgroup` adds a goroutine per directory.** For libraries with thousands of directories, an 8-worker cap bounds memory but the cap is not configurable. **Mitigation**: design Decision 3 evaluates the cap; a follow-up change can make it configurable if real workloads require it.
- **Signature break for `AddResource` / `BatchAddResources` / `CreateLibrary`.** The package is `internal/`, but the migration is mechanical. **Mitigation**: the cmd layer (sole caller) is updated in the same commit; test files are updated; the build fails immediately if a caller is missed.
- **The `io.Writer` parameter is `nil` for non-cmd callers (e.g., tests that don't care about output).** **Mitigation**: the dry-run block is gated on `if stdout != nil`; passing `nil` is a no-op.

## Migration Plan

The change ships in **one PR with 4 atomic phases** (each commit is independently testable):

1. **Phase 1 — Add `io.Writer` parameter to `CreateLibrary` + remove `os.Stdout` write** (tasks 3.1, 3.4). Update `cmd/library_init.go:144` to pass `opts.IO.Out`. Update `creator_test.go`. Verify `mise run test`.
2. **Phase 2 — Add `io.Writer` parameter to `AddResource` / `BatchAddResources` + remove `os.Stdout` write** (tasks 3.2, 3.5, 3.6). Update `cmd/library_add.go:347` to pass `opts.IO.Out`. Update `adder_test.go`. Verify `mise run test` and `mise run test:e2e`.
3. **Phase 3 — Add `EXDEV` fallback** (task 3.3). Add a test in `adder_test.go` that uses a `TMPDIR` on a different filesystem from the library path. Verify `mise run test`.
4. **Phase 4 — `errgroup` + Windows doc** (tasks 3.7, 3.8). Update `adder_test.go` with a `ctx cancellation` test. Verify `mise run test` and `mise run test:race`.

**Rollback strategy**: revert each phase commit independently. Phases 1-2 are signature changes (revert restores the prior signatures and re-adds the `os.Stdout` writes). Phase 3 is a small correctness fix (revert restores the failing behavior). Phase 4 is a performance/correctness change (revert restores the sequential scan).

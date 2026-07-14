## Why

The 2026-07-08 code review identified 4 I/O discipline findings in `internal/library/` that violate the Unix contract and the cross-filesystem safety contract:

1. **B-003** — `internal/library/adder.go:92` writes dry-run output directly to `os.Stdout`, polluting piped consumers' stdout. The project's `forbidigo` pattern at `.golangci.yml:82` enforces `fmt.Fprintf(os.Stdout|Stderr)` against `cmd/**` only; the library package is outside that scope, allowing the leak.
2. **BD-005 / B-004** — `internal/library/creator.go:38` writes dry-run output directly to `os.Stdout`, same issue.
3. **C-009** — `internal/library/adder.go:333`, `internal/library/remover.go:193`, and `internal/library/remover.go:226` use `os.Rename` without an `EXDEV` fallback; cross-filesystem renames (e.g., `/tmp` → `/home`) return a wrapped `*core.FileError` and fail the mutation. **`SaveLibrary` (`internal/library/saver.go:33`) is a separate, related concern**: it uses direct `os.WriteFile` (no temp+rename), which is not EXDEV-vulnerable but is non-atomic — a process kill mid-write leaves a torn `library.yaml`. This change folds `SaveLibrary` into the same helper to gain atomicity, but the rationale is distinct (atomicity, not EXDEV).
4. **C-018** — `internal/library/adder.go:785` `DiscoverOrphans` uses sequential `filepath.WalkDir` inside `scanDirectory`; cancellation is checked per-file but sibling-subtree descent is serial, so `ctx` cancellation propagation is bounded by the slowest subtree's wall time on deep libraries.

The first two are Top-10 fixes (#7 and #8 in the review). The library package must not write to `os.Stdout` — the cmd layer is responsible for rendering user-facing output. Cross-filesystem rename safety and parallel cancellation propagation are correctness fixes for the persistence contract documented in `library-library-persistence/spec.md`.

This change enforces the I/O discipline contract for the library package. It is a **production-code refactor** with spec deltas.

## What Changes

- **MODIFY** `internal/library/adder.go:25` — add `Stdout io.Writer` field to `AddRequest`. Remove the `fmt.Fprintln(os.Stdout, ...)` dry-run block at `adder.go:91-98`; render via the writer, gated on `if opts.Stdout != nil`.
- **MODIFY** `internal/library/adder.go:540` — add `Stdout io.Writer` field to `BatchAddOptions` (forward groundwork for future batch dry-run output).
- **MODIFY** `internal/library/creator.go:16` — add `Stdout io.Writer` field to `CreateOptions`. Remove the `fmt.Fprintln(os.Stdout, ...)` dry-run block at `creator.go:37-45`; render via the writer.
- **MODIFY** `internal/library/requests.go:19` — add `Stdout io.Writer` field to `InitRequest`. Forward `req.Stdout` from `Init` (`creator.go:93-109`) into `CreateLibrary` at `creator.go:101`.
- **MODIFY** `cmd/library_init.go:161-167` — populate `Stdout: opts.IO.Out` on the `library.Init(opts.Ctx, &library.InitRequest{...})` literal.
- **MODIFY** `cmd/library_add.go:352,525,654` — populate `Stdout: opts.IO.Out` on the three request struct literals (`AddRequest` at 352, `BatchAddOptions` at 525 and 654). The `resourceAdder` interface at `cmd/library_add.go:64-68` does not need a signature change since `Stdout` is a struct field.
- **MODIFY** `cmd/library_add.go:82-107` — `libraryAdapter.AddResource` and `libraryAdapter.BatchAddResources` pass `req.Stdout` / `opts.Stdout` through to the underlying package-level functions.
- **MODIFY** `internal/library/saver.go` (new helper next to `SaveLibrary`) — extract `os.WriteFile`-then-`os.Rename` plus `EXDEV` fallback into a single `library.atomicWriteFile(path, data, mode)` helper. Replace the inline patterns at `adder.go:333`, `remover.go:193`, and `remover.go:226` (EXDEV fallback) and at `saver.go:33` (atomicity improvement — `SaveLibrary` switches from non-atomic direct `os.WriteFile` to atomic temp+rename). The helper uses `errors.Is(err, syscall.EXDEV)` and falls back to `io.Copy` + `os.Remove`. Rationale: three EXDEV-vulnerable patterns (~20-line each) and one non-atomic pattern (1-line direct WriteFile) all become one source of truth; per `golang-error-handling`, `errors.Is` on a sentinel is the canonical cross-filesystem detection pattern.
- **MODIFY** `internal/library/adder.go:845` (`scanDirectory`) — refactor from `filepath.WalkDir`-sequential recursion to `errgroup.SetLimit(scanConcurrencyLimit)` parallel sibling-subtree descent (where `scanConcurrencyLimit` is a file-scope `const = 8` declared in `internal/library/adder.go`). Outer `DiscoverOrphans` 4-directory loop (`adder.go:802-810`) is **not** wrapped. Each subtree goroutine checks `ctx.Err()` before yielding. Shared `*DiscoverResult` appends guarded by `sync.Mutex`.
- **DOCUMENT** Windows permission-bit limitation (`0o750` for dirs, `0o600` for `library.yaml` via `SaveLibrary`, `0o644` for resource files and the `CreateLibrary` initial write — all ignored on Windows) at `internal/library/adder.go:104-105`, `internal/library/creator.go:57`, `internal/library/saver.go:21`. Pre-change behavior preserved; this is documentation only (no behavior change on Unix/macOS). Permission-bit unification between `SaveLibrary` and `CreateLibrary` is out of scope.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- **`library-library-persistence`** — require `atomicWriteFile` (temp+rename with `EXDEV` fallback) for all `library.yaml` mutations. Three pre-existing temp+rename sites (`adder.go:333`, `remover.go:193`, `remover.go:226`) gain EXDEV fallback; one pre-existing direct-WriteFile site (`saver.go:33`, `SaveLibrary`) gains atomicity. Document cross-filesystem safety and pre-existing perm-bit values (`0o750` dirs, `0o600` `library.yaml` via `atomicWriteFile` callers, `0o644` `library.yaml` via `CreateLibrary` and resource files; Windows ignores all).
- **`library-library-scaffolding`** — `CreateOptions` SHALL expose a `Stdout io.Writer` field for dry-run output; `internal/library` SHALL NOT write to `os.Stdout` directly.
- **`library-library-resource-import`** — `AddRequest` SHALL expose a `Stdout io.Writer` field for dry-run output.
- **`library-library-batch-add`** — `BatchAddOptions` SHALL expose a `Stdout io.Writer` field (forward groundwork for future batch dry-run output).
- **`library-library-orphan-discovery`** — `scanDirectory` SHALL use `errgroup.SetLimit(scanConcurrencyLimit)` (= 8 concurrent workers) to parallelize sibling-subtree descent; cancellation MUST be observed by each subtree goroutine before processing. The outer `DiscoverOrphans` 4-directory loop is **not** wrapped.

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `internal/library/creator.go:16` (`CreateOptions`) | Add `Stdout io.Writer` field | +3 |
| `internal/library/creator.go:37-45` (dry-run block) | `os.Stdout` → `opts.Stdout` (gated) | -6 / +6 |
| `internal/library/requests.go:19` (`InitRequest`) | Add `Stdout io.Writer` field | +3 |
| `internal/library/creator.go:93-109` (`Init`) | Forward `req.Stdout` into `CreateLibrary` | +1 |
| `internal/library/adder.go:25` (`AddRequest`) | Add `Stdout io.Writer` field | +3 |
| `internal/library/adder.go:540` (`BatchAddOptions`) | Add `Stdout io.Writer` field | +3 |
| `internal/library/adder.go:91-98` (dry-run block) | `os.Stdout` → `opts.Stdout` (gated) | -5 / +5 |
| `internal/library/saver.go` (new `atomicWriteFile` helper) | Extract temp-write + Rename + EXDEV fallback | +40 |
| `internal/library/adder.go:333` (`os.Rename`) | Replace inline pattern with `atomicWriteFile(yamlPath, output, 0o600)` | -20 / +1 |
| `internal/library/remover.go:193` (`os.Rename`) | Replace inline pattern with `atomicWriteFile(yamlPath, output, 0o600)` | -5 / +1 |
| `internal/library/remover.go:226` (`os.Rename`) | Replace inline pattern with `atomicWriteFile(yamlPath, output, 0o600)` | -5 / +1 |
| `internal/library/saver.go:33` (`os.WriteFile`, non-atomic) | Replace direct WriteFile with `atomicWriteFile(yamlPath, data, 0o600)` (atomicity improvement, not EXDEV) | -1 / +1 |
| `internal/library/adder.go:845` (`scanDirectory`) | `errgroup.SetLimit(scanConcurrencyLimit)` parallel walk | +25 / -10 |
| `internal/library/adder.go:104-105` (MkdirAll) | Windows doc comment | +2 |
| `internal/library/creator.go:57` (MkdirAll) | Windows doc comment | +2 |
| `internal/library/saver.go:21` (`MkdirAll`) | Windows doc comment | +2 |
| `internal/library/adder_test.go` | Test updates: pass struct fields; add EXDEV + ctx-cancel tests | +30 / -10 |
| `cmd/library_add.go:352` | Populate `Stdout: opts.IO.Out` on `AddRequest` | +1 |
| `cmd/library_add.go:525` | Populate `Stdout: opts.IO.Out` on `BatchAddOptions` | +1 |
| `cmd/library_add.go:654` | Populate `Stdout: opts.IO.Out` on `BatchAddOptions` | +1 |
| `cmd/library_init.go:161-167` | Populate `Stdout: opts.IO.Out` on `InitRequest` | +1 |
| `go.mod` | Add `golang.org/x/sync` (errgroup dep) | +1 line |

### Affected systems

- **Public API:** struct fields added (source-compatible add; positional struct literals become compile errors and need to switch to keyed). The package is `internal/`, so external consumers see no API break; internal tests using keyed literals are unaffected. Verified via `rg "library\.(AddRequest|BatchAddOptions|CreateOptions|InitRequest)\{" internal/ cmd/ test/` — zero positional literals exist in the codebase.
- **Atomic-write coverage:** all four `library.yaml` write sites (`adder.go:333`, `remover.go:193`, `remover.go:226`, `saver.go:33`) now delegate to `atomicWriteFile`. Pre-change: 3 inline `os.Rename` call sites (EXDEV-vulnerable) + 1 direct `os.WriteFile` site (`saver.go:33`, non-atomic). Post-change: 1 helper-defined `os.Rename` (inside `atomicWriteFile`) plus 4 helper-call sites. Verified via `rg "os\.Rename" internal/library/` returning exactly 1 match post-change and `rg "atomicWriteFile" internal/library/` returning 5 matches (1 def + 4 callers).
- **Cross-filesystem safety:** atomic writes now succeed across filesystems (e.g., when `tmpPath` is on `/tmp` and `yamlPath` is on `/home`). Pre-change behavior was a loud error (not silent) — the `os.Rename` returned `cross-device link` wrapped in `*core.FileError`; the temp file was left behind. Post-change: cross-filesystem `os.Rename` triggers a copy+remove fallback (atomic-or-fail at user-observable level).
- **Recursive cancellation:** `scanDirectory` now processes sibling subtrees concurrently; ctx cancellation is observed at the next goroutine yield (`ctx.Err()` check at goroutine entry) and the errgroup's `WithContext` propagation cancels all in-flight goroutines on first error.
- **Windows perm bits:** documented as pre-existing limitation at the three call sites. No behavior change on Unix/macOS. Windows support remains out of scope (consistent with the existing `isTerminalFile` caveat in `internal/iostreams`).

## Risks

- **EXDEV fallback introduces a small window where the new file exists at `yamlPath` but the temp file is not yet removed.** A crash in that window leaves both files; the next read sees the new file. **Mitigation**: document the window in the `library-library-persistence/spec.md` delta; the next `LoadLibrary` call will read the new file (correct semantics) and ignore the temp file.
- **`errgroup.SetLimit(scanConcurrencyLimit)` (= 8) adds up to 8 sibling-subtree goroutines.** Deeply nested libraries benefit (capped memory); shallow libraries see no effect. **Mitigation**: design Decision 3 evaluates the placement and cap; a follow-up change can make the cap configurable if real workloads require it.
- **Adding `Stdout io.Writer` as a struct field changes the struct shape.** Pre-change struct literals (`&library.AddRequest{...}`) keep working unchanged; tests using positional literals would break. **Mitigation**: the codebase's idiomatic style is keyed struct literals (e.g., `cmd/library_add.go:343-351`); an audit confirms no positional `AddRequest` / `BatchAddOptions` / `CreateOptions` / `InitRequest` literals exist. The build fails immediately if any positional literal was missed.
- **Order of `*DiscoverResult.Orphans` is non-deterministic across parallel goroutines.** Pre-change `filepath.WalkDir` produced directory-then-file order; post-change the order is set by goroutine scheduling. **Mitigation**: `DiscoverResult.Orphans` is an unordered list per the existing contract; tests asserting on order need to switch to set-membership or `len()` assertions. No production consumer depends on order.
- **Permission-bit documentation is a non-fix.** The C-016 finding is documented but not fixed (no `runtime.GOOS` switch); the existing `SaveLibrary` (`0o600`) vs `CreateLibrary` (`0o644`) modes are documented as pre-existing behavior without unifying them. **Mitigation**: design Decision 1 evaluates the `runtime.GOOS` switch as an alternative; if the user prefers the active fix, the scope expands by ~15 LOC and is deferred to a follow-up change.

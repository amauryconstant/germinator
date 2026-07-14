# Tasks — Fix library I/O discipline

Each task ends with `mise run check` passing. Tasks are grouped by phase and ordered so each commit is independently testable.

## 1. Phase 1 — Thread `Stdout io.Writer` through `Init` → `CreateLibrary` and remove `os.Stdout` write

- [x] 1.1 In `internal/library/creator.go`, add a `Stdout io.Writer` field to `CreateOptions` (with doc comment matching the `AddRequest` style).
- [x] 1.2 In `internal/library/requests.go:19`, add a `Stdout io.Writer` field to `InitRequest` (with doc comment).
- [x] 1.3 In `internal/library/creator.go:93-109` (`Init`), forward `req.Stdout` into the `CreateLibrary(CreateOptions{...})` call (the literal at `creator.go:101`).
- [x] 1.4 In `internal/library/creator.go:37-45`, replace the 6 `fmt.Fprintln(os.Stdout, …)` calls with `if opts.Stdout != nil { fmt.Fprintln(opts.Stdout, …) }`. Dry-run block at `creator.go:37`.
- [x] 1.5 In `cmd/library_init.go:161-167`, update the `library.Init(opts.Ctx, &library.InitRequest{...})` literal to include `Stdout: opts.IO.Out`.
- [x] 1.6 There is no `internal/library/creator_test.go`. Existing `cmd/library_init_test.go` uses `runF` injection — verify the file builds with the new `InitRequest.Stdout` field; no `runF`-body edits needed.
- [x] 1.7 Run `rg "fmt\.Fprintln\(os\.Stdout" internal/library/` — must return zero matches. *(now fully satisfied at end of Phase 2.)*
- [x] 1.8 Run `rg "fmt\.Fprintf\(os\.Stdout" internal/library/` — must return zero matches. *(zero matches.)*
- [x] 1.9 Run `rg "os\.Stdout" internal/library/` — must return zero matches. The `os` package import stays in `adder.go` and `creator.go` (used by `MkdirAll`, `WriteFile`, `Rename`, etc.). *(now fully satisfied at end of Phase 2.)*

## 2. Phase 2 — Thread `Stdout io.Writer` through `AddResource` / `BatchAddResources` and remove `os.Stdout` write

- [x] 2.1 In `internal/library/adder.go:25`, add a `Stdout io.Writer` field to `AddRequest` (with doc comment).
- [x] 2.2 In `internal/library/adder.go:540`, add a `Stdout io.Writer` field to `BatchAddOptions` (with doc comment).
- [x] 2.3 In `internal/library/adder.go:91-98` (`AddResource` dry-run block), replace the 5 `fmt.Fprintln(os.Stdout, …)` calls with `if opts.Stdout != nil { fmt.Fprintln(opts.Stdout, …) }`.
- [x] 2.4 In `cmd/library_add.go:64-68` (`resourceAdder` interface) — no signature change needed since `Stdout` is a struct field, not a positional param. Document this in a comment.
- [x] 2.5 (no-op — verify, do not edit) At `cmd/library_add.go:82-107` (`libraryAdapter` methods), `(*libraryAdapter).AddResource` already calls `library.AddResource(ctx, *req)` and `(*libraryAdapter).BatchAddResources` already calls `library.BatchAddResources(ctx, opts)`. Both pass the full struct, so the new `Stdout` field flows through automatically. No code change required at the adapter.
- [x] 2.6 In `cmd/library_add.go:352`, update the `AddResource` `AddRequest` literal to include `Stdout: opts.IO.Out`.
- [x] 2.7 In `cmd/library_add.go:525` and `cmd/library_add.go:654`, update both `BatchAddResources` `BatchAddOptions` literals to include `Stdout: opts.IO.Out`.
- [x] 2.8 In `internal/library/adder_test.go`, update test call sites that exercise dry-run paths to pass either `nil` (for tests that don't care about output) or a `bytes.Buffer` (for dry-run output assertion). *(existing keyed struct literals already yield zero-value `Stdout: nil`; dry-run block is now gated; tests pass unchanged.)*
- [x] 2.9 Run `rg "fmt\.Fprintln\(os\.Stdout" internal/library/` — must return zero matches.
- [x] 2.10 Run `mise run test:e2e` — verify the dry-run output still appears in user-facing output (now via the writer).

## 3. Phase 3 — Add `atomicWriteFile` helper and delegate 4 library.yaml write sites

The helper lives in `internal/library/saver.go` next to `SaveLibrary`. It encapsulates `os.WriteFile` + `os.Rename` plus the `EXDEV` fallback. Three library.yaml write sites (`adder.go:333`, `remover.go:193`, `remover.go:226`) currently use the temp+rename pattern and gain **EXDEV fallback** via the helper. One site (`saver.go:33`, `SaveLibrary`) currently uses direct non-atomic `os.WriteFile` and gains **atomicity** (converts to temp+rename) via the helper. `adder.go` and `remover.go` need **no** new imports — they call the package-internal helper that encapsulates `syscall`/`errors`.

- [x] 3.1 Add `errors`, `syscall`, and `io` to the `internal/library/saver.go` import list (in one diff hunk).

- [x] 3.2 In `internal/library/saver.go` (next to `SaveLibrary` at line 15), add the new helper:

  ```go
  // atomicWriteFile writes data to path with perm atomically via the
  // write-temp-then-rename pattern, falling back to copy+remove on
  // syscall.EXDEV (cross-filesystem rename). This is the single source
  // of truth for library.yaml atomic writes; AddResource, RemoveResource,
  // RemovePreset, and SaveLibrary all delegate here.
  func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
      tmpPath := path + ".tmp"
      if err := os.WriteFile(tmpPath, data, perm); err != nil {
          return gerrors.NewFileError(tmpPath, "write", "failed to write temp", err)
      }
      if err := os.Rename(tmpPath, path); err != nil {
          if errors.Is(err, syscall.EXDEV) {
              // Cross-filesystem: copy + remove (see design Decision 5)
              return atomicWriteFileCrossFS(tmpPath, path, perm)
          }
          return gerrors.NewFileError(path, "rename", "failed to update file", err)
      }
      return nil
  }
  ```

  (Implemented verbatim in `internal/library/saver.go`; production code uses a `defaultRenameFunc` / `renameFunc` seam only so the EXDEV test can inject `syscall.EXDEV` deterministically.)

- [x] 3.3 Replace the inline `os.WriteFile` + `os.Rename` block at `internal/library/adder.go:330-335` with `if err := atomicWriteFile(yamlPath, output, 0o600); err != nil { return err }`. Remove the now-unused `tmpPath` local.

- [x] 3.4 Replace the inline `os.WriteFile` + `os.Rename` block at `internal/library/remover.go:190-195` (`removeResourceFromLibrary`) with `if err := atomicWriteFile(yamlPath, output, 0o600); err != nil { return err }`. Remove the now-unused `tmpPath` local.

- [x] 3.5 Replace the inline `os.WriteFile` + `os.Rename` block at `internal/library/remover.go:223-228` (`removePresetFromLibrary`) with `if err := atomicWriteFile(yamlPath, output, 0o600); err != nil { return err }`. Remove the now-unused `tmpPath` local.

- [x] 3.6 Replace the **direct `os.WriteFile`** at `internal/library/saver.go:33` (`SaveLibrary`'s library.yaml write) with `if err := atomicWriteFile(yamlPath, data, 0o600); err != nil { return err }`. This converts `SaveLibrary` from non-atomic direct write to atomic temp+rename (atomicity improvement, not EXDEV fix). No `tmpPath` local to remove (direct WriteFile had none).

- [x] 3.7 In `internal/library/saver_test.go`, add `TestAtomicWriteFile_EXDEV`, `TestAtomicWriteFile_HappyPath`, and `TestAtomicWriteFile_RenameFailNoEXDEV` tests using the test-only `renameFunc` seam (per design Decision 5: deterministic injection of `syscall.EXDEV` when cross-filesystem setup is unreliable in CI).

## 4. Phase 4 — errgroup into `scanDirectory` (recursive sibling-subtree parallelism) + Windows doc

- [x] 4.1 In `internal/library/adder.go` (`scanDirectory`), refactor the `filepath.WalkDir`-based sequential recursion into an `errgroup.SetLimit(scanConcurrencyLimit)`-bounded parallel walk over sibling subtrees. `const scanConcurrencyLimit = 8` declared at file scope near `scanDirectory`; recursive descent via `scanLevel` (one-level `os.ReadDir` + goroutine fan-out on a shared `*errgroup.Group`). Concurrent writes to `result.Orphans` / `result.Conflicts` slice appends and `result.Summary.TotalScanned` integer increment are all guarded by `result.scanMu sync.Mutex` (added as an unexported field on `DiscoverResult`). `isRegistered` and `checkNameConflict` reads against `lib` are lock-free since `lib` is read-only during the scan. Per-file work factored into `processScanFile` so the mutex critical sections cover only the writes.

  Note: the outer `DiscoverOrphans` 4-directory loop is **NOT** wrapped — N=4 fails the skill's errgroup trigger, and `SetLimit(scanConcurrencyLimit)` on the outer loop would be a no-op.

- [x] 4.2 Promote `golang.org/x/sync` from indirect to direct in `go.mod` (was transitive at v0.19.0; promoted automatically once 4.1's `errgroup` import landed).

- [x] 4.3 In `internal/library/refresher_test.go`, add `TestDiscoverOrphans_CtxCancelled` test that:
  - Creates a library with deeply nested directories (4 top-level dirs × 12 levels deep × 200 `.md` files per leaf — 800 files total).
  - Cancels `ctx` after 1ms (chosen empirically because 50ms let the scan complete cleanly on this fixture; 1ms is mid-flight reliably).
  - Asserts the returned error wraps `context.Canceled`.
  - Asserts `DiscoverOrphans` returns within 500ms post-cancel.
  - Verified with `-count=5` (5 consecutive runs pass).

- [x] 4.4 Above the existing `if err := os.MkdirAll(targetDir, 0o750); err != nil {` block at `internal/library/adder.go:104-105`, added the Unix-only doc comment.

- [x] 4.5 Above the existing `os.MkdirAll(dir, 0o750)` block at `internal/library/creator.go:57`, added the analogous Unix-only doc comment.

- [x] 4.6 Above the existing `os.MkdirAll(lib.RootPath, 0o750)` block at `internal/library/saver.go:21`, added the analogous Unix-only doc comment.

- [x] 4.7 `mise run test:race` — clean (no race conditions on `result.Orphans` / `result.Conflicts` slice appends OR `result.Summary.TotalScanned` integer writes under errgroup; no goroutine leaks).

- [x] 4.8 `rg "fmt\.Fprintln\(os\.Stdout" internal/library/` — final guardrail: 0 matches.

## 5. Verification

- [x] 5.1 Run `mise run build` — no broken imports. *(clean, no diagnostics)*
- [x] 5.2 Run `mise run lint` — must report 0 issues. *(`0 issues.`)*
- [x] 5.3 Run `mise run test` — all unit tests pass. *(all internal/* packages pass; race-clean.)*
- [x] 5.4 Run `mise run test:e2e` — E2E tests pass. *(`ok gitlab.com/amoconst/germinator/test/e2e 2.932s`)*
- [x] 5.5 Run `rg "fmt\.Fprintln\(os\.Stdout|fmt\.Fprintf\(os\.Stdout" internal/library/` — must return zero matches. *(zero matches.)*
- [x] 5.6 Run `rg "CreateLibrary\(|AddResource\(|BatchAddResources\(|library\.Init\(" internal/ cmd/ test/` — all call sites pass a writer via the request struct field. *(verified: `cmd/library_init.go:163` and `cmd/library_add.go:534,664` populate `Stdout: opts.IO.Out`; test sites use zero-value `nil`, which gates the dry-run block as a no-op.)*
- [x] 5.7 Run `rg "os\.Rename" internal/library/` — verify the `EXDEV` fallback is in place. *(matches: 1 comment + 1 call site inside `atomicWriteFile` / `defaultRenameFunc` in `saver.go`.)*
- [x] 5.8 Run `rg "errgroup" internal/library/` — verify the errgroup import and use. *(`errgroup` import in `adder.go:14`; `errgroup.WithContext`, `g.SetLimit(scanConcurrencyLimit)`, `g.Wait()` in `scanDirectory`; `g.Go(...)` calls in `scanLevel`.)*
- [x] 5.9 Run `openspec validate fix-library-io-discipline --strict` — change is coherent. *(`Change 'fix-library-io-discipline' is valid`.)*

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

- [ ] 3.1 Add `errors`, `syscall`, and `io` to the `internal/library/saver.go` import list (in one diff hunk).

- [ ] 3.2 In `internal/library/saver.go` (next to `SaveLibrary` at line 15), add the new helper:

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

  func atomicWriteFileCrossFS(tmpPath, path string, perm os.FileMode) error {
      in, err := os.Open(tmpPath) //nolint:gosec // G304: temp file path is internally controlled
      if err != nil {
          return gerrors.NewFileError(path, "rename", "failed to open temp for copy", err)
      }
      defer in.Close()
      out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
      if err != nil {
          return gerrors.NewFileError(path, "rename", "failed to open target for copy", err)
      }
      defer out.Close()
      if _, err := io.Copy(out, in); err != nil {
          return gerrors.NewFileError(path, "rename", "failed to copy across filesystems", err)
      }
      if err := out.Sync(); err != nil {
          return gerrors.NewFileError(path, "rename", "failed to sync target", err)
      }
      if err := os.Remove(tmpPath); err != nil {
          return gerrors.NewFileError(path, "rename", "failed to remove temp", err)
      }
      return nil
  }
  ```

- [ ] 3.3 Replace the inline `os.WriteFile` + `os.Rename` block at `internal/library/adder.go:330-335` with `if err := atomicWriteFile(yamlPath, output, 0o600); err != nil { return err }`. Remove the now-unused `tmpPath` local.

- [ ] 3.4 Replace the inline `os.WriteFile` + `os.Rename` block at `internal/library/remover.go:190-195` (`removeResourceFromLibrary`) with `if err := atomicWriteFile(yamlPath, output, 0o600); err != nil { return err }`. Remove the now-unused `tmpPath` local.

- [ ] 3.5 Replace the inline `os.WriteFile` + `os.Rename` block at `internal/library/remover.go:223-228` (`removePresetFromLibrary`) with `if err := atomicWriteFile(yamlPath, output, 0o600); err != nil { return err }`. Remove the now-unused `tmpPath` local.

- [ ] 3.6 Replace the **direct `os.WriteFile`** at `internal/library/saver.go:33` (`SaveLibrary`'s library.yaml write) with `if err := atomicWriteFile(yamlPath, data, 0o600); err != nil { return err }`. This converts `SaveLibrary` from non-atomic direct write to atomic temp+rename (atomicity improvement, not EXDEV fix). No `tmpPath` local to remove (direct WriteFile had none).

- [ ] 3.7 In `internal/library/saver_test.go` (where existing `SaveLibrary` tests live), add `TestAtomicWriteFile_EXDEV` test that:
  - Sets `TMPDIR` (or `t.Setenv("TMPDIR", ...)`) to a directory on a different filesystem than the target path when the test environment supports it (e.g., a tmpdir under `/tmp` while the target is on the test filesystem); otherwise stubs the helper via a test-only `renameFunc` seam.
  - Calls `atomicWriteFile(targetPath, []byte("data"), 0o600)` and verifies success.
  - Reads the target back and verifies content equality.
  - Cleans up any temp files left behind.
  - Also verifies `rg "os\.Rename" internal/library/` returns exactly 1 match (inside the helper) and `rg "atomicWriteFile" internal/library/` returns 5 matches (1 def + 4 callers).

## 4. Phase 4 — errgroup into `scanDirectory` (recursive sibling-subtree parallelism) + Windows doc

- [ ] 4.1 In `internal/library/adder.go:845` (`scanDirectory`), refactor the `filepath.WalkDir`-based sequential recursion into an `errgroup.SetLimit(scanConcurrencyLimit)`-bounded parallel walk over sibling subtrees. Declare `const scanConcurrencyLimit = 8` at file scope near the `scanDirectory` signature so the cap is grep-able. Each subtree goroutine processes its own children; the merged result is appended to the shared `*DiscoverResult` via a `sync.Mutex` guarding `result.Orphans` / `result.Conflicts` slice appends AND `result.Summary.TotalScanned` integer writes. Idiomatic pattern per `golang-cli-architecture`:

  ```go
  g, ctx := errgroup.WithContext(ctx)
  g.SetLimit(scanConcurrencyLimit)

  // Walk one level to parallelize sibling-subtree descent.
  entries, err := os.ReadDir(dirPath)
  if err != nil { return err }
  for _, entry := range entries {
      full := filepath.Join(dirPath, entry.Name())
      if entry.IsDir() {
          // 5th arg: DiscoverOptions is unused by the recursive walker; pass zero value.
          g.Go(func() error { return scanDirectory(ctx, full, resType, lib, DiscoverOptions{}, result) })
      } else if strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
          g.Go(func() error {
              if cerr := ctx.Err(); cerr != nil { return fmt.Errorf("scan: %w", cerr) }
              // process file under mutex; mirror existing per-file logic
              return nil
          })
      }
  }
  return g.Wait()
  ```

  Note: the outer `DiscoverOrphans` 4-directory loop (lines 802-810) is **NOT** wrapped — N=4 fails the skill's errgroup trigger, and `SetLimit(scanConcurrencyLimit)` on the outer loop would be a no-op.

- [ ] 4.2 Add `golang.org/x/sync/errgroup` to `go.mod` (verify whether it's already a transitive dependency; add directly if not).
- [ ] 4.3 In `internal/library/refresher_test.go` (where existing `TestDiscoverOrphans` lives at `refresher_test.go:232`), add `TestDiscoverOrphans_CtxCancelled` test that:
  - Creates a library with deeply nested directories (e.g., `skills/sub1/.../sub10/`, at minimum 10 levels).
  - Cancels `ctx` after 50ms.
  - Asserts the returned error wraps `context.Canceled`.
  - Asserts `DiscoverOrphans` returns within 200ms post-cancel (post-change errgroup parallelizes sibling-subtree descent; pre-change is bounded by the slowest subtree's walk time, typically >500ms on deep trees).
- [ ] 4.4 Above the existing `if err := os.MkdirAll(targetDir, 0o750); err != nil {` block at `internal/library/adder.go:104-105` (no behavior change), add a doc comment:

  ```go
  // Unix permission bits (0o750) are no-ops on Windows; Windows support is out of scope.
  if err := os.MkdirAll(targetDir, 0o750); err != nil { ... }
  ```

- [ ] 4.5 Above the existing `os.MkdirAll(dir, 0o750)` block at `internal/library/creator.go:57`, add the analogous Unix-only doc comment.
- [ ] 4.6 Above the existing `os.MkdirAll(libraryDir, 0o750)` block at `internal/library/saver.go:21` (MkdirAll site, matches the persistence spec.md site list), add the analogous Unix-only doc comment. Perm-bit unification between `SaveLibrary` (`0o600`) and `CreateLibrary` (`0o644`) for `library.yaml` is out of scope per the persistence spec delta.
- [ ] 4.7 Run `mise run test:race` — verify no race conditions on `result.Orphans` / `result.Conflicts` slice appends OR `result.Summary.TotalScanned` integer writes under errgroup; verify no goroutine leaks.
- [ ] 4.8 Run `rg "fmt\.Fprintln\(os\.Stdout" internal/library/` again as a final guardrail.

## 5. Verification

- [ ] 5.1 Run `mise run build` — no broken imports.
- [ ] 5.2 Run `mise run lint` — must report 0 issues.
- [ ] 5.3 Run `mise run test` — all unit tests pass.
- [ ] 5.4 Run `mise run test:e2e` — E2E tests pass.
- [ ] 5.5 Run `rg "fmt\.Fprintln\(os\.Stdout|fmt\.Fprintf\(os\.Stdout" internal/library/` — must return zero matches.
- [ ] 5.6 Run `rg "CreateLibrary\(|AddResource\(|BatchAddResources\(|library\.Init\(" internal/ cmd/ test/` — all call sites pass a writer via the request struct field (`Stdout: opts.IO.Out` or `Stdout: nil` for tests).
- [ ] 5.7 Run `rg "os\.Rename" internal/library/` — verify the `EXDEV` fallback is in place.
- [ ] 5.8 Run `rg "errgroup" internal/library/` — verify the errgroup import and use.
- [ ] 5.9 Run `openspec validate fix-library-io-discipline --strict` — change is coherent.

# Tasks — Fix library I/O discipline

Each task ends with `mise run check` passing. Tasks are grouped by phase and ordered so each commit is independently testable.

## 1. Phase 1 — Add `io.Writer` to `CreateLibrary` and remove `os.Stdout` write

- [ ] 1.1 In `internal/library/creator.go:38`, replace the `fmt.Fprintln(os.Stdout, "Would create library at:", opts.Path)` block with a call to the new `stdout` parameter. Specifically: `if stdout != nil { fmt.Fprintln(stdout, "Would create library at:", opts.Path) }`.
- [ ] 1.2 In `internal/library/creator.go`, change `func CreateLibrary(opts CreateOptions) error` to `func CreateLibrary(opts CreateOptions, stdout io.Writer) error`.
- [ ] 1.3 In `internal/library/creator_test.go`, update all `CreateLibrary` test calls to pass `nil` (or a `bytes.Buffer` for dry-run tests).
- [ ] 1.4 In `cmd/library_init.go:144`, update the `CreateLibrary` call to pass `opts.IO.Out` as the new `stdout` parameter.
- [ ] 1.5 Run `rg "fmt\.Fprintln\(os\.Stdout" internal/library/` — must return zero matches.
- [ ] 1.6 Run `rg "fmt\.Fprintf\(os\.Stdout" internal/library/` — must return zero matches.
- [ ] 1.7 Run `rg "os\.Stdout" internal/library/` — only imports (which can be removed) should remain.

## 2. Phase 2 — Add `io.Writer` to `AddResource` / `BatchAddResources` and remove `os.Stdout` write

- [ ] 2.1 In `internal/library/adder.go:92`, replace the `fmt.Fprintln(os.Stdout, "Would add resource:", resourceKey)` block with a call to the new `stdout` parameter.
- [ ] 2.2 In `internal/library/adder.go`, change `func AddResource(ctx context.Context, opts AddOptions) error` to `func AddResource(ctx context.Context, opts AddOptions, stdout io.Writer) error`.
- [ ] 2.3 In `internal/library/adder.go`, change `func BatchAddResources(ctx context.Context, opts *BatchAddOptions) (*BatchAddResult, error)` to `func BatchAddResources(ctx context.Context, opts *BatchAddOptions, stdout io.Writer) (*BatchAddResult, error)`.
- [ ] 2.4 In `internal/library/adder_test.go`, update all `AddResource` / `BatchAddResources` test calls to pass `nil` (or a `bytes.Buffer` for dry-run tests).
- [ ] 2.5 In `cmd/library_add.go:347`, update the `AddResource` call to pass `opts.IO.Out` as the new `stdout` parameter.
- [ ] 2.6 In `cmd/library_add.go`, update any `BatchAddResources` call to pass `opts.IO.Out`.
- [ ] 2.7 Run `rg "fmt\.Fprintln\(os\.Stdout" internal/library/` — must return zero matches.
- [ ] 2.8 Run `mise run test:e2e` — verify the dry-run output still appears in user-facing output (now via the writer).

## 3. Phase 3 — Add `EXDEV` fallback to `os.Rename`

- [ ] 3.1 In `internal/library/adder.go:333`, replace the `os.Rename` block with the `EXDEV` fallback. Specifically:

  ```go
  if err := os.Rename(tmpPath, yamlPath); err != nil {
      if errors.Is(err, syscall.EXDEV) {
          // Cross-filesystem: copy + remove
          in, err := os.Open(tmpPath)
          if err != nil {
              return core.NewFileError(yamlPath, "rename", "failed to open temp for copy", err)
          }
          defer in.Close()
          out, err := os.OpenFile(yamlPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o640)
          if err != nil {
              return core.NewFileError(yamlPath, "rename", "failed to open target for copy", err)
          }
          defer out.Close()
          if _, err := io.Copy(out, in); err != nil {
              return core.NewFileError(yamlPath, "rename", "failed to copy across filesystems", err)
          }
          if err := out.Sync(); err != nil {
              return core.NewFileError(yamlPath, "rename", "failed to sync target", err)
          }
          if err := os.Remove(tmpPath); err != nil {
              return core.NewFileError(yamlPath, "rename", "failed to remove temp", err)
          }
      } else {
          return core.NewFileError(yamlPath, "rename", "failed to update library.yaml", err)
      }
  }
  ```

- [ ] 3.2 Add `syscall` and `errors` to the `internal/library/adder.go` import list (verify they're not already present).
- [ ] 3.3 In `internal/library/adder_test.go`, add `TestSave_CrossFilesystem_EXDEV` test that:
  - Sets `TMPDIR` to a directory on a different mount point than the library path.
  - Calls `Save` (or whatever the public method is) and verifies success.
  - Cleans up the temp directory after the test.

## 4. Phase 4 — errgroup for `DiscoverOrphans` + Windows doc

- [ ] 4.1 In `internal/library/adder.go:785`, wrap the directory scan loop in `errgroup.WithContext` with an 8-worker semaphore:

  ```go
  g, ctx := errgroup.WithContext(ctx)
  sem := make(chan struct{}, 8)
  for _, dir := range dirs {
      sem <- struct{}{}
      g.Go(func() error {
          defer func() { <-sem }()
          if ctx.Err() != nil {
              return ctx.Err()
          }
          // process dir
          return nil
      })
  }
  return g.Wait()
  ```

- [ ] 4.2 Add `golang.org/x/sync/errgroup` to `go.mod` (verify whether it's already a transitive dependency; add directly if not).
- [ ] 4.3 In `internal/library/adder_test.go`, add `TestDiscoverOrphans_CtxCancelled` test that:
  - Creates a library with 100+ nested directories.
  - Cancels the `ctx` after 50ms.
  - Asserts `DiscoverOrphans` returns within 100ms (with `context.Canceled` wrapped).
- [ ] 4.4 In `internal/library/adder.go:105`, add a comment documenting the Unix-only permission bits:

  ```go
  // Unix permission bits (0o750) are ignored on Windows; Windows support is out of scope.
  _ = os.MkdirAll(dir, 0o750)
  ```

- [ ] 4.5 In `internal/library/creator.go:33`, add a similar comment for the `0750` directory creation.
- [ ] 4.6 In `internal/library/saver.go`, add a similar comment for the `0640` file creation.
- [ ] 4.7 Run `mise run test:race` — verify no goroutine leaks in the errgroup path.

## 5. Verification

- [ ] 5.1 Run `mise run build` — no broken imports.
- [ ] 5.2 Run `mise run lint` — must report 0 issues.
- [ ] 5.3 Run `mise run test` — all unit tests pass.
- [ ] 5.4 Run `mise run test:e2e` — E2E tests pass.
- [ ] 5.5 Run `rg "fmt\.Fprintln\(os\.Stdout|fmt\.Fprintf\(os\.Stdout" internal/library/` — must return zero matches.
- [ ] 5.6 Run `rg "CreateLibrary\(|AddResource\(|BatchAddResources\(" internal/ cmd/ test/` — all call sites pass a writer (or `nil` for tests).
- [ ] 5.7 Run `rg "os\.Rename" internal/library/` — verify the `EXDEV` fallback is in place.
- [ ] 5.8 Run `rg "errgroup" internal/library/` — verify the errgroup import and use.
- [ ] 5.9 Run `openspec validate fix-library-io-discipline --strict` — change is coherent.

## 6. Archive

- [ ] 6.1 Apply spec deltas via `osc-sync-specs`.
- [ ] 6.2 Archive this change via `osc-archive-change fix-library-io-discipline`.
- [ ] 6.3 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.

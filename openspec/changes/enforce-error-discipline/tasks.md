# Tasks — Enforce error-handling discipline

Each task ends with `mise run check` passing. Tasks are grouped by phase and ordered so each commit is independently testable.

## 1. Phase 1 — Exit code + dispatch set + dead code + UsageError

- [ ] 1.1 In `internal/cmdutil/exit.go:73`, change `*core.NotFoundError` mapping from `ExitCodeUsage` to `ExitCodeError` (1).
- [ ] 1.2 In `internal/cmdutil/exit.go:24`, replace the `cobraUsagePrefixes` substring list with `errors.As`-based dispatch. Use `errors.As(err, &pflagErr)` and `errors.As(err, &cobraFlagErr)` first; fall back to a narrow prefix list of 3-4 stable Cobra strings.
- [ ] 1.3 In `internal/cmdutil/exit_test.go:58`, update the `*core.NotFoundError` expectation to expect `ExitCodeError` (1). Add a test for the new `*core.UsageError` mapping to `ExitCodeUsage` (2).
- [ ] 1.4 In `internal/output/errors.go:21-50`, add `case *core.InitializeError` to the `FormatError` switch. Render `"Error: initialize failed: <ref>"`.
- [ ] 1.5 In `internal/output/errors.go:21-50`, add `case *core.UsageError` to the `FormatError` switch. Render `"Error: <flag>: <reason>"`.
- [ ] 1.6 In `internal/output/exporter.go:172`, delete `var _ = io.EOF`. Remove the `io` import on line 7 (verify zero other uses via `rg "io\." internal/output/exporter.go`).
- [ ] 1.7 In `internal/core/errors.go`, add `UsageError` type, `NewUsageError(flag, reason string) *UsageError` constructor, and `Error()` method per the `errors-typed-errors/spec.md` delta.

## 2. Phase 2 — Single-handling rule cleanup

- [ ] 2.1 In `cmd/library_remove.go:170`, delete the inline `output.FormatError(opts.IO, err)` call. Return the wrapped error so `main.go:43` formats it once.
- [ ] 2.2 In `cmd/library_remove.go:178`, delete the inline `output.FormatError(opts.IO, err)` call (same pattern as 2.1).
- [ ] 2.3 In `cmd/library_add.go:343`, delete the inline `output.FormatError(opts.IO, err)` call.
- [ ] 2.4 In `cmd/library_add.go:537`, delete the inline `output.FormatError(opts.IO, err)` call.
- [ ] 2.5 In `cmd/library_add.go:681`, delete the inline `output.FormatError(opts.IO, err)` call.
- [ ] 2.6 In `cmd/library_add.go:694`, delete the inline `output.FormatError(opts.IO, err)` call.
- [ ] 2.7 In `cmd/init.go:205`, delete the inline `output.FormatError(opts.IO, partialErr)` call.
- [ ] 2.8 In `cmd/init.go:209`, delete the inline `output.FormatError(opts.IO, partialErr)` call.
- [ ] 2.9 Run `rg "output\.FormatError" cmd/` — must show only the `main.go` central handler and any non-error-path calls (e.g., `cmd/library_init.go` may use it for a non-error path).
- [ ] 2.10 Run `mise run test:e2e` — capture stderr and verify no double-output for any of the 7 sites.

## 3. Phase 3 — Type migration + error chain

- [ ] 3.1 In `internal/library/remover.go:83`, replace `gerrors.NewFileError(opts.LibraryPath, "access", fmt.Sprintf("resource %s not found", opts.Ref), nil)` with `gerrors.NewNotFoundError("library ref", opts.Ref)`.
- [ ] 3.2 In `internal/library/remover.go:103`, replace the silent `os.IsNotExist` swallow with a surface-as-`NotFoundError` path. Specifically: `if errors.Is(err, os.ErrNotExist) { return nil, core.NewNotFoundError("library file", path) }`.
- [ ] 3.3 In `internal/library/resolver.go:67`, replace `gerrors.NewConfigError("preset", name, "preset not found")` with `gerrors.NewNotFoundError("preset", name)`.
- [ ] 3.4 In `cmd/init.go:178`, remove the cross-package translation from `*core.ConfigError` to `*core.NotFoundError` (no longer needed; the resolver returns `NotFoundError` directly).
- [ ] 3.5 In `cmd/library_add.go:534`, add `ErrorType string` and `Cause error` fields to `BatchFailureInfo` per the `errors-enhanced-errors/spec.md` delta. Update the `core.NewOperationError("add", f.Source, errors.New(f.Error))` call to populate the new fields.
- [ ] 3.6 In `cmd/library_add.go:82`, replace `errEmptyResources = errors.New("flag needs an argument: --resources ...")` with `errEmptyResources = core.NewUsageError("--resources", "must be non-empty list of refs")`.
- [ ] 3.7 Update `cmd/library_add.go:120` and any other call site of `errEmptyResources` to handle the new typed error. Specifically: return the error directly (not wrap it in `OperationError`); `main.go` formats it via `FormatError`'s new `UsageError` case.
- [ ] 3.8 Run `rg "NewFileError.*not found"` — must return zero matches.
- [ ] 3.9 Run `rg "NewConfigError.*preset not found"` — must return zero matches.

## 4. Phase 4 — Trivial folds + spec deltas

- [ ] 4.1 Create `internal/config/scaffold.go` with `func WriteDefault(path string, force bool) error` that wraps `os.Stat` / `os.MkdirAll(0o750)` / `os.WriteFile(0o600)`.
- [ ] 4.2 In `cmd/library_init.go:144`, replace the inline `os.Stat` / `os.MkdirAll(0o750)` / `os.WriteFile(0o600)` block with a call to `config.WriteDefault(opts.Path, opts.Force)`.
- [ ] 4.3 In `cmd/canonicalize.go:94`, add `core.ValidateDocumentType(opts.Type)` call before the `default` branch in `validateCanonicalDoc`. Return `*core.UsageError` on invalid type so the user gets an exit-code-2 error before the parse fails.
- [ ] 4.4 In `cmd/library.go:11`, drop the dead `runF func(*libraryResourcesOptions) error` parameter from `NewLibraryCommand`. Update all call sites to omit the parameter.
- [ ] 4.5 Apply spec deltas at `openspec/changes/enforce-error-discipline/specs/*` via `osc-sync-specs`.

## 5. Verification

- [ ] 5.1 Run `mise run build` — no broken imports.
- [ ] 5.2 Run `mise run lint` — must report 0 issues.
- [ ] 5.3 Run `mise run test` — all unit tests pass; the new `UsageError` constructor and `InitializeError` dispatch case have test coverage ≥ 70%.
- [ ] 5.4 Run `mise run test:e2e` — E2E tests pass; verify no double-output in captured stderr.
- [ ] 5.5 Run `rg "output\.FormatError" cmd/` — only `main.go` and any non-error-path callers should appear.
- [ ] 5.6 Run `rg "errEmptyResources\s*=" cmd/` — must show a single declaration using `core.NewUsageError`.
- [ ] 5.7 Run `openspec validate enforce-error-discipline --strict` — change is coherent.
- [ ] 5.8 Manually test the exit-code change: run `germinator library show nonexistent-ref` and verify exit code is 1 (was 2). Update any test fixtures that hard-code `exit 2` for not-found scenarios.

## 6. Archive

- [ ] 6.1 Update CHANGELOG.md with a **BREAKING** entry: "`*core.NotFoundError` now maps to exit code 1 (operational error) instead of exit code 2 (usage error). Scripts that special-case `exit 2` for not-found scenarios must update to `exit 1`."
- [ ] 6.2 Archive this change via `osc-archive-change enforce-error-discipline`.
- [ ] 6.3 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.

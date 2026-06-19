# Tasks — Migrate init command

**Slice 5 of 9.** Migrates `cmd/init.go` to the new pattern. Wires `core.PartialSuccessError` into the error pipeline (already done in change-1's `cmdutil.ExitCodeFor` and `output.FormatError`; this change adds the integration test). Preserves exit-code semantics (0 if any succeeded, 1 if all failed).

Each task ends with `mise run check` passing.

## 5.1 Verify foundation PartialSuccessError wiring

- [ ] 5.1.1 Confirm `cmdutil.ExitCodeFor(&core.PartialSuccessError{Succeeded: 1, Failed: 1})` returns `ExitCodeSuccess` (0) — already tested in change-1 task 1.1.16
- [ ] 5.1.2 Confirm `output.FormatError(io, &core.PartialSuccessError{Succeeded: 1, Failed: 1})` writes `partial success: 1 succeeded, 1 failed` to stderr followed by per-resource error lines — already tested in change-1 task 1.1.11
- [ ] 5.1.3 Confirm `cmdutil.ExitCodeFor(&core.PartialSuccessError{Succeeded: 0, Failed: 2})` returns `ExitCodeError` (1) — already tested in change-1

## 5.2 Migrate `cmd/init.go`

- [ ] 5.2.1 In `cmd/init.go`, define `initOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Initializer func() (Initializer, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`
- [ ] 5.2.2 Declare the `Initializer` interface in `cmd/init.go` (one method: `Initialize(ctx, *InitializeRequest) ([]core.InitializeResult, error)`); import `InitializeRequest` from `internal/application/requests.go`
- [ ] 5.2.3 Implement `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`:
  - Add flags: `--platform`, `--output-dir`, `--resources` ([]string), `--preset`, `--dry-run`, `--force`
  - In `RunE`: construct `opts`, populate from `f.IOStreams`, `c.Context()`, `f.Library`, `f.Initializer`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runInit(opts)`
- [ ] 5.2.4 Implement `runInit(opts *initOptions) error`:
  - Validate platform via `core.ValidatePlatform(opts.Platform)`
  - If `opts.Preset != ""`: call `lib.ResolvePreset(opts.Ctx, opts.Preset)` to expand to refs; merge with `opts.Refs`
  - Construct `&InitializeRequest{Refs: refs, Platform: opts.Platform, OutputDir: opts.OutputDir, DryRun: opts.DryRun, Force: opts.Force}`
  - Call `f.Initializer().Initialize(opts.Ctx, req)` to get `([]core.InitializeResult, error)`
  - If the transport error is non-nil: return it wrapped
  - Count successes and failures from the result slice
  - If `Succeeded > 0 && Failed == 0`: return nil (exit 0)
  - If `Succeeded > 0 && Failed > 0`: build `[]core.InitializeError` from the failures and return `core.NewPartialSuccessError(succeeded, failed, errs)` (exit 0)
  - If `Succeeded == 0 && Failed > 0`: return `core.NewPartialSuccessError(0, failed, errs)` (exit 1)
  - Print per-resource status to `opts.IO.Out` (successes) or `opts.IO.ErrOut` via `output.FormatError` (failures)
- [ ] 5.2.5 Update `internal/service/initializer.go` to return `[]core.InitializeResult` properly (the file itself is deleted in change-7)
- [ ] 5.2.6 Convert `cmd/init_test.go` (or add new tests) to `iostreams.Test()` + `runF` injection

## 5.3 Add partial-success test cases

- [ ] 5.3.1 Add test case: all-success → `runInit(opts)` returns nil; `cmdutil.ExitCodeFor(nil) == 0`
- [ ] 5.3.2 Add test case: partial success (1 succeeded, 1 failed) → `runInit(opts)` returns `*core.PartialSuccessError{Succeeded: 1, Failed: 1}`; `cmdutil.ExitCodeFor(err) == 0`; `output.FormatError(io, err)` writes `partial success: 1 succeeded, 1 failed`
- [ ] 5.3.3 Add test case: all-failed (0 succeeded, 2 failed) → `runInit(opts)` returns `*core.PartialSuccessError{Succeeded: 0, Failed: 2}`; `cmdutil.ExitCodeFor(err) == 1`
- [ ] 5.3.4 Add test case: preset expansion → `--preset git-workflow` expands to refs; each ref processed; partial-success logic applies

## 5.4 Update delta specs

- [ ] 5.4.1 Update `openspec/changes/scaffold-cli-foundation/specs/cli/init-command/spec.md` (if it was created in change-1) to mark the new pattern as **fulfilled by this change**
- [ ] 5.4.2 Update `openspec/changes/migrate-init-command/specs/library/partial-initialization/spec.md` (this change) to reflect the new `Initialize` contract

## 5.5 Verification

- [ ] 5.5.1 Run `mise run lint` — confirm no new violations
- [ ] 5.5.2 Run `mise run test` — confirm all unit tests pass (including the new partial-success cases)
- [ ] 5.5.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 5.5.4 Run `mise run test:coverage` — confirm coverage for `cmd/init.go` ≥ 70%
- [ ] 5.5.5 Run `mise run test:e2e` — confirm E2E tests for init pass (including partial-success scenarios)
- [ ] 5.5.6 Smoke-test:
  - `germinator init --platform opencode --resources skill/commit` → exit 0, success message
  - `germinator init --platform opencode --resources skill/commit,skill/invalid` → exit 0 with `partial success: 1 succeeded, 1 failed`
  - `germinator init --platform opencode --resources skill/invalid1,skill/invalid2` → exit 1 with `partial success: 0 succeeded, 2 failed`
  - `germinator init --platform opencode --preset <existing-preset>` → expand and process
  - `germinator init --platform opencode --preset <existing-preset> --dry-run` → preview only
- [ ] 5.5.7 Update `cmd/AGENTS.md` with the `init` example, including the partial-success pattern
- [ ] 5.5.8 Confirm `legacyBridge` still works for non-migrated commands

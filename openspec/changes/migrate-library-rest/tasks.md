# Tasks — Migrate remaining library commands and delete legacy shell

**Slice 7 of 9.** Migrates the four remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) and deletes the entire legacy shell (`internal/service/`, `internal/application/`, `legacyBridge`, `cmd/error_formatter.go`, `cmd/verbose.go`). **Structural turning point** of the migration.

Each task ends with `mise run check` passing.

## 7.1 Migrate `cmd/library/init.go`

- [ ] 7.1.1 In `cmd/library/init.go`, define `libraryInitOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Path string`, `Force bool`, `DryRun bool`, `Output string`
- [ ] 7.1.2 Declare the `Library` interface in `cmd/library/init.go` with methods called: `Init(ctx, *InitRequest) error`
- [ ] 7.1.3 Implement `NewCmdLibraryInit(f *cmdutil.Factory, runF func(*libraryInitOptions) error) *cobra.Command`:
  - Add flags: `--path`, `--force`, `--dry-run`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runLibraryInit(opts)`
- [ ] 7.1.4 Implement `runLibraryInit(opts *libraryInitOptions) error`:
  - Call `lib.Init(opts.Ctx, &InitRequest{Path: opts.Path, Force: opts.Force, DryRun: opts.DryRun})`
  - Dispatch on `opts.Output` for the result
- [ ] 7.1.5 Convert `cmd/library/init_test.go` to `iostreams.Test()` + `runF` injection
- [ ] 7.1.6 Run `mise run check`

## 7.2 Migrate `cmd/library/refresh.go`

- [ ] 7.2.1 In `cmd/library/refresh.go`, define `refreshOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `DryRun bool`, `Force bool`, `Output string`
- [ ] 7.2.2 Declare the `Library` interface with methods called: `Refresh(ctx, *RefreshRequest) (*RefreshResult, error)`
- [ ] 7.2.3 Implement `NewCmdRefresh(f, runF)` and `runRefresh(opts)`:
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Call `lib.Refresh(opts.Ctx, &RefreshRequest{DryRun: opts.DryRun, Force: opts.Force})`
  - Dispatch on `opts.Output` for the result (per-resource status: updated, unchanged, conflict)
- [ ] 7.2.4 Convert `cmd/library/refresh_test.go` to `iostreams.Test()` + `runF` injection
- [ ] 7.2.5 Run `mise run check`

## 7.3 Migrate `cmd/library/remove.go`

- [ ] 7.3.1 In `cmd/library/remove.go`, define `removeOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `ResourceType string`, `ResourceName string`, `PresetName string`, `Force bool`, `Output string`
- [ ] 7.3.2 Declare the `Library` interface with methods called: `RemoveResource(ctx, *RemoveResourceRequest) error`, `RemovePreset(ctx, *RemovePresetRequest) error`
- [ ] 7.3.3 Implement `NewCmdRemove(f *cmdutil.Factory, runF func(*removeOptions) error) *cobra.Command`:
  - Add sub-commands: `library remove resource` (with `--type` + `--name`) and `library remove preset` (with `--name`)
  - Add `--force` flag
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` on the parent
  - Populate `opts` in `RunE` based on which sub-command was invoked
  - Call `runF(opts)` if non-nil, else `runRemove(opts)`
- [ ] 7.3.4 Implement `runRemove(opts *removeOptions) error`:
  - If `opts.PresetName != ""`: call `lib.RemovePreset(...)`
  - Else: call `lib.RemoveResource(...)` with type + name
  - Dispatch on `opts.Output` for the result
- [ ] 7.3.5 Convert `cmd/library/remove_test.go` (or add tests) to `iostreams.Test()` + `runF` injection; cover both resource and preset removal
- [ ] 7.3.6 Run `mise run check`

## 7.4 Migrate `cmd/library/validate.go`

- [ ] 7.4.1 In `cmd/library/validate.go`, define `libraryValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Fix bool`, `Output string`
- [ ] 7.4.2 Declare the `Library` interface with methods called: `Validate(ctx, *ValidateRequest) (*ValidateResult, error)`
- [ ] 7.4.3 Implement `NewCmdLibraryValidate(f, runF)` and `runLibraryValidate(opts)`:
  - Add `--fix` flag
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Call `lib.Validate(opts.Ctx, &ValidateRequest{Fix: opts.Fix})`
  - If validation errors exist, render each via `output.FormatError`
  - Dispatch on `opts.Output` for the result
- [ ] 7.4.4 Convert `cmd/library/validate_test.go` to `iostreams.Test()` + `runF` injection; cover both with and without `--fix`
- [ ] 7.4.5 Run `mise run check`

## 7.5 Delete legacy shell

- [ ] 7.5.1 Run `rg "internal/service" .` to verify zero remaining references; delete `internal/service/` directory tree (10+ files including `transformer.go`, `initializer.go`, all `*_test.go`, all `*_mock_test.go`)
- [ ] 7.5.2 Run `rg "internal/application" .` to verify zero remaining references; delete `internal/application/` directory tree (3 files: `interfaces.go`, `requests.go`, `results.go`)
- [ ] 7.5.3 Delete `legacyBridge` shim from `main.go` (struct definition + population code)
- [ ] 7.5.4 Delete `cmd/error_formatter.go` and `cmd/verbose.go`
- [ ] 7.5.5 Delete `internal/service/*_mock_test.go` files (already deleted with `internal/service/` in task 7.5.1; this task is a no-op confirmation)
- [ ] 7.5.6 Run `rg "ServiceContainer|CommandConfig|ErrorFormatter|Verbosity" cmd/ main.go` to verify zero remaining references to legacy types
- [ ] 7.5.7 Update every `cmd/library/*.go` file to import the request/result types from new locations (if previously imported from `internal/application/requests.go`, move to per-package DTOs or to `internal/contracts/` if needed; otherwise keep as private types in the command file)

## 7.6 Update delta specs (final fulfillment)

- [ ] 7.6.1 Update `openspec/changes/scaffold-cli-foundation/specs/application/dependency-injection/spec.md`: mark `ServiceContainer` removal as **fulfilled**
- [ ] 7.6.2 Update `openspec/changes/scaffold-cli-foundation/specs/cli/exit-codes/spec.md`: mark `CategorizeError` enum removal as **fulfilled**
- [ ] 7.6.3 Update `openspec/changes/scaffold-cli-foundation/specs/cli/framework/spec.md`: confirm `CommandConfig` removal (was fulfilled in change-2)
- [ ] 7.6.4 Update `openspec/changes/scaffold-cli-foundation/specs/cli/verbose-output/spec.md`: mark `VerbosePrint` removal as **fulfilled**
- [ ] 7.6.5 Update `openspec/changes/scaffold-cli-foundation/specs/cli/error-formatting/spec.md`: mark `ErrorFormatter` removal as **fulfilled**

## 7.7 Verification

- [ ] 7.7.1 Run `mise run lint` — confirm no new violations
- [ ] 7.7.2 Run `mise run test` — confirm all unit tests pass (no mocks needed)
- [ ] 7.7.3 Run `mise run build` — confirm `bin/germinator` builds without `internal/service/` or `internal/application/`
- [ ] 7.7.4 Run `mise run test:coverage` — confirm coverage maintained for `cmd/library/` ≥ 70%
- [ ] 7.7.5 Run `mise run test:full` (unit + e2e)
- [ ] 7.7.6 Smoke-test every library subcommand:
  - `germinator library init --path /tmp/lib --dry-run`
  - `germinator library resources`
  - `germinator library resources --output json`
  - `germinator library presets`
  - `germinator library show <ref>`
  - `germinator library add <file> --type skill --name test`
  - `germinator library create preset <name> --resources skill/x`
  - `germinator library refresh`
  - `germinator library refresh --dry-run`
  - `germinator library remove resource skill/test --force`
  - `germinator library remove preset <name> --force`
  - `germinator library validate`
  - `germinator library validate --fix`
- [ ] 7.7.7 Verify `cmd/`, `internal/library/`, `internal/core/`, `internal/cmdutil/`, `internal/iostreams/`, `internal/output/` have no imports of `internal/service/` or `internal/application/`
- [ ] 7.7.8 Update `cmd/library/AGENTS.md` to document all seven library subcommands in their final form
- [ ] 7.7.9 Confirm non-library commands (`adapt`, `validate`, `canonicalize`, `init`, `config`, `completion`, `version`) still work

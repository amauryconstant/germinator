# Tasks — Migrate library add and library create

**Slice 6 of 9.** Migrates `cmd/library/add.go` (with three modes: explicit, discover, batch) and `cmd/library/create.go`. Adds `core.CanInstallResource` to `internal/core/rules.go`. Adds `ctx` to library I/O functions for cancellation safety.

Each task ends with `mise run check` passing.

## 6.1 Add `core.CanInstallResource`

- [ ] 6.1.1 In `internal/core/rules.go`, add `CanInstallResource(ref string) error`:
  - Use `strings.Cut(ref, "/")` to parse `type/name`
  - If no `/` found or `type` is empty: return `*core.ValidationError{Message: "ref must be type/name"}`
  - If `name` is empty: return `*core.ValidationError{Message: "ref name must be non-empty"}`
  - If `type` is not in `{skill, agent, command, memory}`: return `*core.ValidationError{Message: "ref type must be one of skill, agent, command, memory"}`
  - Otherwise return nil
- [ ] 6.1.2 Add test cases in `internal/core/rules_test.go`:
  - Valid: `skill/commit`, `agent/reviewer`, `command/build`, `memory/project`
  - Invalid: `skills/commit` (wrong type), `skill/` (empty name), `skill` (no slash), `` (empty), `/commit` (empty type)
- [ ] 6.1.3 Confirm `internal/core/rules.go` still imports only stdlib (no `internal/library/`) — depguard enforces

## 6.2 Add `ctx` to library I/O functions

- [ ] 6.2.1 Update `internal/library/loader.go` (`LoadLibrary`): add `ctx context.Context` as the first parameter; check `ctx.Err()` after each I/O operation; return wrapped `ctx.Err()` on cancellation
- [ ] 6.2.2 Update `internal/library/discovery.go` (`DiscoverOrphans`): add `ctx context.Context` as the first parameter; check `ctx.Err()` between files; return partial results + `ctx.Err()` on cancellation
- [ ] 6.2.3 Update `internal/library/adder.go` (`BatchAddResources`): add `ctx context.Context` as the first parameter; check `ctx.Err()` after each file in the loop; return partial results + `ctx.Err()` on cancellation
- [ ] 6.2.4 Update all callers of these three functions (currently only `cmd/library/add.go` and `cmd/library_init.go`) to pass `ctx`
- [ ] 6.2.5 Run `mise run check`; confirm zero issues from the API change

## 6.3 Migrate `cmd/library/add.go`

- [ ] 6.3.1 In `cmd/library/add.go`, define `addOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `InputPaths []string`, `Name string`, `Description string`, `Type string`, `Platform string`, `Discover bool`, `Batch bool`, `Force bool`, `DryRun bool`, `Output string`
- [ ] 6.3.2 Declare the `Library` interface in `cmd/library/add.go` with methods called: `AddResource(ctx, *AddRequest) error`, `DiscoverOrphans(ctx, opts) ([]Orphan, error)`, `BatchAddResources(ctx, opts) ([]AddResult, error)`
- [ ] 6.3.3 Implement `NewCmdAdd(f *cmdutil.Factory, runF func(*addOptions) error) *cobra.Command`:
  - Add flags: `--name`, `--description`, `--type`, `--platform`, `--discover`, `--batch`, `--force`, `--dry-run`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runAdd(opts)`
- [ ] 6.3.4 Implement `runAdd(opts *addOptions) error`:
  - **Mode 1 (explicit files):** for each `InputPath`, call `lib.AddResource(opts.Ctx, ...)`; on error, collect into `[]core.InitializeError`; on success, add to successes list. If any failed, return `core.NewPartialSuccessError(...)`.
  - **Mode 2 (`--discover`):** call `lib.DiscoverOrphans(opts.Ctx, ...)`; for each orphan, validate ref via `core.CanInstallResource(orphan.Ref)`; if valid, add; if invalid, skip and collect error. Return `*core.PartialSuccessError` on partial success.
  - **Mode 3 (`--discover --batch --force`):** same as Mode 2 but in a continuous loop; `--force` overrides existing resources; on cancellation, collect partial successes and return wrapped `ctx.Err()`.
- [ ] 6.3.5 Convert `cmd/library/add_test.go` to `iostreams.Test()` + `runF` injection; cover all three modes
- [ ] 6.3.6 Add explicit cancellation test: trigger batch mode with a slow operation, cancel context, verify partial results + non-nil error
- [ ] 6.3.7 Run `mise run check`; confirm all three modes work end-to-end

## 6.4 Migrate `cmd/library/create.go`

- [ ] 6.4.1 In `cmd/library/create.go`, define `createPresetOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Resources []string`, `Description string`, `Force bool`
- [ ] 6.4.2 Declare the `Library` interface with methods called: `CreatePreset(ctx, *CreatePresetRequest) error`
- [ ] 6.4.3 Implement `NewCmdCreatePreset(f *cmdutil.Factory, runF func(*createPresetOptions) error) *cobra.Command`:
  - Add flags: `--resources` (required), `--description`, `--force`
  - **No `--output` flag** (legacy didn't have `--json`)
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runCreatePreset(opts)`
- [ ] 6.4.4 Implement `runCreatePreset(opts *createPresetOptions) error`:
  - Validate `opts.Resources` is non-empty (return `*pflag.InvalidValueError` or similar usage error if empty)
  - For each ref in `opts.Resources`, validate via `core.CanInstallResource(ref)` (fast-fail)
  - Call `lib.CreatePreset(opts.Ctx, &CreatePresetRequest{Name: presetName, Resources: opts.Resources, Description: opts.Description, Force: opts.Force})`
- [ ] 6.4.5 Convert `cmd/library/create_test.go` to `iostreams.Test()` + `runF` injection
- [ ] 6.4.6 Run `mise run check`; confirm `germinator library create preset <name> --resources skill/x,agent/y` works

## 6.5 Delete legacy files

- [ ] 6.5.1 Delete `internal/service/adder.go` if present
- [ ] 6.5.2 Delete `internal/service/creator.go` if present
- [ ] 6.5.3 Confirm `internal/service/` still contains `transformer.go` (used by `adapt`) and `initializer.go` (used by `init`); these are deleted in change-7

## 6.6 Update delta spec

- [ ] 6.6.1 Update `openspec/changes/wire-factory-and-pilots/specs/library/library-json-output/spec.md` (or this change's own delta spec) to add `--output` to `library add`

## 6.7 Verification

- [ ] 6.7.1 Run `mise run lint` — confirm no new violations
- [ ] 6.7.2 Run `mise run test` — confirm all unit tests pass (including cancellation tests)
- [ ] 6.7.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 6.7.4 Run `mise run test:coverage` — confirm coverage for `cmd/library/add.go`, `cmd/library/create.go`, and `internal/core/rules.go` ≥ 70%
- [ ] 6.7.5 Run `mise run test:e2e` — confirm E2E tests for add and create pass (including all three add modes)
- [ ] 6.7.6 Smoke-test:
  - `germinator library add <file> --type skill --name test`
  - `germinator library add <file> --type skill --name test --dry-run`
  - `germinator library add --discover`
  - `germinator library add --discover --batch --force`
  - `germinator library add --discover --batch --force --output json`
  - `germinator library create preset test-preset --resources skill/x,agent/y`
  - `germinator library create preset test-preset --resources skill/x,agent/y --description "Test preset"`
  - `germinator library create preset test-preset --resources ""` (should fail with usage error)
- [ ] 6.7.7 Update `cmd/library/AGENTS.md` with the three modes of `add` documented
- [ ] 6.7.8 Confirm `legacyBridge` still works for non-migrated commands (init, library init, library refresh, library remove, library validate, config, completion, version)

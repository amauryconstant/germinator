# Tasks — Migrate remaining library commands and delete legacy shell

**Slice 7 of 9.** Migrates the four remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) and deletes the entire legacy shell (`internal/service/`, `internal/application/`, `legacyBridge`, `cmd/error_formatter.go`, `cmd/verbose.go`). **Structural turning point** of the migration.

Each task ends with `mise run check` passing.

## Task ordering

Tasks execute in numeric order with the following critical paths:

- **7.0 (methods-on-`*Library`)** blocks **7.1–7.4** because each `runXxx` calls `lib.SomeMethod(...)` on the loaded `*library.Library` returned by `opts.Library()`. The `(*Library) Refresh/RemoveResource/RemovePreset/Validate/Fix` methods and the `library.Request*`/`Result*` types must exist before the cmd file compiles.
- **7.0 (methods-on-`*Library`)** also defines `library.Init(ctx, *InitRequest) error` and the `InitRequest` type for task 7.1.
- **7.1–7.4 (command migrations)** block **7.5** because the legacy shell (`internal/service/`, `internal/application/`, `legacyBridge`, etc.) can only be deleted once all consumers have moved to the new pattern.
- **7.4.7 (formatter helpers move)** blocks **7.5.12 (`cmd/library_formatters.go` deletion)** — the helpers must exist in `internal/output/library.go` before the original file can be removed.
- **7.5.x (legacy shell delete)** blocks **7.6** because the `## Fulfilled` annotations on prior specs reference actual symbols that must be gone first.
- **7.6 + 7.5** block **7.7** (verification) because lint baseline regeneration (7.7.10) depends on the final shape of `cmd/`.

## 7.0 Add methods on `*library.Library` (slice-6 forward path)

- [x] 7.0.1 In `internal/library/requests.go` (new file), define the request/result types (parallels `library.CreatePresetRequest` from `internal/library/creator.go:99`): `InitRequest{Path, Force, DryRun}`, `RefreshRequest{DryRun, Force}`, `RemoveResourceRequest{Ref, Force}` (`Ref` is `"type/name"`), `RemovePresetRequest{Name, Force}`, `ValidateRequest{Fix}`, `FixRequest{}`, `RefreshResult` (already exists — extend with `Unchanged []RefreshUnchanged` per design Decision 7 in this change), `FixResult{RemovedEntries []string; StrippedRefs []string}`, and `RefreshUnchanged{Ref, LastSynced string}`
- [x] 7.0.2 In `internal/library/library.go` (or split across `refresher.go`/`remover.go`/`validator.go`), add 5 methods on `*library.Library`:
  - `(lib *Library) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResult, error)` — assertion: `lib != nil && lib.RootPath != ""` at entry
  - `(lib *Library) RemoveResource(ctx context.Context, req *RemoveResourceRequest) error`
  - `(lib *Library) RemovePreset(ctx context.Context, req *RemovePresetRequest) error`
  - `(lib *Library) Validate(ctx context.Context, req *ValidateRequest) (*ValidateResult, error)` — when `req.Fix`, also call `lib.Fix(ctx, &FixRequest{})`; merge the `*FixResult` into the returned payload when `--fix` and `--output json` are combined
  - `(lib *Library) Fix(ctx context.Context, req *FixRequest) (*FixResult, error)`
  - Each method mirrors the slice-6 `(*Library).CreatePreset` pattern at `internal/library/creator.go:145`
- [x] 7.0.3 In `internal/library/library.go`, add package-level function `library.Init(ctx context.Context, req *InitRequest) error` that maps `InitRequest` fields to `CreateOptions` and calls `library.CreateLibrary`. `init` does not fit as a `*Library` method; the cmd layer uses this function directly without an interface
- [x] 7.0.4 Refactor the existing package-level functions (`library.RefreshLibrary`, `library.RemoveResource`, `library.RemovePreset`, `library.ValidateLibrary`, `library.FixLibrary`) so each preserves its existing public signature (any external caller still compiles) but delegates internally to the new `(*Library) X` method via a thin loader wrapper: `loadWrapper := func(opts ...)(*Library, error){ return LoadLibrary(ctx, opts.Path) }; return wrapper.Refresh(ctx, &req)`. Keep `cmd.libraryInitOptions.Library` field unused (drop entirely per design Decision 6 — `runLibraryInit` calls `library.Init` directly)
- [x] 7.0.5 In `cmdutil/factory.go`, leave the `Factory.Library` field as `func() (*library.Library, error)` — **no signature change**. The four new methods on `*Library` make `*library.Library` directly satisfy the cmd-side interfaces declared in tasks 7.1–7.4
- [x] 7.0.6 Add unit tests in `internal/library/methods_test.go` (table-driven; no mocks) covering each new method (success + each error path)
- [x] 7.0.7 Run `mise run check`

## 7.1 Migrate `cmd/library_init.go`

- [x] 7.1.1 In `cmd/library_init.go`, define `libraryInitOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `Path string`, `Force bool`, `DryRun bool`, `Output string` (no `Library` field — `init` calls `library.CreateLibrary` package function directly per design Decision 6)
- [x] 7.1.2 Implement `NewCmdLibraryInit(f *cmdutil.Factory, runF func(*libraryInitOptions) error) *cobra.Command`:
  - Add flags: `--path`, `--force`, `--dry-run`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runLibraryInit(opts)`
- [x] 7.1.3 Implement `runLibraryInit(opts *libraryInitOptions) error`:
  - Call `library.Init(opts.Ctx, &InitRequest{Path: opts.Path, Force: opts.Force, DryRun: opts.DryRun})` (which maps to `library.CreateLibrary`)
  - Dispatch on `opts.Output` for the result (mirror the resource-add plain/json/table pattern from `cmd/library_add.go`)
- [x] 7.1.4 Convert `cmd/library_init_test.go` to `iostreams.Test()` + `runF` injection
- [x] 7.1.5 Run `mise run check`

## 7.2 Migrate `cmd/library_refresh.go`

- [x] 7.2.1 In `cmd/library_refresh.go`, define `refreshOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `DryRun bool`, `Force bool`, `Output string`
- [x] 7.2.2 Declare the `refresherLibrary` interface with methods called: `Refresh(ctx, *RefreshRequest) (*RefreshResult, error)`; add compile-time check `var _ refresherLibrary = (*library.Library)(nil)` (matches the slice-6 `presetWriter` pattern at `cmd/library_create.go:58`)
- [x] 7.2.3 Implement `NewCmdRefresh(f, runF)` and `runRefresh(opts)`:
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Resolve the lazy library once: `lib, err := opts.Library()` (handle error via `output.FormatError`)
  - Call `lib.Refresh(opts.Ctx, &RefreshRequest{DryRun: opts.DryRun, Force: opts.Force})`
  - Dispatch on `opts.Output` for the result (per-resource status sections: Refreshed, Unchanged, Skipped, Errors)
- [x] 7.2.4 Convert `cmd/library_refresh_test.go` to `iostreams.Test()` + `runF` injection
- [x] 7.2.5 Run `mise run check`
- [x] 7.2.6 In `internal/library/refresher.go`, add `Unchanged []RefreshUnchanged` field to `RefreshResult`; populate when a resource is scanned but produces no change
- [x] 7.2.7 In `cmd/library_refresh.go`, render the new `Unchanged:` section in plain output

## 7.3 Migrate `cmd/library_remove.go`

- [x] 7.3.1 In `cmd/library_remove.go`, define `removeOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Ref string` (positional `<ref>` arg, e.g. `"skill/commit"`), `PresetName string` (positional `<name>` arg), `Force bool`, `Output string`. **No `ResourceType`/`ResourceName` fields** — the legacy positional `<ref>` argument is preserved (no breaking CLI change).
- [x] 7.3.2 Declare the `removerLibrary` interface with methods: `RemoveResource(ctx, *RemoveResourceRequest) error`, `RemovePreset(ctx, *RemovePresetRequest) error`; add compile-time check `var _ removerLibrary = (*library.Library)(nil)`
- [x] 7.3.3 Implement `NewCmdRemove(f *cmdutil.Factory, runF func(*removeOptions) error) *cobra.Command`:
  - Add sub-commands: `library remove resource <ref>` (positional `<ref>`, parsed via `library.ParseRef`) and `library remove preset <name>` (positional `<name>`)
  - Add `--force` flag on the parent (inherited by both sub-commands)
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` on the parent
  - Populate `opts` in each sub-command's `RunE`: `opts.Ref = args[0]` for resource, `opts.PresetName = args[0]` for preset
  - Call `runF(opts)` if non-nil, else `runRemove(opts)`
- [x] 7.3.4 Implement `runRemove(opts *removeOptions) error`:
  - Resolve the lazy library once: `lib, err := opts.Library()` (handle error via `output.FormatError`)
  - If `opts.PresetName != ""`: call `lib.RemovePreset(opts.Ctx, &RemovePresetRequest{Name: opts.PresetName, Force: opts.Force})`
  - Else: call `lib.RemoveResource(opts.Ctx, &RemoveResourceRequest{Ref: opts.Ref, Force: opts.Force})` (the method parses `Ref` via `library.ParseRef`)
  - Dispatch on `opts.Output` for the result
- [x] 7.3.5 Convert `cmd/library_remove_test.go` (or add tests) to `iostreams.Test()` + `runF` injection; cover both resource and preset removal
- [x] 7.3.6 Run `mise run check`

## 7.4 Migrate `cmd/library_validate.go`

- [x] 7.4.1 In `cmd/library_validate.go`, define `libraryValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Fix bool`, `Output string`
- [x] 7.4.2 Declare the `validatorLibrary` interface with methods called: `Validate(ctx, *ValidateRequest) (*ValidateResult, error)`; add compile-time check `var _ validatorLibrary = (*library.Library)(nil)`
- [x] 7.4.3 Implement `NewCmdLibraryValidate(f, runF)` and `runLibraryValidate(opts)`:
  - Add `--fix` flag
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Resolve the lazy library once: `lib, err := opts.Library()`
  - Call `lib.Validate(opts.Ctx, &ValidateRequest{Fix: opts.Fix})` (when `req.Fix`, the method internally calls `lib.Fix(ctx, &FixRequest{})` per task 7.0.2)
  - If validation errors exist, render each via `output.FormatError`
  - Dispatch on `opts.Output` for the result
- [x] 7.4.4 Convert `cmd/library_validate_test.go` to `iostreams.Test()` + `runF` injection; cover both with and without `--fix`
- [x] 7.4.5 Run `mise run check`
- [x] 7.4.6 In `cmd/library_validate.go`, when `--fix` is set with `--output json`, include `RemovedEntries []string` and `StrippedRefs []string` in the JSON payload (sourced from `*FixResult` returned by `lib.Fix`)
- [x] 7.4.7 Move formatter helpers from `cmd/library_formatters.go` to `internal/output/library.go`; update imports in the 4 migrated command files

## 7.5 Delete legacy shell

- [x] 7.5.1 Run `rg "internal/service" .` to verify zero remaining references; delete `internal/service/` directory tree (10+ files including `transformer.go`, `initializer.go`, all `*_test.go`)
- [x] 7.5.2 Run `rg "internal/application" .` to verify zero remaining references; delete `internal/application/` directory tree (3 files: `interfaces.go`, `requests.go`, `results.go`)
- [x] 7.5.3 Run `rg "application\.(Transformer|Validator|Canonicalizer|Initiator|.*Request|.*Result)" .` to verify zero remaining references to application-package symbols; fix any remaining call sites
- [x] 7.5.4 Delete `cmd/legacy_bridge.go` (file) entirely
- [x] 7.5.5 Drop the `bridge *LegacyBridge` parameter from every cmd constructor (`NewRootCommand`, `NewVersionCommand`, `NewLibraryCommand`, `NewCompletionCommand`, `NewConfigCommand`, `NewConfigInitCommand`, `NewConfigValidateCommand`) and the inner call sites in `cmd/root.go`. Stage C3 will update `main.go` to drop its `cmd.LegacyBridge{}` literal construction and to pass `cmd.NewRootCommand(f)` etc. without `bridge`.
- [x] 7.5.6 All `cmd/` constructor signatures now take only `f *cmdutil.Factory` (and the per-command `runF`/`libraryPath` typed arg); the `bridge *LegacyBridge` parameter is gone from `NewRootCommand`, `NewLibraryCommand`, `NewVersionCommand`, `NewCompletionCommand`, `NewConfigCommand`, `NewConfigInitCommand`, `NewConfigValidateCommand`. `main.go` call sites land in Stage C3 alongside the `cmd.LegacyBridge{}` removal.
- [x] 7.5.7 Delete `cmd/error_formatter.go` (no consumer after `legacyBridge` removed)
- [x] 7.5.8 Delete `cmd/verbose.go` (no consumer after `legacyBridge` removed)
- [x] 7.5.9 Delete `cmd/legacy_test_helpers_test.go` (no consumers after `legacyBridge` removed)
- [x] 7.5.10 Update `cmd/cmd_test.go` to remove `newTestBridge()` calls; convert to `runF` + `iostreams.Test()` pattern
- [x] 7.5.11 Run `rg "ServiceContainer|CommandConfig|ErrorFormatter|Verbosity|LegacyBridge|legacyBridge" cmd/ main.go` to verify zero remaining references to legacy symbols
- [x] 7.5.12 Delete `cmd/library_formatters.go` (helpers moved to `internal/output/library.go` in task 7.4.7)

## 7.6 Update delta specs (final fulfillment)

- [x] 7.6.1 Update `openspec/changes/scaffold-cli-foundation/specs/application/dependency-injection/spec.md`: mark `ServiceContainer` removal as **fulfilled**
- [x] 7.6.2 Update `openspec/changes/scaffold-cli-foundation/specs/cli/exit-codes/spec.md`: mark `CategorizeError` enum removal as **fulfilled**
- [x] 7.6.3 Update `openspec/changes/scaffold-cli-foundation/specs/cli/framework/spec.md`: confirm `CommandConfig` removal (was fulfilled in change-2)
- [x] 7.6.4 Update `openspec/changes/scaffold-cli-foundation/specs/cli/verbose-output/spec.md`: mark `VerbosePrint` removal as **fulfilled**
- [x] 7.6.5 Update `openspec/changes/scaffold-cli-foundation/specs/cli/error-formatting/spec.md`: mark `ErrorFormatter` removal as **fulfilled**
- [x] 7.6.6 Update `openspec/changes/scaffold-cli-foundation/specs/application/service-contracts/spec.md`: mark `Transformer`/`Validator`/`Canonicalizer`/`Initializer` removal as **fulfilled**

## 7.7 Verification

- [x] 7.7.1 Run `mise run lint` — confirm no new violations
- [x] 7.7.2 Run `mise run test` — confirm all unit tests pass (no mocks needed)
- [x] 7.7.3 Run `mise run build` — confirm `bin/germinator` builds without `internal/service/` or `internal/application/`
- [x] 7.7.4 Run `mise run test:coverage` — confirm coverage maintained for `cmd/library_*.go` ≥ 70%
- [x] 7.7.5 Run `mise run test:full` (unit + e2e)
- [x] 7.7.6 Smoke-test every library subcommand with exit-code assertions:
  - `germinator library init --path /tmp/lib --dry-run` → expect exit 0
  - `germinator library resources` → expect exit 0
  - `germinator library resources --output json` → expect exit 0
  - `germinator library presets` → expect exit 0
  - `germinator library show <ref>` → expect exit 0
  - `germinator library add <file> --type skill --name test` → expect exit 0
  - `germinator library create preset <name> --resources skill/x` → expect exit 0
  - `germinator library refresh` → expect exit 0 (or 1 if conflicts)
  - `germinator library refresh --dry-run` → expect exit 0
  - `germinator library remove resource skill/test --force` → expect exit 0
  - `germinator library remove preset <name> --force` → expect exit 0
  - `germinator library validate` → expect exit 0 (or 5 if issues found)
  - `germinator library validate --fix` → expect exit 0
- [x] 7.7.7 Verify `cmd/`, `internal/library/`, `internal/core/`, `internal/cmdutil/`, `internal/iostreams/`, `internal/output/` have no imports of `internal/service/` or `internal/application/`
- [x] 7.7.8 Confirm non-library commands (`adapt`, `validate`, `canonicalize`, `init` (resource installation — regression only, not migrated), `config`, `completion`, `version`) still work
- [x] 7.7.9 Regenerate `cmd/testdata/lint_baseline.txt` (run `mise run lint > cmd/testdata/lint_baseline.txt 2>&1`) and commit the updated baseline
- [x] 7.7.10 Run `mise run test:release` to confirm GoReleaser build still succeeds without `internal/service/`/`internal/application/`

> **AGENTS.md documentation updates** for the four migrated commands are handled by the `osx-maintain-ai-docs` skill (run after archive), not as a numbered task in this list — per the root `AGENTS.md` convention.

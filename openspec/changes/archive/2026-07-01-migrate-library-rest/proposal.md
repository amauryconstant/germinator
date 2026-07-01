# Migrate remaining library commands and delete legacy shell

## Why

This is the **structural turning point** of the migration. Once all library commands are migrated (no consumer of `internal/service/` or `internal/application/` remains), the legacy shell can be deleted in one change: `internal/service/` and `internal/application/` (the eager-wiring + service-interface layers), the `legacyBridge` shim in `main.go`, and the legacy `cmd/error_formatter.go` + `cmd/verbose.go` (kept alive only to support non-migrated commands via `legacyBridge`).

## What Changes

### Migrate remaining library commands

- **MIGRATE** `cmd/library_init.go`:
  - Declare `libraryInitOptions`: `IO`, `Ctx`, `Path string`, `Force bool`, `DryRun bool`, `Output string`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - `runLibraryInit` calls `library.CreateLibrary(...)` (package function) directly — `init` creates a fresh library, so no `*Library` receiver is involved; no adapter shim or interface is needed for this command
  - Implement `NewCmdLibraryInit(f, runF)` and `runLibraryInit(opts)`
- **MIGRATE** `cmd/library_refresh.go`:
  - Declare `refreshOptions`: `IO`, `Library func() (*library.Library, error)`, `Ctx`, `DryRun bool`, `Force bool`, `Output string`
  - Define `refresherLibrary` interface with `Refresh(ctx, *RefreshRequest) (*RefreshResult, error)`; satisfied directly by `*library.Library` (new method on `*Library`)
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdRefresh(f, runF)` and `runRefresh(opts)`
- **MIGRATE** `cmd/library_remove.go`:
  - Declare `removeOptions`: `IO`, `Library func() (*library.Library, error)`, `Ctx`, `Ref string`, `PresetName string`, `Force bool`, `Output string`
  - Sub-command dispatch (`resource <ref>` keeps the legacy positional `<ref>` argument; `preset <name>` likewise)
  - Define `removerLibrary` interface with `RemoveResource(ctx, *RemoveResourceRequest) error` and `RemovePreset(ctx, *RemovePresetRequest) error`; satisfied directly by `*library.Library`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdRemove(f, runF)` and `runRemove(opts)`
  - **No breaking CLI change:** `germinator library remove resource <ref>` and `germinator library remove preset <name>` keep their positional args; only `--force` is added
- **MIGRATE** `cmd/library_validate.go`:
  - Declare `libraryValidateOptions`: `IO`, `Library func() (*library.Library, error)`, `Ctx`, `Fix bool`, `Output string`
  - Define `validatorLibrary` interface with `Validate(ctx, *ValidateRequest) (*ValidateResult, error)`; satisfied directly by `*library.Library`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdLibraryValidate(f, runF)` and `runLibraryValidate(opts)` with `--fix` support

### Add methods on `*library.Library` (follows slice-6 forward path)

- **ADD** the following methods to `*library.Library` (mirroring the slice-6 `(*Library).CreatePreset` precedent at `internal/library/creator.go:145`):
  - `(lib *Library) Refresh(ctx, *RefreshRequest) (*RefreshResult, error)` — uses `lib.RootPath`, delegates to a new internal `refreshLibraryFromLib` helper; existing package-level `library.RefreshLibrary(opts RefreshOptions)` is refactored to delegate to the method form via a thin path-loading wrapper (preserves any external callers)
  - `(lib *Library) RemoveResource(ctx, *RemoveResourceRequest) error`
  - `(lib *Library) RemovePreset(ctx, *RemovePresetRequest) error`
  - `(lib *Library) Validate(ctx, *ValidateRequest) (*ValidateResult, error)`
  - `(lib *Library) Fix(ctx, *FixRequest) (*FixResult, error)` — required because `Validate` with `req.Fix` needs to mutate; tests the design Decision 3a fix report
- **ADD** a thin `library.Init(ctx, *InitRequest) error` package function that maps to `library.CreateLibrary(CreateOptions{...})`. `init` does not fit as a `*Library` method because there is no pre-existing `*Library`; the cmd layer calls the package function directly without an interface
- **ADD** request/result types in `internal/library/requests.go` (new file): `InitRequest`, `RefreshRequest`, `RemoveResourceRequest`, `RemovePresetRequest`, `ValidateRequest`, `FixRequest`, plus result types where the operation returns data
- **No changes** to `cmdutil.Factory.Library`: the existing `func() (*library.Library, error)` field already returns a `*library.Library`, which now satisfies all four `*Library` interfaces declared in the cmd files

> **Why methods-on-`*Library` instead of `*LibraryService`:** slice-6 explicitly noted in `cmd/canonical-examples/AGENTS.md:160` that "a future slice that converts the package functions to methods on `*library.Library` will allow the compile-time check against the concrete type instead of the adapter." This slice delivers that path and removes the parallel-type anti-pattern that slice-7 is otherwise reintroducing in mirror form.

### Delete legacy shell

- **DELETE** `internal/service/` entirely (after confirming no remaining references)
- **DELETE** `internal/application/` entirely
- **DELETE** `cmd/legacy_bridge.go` and the `legacyBridge` shim in `main.go`
- **DELETE** `cmd/error_formatter.go` (no consumer after `legacyBridge` removed)
- **DELETE** `cmd/verbose.go` (no consumer after `legacyBridge` removed)

### Update tests

- **CONVERT** all remaining library command tests to `iostreams.Test()` + `runF` injection
- **DELETE** all `internal/service/*_mock_test.go` files (mocks no longer needed)

## Capabilities

### Modified (final)

- **`library/library-refresh`** (delta) — `--output` flag is added to `library refresh`; `--dry-run` and `--force` are preserved; new `Unchanged` section in plain output; command operates on `(*library.Library).Refresh` (new method on `*Library`)
- **`library/library-scaffolding`** (delta) — `--output` flag is added to `library init`; `--path`, `--force`, `--dry-run` are preserved; command calls `library.CreateLibrary` package function directly (no `*Library` receiver is involved in `init`)
- **`library/library-remove-resource`** (delta) — `--output` flag is added to `library remove resource`; positional `<ref>` argument is preserved (no breaking CLI change); `--force` flag is added; command operates on `(*library.Library).RemoveResource` (new method)
- **`library/library-remove-preset`** (delta) — `--output` flag is added to `library remove preset`; positional `<name>` argument is preserved; `--force` flag is added; command operates on `(*library.Library).RemovePreset` (new method)
- **`library/library-validation`** (delta) — `--output` flag is added to `library validate`; `--fix` flag is preserved; `--fix` + `--output json` emits a machine-readable fix report; command operates on `(*library.Library).Validate` and `(*library.Library).Fix` (new methods)

### Fulfilled (these were delta specs from earlier changes)

- **`application/dependency-injection`** — `ServiceContainer` and `internal/application/` are now **fully removed**
- **`application/service-contracts`** — `Transformer`, `Validator`, `Canonicalizer`, `Initializer` interfaces and their `*Request`/`*Result` types are now **fully removed**
- **`cli/exit-codes`** — `CategorizeError` and the `Category*` enum are now **fully removed**
- **`cli/verbose-output`** — `Verbosity` type and `VerbosePrint`/`VeryVerbosePrint` helpers are now **fully removed**
- **`cli/error-formatting`** — `ErrorFormatter` struct is now **fully removed**

### Implicit (covered by other modified capabilities)

The new `(lib *Library) Refresh/RemoveResource/RemovePreset/Validate/Fix` methods mirror the existing `(*Library).CreatePreset` precedent (slice-6, `internal/library/creator.go:145`) and are an additive API. They do not change existing public behavior; new request types (`library.RefreshRequest`, `library.RemoveResourceRequest`, `library.RemovePresetRequest`, `library.ValidateRequest`, `library.InitRequest`) are thin parameter structs, parallel to the existing `library.CreatePresetRequest`. The public `*library.Library` data shape is unchanged. **No new top-level capability is needed** beyond the modified `library-library-*` deltas listed above.

## Out of scope (deferred)

- Migrating `config init`, `config validate` — change-8
- Migrating `completion`, `version`, deleting `internal/models/`, finalizing `AGENTS.md` + CHANGELOG — change-9

## Impact

### Affected code

- **Rewritten (4 files):** `cmd/library_init.go`, `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_validate.go`
- **Modified (new file):** `internal/library/requests.go` (defines `*Request` / `*Result` types for the new `*Library` methods)
- **Modified (1+ file):** `internal/library/library.go` (or split across `refresher.go`/`remover.go`/`validator.go`) — add five methods on `*Library`: `Refresh`, `RemoveResource`, `RemovePreset`, `Validate`, `Fix`. Each method mirrors the slice-6 `(*Library).CreatePreset` pattern at `internal/library/creator.go:145`
- **Modified (1 file):** `main.go` (remove `legacyBridge` construction + drop `bridge` arg from `NewRootCommand`)
- **Modified (3+ files):** `cmd/root.go`, `cmd/library.go`, and any other constructor still accepting `bridge *LegacyBridge` — drop the `bridge` parameter
- **Deleted (entire directory):** `internal/service/` (~10 files + tests)
- **Deleted (entire directory):** `internal/application/` (3 files)
- **Deleted (3 files):** `cmd/legacy_bridge.go`, `cmd/error_formatter.go`, `cmd/verbose.go`
- **Deleted (1 file):** `cmd/legacy_test_helpers_test.go` (no consumers after `legacyBridge` removed)
- **Modified (1 file):** `cmd/cmd_test.go` (remove `newTestBridge()` calls; convert to `runF` + `iostreams.Test()`)
- **Modified (4 files):** `cmd/library_{init,refresh,remove,validate}_test.go` (converted to new pattern)

### Affected systems

- **Library commands:** `--output` flag is added to `library init`, `library refresh`, `library remove`, `library validate` (additive; default `plain` preserves current output)
- **CLI surface:** `germinator library remove resource <ref>` and `germinator library remove preset <name>` keep their positional args; only `--force` is added. **No breaking CLI change.**
- **Build:** `mise run build` succeeds without `internal/service/` or `internal/application/`

## Risks

- **Mass deletion is risky** — deleting 2 entire directories + 3 files could miss a reference. **Mitigation:** `rg "internal/service" .` and `rg "internal/application" .` are run in tasks 7.5.1 and 7.5.2; any remaining reference is fixed in the same change.
- **Mocks deletion breaks tests that depended on them** — the `internal/service/*_mock_test.go` mocks are referenced by `cmd/cmd_test.go` and possibly other test files. **Mitigation:** task 7.5.5 confirms no remaining references; affected tests are converted to use `iostreams.Test()` + `runF` injection.
- **`legacyBridge` deletion is the riskiest single change in this slice** — `main.go` is the only composition root; removing the bridge means non-migrated commands will fail to compile if any still depend on it. **Mitigation:** by this change, ALL commands have been migrated (changes 2-6); `legacyBridge` has no consumers; `mise run check` is the gate.
- **`cmd/error_formatter.go` and `cmd/verbose.go` may have been referenced from tests** — task 7.5.6 verifies no remaining references.
- **Method-on-`*Library` requires `lib.RootPath` to be set** — the loaded `*Library` returned by `cmdutil.Factory.Library` already carries `RootPath` (populated by `LoadLibrary`). **Mitigation:** tasks in 7.0 explicitly assert `lib != nil && lib.RootPath != ""` at the entry of each new method.

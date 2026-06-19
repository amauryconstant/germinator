# Migrate remaining library commands and delete legacy shell

## Why

This is the **structural turning point** of the migration. Once all library commands are migrated (no consumer of `internal/service/` or `internal/application/` remains), the legacy shell can be deleted in one change: `internal/service/` and `internal/application/` (the eager-wiring + service-interface layers), the `legacyBridge` shim in `main.go`, and the legacy `cmd/error_formatter.go` + `cmd/verbose.go` (kept alive only to support non-migrated commands via `legacyBridge`).

## What Changes

### Migrate remaining library commands

- **MIGRATE** `cmd/library/init.go`:
  - Declare `libraryInitOptions`: `IO`, `Library`, `Ctx`, `Path string`, `Force bool`, `DryRun bool`, `Output string`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdLibraryInit(f, runF)` and `runLibraryInit(opts)`
- **MIGRATE** `cmd/library/refresh.go`:
  - Declare `refreshOptions`: `IO`, `Library`, `Ctx`, `DryRun bool`, `Force bool`, `Output string`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdRefresh(f, runF)` and `runRefresh(opts)`
- **MIGRATE** `cmd/library/remove.go`:
  - Declare `removeOptions`: `IO`, `Library`, `Ctx`, `ResourceType string`, `ResourceName string`, `PresetName string`, `Force bool`, `Output string`
  - Sub-command dispatch (resource vs preset)
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdRemove(f, runF)` and `runRemove(opts)`
- **MIGRATE** `cmd/library/validate.go`:
  - Declare `libraryValidateOptions`: `IO`, `Library`, `Ctx`, `Fix bool`, `Output string`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdLibraryValidate(f, runF)` and `runLibraryValidate(opts)` with `--fix` support

### Delete legacy shell

- **DELETE** `internal/service/` entirely (after confirming no remaining references)
- **DELETE** `internal/application/` entirely
- **DELETE** `legacyBridge` shim from `main.go`
- **DELETE** `cmd/error_formatter.go` (no consumer after `legacyBridge` removed)
- **DELETE** `cmd/verbose.go` (no consumer after `legacyBridge` removed)

### Update tests

- **CONVERT** all remaining library command tests to `iostreams.Test()` + `runF` injection
- **DELETE** all `internal/service/*_mock_test.go` files (mocks no longer needed)

## Capabilities

### Modified (final)

- **`library/library-refresh`** (delta) — `--output` flag is added to `library refresh`; `--dry-run` and `--force` are preserved
- **`library/library-remove-resource`** (delta) — `--output` flag is added to `library remove resource`
- **`library/library-remove-preset`** (delta) — `--output` flag is added to `library remove preset`
- **`library/library-validation`** (delta) — `--output` flag is added to `library validate`; `--fix` flag is preserved

### Fulfilled (these were delta specs from earlier changes)

- **`application/dependency-injection`** — `ServiceContainer` and `internal/application/` are now **fully removed**
- **`cli/exit-codes`** — `CategorizeError` and the `Category*` enum are now **fully removed**
- **`cli/verbose-output`** — `Verbosity` type and `VerbosePrint`/`VeryVerbosePrint` helpers are now **fully removed**
- **`cli/error-formatting`** — `ErrorFormatter` struct is now **fully removed**

## Out of scope (deferred)

- Migrating `config init`, `config validate` — change-8
- Migrating `completion`, `version`, deleting `internal/models/`, finalizing `AGENTS.md` + CHANGELOG — change-9

## Impact

### Affected code

- **Rewritten (4 files):** `cmd/library/init.go`, `cmd/library/refresh.go`, `cmd/library/remove.go`, `cmd/library/validate.go`
- **Modified (1 file):** `main.go` (remove `legacyBridge`)
- **Deleted (entire directory):** `internal/service/` (~10 files + tests)
- **Deleted (entire directory):** `internal/application/` (3 files)
- **Deleted (2 files):** `cmd/error_formatter.go`, `cmd/verbose.go`
- **Modified (4 files):** `cmd/library/{init,refresh,remove,validate}_test.go` (converted to new pattern)
- **Deleted (4 files):** `internal/service/*_mock_test.go` (mocks no longer needed)

### Affected systems

- **Library commands:** `--output` flag is added to `library init`, `library refresh`, `library remove`, `library validate` (additive; default `plain` preserves current output)
- **Build:** `mise run build` succeeds without `internal/service/` or `internal/application/`

## Risks

- **Mass deletion is risky** — deleting 4 entire directories + 2 files could miss a reference. **Mitigation:** `rg "internal/service" .` and `rg "internal/application" .` are run in tasks 7.5.1 and 7.5.2; any remaining reference is fixed in the same change.
- **Mocks deletion breaks tests that depended on them** — the `internal/service/*_mock_test.go` mocks are referenced by `cmd/cmd_test.go` and possibly other test files. **Mitigation:** task 7.5.5 confirms no remaining references; affected tests are converted in tasks 7.4.x to use `iostreams.Test()` + `runF` injection.
- **`legacyBridge` deletion is the riskiest single change in this slice** — `main.go` is the only composition root; removing the bridge means non-migrated commands will fail to compile if any still depend on it. **Mitigation:** by this change, ALL commands have been migrated (changes 2-6); `legacyBridge` has no consumers; `mise run check` is the gate.
- **`cmd/error_formatter.go` and `cmd/verbose.go` may have been referenced from tests** — task 7.5.6 verifies no remaining references.

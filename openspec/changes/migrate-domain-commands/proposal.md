# Migrate validate and canonicalize

## Why

After change-2 (`wire-factory-and-pilots`) proves the new pattern works for `adapt` and `library resources`, the next step is to migrate the two remaining core domain commands (`validate` and `canonicalize`). These commands follow the same template as the pilots but don't depend on the library. Migrating them completes the core domain command set and demonstrates the pattern extends beyond the pilots.

## What Changes

### Migrate `cmd/validate.go`

- **MIGRATE** `cmd/validate.go`:
  - Declare `validateOptions`: `IO *iostreams.IOStreams`, `Validator func() (Validator, error)`, `Ctx context.Context`, `InputPath string`, `Platform string`
  - Declare the `Validator` interface in the same file (one method: `Validate(ctx, *ValidateRequest) (*core.ValidateResult, error)`)
  - Implement `NewCmdValidate(f *cmdutil.Factory, runF func(*validateOptions) error) *cobra.Command`
  - Implement `runValidate(opts *validateOptions) error`: validate platform, call `validator.Validate`, iterate `result.Errors` and call `output.FormatError` for each; return the first typed error

### Migrate `cmd/canonicalize.go`

- **MIGRATE** `cmd/canonicalize.go`:
  - Declare `canonicalizeOptions`: `IO`, `Canonicalizer func() (Canonicalizer, error)`, `Ctx`, `InputPath`, `OutputPath`, `Platform`, `DocType`
  - Declare the `Canonicalizer` interface in the same file (one method: `Canonicalize(ctx, *CanonicalizeRequest) (*core.CanonicalizeResult, error)`)
  - Implement `NewCmdCanonicalize(f, runF)` and `runCanonicalize(opts)`

### Delete legacy service implementations

- **DELETE** `internal/service/validator.go` (logic moves to the migrated `cmd/validate.go` or stays as a small file)
- **DELETE** `internal/service/canonicalizer.go`
- **DELETE** corresponding `internal/service/{validator,canonicalizer}_test.go` (tests move to the new locations)

### Update `legacyBridge` in main.go

- The `legacyBridge` already calls `f.Validator()` and `f.Canonicalizer()` (which now return the migrated implementations); no main.go changes needed beyond confirming the wiring works.

## Capabilities

### Modified

- None — the migration follows the `application/command-options-pattern` and `cli-factory` capabilities established in change-1.

## Out of scope (deferred)

- Migrating any library command — changes 4, 6, 7
- Migrating `init` — change-5
- Migrating config / completion / version — changes 8, 9
- Deleting `internal/service/` entirely (other service files remain for `Transformer` and `Initializer`, used by `init` and `library add`) — change-7
- Deleting `internal/application/` — change-7

## Impact

### Affected code

- **Rewritten (1 file):** `cmd/validate.go`
- **Rewritten (1 file):** `cmd/canonicalize.go`
- **Deleted (2 files):** `internal/service/validator.go`, `internal/service/canonicalizer.go`
- **Deleted (2 files):** `internal/service/validator_test.go`, `internal/service/canonicalizer_test.go`
- **Modified (1 file):** `cmd/validate_test.go` (converted to `iostreams.Test()` + `runF` injection)
- **Modified (1 file):** `cmd/canonicalize_test.go` (converted similarly)

### Affected systems

- **CLI behavior:** no externally observable changes (validate/canonicalize output remains identical)
- **`legacyBridge`:** unchanged (still references `*ServiceContainer` for `Transformer` and `Initializer`)

## Risks

- **Validator returns multiple errors** — the legacy validator returned `*ValidationResult` with `Errors []ValidationError`; the new `runValidate` must iterate and render each. **Mitigation:** task 3.1.4 explicitly handles this; tests in `validate_test.go` cover both single-error and multi-error cases.
- **Test surface rewrite** — `validator_test.go` (555 lines) and `canonicalizer_test.go` (241 lines) are deleted; their assertions move to `cmd/validate_test.go` and `cmd/canonicalize_test.go`. **Mitigation:** tests are converted in-place; golden file tests are preserved by moving them to the new locations (or kept in `internal/service/` if their imports are updated).
- **`internal/service/` is half-deleted** — after this change, only `transformer.go` and `initializer.go` remain in `internal/service/`. **Mitigation:** these are the last two services needed (used by `adapt` and `library add`/`init`); the directory is fully deleted in change-7.

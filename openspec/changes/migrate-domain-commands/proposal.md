# Migrate validate and canonicalize

## Why

After change-2 (`wire-factory-and-pilots` — see `openspec/changes/archive/2026-06-26-wire-factory-and-pilots/proposal.md`) proves the new pattern works for `adapt` and `library resources`, the next step is to migrate the two remaining core domain commands (`validate` and `canonicalize`). These commands follow the same template as the pilots but don't depend on the library. Migrating them completes the core domain command set and demonstrates the pattern extends beyond the pilots.

## What Changes

### Migrate `cmd/validate.go`

- **MIGRATE** `cmd/validate.go`:
  - Declare `validateOptions`: `IO *iostreams.IOStreams`, `Validator func() (Validator, error)`, `Ctx context.Context`, `InputPath string`, `Platform string`
  - Declare the `Validator` interface in the same file (one method: `Validate(ctx, *ValidateRequest) (*core.ValidateResult, error)`)
  - Declare `type ValidateRequest = application.ValidateRequest` (alias to avoid import-name collision with the local interface's method parameter; matches `cmd/adapt.go:41`)
  - Implement `validateValidator(f *cmdutil.Factory) func() (Validator, error)` wrapper that calls `f.Validator()` and type-asserts to the local `Validator` interface (same pattern as `cmd/adapt.go:91-102`)
  - Implement `NewCmdValidate(f *cmdutil.Factory, runF func(*validateOptions) error) *cobra.Command`; in `RunE`, populate `opts.Validator = validateValidator(f)`
  - Implement `runValidate(opts *validateOptions) error`: validate platform via `core.ValidatePlatform`, call `validator.Validate`, iterate `result.Errors` and call `output.FormatError(opts.IO, err)` for each; return the first typed error so `cmdutil.ExitCodeFor` maps it to `1`

### Migrate `cmd/canonicalize.go`

- **MIGRATE** `cmd/canonicalize.go`:
  - Declare `canonicalizeOptions`: `IO`, `Canonicalizer func() (Canonicalizer, error)`, `Ctx`, `InputPath`, `OutputPath`, `Platform`, `DocType`
  - Declare the `Canonicalizer` interface in the same file (one method: `Canonicalize(ctx, *CanonicalizeRequest) (*core.CanonicalizeResult, error)`)
  - Declare `type CanonicalizeRequest = application.CanonicalizeRequest` (alias to avoid import-name collision; matches `cmd/adapt.go:41`)
  - Implement `canonicalizeCanonicalizer(f *cmdutil.Factory) func() (Canonicalizer, error)` wrapper that calls `f.Canonicalizer()` and type-asserts to the local `Canonicalizer` interface (same pattern as `cmd/adapt.go:91-102`)
  - Implement `NewCmdCanonicalize(f, runF)` and `runCanonicalize(opts)`; in `RunE`, populate `opts.Canonicalizer = canonicalizeCanonicalizer(f)`

### Delete legacy service implementations

- **DELETE** `internal/service/validator.go` (logic moves to the migrated `cmd/validate.go` or stays as a small file)
- **DELETE** `internal/service/canonicalizer.go`
- **DELETE** corresponding `internal/service/{validator,canonicalizer}_test.go` (tests move to the new locations)
- **VERIFY** post-deletion: `find internal/service -type f` returns only `transformer.go`, `initializer.go`, and their `_test.go` files.

### Verification (no code change)

The `legacyBridge` continues to call `f.Validator()` and `f.Canonicalizer()`, which now resolve to the migrated implementations. Verification is covered by task 3.4.6 (legacyBridge smoke-test).

## Capabilities

### Modified

- **`cli-framework`** — The `validate` and `canonicalize` commands adopt the `NewCmdXxx(f, runF)` + `runXxx(opts)` pattern with local `Validator` / `Canonicalizer` interfaces and lazy `func() (Validator, error)` fields. They no longer reference `*CommandConfig`.

> **Note on delta-spec layout:** this change stores the delta in a flat folder (`specs/framework/spec.md`) because the `openspec` CLI requires flat layout for delta discovery. When synced to `openspec/specs/`, this delta lands under `cli-framework/` per `openspec/specs/AGENTS.md`. This matches the pattern from change-1 (`archive/2026-06-24-scaffold-cli-foundation/proposal.md:69-83`).

## Out of scope (deferred)

- Migrating any library command — changes 4, 6, 7
- Migrating `init` — change-5
- Migrating config / completion / version — changes 8, 9
- Deleting `internal/service/` entirely (other service files remain for `Transformer` and `Initializer`, used by `init` and `library add`) — change-7
  - Consumers of `legacyBridge` after this change: `init`, `library init`, `library add`, `library create`, `library refresh`, `library remove`, `library validate`, `config`, `completion`, `version` (full smoke list in task 3.4.6)
- Deleting `internal/application/` — change-7

## Impact

### Affected code

- **Rewritten (1 file):** `cmd/validate.go`
- **Rewritten (1 file):** `cmd/canonicalize.go`
- **Deleted (2 files):** `internal/service/validator.go`, `internal/service/canonicalizer.go`
- **Deleted (2 files):** `internal/service/validator_test.go`, `internal/service/canonicalizer_test.go`
- **Modified (1 file):** `cmd/validate_test.go` (converted from `internal/service/validator_test.go` — 555 lines — to `iostreams.Test()` + `runF` injection)
- **Modified (1 file):** `cmd/canonicalize_test.go` (converted from `internal/service/canonicalizer_test.go` — 241 lines — similarly)

### Affected systems

- **CLI behavior:** no externally observable changes (validate/canonicalize output remains identical)
- **`legacyBridge`:** unchanged (still references `*ServiceContainer` for `Transformer` and `Initializer`)

## Risks

- **Validator returns multiple errors** — the legacy validator returned `*ValidationResult` with `Errors []ValidationError`; the new `runValidate` must iterate and render each. **Mitigation:** task 3.1.4 explicitly handles this; tests in `validate_test.go` cover both single-error and multi-error cases.
- **Test surface rewrite** — `validator_test.go` (555 lines) and `canonicalizer_test.go` (241 lines) are deleted; their assertions move to `cmd/validate_test.go` and `cmd/canonicalize_test.go`. **Mitigation:** tests are converted in-place; golden file tests are preserved by moving them to `cmd/canonicalize_golden_test.go` (the `internal/service/` location is rejected per design Decision 4 because the implementation file is deleted in this change).
- **`internal/service/` is half-deleted** — after this change, only `transformer.go` and `initializer.go` remain in `internal/service/`. **Mitigation:** these are the last two services needed (used by `adapt` and `library add`/`init`); the directory is fully deleted in change-7.

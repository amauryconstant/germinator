# Tasks — Migrate validate and canonicalize

**Slice 3 of 9.** Migrates `cmd/validate.go` and `cmd/canonicalize.go` to the new pattern. Deletes `internal/service/validator.go` and `internal/service/canonicalizer.go`. The `legacyBridge` continues to support the remaining non-migrated commands.

Each task ends with `mise run check` passing.

## 3.1 Migrate `cmd/validate.go`

- [ ] 3.1.1 In `cmd/validate.go`, define `validateOptions` struct with fields: `IO *iostreams.IOStreams`, `Validator func() (Validator, error)`, `Ctx context.Context`, `InputPath string`, `Platform string`
- [ ] 3.1.2 Declare the `Validator` interface in `cmd/validate.go` (one method: `Validate(ctx, *ValidateRequest) (*core.ValidateResult, error)`); declare `type ValidateRequest = application.ValidateRequest` as a type alias (matches `cmd/adapt.go:41` pattern) to avoid import-name collision with the local interface's `Validate` method parameter
- [ ] 3.1.3 Implement `NewCmdValidate(f *cmdutil.Factory, runF func(*validateOptions) error) *cobra.Command`:
  - Add `--platform` string flag
  - In `RunE`: construct `opts`, populate from `f.IOStreams`, `c.Context()`, and parsed flags; set `opts.Validator = validateValidator(f)` (the wrapper from the proposal)
  - Call `runF(opts)` if non-nil, else `runValidate(opts)`
- [ ] 3.1.4 Implement `runValidate(opts *validateOptions) error`:
  - Validate platform via `core.ValidatePlatform(opts.Platform)`
  - Move the validator logic from `internal/service/validator.go` into `cmd/validate.go` as a private helper `validateDocument(...)` (or keep it inline)
  - Call the validator; if `result.Errors` is non-empty, iterate and call `output.FormatError(opts.IO, err)` for each; return the first typed error so `cmdutil.ExitCodeFor` maps it to `ExitCodeError` (1)
- [ ] 3.1.5 Convert `cmd/validate_test.go` to `iostreams.Test()` + `runF` injection:
  - `NewCmdValidate(f, runF)` with `runF` capturing `*validateOptions`
  - `runValidate(opts)` directly with fake Factory and buffer-backed IOStreams
  - Cover both single-error and multi-error cases
- [ ] 3.1.6 Run `mise run check`; confirm `germinator validate` produces byte-identical output to the pre-change build for both platforms and both valid + invalid inputs

## 3.2 Migrate `cmd/canonicalize.go`

- [ ] 3.2.1 In `cmd/canonicalize.go`, define `canonicalizeOptions` struct with fields: `IO *iostreams.IOStreams`, `Canonicalizer func() (Canonicalizer, error)`, `Ctx context.Context`, `InputPath string`, `OutputPath string`, `Platform string`, `DocType string`
- [ ] 3.2.2 Declare the `Canonicalizer` interface in `cmd/canonicalize.go` (one method: `Canonicalize(ctx, *CanonicalizeRequest) (*core.CanonicalizeResult, error)`); declare `type CanonicalizeRequest = application.CanonicalizeRequest` as a type alias (matches `cmd/adapt.go:41` pattern)
- [ ] 3.2.3 Implement `NewCmdCanonicalize(f, runF)` and `runCanonicalize(opts)`:
  - Move the canonicalizer logic from `internal/service/canonicalizer.go` into `cmd/canonicalize.go` as a private helper
  - The helper reads `opts.InputPath`, parses, canonicalizes, and writes to `opts.OutputPath`
  - In `RunE`: set `opts.Canonicalizer = canonicalizeCanonicalizer(f)` (the wrapper from the proposal)
- [ ] 3.2.4 Convert `cmd/canonicalize_test.go` (or move `internal/service/canonicalizer_test.go` here) to `iostreams.Test()` + `runF` injection
- [ ] 3.2.5 Move `internal/service/canonicalizer_golden_test.go` to `cmd/canonicalize_golden_test.go` (the test cannot stay in `internal/service/` because `internal/service/canonicalizer.go` is deleted in task 3.3.2); update package decl to `cmd_test`; ensure golden file tests still pass byte-identically
- [ ] 3.2.6 Run `mise run check`; confirm `germinator canonicalize` produces byte-identical output for all fixtures

## 3.3 Delete legacy service implementations

- [ ] 3.3.1 Delete `internal/service/validator.go` (logic moved to `cmd/validate.go` in task 3.1.4)
- [ ] 3.3.2 Delete `internal/service/canonicalizer.go` (logic moved to `cmd/canonicalize.go` in task 3.2.3)
- [ ] 3.3.3 Delete `internal/service/validator_test.go` and `internal/service/canonicalizer_test.go` (tests converted to new locations in tasks 3.1.5 and 3.2.4)
- [ ] 3.3.4 Confirm `internal/service/` still contains `transformer.go` and `initializer.go` (used by `adapt` and `init` / `library add`); these are deleted in change-7

## 3.4 Verification

- [ ] 3.4.1 Run `mise run lint` — confirm no new violations
- [ ] 3.4.2 Run `mise run test` — confirm all unit tests pass (including the moved tests)
- [ ] 3.4.3 Run `mise run test:coverage` — confirm coverage for `cmd/validate.go`, `cmd/canonicalize.go`, and their `_test.go` files (converted from `internal/service/{validator,canonicalizer}_test.go`) ≥ 70%
- [ ] 3.4.4 Run `mise run test:e2e` — confirm E2E tests for validate and canonicalize pass
- [ ] 3.4.5 Smoke-test every command end-to-end (regression check from change-2):
  - All commands listed in change-2 task 2.7.6 still work
  - `germinator validate <valid-file>` exits 0
  - `germinator validate <invalid-file>` exits 1 with formatted error
  - `germinator canonicalize <input> <output>` produces byte-identical canonical YAML
- [ ] 3.4.6 Confirm `legacyBridge` still works for non-migrated commands (init, library init, library add, library create, library refresh, library remove, library validate, config, completion, version)

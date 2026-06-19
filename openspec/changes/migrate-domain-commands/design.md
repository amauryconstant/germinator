# Design — Migrate validate and canonicalize

## Context

After change-2 (`wire-factory-and-pilots`) establishes the new pattern with `adapt` and `library resources`, change-3 migrates the next two commands (`validate` and `canonicalize`). These commands share the new pattern (one-file-per-command with `Options` + `NewCmdXxx` + `runXxx`) but differ in their error handling: `validate` returns a list of typed errors and must render each one via `output.FormatError`.

## Goals / Non-Goals

**Goals:**

- `cmd/validate.go` follows the `NewCmdValidate(f, runF) + runValidate(opts)` pattern.
- `cmd/canonicalize.go` follows the same pattern.
- `internal/service/validator.go` and `internal/service/canonicalizer.go` are deleted.
- All validate/canonicalize tests are converted to `iostreams.Test()` + `runF` injection.
- Output remains byte-identical to the pre-change build for all input files.

**Non-Goals:**

- Adding `--output` flag to validate/canonicalize — they produce structured output (errors or canonical YAML) that doesn't benefit from JSON/table formatting.
- Migrating library commands — changes 4, 6, 7.
- Migrating `init` — change-5 (has unique partial-success semantics).
- Deleting `internal/service/` entirely — change-7 (after `Transformer` and `Initializer` consumers migrate).
- Deleting `internal/application/` — change-7.

## Decisions

### 1. Validation errors are rendered via FormatError, not logged

**Choice**: `runValidate` iterates `result.Errors` and calls `output.FormatError(opts.IO, err)` for each error in the result. The function returns the first typed error to the caller so `cmdutil.ExitCodeFor` maps it to `ExitCodeError` (1).

**Rationale**: matches the foundation's `cli/error-formatting` capability; gives users a uniform error rendering experience; lets `cmdutil.ExitCodeFor` work without special-casing validation.

**Alternatives considered**:

- Render errors inside `runValidate` and return nil → `cmdutil.ExitCodeFor` would return 0 (success), losing the "non-zero exit on validation failure" semantic.
- Accumulate errors and return a single `*core.ValidationError` → matches the existing pattern; tests in task 3.1.4 cover this case explicitly.

### 2. `internal/service/validator.go` logic moves into the command file

**Choice**: The validation logic (~99 lines of `internal/service/validator.go`) moves directly into `cmd/validate.go` as private helpers called by `runValidate`. No new `internal/validator/` package is created.

**Rationale**: the validator is a small, parameterless function; encapsulating it in a separate package adds ceremony without proportional benefit. The command file is the natural home per the foundation's "implementations live next to the caller" principle.

**Alternatives considered**:

- Create `internal/validator/` package → rejected; not needed until the validator gains significant logic that doesn't fit in the command file.

### 3. Canonicalizer keeps its input/output file parameters

**Choice**: `runCanonicalize(opts)` reads `opts.InputPath`, writes to `opts.OutputPath`, and returns the canonicalized document. The logic from `internal/service/canonicalizer.go` moves to `cmd/canonicalize.go`.

**Rationale**: the canonicalizer is the simplest of the four core services (no external dependencies); keeping the logic in the command file is the cleanest expression.

### 4. Golden file tests move to `internal/service/` survivors or new locations

**Choice**: The existing `internal/service/canonicalizer_golden_test.go` (172 lines) is moved to `cmd/canonicalize_golden_test.go` (or kept in `internal/service/` with updated imports). Golden file tests for `validator` (none exist; validator uses unit tests, not golden files) move to `cmd/validate_test.go`.

**Rationale**: golden file tests assert byte-identical output for known inputs; moving them with the implementation keeps the regression check close to the code under test.

## Risks / Trade-offs

- **Test conversion may lose coverage** — converting ~555 + 241 = ~800 lines of tests risks dropping edge cases. **Mitigation:** tasks 3.1.5 and 3.2.4 explicitly call out coverage preservation; `mise run test:coverage` is the gate.
- **Golden file tests may need updates if the implementation changes** — moving tests without moving logic could expose unintended behavior changes. **Mitigation:** the implementation moves with the tests; byte-identical output is asserted explicitly in task 3.1.6 and 3.2.5.
- **`internal/service/` is partially emptied** — the directory now has only `transformer.go` and `initializer.go`. **Mitigation:** `cmd/adapt.go` (already migrated) and `init` (change-5) / `library add` (change-6) still depend on these; deletion happens in change-7.

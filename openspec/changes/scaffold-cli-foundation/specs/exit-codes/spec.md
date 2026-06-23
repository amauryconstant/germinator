# exit-codes Specification (delta)

## MODIFIED Requirements

### Requirement: Exit code constants collapsed

The seven exit codes (`0, 1, 2, 3, 4, 5, 6`) SHALL be collapsed to three (`0, 1, 2`):

- `0` = `ExitCodeSuccess`
- `1` = `ExitCodeError` (any operational error)
- `2` = `ExitCodeUsage` (usage, flag, or argument error)

The four removed codes (`3, 4, 5, 6`) and their semantic meaning SHALL be **removed**; semantic meaning lives in typed errors (`core.ParseError`, `core.ValidationError`, etc.) and is dispatched in `output.FormatError` via `errors.As`.

#### Scenario: Only three exit code constants exist

- **WHEN** the codebase is inspected
- **THEN** exactly three exit code constants SHALL be defined in `internal/cmdutil/exit.go`
- **AND** the `ExitCodeConfig`, `ExitCodeGit`, `ExitCodeValidation`, `ExitCodeNotFound` constants SHALL NOT exist
- **AND** no other exit code constant SHALL be defined anywhere else in the codebase

> **Status (slice 1 / foundation):** `cmdutil.ExitCode` type and the three constants exist with full table-driven tests; `main.go` does not yet use `cmdutil.ExitCodeFor`. The legacy seven-code surface and the `Category*` enum are removed in changes 2 and 7 as noted in the REMOVED section.

### Requirement: ExitCodeFor function

The `cmdutil.ExitCodeFor(err error) ExitCode` function SHALL map an error to an exit code:

- `nil` → `ExitCodeSuccess` (0)
- `*pflag.NotExistError`, `*pflag.ValueRequiredError`, `*pflag.InvalidValueError`, `*pflag.InvalidSyntaxError` (typed errors that exist in pflag v1.0.10) → `ExitCodeUsage` (2)
- Cobra usage errors not wrapped by pflag (detected via `strings.HasPrefix` on the error message against known prefixes such as `"unknown flag"`, `"flag needs an argument"`, `"invalid argument"`, `"bad flag syntax"`) → `ExitCodeUsage` (2)
- `*core.PartialSuccessError` with `Succeeded > 0` → `ExitCodeSuccess` (0)
- `*core.PartialSuccessError` with `Succeeded == 0` → `ExitCodeError` (1)
- All other errors → `ExitCodeError` (1)

#### Scenario: ExitCodeFor table-driven

- **WHEN** `cmdutil.ExitCodeFor` is called with each error type from the list above
- **THEN** it SHALL return the corresponding exit code

> **Status (slice 1 / foundation):** the function exists with full table-driven tests in `internal/cmdutil/exit_test.go`. The legacy `Category*` enum and `CategorizeError` function are removed in changes 2 and 7 as noted in the REMOVED section.

## REMOVED Requirements

### Requirement: CategorizeError enum and function

**Reason**: The `Category*` enum-based dispatch cannot distinguish between error *kinds* that share an exit code (e.g., a `ValidationError` and a `ConfigError` both return `ExitCodeError` but render differently). Typed errors with `errors.As` dispatch in `output.FormatError` provide the same routing plus better error wrapping via `%w`.

**Migration**: Replace `cmd.CategorizeError(err)` calls with `errors.As(err, &core.ValidationError{})` (or the appropriate typed assertion). The exit code is then derived via `cmdutil.ExitCodeFor(err)`. Per-error rendering lives in `internal/output/errors.go` as private helpers (`formatValidationError`, `formatConfigError`, etc.).

The `Category*` enum (`CategoryCobra`, `CategoryConfig`, `CategoryValidation`, `CategoryNotFound`, `CategoryGit`, `CategoryGeneric`) and the `CategorizeError(err) Category` function SHALL be **removed**. Typed-error dispatch replaces enum-based categorization.

#### Scenario: CategorizeError removed

- **WHEN** the codebase is inspected
- **THEN** `CategorizeError` SHALL NOT be defined
- **AND** no `Category*` enum value SHALL be defined
- **AND** the `cmd/error_handler.go` file SHALL be deleted (deletion happens in change-2; current file remains until then)

> **Status (slice 1 / foundation):** `CategorizeError` and the `Category*` enum still exist in `cmd/error_handler.go`. Removal happens in change-2 (delete `cmd/error_handler.go`) for the file itself, and in change-7 (delete `legacyBridge`) for any remaining references. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

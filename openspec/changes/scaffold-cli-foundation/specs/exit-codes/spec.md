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

> **Status (slice 1 / foundation):** `cmdutil.ExitCode` type and the three constants exist with table-driven tests; `main.go` does not yet use `cmdutil.ExitCodeFor`. Wiring happens in change-2.

### Requirement: ExitCodeFor function

The `cmdutil.ExitCodeFor(err error) ExitCode` function SHALL map an error to an exit code:

- `nil` → `ExitCodeSuccess` (0)
- `*pflag.NotExistError`, `*pflag.ValueRequiredError`, `*pflag.InvalidValueError`, `*pflag.InvalidSyntaxError` → `ExitCodeUsage` (2)
- Cobra usage errors (detected via string match on the error message prefix) → `ExitCodeUsage` (2)
- `*core.PartialSuccessError` with `Succeeded > 0` → `ExitCodeSuccess` (0)
- `*core.PartialSuccessError` with `Succeeded == 0` → `ExitCodeError` (1)
- All other errors → `ExitCodeError` (1)

#### Scenario: ExitCodeFor table-driven

- **WHEN** `cmdutil.ExitCodeFor` is called with each error type from the list above
- **THEN** it SHALL return the corresponding exit code

> **Status (slice 1 / foundation):** the function exists with full table-driven tests in `internal/cmdutil/exit_test.go`.

## REMOVED Requirements

### Requirement: CategorizeError enum and function

The `Category*` enum (`CategoryCobra`, `CategoryConfig`, `CategoryValidation`, `CategoryNotFound`, `CategoryGit`, `CategoryGeneric`) and the `CategorizeError(err) Category` function SHALL be **removed**. Typed-error dispatch replaces enum-based categorization.

#### Scenario: CategorizeError removed

- **WHEN** the codebase is inspected
- **THEN** `CategorizeError` SHALL NOT be defined
- **AND** no `Category*` enum value SHALL be defined
- **AND** the `cmd/error_handler.go` file SHALL be deleted (deletion happens in change-2; current file remains until then)

> **Status (slice 1 / foundation):** `CategorizeError` and the `Category*` enum still exist in `cmd/error_handler.go`. Removal happens in change-2 (delete `cmd/error_handler.go`) for the file itself, and in change-7 (delete `legacyBridge`) for any remaining references.

## ADDED Requirements

### Requirement: Exit code deprecation warning (canary)

While consumers migrate from the seven-code scheme to the three-code scheme, `cmdutil.ExitCodeFor` SHALL emit a one-version deprecation warning to `Logger.Warn` when the `EXIT_CODE_LEGACY` env var is set OR stderr is a TTY. The warning SHALL be emitted at most once per process.

#### Scenario: Deprecation warning emitted

- **GIVEN** the `EXIT_CODE_LEGACY` env var is set
- **WHEN** `cmdutil.ExitCodeFor` is called for the first time in the process
- **THEN** `opts.IO.Logger.Warn(...)` SHALL be called with a message describing the exit-code collapse and the migration target

#### Scenario: Warning emitted at most once

- **GIVEN** the `EXIT_CODE_LEGACY` env var is set
- **WHEN** `cmdutil.ExitCodeFor` is called multiple times
- **THEN** the warning SHALL be emitted on the first call only

> **Status (slice 1 / foundation):** the canary is deferred to change-2. This requirement documents the behavior change that change-2 will implement.

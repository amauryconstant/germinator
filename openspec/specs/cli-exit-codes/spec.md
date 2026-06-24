# exit-codes Specification

## Purpose

Define semantic exit codes for germinator CLI to enable programmatic error handling by scripts and tools. The CLI uses a small, fixed set of exit codes; semantic meaning lives in typed errors (`core.ParseError`, `core.ValidationError`, etc.) and is dispatched in `output.FormatError` via `errors.As`.

## Requirements

### Requirement: Exit code constants collapsed

The seven exit codes (`0, 1, 2, 3, 4, 5, 6`) SHALL be collapsed to three (`0, 1, 2`):

- `0` = `ExitCodeSuccess`
- `1` = `ExitCodeError` (any operational error)
- `2` = `ExitCodeUsage` (usage, flag, or argument error)

The four removed codes (`3, 4, 5, 6`) and their semantic meaning SHALL be **removed**; semantic meaning lives in typed errors and is dispatched in `output.FormatError` via `errors.As`.

#### Scenario: Only three exit code constants exist

- **WHEN** the codebase is inspected
- **THEN** exactly three exit code constants SHALL be defined in `internal/cmdutil/exit.go`
- **AND** the `ExitCodeConfig`, `ExitCodeGit`, `ExitCodeValidation`, `ExitCodeNotFound` constants SHALL NOT exist
- **AND** no other exit code constant SHALL be defined anywhere else in the codebase

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

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
- `*core.NotFoundError` → `ExitCodeUsage` (2) (the lookup miss is treated as user input that did not resolve, not an operational error)
- Cobra usage errors not wrapped by pflag (detected via `strings.Contains` on the error message against any of the 12 known substrings: `"unknown flag"`, `"flag needs an argument"`, `"invalid argument"`, `"bad flag syntax"`, `"no such flag"`, `"invalid syntax"`, `"unknown shorthand flag"`, `"required flag"`, `"requires at least"`, `"requires exactly"`, `"accepts at most"`, `"requires at most"`) → `ExitCodeUsage` (2)
- `*core.PartialSuccessError` with `Succeeded > 0` → `ExitCodeSuccess` (0)
- `*core.PartialSuccessError` with `Succeeded == 0` → `ExitCodeError` (1)
- All other errors → `ExitCodeError` (1)

#### Scenario: ExitCodeFor table-driven

- **WHEN** `cmdutil.ExitCodeFor` is called with each error type from the list above
- **THEN** it SHALL return the corresponding exit code

### Requirement: Exit code deprecation canary

The `germinator` post-`Execute` error path SHALL emit a one-time deprecation warning to stderr when the resolved exit code is `1` AND at least one of the following conditions holds:

- The `EXIT_CODE_LEGACY` environment variable is set to a non-empty value, OR
- stderr is a TTY (per `IOStreams.IsStderrTTY()`).

The warning SHALL be emitted at most once per process via `sync.Once` and SHALL be suppressed in non-TTY, non-`EXIT_CODE_LEGACY` environments (typical CI). The helper is exposed as `internal/warning.MaybeWarnLegacyExitCode(io *iostreams.IOStreams)` and is invoked from `main.go` immediately before `os.Exit(int(cmdutil.ExitCodeFor(err)))`. The warning is written to `io.ErrOut` via `io.Warnf(...)` (a method on `IOStreams` that wraps `Styles.Warning`); it does NOT depend on `IOStreams.Logger` (which is gated on `GERMINATOR_DEBUG`). `cmdutil.ExitCodeFor` remains a pure function with no side effects and no logger parameter.

#### Scenario: Canary fires in interactive session

- **WHEN** the process exits with code `1` AND (stderr is a TTY OR `EXIT_CODE_LEGACY` is set to a non-empty value)
- **THEN** the deprecation warning SHALL be written to stderr exactly once

#### Scenario: Canary suppressed in non-interactive, non-env-var invocation

- **WHEN** the process exits with code `1` AND stderr is not a TTY AND `EXIT_CODE_LEGACY` is unset
- **THEN** no deprecation warning SHALL be emitted

#### Scenario: Canary fires when explicitly requested

- **WHEN** `EXIT_CODE_LEGACY=1` is set AND the process exits with code `1`
- **THEN** the deprecation warning SHALL be written to stderr exactly once

#### Scenario: Single emission per process

- **GIVEN** `MaybeWarnLegacyExitCode` has already been called once during the current process
- **WHEN** any subsequent exit code is `1`
- **THEN** the deprecation warning SHALL NOT be emitted again

#### Scenario: ResetCanaryForTest resets once-state

- **GIVEN** `MaybeWarnLegacyExitCode` has been called once during the current process
- **WHEN** `ResetCanaryForTest()` is invoked
- **THEN** the next call to `MaybeWarnLegacyExitCode` SHALL be permitted to emit the warning (subject to gate conditions)

#### Scenario: Warning emission is independent of Logger

- **WHEN** `MaybeWarnLegacyExitCode` is called with an `IOStreams` whose `Logger` field is nil
- **THEN** the function SHALL still write the warning to `io.ErrOut` (the canary does not depend on the Logger)

#### Scenario: Exit code 2 does not trigger the canary

- **WHEN** the process exits with code `2` (`ExitCodeUsage`) under any TTY or env-var conditions
- **THEN** no deprecation warning SHALL be emitted (the canary is gated on exit code `1` only)

## Fulfilled

**Change:** `migrate-library-rest` (slice 7 of 9)
**Date:** 2026-07-01

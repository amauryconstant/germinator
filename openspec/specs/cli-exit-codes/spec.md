# exit-codes Specification

> **Cross-references:** this change also modifies `errors-typed-errors`, `cli-error-formatting`, `errors-enhanced-errors`. See those delta specs.

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
- `*pflag.NotExistError`, `*pflag.ValueRequiredError`, `*pflag.InvalidValueError`, `*pflag.InvalidSyntaxError` (typed errors that exist in pflag `v1.0.10`) → `ExitCodeUsage` (2)
- `*core.NotFoundError` → `ExitCodeError` (1) (CORRECTED — was `ExitCodeUsage` (2); a lookup miss is a runtime state, not a user-input validation error)
- `*core.UsageError` → `ExitCodeUsage` (2) (`UsageError` is introduced for CLI flag validation errors; CLI flag validation maps to usage exit code)
- `*core.CobraUsageError` → `ExitCodeUsage` (2) (sentinel for Cobra arg-validation errors that the typed pflag dispatch does not cover; `cobra.ExactArgs`/`MinimumNArgs`/`MaximumNArgs`/`RangeArgs` failures and `required flag(s) "..."` errors)
- `*core.PartialSuccessError` with `Succeeded > 0` → `ExitCodeSuccess` (0)
- `*core.PartialSuccessError` with `Succeeded == 0` → `ExitCodeError` (1)
- `*config.WriteError` → `ExitCodeError` (1) (`*config.WriteError` carries `op`, `path`, `cause`; an I/O failure during config scaffolding is an operational error, not a user-input validation error)
- All other errors → `ExitCodeError` (1)

The four `*pflag.*Error` types are verified stable in pflag `v1.0.10` (`pflag/errors.go:21-149`); if a future pflag version drops or renames any of them, dispatch falls through to `ExitCodeError` (1) and the corresponding test row in `internal/cmdutil/exit_test.go` fails fast.

#### Scenario: ExitCodeFor table-driven

- **WHEN** `cmdutil.ExitCodeFor` is called with each error type from the list above
- **THEN** it SHALL return the corresponding exit code

#### Scenario: NotFoundError returns ExitCodeError (1)

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeError` (1) — the prior mapping (2) was a semantic error. This is a BREAKING change for any script that special-cases `exit 2` for not-found scenarios; the CHANGELOG calls it out.

#### Scenario: UsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2) — `UsageError` is a CLI flag validation error and maps to the usage exit code.

#### Scenario: CobraUsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.CobraUsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2) — the sentinel wraps Cobra arg-validation errors emitted as `fmt.Errorf` strings, which historically landed on the substring fallback.

#### Scenario: Typed-error dispatch takes precedence over generic fallback

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with an error that wraps `*pflag.NotExistError` (or one of the other three `*pflag.*Error` types)
- **THEN** it SHALL return `ExitCodeUsage` (2) via the `errors.As` dispatch
- **AND** the dispatch SHALL NOT rely on substring matching (the substring fallback is dropped)

### Requirement: UsageError exit-code mapping

`*core.UsageError` SHALL map to `ExitCodeUsage` (2) in `cmdutil.ExitCodeFor`. Consistent with the existing mapping for `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` and `*core.CobraUsageError` — CLI flag validation errors are user-input problems and produce exit code 2.

#### Scenario: UsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

### Requirement: CobraUsageError exit-code mapping

`*core.CobraUsageError` SHALL map to `ExitCodeUsage` (2) in `cmdutil.ExitCodeFor`. Replaces the substring-prefix dispatch fallback that previously mapped errors emitted by `cobra.ExactArgs`, `cobra.MinimumNArgs`, `cobra.MaximumNArgs`, `cobra.RangeArgs`, and `cobra.MarkFlagRequired`-derived `"required flag(s) \"...\" not set"` strings.

#### Scenario: CobraUsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.CobraUsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

### Requirement: WriteError exit-code mapping

`*config.WriteError` SHALL map to `ExitCodeError` (1) in `cmdutil.ExitCodeFor`. An I/O failure during config scaffolding (e.g., `WriteDefault` path) is an operational error, not a user-input validation error; the error wraps the underlying `*os.PathError` via `cause` for inspection.

#### Scenario: WriteError returns ExitCodeError (1)

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*config.WriteError`
- **THEN** it SHALL return `ExitCodeError` (1) — a config I/O failure is an operational error, not a CLI usage error.

### Requirement: Substring dispatch fallback removal

`cmdutil.ExitCodeFor` SHALL NOT use substring-prefix matching against Cobra error messages. The pre-change implementation used a 12-prefix substring-match list; the post-change implementation uses typed dispatch via `errors.As` against `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` plus the new `*core.CobraUsageError` sentinel.

#### Scenario: No substring matching in ExitCodeFor

- **WHEN** `cmdutil.ExitCodeFor` is inspected
- **THEN** it SHALL NOT contain a `cobraUsagePrefixes` slice
- **AND** it SHALL NOT contain a `hasCobraUsagePrefix` helper
- **AND** the dispatch SHALL rely on `errors.As` against the typed errors listed above

## Fulfilled

**Change:** `migrate-library-rest` (slice 7 of 9)
**Date:** 2026-07-01

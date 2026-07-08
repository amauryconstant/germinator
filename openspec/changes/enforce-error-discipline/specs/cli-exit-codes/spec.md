# exit-codes Specification (delta)

## MODIFIED Requirements

### Requirement: ExitCodeFor function

The `cmdutil.ExitCodeFor(err error) ExitCode` function SHALL map an error to an exit code:

- `nil` → `ExitCodeSuccess` (0)
- `*pflag.NotExistError`, `*pflag.ValueRequiredError`, `*pflag.InvalidValueError`, `*pflag.InvalidSyntaxError` (typed errors that exist in pflag v1.0.10) → `ExitCodeUsage` (2)
- `*core.NotFoundError` → `ExitCodeError` (1) (CORRECTED — was `ExitCodeUsage` (2) prior to change `enforce-error-discipline`; a lookup miss is a runtime state, not a user-input validation error)
- `*core.UsageError` → `ExitCodeUsage` (2) (NEW — `UsageError` introduced in change `enforce-error-discipline`)
- Cobra usage errors not wrapped by pflag (detected via `errors.As` against `*cobra.FlagError` / `*pflag.Error`) → `ExitCodeUsage` (2) (CHANGED — the pre-change implementation used substring matching against a list of 12 Cobra error prefixes; the post-change implementation uses typed-error dispatch with a narrow fallback prefix list)
- `*core.PartialSuccessError` with `Succeeded > 0` → `ExitCodeSuccess` (0)
- `*core.PartialSuccessError` with `Succeeded == 0` → `ExitCodeError` (1)
- All other errors → `ExitCodeError` (1)

**Change**: corrected the `NotFoundError` mapping (1 instead of 2); added `UsageError` mapping; replaced substring matching with `errors.As`-based dispatch. The prior mapping was semantically wrong (per the 2026-07-08 review: "the lookup miss is treated as user input that did not resolve, not an operational error" was the original rationale, but the review correctly identifies this as confused — a "not found" is a runtime state, not user input). The Cobra substring matching was brittle to upstream wording drift; `errors.As` dispatch against typed errors is robust.

#### Scenario: ExitCodeFor table-driven

- **WHEN** `cmdutil.ExitCodeFor` is called with each error type from the list above
- **THEN** it SHALL return the corresponding exit code

#### Scenario: NotFoundError returns ExitCodeError (1)

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeError` (1) — the prior mapping (2) was a semantic error and is corrected in change `enforce-error-discipline`. This is a BREAKING change for any script that special-cases `exit 2` for not-found scenarios; the CHANGELOG calls it out.

#### Scenario: UsageError returns ExitCodeUsage (2)

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2) — `UsageError` is a CLI flag validation error and maps to the usage exit code.

#### Scenario: Typed-error dispatch takes precedence

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with an error that wraps a `*pflag.Error` or `*cobra.FlagError`
- **THEN** it SHALL return `ExitCodeUsage` (2) via the `errors.As` dispatch
- **AND** the substring matching fallback SHALL be reached ONLY when no typed error is detected (the fallback list is narrowed to the 3-4 most common prefixes that Cobra has consistently used)

## ADDED Requirements

### Requirement: UsageError exit-code mapping

`*core.UsageError` SHALL map to `ExitCodeUsage` (2) in `cmdutil.ExitCodeFor`. This is consistent with the existing mapping for `*pflag.Error` and `*cobra.FlagError` — CLI flag validation errors are user-input problems and produce exit code 2.

**Change**: NEW requirement. The type is introduced in change `enforce-error-discipline`.

#### Scenario: UsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

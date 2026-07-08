# typed-errors Specification (delta)

## ADDED Requirements

### Requirement: UsageError type

The system SHALL provide a `UsageError` type for CLI flag validation errors that are not caught by Cobra's flag parsing. `UsageError` carries the flag name and the reason as exported fields and maps to `ExitCodeUsage` (2) via `cmdutil.ExitCodeFor`. It is distinct from `ValidationError` (which is for document validation, not CLI flag validation) and from `ConfigError` (which is for config-file errors).

**Change**: NEW requirement. `UsageError` is added in change `enforce-error-discipline` to replace the string-encoded `errEmptyResources = errors.New("flag needs an argument: --resources ...")` pattern. The migration is in the `cmd/library_add.go:82` site.

#### Scenario: UsageError has Flag and Reason fields

- **WHEN** `NewUsageError(flag, reason string)` is called
- **THEN** it SHALL return a `*UsageError{Flag: flag, Reason: reason}`
- **AND** the `Flag` and `Reason` fields SHALL be exported

#### Scenario: UsageError Error format

- **WHEN** `err.Error()` is called on a `*UsageError{Flag: "--resources", Reason: "must be non-empty list of refs"}`
- **THEN** it SHALL return the string `"flag needs an argument: --resources (must be non-empty list of refs)"`

#### Scenario: UsageError maps to ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

#### Scenario: UsageError is in FormatError dispatch set

- **WHEN** `output.FormatError(io, err)` is called with `*core.UsageError`
- **THEN** the rendered message SHALL be `"Error: <flag>: <reason>"` written to `io.ErrOut`

## MODIFIED Requirements

### Requirement: NotFoundError type

The system SHALL provide a `NotFoundError` type for missing-entity lookups (library refs, presets). It carries `Entity` and `Key` as exported fields and maps to exit code 1 (operational error) via `cmdutil.ExitCodeFor`.

**Change**: clarify that `NotFoundError` maps to `ExitCodeError` (1), not `ExitCodeUsage` (2). The prior mapping (2) was semantically wrong per the 2026-07-08 review: "not found" is a runtime state, not a user-input validation error. The change `enforce-error-discipline` updates `internal/cmdutil/exit.go:73` and `internal/cmdutil/exit_test.go:58` accordingly.

#### Scenario: NewNotFoundError constructor

- **WHEN** `NewNotFoundError(entity, key string)` is called
- **THEN** it SHALL return a `*NotFoundError{Entity: entity, Key: key}`

#### Scenario: NotFoundError Error format

- **WHEN** `err.Error()` is called on a `*NotFoundError{Key: "ghost"}`
- **THEN** it SHALL return the string `"not found: ghost"`

#### Scenario: NotFoundError maps to ExitCodeError

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeError` (1) — `*core.NotFoundError` represents a runtime lookup miss, not a user-input validation error; the prior mapping (2) was semantically wrong and is corrected in change `enforce-error-discipline`.

# E2E Canonicalize Tests Specification

## Purpose

End-to-end tests for the canonicalize command to validate platform document canonicalization through actual binary execution.

> **Exit-code semantics:** All scenarios below use the `cli-exit-codes` mapping (`internal/cmdutil/exit.go:ExitCodeFor`). Specifically:
>
> - Missing required flags (`--platform`, `--type`) → `ExitCodeUsage` (2) (Cobra argument error)
> - Invalid flag values (`--platform invalid`, `--type invalid`) → `ExitCodeUsage` (2) (pflag `InvalidValueError`)
> - Nonexistent input file → `ExitCodeError` (1) (`*core.FileError` not flagged as not-found because canonicalize may surface other causes)
>
> Where the spec says "exit code SHALL be 1", replace with the more precise mapping above per `cli-exit-codes`.

## Requirements

### Requirement: Canonicalize Command E2E Tests

The canonicalize command SHALL be tested for all expected behaviors.

#### Scenario: Canonicalize valid document succeeds
- **WHEN** `germinator canonicalize <valid-platform-doc> <output> --platform opencode --type agent` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "Successfully canonicalized"
- **AND** output file SHALL be created

#### Scenario: Canonicalize with missing platform flag fails
- **WHEN** `germinator canonicalize <doc> <output> --type agent` is executed without `--platform`
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain "required" or "platform"

#### Scenario: Canonicalize with missing type flag fails
- **WHEN** `germinator canonicalize <doc> <output> --platform opencode` is executed without `--type`
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain "required" or "type"

#### Scenario: Canonicalize with invalid platform fails
- **WHEN** `germinator canonicalize <doc> <output> --platform invalid --type agent` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL indicate the platform is invalid

#### Scenario: Canonicalize with invalid type fails
- **WHEN** `germinator canonicalize <doc> <output> --platform opencode --type invalid` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL indicate the type is invalid
#### Scenario: Canonicalize nonexistent file fails

- **WHEN** `germinator canonicalize nonexistent.yaml <output> --platform opencode --type agent` is executed
- **THEN** exit code SHALL be 1 (`ExitCodeError` via `*core.FileError`)
- **AND** stderr SHALL contain an error message

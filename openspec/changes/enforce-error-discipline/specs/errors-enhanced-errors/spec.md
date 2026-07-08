# enhanced-errors Specification (delta)

## ADDED Requirements

### Requirement: BatchFailureInfo carries typed-error chain

The `BatchFailureInfo` struct in `internal/library/adder.go` SHALL expose `ErrorType` and `Cause` fields in addition to the existing `Path` and `Error` fields. The new fields preserve the typed-error chain so downstream code can use `errors.Is` / `errors.As` against the original error rather than matching the stringified `Error` field.

**Change**: NEW requirement. The pre-change `BatchFailureInfo` had only `Path string` and `Error string` fields; the new fields are added in change `enforce-error-discipline` to fix the typed-error chain loss in the `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` pattern at `cmd/library_add.go:534`.

#### Scenario: BatchFailureInfo has four fields

- **WHEN** a `BatchFailureInfo` is constructed
- **THEN** it SHALL have the following fields:
  - `Path string` — the file path that failed
  - `Error string` — the stringified error message (preserved for JSON consumers)
  - `ErrorType string` — the type name of the typed error (e.g., `"FileError"`, `"NotFoundError"`, `"ParseError"`, `""` if not a typed error)
  - `Cause error` — the original typed error (json tag `omitempty` so it does not appear in serialized output when nil)

#### Scenario: ErrorType is the type name of the cause

- **WHEN** a `BatchFailureInfo` is built from a typed error (e.g., `*core.FileError`)
- **THEN** `ErrorType` SHALL be the result of `fmt.Sprintf("%T", cause)` with the `*core.` prefix stripped
- **AND** `Cause` SHALL be the original typed error

#### Scenario: Cause is omitempty in JSON output

- **WHEN** `BatchFailureInfo` is serialized to JSON via `output.NewJSONExporter`
- **THEN** a `Cause` field with a nil value SHALL NOT appear in the JSON output
- **AND** a `Cause` field with a non-nil value SHALL appear as the underlying error's `Error()` string (via the standard `error` JSON marshalling)

#### Scenario: Cause supports errors.Is / errors.As

- **WHEN** a downstream caller calls `errors.Is(failure.Cause, sentinel)` or `errors.As(failure.Cause, &target)`
- **THEN** the call SHALL succeed against the original typed error
- **AND** the call SHALL match the chain if the original error was wrapped via `fmt.Errorf("...: %w", inner)`

# enhanced-errors Specification (delta)

> **Cross-references:** this change also modifies `errors-typed-errors`, `cli-error-formatting`, `cli-exit-codes`. See those delta specs.

## REMOVED Requirements

None.

## ADDED Requirements

### Requirement: BatchFailureInfo carries typed-error chain

The `BatchFailureInfo` struct declared in `internal/library/adder.go:541-544` SHALL expose `ErrorType` and `Cause` fields in addition to the existing `Source` and `Error` fields. The new fields preserve the typed-error chain so downstream code can use `errors.Is` / `errors.As` against the original error rather than matching the stringified `Error` field.

**Change**: NEW requirement. The pre-change `BatchFailureInfo` had only `Source string` and `Error string` fields; the new fields are added in change `enforce-error-discipline` to fix the typed-error chain loss in the prior `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` pattern at `cmd/library_add.go:527` (`runAddBatchFiles` failure path) and `cmd/library_add.go:685` (`collectDiscoverFailures` batch failure path) — which discarded the typed cause.

#### Scenario: BatchFailureInfo has four fields

- **WHEN** a `BatchFailureInfo` is constructed
- **THEN** it SHALL have the following fields:
  - `Source string \`json:"source"\`` — the file path that failed (existing field).
  - `Error string \`json:"error"\`` — the stringified error message (preserved for JSON consumers, existing field).
  - `ErrorType string \`json:"errorType,omitempty"\`` — the type name of the typed error (e.g., `"FileError"`, `"NotFoundError"`, `"ParseError"`), the `*core.` (or `*gerrors.` here) prefix is stripped, empty string when no cause.
  - `Cause error \`json:"cause,omitempty"\`` — the original typed error; the `omitempty` tag ensures it does not appear in serialized output when nil.

#### Scenario: ErrorType is the type name of the cause

- **WHEN** a `BatchFailureInfo` is built from a typed error (e.g., `*gerrors.NotFoundError`)
- **THEN** `ErrorType` SHALL be the canonical name of the cause's outermost concrete type, computed via a direct `switch cause.(type)` in `internal/library/adder.go` (e.g., `*gerrors.NotFoundError` → `"NotFoundError"`, `*gerrors.FileError` → `"FileError"`, `*gerrors.ValidationError` → `"ValidationError"`, `*gerrors.ParseError` → `"ParseError"`, `*gerrors.ConfigError` → `"ConfigError"`, `*gerrors.OperationError` → `"OperationError"`, `*gerrors.InitializeError` → `"InitializeError"`, `*gerrors.PartialSuccessError` → `"PartialSuccessError"`, `*gerrors.UsageError` → `"UsageError"`, `*gerrors.CobraUsageError` → `"CobraUsageError"`, `*os.PathError` → `"PathError"`, default → `fmt.Sprintf("%T", cause)`). The dispatch is a direct type-switch (not `errors.As`), so the label reflects the cause's outermost concrete type — for a typed error that wraps another typed error via `fmt.Errorf("...: %w", inner)`, the label is the outermost type (`*OperationError`), not the wrapped cause (`*NotFoundError`); for a non-pointer value (none of `core.*Error` are addressed by value), `fmt.Sprintf("%T", cause)` returns the underlying struct name.
- **AND** `Cause` SHALL be the original typed error
- **AND** when `Cause` is nil, `ErrorType` SHALL be the empty string

#### Scenario: Cause is omitempty in JSON output

- **WHEN** `BatchFailureInfo` is serialized to JSON via `output.NewJSONExporter`
- **THEN** a `Cause` field with a nil value SHALL NOT appear in the JSON output
- **AND** a `Cause` field with a non-nil `*core.*Error` value SHALL appear as `{"error": "<Error()>"}` (because every `core.*Error` type implements `MarshalJSON`; see `errors-typed-errors/spec.md`)

#### Scenario: Cause MUST be a typed error that implements `json.Marshaler`

- **WHEN** a `BatchFailureInfo.Cause` is assigned at any of the 5 population sites in `internal/library/adder.go`
- **THEN** the cause SHALL be a `*core.*Error` typed error (or any other type that implements `json.Marshaler`)
- **AND** non-typed causes (e.g., `*os.PathError`, plain `errors.New(...)`) SHALL be wrapped in `*core.FileError` (or another typed error) at the population site BEFORE assignment to `f.Cause`
- **AND** the JSON serialization contract above assumes typed causes — non-typed causes would marshal as `{}` per stdlib `json.Marshaler` precedence rules, which defeats the typed-error-chain preservation contract

#### Scenario: Stdlib cancellation cause wrapped as typed cause

- **WHEN** a `BatchFailureInfo.Cause` would otherwise be a stdlib `context.Canceled` (returned by `ctx.Err()` during `processBatchAddFile`'s pre-flight cancellation check at `internal/library/adder.go:667`)
- **THEN** the population site SHALL wrap the stdlib error in `*core.FileError{path: source, op: "context", message: cerr.Error(), cause: cerr}` BEFORE assigning to `f.Cause`
- **AND** `f.ErrorType` SHALL be `"FileError"` (the wrapped type, not `"*errors.errorString"` from the unwrapped stdlib default)
- **AND** the original `cerr` SHALL remain reachable via `errors.Is(f.Cause.Unwrap(), context.Canceled)` so cancellation observability is preserved
- **Rationale**: the `context.Canceled` value does not implement `json.Marshaler`; without wrapping, the JSON wire format for `cause` would marshal as `{}`, violating the typed-cause contract.

#### Scenario: Cause supports errors.Is / errors.As

- **WHEN** a downstream caller calls `errors.Is(failure.Cause, sentinel)` or `errors.As(failure.Cause, &target)`
- **THEN** the call SHALL succeed against the original typed error
- **AND** the call SHALL match the chain if the original error was wrapped via `fmt.Errorf("...: %w", inner)`

#### Scenario: Population sites cover all BatchFailureInfo literals

- **WHEN** the change `enforce-error-discipline` lands
- **THEN** every `BatchFailureInfo{...}` literal at `internal/library/adder.go:667, :684, :702, :720, :784` SHALL populate `ErrorType` (via the typed switch in `adder.go`) and `Cause` (when the cause is known) in addition to `Source` and `Error`

### Requirement: BatchFailureInfo JSON wire-format compatibility

The additive `ErrorType` and `Cause` fields SHALL preserve backward compatibility with JSON consumers that parse the legacy 2-field shape (`source`, `error`).

#### Scenario: Legacy consumer sees no breaking change

- **WHEN** a downstream JSON consumer reads the `source` and `error` fields from a serialized `BatchFailureInfo`
- **THEN** the consumer SHALL observe the same values as before the change
- **AND** the consumer SHALL ignore the new `errorType` and `cause` fields

## MODIFIED Requirements

None.

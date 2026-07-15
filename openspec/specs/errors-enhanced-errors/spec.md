> **Cross-references:** this change also modifies `errors-typed-errors`, `cli-error-formatting`, `cli-exit-codes`. See those delta specs.

## Purpose

Apply the immutable builder pattern (private fields, getters, WithSuggestions(), WithContext()) from ValidationError to all error types (ParseError, TransformError, FileError, ConfigError) for complete API alignment.

## Requirements

### Requirement: ParseError with private fields and getters

The system SHALL provide an enhanced `ParseError` in `internal/core/errors.go` with private fields and getter methods.

#### Scenario: ParseError has private fields

- **WHEN** ParseError is examined
- **THEN** it SHALL have private fields: `path`, `message`, `cause`, `suggestions`, `context`
- **AND** these fields SHALL NOT be directly accessible

#### Scenario: ParseError getters work correctly

- **WHEN** `NewParseError("agent.md", "missing delimiter", cause)` is called
- **THEN** `err.Path()` SHALL return "agent.md"
- **AND** `err.Message()` SHALL return "missing delimiter"
- **AND** `err.Cause()` SHALL return the cause error
- **AND** `err.Suggestions()` SHALL return an empty slice
- **AND** `err.Context()` SHALL return an empty string

---

### Requirement: ParseError WithSuggestions builder

The system SHALL provide a `WithSuggestions(suggestions []string) *ParseError` method that returns a new instance.

#### Scenario: WithSuggestions returns new instance

- **WHEN** `err.WithSuggestions([]string{"add --- at start"})` is called
- **THEN** it SHALL return a new ParseError (not modify original)
- **AND** the new error SHALL have the suggestions
- **AND** the original error SHALL remain unchanged

#### Scenario: WithSuggestions chains with WithContext

- **WHEN** `NewParseError(path, msg, cause).WithSuggestions(s).WithContext("parsing agent")` is called
- **THEN** it SHALL return a new ParseError with both suggestions and context
- **AND** all builders SHALL be chainable

---

### Requirement: ParseError WithContext builder

The system SHALL provide a `WithContext(context string) *ParseError` method that returns a new instance.

#### Scenario: WithContext returns new instance

- **WHEN** `err.WithContext("parsing agent.md")` is called
- **THEN** it SHALL return a new ParseError (not modify original)
- **AND** the new error SHALL have the context

---

### Requirement: TransformError with private fields and getters

The system SHALL provide an enhanced `TransformError` in `internal/core/errors.go` with private fields and getter methods.

#### Scenario: TransformError has private fields

- **WHEN** TransformError is examined
- **THEN** it SHALL have private fields: `operation`, `platform`, `message`, `cause`, `suggestions`, `context`
- **AND** these fields SHALL NOT be directly accessible

#### Scenario: TransformError getters work correctly

- **WHEN** `NewTransformError("render", "opencode", "template failed", cause)` is called
- **THEN** `err.Operation()` SHALL return "render"
- **AND** `err.Platform()` SHALL return "opencode"
- **AND** `err.Message()` SHALL return "template failed"
- **AND** `err.Cause()` SHALL return the cause error
- **AND** `err.Suggestions()` SHALL return an empty slice
- **AND** `err.Context()` SHALL return an empty string

---

### Requirement: TransformError WithSuggestions builder

The system SHALL provide a `WithSuggestions(suggestions []string) *TransformError` method.

#### Scenario: WithSuggestions returns new instance

- **WHEN** `err.WithSuggestions([]string{"check template directory"})` is called
- **THEN** it SHALL return a new TransformError (not modify original)
- **AND** the new error SHALL have the suggestions

---

### Requirement: TransformError WithContext builder

The system SHALL provide a `WithContext(context string) *TransformError` method.

#### Scenario: WithContext returns new instance

- **WHEN** `err.WithContext("processing agent.md")` is called
- **THEN** it SHALL return a new TransformError (not modify original)
- **AND** the new error SHALL have the context

---

### Requirement: FileError with private fields and getters

The system SHALL provide an enhanced `FileError` in `internal/core/errors.go` with private fields and getter methods.

#### Scenario: FileError has private fields

- **WHEN** FileError is examined
- **THEN** it SHALL have private fields: `path`, `operation`, `message`, `cause`, `suggestions`, `context`
- **AND** these fields SHALL NOT be directly accessible

#### Scenario: FileError getters work correctly

- **WHEN** `NewFileError("agent.md", "read", "file not found", cause)` is called
- **THEN** `err.Path()` SHALL return "agent.md"
- **AND** `err.Operation()` SHALL return "read"
- **AND** `err.Message()` SHALL return "file not found"
- **AND** `err.Cause()` SHALL return the cause error
- **AND** `err.Suggestions()` SHALL return an empty slice
- **AND** `err.Context()` SHALL return an empty string

---

### Requirement: FileError WithSuggestions builder

The system SHALL provide a `WithSuggestions(suggestions []string) *FileError` method.

#### Scenario: WithSuggestions returns new instance

- **WHEN** `err.WithSuggestions([]string{"check .claude/ directory"})` is called
- **THEN** it SHALL return a new FileError (not modify original)
- **AND** the new error SHALL have the suggestions

---

### Requirement: FileError WithContext builder

The system SHALL provide a `WithContext(context string) *FileError` method.

#### Scenario: WithContext returns new instance

- **WHEN** `err.WithContext("loading agent configuration")` is called
- **THEN** it SHALL return a new FileError (not modify original)
- **AND** the new error SHALL have the context

---

### Requirement: ConfigError with private fields and getters

The system SHALL provide an enhanced `ConfigError` in `internal/core/errors.go` with private fields and getter methods.

#### Scenario: ConfigError has private fields

- **WHEN** ConfigError is examined
- **THEN** it SHALL have private fields: `field`, `value`, `message`, `suggestions`, `context`
- **AND** these fields SHALL NOT be directly accessible
- **AND** the field SHALL be named `suggestions` (renamed from `Available`)

#### Scenario: ConfigError getters work correctly

- **WHEN** `NewConfigError("platform", "invalid", "unknown platform")` is called
- **THEN** `err.Field()` SHALL return "platform"
- **AND** `err.Value()` SHALL return "invalid"
- **AND** `err.Message()` SHALL return "unknown platform"
- **AND** `err.Suggestions()` SHALL return an empty slice
- **AND** `err.Context()` SHALL return an empty string

---

### Requirement: ConfigError WithSuggestions builder

The system SHALL provide a `WithSuggestions(suggestions []string) *ConfigError` method.

#### Scenario: WithSuggestions returns new instance

- **WHEN** `err.WithSuggestions([]string{"opencode", "claude-code"})` is called
- **THEN** it SHALL return a new ConfigError (not modify original)
- **AND** the new error SHALL have the suggestions

#### Scenario: WithSuggestions replaces old Available field

- **WHEN** ConfigError is created with `NewConfigError("platform", "", "required").WithSuggestions([]string{"opencode", "claude-code"})`
- **THEN** `err.Suggestions()` SHALL return ["opencode", "claude-code"]
- **AND** this SHALL replace the old `Available` field functionality

---

### Requirement: ConfigError WithContext builder

The system SHALL provide a `WithContext(context string) *ConfigError` method.

#### Scenario: WithContext returns new instance

- **WHEN** `err.WithContext("validating CLI flags")` is called
- **THEN** it SHALL return a new ConfigError (not modify original)
- **AND** the new error SHALL have the context

---

### Requirement: ConfigError constructor signature change

The system SHALL change the `NewConfigError` constructor signature to remove the `available` parameter.

#### Scenario: New constructor takes three parameters

- **WHEN** `NewConfigError("platform", "invalid", "unknown platform")` is called
- **THEN** it SHALL create a ConfigError with field="platform", value="invalid", message="unknown platform"
- **AND** suggestions SHALL be empty
- **AND** context SHALL be empty

#### Scenario: Old constructor signature no longer exists

- **WHEN** code tries to use the old 4-parameter NewConfigError signature
- **THEN** it SHALL fail to compile

#### Scenario: Available options added via builder

- **WHEN** config error needs to show valid options
- **THEN** it SHALL use `NewConfigError(field, value, message).WithSuggestions([]string{...})`
- **AND** NOT pass available options in constructor

---

### Requirement: All error types use immutable builders

All error type builders (WithSuggestions, WithContext) SHALL be immutable and return new instances.

#### Scenario: ParseError builders are immutable

- **WHEN** `err1 := NewParseError(path, msg, cause)` and `err2 := err1.WithSuggestions(s)` are called
- **THEN** `err1` SHALL remain unchanged
- **AND** `err2` SHALL be a new instance with suggestions

#### Scenario: TransformError builders are immutable

- **WHEN** `err1 := NewTransformError(op, platform, msg, cause)` and `err2 := err1.WithContext("context")` are called
- **THEN** `err1` SHALL remain unchanged
- **AND** `err2` SHALL be a new instance with context

#### Scenario: FileError builders are immutable

- **WHEN** `err1 := NewFileError(path, op, msg, cause)` and `err2 := err1.WithSuggestions(s)` are called
- **THEN** `err1` SHALL remain unchanged
- **AND** `err2` SHALL be a new instance with suggestions

#### Scenario: ConfigError builders are immutable

- **WHEN** `err1 := NewConfigError(field, value, msg)` and `err2 := err1.WithSuggestions(s)` are called
- **THEN** `err1` SHALL remain unchanged
- **AND** `err2` SHALL be a new instance with suggestions

---

### Requirement: Suggestions getter returns copy

All error types' `Suggestions()` getter SHALL return a copy of the suggestions slice, not the original.

#### Scenario: ParseError Suggestions returns copy

- **WHEN** `s := err.Suggestions()` is called and `s[0] = "modified"`
- **THEN** the original error's suggestions SHALL remain unchanged

#### Scenario: ConfigError Suggestions returns copy

- **WHEN** `s := err.Suggestions()` is called and `s[0] = "modified"`
- **THEN** the original error's suggestions SHALL remain unchanged

---

### Requirement: Error methods include context and suggestions

All error types' `Error()` method SHALL include context and suggestions in the formatted output.

#### Scenario: ParseError Error includes suggestions

- **WHEN** `err.Error()` is called on a ParseError with suggestions
- **THEN** the string SHALL contain the suggestions formatted as "Hint: <suggestion>"

#### Scenario: ParseError Error includes context

- **WHEN** `err.WithContext("loading config").Error()` is called
- **THEN** the string SHALL contain "Context: loading config"

#### Scenario: ConfigError Error includes suggestions

- **WHEN** `err.Error()` is called on a ConfigError with suggestions
- **THEN** the string SHALL contain the suggestions formatted as "Hint: <suggestion>"
- **AND** this SHALL replace the old "Available: ..." format

---

### Requirement: ValidationError Unwrap method for API consistency

The system SHALL provide an `Unwrap() error` method on ValidationError that returns nil.

#### Scenario: ValidationError Unwrap returns nil

- **WHEN** `err.Unwrap()` is called on a ValidationError
- **THEN** it SHALL return nil
- **AND** this SHALL provide API consistency with ParseError, TransformError, and FileError

### Requirement: BatchFailureInfo carries typed-error chain

The `BatchFailureInfo` struct declared in `internal/library/adder.go:541-544` SHALL expose `ErrorType` and `Cause` fields in addition to the existing `Source` and `Error` fields. The new fields preserve the typed-error chain so downstream code can use `errors.Is` / `errors.As` against the original error rather than matching the stringified `Error` field.

#### Scenario: BatchFailureInfo has four fields

- **WHEN** a `BatchFailureInfo` is constructed
- **THEN** it SHALL have the following fields:
  - `Source string` — the file path that failed (existing field).
  - `Error string` — the stringified error message (preserved for JSON consumers, existing field).
  - `ErrorType string` — the type name of the typed error (e.g., `"FileError"`, `"NotFoundError"`, `"ParseError"`), the `*core.` prefix is stripped, empty string when no cause.
  - `Cause error` — the original typed error; `omitempty` in JSON ensures it does not appear when nil.

#### Scenario: ErrorType is the type name of the cause

- **WHEN** a `BatchFailureInfo` is built from a typed error (e.g., `*core.NotFoundError`)
- **THEN** `ErrorType` SHALL be the canonical name of the cause's outermost concrete type (e.g., `*core.NotFoundError` → `"NotFoundError"`, `*core.FileError` → `"FileError"`, `*core.ValidationError` → `"ValidationError"`, `*core.ParseError` → `"ParseError"`, `*core.ConfigError` → `"ConfigError"`, `*core.OperationError` → `"OperationError"`, `*core.InitializeError` → `"InitializeError"`, `*core.PartialSuccessError` → `"PartialSuccessError"`, `*core.UsageError` → `"UsageError"`, `*core.CobraUsageError` → `"CobraUsageError"`, `*os.PathError` → `"PathError"`, default → `fmt.Sprintf("%T", cause)`). The dispatch is a direct type-switch, so the label reflects the outermost type.
- **AND** `Cause` SHALL be the original typed error
- **AND** when `Cause` is nil, `ErrorType` SHALL be the empty string

#### Scenario: Cause is omitempty in JSON output

- **WHEN** `BatchFailureInfo` is serialized to JSON via `output.NewJSONExporter`
- **THEN** a `Cause` field with a nil value SHALL NOT appear in the JSON output
- **AND** a `Cause` field with a non-nil `*core.*Error` value SHALL appear as `{"error": "<Error()>"}`

#### Scenario: Cause MUST be a typed error that implements `json.Marshaler`

- **WHEN** a `BatchFailureInfo.Cause` is assigned at any of the 5 population sites in `internal/library/adder.go`
- **THEN** the cause SHALL be a `*core.*Error` typed error (or any other type that implements `json.Marshaler`)
- **AND** non-typed causes (e.g., `*os.PathError`, plain `errors.New(...)`) SHALL be wrapped in `*core.FileError` at the population site BEFORE assignment to `f.Cause`

#### Scenario: Cause supports errors.Is / errors.As

- **WHEN** a downstream caller calls `errors.Is(failure.Cause, sentinel)` or `errors.As(failure.Cause, &target)`
- **THEN** the call SHALL succeed against the original typed error
- **AND** the call SHALL match the chain if the original error was wrapped via `fmt.Errorf("...: %w", inner)`

### Requirement: BatchFailureInfo JSON wire-format compatibility

The additive `ErrorType` and `Cause` fields SHALL preserve backward compatibility with JSON consumers that parse the legacy 2-field shape (`source`, `error`).

#### Scenario: Legacy consumer sees no breaking change

- **WHEN** a downstream JSON consumer reads the `source` and `error` fields from a serialized `BatchFailureInfo`
- **THEN** the consumer SHALL observe the same values as before the change
- **AND** the consumer SHALL ignore the new `errorType` and `cause` fields

# typed-errors Specification

## Purpose

Define domain-specific error types with structured fields for parse, validation, transform, file, config, not-found, operation, and partial-success conditions in germinator. All error types live in `internal/core/errors.go` and follow an immutable builder pattern: private fields, accessor methods, fluent `WithSuggestions`/`WithContext` builders.

## Requirements

### Requirement: ParseError type

The system SHALL provide a `ParseError` type for document parsing failures with private fields and accessor methods.

#### Scenario: ParseError has private fields

- **WHEN** `ParseError` is examined
- **THEN** it SHALL have private fields: `path`, `message`, `cause`, `suggestions`, `context`
- **AND** these fields SHALL NOT be directly accessible from outside `internal/core`

#### Scenario: ParseError getters return field values

- **WHEN** `NewParseError("agent.md", "missing delimiter", cause)` is called
- **THEN** `err.Path()` SHALL return `"agent.md"`
- **AND** `err.Message()` SHALL return `"missing delimiter"`
- **AND** `err.Cause()` SHALL return the cause error (may be nil)
- **AND** `err.Suggestions()` SHALL return a copy of the suggestions slice (empty slice when unset)
- **AND** `err.Context()` SHALL return an empty string when unset

#### Scenario: ParseError Unwrap exposes cause

- **WHEN** `errors.Unwrap(err)` is called on a `ParseError` with a non-nil cause
- **THEN** it SHALL return the cause error

### Requirement: ValidationError type

The system SHALL provide a `ValidationError` type for document validation failures with private fields, four-argument constructor, and getter methods.

#### Scenario: ValidationError has private fields

- **WHEN** `ValidationError` is examined
- **THEN** it SHALL have private fields: `request`, `field`, `value`, `message`, `suggestions`, `context`

#### Scenario: NewValidationError constructor takes four parameters

- **WHEN** `NewValidationError(request, field, value, message string)` is called
- **THEN** it SHALL return a `*ValidationError` with `request`, `field`, `value`, and `message` populated
- **AND** suggestions SHALL be empty
- **AND** context SHALL be empty
- **AND** the old three-parameter signature `NewValidationError(message, field, suggestions)` SHALL NOT exist

#### Scenario: ValidationError getters return field values

- **WHEN** `err := NewValidationError("Agent", "name", "invalid", "name is required")` is called
- **THEN** `err.Request()` SHALL return `"Agent"`
- **AND** `err.Field()` SHALL return `"name"`
- **AND** `err.Value()` SHALL return `"invalid"`
- **AND** `err.Message()` SHALL return `"name is required"`
- **AND** `err.Suggestions()` SHALL return an empty slice
- **AND** `err.Context()` SHALL return an empty string

#### Scenario: ValidationError Error format includes request and field

- **WHEN** `err.Error()` is called on a `ValidationError` with request `"Agent"` and field `"name"`
- **THEN** the rendered string SHALL contain `validation failed for Agent.name`
- **AND** the rendered string SHALL contain the message

### Requirement: TransformError type

The system SHALL provide a `TransformError` type for transformation pipeline failures with private fields and accessor methods.

#### Scenario: TransformError has private fields

- **WHEN** `TransformError` is examined
- **THEN** it SHALL have private fields: `operation`, `platform`, `message`, `cause`, `suggestions`, `context`

#### Scenario: NewTransformError constructor takes four parameters

- **WHEN** `NewTransformError(operation, platform, message string, cause error)` is called
- **THEN** it SHALL return a `*TransformError` with all four parameters populated

### Requirement: FileError type

The system SHALL provide a `FileError` type for file I/O failures with private fields and accessor methods.

#### Scenario: FileError has private fields

- **WHEN** `FileError` is examined
- **THEN** it SHALL have private fields: `path`, `operation`, `message`, `cause`, `suggestions`, `context`

#### Scenario: NewFileError constructor takes four parameters

- **WHEN** `NewFileError(path, operation, message string, cause error)` is called
- **THEN** it SHALL return a `*FileError` with path, operation, message, and cause populated

#### Scenario: FileError IsNotFound helper

- **WHEN** `IsNotFound()` is called on a `FileError` whose message or wrapped cause contains `"not found"`, `"does not exist"`, or `"no such file"` (case-insensitive)
- **THEN** it SHALL return `true`
- **AND** otherwise it SHALL return `false`

### Requirement: ConfigError type

The system SHALL provide a `ConfigError` type for configuration and CLI errors. The constructor takes three parameters; valid options are added via the `WithSuggestions` builder.

#### Scenario: ConfigError has private fields

- **WHEN** `ConfigError` is examined
- **THEN** it SHALL have private fields: `field`, `value`, `message`, `suggestions`, `context`
- **AND** the field SHALL be named `suggestions` (replacing the legacy `available` field)

#### Scenario: NewConfigError constructor takes three parameters

- **WHEN** `NewConfigError(field, value, message string)` is called
- **THEN** it SHALL return a `*ConfigError` with `field`, `value`, and `message` populated
- **AND** suggestions SHALL be empty
- **AND** the old four-parameter signature `NewConfigError(field, value, available, message)` SHALL NOT exist

#### Scenario: Available options added via WithSuggestions

- **WHEN** a config error needs to show valid options
- **THEN** the code SHALL call `NewConfigError(field, value, message).WithSuggestions([]string{...})`
- **AND** the constructor SHALL NOT accept an `available` parameter

### Requirement: NotFoundError type

The system SHALL provide a `NotFoundError` type for missing-entity lookups (library refs, presets). It carries `Entity` and `Key` as exported fields and maps to exit code 2 (usage) via `cmdutil.ExitCodeFor`.

#### Scenario: NewNotFoundError constructor

- **WHEN** `NewNotFoundError(entity, key string)` is called
- **THEN** it SHALL return a `*NotFoundError{Entity: entity, Key: key}`

#### Scenario: NotFoundError Error format

- **WHEN** `err.Error()` is called on a `*NotFoundError{Key: "ghost"}`
- **THEN** it SHALL return the string `"not found: ghost"`

#### Scenario: NotFoundError maps to ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

### Requirement: OperationError type

The system SHALL provide an `OperationError` type for per-operation failures (library orphan discovery, file operations) that wraps an optional cause and exposes `Op`, `Resource`, and `Cause` as exported fields.

#### Scenario: NewOperationError constructor

- **WHEN** `NewOperationError(op, resource string, cause error)` is called
- **THEN** it SHALL return a `*OperationError{Op: op, Resource: resource, Cause: cause}`

#### Scenario: OperationError Unwrap exposes cause

- **WHEN** `errors.Unwrap(err)` is called on an `OperationError` with a non-nil cause
- **THEN** it SHALL return the cause error

### Requirement: InitializeError type

The system SHALL provide an `InitializeError` type for per-resource installation failures with private fields, builder methods, and accessor methods.

#### Scenario: InitializeError has private fields

- **WHEN** `InitializeError` is examined
- **THEN** it SHALL have private fields: `ref`, `inputPath`, `outputPath`, `cause`, `suggestions`, `context`

#### Scenario: NewInitializeError constructor

- **WHEN** `NewInitializeError(ref, inputPath, outputPath string, cause error)` is called
- **THEN** it SHALL return a `*InitializeError` with all four parameters populated

### Requirement: PartialSuccessError type

The system SHALL provide a `PartialSuccessError` type for aggregated batch outcomes where some operations succeed and others fail.

#### Scenario: NewPartialSuccessError constructor

- **WHEN** `NewPartialSuccessError(succeeded, failed int, errs []InitializeError)` is called
- **THEN** it SHALL return a `*PartialSuccessError` with `succeeded`, `failed`, and `errors` populated
- **AND** the aggregate `Error()` format SHALL be `"partial success: N succeeded, M failed"`

#### Scenario: PartialSuccessError exit-code semantics

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*PartialSuccessError{Succeeded: 3, Failed: 1}`
- **THEN** it SHALL return `ExitCodeSuccess` (0)
- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*PartialSuccessError{Succeeded: 0, Failed: N}`
- **THEN** it SHALL return `ExitCodeError` (1)

### Requirement: Immutable builder pattern

All error types' `WithSuggestions` and `WithContext` methods SHALL be immutable builders that return new instances without modifying the original.

#### Scenario: WithSuggestions returns new instance

- **WHEN** `err2 := err1.WithSuggestions([]string{"hint"})` is called
- **THEN** `err2` SHALL be a new error instance with suggestions populated
- **AND** `err1` SHALL remain unchanged (its suggestions remain empty)

#### Scenario: WithSuggestions chains with WithContext

- **WHEN** `NewParseError(p, m, c).WithSuggestions(s).WithContext("ctx")` is called
- **THEN** the chain SHALL return a new error with both suggestions and context set
- **AND** all builders SHALL be chainable

### Requirement: Suggestions getter returns a copy

All error types' `Suggestions()` getter SHALL return a copy of the underlying slice, not the original, so external mutation cannot corrupt the error state.

#### Scenario: Suggestions copy is independent

- **WHEN** `s := err.Suggestions()` is called and the caller mutates `s[0]`
- **THEN** the original error's suggestions SHALL remain unchanged

### Requirement: Error wrapping support

All error types SHALL support Go's error-wrapping conventions via `Unwrap()` and shall be detectable through `errors.As` and `errors.Is`.

#### Scenario: errors.As matches typed errors

- **WHEN** `errors.As(err, &target)` is called with `var target *core.ParseError`
- **THEN** it SHALL return `true` when `err` is or wraps a `*core.ParseError`
- **AND** `target` SHALL receive the unwrapped `*core.ParseError` value

#### Scenario: errors.Is traverses wrap chain

- **WHEN** a typed error wraps a sentinel via `fmt.Errorf("...: %w", sentinel)` and `errors.Is(err, sentinel)` is called
- **THEN** it SHALL return `true`

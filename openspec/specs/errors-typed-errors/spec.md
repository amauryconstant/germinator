# typed-errors Specification

> **Cross-references:** this change also modifies `cli-error-formatting`, `cli-exit-codes`, `errors-enhanced-errors`. See those delta specs.

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

The system SHALL provide a `NotFoundError` type for missing-entity lookups (library refs, presets, library.yaml, source files, library files). It carries `Entity` and `Key` as exported fields and maps to exit code 1 (operational error) via `cmdutil.ExitCodeFor`.

**Change**: clarify that `NotFoundError` maps to `ExitCodeError` (1), not `ExitCodeUsage` (2). The prior mapping (2) was semantically wrong: "not found" is a runtime state, not a user-input validation error.

#### Scenario: NewNotFoundError constructor

- **WHEN** `NewNotFoundError(entity, key string)` is called
- **THEN** it SHALL return a `*NotFoundError{Entity: entity, Key: key}`

#### Scenario: NotFoundError Error format

- **WHEN** `err.Error()` is called on a `*NotFoundError{Key: "ghost"}`
- **THEN** it SHALL return the string `"not found: ghost"`

#### Scenario: NotFoundError maps to ExitCodeError

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeError` (1) — `*core.NotFoundError` represents a runtime lookup miss, not a user-input validation error.

#### Scenario: NotFoundError implements json.Marshaler

- **WHEN** `json.Marshal(*core.NotFoundError{Key: "ghost"})` is called
- **THEN** it SHALL return the JSON bytes `{"error": "not found: ghost"}`

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

### Requirement: UsageError type

The system SHALL provide a `UsageError` type for CLI flag validation errors that are not caught by Cobra's `MarkFlagRequired`, `Args` validators, or pflag's typed errors. `UsageError` carries the flag name, the reason, and optional suggestions as private fields exposed via accessor methods (`Flag()`, `Reason()`, `Suggestions()`) following the project's builder pattern.

`UsageError` maps to `ExitCodeUsage` (2) via `cmdutil.ExitCodeFor`. `Unwrap()` returns `nil` — `UsageError` is a leaf error.

#### Scenario: UsageError has private Flag, Reason, and Suggestions fields

- **WHEN** `NewUsageError(flag, reason string)` is called
- **THEN** it SHALL return a `*UsageError` with private `flag`, `reason`, and (initially nil) `suggestions` fields populated
- **AND** `e.Flag()` SHALL return the flag
- **AND** `e.Reason()` SHALL return the reason
- **AND** `e.Suggestions()` SHALL return a defensive copy of the suggestions slice (or nil if no suggestions are set)

#### Scenario: UsageError Error format

- **WHEN** `err.Error()` is called on a `*UsageError{flag: "--resources", reason: "must be non-empty list of refs"}`
- **THEN** it SHALL return the string `"--resources: must be non-empty list of refs"`

#### Scenario: UsageError follows Go error-string convention

- **WHEN** a `*UsageError` is constructed via `NewUsageError(flag, reason)` for any input
- **THEN** the rendered `Error()` SHALL be a single line in the form `<flag>: <reason>`, where `<flag>` starts with `--` and the rest of the flag segment is lowercase kebab-case, and `<reason>` starts with a lowercase letter, contains no trailing `.`, `!`, or `?`
- **AND** the godoc for `*UsageError` SHALL explicitly state the convention.

#### Scenario: UsageError WithSuggestions builder returns a new instance

- **WHEN** `e.WithSuggestions([]string{"hint1", "hint2"})` is called on a `*UsageError`
- **THEN** the returned `*UsageError` SHALL have the same `flag` and `reason` as `e`
- **AND** the returned `*UsageError`'s `Suggestions()` SHALL return `[]string{"hint1", "hint2"}`
- **AND** the original `e` SHALL NOT be modified (immutable builder)

#### Scenario: UsageError is in FormatError dispatch set

- **WHEN** `output.FormatError(io, err)` is called with `*core.UsageError`
- **THEN** the rendered message SHALL be `"Error: --resources: must be non-empty list of refs"` written to `io.ErrOut`

#### Scenario: UsageError implements json.Marshaler

- **WHEN** `json.Marshal(*core.UsageError{...})` is called
- **THEN** it SHALL return the JSON bytes `{"error": "--resources: must be non-empty list of refs"}`

### Requirement: CobraUsageError sentinel

The system SHALL provide a `CobraUsageError` sentinel that wraps the underlying Cobra arg-validation error. Commands wrap the error returned by `cobra.ExactArgs`/`MinimumNArgs`/`MaximumNArgs`/`RangeArgs` failures (currently emitted as `fmt.Errorf` strings via `cobra/args.go`) with `MustNewCobraUsageError(err)` so `cmdutil.ExitCodeFor` can match the typed error and return `ExitCodeUsage` (2). The `Must*` prefix telegraphs the panic on nil cause.

#### Scenario: CobraUsageError wraps an existing error

- **WHEN** `MustNewCobraUsageError(err)` is called with a non-nil `err`
- **THEN** the returned `*CobraUsageError` SHALL expose the underlying error via `Unwrap()`
- **AND** the constructor SHALL NOT panic (the cause is non-nil)

#### Scenario: CobraUsageError Error format

- **WHEN** `err.Error()` is called on a `*CobraUsageError` wrapping `errors.New("requires at least 1 arg(s), only received 0")`
- **THEN** it SHALL return the wrapped error's `Error()` string verbatim

#### Scenario: MustNewCobraUsageError panics on nil cause

- **WHEN** `MustNewCobraUsageError(nil)` is called
- **THEN** it SHALL panic with a message indicating the cause is required — a nil cause is a programmer error, not a recoverable state.

### Requirement: ValidateDocumentType helper

The system SHALL provide a `ValidateDocumentType(docType string) error` helper in `internal/core/rules.go` that validates a bare document type against the canonical resource-type set `{skill, agent, command, memory}`.

`CanInstallResource(ref)` validates the `"type/name"` ref shape and is the WRONG guardrail for bare document types; `ValidateDocumentType` is the new sibling helper for the bare-type case.

#### Scenario: ValidateDocumentType accepts canonical types

- **WHEN** `core.ValidateDocumentType("agent")` is called (and similarly for `"command"`, `"skill"`, `"memory"`)
- **THEN** it SHALL return nil

#### Scenario: ValidateDocumentType rejects the plural form

- **WHEN** `core.ValidateDocumentType("skills")` is called (the plural form)
- **THEN** it SHALL return a `*core.ValidationError`
- **AND** the error SHALL include a suggestion listing the canonical types

#### Scenario: ValidateDocumentType rejects unknown / empty input

- **WHEN** `core.ValidateDocumentType("bot")` or `core.ValidateDocumentType("")` is called
- **THEN** it SHALL return a `*core.ValidationError`

### Requirement: MarshalJSON on all core typed errors

All typed errors defined in `internal/core/errors.go` SHALL implement `MarshalJSON() ([]byte, error)`. The complete set is:

- **Existing (9)**: `ParseError`, `ValidationError`, `TransformError`, `FileError`, `ConfigError`, `NotFoundError`, `OperationError`, `InitializeError`, `PartialSuccessError`.
- **New (2)**: `UsageError`, `CobraUsageError`.

Each `MarshalJSON()` SHALL return the JSON bytes `{"error": "<Error()>"}`.

#### Scenario: MarshalJSON returns structured JSON

- **WHEN** `json.Marshal(*core.NotFoundError{Key: "ghost"})` is called
- **THEN** it SHALL return the JSON bytes `{"error": "not found: ghost"}`
- **AND** the underlying error's `Error()` string SHALL be the value of the `error` field

#### Scenario: MarshalJSON for wrapped typed errors

- **WHEN** `json.Marshal(&core.OperationError{Op: "add", Resource: "skill/commit", Cause: ...})` is called
- **THEN** the rendered JSON SHALL be `{"error": "add: skill/commit"}` (the single-key shape, delegating to `e.Error()`)
- **AND** cause chains SHALL NOT be re-encoded recursively.

#### Scenario: MarshalJSON contract for all 11 typed errors

- **WHEN** `json.Marshal(e)` is called for each of the 11 typed errors (ParseError, ValidationError, TransformError, FileError, ConfigError, NotFoundError, OperationError, InitializeError, PartialSuccessError, UsageError, CobraUsageError)
- **THEN** the rendered JSON SHALL be `{"error": "<e.Error()>"}`
- **AND** the underlying error's `Error()` string SHALL be the value of the `error` field

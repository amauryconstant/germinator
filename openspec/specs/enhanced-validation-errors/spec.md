## Purpose

Replace the existing ValidationError with an enhanced version that supports immutable builders for fluent error construction, private fields with getters, and better context/suggestion support.

## Requirements

### Requirement: ValidationError with private fields

The system SHALL provide an enhanced `ValidationError` in `internal/errors/types.go` with private fields.

#### Scenario: ValidationError has private fields

- **WHEN** ValidationError is examined
- **THEN** it SHALL have private fields: `request`, `field`, `value`, `message`, `suggestions`, `context`
- **AND** these fields SHALL NOT be directly accessible

---

### Requirement: NewValidationError constructor

The system SHALL provide a `NewValidationError(request, field, value, message string) *ValidationError` function.

#### Scenario: NewValidationError creates error with all params

- **WHEN** `NewValidationError("Agent", "name", "invalid", "name is required")` is called
- **THEN** the returned ValidationError SHALL have request="Agent", field="name", value="invalid", message="name is required"
- **AND** suggestions SHALL be empty
- **AND** context SHALL be empty

---

### Requirement: WithSuggestions immutable builder

The system SHALL provide a `WithSuggestions(suggestions []string) *ValidationError` method.

#### Scenario: WithSuggestions returns new instance

- **WHEN** `err.WithSuggestions([]string{"try this"})` is called
- **THEN** it SHALL return a new ValidationError (not modify original)
- **AND** the new error SHALL have the suggestions
- **AND** the original error SHALL remain unchanged

#### Scenario: WithSuggestions chains

- **WHEN** multiple WithSuggestions calls are chained
- **THEN** each SHALL return a new instance

---

### Requirement: WithContext immutable builder

The system SHALL provide a `WithContext(context string) *ValidationError` method.

#### Scenario: WithContext returns new instance

- **WHEN** `err.WithContext("additional info")` is called
- **THEN** it SHALL return a new ValidationError (not modify original)
- **AND** the new error SHALL have the context

---

### Requirement: ValidationError getter methods

The system SHALL provide getter methods for all ValidationError fields.

#### Scenario: Field getter returns field name

- **WHEN** `err.Field()` is called
- **THEN** it SHALL return the field value

#### Scenario: Value getter returns invalid value

- **WHEN** `err.Value()` is called
- **THEN** it SHALL return the value that failed validation

#### Scenario: Message getter returns error message

- **WHEN** `err.Message()` is called
- **THEN** it SHALL return the validation message

#### Scenario: Request getter returns request context

- **WHEN** `err.Request()` is called
- **THEN** it SHALL return the request type name

#### Scenario: Suggestions getter returns copy

- **WHEN** `err.Suggestions()` is called
- **THEN** it SHALL return a copy of the suggestions slice (not the original)

#### Scenario: Context getter returns context

- **WHEN** `err.Context()` is called
- **THEN** it SHALL return the context string

---

### Requirement: ValidationError Error() method

The system SHALL provide an `Error() string` method that formats the error.

#### Scenario: Error includes request and field

- **WHEN** `err.Error()` is called on a ValidationError with request="Agent", field="name", value="invalid"
- **THEN** the string SHALL contain "validation failed for Agent.name"
- **AND** the string SHALL contain the message
- **AND** the string SHALL contain the value

#### Scenario: Error includes suggestions

- **WHEN** `err.Error()` is called on a ValidationError with suggestions
- **THEN** the string SHALL contain "💡 " followed by each suggestion

---

### Requirement: Remove old NewValidationError signature

The system SHALL remove the old `NewValidationError(message, field string, suggestions []string)` function.

#### Scenario: Old constructor no longer exists

- **WHEN** code tries to use the old 3-parameter NewValidationError
- **THEN** it SHALL fail to compile

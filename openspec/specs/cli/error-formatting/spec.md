# error-formatting Specification

## Purpose

Provide composable error formatting with type-specific output and contextual hints for better user experience.

## Requirements

### Requirement: Error Formatter Interface

The system SHALL provide an ErrorFormatter for consistent error output.

#### Scenario: Format any error

- **WHEN** ErrorFormatter.Format receives any error
- **THEN** it SHALL return a formatted string for stderr output
- **AND** it SHALL include "Error:" prefix
- **AND** it SHALL end with a newline

#### Scenario: Format unknown error type

- **WHEN** ErrorFormatter.Format receives an unregistered error type
- **THEN** it SHALL use default formatting
- **AND** default formatting SHALL be "Error: <error message>"

---

### Requirement: Type-Specific Formatting

The system SHALL register type-specific formatters for each error type.

#### Scenario: Format ParseError

- **WHEN** formatting a ParseError
- **THEN** output SHALL include "Parse error:" prefix
- **AND** output SHALL include the file path
- **AND** output SHALL include the descriptive message

#### Scenario: Format ValidationError with suggestions

- **WHEN** formatting a ValidationError with suggestions
- **THEN** output SHALL include "Validation error:" prefix
- **AND** output SHALL include the validation message
- **AND** output SHALL include "Hint:" prefixed suggestions

#### Scenario: Format ValidationError without suggestions

- **WHEN** formatting a ValidationError without suggestions
- **THEN** output SHALL include the validation message
- **AND** output SHALL NOT include "Hint:" lines

#### Scenario: Format TransformError

- **WHEN** formatting a TransformError
- **THEN** output SHALL include "Transform error:" prefix
- **AND** output SHALL include the operation that failed
- **AND** output SHALL include the platform if available

#### Scenario: Format FileError

- **WHEN** formatting a FileError
- **THEN** output SHALL include "File error:" prefix
- **AND** output SHALL include the operation (read/write)
- **AND** output SHALL include the file path

#### Scenario: Format ConfigError

- **WHEN** formatting a ConfigError
- **THEN** output SHALL include "Config error:" prefix
- **AND** output SHALL include available options if applicable

---

### Requirement: Formatter Registration

The system SHALL allow registration of custom formatters.

#### Scenario: Register formatter for type

- **WHEN** RegisterFormatter is called with a type and function
- **THEN** the formatter SHALL be stored for that type
- **AND** Format SHALL use the registered formatter for matching errors

#### Scenario: Formatter matches wrapped errors

- **WHEN** formatting a wrapped typed error
- **THEN** the formatter SHALL detect the type using errors.As
- **AND** the type-specific formatter SHALL be used

---

### Requirement: Multiple Validation Errors

The system SHALL format multiple validation errors clearly.

#### Scenario: Format validation error list

- **WHEN** multiple validation errors exist
- **THEN** each error SHALL be formatted on a separate line
- **AND** each error SHALL be numbered or bulleted
- **AND** suggestions SHALL appear after their associated error

---

### Requirement: Error Cause Chain

The system SHALL optionally include error cause in output.

#### Scenario: Include cause for debugging

- **WHEN** an error has an underlying cause
- **THEN** formatted output MAY include cause details
- **AND** cause SHALL be indented or clearly separated

---

### Requirement: Error Formatter Constructor

The system SHALL provide a constructor for ErrorFormatter.

#### Scenario: NewErrorFormatter creates formatter

- **WHEN** NewErrorFormatter() is called
- **THEN** it SHALL return an ErrorFormatter with all type formatters registered
- **AND** the formatter SHALL be ready to use immediately

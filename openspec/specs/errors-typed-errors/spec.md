# typed-errors Specification

## Purpose

Define domain-specific error types with structured fields for parse, validation, transform, file, and config errors in germinator.

## Requirements

### Requirement: Parse Error Type

The system SHALL provide a ParseError type for document parsing failures.

#### Scenario: ParseError with invalid YAML

- **WHEN** YAML parsing fails for a document
- **THEN** ParseError SHALL contain the file path
- **AND** ParseError SHALL contain a descriptive message
- **AND** ParseError SHALL wrap the underlying cause
- **AND** ParseError.Error() SHALL return a formatted string

#### Scenario: ParseError with unrecognized document type

- **WHEN** document type cannot be detected from filename
- **THEN** ParseError SHALL contain the file path
- **AND** ParseError message SHALL list expected patterns
- **AND** ParseError.Unwrap() SHALL return nil

---

### Requirement: Validation Error Type

The system SHALL provide a ValidationError type for document validation failures.

#### Scenario: ValidationError with field errors

- **WHEN** document validation fails
- **THEN** ValidationError SHALL contain a descriptive message
- **AND** ValidationError SHALL contain optional field name
- **AND** ValidationError SHALL support suggestions list
- **AND** ValidationError.Error() SHALL return formatted message

#### Scenario: ValidationError with suggestions

- **WHEN** ValidationError has suggestions
- **THEN** Suggestions() SHALL return list of hint strings
- **AND** each suggestion SHALL be actionable guidance

---

### Requirement: Transform Error Type

The system SHALL provide a TransformError type for transformation pipeline failures.

#### Scenario: TransformError with template failure

- **WHEN** template rendering fails
- **THEN** TransformError SHALL contain the template name
- **AND** TransformError SHALL contain a descriptive message
- **AND** TransformError SHALL wrap the underlying cause

#### Scenario: TransformError with platform conversion failure

- **WHEN** platform-specific conversion fails
- **THEN** TransformError SHALL contain the platform name
- **AND** TransformError SHALL contain the operation that failed

---

### Requirement: File Error Type

The system SHALL provide a FileError type for file I/O failures.

#### Scenario: FileError with read failure

- **WHEN** file read fails
- **THEN** FileError SHALL contain the file path
- **AND** FileError SHALL contain the operation ("read")
- **AND** FileError SHALL wrap the underlying cause

#### Scenario: FileError with write failure

- **WHEN** file write fails
- **THEN** FileError SHALL contain the file path
- **AND** FileError SHALL contain the operation ("write")
- **AND** FileError SHALL wrap the underlying cause

#### Scenario: FileError IsNotFound helper

- **WHEN** FileError represents a file not found condition
- **THEN** IsNotFound() SHALL return true
- **AND** detection SHALL check for "not found" or "does not exist" in message

---

### Requirement: Config Error Type

The system SHALL provide a ConfigError type for configuration and CLI errors.

#### Scenario: ConfigError with invalid platform

- **WHEN** an invalid platform is specified
- **THEN** ConfigError SHALL contain the invalid value
- **AND** ConfigError SHALL contain available options
- **AND** ConfigError message SHALL list valid platforms

#### Scenario: ConfigError with missing required flag

- **WHEN** a required flag is missing
- **THEN** ConfigError SHALL contain the flag name
- **AND** ConfigError SHALL be categorized as usage error

---

### Requirement: Error Wrapping Support

All error types SHALL support Go's error wrapping conventions.

#### Scenario: errors.As for typed errors

- **WHEN** checking error type with errors.As
- **THEN** typed errors SHALL be correctly matched
- **AND** the target pointer SHALL receive the error value

#### Scenario: errors.Is for error comparison

- **WHEN** wrapped errors are compared with errors.Is
- **THEN** comparison SHALL work through the wrap chain

---

### Requirement: Error Constructor Functions

The system SHALL provide constructor functions for each error type.

#### Scenario: NewParseError constructor

- **WHEN** NewParseError(path, message, cause) is called
- **THEN** it SHALL return a ParseError with all fields populated

#### Scenario: NewValidationError constructor

- **WHEN** NewValidationError(message, field, suggestions) is called
- **THEN** it SHALL return a ValidationError with all fields populated

#### Scenario: NewTransformError constructor

- **WHEN** NewTransformError(operation, platform, message, cause) is called
- **THEN** it SHALL return a TransformError with all fields populated

#### Scenario: NewFileError constructor

- **WHEN** NewFileError(path, operation, message, cause) is called
- **THEN** it SHALL return a FileError with all fields populated

#### Scenario: NewConfigError constructor

- **WHEN** NewConfigError(field, value, available, message) is called
- **THEN** it SHALL return a ConfigError with all fields populated

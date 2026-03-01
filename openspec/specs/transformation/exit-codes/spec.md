# exit-codes Specification

## Purpose

Define semantic exit codes for germinator CLI to enable programmatic error handling by scripts and tools.

## Requirements

### Requirement: Exit Code Constants

The system SHALL define semantic exit code constants.

#### Scenario: Exit code values

- **WHEN** exit codes are defined
- **THEN** ExitCodeSuccess SHALL be 0
- **AND** ExitCodeError SHALL be 1
- **AND** ExitCodeUsage SHALL be 2
- **AND** ExitCodeParse SHALL be 3

---

### Requirement: Error Categories

The system SHALL define error categories for exit code mapping.

#### Scenario: Category enumeration

- **WHEN** error categories are defined
- **THEN** CategoryCobra SHALL exist for Cobra framework errors
- **AND** CategoryConfig SHALL exist for configuration errors
- **AND** CategoryParse SHALL exist for parsing errors
- **AND** CategoryValidation SHALL exist for validation errors
- **AND** CategoryTransform SHALL exist for transformation errors
- **AND** CategoryFile SHALL exist for file I/O errors
- **AND** CategoryGeneric SHALL exist for unclassified errors

---

### Requirement: Error Categorization Function

The system SHALL provide a function to categorize errors.

#### Scenario: Categorize ParseError

- **WHEN** CategorizeError receives a ParseError
- **THEN** it SHALL return CategoryParse

#### Scenario: Categorize ValidationError

- **WHEN** CategorizeError receives a ValidationError
- **THEN** it SHALL return CategoryValidation

#### Scenario: Categorize TransformError

- **WHEN** CategorizeError receives a TransformError
- **THEN** it SHALL return CategoryTransform

#### Scenario: Categorize FileError

- **WHEN** CategorizeError receives a FileError
- **THEN** it SHALL return CategoryFile

#### Scenario: Categorize ConfigError

- **WHEN** CategorizeError receives a ConfigError
- **THEN** it SHALL return CategoryConfig

#### Scenario: Categorize wrapped errors

- **WHEN** CategorizeError receives a wrapped typed error
- **THEN** it SHALL detect the underlying type using errors.As
- **AND** it SHALL return the correct category

#### Scenario: Categorize unknown error

- **WHEN** CategorizeError receives an unclassified error
- **THEN** it SHALL return CategoryGeneric

---

### Requirement: Exit Code Mapping

The system SHALL provide a function to map errors to exit codes.

#### Scenario: Parse error exit code

- **WHEN** GetExitCodeForError receives CategoryParse
- **THEN** it SHALL return ExitCodeParse (3)

#### Scenario: Config error exit code

- **WHEN** GetExitCodeForError receives CategoryConfig
- **THEN** it SHALL return ExitCodeUsage (2)

#### Scenario: Validation error exit code

- **WHEN** GetExitCodeForError receives CategoryValidation
- **THEN** it SHALL return ExitCodeUsage (2)

#### Scenario: Transform error exit code

- **WHEN** GetExitCodeForError receives CategoryTransform
- **THEN** it SHALL return ExitCodeError (1)

#### Scenario: File error exit code

- **WHEN** GetExitCodeForError receives CategoryFile
- **THEN** it SHALL return ExitCodeError (1)

#### Scenario: Cobra error exit code

- **WHEN** GetExitCodeForError receives CategoryCobra
- **THEN** it SHALL return ExitCodeUsage (2)

#### Scenario: Generic error exit code

- **WHEN** GetExitCodeForError receives CategoryGeneric
- **THEN** it SHALL return ExitCodeError (1)

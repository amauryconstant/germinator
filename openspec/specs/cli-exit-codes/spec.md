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
- **AND** ExitCodeConfig SHALL be 3
- **AND** ExitCodeGit SHALL be 4
- **AND** ExitCodeValidation SHALL be 5
- **AND** ExitCodeNotFound SHALL be 6

---

### Requirement: Error Categories

The system SHALL define error categories for exit code mapping.

#### Scenario: Category enumeration

- **WHEN** error categories are defined
- **THEN** CategoryCobra SHALL exist for Cobra framework errors
- **AND** CategoryConfig SHALL exist for configuration errors (renamed from CategoryParse)
- **AND** CategoryValidation SHALL exist for validation errors
- **AND** CategoryNotFound SHALL exist for not-found errors (NEW)
- **AND** CategoryGit SHALL exist for git-related errors (NEW)
- **AND** CategoryGeneric SHALL exist for unclassified errors

#### Scenario: CategoryParse renamed to CategoryConfig

- **GIVEN** current code has `CategoryParse` for parse errors
- **WHEN** the migration is complete
- **THEN** `CategoryParse` SHALL be renamed to `CategoryConfig`
- **AND** parse errors SHALL map to the Config exit code (3)

---

### Requirement: Error Categorization Function

The system SHALL provide a function to categorize errors.

#### Scenario: Categorize ParseError

- **WHEN** CategorizeError receives a ParseError
- **THEN** it SHALL return CategoryConfig (renamed from CategoryParse)

#### Scenario: Categorize ValidationError

- **WHEN** CategorizeError receives a ValidationError
- **THEN** it SHALL return CategoryValidation

#### Scenario: Categorize ConfigError

- **WHEN** CategorizeError receives a ConfigError
- **THEN** it SHALL return CategoryConfig

#### Scenario: Categorize GitError

- **WHEN** CategorizeError receives a GitError
- **THEN** it SHALL return CategoryGit

#### Scenario: Categorize NotFoundError

- **WHEN** CategorizeError receives a NotFoundError
- **THEN** it SHALL return CategoryNotFound

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

#### Scenario: Config error exit code

- **WHEN** GetExitCodeForError receives CategoryConfig
- **THEN** it SHALL return ExitCodeConfig (3)

#### Scenario: Validation error exit code

- **WHEN** GetExitCodeForError receives CategoryValidation
- **THEN** it SHALL return ExitCodeValidation (5)

#### Scenario: NotFound error exit code

- **WHEN** GetExitCodeForError receives CategoryNotFound
- **THEN** it SHALL return ExitCodeNotFound (6)

#### Scenario: Git error exit code

- **WHEN** GetExitCodeForError receives CategoryGit
- **THEN** it SHALL return ExitCodeGit (4)

#### Scenario: Cobra error exit code

- **WHEN** GetExitCodeForError receives CategoryCobra
- **THEN** it SHALL return ExitCodeUsage (2)

#### Scenario: Generic error exit code

- **WHEN** GetExitCodeForError receives CategoryGeneric
- **THEN** it SHALL return ExitCodeError (1)

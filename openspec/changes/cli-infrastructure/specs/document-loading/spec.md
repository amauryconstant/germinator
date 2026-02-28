# document-loading Specification (Delta)

## Purpose

Load and validate documents from files, detecting type from filename patterns.

## MODIFIED Requirements

### Requirement: Filename-Based Type Detection

The system SHALL detect document type from filename patterns only.

#### Scenario: Agent detected from agent- prefix

- **GIVEN** a file named "agent-coder.md"
- **WHEN** LoadDocument analyzes the filename
- **THEN** it SHALL detect type as "agent"

#### Scenario: Agent detected from -agent suffix

- **GIVEN** a file named "code-review-agent.md"
- **WHEN** LoadDocument analyzes the filename
- **THEN** it SHALL detect type as "agent"

#### Scenario: Command detected from command- prefix or -command suffix

- **GIVEN** a file named "command-test.md" or "lint-command.md"
- **WHEN** LoadDocument analyzes the filename
- **THEN** it SHALL detect type as "command"

#### Scenario: Memory detected from memory- prefix or -memory suffix

- **GIVEN** a file named "memory-go.md" or "api-memory.md"
- **WHEN** LoadDocument analyzes the filename
- **THEN** it SHALL detect type as "memory"

#### Scenario: Skill detected from skill- prefix or -skill suffix

- **GIVEN** a file named "skill-review.md" or "refactor-skill.md"
- **WHEN** LoadDocument analyzes the filename
- **THEN** it SHALL detect type as "skill"

#### Scenario: Unrecognizable filename returns typed error

- **GIVEN** a file named "my-document.md"
- **WHEN** LoadDocument analyzes the filename
- **THEN** it SHALL return a ParseError
- **AND** ParseError.Path SHALL be the filepath
- **AND** ParseError.Message SHALL list expected filename patterns
- **AND** ParseError.Unwrap() SHALL return nil

---

### Requirement: Document Loading Workflow

LoadDocument SHALL combine type detection, parsing, and validation in one function.

#### Scenario: LoadDocument parses and validates valid document

- **GIVEN** a valid Agent document file
- **WHEN** LoadDocument is called with the filepath
- **THEN** it SHALL detect document type from filename
- **AND** it SHALL call ParseDocument with detected type
- **AND** it SHALL call document.Validate() on the parsed document
- **AND** it SHALL return the validated document

#### Scenario: LoadDocument returns typed validation errors

- **GIVEN** a document with missing required fields
- **WHEN** LoadDocument is called with the filepath
- **THEN** it SHALL detect type and parse the document
- **AND** it SHALL call document.Validate()
- **AND** it SHALL return ValidationError instances

#### Scenario: LoadDocument returns typed parse errors

- **GIVEN** a document with invalid YAML syntax
- **WHEN** LoadDocument is called with the filepath
- **THEN** it SHALL return a ParseError
- **AND** ParseError.Path SHALL be the filepath
- **AND** ParseError.Cause SHALL wrap the YAML parsing error

---

## ADDED Requirements

### Requirement: Typed Error Returns

LoadDocument SHALL return typed errors for different failure modes.

#### Scenario: File not found returns FileError

- **GIVEN** a filepath that does not exist
- **WHEN** LoadDocument is called
- **THEN** it SHALL return a FileError
- **AND** FileError.Path SHALL be the filepath
- **AND** FileError.Operation SHALL be "read"
- **AND** FileError.IsNotFound() SHALL return true

#### Scenario: Permission denied returns FileError

- **GIVEN** a filepath with insufficient permissions
- **WHEN** LoadDocument is called
- **THEN** it SHALL return a FileError
- **AND** FileError.Operation SHALL be "read"

#### Scenario: YAML parse failure returns ParseError

- **GIVEN** a file with invalid YAML syntax
- **WHEN** LoadDocument is called
- **THEN** it SHALL return a ParseError
- **AND** ParseError.Path SHALL be the filepath
- **AND** ParseError.Message SHALL describe the parse failure
- **AND** ParseError.Cause SHALL wrap the underlying YAML error

#### Scenario: Frontmatter parse failure returns ParseError

- **GIVEN** a file with invalid frontmatter structure
- **WHEN** LoadDocument is called
- **THEN** it SHALL return a ParseError
- **AND** ParseError.Path SHALL be the filepath
- **AND** ParseError.Message SHALL describe the frontmatter issue

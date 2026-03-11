# document-loading Specification

## Purpose

Load and validate documents from files, detecting type from filename patterns.

## Requirements
### Requirement: Filename-Based Type Detection

The system SHALL detect document type from filename patterns only.

#### Scenario: Agent detected from agent- prefix
**Given** a file named "agent-coder.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL detect type as "agent"

#### Scenario: Agent detected from -agent suffix
**Given** a file named "code-review-agent.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL detect type as "agent"

#### Scenario: Command detected from command- prefix or -command suffix
**Given** a file named "command-test.md" or "lint-command.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL detect type as "command"

#### Scenario: Memory detected from memory- prefix or -memory suffix
**Given** a file named "memory-go.md" or "api-memory.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL detect type as "memory"

#### Scenario: Skill detected from skill- prefix or -skill suffix
**Given** a file named "skill-review.md" or "refactor-skill.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL detect type as "skill"

#### Scenario: Unrecognizable filename returns typed error
**Given** a file named "my-document.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL return a ParseError
**And** ParseError.Path SHALL be the filepath
**And** ParseError.Message SHALL list expected filename patterns
**And** ParseError.Unwrap() SHALL return nil

---

### Requirement: Document Loading Workflow

LoadDocument SHALL combine type detection, parsing, and validation in one function.

#### Scenario: LoadDocument parses and validates valid document
**Given** a valid Agent document file
**When** LoadDocument is called with the filepath
**Then** it SHALL detect document type from filename
**And** it SHALL call ParseDocument with detected type
**And** it SHALL call document.Validate() on the parsed document
**And** it SHALL return the validated document

#### Scenario: LoadDocument returns validation errors
**Given** a document with missing required fields
**When** LoadDocument is called with the filepath
**Then** it SHALL detect type and parse the document
**And** it SHALL call document.Validate()
**And** it SHALL return validation errors

#### Scenario: LoadDocument returns validation errors
**Given** a document with missing required fields
**When** LoadDocument is called with the filepath
**Then** it SHALL detect type and parse the document
**And** it SHALL call document.Validate()
**And** it SHALL return ValidationError instances

#### Scenario: LoadDocument returns typed parse errors
**Given** a document with invalid YAML syntax
**When** LoadDocument is called with the filepath
**Then** it SHALL return a ParseError
**And** ParseError.Path SHALL be the filepath
**And** ParseError.Cause SHALL wrap the YAML parsing error

---

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

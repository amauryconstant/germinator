# document-loading Specification

## Purpose
TBD - created by archiving change add-core-infrastructure. Update Purpose after archive.
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

#### Scenario: Unrecognizable filename returns error
**Given** a file named "my-document.md"
**When** LoadDocument analyzes the filename
**Then** it SHALL return an error
**And** it SHALL list expected filename patterns in error message

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

#### Scenario: LoadDocument returns parse errors
**Given** a document with invalid YAML syntax
**When** LoadDocument is called with the filepath
**Then** it SHALL return the YAML parsing error


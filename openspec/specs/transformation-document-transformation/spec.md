# document-transformation Specification

## Purpose

Transform documents between platform formats with validation and serialization.

## Requirements
### Requirement: Transformation Pipeline Function

The system SHALL provide a function that orchestrates complete transformation workflow.

#### Scenario: Transform valid document
**Given** a valid agent document file
**When** TransformDocument(input, output, "claude-code") is called
**Then** it SHALL parse the document
**And** it SHALL validate the document
**And** it SHALL serialize the document
**And** it SHALL write the output to the specified file
**And** it SHALL return nil (success)

#### Scenario: Transform document fails on validation
**Given** an invalid command document file
**When** TransformDocument(input, output, "claude-code") is called
**Then** it SHALL parse the document successfully
**And** it SHALL validate the document and receive errors
**And** it SHALL return ValidationError instances
**And** it SHALL NOT create the output file

#### Scenario: Transform document fails on parsing
**Given** a file with invalid YAML frontmatter
**When** TransformDocument(input, output, "claude-code") is called
**Then** it SHALL attempt to parse the document
**And** it SHALL return a ParseError with filepath and cause
**And** it SHALL NOT create the output file

#### Scenario: Transform document with write error
**Given** a valid document and an output path to a read-only directory
**When** TransformDocument(input, output, "claude-code") is called
**Then** it SHALL successfully parse, validate, and serialize
**And** it SHALL attempt to write to the output file
**And** it SHALL return a FileError with path and "write" operation

---

### Requirement: Workflow Orchestration Order

The transformation pipeline SHALL execute steps in a specific order: Parse → Validate → Serialize → Write.

#### Scenario: Verify workflow order
**Given** a transformation is initiated
**When** the workflow executes
**Then** it SHALL first parse the document
**And** validation SHALL occur after parsing
**And** serialization SHALL occur after successful validation
**And** file writing SHALL occur after successful serialization
**And** each step SHALL complete before the next begins

#### Scenario: Fail fast on parse error
**Given** a file with invalid structure
**When** TransformDocument is called
**Then** parsing SHALL fail
**And** validation SHALL NOT be attempted
**And** serialization SHALL NOT be attempted
**And** file writing SHALL NOT be attempted

#### Scenario: Fail fast on validation error
**Given** a file with valid structure but invalid content
**When** TransformDocument is called
**Then** parsing SHALL succeed
**And** validation SHALL fail
**And** serialization SHALL NOT be attempted
**And** file writing SHALL NOT be attempted

---

### Requirement: Output File Writing

The transformation pipeline SHALL write the transformed document to the specified output file.

#### Scenario: Write output file
**Given** a transformed document string
**When** the transformation completes successfully
**Then** it SHALL create the output file if it doesn't exist
**Or** it SHALL overwrite the output file if it exists
**And** the file SHALL contain the complete transformed document

---

### Requirement: Typed Transform Errors

TransformDocument SHALL return typed errors for transformation failures.

#### Scenario: Template rendering failure returns TransformError

- **GIVEN** a template file is missing or invalid
- **WHEN** TransformDocument attempts to render
- **THEN** it SHALL return a TransformError
- **AND** TransformError.Operation SHALL be "render"
- **AND** TransformError.Platform SHALL be the target platform

#### Scenario: Platform adapter failure returns TransformError

- **GIVEN** platform-specific conversion fails
- **WHEN** TransformDocument processes the document
- **THEN** it SHALL return a TransformError
- **AND** TransformError.Operation SHALL describe the failed step

---

### Requirement: Typed Validation Errors

TransformDocument SHALL return typed ValidationErrors for validation failures.

#### Scenario: Missing required field returns ValidationError

- **GIVEN** a document missing a required field
- **WHEN** TransformDocument validates the document
- **THEN** it SHALL return a ValidationError
- **AND** ValidationError.Field SHALL identify the missing field
- **AND** ValidationError.Message SHALL describe the requirement

#### Scenario: Invalid field value returns ValidationError with suggestions

- **GIVEN** a document with an invalid enum value
- **WHEN** TransformDocument validates the document
- **THEN** it SHALL return a ValidationError
- **AND** ValidationError.Suggestions SHALL list valid values

---

### Requirement: Canonicalization Error Handling

CanonicalizeDocument SHALL return typed errors for canonicalization failures.

#### Scenario: Platform parsing failure returns ParseError

- **GIVEN** a platform document with invalid structure
- **WHEN** CanonicalizeDocument is called
- **THEN** it SHALL return a ParseError
- **AND** ParseError.Path SHALL be the input filepath

#### Scenario: Canonical validation failure returns ValidationError

- **GIVEN** a platform document with invalid content
- **WHEN** CanonicalizeDocument validates
- **THEN** it SHALL return ValidationError instances

#### Scenario: Canonical write failure returns FileError

- **GIVEN** output path is not writable
- **WHEN** CanonicalizeDocument attempts to write
- **THEN** it SHALL return a FileError
- **AND** FileError.Operation SHALL be "write"

---

### Requirement: ValidateDocument Typed Errors

ValidateDocument SHALL return typed errors for validation failures.

#### Scenario: Invalid platform returns ConfigError

- **GIVEN** an invalid platform name
- **WHEN** ValidateDocument is called
- **THEN** it SHALL return a ConfigError
- **AND** ConfigError.Field SHALL be "platform"
- **AND** ConfigError.Available SHALL list valid platforms

#### Scenario: Parse failure returns ParseError

- **GIVEN** a file that cannot be parsed
- **WHEN** ValidateDocument is called
- **THEN** it SHALL return a ParseError
- **AND** ParseError.Path SHALL be the filepath

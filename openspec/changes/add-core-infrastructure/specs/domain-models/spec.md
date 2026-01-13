# domain-models Specification

## Purpose

Define document models (Agent, Command, Memory, Skill) with YAML field mapping and struct-based validation methods.

## ADDED Requirements

### Requirement: Document Models

The project SHALL define Agent, Command, Memory, and Skill structs with all frontmatter fields.

#### Scenario: Agent struct has all required fields
**Given** Agent struct is defined
**When** a developer inspects the struct
**Then** it SHALL have an ID string field with yaml tag "id"
**And** it SHALL have a LastChanged string field with yaml tag "last_changed"
**And** it SHALL have a Model string field with yaml tag "model"
**And** it SHALL have specialization fields (PrimaryCapability, ExpertiseDomain, ScopeBoundaries)
**And** it SHALL have selection criteria fields (Triggers, UseWhen, AvoidWhen, ComplexityLevel)
**And** it SHALL have compatibility fields (RequiredCapabilities, OptionalCapabilities, InputExpectations, OutputFormat)
**And** it SHALL have FilePath and Content string fields

#### Scenario: Command struct has all required fields
**Given** Command struct is defined
**When** a developer inspects the struct
**Then** it SHALL have Name, Description, Version, Category, LastChanged string fields with yaml tags
**And** it SHALL have Tools, Files, Args []string fields with yaml tags
**And** it SHALL have FilePath and Content string fields

#### Scenario: Memory struct has all required fields
**Given** Memory struct is defined
**When** a developer inspects the struct
**Then** it SHALL have Title, Description, AppliesTo, LastChanged string fields with yaml tags
**And** it SHALL have FilePath and Content string fields

#### Scenario: Skill struct has all required fields
**Given** Skill struct is defined
**When** a developer inspects the struct
**Then** it SHALL have Name, Description, LastChanged string fields with yaml tags
**And** it SHALL have FilePath and Content string fields

---

### Requirement: Document Validation Methods

Each document struct SHALL implement a Validate() method that checks struct fields.

#### Scenario: Agent Validate checks required fields
**Given** an Agent struct with some fields missing
**When** Agent.Validate() is called
**Then** it SHALL return errors for missing required fields (id, last_changed, model, etc.)

#### Scenario: Memory Validate checks enum values
**Given** a Memory struct with applies_to set to "invalid"
**When** Memory.Validate() is called
**Then** it SHALL return an error
**And** it SHALL include valid enum values in error message

#### Scenario: Validate returns multiple errors
**Given** a document with multiple validation failures
**When** Validate() is called
**Then** it SHALL return an error slice with all failures
**And** each error SHALL be independently actionable

---

### Requirement: YAML Frontmatter Unmarshaling

The document models SHALL support YAML frontmatter parsing with correct field mapping.

#### Scenario: Valid YAML unmarshals correctly
**Given** a document with valid YAML frontmatter
**When** YAML is unmarshaled into a document struct
**Then** all fields SHALL be populated with correct values
**And** type conversions SHALL be handled appropriately

#### Scenario: Invalid YAML returns error
**Given** a document with invalid YAML frontmatter
**When** YAML is unmarshaled into a document struct
**Then** an error SHALL be returned
**And** it SHALL indicate the line number of parse failure

---

### Requirement: Single Models File

All document models SHALL be defined in a single `models.go` file.

#### Scenario: Models file contains all document types
**Given** `pkg/models/models.go` file exists
**When** the file is inspected
**Then** it SHALL define Agent struct
**And** it SHALL define Command struct
**And** it SHALL define Memory struct
**And** it SHALL define Skill struct

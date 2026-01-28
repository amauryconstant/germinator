# domain-models Specification

## Purpose
Define document models representing AI coding assistant documents (Agent, Command, Memory, Skill) following Claude Code format.

## Requirements
### Requirement: Document Models

The project SHALL define Agent, Command, Memory, and Skill structs with all frontmatter fields matching Claude Code document format.

#### Scenario: Agent struct has all required fields
**Given** Agent struct is defined
**When** a developer inspects the struct
**Then** it SHALL have a Name string field with yaml tag "name"
**And** it SHALL have a Description string field with yaml tag "description"
**And** it SHALL have a Tools []string field with yaml tag "tools"
**And** it SHALL have a DisallowedTools []string field with yaml tag "disallowedTools"
**And** it SHALL have a Model string field with yaml tag "model"
**And** it SHALL have a PermissionMode string field with yaml tag "permissionMode"
**And** it SHALL have a Skills []string field with yaml tag "skills"
**And** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Command struct has all required fields
**Given** Command struct is defined
**When** a developer inspects the struct
**Then** it SHALL have AllowedTools []string field with yaml tag "allowed-tools"
**And** it SHALL have ArgumentHint string field with yaml tag "argument-hint"
**And** it SHALL have Context string field with yaml tag "context"
**And** it SHALL have Agent string field with yaml tag "agent"
**And** it SHALL have Description string field with yaml tag "description"
**And** it SHALL have Model string field with yaml tag "model"
**And** it SHALL have DisableModelInvocation bool field with yaml tag "disable-model-invocation"
**And** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Memory struct has all required fields
**Given** Memory struct is defined
**When** a developer inspects the struct
**Then** it SHALL have Paths []string field with yaml tag "paths"
**And** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Skill struct has all required fields
**Given** Skill struct is defined
**When** a developer inspects the struct
**Then** it SHALL have Name string field with yaml tag "name"
**And** it SHALL have Description string field with yaml tag "description"
**And** it SHALL have AllowedTools []string field with yaml tag "allowed-tools"
**And** it SHALL have Model string field with yaml tag "model"
**And** it SHALL have Context string field with yaml tag "context"
**And** it SHALL have Agent string field with yaml tag "agent"
**And** it SHALL have UserInvocable bool field with yaml tag "user-invocable"
**And** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

---

### Requirement: Document Validation Methods

Each document struct SHALL implement a Validate() method that checks struct fields for validity.

#### Scenario: Agent Validate checks required fields
**Given** an Agent struct with some fields missing
**When** Agent.Validate() is called
**Then** it SHALL return error if name is missing
**And** it SHALL return error if name does not match regex `^[a-z-]+$`
**And** it SHALL return error if description is missing
**And** it SHALL return error if model is not one of: sonnet, opus, haiku, inherit
**And** it SHALL return error if permissionMode is not one of: default, acceptEdits, dontAsk, bypassPermissions, plan

#### Scenario: Command Validate checks context field
**Given** a Command struct
**When** Command.Validate() is called
**Then** it SHALL return error if context is specified and not "fork"

#### Scenario: Memory Validate is no-op
**Given** a Memory struct
**When** Memory.Validate() is called
**Then** it SHALL always return nil (memory has no validation rules)

#### Scenario: Skill Validate checks required fields
**Given** a Skill struct
**When** Skill.Validate() is called
**Then** it SHALL return error if name is missing
**And** it SHALL return error if name exceeds 64 characters
**And** it SHALL return error if name does not match regex `^[a-z0-9-]+$`
**And** it SHALL return error if description is missing
**And** it SHALL return error if description exceeds 1024 characters
**And** it SHALL return error if context is specified and not "fork"

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
**And** kebab-case YAML keys SHALL map to camelCase struct fields

#### Scenario: Invalid YAML returns error
**Given** a document with invalid YAML frontmatter
**When** YAML is unmarshaled into a document struct
**Then** an error SHALL be returned
**And** it SHALL indicate the line number of parse failure

---

### Requirement: Models Location

All document models SHALL be defined in the internal/models package.

#### Scenario: Models file contains all document types
**Given** `internal/models/models.go` file exists
**When** the file is inspected
**Then** it SHALL define Agent struct
**And** it SHALL define Command struct
**And** it SHALL define Memory struct
**And** it SHALL define Skill struct
**And** it SHALL have Validate() method for each struct

#### Scenario: Models package has documentation
**Given** `internal/models/doc.go` file exists
**When** the file is inspected
**Then** it SHALL describe the package's purpose as containing document model definitions

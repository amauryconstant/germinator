# domain-models Specification

## Purpose
Define canonical document models (Agent, Command, Memory, Skill) with domain-driven field names independent of platform specifics.

## Requirements
### Requirement: Document Models

The project SHALL define canonical document models (Agent, Command, Memory, Skill) with domain-driven field names independent of platform specifics.

#### Scenario: Canonical Agent struct

- **GIVEN** Canonical Agent struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Name string field with yaml tag "name"
- **AND** it SHALL have Description string field with yaml tag "description"
- **AND** it SHALL have Tools []string field with yaml tag "tools" (lowercase names)
- **AND** it SHALL have DisallowedTools []string field with yaml tag "disallowedTools" (lowercase names)
- **AND** it SHALL have PermissionPolicy PermissionPolicy enum field with yaml tag "permissionPolicy" (not string)
- **AND** it SHALL have Behavior AgentBehavior struct field with yaml tag "behavior"
- **AND** it SHALL have Model string field with yaml tag "model"
- **AND** it SHALL have Targets map[string]PlatformConfig field with yaml tag "targets"
- **AND** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical AgentBehavior struct

- **GIVEN** Canonical AgentBehavior struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Mode string field with yaml tag "mode" (values: primary, subagent, all)
- **AND** it SHALL have Temperature *float64 field with yaml tag "temperature"
- **AND** it SHALL have MaxSteps int field with yaml tag "maxSteps"
- **AND** it SHALL have Prompt string field with yaml tag "prompt"
- **AND** it SHALL have Hidden bool field with yaml tag "hidden"
- **AND** it SHALL have Disabled bool field with yaml tag "disabled"

#### Scenario: Canonical Command struct

- **GIVEN** Canonical Command struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Name string field with yaml tag "name"
- **AND** it SHALL have Description string field with yaml tag "description"
- **AND** it SHALL have Tools []string field with yaml tag "tools" (lowercase names)
- **AND** it SHALL have Execution CommandExecution struct field with yaml tag "execution"
- **AND** it SHALL have Arguments CommandArguments struct field with yaml tag "arguments"
- **AND** it SHALL have Model string field with yaml tag "model"
- **AND** it SHALL have Targets map[string]PlatformConfig field with yaml tag "targets"
- **AND** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical CommandExecution struct

- **GIVEN** Canonical CommandExecution struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Context string field with yaml tag "context"
- **AND** it SHALL have Subtask bool field with yaml tag "subtask"
- **AND** it SHALL have Agent string field with yaml tag "agent"

#### Scenario: Canonical CommandArguments struct

- **GIVEN** Canonical CommandArguments struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Hint string field with yaml tag "hint"

#### Scenario: Canonical Memory struct

- **GIVEN** Canonical Memory struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Paths []string field with yaml tag "paths"
- **AND** it SHALL have Content string field with yaml tag "content"
- **AND** it SHALL have FilePath string field (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical Skill struct

- **GIVEN** Canonical Skill struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Name string field with yaml tag "name"
- **AND** it SHALL have Description string field with yaml tag "description"
- **AND** it SHALL have Tools []string field with yaml tag "tools" (lowercase names)
- **AND** it SHALL have Extensions SkillExtensions struct field with yaml tag "extensions"
- **AND** it SHALL have Execution SkillExecution struct field with yaml tag "execution"
- **AND** it SHALL have Model string field with yaml tag "model"
- **AND** it SHALL have Targets map[string]PlatformConfig field with yaml tag "targets"
- **AND** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical SkillExtensions struct

- **GIVEN** Canonical SkillExtensions struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have License string field with yaml tag "license"
- **AND** it SHALL have Compatibility []string field with yaml tag "compatibility"
- **AND** it SHALL have Metadata map[string]string field with yaml tag "metadata"
- **AND** it SHALL have Hooks map[string]string field with yaml tag "hooks"

#### Scenario: Canonical SkillExecution struct

- **GIVEN** Canonical SkillExecution struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Context string field with yaml tag "context"
- **AND** it SHALL have Agent string field with yaml tag "agent"
- **AND** it SHALL have UserInvocable bool field with yaml tag "userInvocable"

---

### Requirement: Document Validation Methods

Each document struct SHALL implement a Validate() method that checks struct fields for validity.

#### Scenario: Agent Validate checks required fields
**Given** an Agent struct with some fields missing
**When** Agent.Validate() is called
**Then** it SHALL return error if name is missing
**And** it SHALL return error if name does not match regex `^[a-z-]+$`
**And** it SHALL return error if description is missing
**And** it SHALL return error if permissionPolicy is not one of: restrictive, balanced, permissive, analysis, unrestricted

#### Scenario: Command Validate checks context field
**Given** a Command struct
**When** Command.Validate() is called
**Then** it SHALL return error if execution.context is specified and not "fork"

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
**And** it SHALL return error if execution.context is specified and not "fork"

#### Scenario: Validate returns multiple errors
**Given** a document with multiple validation failures
**When** Validate() is called
**Then** it SHALL return an error slice with all failures
**And** each error SHALL be independently actionable

---

### Requirement: YAML Frontmatter Unmarshaling

The document models SHALL support YAML frontmatter parsing with correct field mapping.

#### Scenario: Valid YAML unmarshals correctly
- **GIVEN** A document with valid YAML frontmatter
- **WHEN** YAML is unmarshaled into a document struct
- **THEN** All fields SHALL be populated with correct values
- **AND** Type conversions SHALL be handled appropriately
- **AND** Kebab-case YAML keys SHALL map to camelCase struct fields

#### Scenario: Invalid YAML returns error
- **GIVEN** A document with invalid YAML frontmatter
- **WHEN** YAML is unmarshaled into a document struct
- **THEN** An error SHALL be returned
- **AND** It SHALL indicate the line number of parse failure

---

### Requirement: Models Location

All document models SHALL be defined in the internal/models package.

#### Scenario: Models file contains all document types
- **GIVEN** `internal/models/models.go` file exists
- **WHEN** The file is inspected
- **THEN** It SHALL define Agent struct
- **AND** It SHALL define Command struct
- **AND** It SHALL define Memory struct
- **AND** It SHALL define Skill struct
- **AND** It SHALL have Validate() method for each struct
- **AND** It SHALL define Behavior, Extensions, Execution structs as needed

#### Scenario: Models package has documentation
- **GIVEN** `internal/models/doc.go` file exists
- **WHEN** The file is inspected
- **THEN** It SHALL describe the package's purpose as containing canonical document model definitions

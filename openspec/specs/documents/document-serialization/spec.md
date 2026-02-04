# document-serialization Specification

## Purpose
TBD - created by archiving change add-document-processing-infrastructure. Update Purpose after archive.
## Requirements
### Requirement: Template Rendering Function

The system SHALL provide a function that renders document structs to YAML frontmatter and markdown body format.

#### Scenario: Render agent document to template
**Given** an Agent struct with populated fields
**When** RenderDocument(agent, "claude-code") is called
**Then** it SHALL return a complete formatted document string
**And** the string SHALL contain YAML frontmatter with all agent fields
**And** the string SHALL preserve the agent's markdown body content

#### Scenario: Render command document to template
**Given** a Command struct with populated fields
**When** RenderDocument(command, "claude-code") is called
**Then** it SHALL return a complete formatted document string
**And** the string SHALL contain YAML frontmatter with all command fields
**And** the string SHALL preserve the command's markdown body content

#### Scenario: Render skill document to template
**Given** a Skill struct with populated fields
**When** RenderDocument(skill, "claude-code") is called
**Then** it SHALL return a complete formatted document string
**And** the string SHALL contain YAML frontmatter with all skill fields
**And** the string SHALL preserve the skill's markdown body content

#### Scenario: Render memory document to template
**Given** a Memory struct with populated fields
**When** RenderDocument(memory, "claude-code") is called
**Then** it SHALL return a complete formatted document string
**And** the string SHALL contain YAML frontmatter with memory paths (if present)
**And** the string SHALL preserve the memory's markdown body content

#### Scenario: Handle missing template
**Given** a document type and platform
**When** RenderDocument is called with a non-existent template
**Then** it SHALL return an error
**And** the error message SHALL indicate the missing template file

---

### Requirement: Markdown Body Preservation

The template rendering SHALL preserve the original markdown body content exactly.

#### Scenario: Preserve markdown formatting
**Given** a document with markdown body content containing formatting, code blocks, and links
**When** RenderDocument is called
**Then** the rendered output SHALL preserve all markdown formatting exactly
**And** code blocks SHALL be unchanged
**And** links SHALL be unchanged

---

### Requirement: Template Function Registration
The system SHALL register custom template functions for platform-specific transformations.

#### Scenario: Register template functions
- **GIVEN** RenderDocument function is called
- **WHEN** a template is parsed
- **THEN** it SHALL create a FuncMap with registered functions
- **AND** transformPermissionMode function SHALL be available

### Requirement: Permission Transformation Function
The system SHALL provide transformPermissionMode template function to convert Claude Code permissionMode enum to OpenCode permission objects.

#### Scenario: Transform permissionMode to OpenCode format
- **GIVEN** an Agent with permissionMode "acceptEdits"
- **WHEN** {{transformPermissionMode .PermissionMode}} is used in template
- **THEN** it SHALL output permission object with edit: {"*": "allow"} and bash: {"*": "ask"}

#### Scenario: Transform permissionMode "default"
- **WHEN** {{transformPermissionMode "default"}} is called
- **THEN** it SHALL output permission: {"edit": {"*": "ask"}, "bash": {"*": "ask"}}

### Requirement: Template Function Documentation
Template functions SHALL have Go documentation describing their purpose and usage.

#### Scenario: Function documentation exists
- **GIVEN** template functions are implemented in internal/core/template_funcs.go
- **WHEN** code is inspected
- **THEN** transformPermissionMode function SHALL have Go doc comment
- **AND** each doc SHALL describe purpose, parameters, and return value

### Requirement: Template Function Testing
Template functions SHALL have unit tests covering all scenarios.

#### Scenario: Unit tests pass for transformPermissionMode
- **GIVEN** transformPermissionMode function is tested
- **WHEN** unit tests are run
- **THEN** all 5 Claude Code modes SHALL be tested
- **AND** unknown mode SHALL be tested
- **AND** all tests SHALL pass


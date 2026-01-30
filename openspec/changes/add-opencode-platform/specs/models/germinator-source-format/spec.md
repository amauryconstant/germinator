## ADDED Requirements

### Requirement: Germinator format serves as canonical source
The system SHALL use a single canonical YAML format containing ALL platform fields (Claude Code + OpenCode) as the source of truth for transformations.

#### Scenario: Source format contains all field types
- **GIVEN** A Germinator source YAML file
- **WHEN** The file is parsed
- **THEN** Common fields SHALL be present (Name, Description, Model, Content)
- **AND** Claude Code-specific fields SHALL be parseable (Tools, PermissionMode, Skills)
- **AND** OpenCode-specific fields SHALL be parseable (Mode, Temperature, MaxSteps, Hidden, Prompt, Disable)

### Requirement: All platform fields MUST be parseable from YAML
The system SHALL NOT use `yaml:"-"` tags to prevent parsing of platform-specific fields.

#### Scenario: OpenCode fields are parseable from YAML
- **GIVEN** An Agent struct with OpenCode-specific fields (Mode, Temperature, MaxSteps)
- **WHEN** YAML is unmarshaled into a struct
- **THEN** Mode SHALL be parsed from "mode" YAML key
- **AND** Temperature SHALL be parsed from "temperature" YAML key
- **AND** MaxSteps SHALL be parsed from "maxSteps" YAML key

#### Scenario: Claude Code fields remain parseable
- **GIVEN** An Agent struct with Claude Code-specific fields (PermissionMode, Skills)
- **WHEN** YAML is unmarshaled into a struct
- **THEN** PermissionMode SHALL be parsed from "permissionMode" YAML key
- **AND** Skills SHALL be parsed from "skills" YAML key

### Requirement: Transformation is unidirectional
The system SHALL support unidirectional transformation from Germinator format to target platform formats (Claude Code OR OpenCode).

#### Scenario: Transform to Claude Code omits OpenCode fields
- **GIVEN** A Germinator source with both Claude Code and OpenCode fields
- **WHEN** RenderDocument is called with platform="claude-code"
- **THEN** Claude Code-specific fields SHALL be included (permissionMode, skills)
- **AND** OpenCode-specific fields SHALL be omitted (mode, temperature, maxSteps)
- **AND** Common fields SHALL be included (name, description, model)

#### Scenario: Transform to OpenCode omits Claude Code fields
- **GIVEN** A Germinator source with both Claude Code and OpenCode fields
- **WHEN** RenderDocument is called with platform="opencode"
- **THEN** OpenCode-specific fields SHALL be included (mode, temperature, maxSteps)
- **AND** Claude Code-specific fields SHALL be omitted (permissionMode, skills)
- **AND** Common fields SHALL be included (name, description, model)

### Requirement: Templates filter platform-specific fields
The system SHALL use Go templates to conditionally render platform-specific fields based on the target platform.

#### Scenario: Claude Code template conditionally renders fields
- **GIVEN** An Agent with both Claude Code and OpenCode fields
- **WHEN** The Claude Code agent template is rendered
- **THEN** Template SHALL use {{if .PermissionMode}} for Claude Code fields
- **AND** Template SHALL NOT render OpenCode fields
- **AND** Template SHALL render common fields

#### Scenario: OpenCode template conditionally renders fields
- **GIVEN** An Agent with both Claude Code and OpenCode fields
- **WHEN** The OpenCode agent template is rendered
- **THEN** Template SHALL use {{if .Mode}} for OpenCode fields
- **AND** Template SHALL NOT render Claude Code fields
- **AND** Template SHALL render common fields

### Requirement: All model fields MUST have YAML and JSON tags
The system SHALL ensure all fields in Agent, Command, Skill, and Memory models have both `yaml:"field,omitempty"` and `json:"field,omitempty"` tags.

#### Scenario: Agent model tags are correct
- **GIVEN** An Agent model definition
- **WHEN** The struct is inspected
- **THEN** OpenCode-specific fields SHALL have yaml:"field,omitempty" tags (not yaml:"-")
- **AND** Claude Code-specific fields SHALL have yaml:"field,omitempty" tags
- **AND** All fields SHALL have json:"field,omitempty" tags

#### Scenario: Command model tags are correct
- **GIVEN** A Command model definition
- **WHEN** The struct is inspected
- **THEN** Subtask field SHALL have yaml:"subtask,omitempty" tag
- **AND** AllowedTools field SHALL have yaml:"allowed-tools,omitempty" tag
- **AND** All fields SHALL have json tags with omitempty

### Requirement: Source YAML files can contain optional platform fields
The system SHALL allow source YAML files to omit optional platform-specific fields that are not needed for the target transformation.

#### Scenario: Minimal agent with only common fields
- **GIVEN** A Germinator source YAML with only common fields
```yaml
---
name: minimal-agent
description: A minimal agent
model: anthropic/claude-sonnet-4-20250514
---
Agent content
```
- **WHEN** The file is parsed
- **THEN** Common fields SHALL be parsed correctly
- **AND** Optional platform fields SHALL be empty strings or zero values

#### Scenario: Transform minimal agent to Claude Code
- **GIVEN** A minimal agent with only common fields
- **WHEN** RenderDocument is called with platform="claude-code"
- **THEN** Common fields SHALL be rendered
- **AND** Empty Claude Code fields SHALL NOT be rendered

#### Scenario: Transform minimal agent to OpenCode with default mode
- **GIVEN** A minimal agent with only common fields
- **WHEN** RenderDocument is called with platform="opencode"
- **THEN** Common fields SHALL be rendered
- **AND** Mode field SHALL default to "all" if not specified

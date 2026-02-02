## ADDED Requirements

### Requirement: Germinator format serves as canonical source
The system SHALL use a single canonical YAML format containing ALL platform fields (Claude Code + OpenCode) as source of truth for transformations.

#### Scenario: Source format contains all field types
- **GIVEN** A Germinator source YAML file
- **WHEN** The file is parsed
- **THEN** Common fields SHALL be present (Name, Description, Model, Content)
- **AND** Claude Code-specific fields SHALL be parseable (Tools, PermissionMode, Skills)
- **AND** OpenCode-specific fields SHALL be parseable (Mode, Temperature (*float64), Steps, Hidden, Prompt, Disable)

### Requirement: All platform fields MUST be parseable from YAML
The system SHALL NOT use `yaml:"-"` tags to prevent parsing of platform-specific fields.

#### Scenario: Platform fields parseable with proper YAML tags
- **GIVEN** Models with platform-specific fields (OpenCode: Mode, Temperature; Claude Code: PermissionMode, Skills)
- **WHEN** YAML is unmarshaled into structs
- **THEN** All fields SHALL parse from their respective YAML keys (mode, temperature, permissionMode, skills)
- **AND** All fields SHALL have `yaml:"field,omitempty"` tags (not `yaml:"-"`)

### Requirement: Transformation is unidirectional
The system SHALL support unidirectional transformation from Germinator format to target platform formats (Claude Code OR OpenCode).

#### Scenario: Transform to Claude Code omits OpenCode fields
- **GIVEN** A Germinator source with both Claude Code and OpenCode fields
- **WHEN** RenderDocument is called with platform="claude-code"
- **THEN** Claude Code-specific fields SHALL be included (permissionMode, skills)
- **AND** OpenCode-specific fields SHALL be omitted (mode, temperature, steps, hidden, prompt, disable)
- **AND** Common fields SHALL be included (name, description, model)

#### Scenario: Transform to OpenCode omits Claude Code fields
- **GIVEN** A Germinator source with both Claude Code and OpenCode fields
- **WHEN** RenderDocument is called with platform="opencode"
- **THEN** OpenCode-specific fields SHALL be included (mode, temperature, steps, hidden, prompt, disable)
- **AND** Claude Code-specific fields SHALL be omitted (permissionMode, skills)
- **AND** Common fields SHALL be included (name, description, model)

### Requirement: Templates filter platform-specific fields
The system SHALL use Go templates to conditionally render platform-specific fields based on target platform.

#### Scenario: Templates conditionally render fields
- **GIVEN** A model with both Claude Code and OpenCode fields
- **WHEN** A platform template is rendered
- **THEN** Claude Code template SHALL use `{{if .PermissionMode}}` for Claude Code fields and NOT render OpenCode fields
- **AND** OpenCode template SHALL use `{{if .Mode}}` for OpenCode fields and NOT render Claude Code fields
- **AND** Both templates SHALL render common fields

### Requirement: All model fields MUST have YAML and JSON tags
The system SHALL ensure all fields in Agent, Command, Skill, and Memory models have both `yaml:"field,omitempty"` and `json:"field,omitempty"` tags.

#### Scenario: Model fields have proper tags
- **GIVEN** Model definitions (Agent, Command, Skill, Memory)
- **WHEN** The structs are inspected
- **THEN** All fields SHALL have `yaml:"field,omitempty"` tags
- **AND** All fields SHALL have `json:"field,omitempty"` tags
- **AND** No fields SHALL have `yaml:"-"` tags that prevent parsing

### Requirement: Source YAML files can contain optional platform fields
The system SHALL allow source YAML files to omit optional platform-specific fields that are not needed for target transformation.

#### Scenario: Minimal agent with only common fields
- **GIVEN** A Germinator source YAML with only common fields (name, description, model, content)
- **WHEN** The file is parsed and transformed to OpenCode
- **THEN** Common fields SHALL be parsed correctly
- **AND** Empty platform fields SHALL be zero values (empty strings, nil pointers)
- **AND** Mode field SHALL default to "all" in OpenCode output
- **AND** Empty platform fields SHALL NOT be rendered

#### Scenario: Temperature nil vs 0.0 distinction
- **GIVEN** Agent model with Temperature field as *float64 pointer
- **WHEN** Temperature is nil
- **THEN** Temperature field SHALL NOT be rendered in OpenCode output (user didn't set it)
- **AND** OpenCode will use model's default temperature
- **WHEN** Temperature is 0.0
- **THEN** Temperature field SHALL be rendered as `temperature: 0.0` in OpenCode output
- **AND** User explicitly requested deterministic low-temperature behavior


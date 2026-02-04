# platform-agnostic-models Specification

## Purpose
Platform-agnostic models support both Claude Code and OpenCode fields in single structs.

## Requirements

### Requirement: Platform-specific fields MUST be parseable from YAML
The system SHALL NOT use `yaml:"-"` tags to prevent parsing of platform-specific fields. All platform fields SHALL have proper YAML tags (`yaml:"field,omitempty"`) for full parseability.

#### Scenario: Platform fields parseable with proper YAML tags
- **GIVEN** Models with platform-specific fields (OpenCode: Mode, Temperature *float64, MaxSteps; Claude Code: PermissionMode, Skills)
- **WHEN** YAML is unmarshaled into structs
- **THEN** All fields SHALL parse from their respective YAML keys
- **AND** Temperature SHALL be parsed as *float64 pointer to distinguish nil from 0.0

### Requirement: Models capture core document concepts without platform constraints
Platform-agnostic models support both Claude Code and OpenCode fields in single structs.

#### Scenario: Platform-agnostic models support both platforms
- **GIVEN** Platform-agnostic models with all platform fields
- **WHEN** Fields are parsed from Germinator YAML
- **THEN** Common fields (Name, Description, Model) are preserved for both platforms
- **AND** Platform-specific fields (PermissionMode vs Mode, Temperature, etc.) are available when needed

### Requirement: Tool configuration uses flat arrays as single internal representation
Tool configuration uses Claude Code's flat array format (tools, disallowedTools, allowedTools) in models. Templates transform to platform-specific formats.

#### Scenario: Tools represented as flat arrays in model
- **GIVEN** An Agent model
- **WHEN** Tools field is populated
- **THEN** Tools SHALL be a string slice ([]string)
- **AND** Templates SHALL transform flat arrays to platform-specific formats (arrays for Claude Code, objects for OpenCode)

### Requirement: Model IDs are user-provided, platform-specific strings
Model IDs (e.g., anthropic/claude-sonnet-4-20250514) are provided by user and preserved without normalization.

#### Scenario: Model IDs preserved without normalization
- **GIVEN** An Agent model with Model="anthropic/claude-sonnet-4-20250514"
- **WHEN** Model is serialized to platform output
- **THEN** Model ID SHALL be preserved exactly as provided
- **AND** No normalization or mapping SHALL be applied
- **AND** User is responsible for providing correct format for target platform

### Requirement: Agent model supports common and platform-specific fields
Agent model includes required common fields and platform-specific fields for both Claude Code and OpenCode.

#### Scenario: Agent common fields
- **GIVEN** A new Agent model
- **WHEN** Common fields are populated
- **THEN** Name and Description are required for all platforms
- **AND** Model field is optional and platform-specific

#### Scenario: Agent OpenCode-specific fields
- **GIVEN** An Agent model with OpenCode-specific fields
- **WHEN** Mode, Temperature (*float64 pointer), MaxSteps, Hidden, Prompt, Disable are set
- **THEN** Fields are preserved in Go struct with proper YAML tags
- **AND** Temperature pointer distinguishes nil (not set) from 0.0 (deterministic mode)
- **AND** Templates handle field filtering based on target platform

#### Scenario: Agent Claude Code-specific fields
- **GIVEN** An Agent model with Claude Code-specific fields
- **WHEN** PermissionMode and Skills are set
- **THEN** Fields are preserved in Go struct with proper YAML tags
- **AND** Templates handle field filtering based on target platform

### Requirement: Temperature Field as *float64 Pointer to Distinguish "Not Set" from 0.0
The system SHALL use `*float64` pointer type for Temperature field to distinguish between nil (not set) and 0.0 (valid deterministic value).

#### Scenario: Temperature pointer handling
- **GIVEN** Temperature is *float64 pointer
- **WHEN** Validating and rendering
- **THEN** nil (not set) → field omitted from output (template checks `{{if .Temperature}}`)
- **AND** 0.0 (deterministic) → rendered as "temperature: 0.0"
- **AND** 1.0 (max randomness) → rendered as "temperature: 1.0"
- **AND** Validation SHALL check `agent.Temperature != nil` before accessing value
- **AND** Validation SHALL only apply range checks (0.0-1.0) when Temperature is not nil

### Requirement: Command model supports common and platform-specific fields
Command model includes required common fields and platform-specific fields.

#### Scenario: Command common fields
- **GIVEN** A new Command model
- **WHEN** Name and Description are populated
- **THEN** Fields are required for all platforms

#### Scenario: Command platform-specific fields
- **GIVEN** A Command model with platform-specific fields
- **WHEN** Subtask (OpenCode) or AllowedTools, Context, Agent, DisableModelInvocation (Claude Code) are set
- **THEN** Fields are preserved in Go struct with proper YAML tags
- **AND** Templates handle field filtering based on target platform

### Requirement: Skill model supports common and platform-specific fields
Skill model includes required common fields and platform-specific fields.

#### Scenario: Skill common fields
- **GIVEN** A new Skill model
- **WHEN** Name and Description are populated
- **THEN** Fields are required for all platforms

#### Scenario: Skill platform-specific fields
- **GIVEN** A Skill model with platform-specific fields
- **WHEN** License, Compatibility, Metadata, Hooks (OpenCode) or Model, Context, Agent, UserInvocable (Claude Code) are set
- **THEN** Fields are preserved in Go struct with proper YAML tags
- **AND** Templates handle field filtering based on target platform

### Requirement: Memory model supports dual storage modes
Memory model supports both file paths and narrative content fields.

#### Scenario: Memory dual storage modes
- **GIVEN** A Memory model
- **WHEN** Both Paths and Content are set
- **THEN** Both fields are preserved
- **AND** Templates can render either or both based on platform requirements

## MODIFIED Requirements

### Requirement: Platform-specific fields MUST be parseable from YAML
The system SHALL NOT use `yaml:"-"` tags to prevent parsing of platform-specific fields. All platform fields SHALL have proper YAML tags (`yaml:"field,omitempty"`) for full parseability.

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

### Requirement: Models must capture core document concepts without platform constraints
### Requirement: Tool configuration uses flat arrays (tools, disallowedTools, allowedTools) as single internal representation
### Requirement: Model IDs are user-provided, platform-specific strings (e.g., anthropic/claude-sonnet-4-20250514)

#### Scenario: Agent model has common required fields
- **GIVEN** A new Agent model
- **WHEN** Fields are populated with common values
- **THEN** Name and Description are required for all platforms
- **AND** Model field is optional and platform-specific

#### Scenario: Agent model supports platform-specific fields
- **GIVEN** An Agent model with OpenCode-specific fields
- **WHEN** Mode, Temperature, MaxSteps, Hidden, Prompt, Disable are set
- **THEN** Fields are preserved in Go struct
- **AND** Fields are parseable from YAML via proper yaml tags
- **AND** Templates handle field filtering based on target platform

#### Scenario: Agent model supports Claude Code-specific fields
- **GIVEN** An Agent model with Claude Code-specific fields
- **WHEN** PermissionMode and Skills are set
- **THEN** Fields are preserved in Go struct
- **AND** Fields are parseable from YAML via proper yaml tags
- **AND** Templates handle field filtering based on target platform

#### Scenario: Command model has common required fields
- **GIVEN** A new Command model
- **WHEN** Name and Description are populated
- **THEN** Fields are required for all platforms

#### Scenario: Command model supports platform-specific fields
- **GIVEN** A Command model with OpenCode Subtask field
- **WHEN** Subtask is set to true
- **THEN** Field is preserved in Go struct
- **AND** Field is parseable from YAML via proper yaml tags
- **AND** Templates handle field filtering based on target platform

#### Scenario: Skill model supports platform-specific fields
- **GIVEN** A Skill model with OpenCode-specific fields
- **WHEN** License, Compatibility, Metadata, Hooks are set
- **THEN** Fields are preserved in Go struct
- **AND** Fields are parseable from YAML via proper yaml tags
- **AND** Templates handle field filtering based on target platform

#### Scenario: Memory model supports dual storage modes
- **GIVEN** A Memory model
- **WHEN** Both Paths and Content are set
- **THEN** Both fields are preserved
- **AND** Templates can render either or both based on platform requirements

#### Scenario: Tool configuration uses flat arrays
- **GIVEN** An Agent model
- **WHEN** Tools and DisallowedTools are populated
- **THEN** Both fields are string slices
- **AND** Templates transform to platform-specific formats (arrays or objects)

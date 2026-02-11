# germinator-source-format Specification

## Purpose
Germinator source format serves as canonical source expressing configuration intent, not platform-specific fields.

## Requirements

### Requirement: Germinator format serves as canonical source

The system SHALL use a domain-driven canonical YAML format expressing configuration intent, NOT Claude Code format with platform-specific fields.

#### Scenario: Canonical format uses domain-driven fields

- **GIVEN** A canonical source YAML file
- **WHEN** The file is parsed
- **THEN** Common fields SHALL be present (Name, Description, Model, Content)
- **AND** PermissionPolicy enum SHALL be present (restrictive, balanced, permissive, analysis, unrestricted)
- **AND** Behavior object SHALL group settings (mode, temperature, maxSteps, prompt, hidden, disabled)
- **AND** Targets section SHALL contain platform-specific overrides
- **AND** Claude Code-specific fields (permissionMode, skills) SHALL NOT be at top level
- **AND** OpenCode-specific fields (mode, temperature, steps, hidden, prompt, disable) SHALL NOT be at top level (except in behavior)

#### Scenario: Tools use split lists with lowercase names

- **GIVEN** A canonical source YAML file
- **WHEN** The file is parsed
- **THEN** Tools array SHALL contain lowercase tool names
- **AND** DisallowedTools array SHALL contain lowercase tool names

#### Scenario: Model is simple string with full ID

- **GIVEN** A canonical source YAML file
- **WHEN** The file is parsed
- **THEN** Model SHALL be a string with provider/id format (e.g., "anthropic/claude-sonnet-4-20250514")
- **AND** No alias resolution or normalization SHALL occur
- **AND** Model ID SHALL be passed to output as-is

#### Scenario: Targets section for platform overrides

- **GIVEN** A canonical source YAML with targets section
- **WHEN** The file is parsed
- **THEN** targets.claude-code SHALL contain Claude Code-specific fields (skills array, disableModelInvocation bool)
- **AND** targets.opencode SHALL contain OpenCode-specific overrides (can override behavior object fields)
- **AND** Other platform keys in targets SHALL be supported for future extensibility
- **AND** Empty targets section SHALL NOT cause parsing errors

### Requirement: Temperature nil vs 0.0 distinction

This requirement applies to behavior.temperature pointer field in canonical format. Same logic applies (nil vs 0.0 distinction).

#### Scenario: Behavior temperature nil vs 0.0

- **GIVEN** Agent model with behavior.temperature as *float64 pointer
- **WHEN** behavior.temperature is nil
- **THEN** behavior.temperature field SHALL NOT be rendered in OpenCode output (user didn't set it)
- **AND** OpenCode will use model's default temperature
- **WHEN** behavior.temperature is 0.0
- **THEN** behavior.temperature field SHALL be rendered as `temperature: 0.0` in OpenCode output
- **AND** User explicitly requested deterministic low-temperature behavior
- **AND** Template rendering checks nil presence (not zero value)
- **AND** Distinction between unset and explicitly set to 0.0 is preserved

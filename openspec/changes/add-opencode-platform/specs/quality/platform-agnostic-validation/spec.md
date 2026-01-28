## ADDED Requirements

### Requirement: All models must implement Validate(platform string) method
### Requirement: Platform parameter is always required
### Requirement: Validation returns []error (all errors, not just first)
### Requirement: Common validation: required fields, data types, format constraints
### Requirement: Platform-specific validation: mode values, temperature ranges, skill name formats

#### Scenario: Validate requires platform parameter
- **GIVEN** Any model (Agent, Command, Skill, Memory)
- **WHEN** Validate("") is called with empty platform
- **THEN** Validation returns error "platform is required"
- **AND** No other validation occurs

#### Scenario: Validate returns multiple errors
- **GIVEN** An Agent model with missing name and description
- **WHEN** Validate(platform) is called
- **THEN** Validation returns []error containing both "name is required" and "description is required"
- **AND** Both errors are returned (not just first)

#### Scenario: Agent common validation checks required fields
- **GIVEN** An Agent model
- **WHEN** Name or Description is empty
- **THEN** Validation returns error "name is required" or "description is required"

#### Scenario: Agent skill name regex validation
- **GIVEN** A Skill model with name "invalid--name"
- **WHEN** Validate(platform) is called
- **THEN** Validation returns error about invalid skill name format
- **AND** Regex is ^[a-z0-9]+(-[a-z0-9]+)*$

#### Scenario: Agent name format validation
- **GIVEN** An Agent model with name "invalid--agent-name"
- **WHEN** Validate(platform) is called
- **THEN** Validation returns error about invalid name format
- **AND** Regex is ^[a-z0-9]+(-[a-z0-9]+)*$
- **AND** Applies to all platforms

#### Scenario: Agent platform-specific validation for OpenCode
- **GIVEN** An Agent model with mode="invalid", temperature=2.0, maxSteps=0
- **WHEN** Validate("opencode") is called
- **THEN** Validation returns errors for invalid mode, temperature out of range, maxSteps too low
- **AND** Valid modes: primary, subagent, all
- **AND** Temperature range: 0.0-1.0
- **AND** MaxSteps must be >= 1

#### Scenario: Agent platform-specific validation for Claude Code
- **GIVEN** An Agent model with model="invalid"
- **WHEN** Validate("claude-code") is called
- **THEN** Validation returns error for invalid model (if validation exists)
- **AND** Validation passes (model validation removed to support full platform-specific IDs)

#### Scenario: Unknown platform returns error
- **GIVEN** Any model
- **WHEN** Validate("nonexistent-platform") is called
- **THEN** Validation returns error "unknown platform: nonexistent-platform"
- **AND** Error message lists available platforms (claude-code, opencode)

#### Scenario: Memory validation requires paths or content
- **GIVEN** A Memory model with empty paths and empty content
- **WHEN** Validate(platform) is called
- **THEN** Validation returns error "paths or content is required"

#### Scenario: Validation passes with valid data
- **GIVEN** Any model with all required fields populated and valid values
- **WHEN** Validate(platform) is called with valid platform
- **THEN** Validation returns nil (no errors)

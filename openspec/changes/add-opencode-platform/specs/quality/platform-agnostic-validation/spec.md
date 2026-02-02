## ADDED Requirements

### Requirement: All models must implement Validate(platform string) method
All model validation methods require a platform parameter to apply platform-specific rules.

#### Scenario: Validate requires platform parameter
- **GIVEN** Any model (Agent, Command, Skill, Memory)
- **WHEN** Validate("") is called with empty platform
- **THEN** Validation SHALL return error "platform is required (available: claude-code, opencode)"
- **AND** No other validation SHALL occur

### Requirement: Platform parameter is always required
Validation functions must reject empty or missing platform values.

#### Scenario: Unknown platform returns error
- **GIVEN** Any model
- **WHEN** Validate("nonexistent-platform") is called
- **THEN** Validation SHALL return error "unknown platform: nonexistent-platform"
- **AND** Error message SHALL list available platforms (claude-code, opencode)

### Requirement: Validation returns []error (all errors, not just first)
Validation functions collect all validation errors rather than failing on the first error.

#### Scenario: Validate returns multiple errors
- **GIVEN** An Agent model with missing name and description
- **WHEN** Validate(platform) is called
- **THEN** Validation SHALL return []error containing both "name is required" and "description is required"

### Requirement: Common validation applies to all platforms
Required fields, data types, and format constraints are validated regardless of platform.

#### Scenario: Agent common validation
- **GIVEN** An Agent model
- **WHEN** Name or Description is empty
- **THEN** Validation SHALL return error "name is required" or "description is required"
- **AND** Agent name format SHALL be validated against regex `^[a-z0-9]+(-[a-z0-9]+)*$`

#### Scenario: Skill name format validation
- **GIVEN** A Skill model
- **WHEN** Validate(platform) is called with invalid name (e.g., "invalid--name")
- **THEN** Validation SHALL return error about invalid name format
- **AND** Regex SHALL be `^[a-z0-9]+(-[a-z0-9]+)*$`

#### Scenario: Memory validation requires paths or content
- **GIVEN** A Memory model
- **WHEN** paths=[], content="" (both empty)
- **THEN** Validation SHALL return error "paths or content is required"

### Requirement: Platform-specific validation applied by platform
Each platform has unique validation rules applied when that platform is specified.

#### Scenario: Agent OpenCode platform validation
- **GIVEN** An Agent with mode="invalid", temperature=2.0, steps=0
- **WHEN** Validate("opencode") is called
- **THEN** Validation SHALL call validateOpenCodeAgent and return errors for invalid mode, temperature out of range, steps too low
- **AND** Valid modes: primary, subagent, all
- **AND** Temperature range: 0.0-1.0
- **AND** Steps must be >= 1

#### Scenario: Agent Claude Code platform validation
- **GIVEN** An Agent model
- **WHEN** Validate("claude-code") is called
- **THEN** Validation SHALL apply Claude Code-specific rules (if any)
- **AND** Model field validation is removed (allows full platform-specific IDs)

#### Scenario: Validation passes with valid data
- **GIVEN** Any model with all required fields populated and valid values
- **WHEN** Validate(platform) is called with valid platform
- **THEN** Validation SHALL return nil (no errors)

---

### Requirement: Platform Constants to Eliminate Hardcoded Strings
The system SHALL define platform identifiers as package-level constants to avoid hardcoded string comparisons throughout the codebase.

#### Scenario: Platform constants defined
- **GIVEN** Platform constants
- **WHEN** Constants are defined
- **THEN** `PlatformClaudeCode = "claude-code"` SHALL be defined
- **AND** `PlatformOpenCode = "opencode"` SHALL be defined

#### Scenario: Validation functions use platform constants
- **GIVEN** Platform constants are defined
- **WHEN** Validation logic checks platform
- **THEN** Hardcoded string literals like "claude-code" SHALL NOT be used
- **AND** Platform constants SHALL be used instead
- **AND** Error messages SHALL reference constants

### Requirement: Shared ValidatePlatform helper function
The system SHALL extract platform requirement validation to a shared helper function to eliminate code duplication.

#### Scenario: ValidatePlatform validates platform parameter
- **GIVEN** A shared ValidatePlatform(platform string) function
- **WHEN** Called with empty platform
- **THEN** It SHALL return error: "platform is required (available: claude-code, opencode)"
- **AND** It SHALL use platform constants in error message

#### Scenario: ValidatePlatform handles unknown platforms
- **GIVEN** ValidatePlatform function
- **WHEN** Called with "invalid-platform"
- **THEN** It SHALL return error listing available platforms
- **AND** It SHALL use platform constants for comparison

#### Scenario: All model Validate methods use shared function
- **GIVEN** ValidatePlatform function exists
- **WHEN** Agent.Validate(platform), Command.Validate(platform), Skill.Validate(platform), Memory.Validate(platform) are called
- **THEN** All SHALL call ValidatePlatform(platform) for platform requirement checks
- **AND** 120+ lines of duplicated code SHALL be eliminated

### Requirement: Architecture Separation - Platform Validation in Services
The system SHALL move OpenCode-specific validation functions from models package to services package for proper separation of concerns.

#### Scenario: ValidateOpenCodeAgent in services package
- **GIVEN** OpenCode validation functions
- **WHEN** OpenCode-specific validation is needed
- **THEN** ValidateOpenCodeAgent function SHALL reside in internal/services/transformer.go
- **AND** NOT in internal/models/models.go

#### Scenario: Model Validate methods delegate to services
- **GIVEN** OpenCode validation in services package
- **WHEN** Agent.Validate("opencode") is called
- **THEN** It SHALL call services.ValidateOpenCodeAgent(*Agent) for business rule validation
- **AND** Model validation (structure) remains in models package
- **AND** Business rule validation (temperature range, mode values) is in services package

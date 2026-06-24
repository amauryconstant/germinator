# platform-adapters Specification

## Purpose

Bidirectional conversion between canonical format and platform-specific formats (Claude Code, OpenCode), handling field name conversions, permission policy mapping, and platform-specific field filtering.

## Requirements

### Requirement: Adapter Interface Definition

The system SHALL define a common adapter interface for platform-specific conversion logic.

#### Scenario: Adapter has canonical conversion methods

- **GIVEN** A platform adapter implementation
- **WHEN** Inspected for available methods
- **THEN** ToCanonical() method SHALL exist (parse platform format into canonical)
- **AND** FromCanonical() method SHALL exist (render canonical to platform format)
- **AND** Methods SHALL accept document structs and return converted structs or strings

#### Scenario: Adapter provides permission mapping

- **GIVEN** A platform adapter implementation
- **WHEN** Inspected for permission mapping capability
- **THEN** PermissionPolicyToPlatform() method SHALL exist
- **AND** Method SHALL convert canonical PermissionPolicy enum to platform-specific value
- **AND** Method SHALL use mapping table (no complex logic)

#### Scenario: Adapter provides field name conversion

- **GIVEN** A platform adapter implementation
- **WHEN** Inspected for field name conversion capability
- **THEN** ConvertToolNameCase() method SHALL exist for platform-specific case conversion
- **AND** Claude Code adapter SHALL convert to PascalCase
- **AND** OpenCode adapter SHALL convert to lowercase
- **AND** Method SHALL use consistent rules (no special cases)

#### Scenario: PermissionAction enum definition

- **GIVEN** PermissionAction enum is defined in adapters package
- **WHEN** Inspected
- **THEN** PermissionAction SHALL have three string constants: Allow, Ask, Deny
- **AND** Values SHALL be used instead of string literals in code
- **AND** Values SHALL match OpenCode permission action strings ("allow", "ask", "deny")

### Requirement: Claude Code Adapter

The system SHALL implement Claude Code adapter converting between canonical and Claude Code formats.

#### Scenario: Claude Code adapter parses Claude Code YAML to canonical

- **GIVEN** A Claude Code agent YAML with permissionMode enum and tools array (PascalCase)
- **WHEN** ClaudeCodeAdapter.ToCanonical() is called
- **THEN** Output SHALL be canonical Agent struct with PermissionPolicy enum
- **AND** Tools SHALL be lowercase (canonical uses lowercase)
- **AND** Skills SHALL be placed in targets.claude-code section
- **AND** Field names SHALL match canonical structure (name, description, tools, behavior, model)

#### Scenario: Claude Code adapter renders canonical to Claude Code format

- **GIVEN** A canonical Agent struct with PermissionPolicy enum and tools array (lowercase)
- **WHEN** ClaudeCodeAdapter.FromCanonical() is called
- **THEN** Output SHALL be Claude Code YAML with permissionMode enum (converted from policy)
- **AND** Tools SHALL be PascalCase (converted from canonical lowercase)
- **AND** Skills SHALL be rendered from targets.claude-code.skills array
- **AND** Field names SHALL match Claude Code structure (permissionMode, skills, disallowedTools)

#### Scenario: Claude Code adapter converts permission policies

- **GIVEN** Canonical PermissionPolicy value (restrictive, balanced, permissive, analysis, unrestricted)
- **WHEN** ClaudeCodeAdapter.PermissionPolicyToPlatform() is called
- **THEN** restrictive SHALL convert to "default"
- **AND** balanced SHALL convert to "acceptEdits"
- **AND** permissive SHALL convert to "dontAsk"
- **AND** analysis SHALL convert to "plan"
- **AND** unrestricted SHALL convert to "bypassPermissions"

#### Scenario: Claude Code adapter converts tool case

- **GIVEN** Canonical tool name (lowercase, e.g., "bash", "read")
- **WHEN** ClaudeCodeAdapter.ConvertToolNameCase() is called
- **THEN** Output SHALL be PascalCase ("Bash", "Read")
- **AND** Conversion SHALL be consistent for all built-in tools
- **AND** Special characters in tool names SHALL be preserved

#### Scenario: Claude Code adapter handles targets section

- **GIVEN** A canonical Agent struct with targets.claude-code section
- **WHEN** ClaudeCodeAdapter.FromCanonical() is called
- **THEN** targets.claude-code.skills SHALL be rendered as skills array at top level
- **AND** targets.claude-code.disableModelInvocation SHALL be rendered as disable-model-invocation field
- **AND** Other platform targets sections SHALL be omitted
- **AND** Empty targets.claude-code section SHALL NOT add Claude Code fields

### Requirement: OpenCode Adapter

The system SHALL implement OpenCode adapter converting between canonical and OpenCode formats.

#### Scenario: OpenCode adapter parses OpenCode YAML to canonical

- **GIVEN** An OpenCode agent YAML with mode field and tools map (lowercase)
- **WHEN** OpenCodeAdapter.ToCanonical() is called
- **THEN** Output SHALL be canonical Agent struct
- **AND** Tools map SHALL be split into tools and disallowedTools arrays (true values → tools array, false values → disallowedTools array)
- **AND** Mode SHALL be placed in behavior.mode field
- **AND** Temperature, MaxSteps, Hidden, Prompt, Disabled SHALL be placed in behavior object
- **AND** Field names SHALL match canonical structure

#### Scenario: OpenCode adapter renders canonical to OpenCode format

- **GIVEN** A canonical Agent struct with Behavior object and tools arrays
- **WHEN** OpenCodeAdapter.FromCanonical() is called
- **THEN** Output SHALL be OpenCode YAML with mode field from behavior.mode
- **AND** Tools SHALL be converted to map format (lowercase: boolean)
- **AND** DisallowedTools SHALL be converted to map format (lowercase: false)
- **AND** Behavior fields (temperature, maxSteps, hidden, prompt, disabled) SHALL be rendered at top level
- **AND** PermissionPolicy SHALL be converted to permission object (via PermissionPolicyToPlatform())

#### Scenario: OpenCode adapter converts permission policies

- **GIVEN** Canonical PermissionPolicy value (restrictive, balanced, permissive, analysis, unrestricted)
- **WHEN** OpenCodeAdapter.PermissionPolicyToPlatform() is called
- **THEN** restrictive SHALL return PermissionMap{Edit: Ask, Bash: Ask, Read: Ask, ...}
- **AND** balanced SHALL return PermissionMap{Edit: Allow, Bash: Ask, Read: Allow, ...}
- **AND** permissive SHALL return PermissionMap{Edit: Allow, Bash: Allow, Read: Allow, ...}
- **AND** analysis SHALL return PermissionMap{Edit: Deny, Bash: Deny, Read: Allow, ...}
- **AND** unrestricted SHALL return PermissionMap{Edit: Allow, Bash: Allow, Read: Allow, ...}

#### Scenario: OpenCode adapter maintains tool case

- **GIVEN** Canonical tool name (lowercase, e.g., "bash", "read")
- **WHEN** OpenCodeAdapter.ConvertToolNameCase() is called
- **THEN** Output SHALL be lowercase ("bash", "read")
- **AND** No case conversion SHALL occur (already lowercase)
- **AND** Method SHALL be provided for consistency with Claude Code adapter

#### Scenario: OpenCode adapter splits tool lists

- **GIVEN** Canonical Agent with tools ["bash", "read"] and disallowedTools ["write", "execute"]
- **WHEN** OpenCodeAdapter.FromCanonical() is called
- **THEN** Output SHALL be tools map: {"bash": true, "read": true, "write": false, "execute": false}
- **AND** True values SHALL come from tools array
- **AND** False values SHALL come from disallowedTools array
- **AND** Tool appearing in both SHALL use tools array value (allowed takes precedence)

#### Scenario: OpenCode adapter flattens behavior object

- **GIVEN** Canonical Agent with behavior.mode="primary", behavior.temperature=0.3, behavior.maxSteps=25
- **WHEN** OpenCodeAdapter.FromCanonical() is called
- **THEN** Output SHALL have mode: primary at top level
- **AND** Output SHALL have temperature: 0.3 at top level
- **AND** Output SHALL have maxSteps: 25 at top level (renamed from behavior.maxSteps)
- **AND** Nested behavior object SHALL NOT be present in output

#### Scenario: OpenCode adapter handles targets section

- **GIVEN** A canonical Agent struct with targets.opencode section
- **WHEN** OpenCodeAdapter.FromCanonical() is called
- **THEN** targets.opencode section values SHALL override canonical behavior object fields
- **AND** Other platform targets sections SHALL be omitted
- **AND** Empty targets.opencode section SHALL NOT override behavior fields

### Requirement: Shared Helper Functions

The adapters package SHALL provide shared helper functions for common conversion logic.

#### Scenario: Tool name case conversion

- **GIVEN** Tool name in any case (PascalCase, camelCase, lowercase, UPPER_CASE)
- **WHEN** Helper functions are used
- **THEN** ToPascalCase() SHALL convert "bash" → "Bash"
- **AND** ToLowerCase() SHALL convert "Bash" → "bash"
- **AND** Functions SHALL handle multiple words (e.g., "webFetch" → "WebFetch" / "webfetch")

#### Scenario: Permission mapping data structure

- **GIVEN** PermissionPolicy mapping table is defined
- **WHEN** Inspected
- **THEN** Table SHALL be map[PermissionPolicy]PermissionMapping struct
- **AND** PermissionMapping SHALL contain ClaudeCode string field
- **AND** PermissionMapping SHALL contain OpenCode PermissionMap field (map[string]PermissionAction enum values)
- **AND** Mapping SHALL be initialized at package level

#### Scenario: OpenCode permission map structure

- **GIVEN** OpenCode PermissionMap is defined
- **WHEN** Inspected
- **THEN** Map SHALL contain all tool permissions (edit, bash, read, grep, glob, list, webfetch, websearch)
- **AND** Values SHALL be PermissionAction enum values (Allow, Ask, Deny)
- **AND** Default action SHALL be provided for each policy

### Requirement: Adapter Error Handling

Adapters SHALL return clear errors for conversion failures with contextual information.

#### Scenario: Invalid permission policy

- **GIVEN** Unknown PermissionPolicy value in canonical document
- **WHEN** Adapter.PermissionPolicyToPlatform() is called
- **THEN** Error SHALL be returned with unknown policy value
- **AND** Error message SHALL list valid policy values
- **AND** Conversion SHALL NOT proceed

#### Scenario: Tool name with special characters

- **GIVEN** Canonical tool name with hyphens or numbers (e.g., "web-search", "lsp")
- **WHEN** Adapter.ConvertToolNameCase() is called
- **THEN** Conversion SHALL preserve hyphens and numbers
- **AND** PascalCase SHALL handle consecutive capitals appropriately (WebSearch)
- **AND** No error SHALL be raised (valid tool names)

#### Scenario: Missing required fields in canonical document

- **GIVEN** Canonical Agent missing required name or description fields
- **WHEN** Adapter.FromCanonical() is called
- **THEN** Error SHALL be returned indicating missing field
- **AND** Error message SHALL specify which field is required
- **AND** Output SHALL NOT be generated

### Requirement: Adapter Testing

Adapters SHALL have comprehensive unit test coverage for all conversion methods.

#### Scenario: Test Claude Code adapter conversion

- **GIVEN** Test suite with Claude Code and canonical fixtures
- **WHEN** Tests are run
- **THEN** ToCanonical() SHALL be tested with various Claude Code formats
- **AND** FromCanonical() SHALL be tested with various canonical agents
- **AND** Permission policy conversion SHALL be tested for all enum values
- **AND** Tool case conversion SHALL be tested with edge cases

#### Scenario: Test OpenCode adapter conversion

- **GIVEN** Test suite with OpenCode and canonical fixtures
- **WHEN** Tests are run
- **THEN** ToCanonical() SHALL be tested with various OpenCode formats
- **AND** FromCanonical() SHALL be tested with various canonical agents
- **AND** Permission policy conversion SHALL be tested for all enum values
- **AND** Tool list splitting SHALL be tested with overlapping tools
- **AND** Behavior flattening SHALL be tested with various field combinations

### Requirement: Extensibility for New Platforms

Adapter pattern SHALL support easy addition of new platforms without core code changes.

#### Scenario: New platform adapter structure

- **GIVEN** Need to add a new platform (e.g., "future-platform")
- **WHEN** Implementing new adapter
- **THEN** New adapter SHALL implement same interface as ClaudeCodeAdapter and OpenCodeAdapter
- **AND** ToCanonical() SHALL parse future-platform YAML into canonical format
- **AND** FromCanonical() SHALL render canonical format to future-platform YAML
- **AND** No changes SHALL be required to core models or validation logic
- **AND** Only new template set SHALL be added for the platform

#### Scenario: Platform-specific permission mapping

- **GIVEN** New platform with permission system different from Claude Code/OpenCode
- **WHEN** Implementing PermissionPolicyToPlatform() for new platform
- **THEN** Method SHALL map 5 canonical policies to platform-specific values
- **AND** Mapping SHALL be documented in adapter code
- **AND** PermissionPolicy enum SHALL NOT be extended (canonical stays stable)

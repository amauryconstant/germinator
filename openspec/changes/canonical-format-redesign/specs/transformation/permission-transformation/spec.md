## MODIFIED Requirements

### Requirement: Claude Code to OpenCode Permission Transformation

The system SHALL provide simple permission policy mapping tables converting canonical PermissionPolicy enum to platform-specific values, NOT complex transformation functions.

#### Scenario: Map restrictive policy to Claude Code

- **GIVEN** Canonical PermissionPolicy value is "restrictive"
- **WHEN** Adapter.PermissionPolicyToPlatform("claude-code") is called
- **THEN** Output SHALL be "default" string
- **AND** No complex logic SHALL be used
- **AND** Mapping is table lookup: restrictve → default

#### Scenario: Map restrictive policy to OpenCode

- **GIVEN** Canonical PermissionPolicy value is "restrictive"
- **WHEN** Adapter.PermissionPolicyToPlatform("opencode") is called
- **THEN** Output SHALL be PermissionMap{Edit: Ask, Bash: Ask, Read: Ask, ...}
- **AND** No complex logic SHALL be used
- **AND** Mapping is table lookup: restrictve → permission map

#### Scenario: Map balanced policy to Claude Code

- **GIVEN** Canonical PermissionPolicy value is "balanced"
- **WHEN** Adapter.PermissionPolicyToPlatform("claude-code") is called
- **THEN** Output SHALL be "acceptEdits" string
- **AND** Mapping is table lookup: balanced → acceptEdits

#### Scenario: Map balanced policy to OpenCode

- **GIVEN** Canonical PermissionPolicy value is "balanced"
- **WHEN** Adapter.PermissionPolicyToPlatform("opencode") is called
- **THEN** Output SHALL be PermissionMap{Edit: Allow, Bash: Ask, Read: Allow, ...}
- **AND** Mapping is table lookup: balanced → permission map

#### Scenario: Map permissive policy to Claude Code

- **GIVEN** Canonical PermissionPolicy value is "permissive"
- **WHEN** Adapter.PermissionPolicyToPlatform("claude-code") is called
- **THEN** Output SHALL be "dontAsk" string
- **AND** Mapping is table lookup: permissive → dontAsk

#### Scenario: Map permissive policy to OpenCode

- **GIVEN** Canonical PermissionPolicy value is "permissive"
- **WHEN** Adapter.PermissionPolicyToPlatform("opencode") is called
- **THEN** Output SHALL be PermissionMap{Edit: Allow, Bash: Allow, Read: Allow, ...}
- **AND** Mapping is table lookup: permissive → permission map

#### Scenario: Map analysis policy to Claude Code

- **GIVEN** Canonical PermissionPolicy value is "analysis"
- **WHEN** Adapter.PermissionPolicyToPlatform("claude-code") is called
- **THEN** Output SHALL be "plan" string
- **AND** Mapping is table lookup: analysis → plan
- **AND** "analysis" maps to "plan" because both represent read-only exploration mode

#### Scenario: Map analysis policy to OpenCode

- **GIVEN** Canonical PermissionPolicy value is "analysis"
- **WHEN** Adapter.PermissionPolicyToPlatform("opencode") is called
- **THEN** Output SHALL be PermissionMap{Edit: Deny, Bash: Deny, Read: Allow, ...}
- **AND** Mapping is table lookup: analysis → permission map

#### Scenario: Map unrestricted policy to Claude Code

- **GIVEN** Canonical PermissionPolicy value is "unrestricted"
- **WHEN** Adapter.PermissionPolicyToPlatform("claude-code") is called
- **THEN** Output SHALL be "bypassPermissions" string
- **AND** Mapping is table lookup: unrestricted → bypassPermissions
- **AND** "unrestricted" maps to "bypassPermissions" because both represent "allow all without restrictions"

#### Scenario: Map unrestricted policy to OpenCode

- **GIVEN** Canonical PermissionPolicy value is "unrestricted"
- **WHEN** Adapter.PermissionPolicyToPlatform("opencode") is called
- **THEN** Output SHALL be PermissionMap{Edit: Allow, Bash: Allow, Read: Allow, ...}
- **AND** Mapping is table lookup: unrestricted → permission map

#### Scenario: Map unknown permission policy

- **GIVEN** Canonical PermissionPolicy value is not valid enum value
- **WHEN** Adapter.PermissionPolicyToPlatform(platform) is called for any platform
- **THEN** Error SHALL be returned
- **AND** Error message SHALL list valid policy values
- **AND** Conversion SHALL NOT proceed

### Requirement: PermissionAction Enum Definition

The system SHALL define PermissionAction enum for type-safe permission values in OpenCode format.

#### Scenario: PermissionAction enum has three values

- **GIVEN** PermissionAction enum is defined
- **WHEN** Inspected
- **THEN** PermissionAction SHALL have three string constants: Allow, Ask, Deny
- **AND** Values SHALL be used instead of string literals in code
- **AND** Values SHALL match OpenCode permission action strings ("allow", "ask", "deny")

### Requirement: Permission Transformation Uses Mapping Table

Permission transformation SHALL use struct/map lookup tables instead of complex switch statements or YAML generation functions.

#### Scenario: Mapping table structure

- **GIVEN** PermissionPolicy mapping table is defined
- **WHEN** Inspected
- **THEN** Table SHALL be map[PermissionPolicy]PermissionMapping struct
- **AND** PermissionMapping SHALL contain ClaudeCode string field
- **AND** PermissionMapping SHALL contain OpenCode PermissionMap field (map[string]PermissionAction)
- **AND** Mapping SHALL be initialized at package level

#### Scenario: OpenCode permission map structure

- **GIVEN** OpenCode PermissionMap is defined
- **WHEN** Inspected
- **THEN** Map SHALL contain all tool permissions (edit, bash, read, grep, glob, list, webfetch, websearch)
- **AND** Values SHALL be PermissionAction enum (Allow, Ask, Deny)
- **AND** Default action SHALL be provided for each policy

## REMOVED Requirements

### Requirement: Claude Code to OpenCode Permission Transformation

**Reason**: Replaced with simple permission policy mapping table in adapters. Old approach used 103-line transformPermissionMode() function with complex YAML string building.

**Migration**:
- Use adapter.PermissionPolicyToPlatform(platform) method for conversion
- Return platform-specific value directly from mapping table
- Remove transformPermissionMode() function from core package
- Remove all tests for transformPermissionMode() function
- Create tests for adapter.PermissionPolicyToPlatform() method

### Requirement: Permission Transformation Limitations

**Reason**: These limitations are documented but transformation approach (mapping table vs function) doesn't change them. Limitations are due to fundamental differences between platform permission systems.

**Migration**:
- Keep limitation documentation in new specs
- Document that fine-grained permissions (command-level rules) are delayed to future change
- Maintain same coverage (8 tools: edit, bash, read, grep, glob, list, webfetch, websearch)
- No change to limitation descriptions or documentation approach

### Requirement: Expanded Permission Coverage

**Reason**: Same 8 tools covered, mapping table approach doesn't change coverage. Only implementation changes from function to table lookup.

**Migration**:
- Keep test coverage for all 8 tools
- Ensure mapping table has all tools defined
- No change to tools covered or test scenarios

### Requirement: Permission Transformation Testing

**Reason**: Testing approach changes from testing transformPermissionMode() function to testing adapter.PermissionPolicyToPlatform() methods.

**Migration**:
- Create unit tests for ClaudeCodeAdapter.PermissionPolicyToPlatform("claude-code")
- Create unit tests for OpenCodeAdapter.PermissionPolicyToPlatform("opencode")
- Test all 5 policy values (restrictive, balanced, permissive, analysis, unrestricted)
- Test unknown policy value error case
- Remove tests for transformPermissionMode() function

## ADDED Requirements

### Requirement: Render YAML frontmatter with agent configuration
OpenCode agent template MUST render YAML frontmatter with all Agent fields in correct format.

#### Scenario: Minimal agent transformation
- **GIVEN** Agent with name="code-reviewer", description="Reviews code", content="You are a reviewer..."
- **WHEN** Rendered to OpenCode format
- **THEN** Output does NOT contain name field (OpenCode uses filename)
- **AND** Output contains description in YAML frontmatter
- **AND** mode field is NOT rendered (empty, uses platform default)
- **AND** No tools map (empty tools)
- **AND** Content follows frontmatter

### Requirement: Transform tools/disallowedTools arrays to {tool: true|false} map with lowercase names
Tools and disallowedTools arrays transform to OpenCode object format with lowercase tool names.

#### Scenario: Agent with tools transformation
- **GIVEN** Agent with tools=["Bash", "Grep", "Read"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains tools map: {bash: true, grep: true, read: true}
- **AND** Tool names are converted to lowercase for OpenCode
- **AND** Each tool is on separate line with ": true" suffix

#### Scenario: Agent with disallowedTools transformation
- **GIVEN** Agent with disallowedTools=["Write", "Edit"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains disallowedTools map: {write: false, edit: false}
- **AND** Tool names are converted to lowercase for OpenCode
- **AND** Each tool is on separate line with ": false" suffix

#### Scenario: Agent with mixed tools transformation
- **GIVEN** Agent with tools=["Bash"], disallowedTools=["Write"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains tools map: {bash: true}
- **AND** Output contains disallowedTools map: {write: false}
- **AND** All tool names are converted to lowercase

### Requirement: Transform permissionMode to permission object
Claude Code permissionMode enum transforms to OpenCode permission object using transformPermissionMode() function.

#### Scenario: Permission mode transformations
- **GIVEN** Agent with permissionMode
- **WHEN** Rendered to OpenCode format
- **THEN** permissionMode="acceptEdits" → {"edit": {"*": "allow"}, "bash": {"*": "ask"}}
- **AND** permissionMode="default" → {"edit": {"*": "ask"}, "bash": {"*": "ask"}}
- **AND** permissionMode="dontAsk" → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
- **AND** permissionMode="bypassPermissions" → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
- **AND** permissionMode="plan" → {"edit": {"*": "deny"}, "bash": {"*": "deny"}}

### Requirement: Include OpenCode-specific agent fields
OpenCode agent fields (mode, temperature, steps, hidden, prompt, disable) are rendered when set.

#### Scenario: Agent mode field
- **GIVEN** Agent with mode="subagent"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "mode: subagent"
- **AND** No default value used (explicit mode is preserved)

#### Scenario: Agent with mode empty
- **GIVEN** Agent with mode="" (empty)
- **WHEN** Rendered to OpenCode format
- **THEN** mode field SHALL NOT be rendered in output
- **AND** OpenCode uses platform default (implicitly "all")
- **AND** This allows idiomatically omitting optional fields

#### Scenario: Agent temperature field
- **GIVEN** Agent with temperature=0.3 (non-zero)
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "temperature: 0.3"

#### Scenario: Agent temperature nil vs 0.0
- **GIVEN** Agent with Temperature field as *float64 pointer
- **WHEN** Temperature is nil (not set)
- **THEN** Temperature field SHALL NOT be rendered in output
- **AND** OpenCode uses model's default temperature
- **WHEN** Temperature is 0.0 (explicitly set)
- **THEN** Output contains "temperature: 0.0"
- **AND** User explicitly requested deterministic low-temperature behavior
- **AND** Template checks for nil presence, not zero value

#### Scenario: Agent steps field
- **GIVEN** Agent with steps=50
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "steps: 50"

#### Scenario: Agent hidden, prompt, disable fields
- **GIVEN** Agent with hidden=true, prompt="You are expert...", disable=true
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "hidden: true" and "disable: true" when true
- **AND** Output does NOT contain hidden/disable when false
- **AND** This prevents redundant output of false values
- **AND** Prompt field is rendered when non-empty string

### Requirement: Render content after YAML frontmatter
Agent markdown content follows YAML frontmatter separator.

#### Scenario: Content rendering
- **GIVEN** Agent with content="You are a code reviewer..."
- **WHEN** Rendered to OpenCode format
- **THEN** Content SHALL follow "---" separator after YAML frontmatter

### Requirement: OpenCode agent template MUST NOT output name field
The OpenCode agent template SHALL NOT output `name` field as OpenCode uses filename as identifier.

#### Scenario: Name field omitted in OpenCode template
- **GIVEN** Agent with Name="code-reviewer"
- **WHEN** Rendering to OpenCode platform
- **THEN** Template SHALL NOT output name field in frontmatter
- **AND** OpenCode uses filename (e.g., code-reviewer.md) as agent identifier
- **AND** Claude Code template outputs name field in frontmatter

### Requirement: Omit Claude Code-specific fields
OpenCode agent template omits Claude Code-specific fields (skills list).

#### Scenario: Skills list omitted for OpenCode
- **GIVEN** Agent with skills=["skill-creator", "refactoring"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output does NOT contain skills field
- **AND** Skills are independent documents in OpenCode

### Requirement: Full model ID preservation
Model IDs are preserved exactly as provided without transformation.

#### Scenario: Agent with full model ID
- **GIVEN** Agent with model="anthropic/claude-sonnet-4-20250514"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "model: anthropic/claude-sonnet-4-20250514"
- **AND** Model ID is preserved exactly as provided

---

### Requirement: Tool Name Case Transformation for OpenCode
The system SHALL convert tool names from any case (PascalCase, camelCase, etc.) to lowercase when rendering OpenCode output using Sprig library functions.

#### Scenario: Tool names converted to lowercase for OpenCode
- **GIVEN** Agent with Tools: ["Bash", "Read", "Edit"] or Command with AllowedTools: ["bash", "Write", "GREP"]
- **WHEN** Rendering to OpenCode platform
- **THEN** Tool names SHALL be lowercase in output (bash, read, edit) using Sprig's `lower` function
- **AND** Claude Code platform SHALL preserve original case (Bash, Read, Edit)
- **AND** Template SHALL use `{{range .Tools}}{{. | lower}}: true{{end}}` for OpenCode output

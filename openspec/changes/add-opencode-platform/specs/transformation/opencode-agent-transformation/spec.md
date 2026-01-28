## ADDED Requirements

### Requirement: Render YAML frontmatter with agent configuration
### Requirement: Transform tools/disallowedTools arrays to {tool: true|false} map
### Requirement: Transform permissionMode to permission object using transformPermissionModeToOpenCode()
### Requirement: Include OpenCode-specific fields: mode, temperature, maxSteps, hidden, prompt, disable
### Requirement: Omit Claude Code-specific fields: skills list
### Requirement: Default mode to "all" if empty
### Requirement: Render content after YAML frontmatter

#### Scenario: Minimal agent transformation
- **GIVEN** Agent with name="code-reviewer", description="Reviews code", content="You are a reviewer..."
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains YAML frontmatter with name and description
- **AND** mode defaults to "all"
- **AND** No tools map (empty tools)
- **AND** Content follows frontmatter

#### Scenario: Agent with tools transformation
- **GIVEN** Agent with tools=["bash", "grep", "read"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains tools map: {bash: true, grep: true, read: true}
- **AND** Each tool is on separate line with ": true" suffix

#### Scenario: Agent with disallowedTools transformation
- **GIVEN** Agent with disallowedTools=["write", "edit"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains disallowedTools map: {write: false, edit: false}
- **AND** Each tool is on separate line with ": false" suffix

#### Scenario: Agent with mixed tools transformation
- **GIVEN** Agent with tools=["bash"], disallowedTools=["write"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains tools map: {bash: true}
- **AND** Output contains disallowedTools map: {write: false}

#### Scenario: Agent with permission mode transformation
- **GIVEN** Agent with permissionMode="acceptEdits"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains permission object
- **AND** Permission maps to: {"edit": {"*": "allow"}, "bash": {"*": "ask"}}
- **AND** transformPermissionModeToOpenCode() function is called

#### Scenario: Agent with permission mode "default"
- **GIVEN** Agent with permissionMode="default"
- **WHEN** Rendered to OpenCode format
- **THEN** Permission maps to: {"edit": {"*": "ask"}, "bash": {"*": "ask"}}

#### Scenario: Agent with permission mode "dontAsk"
- **GIVEN** Agent with permissionMode="dontAsk"
- **WHEN** Rendered to OpenCode format
- **THEN** Permission maps to: {"edit": {"*": "deny"}, "bash": {"*": "deny"}}

#### Scenario: Agent with OpenCode mode field
- **GIVEN** Agent with mode="subagent"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "mode: subagent"
- **AND** No default value used (explicit mode is preserved)

#### Scenario: Agent with temperature field
- **GIVEN** Agent with temperature=0.3
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "temperature: 0.3"

#### Scenario: Agent with maxSteps field
- **GIVEN** Agent with maxSteps=50
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "maxSteps: 50"

#### Scenario: Agent with hidden field
- **GIVEN** Agent with hidden=true
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "hidden: true"

#### Scenario: Agent with prompt field
- **GIVEN** Agent with prompt="You are expert..."
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "prompt: \"You are expert...\""

#### Scenario: Agent with disable field
- **GIVEN** Agent with disable=true
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "disable: true"

#### Scenario: Agent with full model ID
- **GIVEN** Agent with model="anthropic/claude-sonnet-4-20250514"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "model: anthropic/claude-sonnet-4-20250514"
- **AND** Model ID is preserved exactly as provided

#### Scenario: Agent omits Claude Code skills list
- **GIVEN** Agent with skills=["skill-creator", "refactoring"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain skills field
- **AND** No warning is logged (silent skip)

#### Scenario: Agent with all fields transformation
- **GIVEN** Agent with all fields populated (name, description, tools, disallowedTools, permissionMode, mode, temperature, maxSteps, hidden, prompt, disable, model, content)
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains all fields in proper order
- **AND** YAML frontmatter is valid
- **AND** Content follows "---" separator
- **AND** Skills list is omitted

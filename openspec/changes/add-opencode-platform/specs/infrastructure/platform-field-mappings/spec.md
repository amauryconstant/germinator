## ADDED Requirements

### Requirement: Document all Agent, Command, Skill, Memory fields
### Requirement: Indicate field type: common, Claude Code-specific, OpenCode-specific
### Requirement: Document transformation logic for non-trivial mappings
### Requirement: Document skipped fields (Claude Code → OpenCode)

#### Scenario: Agent common fields map directly
- **GIVEN** Agent model
- **WHEN** Comparing Claude Code and OpenCode schemas
- **THEN** Name, Description map directly
- **AND** Model maps directly (user-provided platform-specific ID)

#### Scenario: Agent tool configuration transformation
- **GIVEN** Agent model with Tools and DisallowedTools
- **WHEN** Transforming to OpenCode
- **THEN** Tools array → {tool: true} map
- **AND** DisallowedTools array → {tool: false} map
- **AND** Tool names are converted to lowercase for OpenCode using Sprig's `lower` function (Bash → bash)
- **AND** Claude Code format: flat arrays (tools, disallowedTools) with original case preserved

#### Scenario: Agent permission mode transformation
- **GIVEN** Agent model with PermissionMode
- **WHEN** Transforming to OpenCode
- **THEN** PermissionMode enum → permission object
- **AND** transformPermissionModeToOpenCode() function handles mapping
- **AND** default → {"edit": {"*": "ask"}, "bash": {"*": "ask"}}
- **AND** acceptEdits → {"edit": {"*": "allow"}, "bash": {"*": "ask"}}
- **AND** dontAsk → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
- **AND** bypassPermissions → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
- **AND** plan → {"edit": {"*": "deny"}, "bash": {"*": "deny"}}
- **AND** Note: Basic approximation only, command-level rules not supported

#### Scenario: Agent skipped fields (Claude Code → OpenCode)
- **GIVEN** Agent model with Skills list
- **WHEN** Transforming to OpenCode
- **THEN** Skills field is skipped (no OpenCode equivalent)
- **AND** Skills are independent documents in OpenCode

#### Scenario: Agent OpenCode-specific fields
- **GIVEN** Agent model with Mode, Temperature, MaxSteps, Hidden, Prompt, Disable
- **WHEN** Serializing to OpenCode
- **THEN** All fields are included if set
- **AND** Mode defaults to "all" if empty
- **AND** Temperature must be 0.0-1.0
- **AND** MaxSteps must be >= 1
- **AND** Fields are omitted from Claude Code serialization

#### Scenario: Command common fields map directly
- **GIVEN** Command model
- **WHEN** Comparing Claude Code and OpenCode schemas
- **THEN** Name, Description map directly

#### Scenario: Command skipped fields (Claude Code → OpenCode)
- **GIVEN** Command model with AllowedTools, ArgumentHint, Context, DisableModelInvocation
- **WHEN** Transforming to OpenCode
- **THEN** All fields are skipped (no OpenCode equivalent)
- **AND** Note: OpenCode doesn't support allowedTools lists

#### Scenario: Command DisallowedTools field
- **GIVEN** Command model with DisallowedTools field populated
- **WHEN** Transforming to OpenCode
- **THEN** DisallowedTools is preserved in model
- **AND** Field is omitted from OpenCode serialization (no current equivalent)
- **AND** Field is preserved in Claude Code serialization

#### Scenario: Command OpenCode-specific fields
- **GIVEN** Command model with Subtask
- **WHEN** Serializing to OpenCode
- **THEN** Subtask is included if set
- **AND** Field is omitted from Claude Code serialization

#### Scenario: Skill common fields map directly
- **GIVEN** Skill model
- **WHEN** Comparing Claude Code and OpenCode schemas
- **THEN** Name, Description map directly

#### Scenario: Skill skipped fields (Claude Code → OpenCode)
- **GIVEN** Skill model with AllowedTools, UserInvocable
- **WHEN** Transforming to OpenCode
- **THEN** Fields are skipped (no direct OpenCode equivalent)

#### Scenario: Skill OpenCode-specific fields
- **GIVEN** Skill model with License, Compatibility, Metadata, Hooks
- **WHEN** Serializing to OpenCode
- **THEN** All fields are included if set
- **AND** Fields are omitted from Claude Code serialization
- **AND** Compatibility is rendered as YAML list
- **AND** Metadata and Hooks are rendered as YAML maps

#### Scenario: Memory fields map directly
- **GIVEN** Memory model
- **WHEN** Comparing Claude Code and OpenCode schemas
- **THEN** Paths maps to @ file references in AGENTS.md
- **AND** Content maps to project context narrative
- **AND** Both fields can be present simultaneously

#### Scenario: Tool name case conversion for OpenCode
- **GIVEN** Tool names in PascalCase (Bash, Read, Edit) or mixed case
- **WHEN** Transforming to OpenCode
- **THEN** All tool names are converted to lowercase using Sprig's `lower` function (bash, read, edit)
- **AND** OpenCode platform requires lowercase tool names
- **AND** Claude Code platform preserves original case

#### Scenario: Tool name case preservation for Claude Code
- **GIVEN** Tool names in PascalCase (Bash, Read, Edit) or mixed case
- **WHEN** Transforming to Claude Code
- **THEN** Original case is preserved (Bash, Read, Edit)
- **AND** No case conversion is applied

#### Scenario: Agent name field handling for OpenCode
- **GIVEN** Agent model with Name field
- **WHEN** Transforming to OpenCode
- **THEN** Name field is NOT rendered in frontmatter
- **AND** OpenCode uses filename as agent identifier
- **AND** Claude Code templates render name field in frontmatter

#### Scenario: Command name field handling for OpenCode
- **GIVEN** Command model with Name field
- **WHEN** Transforming to OpenCode
- **THEN** Name field is NOT rendered in frontmatter
- **AND** OpenCode uses filename as command identifier
- **AND** Claude Code templates render name field in frontmatter


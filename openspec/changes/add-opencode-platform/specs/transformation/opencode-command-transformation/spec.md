## ADDED Requirements

### Requirement: Render YAML frontmatter with command configuration
### Requirement: Include command template content body (markdown after YAML frontmatter)
### Requirement: Support $ARGUMENTS placeholder in content for argument substitution
### Requirement: Include OpenCode-specific field: subtask
### Requirement: Omit Claude Code-specific fields: allowedTools, argumentHint, context, disableModelInvocation
### Requirement: Render content after YAML frontmatter

#### Scenario: Minimal command transformation
- **GIVEN** Command with name="run-tests", description="Run all tests", content="npm test"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains YAML frontmatter with name and description
- **AND** Output contains command template in content body (markdown after YAML frontmatter)
- **AND** Content is preserved as-is

#### Scenario: Command with $ARGUMENTS placeholder
- **GIVEN** Command with content="git $ARGUMENTS"
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves "$ARGUMENTS" placeholder in content
- **AND** Template content body contains original content

#### Scenario: Command with agent field
- **GIVEN** Command with agent="code-reviewer"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "agent: code-reviewer"

#### Scenario: Command with model field
- **GIVEN** Command with model="anthropic/claude-sonnet-4-20250514"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "model: anthropic/claude-sonnet-4-20250514"

#### Scenario: Command with subtask field
- **GIVEN** Command with subtask=true
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "subtask: true"

#### Scenario: Command with subtask false
- **GIVEN** Command with subtask=false
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "subtask: false"

#### Scenario: Command omits allowedTools
- **GIVEN** Command with allowedTools=["bash", "read"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain allowedTools field
- **AND** No warning is logged (silent skip)

#### Scenario: Command omits argumentHint
- **GIVEN** Command with argumentHint="file path"
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain argumentHint field
- **AND** No warning is logged (silent skip)

#### Scenario: Command omits context field
- **GIVEN** Command with context="fork"
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain context field
- **AND** No warning is logged (silent skip)

#### Scenario: Command omits disableModelInvocation
- **GIVEN** Command with disableModelInvocation=true
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain disableModelInvocation field
- **AND** No warning is logged (silent skip)

#### Scenario: Command with full model ID
- **GIVEN** Command with model="openai/gpt-4-1"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "model: openai/gpt-4-1"
- **AND** Model ID is preserved exactly as provided

#### Scenario: Command with all optional fields
- **GIVEN** Command with name, description, content, agent, model, subtask all populated
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains all applicable fields
- **AND** YAML frontmatter is valid
- **AND** Content follows "---" separator
- **AND** Claude Code-specific fields are omitted

#### Scenario: Command preserves content indentation
- **GIVEN** Command with multi-line content:
    ```
    npm run build
    npm run test
    ```
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves original indentation
- **AND** Multi-line content is rendered correctly in content body

#### Scenario: Command preserves special characters in content
- **GIVEN** Command with content containing special characters ($, *, #, etc.)
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves all special characters
- **AND** Content is not escaped incorrectly

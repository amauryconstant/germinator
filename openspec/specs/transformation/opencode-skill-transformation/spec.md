# opencode-skill-transformation Specification

## Purpose
Transformation logic for converting Skill models from Germinator source format to OpenCode target format.

## Requirements

### Requirement: Render YAML frontmatter with skill configuration
### Requirement: Include OpenCode-specific fields: license, compatibility, metadata, hooks
### Requirement: Compatibility field rendered as YAML list
### Requirement: Metadata field rendered as YAML key-value map
### Requirement: Hooks field rendered as YAML map
### Requirement: Omit Claude Code-specific fields: allowedTools, userInvocable (plus some with different semantics)
### Requirement: Render content after YAML frontmatter

#### Scenario: Minimal skill transformation
- **GIVEN** Skill with name="git-workflow", description="Git operations", content="Provides git commands..."
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains YAML frontmatter with name and description
- **AND** No optional fields present
- **AND** Content follows frontmatter

#### Scenario: Skill with license field
- **GIVEN** Skill with license="MIT"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "license: MIT"

#### Scenario: Skill with compatibility list
- **GIVEN** Skill with compatibility=["claude-code", "opencode"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains compatibility field as YAML list:
    ```
    compatibility:
      - claude-code
      - opencode
    ```

#### Scenario: Skill with metadata map
- **GIVEN** Skill with metadata={"author": "Team", "version": "1.0"}
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains metadata field as YAML map:
    ```
    metadata:
      author: Team
      version: "1.0"
    ```

#### Scenario: Skill with hooks map
- **GIVEN** Skill with hooks={"pre": "validate", "post": "cleanup"}
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains hooks field as YAML map:
    ```
    hooks:
      pre: validate
      post: cleanup
    ```

#### Scenario: Skill with all OpenCode fields
- **GIVEN** Skill with license, compatibility, metadata, hooks all populated
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains all fields in proper order
- **AND** YAML frontmatter is valid
- **AND** Content follows "---" separator

#### Scenario: Skill omits allowedTools
- **GIVEN** Skill with allowedTools=["bash", "read"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain allowedTools field
- **AND** No warning is logged (silent skip)

#### Scenario: Skill omits userInvocable
- **GIVEN** Skill with userInvocable=true
- **WHEN** Rendered to OpenCode format
- **THEN** Output does not contain userInvocable field
- **AND** No warning is logged (silent skip)

#### Scenario: Skill with full model ID
- **GIVEN** Skill with model="anthropic/claude-sonnet-4-20250514"
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains "model: anthropic/claude-sonnet-4-20250514"
- **AND** Model ID is preserved exactly as provided

#### Scenario: Skill with multi-line content
- **GIVEN** Skill with multi-line content:
    ```
    This skill provides:
    - Feature 1
    - Feature 2
    ```
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves original formatting
- **AND** Multi-line content follows frontmatter correctly

#### Scenario: Skill preserves markdown in content
- **GIVEN** Skill content with markdown syntax (#, **, -, etc.)
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves all markdown characters
- **AND** Markdown is not escaped or corrupted

#### Scenario: Skill directory structure requirement
- **GIVEN** Skill with name="git-workflow"
- **WHEN** Writing output to OpenCode format
- **THEN** Output path is .opencode/skills/git-workflow/SKILL.md
- **AND** Directory structure is created automatically by CLI

#### Scenario: Skill with special characters in name
- **GIVEN** Skill with name="code-review-tool"
- **WHEN** Rendered to OpenCode format
- **THEN** Output path uses kebab-case name in directory
- **AND** Name is validated against regex ^[a-z0-9]+(-[a-z0-9]+)*$

#### Scenario: Skill with empty optional fields
- **GIVEN** Skill with license="", compatibility=[], metadata={}
- **WHEN** Rendered to OpenCode format
- **THEN** Empty optional fields are omitted from output
- **AND** Only non-empty fields are rendered

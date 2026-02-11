# template-rendering Specification

## Purpose

Template-based rendering of canonical documents to platform-specific output formats. Uses Go templates with Sprig functions and custom template functions to generate YAML frontmatter and markdown content for both Claude Code and OpenCode platforms.

## Requirements

### Requirement: Template Organization

The system SHALL organize templates by platform and document type in a predictable directory structure.

#### Scenario: Template directory structure

- **GIVEN** Templates are organized in the codebase
- **WHEN** Template directories are inspected
- **THEN** Templates SHALL be located at `config/templates/{platform}/{docType}.tmpl`
- **AND** Platform directories SHALL include: `claude-code`, `opencode`
- **AND** Document types SHALL include: `agent`, `command`, `memory`, `skill`

#### Scenario: Template discovery at runtime

- **GIVEN** A platform and document type are specified
- **WHEN** The system needs to render a document
- **THEN** System SHALL construct template path from platform and document type
- **AND** System SHALL search from current working directory first
- **AND** System SHALL fall back to relative path `../../config/templates/` if not found

---

### Requirement: Template Functions

The system SHALL provide custom template functions for platform-specific transformations via Go templates.

#### Scenario: permissionPolicyToPlatform function converts policies

- **GIVEN** A canonical PermissionPolicy value (restrictive, balanced, permissive, analysis, unrestricted)
- **WHEN** `{{permissionPolicyToPlatform .Policy "claude-code"}}` is called in template
- **THEN** restrictive SHALL convert to "default"
- **AND** balanced SHALL convert to "acceptEdits"
- **AND** permissive SHALL convert to "dontAsk"
- **AND** analysis SHALL convert to "plan"
- **AND** unrestricted SHALL convert to "bypassPermissions"

#### Scenario: permissionPolicyToPlatform for OpenCode returns map

- **GIVEN** A canonical PermissionPolicy value
- **WHEN** `{{permissionPolicyToPlatform .Policy "opencode"}}` is called in template
- **THEN** Output SHALL be OpenCode permission map YAML string
- **AND** Map SHALL contain all tools (edit, bash, read, grep, glob, list, webfetch, websearch)
- **AND** Each tool SHALL have "*" wildcard action
- **AND** restrictive SHALL return all tools as "ask"
- **AND** balanced SHALL return edit/read as "allow", bash as "ask"
- **AND** permissive SHALL return all tools as "allow"
- **AND** analysis SHALL return edit/bash as "deny", read as "allow"
- **AND** unrestricted SHALL return all tools as "allow"

#### Scenario: convertToolNameCase for Claude Code

- **GIVEN** A canonical tool name (lowercase, e.g., "bash", "read", "web-fetch")
- **WHEN** `{{convertToolNameCase "bash" "claude-code"}}` is called in template
- **THEN** Output SHALL be PascalCase ("Bash", "Read", "WebFetch")
- **AND** Conversion SHALL be consistent for all built-in tools
- **AND** Special characters in tool names SHALL be preserved (hyphens, numbers)
- **AND** Consecutive capitals SHALL be handled appropriately (WebSearch)

#### Scenario: convertToolNameCase for OpenCode

- **GIVEN** A canonical tool name (lowercase, e.g., "bash", "read")
- **WHEN** `{{convertToolNameCase "bash" "opencode"}}` is called in template
- **THEN** Output SHALL be lowercase ("bash", "read")
- **AND** No case conversion SHALL occur (already lowercase)
- **AND** Method SHALL be provided for consistency with Claude Code template

#### Scenario: Sprig functions available

- **GIVEN** A template is being rendered
- **WHEN** Template uses string manipulation
- **THEN** Sprig functions SHALL be available: lower, upper, trim, join, etc.
- **AND** Functions work as documented in Sprig library

---

### Requirement: Agent Template Rendering

The agent template SHALL render YAML frontmatter with agent configuration in platform-specific format.

#### Scenario: Claude Code agent renders PascalCase tools

- **GIVEN** A canonical Agent struct with tools ["bash", "read"] and permissionPolicy "balanced"
- **WHEN** Agent template is rendered for Claude Code platform
- **THEN** Output SHALL contain permissionMode: "acceptEdits"
- **AND** Tools SHALL be PascalCase: ["Bash", "Read"]
- **AND** Skills SHALL be rendered from targets.claude-code.skills array
- **AND** Field names SHALL match Claude Code structure

#### Scenario: OpenCode agent renders map format tools

- **GIVEN** A canonical Agent struct with tools ["bash", "read"] and disallowedTools ["write", "execute"]
- **WHEN** Agent template is rendered for OpenCode platform
- **THEN** Tools SHALL be converted to map: {bash: true, read: true}
- **AND** DisallowedTools SHALL be converted to map: {write: false, execute: false}
- **AND** Tool names SHALL be lowercase
- **AND** True values SHALL come from tools array
- **AND** False values SHALL come from disallowedTools array

#### Scenario: OpenCode agent omits name field

- **GIVEN** A canonical Agent with name="code-reviewer"
- **WHEN** Agent template is rendered for OpenCode platform
- **THEN** Template SHALL NOT output name field in frontmatter
- **AND** OpenCode uses filename (e.g., code-reviewer.md) as agent identifier

#### Scenario: OpenCode agent omits skills field

- **GIVEN** A canonical Agent with skills=["skill-creator", "refactoring"]
- **WHEN** Agent template is rendered for OpenCode platform
- **THEN** Output SHALL NOT contain skills field
- **AND** Skills are independent documents in OpenCode

#### Scenario: OpenCode agent renders behavior fields at top level

- **GIVEN** A canonical Agent with behavior.mode="primary", behavior.temperature=0.3, behavior.maxSteps=25
- **WHEN** Agent template is rendered for OpenCode platform
- **THEN** Output SHALL have mode: primary at top level
- **AND** Output SHALL have temperature: 0.3 at top level
- **AND** Output SHALL have maxSteps: 25 at top level
- **AND** Nested behavior object SHALL NOT be present in output

#### Scenario: OpenCode agent temperature nil vs 0.0

- **GIVEN** A canonical Agent with Temperature field as *float64 pointer
- **WHEN** Temperature is nil (not set)
- **THEN** Temperature field SHALL NOT be rendered in output
- **AND** OpenCode uses model's default temperature
- **WHEN** Temperature is 0.0 (explicitly set)
- **THEN** Output contains "temperature: 0.0"
- **AND** Template checks for nil presence, not zero value

#### Scenario: OpenCode agent omits false boolean fields

- **GIVEN** A canonical Agent with hidden=false, disabled=false
- **WHEN** Agent template is rendered for OpenCode platform
- **THEN** Output does NOT contain hidden/disable when false
- **AND** This prevents redundant output of false values

#### Scenario: Agent content follows frontmatter

- **GIVEN** A canonical Agent with content="You are a code reviewer..."
- **WHEN** Agent template is rendered
- **THEN** Content SHALL follow "---" separator after YAML frontmatter
- **AND** Content is rendered as-is (markdown formatting preserved)

---

### Requirement: Command Template Rendering

The command template SHALL render YAML frontmatter with command configuration and preserve content.

#### Scenario: OpenCode command omits name field

- **GIVEN** A canonical Command with name="run-tests"
- **WHEN** Command template is rendered for OpenCode platform
- **THEN** Output does NOT contain name field in frontmatter
- **AND** OpenCode uses filename (e.g., run-tests.md) as command identifier

#### Scenario: Command preserves $ARGUMENTS placeholder

- **GIVEN** A canonical Command with content="git $ARGUMENTS"
- **WHEN** Command template is rendered
- **THEN** Output preserves "$ARGUMENTS" placeholder in content
- **AND** Content body contains original content

#### Scenario: OpenCode command includes subtask field

- **GIVEN** A canonical Command with subtask=true
- **WHEN** Command template is rendered for OpenCode platform
- **THEN** Output contains "subtask: true"
- **AND** WHEN subtask=false, output contains "subtask: false"

#### Scenario: OpenCode command omits Claude Code fields

- **GIVEN** A canonical Command with allowedTools=["bash", "read"], argumentHint="file path", context="fork", disableModelInvocation=true
- **WHEN** Command template is rendered for OpenCode platform
- **THEN** Output does NOT contain allowedTools field
- **AND** Output does NOT contain argumentHint field
- **AND** Output does NOT contain context field
- **AND** Output does NOT contain disableModelInvocation field

#### Scenario: Command preserves multi-line content

- **GIVEN** A canonical Command with multi-line content:
    ```
    npm run build
    npm run test
    ```
- **WHEN** Command template is rendered
- **THEN** Output preserves original indentation
- **AND** Multi-line content is rendered correctly in content body

#### Scenario: Command preserves special characters

- **GIVEN** A canonical Command with content containing special characters ($, *, #, etc.)
- **WHEN** Command template is rendered
- **THEN** Output preserves all special characters
- **AND** Content is not escaped incorrectly

---

### Requirement: Memory Template Rendering

The memory template SHALL render OpenCode AGENTS.md format as plain markdown without YAML frontmatter.

#### Scenario: Memory renders as plain markdown

- **GIVEN** A canonical Memory with paths and content
- **WHEN** Memory template is rendered for OpenCode platform
- **THEN** Output is plain markdown (no YAML frontmatter)
- **AND** File format is AGENTS.md style

#### Scenario: Memory paths to @ file references

- **GIVEN** A canonical Memory with paths=["README.md", "package.json", "src/main.go"]
- **WHEN** Memory template is rendered
- **THEN** Paths rendered as @ file references:
    ```
    @README.md
    @package.json
    @src/main.go
    ```
- **AND** Each path is on its own line
- **AND** Glob patterns are preserved (e.g., src/**/*.go)

#### Scenario: Memory content as narrative

- **GIVEN** A canonical Memory with content="This project is a CLI tool for..."
- **WHEN** Memory template is rendered
- **THEN** Content is rendered as-is (markdown formatting preserved)
- **AND** Content follows file references (if both present)

#### Scenario: Memory teaching instructions

- **GIVEN** A canonical Memory with any configuration
- **WHEN** Memory template is rendered for OpenCode platform
- **THEN** Output includes explicit teaching instructions for Read tool usage
- **AND** Instructions explain how to use @ file references
- **AND** Example: "To reference files, use @ followed by the file path"

#### Scenario: Memory preserves markdown in content

- **GIVEN** A canonical Memory content with markdown syntax:
    ```
    # Project Overview

    This is a **CLI tool** that:

    - Transforms documents
    - Validates schemas
    ```
- **WHEN** Memory template is rendered
- **THEN** Output preserves all markdown characters
- **AND** Markdown is not escaped or corrupted
- **AND** Formatting is maintained

#### Scenario: Memory handles empty paths and content

- **GIVEN** A canonical Memory with paths=[] and content=""
- **WHEN** Memory template is rendered for OpenCode platform
- **THEN** Output contains minimal AGENTS.md structure
- **AND** File references section is empty or omitted
- **AND** Project context section is empty or omitted
- **AND** Teaching instructions are still included

---

### Requirement: Skill Template Rendering

The skill template SHALL render YAML frontmatter with skill configuration including OpenCode-specific fields.

#### Scenario: Minimal skill transformation

- **GIVEN** A canonical Skill with name="git-workflow", description="Git operations", content="Provides git commands..."
- **WHEN** Skill template is rendered for OpenCode platform
- **THEN** Output contains YAML frontmatter with name and description
- **AND** No optional fields present
- **AND** Content follows frontmatter

#### Scenario: Skill with OpenCode-specific fields

- **GIVEN** A canonical Skill with license="MIT", compatibility=["claude-code", "opencode"], metadata={"author": "Team", "version": "1.0"}, hooks={"pre": "validate", "post": "cleanup"}
- **WHEN** Skill template is rendered for OpenCode platform
- **THEN** Output contains license field as string
- **AND** Output contains compatibility field as YAML list
- **AND** Output contains metadata field as YAML key-value map
- **AND** Output contains hooks field as YAML map
- **AND** All fields rendered in proper order

#### Scenario: Skill compatibility list rendering

- **GIVEN** A canonical Skill with compatibility=["claude-code", "opencode"]
- **WHEN** Skill template is rendered
- **THEN** Output contains compatibility field as YAML list:
    ```
    compatibility:
      - claude-code
      - opencode
    ```

#### Scenario: Skill metadata map rendering

- **GIVEN** A canonical Skill with metadata={"author": "Team", "version": "1.0"}
- **WHEN** Skill template is rendered
- **THEN** Output contains metadata field as YAML map:
    ```
    metadata:
      author: Team
      version: "1.0"
    ```

#### Scenario: Skill hooks map rendering

- **GIVEN** A canonical Skill with hooks={"pre": "validate", "post": "cleanup"}
- **WHEN** Skill template is rendered
- **THEN** Output contains hooks field as YAML map:
    ```
    hooks:
      pre: validate
      post: cleanup
    ```

#### Scenario: OpenCode skill omits Claude Code fields

- **GIVEN** A canonical Skill with allowedTools=["bash", "read"], userInvocable=true
- **WHEN** Skill template is rendered for OpenCode platform
- **THEN** Output does NOT contain allowedTools field
- **AND** Output does NOT contain userInvocable field
- **AND** No warning is logged (silent skip)

#### Scenario: Skill directory structure requirement

- **GIVEN** A canonical Skill with name="git-workflow"
- **WHEN** Writing output to OpenCode format
- **THEN** Output path is .opencode/skills/git-workflow/SKILL.md
- **AND** Directory structure is created automatically by CLI

#### Scenario: Skill with special characters in name

- **GIVEN** A canonical Skill with name="code-review-tool"
- **WHEN** Skill template is rendered for OpenCode platform
- **THEN** Output path uses kebab-case name in directory
- **AND** Name is validated against regex ^[a-z0-9]+(-[a-z0-9]+)*$

#### Scenario: Skill preserves markdown in content

- **GIVEN** A canonical Skill content with markdown syntax (#, **, -, etc.)
- **WHEN** Skill template is rendered
- **THEN** Output preserves all markdown characters
- **AND** Markdown is not escaped or corrupted

---

### Requirement: Template Execution

Templates SHALL be executed with proper context and error handling.

#### Scenario: Template receives Doc and Adapter context

- **GIVEN** A canonical document and platform adapter are available
- **WHEN** Template is executed
- **THEN** Template SHALL receive context with Doc field containing canonical struct
- **AND** Template SHALL receive context with Adapter field providing platform methods
- **AND** Context structure enables template function calls

#### Scenario: Template error messages

- **GIVEN** A template has a syntax error or missing function
- **WHEN** Template execution is attempted
- **THEN** Error SHALL be returned with template filename
- **AND** Error SHALL include line number where error occurred
- **AND** Error message SHALL be clear enough for debugging

#### Scenario: Empty optional fields are omitted

- **GIVEN** A canonical document with empty optional fields (nil pointers, empty strings, empty maps/lists)
- **WHEN** Template is rendered
- **THEN** Empty optional fields SHALL be omitted from output
- **AND** Only non-empty fields are rendered
- **AND** This prevents cluttering output with null/empty values

#### Scenario: Template function errors

- **GIVEN** A template function receives invalid input (e.g., unknown permission policy)
- **WHEN** Template function is called
- **THEN** Function SHALL return error or empty string
- **AND** Template execution SHOULD NOT crash
- **AND** Output MAY be incomplete but error is surfaced

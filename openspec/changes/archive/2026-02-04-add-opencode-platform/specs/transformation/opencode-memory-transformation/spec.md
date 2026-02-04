## ADDED Requirements

### Requirement: Render AGENTS.md format with project context and file references
### Requirement: Transform paths array to @ file references (e.g., @README.md)
### Requirement: Render content as project context narrative
### Requirement: Support both paths-only, content-only, and combined modes
### Requirement: Include explicit teaching instructions for Read tool usage
### Requirement: Format as markdown file (no YAML frontmatter)

#### Scenario: Memory with paths only
- **GIVEN** Memory with paths=["README.md", "package.json", "src/main.go"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains AGENTS.md format
- **AND** Paths rendered as @ file references:
    ```
    @README.md
    @package.json
    @src/main.go
    ```
- **AND** No content section (content is empty)

#### Scenario: Memory with content only
- **GIVEN** Memory with content="This project is a CLI tool for..."
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains project context narrative
- **AND** Content is rendered as-is (markdown formatting preserved)
- **AND** No @ file references (paths is empty)

#### Scenario: Memory with both paths and content
- **GIVEN** Memory with paths=["README.md"] and content="This project..."
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains both @ file references and project context
- **AND** @README.md appears before content section
- **AND** Content follows file references
- **AND** Clear separation between file references and context

#### Scenario: Memory includes teaching instructions
- **GIVEN** Memory with any configuration
- **WHEN** Rendered to OpenCode format
- **THEN** Output includes explicit teaching instructions for Read tool usage
- **AND** Instructions explain how to use @ file references
- **AND** Example: "To reference files, use @ followed by the file path"

#### Scenario: Memory with multiple paths
- **GIVEN** Memory with paths=["README.md", "LICENSE", "docs/usage.md", "src/**/*.go"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output includes all paths as @ references
- **AND** Each path is on its own line
- **AND** Glob patterns are preserved (e.g., src/**/*.go)

#### Scenario: Memory with nested directory paths
- **GIVEN** Memory with paths=["src/models/agent.go", "test/integration_test.go"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves directory structure in @ references
- **AND** @src/models/agent.go appears correctly
- **AND** @test/integration_test.go appears correctly

#### Scenario: Memory preserves markdown in content
- **GIVEN** Memory content with markdown syntax:
    ```
    # Project Overview

    This is a **CLI tool** that:

    - Transforms documents
    - Validates schemas
    ```
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves all markdown characters
- **AND** Markdown is not escaped or corrupted
- **AND** Formatting is maintained

#### Scenario: Memory with empty paths and content
- **GIVEN** Memory with paths=[] and content=""
- **WHEN** Rendered to OpenCode format
- **THEN** Output contains minimal AGENTS.md structure
- **AND** File references section is empty or omitted
- **AND** Project context section is empty or omitted
- **AND** Teaching instructions are still included

#### Scenario: Memory with relative paths
- **GIVEN** Memory with paths=["./README.md", "./config.yaml"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves relative path syntax in @ references
- **AND** @./README.md appears correctly

#### Scenario: Memory with absolute paths
- **GIVEN** Memory with paths=["/home/user/project/README.md"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves absolute path in @ references
- **AND** @/home/user/project/README.md appears correctly

#### Scenario: Memory with long content
- **GIVEN** Memory with content exceeding 1024 characters
- **WHEN** Rendered to OpenCode format
- **THEN** Output renders entire content
- **AND** No truncation occurs
- **AND** Long content is preserved as-is

#### Scenario: Memory teaching instructions format
- **GIVEN** Memory transformation
- **WHEN** Rendered to OpenCode format
- **THEN** Teaching instructions are clear and concise
- **AND** Instructions are formatted as markdown
- **AND** Instructions appear at top of AGENTS.md or in dedicated section

#### Scenario: Memory AGENTS.md header
- **GIVEN** Memory transformation
- **WHEN** Rendered to OpenCode format
- **THEN** Output may include "# AGENTS.md" header
- **AND** File format is clearly identified

#### Scenario: Memory preserves special characters in paths
- **GIVEN** Memory with paths=["file with spaces.md", "config-file.yaml"]
- **WHEN** Rendered to OpenCode format
- **THEN** Output preserves spaces and special characters in @ references
- **AND** @file with spaces.md appears correctly

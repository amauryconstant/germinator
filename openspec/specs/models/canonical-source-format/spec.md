# canonical-source-format Specification

## Purpose

Domain-driven canonical YAML format expressing AI coding assistant configuration intent independent of platform specifics, using permission policies, behavior objects, and target-specific overrides.

## Requirements

### Requirement: Canonical Format Structure

The canonical format SHALL use domain-driven field names expressing intent, not platform-specific terminology or mechanics.

#### Scenario: Agent with canonical structure

- **GIVEN** A canonical agent YAML with name, description, tools, disallowedTools, permissionPolicy, behavior, model, and targets sections
- **WHEN** The file is parsed
- **THEN** Identity fields SHALL be present (name, description)
- **AND** Tools fields SHALL be present (tools, disallowedTools)
- **AND** Policy field SHALL be permissionPolicy enum value
- **AND** Behavior object SHALL group settings (mode, temperature, maxSteps, prompt, hidden, disabled)
- **AND** Model field SHALL be a string (user-provided full ID)
- **AND** Targets section SHALL contain platform-specific overrides

#### Scenario: Command with canonical structure

- **GIVEN** A canonical command YAML with name, description, tools, execution, arguments, model, and targets sections
- **WHEN** The file is parsed
- **THEN** Identity fields SHALL be present (name, description)
- **AND** Tools field SHALL be present (tools array)
- **AND** Execution object SHALL group context, subtask, agent
- **AND** Arguments object SHALL contain hint
- **AND** Model field SHALL be a string
- **AND** Targets section SHALL contain platform-specific overrides

#### Scenario: Skill with canonical structure

- **GIVEN** A canonical skill YAML with name, description, tools, extensions, execution, model, and targets sections
- **WHEN** The file is parsed
- **THEN** Identity fields SHALL be present (name, description)
- **AND** Tools field SHALL be present (tools array)
- **AND** Extensions object SHALL group license, compatibility, metadata, hooks
- **AND** Execution object SHALL group context, agent, userInvocable
- **AND** Model field SHALL be a string
- **AND** Targets section SHALL contain platform-specific overrides

#### Scenario: Memory with canonical structure

- **GIVEN** A canonical memory YAML with paths and content sections
- **WHEN** The file is parsed
- **THEN** Paths field SHALL be an array of file paths
- **OR** Content field SHALL contain narrative markdown text
- **OR** Both fields SHALL be present

### Requirement: Permission Policy Enum

The canonical format SHALL use a permissionPolicy enum expressing security posture independent of platform.

#### Scenario: Valid permission policies

- **GIVEN** A canonical YAML with permissionPolicy field
- **WHEN** The field is parsed
- **THEN** PermissionPolicy SHALL be one of: restrictive, balanced, permissive, analysis, unrestricted
- **AND** Field values SHALL use exact enum names (case-sensitive)

#### Scenario: Omitted permissionPolicy uses platform default

- **GIVEN** A canonical agent YAML without permissionPolicy field
- **WHEN** The file is transformed to a platform
- **THEN** Target platform's default permission behavior SHALL be used
- **AND** No permissionPolicy field SHALL be present in output

### Requirement: Targets Section for Platform Overrides

The canonical format SHALL use a `targets` section containing platform-specific configuration overrides.

#### Scenario: Claude Code specific overrides

- **GIVEN** A canonical agent YAML with targets.claude-code section
- **WHEN** The file is transformed to Claude Code platform
- **THEN** targets.claude-code.skills SHALL be rendered as skills array
- **AND** Other targets sections SHALL be omitted
- **AND** Common fields (permissionPolicy) SHALL be converted to Claude Code permissionMode

#### Scenario: OpenCode specific overrides

- **GIVEN** A canonical agent YAML with targets.opencode section
- **WHEN** The file is transformed to OpenCode platform
- **THEN** targets.opencode section SHALL be rendered if non-empty
- **AND** Other targets sections SHALL be omitted
- **AND** OpenCode-specific behavior fields in targets section SHALL override behavior object

#### Scenario: No targets section (pure canonical)

- **GIVEN** A canonical YAML without targets section
- **WHEN** The file is transformed to any platform
- **THEN** Only common fields SHALL be used
- **AND** Platform-specific overrides SHALL NOT be added to output

### Requirement: Behavior Object for Agent Settings

The canonical format SHALL group agent execution settings into a `behavior` object.

#### Scenario: Behavior object with mode field

- **GIVEN** A canonical agent YAML with behavior.mode field
- **WHEN** The file is parsed
- **THEN** Mode SHALL be one of: primary, subagent, all
- **AND** Field SHALL be nested under behavior key

#### Scenario: Behavior object with temperature field

- **GIVEN** A canonical agent YAML with behavior.temperature field
- **WHEN** The file is parsed
- **THEN** Temperature SHALL be a *float64 pointer (not float64 value)
- **AND** Value (when not nil) SHALL be between 0.0 and 1.0
- **AND** Field SHALL be nested under behavior key

#### Scenario: Behavior temperature nil vs 0.0 distinction

- **GIVEN** Agent model with behavior.temperature as *float64 pointer
- **WHEN** behavior.temperature is nil
- **THEN** behavior.temperature field SHALL NOT be rendered in OpenCode output (user didn't set it)
- **AND** OpenCode will use model's default temperature
- **WHEN** behavior.temperature is 0.0
- **THEN** behavior.temperature field SHALL be rendered as `temperature: 0.0` in OpenCode output
- **AND** User explicitly requested deterministic low-temperature behavior
- **AND** Template rendering checks nil presence (not zero value)
- **AND** Distinction between unset and explicitly set to 0.0 is preserved

#### Scenario: Behavior object with maxSteps field

- **GIVEN** A canonical agent YAML with behavior.maxSteps field
- **WHEN** The file is parsed
- **THEN** MaxSteps SHALL be an integer greater than 0
- **AND** Field SHALL be nested under behavior key

#### Scenario: Behavior object with prompt field

- **GIVEN** A canonical agent YAML with behavior.prompt field
- **WHEN** The file is parsed
- **THEN** Prompt SHALL be a string containing system prompt override
- **AND** Field SHALL be nested under behavior key

#### Scenario: Behavior object with boolean flags

- **GIVEN** A canonical agent YAML with behavior.hidden and behavior.disabled fields
- **WHEN** The file is parsed
- **THEN** Hidden SHALL be a boolean value
- **AND** Disabled SHALL be a boolean value
- **AND** Fields SHALL be nested under behavior key

### Requirement: Split Tool Lists

The canonical format SHALL use separate `tools` and `disallowedTools` arrays for tool access control.

#### Scenario: Tools array for allowed tools

- **GIVEN** A canonical YAML with tools field
- **WHEN** The field is parsed
- **THEN** Tools SHALL be an array of strings
- **AND** Each string SHALL be a tool name (lowercase)

#### Scenario: DisallowedTools array for denied tools

- **GIVEN** A canonical YAML with disallowedTools field
- **WHEN** The field is parsed
- **THEN** DisallowedTools SHALL be an array of strings
- **AND** Each string SHALL be a tool name (lowercase)

#### Scenario: Both tools and disallowedTools present

- **GIVEN** A canonical YAML with both tools and disallowedTools fields
- **WHEN** The file is parsed
- **THEN** Both arrays SHALL be parsed successfully
- **AND** Tools appearing in both arrays SHALL NOT cause validation errors (adapter resolves precedence)

### Requirement: Simple Model String

The canonical format SHALL use a simple `model` string field for user-provided full model IDs.

#### Scenario: Model with provider/id format

- **GIVEN** A canonical YAML with model="anthropic/claude-sonnet-4-20250514"
- **WHEN** The field is parsed
- **THEN** Model SHALL be parsed as a string
- **AND** No alias resolution or normalization SHALL occur
- **AND** Model ID SHALL be passed to output as-is (after case conversion if needed)

#### Scenario: Model with custom provider

- **GIVEN** A canonical YAML with model="openai/gpt-4-1"
- **WHEN** The field is parsed
- **THEN** Model SHALL be parsed as a string
- **AND** Adapter SHALL pass to platform-specific output without transformation
- **AND** Platform SHALL handle provider-specific ID format

#### Scenario: Omitted model uses platform default

- **GIVEN** A canonical YAML without model field
- **WHEN** The file is transformed to a platform
- **THEN** Target platform's default model SHALL be used
- **AND** No model field SHALL be present in output

### Requirement: Extensions Object for Skill Metadata

The canonical format SHALL group skill metadata into an `extensions` object.

#### Scenario: Extensions object with license field

- **GIVEN** A canonical skill YAML with extensions.license field
- **WHEN** The field is parsed
- **THEN** License SHALL be a string (SPDX identifier or custom)
- **AND** Field SHALL be nested under extensions key

#### Scenario: Extensions object with compatibility list

- **GIVEN** A canonical skill YAML with extensions.compatibility field
- **WHEN** The field is parsed
- **THEN** Compatibility SHALL be an array of strings
- **AND** Each string SHALL be a platform name (claude-code, opencode, or future platforms)
- **AND** Field SHALL be nested under extensions key

#### Scenario: Extensions object with metadata map

- **GIVEN** A canonical skill YAML with extensions.metadata field
- **WHEN** The field is parsed
- **THEN** Metadata SHALL be a map of string to string
- **AND** Keys SHALL be metadata property names
- **AND** Values SHALL be string values
- **AND** Field SHALL be nested under extensions key

#### Scenario: Extensions object with hooks map

- **GIVEN** A canonical skill YAML with extensions.hooks field
- **WHEN** The field is parsed
- **THEN** Hooks SHALL be a map of string to string
- **AND** Keys SHALL be hook names (e.g., pre-commit, post-review)
- **AND** Values SHALL be hook commands or references
- **AND** Field SHALL be nested under extensions key

### Requirement: Execution Object for Command/Skill Settings

The canonical format SHALL group execution-related settings into an `execution` object.

#### Scenario: Execution object with context field

- **GIVEN** A canonical command or skill YAML with execution.context field
- **WHEN** The field is parsed
- **THEN** Context SHALL be "fork" or omitted
- **AND** Field SHALL be nested under execution key

#### Scenario: Execution object with subtask field

- **GIVEN** A canonical command YAML with execution.subtask field
- **WHEN** The field is parsed
- **THEN** Subtask SHALL be a boolean value
- **AND** Field SHALL be nested under execution key

#### Scenario: Execution object with agent field

- **GIVEN** A canonical command or skill YAML with execution.agent field
- **WHEN** The field is parsed
- **THEN** Agent SHALL be a string (agent name reference)
- **AND** Field SHALL be nested under execution key

#### Scenario: Execution object with userInvocable field

- **GIVEN** A canonical skill YAML with execution.userInvocable field
- **WHEN** The field is parsed
- **THEN** UserInvocable SHALL be a boolean value
- **AND** Field SHALL be nested under execution key

### Requirement: Arguments Object for Command Hints

The canonical format SHALL use an `arguments` object containing command interface hints.

#### Scenario: Arguments object with hint field

- **GIVEN** A canonical command YAML with arguments.hint field
- **WHEN** The field is parsed
- **THEN** Hint SHALL be a string describing expected arguments
- **AND** Field SHALL be nested under arguments key

### Requirement: Canonical Format is Platform-Agnostic

The canonical format SHALL NOT contain platform-specific terminology, enums, or field names.

#### Scenario: No platform-specific enums

- **GIVEN** A canonical YAML file
- **WHEN** Inspected for field values
- **THEN** PermissionMode (Claude Code) SHALL NOT be present
- **AND** Mode (OpenCode) SHALL NOT be at top level (only in behavior.mode)
- **AND** Only permissionPolicy enum values SHALL be used (restrictive, balanced, permissive, analysis, unrestricted)

#### Scenario: No platform-specific field names at top level

- **GIVEN** A canonical YAML file
- **WHEN** Inspected for top-level field names
- **THEN** skills (Claude Code) SHALL NOT be at top level
- **AND** argument-hint (Claude Code) SHALL be at arguments.hint
- **AND** user-invocable (Claude Code) SHALL be at execution.userInvocable
- **AND** disable-model-invocation (Claude Code) SHALL be at targets.claude-code.disableModelInvocation

### Requirement: Content Section for Markdown Body

All canonical formats SHALL support a markdown content body following YAML frontmatter.

#### Scenario: Content body parsing

- **GIVEN** A canonical YAML file with markdown content after frontmatter
- **WHEN** The file is parsed
- **THEN** Content SHALL be extracted as raw markdown string
- **AND** Content SHALL NOT be parsed as YAML
- **AND** Content SHALL be preserved exactly as written (indentation, line endings)

#### Scenario: Memory content without frontmatter

- **GIVEN** A canonical memory file without YAML frontmatter (pure markdown)
- **WHEN** The file is parsed
- **THEN** Entire content SHALL be treated as Content field
- **AND** Paths field SHALL be empty or omitted
- **AND** No parsing error SHALL occur

# Germinator

A configuration adapter that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. It uses a canonical Germinator YAML format as the source and adapts it to target platforms (Claude Code, OpenCode).

## Overview

Germinator enables users who test different AI coding assistants regularly to:
- Maintain **one source of truth** for their coding assistant setup
- Quickly **switch platforms** without rewriting their configuration
- **Adapt** their setup to new projects easily

## Installation

### Quick Install

The easiest way to install germinator:

```bash
curl -sSL https://gitlab.com/amoconst/germinator/-/raw/main/install.sh | bash
```

### From Source

```bash
go build -o germinator ./cmd
```

### Manual Installation

See [INSTALL.md](INSTALL.md) for detailed installation instructions for Linux, macOS, and Windows, including checksum verification and GPG signature verification.

### Prerequisites

- Go 1.25.5 or later

## Usage

```bash
./germinator --help
```

### Commands

Germinator provides two main commands:

- **validate** - Validate a Germinator source document
- **adapt** - Transform a Germinator source document to a target platform

**Important**: The `--platform` flag is required for both commands. Specify either `claude-code` or `opencode`.

### Examples

```bash
# Validate a Germinator source document
./germinator validate path/to/document.yaml --platform claude-code

# Adapt an agent for Claude Code
./germinator adapt agent.yaml .claude/agents/my-agent.yaml --platform claude-code

# Adapt a skill for OpenCode
./germinator adapt skill.yaml .opencode/skills/my-skill/SKILL.md --platform opencode

# Adapt memory to AGENTS.md
./germinator adapt memory.yaml AGENTS.md --platform opencode
```

## Supported Platforms

Germinator supports transformation to the following platforms:

### Claude Code
- **Agents**: `.claude/agents/<name>.yaml`
- **Commands**: `.claude/commands/<name>.md`
- **Skills**: `.claude/skills/<name>/SKILL.md`
- **Memory**: `.claude/memory/<name>.md`

### OpenCode
- **Agents**: `.opencode/agents/<name>.yaml`
- **Commands**: `.opencode/commands/<name>.md`
- **Skills**: `.opencode/skills/<name>/SKILL.md`
- **Memory**: `AGENTS.md` (memory documents are merged into project-level instructions)

## Document Types

Germinator supports four types of AI coding assistant documents:

| Type | Description |
|------|-------------|
| **Agents** | Specialized AI agents with capabilities, tools, and selection criteria |
| **Commands** | CLI commands with tool references, templates, and execution rules |
| **Memory** | Project context and guidelines for AI assistants |
| **Skills** | Specialized skills and techniques with metadata and hooks |

## Germinator Source Format

The Germinator YAML format is the canonical source containing ALL fields for ALL platforms. Source files include both Claude Code and OpenCode specific fields, enabling unidirectional transformation to either platform.

### Example Agent Source

```yaml
name: my-agent
description: A specialized agent for code review
model: anthropic/claude-sonnet-4-20250514
permissionMode: acceptEdits  # Claude Code field
mode: strict                  # OpenCode field
temperature: 0.7
tools:
  - bash
  - edit
  - read
prompt: You are a code review agent.
```

### Example Skill Source

```yaml
name: git-workflow
description: Git workflow management
context: fork
allowed-tools:
  - bash
compatibility:
  - claude-code
  - opencode
```

## Field Mappings

### Agent

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| name | ✓ | ⚠ omitted (uses filename as identifier) |
| description | ✓ | ✓ |
| model | ✓ | ✓ (full provider-prefixed ID) |
| tools | ✓ | ✓ (converted to lowercase) |
| disallowedTools | ✓ | ✓ (converted to lowercase, set false) |
| permissionMode | ✓ | → Permission object (nested with ask/allow/deny) |
| skills | ✓ | ⚠ (skipped - not supported) |
| mode | - | ✓ (primary/subagent/all, defaults to all) |
| temperature | - | ✓ (*float64 pointer, omits when nil) |
| maxSteps | - | ✓ |
| hidden | - | ✓ (omits when false) |
| prompt | - | ✓ |
| disable | - | ✓ (omits when false) |

### Command

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| name | ✓ | ✓ |
| description | ✓ | ✓ |
| allowed-tools | ✓ | ⚠ (skipped - not supported) |
| disallowed-tools | ✓ | ⚠ (skipped - not supported) |
| subtask | ✓ | ✓ |
| argument-hint | ✓ | ⚠ (skipped - not supported) |
| context | ✓ (fork) | ✓ (fork) |
| agent | ✓ | ✓ |
| model | ✓ | ✓ (full provider-prefixed ID) |
| disable-model-invocation | ✓ | ⚠ (skipped - not supported) |

### Skill

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| name | ✓ | ✓ |
| description | ✓ | ✓ |
| allowed-tools | ✓ | ⚠ (skipped - not supported) |
| disallowed-tools | ✓ | ⚠ (skipped - not supported) |
| license | ✓ | ✓ |
| compatibility | ✓ | ✓ |
| metadata | ✓ | ✓ |
| hooks | ✓ | ✓ |
| model | ✓ | ✓ (full provider-prefixed ID) |
| context | ✓ (fork) | ✓ (fork) |
| agent | ✓ | ✓ |
| user-invocable | ✓ | ⚠ (skipped - not supported) |

### Memory

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| paths | ✓ | → @ file references (one per line) |
| content | ✓ | → Narrative context (rendered as-is) |

## Known Limitations

### Permission Mode Transformation
The transformation from Claude Code's `permissionMode` enum to OpenCode's permission object is a basic approximation:
- `default` → `{edit: {"*": "ask"}, bash: {"*": "ask"}}`
- `acceptEdits` → `{edit: {"*": "allow"}, bash: {"*": "ask"}}`
- `dontAsk` → `{edit: {"*": "allow"}, bash: {"*": "allow"}}`
- `bypassPermissions` → `{edit: {"*": "allow"}, bash: {"*": "allow"}}`
- `plan` → `{edit: {"*": "deny"}, bash: {"*": "deny"}}`
- Only `edit` and `bash` tools are mapped (7+ other OpenCode permissionable tools remain undefined)
- Command-level permission rules are not supported

### Skipped Fields
The following fields are not supported in OpenCode and are silently skipped:
- **Agent**: `skills`
- **Command**: `disableModelInvocation`, `argumentHint`, `allowedTools`, `disallowedTools`
- **Skill**: `userInvocable`, `allowedTools`, `disallowedTools`

### DisallowedTools Forward Compatibility
OpenCode does not support `disallowedTools` in agents. Fields are included for forward compatibility but not used in current transformations.

### Unidirectional Transformation
Transformation is one-way: Germinator format → target platform only. There is no support for bidirectional conversion.

## Breaking Changes and Migration

### Germinator Source Format (v0.4.0+)
**Breaking Change**: Germinator now uses its own canonical YAML format as the source. Existing Claude Code YAML files are incompatible.

**Migration Guide**:
1. Create new Germinator source files with platform-agnostic field names
2. Include both Claude Code and OpenCode specific fields in the source
3. Use `--platform` flag to specify target platform

**Example Migration**:
```yaml
# Old Claude Code format (incompatible)
name: my-agent
description: My agent
permissionMode: default

# New Germinator format
name: my-agent
description: My agent
permissionMode: default  # Claude Code field
mode: strict             # OpenCode field
model: anthropic/claude-sonnet-4-20250514
```

### --platform Flag Requirement (v0.4.0+)
**Breaking Change**: The `--platform` flag is now required for all CLI operations. No default to `claude-code`.

**Migration**: Update all scripts and commands to include `--platform claude-code` or `--platform opencode`.

### Validate() Signature Change (v0.4.0+)
**Breaking Change**: `Validate()` methods now require a `platform string` parameter.

**Migration**: Update custom parsers or validation code to pass the platform parameter:
```go
// Before
errs := agent.Validate()

// After
errs := agent.Validate("claude-code")
```

### Field Name Changes
- No field name changes in this release
- Field casing uses Go conventions (PascalCase) for compatibility

## License

[Add your license here]

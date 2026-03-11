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

Germinator provides the following commands:

- **validate** - Validate a Germinator source document
- **adapt** - Transform a Germinator source document to a target platform
- **canonicalize** - Convert a platform-specific document to canonical Germinator format
- **library** - Manage library resources (list, show)
- **init** - Initialize library resources in a project

**Important**: The `--platform` flag is required for validate, adapt, and canonicalize. Specify either `claude-code` or `opencode`.

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

# Canonicalize a Claude Code agent to Germinator format
./germinator canonicalize .claude/agents/my-agent.yaml agent.yaml --platform claude-code

# List available library resources
./germinator library list

# Initialize library resources to a project
./germinator init --platform opencode --output . --ref agent-base
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

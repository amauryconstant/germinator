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

## Detailed Reference

For complete field mappings and known limitations, see [ARCHITECTURE.md](ARCHITECTURE.md).

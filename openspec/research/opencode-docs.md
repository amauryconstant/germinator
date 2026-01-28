# OpenCode Documentation

Platform documentation for AI coding assistant configuration

## Official Documentation Sources

- [Agents](https://opencode.ai/docs/agents) - Agent configuration
- [Skills](https://opencode.ai/docs/skills) - Agent skills format
- [Permissions](https://opencode.ai/docs/permissions) - Permission system
- [Commands](https://opencode.ai/docs/commands) - Custom commands
- [Tools](https://opencode.ai/docs/tools) - Tool configuration
- [Config](https://opencode.ai/docs/config) - Global configuration
- [Models](https://opencode.ai/docs/models) - Model configuration
- [Rules](https://opencode.ai/docs/rules) - Project rules (AGENTS.md/CLAUDE.md)

---

## Document Types

### Agents

**Files:** `opencode.json` or `.opencode/agents/*.md`

**Frontmatter Fields (Markdown):**

| Field       | Type    | Required | Description                                                     |
| ----------- | ------- | -------- | --------------------------------------------------------------- |
| description | string  | Yes      | Brief description of what agent does                            |
| mode        | string  | No       | `primary`, `subagent`, or `all` (default: `all`)                |
| model       | string  | No       | Model: `provider/model-id` format                               |
| prompt      | string  | No       | System prompt file path or content                              |
| tools       | object  | No       | Tools available to agent (lowercase tool names, boolean values) |
| permissions | object  | No       | Permission overrides                                            |
| temperature | number  | No       | Temperature (0.0-1.0, default: varies)                          |
| maxSteps    | number  | No       | Maximum agentic iterations (must be > 0)                        |
| disable     | boolean | No       | Disable agent                                                   |
| hidden      | boolean | No       | Hide from `@` autocomplete (subagents only)                     |

**Body (Markdown):** Markdown content with system prompt instructions

**JSON Structure:**

```json
{
  "agent": {
    "agent-name": {
      "description": "string",
      "mode": "primary|subagent|all",
      "model": "provider/model-id",
      "prompt": "file or string",
      "tools": {
        "tool-name": true/false
      },
      "permissions": {
        "tool": "allow|ask|deny"
      },
      "temperature": 0.5,
      "maxSteps": 10
    }
  }
}
```

---

### Skills

**Files:** `.opencode/skills/<name>/SKILL.md` or `~/.claude/skills/`

**Frontmatter Fields:**

| Field         | Type   | Required | Description                       |
| ------------- | ------ | -------- | --------------------------------- |
| name          | string | Yes      | Skill identifier                  |
| description   | string | Yes      | What skill does                   |
| license       | string | No       | License identifier                |
| compatibility | string | No       | Platform compatibility            |
| metadata      | object | No       | String-to-string map for metadata |

**Naming Constraints:**

- 1-64 characters
- Lowercase alphanumeric with single hyphen separators
- Regex: `^[a-z0-9]+(-[a-z0-9]+)*$`
- Must match directory name

**Body:** Markdown content with instructions

---

### Commands

**Files:** `opencode.json` or `.opencode/commands/*.md`

**Frontmatter Fields (Markdown):**

| Field       | Type    | Required | Description                |
| ----------- | ------- | -------- | -------------------------- |
| description | string  | No       | Description shown in TUI   |
| agent       | string  | No       | Which agent should execute |
| subtask     | boolean | No       | Force subagent invocation  |
| model       | string  | No       | Override model             |

**Body (Markdown):** Markdown content with instructions

**String Substitutions:**

- `$ARGUMENTS` - All arguments
- `$1`, `$2`, `$3` - Positional arguments
- `` `!command` `` - Shell output injection
- `@filename` - File content inclusion

---

### Permissions

**Configuration:** Configured in JSON

### Permission Actions

- `"allow"` - Run without approval
- `"ask"` - Prompt for approval
- `"deny"` - Block the action

### Available Permissions

| Permission            | Description                                            |
| --------------------- | ------------------------------------------------------ |
| read                  | Reading files (matches file path)                      |
| edit                  | All file modifications (edit, write, patch, multiedit) |
| glob                  | File globbing (matches glob pattern)                   |
| grep                  | Content search (matches regex pattern)                 |
| list                  | Listing directories (matches directory path)           |
| bash                  | Running shell commands (matches parsed commands)       |
| task                  | Launching subagents (matches subagent type)            |
| skill                 | Loading skills (matches skill name)                    |
| lsp                   | Running LSP queries                                    |
| todoread, todowrite   | Todo list operations                                   |
| webfetch              | Fetching URLs (matches URL)                            |
| websearch, codesearch | Web/code search (matches query)                        |
| external_directory    | Paths outside working directory                        |
| doom_loop             | Repeated tool calls                                    |

### Defaults

- Most: `"allow"`
- `doom_loop`, `external_directory`: `"ask"`
- `read`: `.env` files denied by default

### Structure

```json
{
  "permission": {
    "bash": {
      "*": "ask",
      "git *": "allow",
      "grep *": "allow"
    },
    "edit": {
      "*.env": "deny",
      "*.mdx": "allow"
    }
  }
}
```

### Pattern Matching

- Simple wildcard: `*`, `?`
- Last matching rule wins
- Can use home directory expansion: `~/projects/*`

### Agent-Specific Permissions

```json
{
  "agent": {
    "build": {
      "permission": {
        "edit": "ask",
        "bash": {
          "git push": "deny"
        }
      }
    }
  }
}
```

---

### Tools

**Configuration:** Built-in tools configured via permissions

### Built-in Tools

| Tool                | Description            |
| ------------------- | ---------------------- |
| bash                | Execute shell commands |
| edit                | Modify files           |
| write               | Create/overwrite files |
| read                | Read files             |
| grep                | Search content         |
| glob                | Find files             |
| list                | List directories       |
| lsp (experimental)  | LSP queries            |
| patch               | Apply patches          |
| skill               | Load skills            |
| todowrite, todoread | Todo lists             |
| webfetch            | Fetch web content      |
| question            | Ask questions          |

**Configuration:**

- Via `permission` object in config
- Via `tools` object in agent
- Via wildcards for MCP servers: `mymcp_*`

**Note:** Tool names are **lowercase** in OpenCode

---

### Rules

**Files:** `AGENTS.md` or `CLAUDE.md` (Claude Code compatible)

**Format:** Claude Code compatible - no frontmatter, pure markdown

---

### Config

**File:** Global configuration in JSON

**Structure:** Contains agent, permission, tool, and model configurations

---

### Models

**Format:** `provider/model-id`

**Examples:**

- `anthropic/claude-sonnet-4-5`
- `opencode/gpt-5.1-codex`
- `lmstudio/google/gemma-3n-e4b`

**Variants:**

- Built-in variants exist for popular providers
- Custom variants can be defined in config

---

## Permission System

### Permission Object Structure

OpenCode uses a **structured permission object** with tool-specific configurations:

```json
{
  "permission": {
    "bash": {
      "*": "ask",
      "git *": "allow",
      "git push *": "deny"
    }
  }
}
```

**Key Points:**

- `bash` is an **object with command keys**, not a simple string value
- Supports pattern matching with wildcards (`*`, `?`)
- Last matching rule wins
- Can have tool-specific overrides
- No wildcard support for generic permissions (cannot use `"*": "allow"` globally)

---

## Agent Modes

### Mode Values

- **primary** - Main agent you interact with directly (cycle with Tab)
- **subagent** - Specialized assistant invoked by primary or via `@` mention
- **all** - Can function as either (default if not specified)

### Built-in Agents

- **Build** (mode: primary) - Default agent with all tools
- **Plan** (mode: primary) - Restricted agent for analysis
- **General** (mode: subagent) - General-purpose agent
- **Explore** (mode: subagent) - Fast, read-only agent

---

## Validation Constraints

### Skill Names

- 1-64 characters
- Lowercase alphanumeric with single hyphen separators
- Regex: `^[a-z0-9]+(-[a-z0-9]+)*$`
- Must match directory name

### Descriptions

- 1-1024 characters
- Should be specific for agent selection

### Temperature

- Range: 0.0 to 1.0
- Typical ranges:
  - 0.0-0.2: Focused/deterministic
  - 0.3-0.5: Balanced
  - 0.6-1.0: Creative/variable

### MaxSteps

- Must be > 0
- No upper limit specified

### Permissions

- Pattern matching with wildcards
- Last matching rule wins
- Home directory expansion supported

---

## YAML/JSON Examples

### Agent Example (JSON)

```json
{
  "$schema": "https://opencode.ai/config.json",
  "agent": {
    "build": {
      "mode": "primary",
      "model": "anthropic/claude-sonnet-4-20250514",
      "prompt": "{file:./prompts/build.txt}",
      "tools": {
        "write": true,
        "edit": true,
        "bash": true
      }
    },
    "plan": {
      "mode": "primary",
      "model": "anthropic/claude-haiku-4-20250514",
      "tools": {
        "write": false,
        "edit": false,
        "bash": false
      }
    }
  }
}
```

### Agent Example (Markdown)

```yaml
---
description: Reviews code for quality and best practices
mode: subagent
model: anthropic/claude-sonnet-4-20250514
temperature: 0.1
tools:
  write: false
  edit: false
  bash: false
---

You are in code review mode. Focus on:
- Code quality and best practices
- Potential bugs and edge cases
- Performance implications
- Security considerations

Provide constructive feedback without making direct changes.
```

### Skill Example

```yaml
---
name: git-release
description: Create consistent releases and changelogs
license: MIT
compatibility: opencode
metadata:
  audience: maintainers
  workflow: github
---

## What I do
- Draft release notes from merged PRs
- Propose a version bump
- Provide a copy-pasteable `gh release create` command

## When to use me
Use this when you are preparing a tagged release.
```

### Permissions Example

```json
{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "bash": {
      "*": "ask",
      "git *": "allow",
      "grep *": "allow"
    },
    "edit": {
      "*.env": "deny",
      "*.mdx": "allow"
    }
  }
}
```

---

## String Substitutions

### Commands

- `$ARGUMENTS` - All arguments
- `$1`, `$2`, `$3` - Positional arguments
- `` `!command` `` - Shell output injection
- `@filename` - File content inclusion

---

## Comparison with Claude Code

| Aspect            | Claude Code                    | OpenCode                               |
| ----------------- | ------------------------------ | -------------------------------------- |
| Tool Names        | PascalCase (`Bash`, `Read`)    | Lowercase (`bash`, `read`)             |
| Permission System | Enum (`permissionMode`)        | Object with tool-specific config       |
| Permission Values | `default`, `acceptEdits`, etc. | `allow`, `ask`, `deny`                 |
| Model Format      | Alias or full name             | `provider/model-id`                    |
| Tool Config       | Flat arrays                    | Boolean objects                        |
| Agent Modes       | Built-in types                 | `primary`, `subagent`, `all`           |
| Skills            | Optional `name` field          | Required `name` field                  |
| Additional Fields | `hooks` (lifecycle)            | `license`, `compatibility`, `metadata` |
| Temperature       | Not supported                  | Supported (0.0-1.0)                    |
| MaxSteps          | Not supported                  | Supported (> 0)                        |
| Hidden            | Not supported                  | Supported (boolean)                    |

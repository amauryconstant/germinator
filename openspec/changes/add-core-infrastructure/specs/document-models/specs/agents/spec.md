# Agents Model

## Summary

Define data model for Claude Code Agents (subagents).

## Scope

- **Package**: `pkg/models/`
- **Files**: `models.go`, `models_test.go`
- **Testing**: Table-driven tests for validation logic

## Requirements

### Agents Model

### Specification

Agents MUST support the following YAML frontmatter fields:

#### Required Fields
- `name` (string): Unique identifier using lowercase letters and hyphens
- `description` (string): When Claude should delegate to this subagent

#### Optional Fields
- `tools` ([]string): Tools subagent can use. Inherits all tools if omitted.
- `disallowedTools` ([]string): Tools to deny, removed from inherited or specified list.
- `model` (string): Model to use: `sonnet`, `opus`, `haiku`, or `inherit`. Defaults to `sonnet`.
- `permissionMode` (string): Permission mode: `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, or `plan`.
- `skills` ([]string): Skills to load into subagent's context at startup. Subagents don't inherit skills from parent conversation.

### Validation Rules
- Name MUST be lowercase letters and hyphens only (no numbers or other characters)
- Both `name` and `description` are REQUIRED
- `model` MUST be one of: `sonnet`, `opus`, `haiku`, `inherit` (if omitted defaults to `sonnet`)
- `permissionMode` MUST be one of: `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, `plan`

### Examples

#### Basic Agent
```yaml
---
name: code-reviewer
description: Reviews code for quality and best practices.
tools: Read, Glob, Grep, Bash
model: sonnet
---

You are a code reviewer ensuring high standards of code quality and security.
```

#### Agent with tool restrictions
```yaml
---
name: safe-researcher
description: Research agent with restricted capabilities
tools: Read, Grep, Glob
disallowedTools: Write, Edit
permissionMode: default
---
```

#### Agent with permission mode
```yaml
---
name: aggressive-optimizer
description: Performance optimizer with bypassed permissions
permissionMode: bypassPermissions
model: opus
---
```

#### Agent with skills
```yaml
---
name: database-query-agent
description: Expert database query specialist
skills: sql-query, data-analysis
tools: Bash, Read
---

You are a database specialist with access to SQL query and data analysis skills.
```

## References

- [Claude Code Subagents Documentation](https://code.claude.com/docs/en/sub-agents)
- [Claude Code Built-in Subagents](https://code.claude.com/docs/en/sub-agents#built-in-subagents)
- [Claude Code Permission Modes](https://code.claude.com/docs/en/sub-agents#permission-modes)

# Document Models

## Summary

Define data models for Claude Code document types: Skills, Agents, Commands, and Memory.

## Scope

- **Package**: `pkg/models/`
- **Files**: `models.go`, `models_test.go`
- **Testing**: Table-driven tests for validation logic

## Requirements

### Skills Model

### Specification

Skills MUST support the following YAML frontmatter fields:

#### Required Fields
- `name` (string): Skill name. Must use lowercase letters, numbers, and hyphens only (max 64 characters). Should match directory name.
- `description` (string): What Skill does and when to use it (max 1024 characters). Claude uses this to decide when to apply Skill.

#### Optional Fields
- `allowed-tools` ([]string): Tools Claude can use without asking permission when this Skill is active. Supports comma-separated values or YAML-style lists.
- `model` (string): Model to use when this Skill is active (e.g., `claude-sonnet-4-20250514`). Defaults to conversation's model.
- `context` (string): Set to `fork` to run Skill in a forked sub-agent context with its own conversation history.
- `agent` (string): Specify which agent type to use when `context: fork` is set (e.g., `Explore`, `Plan`, `general-purpose`, or a custom agent name). Only applicable when combined with `context: fork`.
- `user-invocable` (boolean): Controls whether Skill appears in slash command menu. Does not affect `Skill` tool or automatic discovery. Defaults to `true`.

### Validation Rules
- Name MUST be lowercase letters, numbers, and hyphens only
- Name MUST NOT exceed 64 characters
- Description MUST NOT exceed 1024 characters
- Both `name` and `description` are REQUIRED

### Examples

#### Basic Skill
```yaml
---
name: explaining-code
description: Explains code with visual diagrams and analogies. Use when explaining how code works, teaching about a codebase, or when user asks "how does this work?"
---
```

#### Skill with tool restrictions
```yaml
---
name: reading-files-safely
description: Read files without making changes. Use when you need read-only file access.
allowed-tools:
  - Read
  - Grep
  - Glob
---
```

#### Skill with fork context
```yaml
---
name: code-analysis
description: Analyze code quality and generate detailed reports
context: fork
---
```

## References

- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills)
- [Claude Code Skills Best Practices](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/best-practices)
- [Agent SDK Skills Reference](https://docs.claude.com/en/docs/agent-sdk/skills)

# Slash Commands Model

## Summary

Define data model for Claude Code Slash Commands.

## Scope

- **Package**: `pkg/models/`
- **Files**: `models.go`, `models_test.go`
- **Testing**: Table-driven tests for validation logic

## Requirements

### Commands Model

### Specification

Commands MUST support the following YAML frontmatter fields:

#### Required Fields
- None (all fields are optional)

#### Optional Fields
- `allowed-tools` ([]string): List of tools command can use.
- `argument-hint` (string): The arguments expected for the slash command (e.g., `add [tagId] | remove [tagId] | list`). This hint is shown to users for auto-completion.
- `context` (string): Set to `fork` to run command in a forked sub-agent context with its own conversation history.
- `agent` (string): Specify which agent type to use when `context: fork` is set (e.g., `Explore`, `Plan`, `general-purpose`). Only applicable when combined with `context: fork`.
- `description` (string): Brief description of command. If omitted, uses first line from prompt.
- `model` (string): Specific model string. Inherits from conversation if omitted.
- `disable-model-invocation` (boolean): Whether to prevent `Skill` tool from calling this command (default: false).

### Command Name Source
- The command name is derived from the **markdown filename** (without `.md` extension), not from a frontmatter field.
- Subdirectories can be used for namespacing (e.g., `.claude/commands/frontend/test.md` creates `/test` with description "(project:frontend)")

### Validation Rules
- No fields are required
- `description` is optional (uses first line of body if omitted)
- Name is NOT validated in frontmatter (comes from filename)

### Examples

#### Basic Command
```yaml
---
description: Review this code for security vulnerabilities.
---

Review this code for security vulnerabilities.
```

#### Command with Arguments
```yaml
---
argument-hint: [pr-number] [priority] [assignee]
description: Review pull request.
---

Review PR #$1 with priority $2 and assign to $3.
```

#### Command with Allowed Tools
```yaml
---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*)
description: Create a git commit.
---

Create a git commit using staged changes.
```

#### Command with Fork Context
```yaml
---
context: fork
agent: Explore
description: Perform comprehensive codebase search.
---

Search the entire codebase and return results.
```

## References

- [Claude Code Slash Commands Documentation](https://code.claude.com/docs/en/slash-commands)

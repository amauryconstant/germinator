# Memory Model

## Summary

Define data model for Claude Code Memory handling.

## Scope

- **Package**: `pkg/models/`
- **Files**: `models.go`, `models_test.go`
- **Testing**: Table-driven tests for parsing logic

## Requirements

### Memory Architecture

Memory in Claude Code is **markdown content** without structured frontmatter, with these exceptions:

### File Types and Locations

#### CLAUDE.md Files
- **Project Memory**: `./CLAUDE.md` or `./.claude/CLAUDE.md` - Team-shared project instructions
- **User Memory**: `~/.claude/CLAUDE.md` - Personal preferences for all projects
- **Enterprise Memory**: Platform-specific path managed by organization

#### CLAUDE.local.md
- **Project Memory (local)**: `./CLAUDE.local.md` - Personal project-specific preferences
- Automatically added to `.gitignore`

#### Rules Files
- **Project Rules**: `./.claude/rules/*.md` - Modular, topic-specific project instructions
- All `.md` files in `.claude/rules/` are automatically loaded as project memory

### Rules Frontmatter

Rules files (`.claude/rules/*.md`) MAY include optional YAML frontmatter:

#### Optional Fields
- `paths` ([]string): Glob patterns for path-specific rules. Supports standard glob patterns like `**/*.ts`, `src/**/*.{ts,tsx}`

### Import Syntax

CLAUDE.md files can import additional files using `@path/to/import` syntax:

- Both relative and absolute paths are allowed
- Max recursion depth: 5 hops
- Imports are NOT evaluated inside markdown code spans and code blocks
- Imports are an alternative to `CLAUDE.local.md`

### Validation Rules
- Memory files are pure markdown content
- No frontmatter validation required (except optional `paths` for rules files)
- `@path/to/import` syntax parsing should be supported
- Import recursion should not exceed 5 hops

### Examples

#### Basic CLAUDE.md
```markdown
# Project Overview

This project uses Go 1.25 with Cobra CLI framework.

## Build Commands

- `go build -o germinator ./cmd`
- `go test ./...`
```

#### CLAUDE.md with Imports
```markdown
See @README for project overview and @package.json for available npm commands.

# Additional Instructions
- git workflow @docs/git-instructions.md
```

#### Rules File with Path Filtering
```yaml
---
paths:
  - "src/api/**/*.ts"
  - "lib/**/*.ts"
  - "tests/**/*.test.ts"
---

# API Development Rules

- All API endpoints must include input validation
- Use standard error response format
- Include OpenAPI documentation comments
```

## References

- [Claude Code Memory Documentation](https://code.claude.com/docs/en/memory)
- [Claude Code Project Memory](https://code.claude.com/docs/en/memory#set-up-project-memory)
- [Claude Code Rules](https://code.claude.com/docs/en/memory#modular-rules-with-claude%2Frules%2F)

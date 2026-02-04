**Location**: `internal/`
**Parent**: See `/AGENTS.md` for project overview

---

# Core Package Patterns

## Structure

- `core/` - Document parsing, loading, serialization, template functions
- `models/` - Document data models and validation
- `services/` - Platform-specific transformation (see `internal/services/AGENTS.md`)

---

# Core Package (`internal/core/`)

**Location**: See `core/AGENTS.md` for implementation details.

Document Loading: `DetectType → ParseDocument → LoadDocument → Validate`
- Detects type from filename patterns (agent-*, command-*, skill-*, memory-*)
- Memory: full content; others: YAML frontmatter between `---`

Serialization: `getDocType → getTemplatePath → template.Execute()`
- Templates at `config/templates/{platform}/{docType}.tmpl`
- Custom `transformPermissionMode`: Claude Code enum → OpenCode permission object
- Sprig functions: lower, upper, trim, join, etc.

---

# Models Package (`internal/models/`)

## Document Types

### Agent
- Tools (lowercase list)
- DisallowedTools (lowercase list, set false in OpenCode)
- PermissionMode (Claude Code enum)
- Model (platform-specific ID)
- Skills (skipped in OpenCode)
- Mode (OpenCode: primary/subagent/all)
- Temperature (*float64, nil omits, 0.0 renders)
- MaxSteps (> 0 for OpenCode)
- Hidden/Disable (omit when false)

### Command
- Tools permissions
- Subtask (boolean)
- Context (fork)
- Agent reference
- Model (platform-specific ID)
- ArgumentHint (skipped in OpenCode)
- DisableModelInvocation (skipped in OpenCode)

### Skill
- Tool permissions
- License
- Compatibility
- Metadata
- Hooks
- Model (platform-specific ID)
- Context (fork)
- Agent reference
- UserInvocable (skipped in OpenCode)

### Memory
- Paths → @ file references (one per line)
- Content → narrative context (rendered as-is)

## Validation

Platform-specific validation rules in `services/AGENTS.md`. Returns `[]error` for multiple issues. Use `errorCount` in tests for exact verification.

---

# Testing Patterns

Table-driven tests with descriptive names. End-to-end: `LoadDocument → Validate → RenderDocument → Verify output`.

See `test/AGENTS.md` for golden file testing patterns.

---

# Error Handling

Validation: Non-fatal returns doc with errors, fatal aborts. Parse/template errors wrapped with `failed to <action>: %w`. See `core/AGENTS.md` for details.

---

# File Organization

## Test Files

- `<package>_test.go` - Unit tests
- `integration_test.go` - Integration tests
- `<package>_golden_test.go` - Golden file tests

## Source Files

- `<package>.go` - Main implementation
- `doc.go` - Package documentation

---

# Constants

See `internal/models/constants.go` for:
- Permission mode enums
- Document type constants
- Platform-specific constants

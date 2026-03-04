**Location**: `internal/`
**Parent**: See `/AGENTS.md` for project overview

---

# Core Package Patterns

## Structure

- `application/` - Service interfaces and DTOs for dependency injection (see `internal/application/AGENTS.md`)
- `config/` - Configuration loading, XDG paths, TOML parsing (Koanf-based)
- `core/` - Document parsing, loading, serialization, template functions
- `errors/` - Typed domain errors (ParseError, ValidationError, TransformError, FileError, ConfigError)
- `models/` - Document data models (validation moved to `internal/validation/`)
- `services/` - Service implementations (see `internal/services/AGENTS.md`)
- `validation/` - Functional validation pipeline with `Result[T]` (see `internal/validation/AGENTS.md`)

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

**Moved to `internal/validation/`**: Standalone validator functions with `Result[T]` pattern.

See `internal/validation/AGENTS.md` for:
- `Result[T]` type for functional error handling
- `ValidationPipeline[T]` for composable validation
- Generic validators: `ValidateAgent()`, `ValidateCommand()`, `ValidateSkill()`, `ValidateMemory()`
- OpenCode validators: `ValidateAgentOpenCode()`, etc.

---

# Testing Patterns

Table-driven tests with descriptive names. End-to-end: `LoadDocument → Validate → RenderDocument → Verify output`.

See `test/AGENTS.md` for golden file testing patterns.

---

# Error Handling

**Typed errors** in `internal/errors/`:
| Type | Fields | Use Case |
|------|--------|----------|
| ParseError | Path, Message, Cause | Malformed YAML, unrecognized document type |
| ValidationError | request, field, value, message, suggestions, context | Invalid field values |
| TransformError | Operation, Platform, Message, Cause | Template/render failures |
| FileError | Path, Operation, Message, Cause | File read/write errors |
| ConfigError | Field, Value, Available, Message | Invalid configuration |

Constructors: `NewParseError`, `NewValidationError`, `NewTransformError`, `NewFileError`, `NewConfigError`

**ValidationError signature**: `NewValidationError(request, field, value, message string)`

**Immutable builders** (return copy):
```go
err := errors.NewValidationError("Agent", "name", "", "name is required")
err.WithSuggestions([]string{"Use lowercase with hyphens"})
err.WithContext("agent definition file")
```

**Getters**: `Field()`, `Value()`, `Message()`, `Request()`, `Suggestions()`, `Context()`

All types implement `Unwrap()` for `errors.As` chains. FileError has `IsNotFound()` helper.

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

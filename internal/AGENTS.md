**Location**: `internal/`
**Parent**: See `/AGENTS.md` for project overview

---

# Core Package Patterns

## Structure

- `application/` - Service interfaces and DTOs for dependency injection (see `internal/application/AGENTS.md`)
- `domain/` - **Consolidated domain layer**: types, errors, validation, results (see `internal/domain/AGENTS.md`)
- `service/` - Service implementations (see `internal/service/AGENTS.md`)
- `infrastructure/` - **Unified infrastructure layer**:
  - `infrastructure/parsing/` - Document loading, parsing, platform detection
  - `infrastructure/serialization/` - Serialization, template functions
  - `infrastructure/adapters/` - Platform adapters (Claude Code, OpenCode)
  - `infrastructure/config/` - Configuration loading, XDG paths, TOML parsing
  - `infrastructure/library/` - Library system, resource management, preset grouping

**Note**: The following packages have been consolidated into `internal/domain/`:
- `errors/` → now `domain/errors.go` (Typed domain errors with builder pattern)
- `models/` → now split into `domain/*.go` (Agent, Command, Skill, Memory, Platform types)
- `validation/` → now `domain/validation.go`, `domain/result.go`, `domain/opencode/` (Validation pipeline and Result[T])

---

# Infrastructure Package (`internal/infrastructure/`)

**Unified infrastructure layer** organized by concern:

## Parsing (`internal/infrastructure/parsing/`)

Document Loading: `DetectType → ParseDocument → LoadDocument → Validate`
- Detects type from filename patterns (agent-*, command-*, skill-*, memory-*)
- Memory: full content; others: YAML frontmatter between `---`

## Serialization (`internal/infrastructure/serialization/`)

Serialization: `getDocType → getTemplatePath → template.Execute()`
- Templates at `config/templates/{platform}/{docType}.tmpl`
- Custom `transformPermissionMode`: Claude Code enum → OpenCode permission object
- Sprig functions: lower, upper, trim, join, etc.

**See also**:
- [infrastructure/parsing/AGENTS.md](infrastructure/parsing/AGENTS.md) for detailed parsing documentation
- [infrastructure/serialization/AGENTS.md](infrastructure/serialization/AGENTS.md) for detailed serialization documentation

---

# Domain Package (`internal/domain/`)

**Consolidated domain layer** containing all business types with no external dependencies.

**Location**: See `domain/AGENTS.md` for complete documentation.

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

**Functional validation pipeline** with `Result[T]` type and composable validators.

Located in `internal/domain/`:
- `validation.go` - Generic validators: `ValidateAgent()`, `ValidateCommand()`, `ValidateSkill()`, `ValidateMemory()`
- `opencode/` - OpenCode-specific validators: `ValidateAgentOpenCode()`, etc.
- `pipeline.go` - `ValidationPipeline[T]` for chaining validators
- `result.go` - `Result[T]` type for functional error handling

## Errors

**Typed domain errors** in `internal/domain/errors.go` with immutable builder pattern.

Error types: ParseError, ValidationError, TransformError, FileError, ConfigError

Features:
- Immutable builder pattern with `WithSuggestions()`, `WithContext()`
- Getters for programmatic access
- `Unwrap()` for error chaining

## Results

**Service result types** in `internal/domain/results.go`:

Types: TransformResult, ValidateResult, CanonicalizeResult, InitializeResult

Used by service interfaces in `internal/application/` for operation outcomes.

---

# Testing Patterns

Table-driven tests with descriptive names. End-to-end: `LoadDocument → Validate → RenderDocument → Verify output`.

See `test/AGENTS.md` for golden file testing patterns.

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

**Note**: `internal/models/` directory remains but only contains constants.go and doc.go (non-domain artifacts).

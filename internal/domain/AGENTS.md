**Location**: `internal/domain/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Domain Package

**Consolidated domain layer** containing all business types with no external dependencies.

## Domain Purity

This package has **zero external dependencies** (only stdlib and internal/domain).

Enforced via depguard in `.golangci.yml`:
```yaml
depguard:
  rules:
    domain:
      files:
        - internal/domain/**
      allow:
        - $gostd
        - internal/domain
```

**Purpose**: Prevents architectural drift - domain types remain pure and independent.

---

# Document Types

| Type | File | Key Fields |
|------|------|------------|
| Agent | `agent.go` | Tools, PermissionMode, Model, Skills, Mode, MaxSteps |
| Command | `command.go` | Tools, Subtask, Context, Agent, Model |
| Skill | `skill.go` | Tools, License, Compatibility, Hooks, Model |
| Memory | `memory.go` | Paths (→ @ refs), Content (narrative) |
| Platform | `platform.go` | Platform, PermissionPolicy |

**Platform-specific behavior**: See `internal/adapters/{platform}/` for transformation logic.

---

# Errors

**Typed domain errors** in `errors.go` with immutable builder pattern.

| Type | Purpose |
|------|---------|
| ParseError | Document parsing failures |
| ValidationError | Validation failures |
| TransformError | Transformation failures |
| FileError | File I/O failures |
| ConfigError | Configuration failures |

## Builder Pattern

```go
err := NewParseError("invalid YAML").
    WithSuggestion("Check YAML syntax").
    WithContext(map[string]any{"line": 42})
```

Features:
- `WithSuggestions()` - Add remediation hints
- `WithContext()` - Add debugging context
- `Unwrap()` - Error chaining support
- Getters for programmatic access (`Suggestions()`, `Context()`)

---

# Result[T]

**Functional error handling** in `result.go`.

```go
type Result[T any] struct {
    Value T
    Error error
}

func (r Result[T]) Unwrap() (T, error) { return r.Value, r.Error }
func (r Result[T]) OrElse(defaultValue T) T { /* ... */ }
```

**Usage**: Prefer over returning `(T, error)` for composable pipelines.

---

# Validation Pipeline

**Composable validators** in `validation.go` and `opencode/`.

## Generic Validators

```go
func ValidateAgent(agent *Agent) Result[*Agent]
func ValidateCommand(cmd *Command) Result[*Command]
func ValidateSkill(skill *Skill) Result[*Skill]
func ValidateMemory(mem *Memory) Result[*Memory]
```

## OpenCode-Specific Validators

Located in `opencode/`:

| Validator | Purpose |
|-----------|---------|
| ValidateAgentOpenCode | OpenCode agent rules |
| ValidateCommandOpenCode | OpenCode command rules |
| ValidateSkillOpenCode | OpenCode skill rules |
| ValidateMemoryOpenCode | OpenCode memory rules |

## ValidationPipeline[T]

```go
pipeline := ValidationPipeline[Agent]{}
pipeline.Add(ValidateAgent)
pipeline.Add(ValidateAgentOpenCode)

result := pipeline.Run(&agent)
if !result.IsOk() {
    return result
}
```

---

# Service Results

**Operation outcome types** in `results.go`.

| Type | Purpose |
|------|---------|
| TransformResult | Transformation output |
| ValidateResult | Validation errors (with `Valid()` method) |
| CanonicalizeResult | Canonicalization output |
| InitializeResult | Per-resource installation result |

**Used by**: Service interfaces in `internal/application/`.

---

# File Organization

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `agent.go` | Agent type |
| `command.go` | Command type |
| `skill.go` | Skill type |
| `memory.go` | Memory type |
| `platform.go` | Platform/PermissionPolicy types |
| `errors.go` | Typed errors with builder |
| `validation.go` | Generic validators |
| `result.go` | Result[T] type |
| `results.go` | Service result types |
| `pipeline.go` | ValidationPipeline[T] |
| `opencode/` | OpenCode validators (subdirectory) |

---

# Migration History

This package consolidates types from previous packages:

| Previous | Current |
|----------|---------|
| `internal/models/canonical/` | Split into `domain/*.go` |
| `internal/errors/` | `domain/errors.go` |
| `internal/validation/` | `domain/validation.go`, `result.go`, `opencode/` |
| `internal/application/results.go` | `domain/results.go` |

**Exception**: `internal/application/requests.go` remains in application layer (depends on `library.Library`).

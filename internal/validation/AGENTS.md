**Location**: `internal/validation/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Validation Package

Functional validation pipeline with `Result[T]` type for composable, early-exit validation.

## Files

| File | Purpose |
|------|---------|
| `result.go` | `Result[T]` type for functional error handling |
| `pipeline.go` | `ValidationPipeline[T]` that chains validators |
| `validators.go` | Generic validators (Agent, Command, Skill, Memory) |
| `opencode/validators.go` | OpenCode-specific validators |

---

# Result[T] Pattern

```go
type Result[T any] struct {
    Value T
    Error error
}

func NewResult[T any](value T) Result[T]
func NewErrorResult[T any](err error) Result[T]
func (r Result[T]) IsSuccess() bool
func (r Result[T]) IsError() bool
```

**Usage**: Enables functional error handling without exceptions. Check `IsError()` before accessing `Value`.

---

# Validation Pipeline

```go
type ValidationFunc[T any] func(T) Result[bool]

func NewValidationPipeline[T any](validations ...ValidationFunc[T]) *ValidationPipeline[T]
func (p *ValidationPipeline[T]) Validate(input T) Result[bool]
```

**Pipeline collects all errors** (no early exit on first failure). Errors joined via `errors.Join()`.

---

# Generic Validators

| Function | Validates |
|----------|-----------|
| `ValidateAgentName()` | Required, matches `^[a-z0-9]+(-[a-z0-9]+)*$` |
| `ValidateAgentDescription()` | Required |
| `ValidateAgentPermissionPolicy()` | Valid enum if specified |
| `ValidateAgent()` | Composes all agent validators |
| `ValidateCommandName()` | Required |
| `ValidateCommandDescription()` | Required |
| `ValidateCommandExecution()` | Context must be `fork` if specified |
| `ValidateCommand()` | Composes all command validators |
| `ValidateSkillName()` | Required, 1-64 chars, matches pattern |
| `ValidateSkillDescription()` | Required, 1-1024 chars |
| `ValidateSkillExecution()` | Context must be `fork` if specified |
| `ValidateSkill()` | Composes all skill validators |
| `ValidateMemory()` | Either paths or content required |

---

# OpenCode Validators

Located in `internal/validation/opencode/`:

| Function | Validates |
|----------|-----------|
| `ValidateAgentMode()` | Mode in `primary`, `subagent`, `all` |
| `ValidateAgentTemperature()` | Range [0.0, 1.0] |
| `ValidateAgentOpenCode()` | Composes OpenCode agent validators |
| `ValidateCommandOpenCode()` | No additional rules (pass-through) |
| `ValidateSkillOpenCode()` | No additional rules (pass-through) |

---

# Service Integration

Services compose generic + platform validators:

```go
// Generic validation
if r := validation.ValidateAgent(doc); r.IsError() {
    return &ValidateResult{Errors: []error{r.Error}}, nil
}

// Platform-specific validation
if req.Platform == "opencode" {
    if r := opencode.ValidateAgentOpenCode(doc); r.IsError() {
        return &ValidateResult{Errors: []error{r.Error}}, nil
    }
}
```

---

# ValidationError Integration

Validators return `Result[bool]` with `ValidationError`:

```go
return NewErrorResult[bool](
    domainerrors.NewValidationError("Agent", "name", a.Name, "name is required"),
)
```

See `internal/AGENTS.md` for `ValidationError` builders (`WithSuggestions`, `WithContext`).

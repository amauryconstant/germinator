**Location**: `internal/core/`
**Parent**: See `/internal/AGENTS.md` for core patterns
**Skill references**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`, `@.opencode/skills/golang-error-handling/SKILL.md` (typed-error builder / `Unwrap` / `errors.As` patterns), `@.opencode/skills/golang-samber-lo/SKILL.md` (allowed collection helpers in this package)

---

# Core Package

**Consolidated core layer** containing all business types with no external dependencies.

## Core Purity

This package has **zero external dependencies** (only stdlib and `samber/lo`). Enforced via `depguard` in `.golangci.yml` (see rule in `/internal/AGENTS.md`). Prevents architectural drift — core types remain pure and independent.

### The one self-import exception

`internal/core/opencode/validators.go` is the only file in the entire `internal/core/` tree with a non-stdlib production import — it imports `gitlab.com/amoconst/germinator/internal/core` to reach the canonical document types (`*core.Agent`, `*core.Command`, etc.) when applying the OpenCode-specific validation rules. The depguard rule accommodates this via its third allow-list entry. **No other file in `internal/core/` may import a project-internal package**; this is the documented exception.

---

# Document Types

| Type | File | Key Fields |
|------|------|------------|
| Agent | `agent.go` | Tools, PermissionMode, Model, Skills, Mode, MaxSteps |
| Command | `command.go` | Tools, Subtask, Context, Agent, Model |
| Skill | `skill.go` | Tools, License, Compatibility, Hooks, Model |
| Memory | `memory.go` | Paths (→ @ refs), Content (narrative) |
| Platform | `platform.go` | Platform, PermissionPolicy |

**Platform-specific behavior**: See `internal/{claude-code,opencode}/` for transformation logic.

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
| NotFoundError | Missing-entity lookup (carries `Entity`, `Key`); `output.FormatError` renders `Error: not found: <key>` to stderr |
| OperationError | Per-operation failure (carries `Op`, `Resource`, `Cause`); `output.FormatError` renders `Error: <op>: <resource>` to stderr |
| InitializeError | Per-resource install failure (carries `Ref`, `InputPath`, `OutputPath`, `Cause`); builder `WithSuggestions`/`WithContext` |
| PartialSuccessError | Aggregated install outcome (`Succeeded`, `Failed`, `[]InitializeError`); `cmdutil.ExitCodeFor` returns 0 when `Succeeded > 0` |
| UsageError | CLI flag validation failure (carries `Flag`, `Reason`, optional `Suggestions`); `output.FormatError` renders `Error: <flag>: <reason>` to stderr; maps to exit 2 |
| CobraUsageError | Sentinel wrapping Cobra arg-validation errors (`MustNewCobraUsageError` panics on nil cause); maps to exit 2 |

Errors carry semantic meaning only — no exit codes. Builder pattern (`WithSuggestions`, `WithContext`, `Unwrap`); see `golang-error-handling` skill for full reference.

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

**Used by**: cmd-layer run functions and shell-package operations that need to surface domain-typed failures.

**Location**: `internal/service/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Services Package

Service implementations for `internal/application` interfaces.

## Files

| File | Purpose |
|------|---------|
| `transformer.go` | Implements `application.Transformer` |
| `validator.go` | Implements `application.Validator` |
| `initializer.go` | Implements `application.Initializer` |
| `canonicalizer.go` | Implements `application.Canonicalizer` |
| `*_test.go` | Unit tests for each service |

---

# Interface Implementation

Each service is a struct with methods implementing `application` interfaces:

```go
type transformer struct{}

func NewTransformer() application.Transformer {
    return &transformer{}
}

func (t *transformer) Transform(ctx context.Context, req *application.TransformRequest) (*application.TransformResult, error) {
    // Implementation
}

// Compile-time interface check
var _ application.Transformer = (*transformer)(nil)
```

**Constructors**: `NewTransformer()`, `NewValidator()`, `NewCanonicalizer()`, `NewInitializer()`

**Pattern**: Constructor returns interface type, implementation is private struct.

---

# Transformation Pipeline

## Transformer.Transform

```
core.LoadDocument → core.RenderDocument → WriteFile
```

**Request**: `TransformRequest{InputPath, OutputPath, Platform}`

**Result**: `TransformResult{OutputPath}`

**Error handling**: Returns wrapped errors at each step.

**File permissions**: `0644` (rw-r--r--)

---

# Initialization Pipeline

## Initializer.Initialize

Batch transformation of library resources to platform-specific output files.

```
library.ResolveResource → core.LoadDocument → core.RenderDocument → WriteFile
```

**Request**: `InitializeRequest{Library, Platform, OutputDir, Refs, DryRun, Force}`

**Result**: `[]InitializeResult` - Per-resource results with Ref, InputPath, OutputPath, Error

**Error handling**: Fail-fast - stops on first error.

**File permissions**: `0644` (rw-r--r--)

**Dry-run mode**: Prints what would be written without creating files.

**Force mode**: Overwrites existing files; otherwise returns error if file exists.

---

# Validation Pipeline

## Validator.Validate

Two-stage validation using `internal/validation/` package:

```
core.DetectType → core.ParseDocument → validation.ValidateAgent() → opencode.ValidateAgentOpenCode()
```

**Request**: `ValidateRequest{InputPath, Platform}`

**Result**: `ValidateResult{Errors []error}` with `Valid() bool` method

**Stage 1**: Generic validation via `validation.ValidateAgent()`, `ValidateCommand()`, etc.

**Stage 2**: Platform-specific validation via `validation/opencode.ValidateAgentOpenCode()`, etc.

**Dual-return pattern**: `error` = fatal (couldn't validate), `result.Errors` = validation issues.

See `internal/validation/AGENTS.md` for Result[T] pattern and validator composition.

---

# OpenCode-Specific Validation

OpenCode validators located in `internal/validation/opencode/`:

| Function | Rules |
|----------|-------|
| `ValidateAgentMode()` | Mode: `primary`, `subagent`, or `all` |
| `ValidateAgentTemperature()` | Range [0.0, 1.0] |
| `ValidateAgentOpenCode()` | Composes mode + temperature |

See `internal/validation/AGENTS.md` for complete validator reference.

---

# Platform-Specific Behavior

## Claude Code

**Transformation**: Direct pass-through (source format based on Claude Code)

**Validation**:
- Model: Short names (sonnet, opus, haiku)
- Permission mode: Enum validation

## OpenCode

**Transformation**: Rendered using OpenCode templates

**Template paths**: `config/templates/opencode/{agent,command,skill,memory}.tmpl`

**Transformations**:
- Tools: Lowercase
- Permission modes: Via `transformPermissionMode` template function
- Agent: Omit `name`, add `mode` (default: all)
- Temperature: Render as float or omit if nil
- Boolean: Omit when false

**Validation**:
- Mode: `primary`, `subagent`, or `all`
- Temperature: [0.0, 1.0]
- MaxSteps: >= 1

---

# Golden File Testing

See `test/AGENTS.md` for golden file testing patterns, fixture structure, and update procedures.

# Error Handling Patterns

**Validation**: Return doc with errors (non-fatal) or error (fatal).

**Transformation**: Wrap at each step: `fmt.Errorf("failed to <action>: %w", err)`

**Platform-specific validation**: Append to error list after generic validation.

---

# Integration with Core Package

**Dependencies**: `core.LoadDocument`, `core.DetectType`, `core.ParseDocument`, `core.RenderDocument`

**Library**: `internal/infrastructure/library/` provides resource loading, resolution, and output path derivation.

**Validation**: `internal/validation/` provides `Result[T]` pattern and composable validators.

**Models**: `internal/models/canonical/` provides document structs (validation methods removed).

---

# Testing Strategies

See `test/AGENTS.md` for table-driven test patterns and golden file testing.

**Unit tests**: Platform-specific validation with errorCount field.

**Integration tests**: End-to-end transformation workflow.

See `internal/infrastructure/adapters/AGENTS.md` for platform-specific adapter patterns.

---

# Platform Constants

See `internal/models/constants.go` for `PlatformClaudeCode`, `PlatformOpenCode`.

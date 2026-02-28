**Location**: `internal/services/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Services Package

Platform-specific business logic for document transformation and validation.

## Files

- `transformer.go` - Document transformation pipeline
- `transformer_golden_test.go` - Golden file tests
- `transformer_test.go` - Unit tests

---

# Transformation Pipeline

## TransformDocument

```
core.LoadDocument → core.RenderDocument → WriteFile
```

**Parameters**: `inputPath`, `outputPath`, `platform` (claude-code or opencode)

**Error handling**: Returns wrapped errors at each step.

**File permissions**: `0644` (rw-r--r--)

---

# Validation Pipeline

## ValidateDocument

Two-stage validation:

```
core.DetectType → core.ParseDocument → Validate(platform) → ValidateOpenCodeType()
```

**Stage 1**: Generic platform validation via model's `Validate(platform)`

**Stage 2**: OpenCode-specific validation:
- `ValidateOpenCodeAgent` - Agent constraints
- `ValidateOpenCodeCommand` - Command constraints
- `ValidateOpenCodeMemory` - Memory constraints
- `ValidateOpenCodeSkill` - Skill constraints

---

# OpenCode-Specific Validation

## ValidateOpenCodeAgent

- `mode`: Must be `primary`, `subagent`, or `all` (if specified)
- `temperature`: Must be in range [0.0, 1.0] (if specified)
- `maxSteps`: Must be >= 1 (if specified)

## ValidateOpenCodeCommand

- `content` (template): Required field

## ValidateOpenCodeMemory

- `paths`: Must have at least one path or content specified

## ValidateOpenCodeSkill

- `name`: Required field

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

**Models**: `internal/models/` provides document structs and validation methods.

---

# Testing Strategies

See `test/AGENTS.md` for table-driven test patterns and golden file testing.

**Unit tests**: Platform-specific validation with errorCount field.

**Integration tests**: End-to-end transformation workflow.

See `internal/adapters/AGENTS.md` for platform-specific adapter patterns.

---

# Platform Constants

See `internal/models/constants.go` for `PlatformClaudeCode`, `PlatformOpenCode`.

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

**Validation errors**: Return document with errors for non-fatal, return error for fatal.

**Transformation errors**: Wrap at each step: `fmt.Errorf("failed to <action>: %w", err)`

Example:
```go
doc, err := core.LoadDocument(inputPath, platform)
if err != nil {
    return fmt.Errorf("failed to load document: %w", err)
}
rendered, err := core.RenderDocument(doc, platform)
if err != nil {
    return fmt.Errorf("failed to render document: %w", err)
}
```

**Platform-specific validation**:
```go
if platform == models.PlatformOpenCode {
    errs2 = ValidateOpenCodeAgent(d)
}
return append(errs, errs2...), nil
```

---

# Integration with Core Package

## Dependencies

Services depends on core:
- `core.LoadDocument` - Document loading and parsing
- `core.DetectType` - Document type detection
- `core.ParseDocument` - Document parsing
- `core.RenderDocument` - Template rendering

## Models Package

Uses models from `internal/models/`:
- `models.Agent`, `models.Command`, `models.Memory`, `models.Skill`
- `models.PlatformOpenCode`, `models.PlatformClaudeCode`
- Validation methods on models

---

# Testing Strategies

## Unit Tests

```go
tests := []struct {
    name       string
    agent      *models.Agent
    errorCount int
}{
    {name: "valid agent", agent: &models.Agent{Mode: "primary"}, errorCount: 0},
    {name: "invalid mode", agent: &models.Agent{Mode: "invalid"}, errorCount: 1},
}
```

## Golden File Tests

See Golden File Testing section above.

## Integration Tests

End-to-end transformation workflow:
```go
func TestTransformDocument(t *testing.T) {
    tests := []struct {
        name     string
        fixture  string
        platform string
    }{
        {name: "agent to opencode", fixture: "agent-full.md", platform: "opencode"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Load, transform, verify
        })
    }
}
```

---

# Platform Constants

See `internal/models/constants.go` for `PlatformClaudeCode`, `PlatformOpenCode`.

**Location**: `internal/core/`
**Parent**: See `/internal/AGENTS.md` for core patterns
**Root**: See `/AGENTS.md` for project overview

---

# Core Package

Document loading, parsing, and serialization for AI coding assistant configurations.

---

# Files

| File | Purpose |
|------|---------|
| `loader.go` | `LoadDocument`, `DetectType` |
| `parser.go` | `ParseDocument` |
| `serializer.go` | `RenderDocument` |
| `*_test.go` | Unit tests for each component |

---

# Document Loading Pipeline

```
DetectType(path) → ParseDocument(content, docType) → LoadDocument(path)
```

## DetectType

Determines document type from filename pattern:

| Pattern | Type |
|---------|------|
| `agent-*.{yaml,yml}` | Agent |
| `command-*.{yaml,yml,md}` | Command |
| `skill-*.{yaml,yml}` / `SKILL.md` | Skill |
| `memory-*.{yaml,yml}` / `AGENTS.md` | Memory |
| Other | Error: unknown document type |

## ParseDocument

Parses raw content into canonical struct:

| Type | Parsing |
|------|---------|
| Agent | YAML frontmatter between `---` delimiters |
| Command | YAML frontmatter + markdown content |
| Skill | YAML frontmatter + markdown content |
| Memory | Full content (no frontmatter) |

Returns: `*canonical.Agent`, `*canonical.Command`, `*canonical.Skill`, or `*canonical.Memory`

## LoadDocument

Full loading pipeline:

1. Read file content
2. Detect type from path
3. Parse content
4. Set `FilePath` on document

Returns: Document interface + error

---

# Serialization Pipeline

```
getDocType(doc) → getTemplatePath(docType, platform) → template.Execute()
```

## RenderDocument

Renders canonical struct to platform-specific output:

```go
func RenderDocument(doc interface{}, platform string) (string, error)
```

| Platform | Template Path |
|----------|---------------|
| claude-code | `config/templates/claude-code/{docType}.tmpl` |
| opencode | `config/templates/opencode/{docType}.tmpl` |

## Template Resolution

1. Try CWD: `./config/templates/{platform}/{docType}.tmpl`
2. Fall back to relative: `../../config/templates/{platform}/{docType}.tmpl`

## Template Context

Templates receive:
- `.Doc` - Canonical document struct
- Template functions (see below)

---

# Template Functions

Available in all templates via Sprig + custom functions:

## Custom Functions

| Function | Purpose |
|----------|---------|
| `permissionPolicyToPlatform` | Convert canonical enum to platform format |
| `convertToolNameCase` | PascalCase (Claude Code) / lowercase (OpenCode) |

## Sprig Functions

Common functions from [Sprig](http://masterminds.github.io/sprig/):

| Function | Example |
|----------|---------|
| `lower` | `{{ .Name | lower }}` |
| `upper` | `{{ .Name | upper }}` |
| `trim` | `{{ .Content | trim }}` |
| `join` | `{{ .Tools | join ", " }}` |
| `default` | `{{ .Mode | default "all" }}` |
| `indent` | `{{ .Content | indent 2 }}` |

---

# Error Handling

| Operation | Error Message |
|-----------|---------------|
| File read | `failed to read file: %w` |
| Parse | `failed to parse document: %w` |
| Template not found | `template file not found: %s` |
| Template parse | `failed to parse template: %w` |
| Template execute | `failed to execute template: %w` |
| Type detection | `unknown document type: %T` |

All errors use typed errors from `internal/errors/`:

```go
// File read error
return nil, errors.NewFileError(path, "read", "failed to read file", err)

// Parse error
return nil, errors.NewParseError(path, "invalid frontmatter", err).
    WithSuggestions([]string{"Check YAML syntax", "Ensure --- delimiters present"})

// Template error
return "", errors.NewTransformError("render", platform, "template execution failed", err)
```

---

# Helper Functions

## extractContentFromYamlLines

Extracts markdown content from YAML lines (for Command/Skill):

```go
func extractContentFromYamlLines(lines []string, frontmatterEnd int) string
```

Returns content after frontmatter, trimmed of leading whitespace.

---

# Testing

| File | Coverage |
|------|----------|
| `loader_test.go` | Type detection, file loading |
| `parser_test.go` | Frontmatter parsing, content extraction |
| `serializer_test.go` | Template rendering, output verification |
| `integration_test.go` | End-to-end pipeline |

### Test Patterns

Table-driven tests with golden file verification:

```go
tests := []struct {
    name     string
    input    string
    wantErr  bool
}{
    {"valid agent", "agent-valid.yaml", false},
    {"missing frontmatter", "agent-no-frontmatter.yaml", true},
}
```

---

# Integration Points

| Package | Usage |
|---------|-------|
| `internal/models/canonical` | Document structs |
| `internal/errors` | Typed errors |
| `internal/adapters` | Platform-specific functions for templates |
| `config/templates/` | Template files |

---

# See Also

- `internal/models/AGENTS.md` - Canonical struct definitions
- `internal/adapters/AGENTS.md` - Platform adapters and template functions
- `config/AGENTS.md` - Template patterns and permission mappings
- `internal/errors/AGENTS.md` - Error types and patterns

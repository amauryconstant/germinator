**Location**: `internal/core/`
**Parent**: See `/internal/AGENTS.md` for overview
**Root**: See `/AGENTS.md` for core standards

---

# Document Loading

DetectType → ParseDocument → LoadDocument. See `../AGENTS.md` for patterns.

**Parsing**: Memory parses full content, others parse YAML frontmatter.

# Serialization

RenderDocument → template.Execute() at `config/templates/{platform}/{docType}.tmpl`. Uses Sprig functions + custom map.

**Template Path Resolution**: Tries CWD, then `../../` relative path.

**Helper**: `extractContentFromYamlLines` extracts content from YAML lines.

# Error Handling

File read: `failed to read file: %w`
Parse: `failed to parse document: %w`
Template not found: `template file not found: %s`
Template parse: `failed to parse template: %w`
Template execute: `failed to execute template: %w`
Type detection: `unknown document type: %T`

# Testing

`loader_test.go`, `parser_test.go`, `serializer_test.go`, `integration_test.go`.

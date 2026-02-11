## Why

Germinator currently supports transforming canonical YAML format to Claude Code or OpenCode formats, but cannot reverse the transformation. Users with existing Claude Code or OpenCode documents must manually rewrite them to canonical format, creating a significant adoption barrier. This limits the value of the canonical format as a single source of truth.

## What Changes

Add reverse transformation capability to convert platform documents (Claude Code or OpenCode) into canonical YAML format:

- Add `canonicalize` CLI command: `germinator canonicalize <input> <output> --platform <claude-code|opencode> --type <agent|command|skill|memory>`
- Implement `internal/core/platform_parser.go` with `ParsePlatformDocument()` function to parse platform YAML files using existing adapter.ToCanonical() methods
- Add `MarshalCanonical()` function to `internal/core/serializer.go` to serialize canonical models to YAML strings
- Create canonical templates in `config/templates/canonical/` directory (agent.tmpl, command.tmpl, skill.tmpl, memory.tmpl)
- Implement `internal/services/canonicalizer.go` with `CanonicalizeDocument()` function orchestrating the reverse transformation pipeline (Parse → Validate → Marshal → Write)
- **Note**: First iteration requires user to specify document type via `--type` flag (no auto-detection), preserves canonical format simplicity but loses fine-grained platform-specific permissions (e.g., OpenCode command-level permission rules collapse to 5 policies)

## Capabilities

### New Capabilities

- `platform-to-canonical`: CLI command and pipeline for transforming Claude Code or OpenCode documents to canonical YAML format, including platform file parsing, canonical model validation, YAML serialization via templates, and file writing with required flags (--platform, --type)

### Modified Capabilities

None. The `platform-adapters` capability (defined in canonical-format-redesign) already includes ToCanonical() methods for both Claude Code and OpenCode adapters. This change leverages existing conversion logic rather than modifying it.

## Impact

**New files:**

- `cmd/canonicalize.go` - CLI command for reverse transformation
- `internal/core/platform_parser.go` - Parse platform YAML files to canonical models
- `internal/services/canonicalizer.go` - Service layer orchestrating reverse transformation pipeline
- `config/templates/canonical/agent.tmpl` - Agent YAML template
- `config/templates/canonical/command.tmpl` - Command YAML template
- `config/templates/canonical/skill.tmpl` - Skill YAML template
- `config/templates/canonical/memory.tmpl` - Memory YAML template

**Modified files:**

- `internal/core/serializer.go` - Add `MarshalCanonical()` function using canonical templates and canonicalTemplateContext struct

**Tests:**

- New unit tests for `ParsePlatformDocument()` covering all document types and platforms
- New unit tests for `MarshalCanonical()` covering all canonical models
- New integration tests for `CanonicalizeDocument()` end-to-end workflow
- New CLI tests for `canonicalize` command

**Dependencies:** None (reuses existing adapters package and gopkg.in/yaml.v3)

---

## Reverse Transformation Pipeline

```
Platform File (.md)
       │
       ▼
┌─────────────────────────────┐
│ ParsePlatformDocument()     │
│ internal/core/platform_     │
│ parser.go                   │
├─────────────────────────────┤
│ • Read YAML frontmatter     │
│ • Instantiate adapter       │
│   (claudecode.New() or      │
│    opencode.New())          │
│ • adapter.ToCanonical()     │
│ • Extract markdown content  │
└─────────────────────────────┘
       │
       ▼
 Canonical Model
       │
       ▼
┌─────────────────────────────┐
│ canonical.*.Validate()      │
│ internal/models/canonical/  │
│ models.go                   │
├─────────────────────────────┤
│ • Required fields           │
│ • Name format regex         │
│ • Permission policy enum   │
│ • Temperature range         │
│ • Paths OR content (Memory) │
└─────────────────────────────┘
       │
       ▼
 MarshalCanonical()
 internal/core/serializer.go
       │
       ▼
 Canonical YAML (.yaml)
       │
       ▼
 Write File (0644)
       │
       ▼
   Output File
```

**Location**: `internal/adapters/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Adapters Package

Platform-specific adapters transforming between canonical models and platform formats.

---

# Adapter Interface

Defined in `adapter.go`:

```go
type Adapter interface {
    ToCanonical(input map[string]interface{}) (*canonical.Agent, *canonical.Command, *canonical.Skill, *canonical.Memory, error)
    FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)
    PermissionPolicyToPlatform(policy canonical.PermissionPolicy) (interface{}, error)
    ConvertToolNameCase(name string) string
}
```

| Method | Purpose |
|--------|---------|
| `ToCanonical` | Parse platform YAML → canonical struct |
| `FromCanonical` | Canonical struct → platform map (for templates) |
| `PermissionPolicyToPlatform` | Convert canonical policy → platform format |
| `ConvertToolNameCase` | Transform tool name casing (PascalCase/lowercase) |

---

# Platform Implementations

## Claude Code (`claude-code/`)

| File | Purpose |
|------|---------|
| `claude_code_adapter.go` | Adapter implementation |
| `doc.go` | Package documentation |

**ToCanonical**: Parses Claude Code YAML fields into canonical struct

**FromCanonical**: Returns map for template rendering

**PermissionPolicyToPlatform**: Returns enum string (`default`, `acceptEdits`, `dontAsk`, `plan`, `bypassPermissions`)

**ConvertToolNameCase**: PascalCase (`Bash`, `Edit`, `Read`)

## OpenCode (`opencode/`)

| File | Purpose |
|------|---------|
| `opencode_adapter.go` | Adapter implementation |
| `doc.go` | Package documentation |

**ToCanonical**: Parses OpenCode YAML/MD frontmatter into canonical struct

**FromCanonical**: Returns map for template rendering

**PermissionPolicyToPlatform**: Returns nested permission object:
```yaml
permission:
  edit:
    "*": ask | allow | deny
  bash:
    "*": ask | allow | deny
  # ... other tools
```

**ConvertToolNameCase**: lowercase (`bash`, `edit`, `read`)

---

# Permission Policy Mappings

## Canonical → Claude Code

| Canonical | Claude Code |
|-----------|-------------|
| restrictive | default |
| balanced | acceptEdits |
| permissive | dontAsk |
| analysis | plan |
| unrestricted | bypassPermissions |

## Canonical → OpenCode

| Canonical | Edit | Bash | Read | Grep | Glob | List | WebFetch | WebSearch |
|-----------|------|------|------|------|------|------|----------|-----------|
| restrictive | ask | ask | ask | ask | ask | ask | ask | ask |
| balanced | allow | ask | allow | allow | allow | allow | allow | allow |
| permissive | allow | allow | allow | allow | allow | allow | allow | allow |
| analysis | deny | deny | allow | allow | allow | allow | allow | allow |
| unrestricted | allow | allow | allow | allow | allow | allow | allow | allow |

---

# Helper Functions (`helpers.go`)

Shared utilities for adapters:

| Function | Purpose |
|----------|---------|
| `extractString` | Safe string extraction from map |
| `extractStringSlice` | Safe string slice extraction |
| `extractBool` | Safe bool extraction with default |
| `extractFloat64` | Safe float64 extraction |
| `extractMap` | Safe nested map extraction |

---

# Usage Pattern

Adapters are used by:

1. **Canonicalization** (`cmd/canonicalize`): `ToCanonical` to create canonical from platform format
2. **Transformation** (`cmd/adapt`): `FromCanonical` + templates to render platform output
3. **Templates**: `PermissionPolicyToPlatform`, `ConvertToolNameCase` as template functions

---

# Testing

| File | Purpose |
|------|---------|
| `claude-code/claude_code_adapter_test.go` | Claude Code adapter tests |
| `opencode/opencode_adapter_test.go` | OpenCode adapter tests |
| `helpers_test.go` | Helper function tests |

Table-driven tests with platform-specific fixtures.

---

# See Also

- `config/AGENTS.md` - Template rendering patterns
- `internal/services/AGENTS.md` - Service layer adapter usage
- `internal/models/canonical/` - Canonical struct definitions

**Location**: `config/`
**Parent**: See `/AGENTS.md` for project overview

---

# Configuration Templates

Go templates for rendering canonical models to target platforms.

---

# Template Structure

`config/templates/{platform}/{docType}.tmpl`
Platforms: `claude-code`, `opencode`
DocTypes: `agent`, `command`, `memory`, `skill`

Resolution: CWD → `../../config/templates/`

---

# Template Context

- `Doc`: Canonical document struct
- `Adapter`: Platform-specific methods via template functions

See `openspec/specs/transformation/template-rendering/spec.md` for detailed requirements.

---

# Template Functions

| Function | Purpose |
|----------|---------|
| `permissionPolicyToPlatform(policy, platform)` | Convert enum to platform format |
| `convertToolNameCase(name, platform)` | PascalCase (Claude Code) / lowercase (OpenCode) |

---

# Permission Policy Mapping

## Claude Code

| Canonical | Output |
|-----------|--------|
| restrictive | default |
| balanced | acceptEdits |
| permissive | dontAsk |
| analysis | plan |
| unrestricted | bypassPermissions |

## OpenCode

| Canonical | Edit | Bash | Read | Grep | Glob | List |
|-----------|------|------|------|------|------|------|
| restrictive | ask | ask | ask | ask | ask | ask |
| balanced | allow | ask | allow | allow | allow | allow |
| permissive | allow | allow | allow | allow | allow | allow |
| analysis | deny | deny | allow | allow | allow | allow |
| unrestricted | allow | allow | allow | allow | allow | allow |

webfetch, websearch follow grep/glob/list pattern.

---

# Template Patterns

**Claude Code**:
- Tools: PascalCase (`Bash`)
- PermissionPolicy → `permissionMode` enum
- `behavior.*` flattened
- YAML frontmatter with `---` delimiters
- `targets.claude-code.*` rendered as platform-specific fields

**OpenCode**:
- Tools: boolean map (`bash: true`)
- PermissionPolicy → nested `permission` object
- Omit false booleans, nil pointers
- `targets.platform.*` for platform-specific overrides

**Common transformations**:
- Direct mapping: `name`, `description`, `model`
- `execution.context` → `context`
- `execution.agent` → `agent`
- `execution.subtask` → `subtask` (if true)

See template files for exact field mappings and omission patterns.

See `internal/adapters/AGENTS.md` for platform-specific adapter patterns.

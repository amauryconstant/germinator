**Location**: `config/`
**Parent**: See `/AGENTS.md` for project overview

---

# Configuration Templates

Platform-specific Go templates for rendering canonical models to target formats.

---

# Template Structure

```
config/templates/
 ├── claude-code/
 │   ├── agent.tmpl
 │   ├── command.tmpl
 │   ├── memory.tmpl
 │   └── skill.tmpl
 └── opencode/
     ├── agent.tmpl
     ├── command.tmpl
     ├── memory.tmpl
     └── skill.tmpl
```

---

# Template Rendering

Templates render canonical models to platform-specific output formats. See `openspec/specs/transformation/template-rendering/spec.md` for detailed requirements.

## Template Context

Templates receive a context with two fields:
- `Doc`: Canonical document struct (Agent, Command, Memory, Skill)
- `Adapter`: Platform adapter providing methods (via template functions)

## Template Discovery

Templates are located at `config/templates/{platform}/{docType}.tmpl`:
- Platforms: `claude-code`, `opencode`
- Document types: `agent`, `command`, `memory`, `skill`

Resolution: Search from CWD first, fall back to `../../config/templates/`.

## Template Functions

Custom functions (in addition to Sprig):
- `permissionPolicyToPlatform(policy, platform)`: Converts PermissionPolicy enum to platform-specific format
- `convertToolNameCase(name, platform)`: Converts tool names to platform-specific case

### Permission Policy Conversion

**Claude Code** (via `permissionPolicyToClaudeCode`):
- `restrictive` → `"default"`
- `balanced` → `"acceptEdits"`
- `permissive` → `"dontAsk"`
- `analysis` → `"plan"`
- `unrestricted` → `"bypassPermissions"`

**OpenCode** (via `permissionPolicyToOpenCode`):
- Returns permission map YAML string with all tools (edit, bash, read, grep, glob, list, webfetch, websearch)
- See Permission Policy Mapping table below for action values

### Tool Case Conversion

**Claude Code**: `convertToolNameCase("bash", "claude-code")` → `"Bash"`
- Converts lowercase to PascalCase

**OpenCode**: `convertToolNameCase("bash", "opencode")` → `"bash"`
- Returns lowercase unchanged

---

# Template Function Maps

Templates use Sprig functions + custom adapter methods for permission mapping and tool case conversion.

Accessed via:
```go
template.New(name).Funcs(createTemplateFuncMap()).Parse(content)
```

Sprig provides: `lower`, `upper`, `trim`, `join`, `gt`, etc.

Custom functions:
- `permissionPolicyToClaudeCode`: Maps canonical PermissionPolicy enum to Claude Code's permissionMode
- `permissionPolicyToOpenCode`: Maps canonical PermissionPolicy enum to OpenCode permission object
- `convertToolNameCase`: Converts tool case based on platform (lowercase → PascalCase/lowercase)

---

# Canonical Model Fields

Templates render from canonical models with these field structures:

## Agent
- `name`, `description`: Identity fields
- `tools`, `disallowedTools`: String arrays (lowercase in canonical)
- `permissionPolicy`: Enum (restrictive, balanced, permissive, analysis, unrestricted)
- `behavior`: Object containing mode, temperature, steps, prompt, hidden, disabled
- `model`: String (full provider ID)
- `targets.platform`: Platform-specific overrides

## Command
- `name`, `description`: Identity fields
- `tools`: String array (lowercase in canonical)
- `execution`: Object containing context, subtask, agent
- `arguments`: Object containing hint
- `model`: String (full provider ID)
- `targets.platform`: Platform-specific overrides

## Skill
- `name`, `description`: Identity fields
- `tools`: String array (lowercase in canonical)
- `extensions`: Object containing license, compatibility, metadata, hooks
- `execution`: Object containing context, agent, userInvocable
- `model`: String (full provider ID)
- `targets.platform`: Platform-specific overrides

## Memory
- `paths`: Array of file paths
- `content`: Markdown narrative

---

# Claude Code Templates

## agent.tmpl

YAML frontmatter with `---` delimiters.

Fields rendered from canonical:
- `name`, `description`, `model`: Direct mapping
- `tools`, `disallowedTools`: Case-converted to PascalCase
- `permissionPolicy`: Mapped to `permissionMode` enum
- `behavior.mode`: Flattened to `mode`
- `behavior.temperature`: Flattened to `temperature`
- `behavior.steps`: Flattened to `maxSteps`
- `behavior.prompt`: Flattened to `prompt`
- `behavior.hidden`, `behavior.disabled`: Flattened
- `targets.claude-code.skills`: Rendered as `skills` array
- `targets.claude-code.disableModelInvocation`: Rendered as `disable-model-invocation`

Content rendered after second `---`.

## command.tmpl

Fields rendered from canonical:
- `name`, `description`, `model`: Direct mapping
- `tools`: Case-converted to PascalCase
- `execution.context`: Flattened to `context`
- `execution.subtask`: Flattened to `subtask` (if true)
- `execution.agent`: Flattened to `agent`
- `arguments.hint`: Flattened to `argument-hint`

Lists use YAML syntax: `- item`

## skill.tmpl

Fields rendered from canonical:
- `name`, `description`, `model`: Direct mapping
- `tools`: Case-converted to PascalCase
- `execution.context`: Flattened to `context`
- `execution.agent`: Flattened to `agent`
- `extensions.license`, `extensions.compatibility`, `extensions.metadata`, `extensions.hooks`: Flattened
- `execution.userInvocable`: Flattened to `user-invocable`

## memory.tmpl

Fields rendered from canonical:
- `paths`: Array of file paths
- `content`: Markdown narrative

---

# OpenCode Templates

## agent.tmpl

YAML frontmatter with field omission patterns.

Fields rendered from canonical:
- `description`: Direct mapping
- `behavior.mode`: Flattened to `mode`
- `model`: Direct mapping
- `tools`, `disallowedTools`: Converted to boolean map (lowercase)
- `permissionPolicy`: Mapped to nested `permission` object
- `behavior.temperature`: Flattened to `temperature` (if not nil)
- `behavior.steps`: Flattened to `maxSteps` (if > 0)
- `behavior.hidden`, `behavior.disabled`: Rendered as `hidden`, `disable` (if true)
- `behavior.prompt`: Flattened to `prompt` (if not empty)

Boolean omission (omit when false):
```go
{{- if .Behavior.Hidden}}
hidden: true
{{- end}}
```

Pointer omission (omit when nil):
```go
{{- if .Behavior.Temperature}}
temperature: {{.Behavior.Temperature}}
{{- end}}
```

Tools transformed to boolean map:
```go
{{- range .Tools}}
  {{convertToolNameCase . "opencode"}}: true
{{- end}}
{{- range .DisallowedTools}}
  {{convertToolNameCase . "opencode"}}: false
{{- end}}
```

Permission policy mapping:
```go
{{- if .Doc.PermissionPolicy}}
permission:
{{.Doc.PermissionPolicy | permissionPolicyToOpenCode}}
{{- end}}
```

## command.tmpl

Fields rendered from canonical:
- `description`: Direct mapping
- `model`: Direct mapping
- `execution.context`: Flattened to `context`
- `execution.subtask`: Flattened to `subtask` (if true)
- `execution.agent`: Flattened to `agent`

Note: `tools`, `arguments.hint` are not rendered for OpenCode commands (platform limitation).

## skill.tmpl

Fields rendered from canonical:
- `name`, `description`, `model`: Direct mapping
- `execution.context`: Flattened to `context`
- `execution.agent`: Flattened to `agent`
- `extensions.license`: Flattened
- `extensions.compatibility`: Flattened to YAML list
- `extensions.metadata`: Flattened to YAML map
- `extensions.hooks`: Flattened to YAML map

Map iteration for metadata/hooks:
```go
{{- range $key, $value := .Extensions.Metadata}}
  {{$key}}: "{{$value}}"
{{- end}}
```

## memory.tmpl

Fields rendered from canonical:
- `paths`: Array of file paths
- `content`: Markdown narrative

Paths rendered as `@file/path` references:
```go
{{if .Doc.Paths}}
To reference files, use @ followed by file path.
{{end}}{{if .Doc.Paths}}{{range .Doc.Paths}}
@{{.}}
{{end}}{{end}}{{if .Doc.Content}}{{.Doc.Content}}{{end}}
```

---

# Permission Policy Mapping

## Canonical → Claude Code

| Canonical | Claude Code |
|----------|-------------|
| restrictive | default |
| balanced | acceptEdits |
| permissive | dontAsk |
| analysis | plan |
| unrestricted | bypassPermissions |

## Canonical → OpenCode

| Canonical | Edit | Bash | Read | Grep | Glob | List |
|----------|------|------|------|------|------|------|
| restrictive | ask | ask | ask | ask | ask | ask |
| balanced | allow | ask | allow | allow | allow | allow |
| permissive | allow | allow | allow | allow | allow | allow |
| analysis | deny | deny | allow | allow | allow | allow |
| unrestricted | allow | allow | allow | allow | allow | allow |

Additional tools in OpenCode: webfetch, websearch (follow same pattern as grep/glob/list)

---

# Template Path Resolution

Templates located at: `config/templates/{platform}/{docType}.tmpl`

Core package resolves from CWD or parent directory (for test flexibility).

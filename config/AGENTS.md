**Location**: `config/`
**Parent**: See `/AGENTS.md` for project overview

---

# Configuration Templates

Platform-specific Go templates for document transformation.

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

# Template Function Maps

Templates use Sprig functions + custom `transformPermissionMode`.

Accessed via:
```go
template.New(name).Funcs(createTemplateFuncMap()).Parse(content)
```

Sprig provides: `lower`, `upper`, `trim`, `join`, etc.

---

# Claude Code Templates

## agent.tmpl

YAML frontmatter with `---` delimiters.
Fields: name, description, tools, disallowedTools, model, permissionMode, skills.
Content rendered after second `---`.

## command.tmpl

Fields: allowed-tools, argument-hint, context, agent, description, model, disable-model-invocation.
Lists use YAML syntax: `- item`

## skill.tmpl

Fields: name, description, allowed-tools, model, context, agent, user-invocable.

## memory.tmpl

Fields: paths (list), content.

---

# OpenCode Templates

## agent.tmpl

YAML frontmatter with field omission patterns:

Boolean omission (omit when false):
```go
{{- if .Hidden}}
hidden: true
{{- end}}
```

Pointer omission (omit when nil):
```go
{{- if .Temperature}}
temperature: {{.Temperature}}
{{- end}}
```

String omission (omit when empty):
```go
{{- if ne .Prompt ""}}
prompt: {{.Prompt}}
{{- end}}
```

Tools transformed to lowercase:
```go
{{- range .Tools}}
  {{. | lower}}: true
{{- end}}
```

DisallowedTools transformed to lowercase with false:
```go
{{- range .DisallowedTools}}
  {{. | lower}}: false
{{- end}}
```

## command.tmpl

Fields: description, agent, subtask, model.
Note: `allowed-tools`, `argument-hint`, `disable-model-invocation` are skipped.

## skill.tmpl

Fields: name, description, license, compatibility (list), metadata (map), hooks (map).

Map iteration for metadata/hooks:
```go
{{- range $key, $value := .Metadata}}
  {{$key}}: "{{$value}}"
{{- end}}
```

## memory.tmpl

Paths rendered as `@file/path` references:
```go
{{if .Paths}}
To reference files, use @ followed by file path.
{{end}}{{if .Paths}}{{range .Paths}}
@{{.}}
{{end}}{{end}}{{if .Content}}{{.Content}}{{end}}
```

---

# Custom Template Function

## transformPermissionMode

Converts Claude Code enum to OpenCode permission object.

**Claude Code modes**:
- `default` → ask for all tools
- `acceptEdits` → allow edit/read/grep/glob/list, ask bash/webfetch/websearch
- `dontAsk` → allow all
- `bypassPermissions` → allow all
- `plan` → deny edit/bash, allow others

**OpenCode output** (YAML with indentation):
```yaml
  edit:
    "*": ask
  bash:
    "*": ask
  ...
```

Used in opencode/agent.tmpl:
```go
{{- if .PermissionMode}}
permission:
{{transformPermissionMode .PermissionMode}}
{{- end}}
```

Returns empty string for unknown modes.

---

# Template Path Resolution

Templates located at: `config/templates/{platform}/{docType}.tmpl`

Core package resolves from CWD or parent directory (for test flexibility).

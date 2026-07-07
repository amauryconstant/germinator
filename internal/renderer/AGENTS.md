**Location**: `internal/renderer/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview

---

# Renderer Package

Platform-agnostic document renderer. Executes Go `text/template` templates against `Canonical*` structs to produce platform-specific (claude-code, opencode) or canonical YAML output. Owns template lookup, Sprig-based func maps, and the per-platform permission-policy / tool-name template helpers.

## Files

| File | Purpose |
|------|---------|
| `serializer.go` | `RenderDocument`, `MarshalCanonical`, `Serializer` struct, template-path resolution, `createTemplateFuncMap`, `createCanonicalTemplateFuncMap` |
| `template_funcs.go` | Reserved for template func registration (currently empty placeholder) |
| `serializer_test.go` | Unit tests |

## Public Surface

- `RenderDocument(doc interface{}, platform string) (string, error)` — renders a `parser.Canonical*` value to a platform-specific document string. Selects template at `config/templates/<platform>/<docType>.tmpl`.
- `MarshalCanonical(doc interface{}) (string, error)` — renders a `parser.Canonical*` value back to canonical Germinator YAML using `config/templates/canonical/<docType>.tmpl`.
- `Serializer` / `NewSerializer()` — method-style wrapper around `RenderDocument` for callers that prefer instance syntax.

### Template function map (custom, beyond Sprig)

Available in platform templates via `createTemplateFuncMap`:

- `permissionPolicyToClaudeCode policy` — `PermissionPolicy` → `"default" | "acceptEdits" | "dontAsk" | "plan" | "bypassPermissions"`
- `permissionPolicyToOpenCode policy` — `PermissionPolicy` → multi-line YAML fragment for the OpenCode `permission:` map (edit/bash/read/grep/glob/list/webfetch/websearch)
- `convertToolNameCase name platform` — tool name canonicalization (PascalCase for claude-code, lowercase for opencode)

Canonical templates use `createCanonicalTemplateFuncMap` (Sprig only — no platform adapters).

## Design Notes

**Template discovery.** `getTemplatePath` / `getCanonicalTemplatePath` look up templates by walking up from `os.Getwd()` until a `go.mod` is found (`findProjectRoot`), then resolving `config/templates/<scope>/<filename>`. This dual-resolution (cwd first, then project root) supports both production invocations from the project root and test invocations from package directories.

**Document-type detection.** `getDocType` is a type switch over `*parser.CanonicalAgent|Command|Skill|Memory`; unknown types return `*core.TransformError`. The renderer trusts the parser's typing and does no content sniffing.

**Dependencies.** Imports `internal/core` (errors), `internal/parser` (Canonical types for `getDocType`), `internal/permission` (for `permission.Map` shape in the OpenCode YAML emitter), and the platform adapters `internal/claude-code` + `internal/opencode` (instantiated inside the func map closures). Uses `Masterminds/sprig/v3` for generic template helpers and `text/template` from stdlib.

**Error wrapping.** Every failure path returns a typed `core.*Error` (`TransformError` for parse/execute failures, `FileError` for missing template files) so `output.FormatError` can dispatch on type in the cmd layer.

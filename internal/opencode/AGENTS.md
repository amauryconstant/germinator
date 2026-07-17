**Location**: `internal/opencode/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# OpenCode Adapter Package

Bidirectional converter between canonical Germinator models and OpenCode's YAML format. Implements the same `ToCanonical` / `FromCanonical` shape as `internal/claude-code`; handles OpenCode-specific differences (boolean tool maps, flattened behavior, nested permission object, kebab-case keys).

## Files

| File | Purpose |
|------|---------|
| `doc.go` | Package doc: permission mapping table, tool-list splitting, behavior flattening, OpenCode-specific fields |
| `opencode_adapter.go` | `Adapter` struct; `New()`; `ToCanonical`, `FromCanonical`, `PermissionPolicyToPlatform`, `ConvertToolNameCase` |
| `opencode_adapter_test.go` | Unit tests |

## Public Surface

- `Adapter` — stateless struct; same shape as `claudecode.Adapter`.
- `New() *Adapter` — constructor.
- `ToCanonical(input map[string]interface{}) (*core.Agent, *core.Command, *core.Skill, *core.Memory, error)` — reads `input["__type"]`, dispatches to the matching `parse*` method. Exactly one returned pointer is non-nil on success.
- `FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)` — inverse: produces the OpenCode-shaped map for template rendering.
- `PermissionPolicyToPlatform(policy core.PermissionPolicy) (interface{}, error)` — returns a `permission.Map` (per-tool `Allow`/`Ask`/`Deny` actions for edit/bash/read/grep/glob/list/webfetch/websearch) via `permission.PermissionPolicyMappings`.
- `ConvertToolNameCase(name string) string` — identity (`permission.ToLowerCase`); OpenCode uses lowercase tool names, matching canonical.

## Design Notes

**OpenCode-specific transformations** (documented in `doc.go`):

- **Tool list as boolean map.** Canonical uses separate `tools` / `disallowedTools` arrays; OpenCode uses a single `{tool: true|false}` map. `parseAgent`/`renderAgent` translate both directions.
- **Behavior flattening.** OpenCode does not support nested `behavior:` objects — the adapter flattens canonical `behavior.*` to top-level keys (`mode`, `temperature`, `maxSteps`, `prompt`, `hidden`, `disable`). Note the renames: canonical `behavior.disabled` → OpenCode `disable`; canonical `behavior.steps` → OpenCode `maxSteps`.
- **Permission dual shape.** OpenCode accepts either a `permissionMode` string (mapped via `mapPermissionModeToPolicy`) or a nested `permission:` object (decoded via `mapPermissionObjectToPolicy`, which inspects edit/bash deny flags to infer the canonical policy). The nested-object path validates action strings via `permission.ValidateActionStrings` before mapping; unknown action values surface as `*core.ConfigError` per the spec scenario "Unknown action string at runtime is rejected" in `openspec/changes/harden-tests-and-coverage/specs/transformation-permission-transformation/spec.md`.
- **Kebab-case keys.** `allowed-tools`, `argument-hint`, `user-invocable` are kebab-cased on the OpenCode side and converted to / from the camelCased canonical fields.

**Dependencies.** `internal/core` (types, errors, `PermissionPolicy`) and `internal/permission` (mappings, `permission.Map`, `ToLowerCase`). No I/O. Returns typed `core.*Error`.

**Adapter contract.** The five-method surface mirrors `internal/claude-code` exactly. `internal/parser.ParsePlatformDocument` selects between the two adapters via the `platform` argument and calls `ToCanonical` through an anonymous interface assertion.

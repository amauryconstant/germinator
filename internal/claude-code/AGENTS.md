**Location**: `internal/claude-code/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview

---

# Claude Code Adapter Package

Bidirectional converter between canonical Germinator models and Claude Code's YAML format. Implements the `ToCanonical` / `FromCanonical` shape shared with `internal/opencode`; consumed by `internal/parser` (parse side) and `internal/renderer` (template-func side).

## Files

| File | Purpose |
|------|---------|
| `doc.go` | Package doc: permission-policy mapping table, tool-name convention, Claude Code-specific fields |
| `claude_code_adapter.go` | `Adapter` struct; `New()`; `ToCanonical`, `FromCanonical`, `PermissionPolicyToPlatform`, `ConvertToolNameCase` |
| `claude_code_adapter_test.go` | Unit tests |

## Public Surface

- `Adapter` — stateless struct; all methods are essentially pure functions.
- `New() *Adapter` — constructor (allocates nothing meaningful; exists for symmetric API with the OpenCode adapter).
- `ToCanonical(input map[string]interface{}) (*core.Agent, *core.Command, *core.Skill, *core.Memory, error)` — reads `input["__type"]` (`"agent"|"command"|"skill"|"memory"`) and dispatches to the matching `parse*` method. Exactly one of the four pointers is non-nil on success.
- `FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)` — inverse: takes a `*core.*` value and produces a `map[string]interface{}` suitable for template rendering. Asserts the concrete type with a type assertion and returns `*core.TransformError` on mismatch.
- `PermissionPolicyToPlatform(policy core.PermissionPolicy) (interface{}, error)` — returns the Claude Code `permissionMode` string (`"default"`, `"acceptEdits"`, `"dontAsk"`, `"plan"`, `"bypassPermissions"`) via `permission.PermissionPolicyMappings`.
- `ConvertToolNameCase(name string) string` — lowercase → PascalCase (`"bash"` → `"Bash"`) via `permission.ToPascalCase`.

## Design Notes

**Adapter contract.** The five-method surface (`ToCanonical`/`FromCanonical`/`PermissionPolicyToPlatform`/`ConvertToolNameCase` plus the constructor) is shared verbatim with `internal/opencode`. `internal/parser.ParsePlatformDocument` calls `ToCanonical` through an inline anonymous interface assertion rather than importing a shared interface type — both adapters must keep the exact signature in sync.

**Permission mapping.** `parseAgent` reads Claude Code's `permissionMode` string and maps it back to canonical via `mapPermissionModeToPolicy` (the inverse of `PermissionPolicyToPlatform`). The mapping table lives in `internal/permission`.

**Claude Code-specific fields.** Skills list (`skills:`), `disableModelInvocation`, and `permissionMode` are read from / written to the `targets["claude-code"]` sub-map on canonical models, so cross-platform fidelity is preserved.

**Dependencies.** `internal/core` (types, errors, `PermissionPolicy`) and `internal/permission` (policy mappings, `ToPascalCase`). No I/O — purely in-memory conversion. Returns typed `core.*Error` so the cmd layer can format via `output.FormatError`.

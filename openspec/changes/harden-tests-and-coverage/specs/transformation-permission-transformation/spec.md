# transformation-permission-transformation Specification (delta)

## ADDED Requirements

### Requirement: Typed permission constants

The `internal/permission/` package SHALL expose typed `Action` constants (`permission.Allow`, `permission.Ask`, `permission.Deny`) that adapter permission maps use instead of raw string literals. The constants are exported and used by `internal/opencode/`, `internal/claude-code/`, and any other adapter that maps canonical permissions to platform-specific actions.

**Change**: NEW requirement. The typed constants already exist at `internal/permission/permissions.go:38-47`; the work in this change is to ensure adapter permission maps consume the constants instead of raw string literals. Pre-change, the adapter maps embed raw string literals (`"allow"`, `"ask"`, `"deny"`) directly, so a typo in a string literal is undetected at compile time. The typed constants centralize the vocabulary; future unrecognised values are caught at runtime via `*core.ConfigError` from `PermissionPolicyMappings` lookup.

#### Scenario: Action constants are typed

- **WHEN** the `internal/permission/permissions.go` file is inspected
- **THEN** it SHALL define `Action` as a typed string: `type Action string`
- **AND** it SHALL define the constants: `const (Allow Action = "allow"; Ask Action = "ask"; Deny Action = "deny")`

#### Scenario: Adapters use typed constants

- **WHEN** the `internal/opencode/opencode_adapter.go` and `internal/claude-code/claude_code_adapter.go` permission maps are inspected
- **THEN** the maps SHALL use `permission.Allow` / `permission.Ask` / `permission.Deny` (not the string literals `"allow"` / `"ask"` / `"deny"`)
- **AND** the maps SHALL be typed as `map[string]permission.Action` (not `map[string]string`)

#### Scenario: No raw string literals in adapter permission maps

- **WHEN** the codebase is searched for `"allow"` / `"ask"` / `"deny"` in `internal/opencode/` and `internal/claude-code/`
- **THEN** zero matches SHALL appear in the adapter permission maps (the constants are used instead)
- **AND** raw string literals SHALL appear only in test fixtures, golden files, and the `permission.Action` constant declarations themselves

#### Scenario: Unknown action string at runtime is rejected

- **WHEN** an adapter receives an action string that is not one of the typed `permission.Action` constants (e.g., a future `"denyUnlessRead"` value)
- **THEN** the adapter SHALL return `*core.ConfigError` from the lookup
- **AND** the error message SHALL list the valid `permission.Action` values (`Allow`, `Ask`, `Deny`)

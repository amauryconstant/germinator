# transformation-permission-transformation Specification (delta)

## MODIFIED Requirements

### Requirement: Typed permission constants

The `internal/permission/` package SHALL define typed `Action` constants (`permission.Allow`, `permission.Ask`, `permission.Deny`) that adapter permission maps use instead of raw string literals. The constants are exported and used by `internal/opencode/`, `internal/claude-code/`, and any other adapter that maps canonical permissions to platform-specific actions.

**Change**: NEW requirement. Pre-change, the adapter maps (e.g., `permission.Map` at `internal/permission/permissions.go:41`) compare raw string literals (`"allow"`, `"ask"`, `"deny"`). A typo in a string literal is undetected at compile time; the typed constants catch typos at compile time.

#### Scenario: Action constants are typed

- **WHEN** the `internal/permission/permissions.go` file is inspected
- **THEN** it SHALL define `Action` as a typed string: `type Action string`
- **AND** it SHALL define the constants: `const (Allow Action = "allow"; Ask Action = "ask"; Deny Action = "deny")`

#### Scenario: Adapters use typed constants

- **WHEN** the `internal/opencode/opencode_adapter.go` and `internal/claude-code/claude_code_adapter.go` permission maps are inspected
- **THEN** the maps SHALL use `permission.Allow` / `permission.Ask` / `permission.Deny` (not the string literals `"allow"` / `"ask"` / `"deny"`)
- **AND** the maps SHALL be typed as `map[string]permission.Action` (not `map[string]string`)

#### Scenario: Typos in Action are caught at compile time

- **WHEN** a contributor writes `permission.Action("alllow")` (typo) in an adapter
- **THEN** the build SHALL fail with `cannot use "alllow" (untyped string constant) as permission.Action value`
- **AND** the contributor SHALL fix the typo (or add a new `Action` constant if the typo is intentional)

#### Scenario: No raw string literals in adapter permission maps

- **WHEN** the codebase is searched for `"allow"` / `"ask"` / `"deny"` in `internal/opencode/` and `internal/claude-code/`
- **THEN** zero matches SHALL appear in the adapter permission maps (the constants are used instead)
- **AND** raw string literals SHALL appear only in test fixtures, golden files, and the `permission.Action` constant declarations themselves

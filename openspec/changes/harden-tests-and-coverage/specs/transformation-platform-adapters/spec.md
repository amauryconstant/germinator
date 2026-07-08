# transformation-platform-adapters Specification (delta)

## MODIFIED Requirements

### Requirement: Adapter contract guard

Both platform adapter packages (`internal/opencode/`, `internal/claude-code/`) SHALL declare a compile-time check that the adapter satisfies the `permission.Adapter` interface:

```go
var _ permission.Adapter = (*Adapter)(nil)
```

The check is placed at the bottom of the adapter's main file (e.g., `opencode_adapter.go`) and is the canonical Go idiom for interface satisfaction verification.

**Change**: NEW requirement. Pre-change, the `permission.Adapter` interface is declared in `internal/permission/adapter.go:8-13` but no compile-time check exists. The pre-change signature drift in either adapter is detected only at the call site, not at compile time.

#### Scenario: OpenCode adapter satisfies permission.Adapter

- **WHEN** `internal/opencode/opencode_adapter.go` is inspected
- **THEN** the file SHALL declare `var _ permission.Adapter = (*Adapter)(nil)`
- **AND** the `*Adapter` type SHALL expose all 5 methods of `permission.Adapter` (`ToCanonical`, `FromCanonical`, `ParsePlatformDocument`, `RenderDocument`, plus the permission-mapping helpers)

#### Scenario: Claude Code adapter satisfies permission.Adapter

- **WHEN** `internal/claude-code/claude_code_adapter.go` is inspected
- **THEN** the file SHALL declare `var _ permission.Adapter = (*Adapter)(nil)`
- **AND** the `*Adapter` type SHALL expose all 5 methods of `permission.Adapter`

#### Scenario: Signature drift is caught at compile time

- **WHEN** a contributor modifies the `permission.Adapter` interface (e.g., renames a method)
- **THEN** the build SHALL fail with `cannot use (*Adapter)(nil) as permission.Adapter` in both adapter files
- **AND** the contributor SHALL update both adapters to match the new interface

### Requirement: Adapter singleton

The `internal/opencode/` and `internal/claude-code/` packages SHALL expose a package-level singleton (`var OpenCode = &Adapter{}` and `var ClaudeCode = &Adapter{}`) instead of a per-call `New()` constructor. The `Adapter` struct is stateless, so a single instance is safe to share.

**Change**: NEW requirement. Pre-change, `New()` returns `&Adapter{}` per call, allocating a new instance each invocation. The singleton eliminates the allocation and makes the stateless nature of the adapter explicit.

#### Scenario: Package-level singleton exposed

- **WHEN** the `internal/opencode/` package is inspected
- **THEN** it SHALL expose `var OpenCode = &Adapter{}` at the package level
- **AND** the package-level `New()` function SHALL NOT exist (deleted)
- **WHEN** the `internal/claude-code/` package is inspected
- **THEN** it SHALL expose `var ClaudeCode = &Adapter{}` at the package level
- **AND** the package-level `New()` function SHALL NOT exist (deleted)

#### Scenario: Singleton is stateless

- **WHEN** the singleton is used concurrently from multiple goroutines
- **THEN** the singleton SHALL NOT exhibit data races (verified by `go test -race`)
- **AND** the singleton SHALL NOT mutate any internal state during `ToCanonical` / `FromCanonical` calls

#### Scenario: Callers use the singleton

- **WHEN** the codebase is searched for `claudecode.New()` or `opencode.New()`
- **THEN** zero matches SHALL appear (the constructor is deleted; callers use the singleton)
- **AND** the callers SHALL import the package and use `claudecode.ClaudeCode` / `opencode.OpenCode` directly

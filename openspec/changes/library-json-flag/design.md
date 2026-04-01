## Context

The `germinator library` command group currently has inconsistent JSON output support:
- `library refresh --json` - exists in `library_refresh.go`
- `library remove --json` - exists in `library_remove.go`
- `library validate --json` - exists in `library_validate.go`
- `library resources --json` - not implemented
- `library presets --json` - not implemented
- `library show --json` - not implemented
- `library init --json` - not implemented
- `library add --json` - not implemented (also missing `--batch` flag)

Currently, each subcommand defines its own `--json` flag locally. This creates duplication and inconsistency. The user wants a single `--json` flag on the parent command that inherits to all subcommands via Cobra's flag inheritance.

## Goals / Non-Goals

**Goals:**
- Add `--json` flag to the parent `library` command in `cmd/library.go`
- Implement JSON output for all subcommands that currently lack it
- Ensure consistent JSON structure across all library subcommands
- Use `json.NewEncoder(c.OutOrStdout()).SetIndent("", "  ")` for JSON formatting

**Non-Goals:**
- Modifying the underlying library data types (existing JSON-capable types are fine)
- Changing behavior of existing JSON output (backward compatibility)
- Adding `--batch` flag to `library add` (separate change)

## Decisions

### Decision 1: Flag Inheritance Approach

**Choice:** Add `--json` as a persistent flag on the parent `library` command using `cmd.PersistentFlags().BoolVar(&opts.json, "json", false, ...)`.

**Rationale:** Cobra's persistent flags are inherited by child commands automatically. This is the cleanest approach - single flag definition on parent, automatic inheritance.

**Alternatives Considered:**
- Add flag to each subcommand individually: Would cause duplication and inconsistency
- Use command hierarchy traversal to find parent flag: Complex and fragile

### Decision 2: JSON Output Check Pattern

**Choice:** Each subcommand's `RunE` function checks `c.Flags().Changed("json")` or retrieves the flag value via `c.Flags().GetBool("json")`.

**Rationale:** This follows the existing pattern in `library_refresh.go` where the flag is checked within the `RunE` function.

**Alternatives Considered:**
- Pre-execute hook to intercept: Overly complex
- Global command configuration: Breaks the existing command pattern

### Decision 3: JSON Structure for List Commands

**Choice:**
- `library resources --json` outputs: `{"resources": [...]}`
- `library presets --json` outputs: `{"presets": [...]}`

**Rationale:** Consistent with the user's requirements. Wrapping arrays in objects provides future extensibility for adding metadata (e.g., count, pagination).

### Decision 4: Existing JSON Implementations

**Choice:** Keep existing JSON output in `refresh`, `remove`, and `validate` as-is. The parent flag will also be available, allowing both patterns to work.

**Rationale:** Backward compatibility. Existing scripts using `--json` continue to work.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| Inherited flag with same name as local flag | Conflict if a subcommand already has a local `--json` flag | Remove local `--json` flags from refresh, remove, validate (they'll inherit from parent) |
| JSON output format differs between subcommands | Inconsistent API for consumers | Standardize structures: all use `{"<key>": [...]} or `{"<key>": {...}}` |
| Breaking existing `--json` behavior | Scripts relying on current output format | Keep existing implementations; parent flag adds to capability |

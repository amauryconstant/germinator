## Context

Germinator transforms AI coding assistant documents between platforms. Users currently manage skills, agents, and commands as individual files with no central indexing or discovery mechanism. Additionally, users starting new projects must manually create these configurations, which is repetitive and error-prone.

This design introduces:
1. **A Library System** that stores canonical resources in a structured directory with an index file (`library.yaml`)
2. **An Init Command** that installs pre-defined resources from the library to target projects

The key insight: **init is a batch transform** - it reuses the existing `LoadDocument → RenderDocument` pipeline for each resource, adding only library management and output path derivation.

**Constraints:**
- No external dependencies beyond existing stack (Go 1.25, Cobra, gopkg.in/yaml.v3)
- No remote libraries (local filesystem only)
- No resource modification through CLI (read-only operations for library commands)
- No default library shipped (user must create their own)
- Reuse existing transformation pipeline (no new adapters/templates)
- Library uses canonical YAML format (same as adapt command input)

**Assumptions:**
- Users manage their library manually (add/remove files, edit library.yaml)
- Resource files are valid canonical YAML (skills, agents, commands)
- Single library per user (no multi-library support)
- Users opt into library approach; explicit is clearer

## Goals / Non-Goals

**Goals:**
- Define and parse `library.yaml` index format
- Load library from filesystem path
- Resolve resource names to file paths via index lookup
- Resolve preset names to lists of resource names
- Discover library path via flag > env > default priority
- Provide CLI commands: `library resources`, `library presets`, `library show`
- Install multiple resources from a library in one command
- Support explicit resource selection (`--resources commit,merge-request`)
- Support presets for bundled resources (`--preset git-workflow`)
- Derive platform-specific output paths from resource name and type

**Non-Goals:**
- Library modification commands (`add`, `remove`, `init` for library itself)
- Remote library support (git, http, etc.)
- Resource validation at library level (delegated to existing validation)
- Preset installation or composition beyond listing
- Integration with adapt/validate commands
- Interactive TUI for resource selection
- Library synchronization tooling
- Shipping a default/sample library

## Decisions

### D1: Package Location

**Choice:** `internal/library/`

**Rationale:** Follows existing architecture pattern. Library is internal implementation detail, not exported. Mirrors structure of `internal/core/` and `internal/services/`.

**Alternatives:**
- `pkg/library/` - rejected (not meant for external consumption)
- `internal/core/library.go` - rejected (separate domain, deserves own package)

### D2: File Structure

**Choice:** Five files with single responsibility:
```
internal/library/
├── library.go     # Types (Library, Resource, Preset)
├── loader.go      # LoadLibrary(path)
├── resolver.go    # ResolveResource(), ResolvePreset(), GetOutputPath()
├── lister.go      # ListResources(), ListPresets()
└── discovery.go   # FindLibrary()
```

**Rationale:** Matches existing patterns (`core/` has loader.go, parser.go, serializer.go). Easier to navigate and test. Each file has clear purpose.

**Alternatives:**
- Single `library.go` - rejected (mixes concerns, harder to test)
- Split by feature - rejected (over-engineering for 5 functions)

### D3: Resolver Return Value

**Choice:** Return absolute file path only (`string, error`) for ResolveResource

**Rationale:** Simplest contract. Caller decides what to do with path (load, validate, display). Keeps resolver focused on lookup, not I/O.

**Alternatives:**
- Return parsed resource (`*canonical.Skill, error`) - rejected (resolver shouldn't know about models)
- Return `io.Reader` - rejected (caller loses path information)

### D4: Error Style

**Choice:** Terse errors, no suggestions

**Rationale:** v1 simplicity. User preference from exploration session.

**Examples:**
- `resource not found: foo`
- `preset not found: bar`
- `library not found: /path/to/library`

**Alternatives:**
- Did-you-mean suggestions - rejected (complexity for v1)
- Helpful hints ("run `list` to see available") - rejected (preference for terse)

### D5: Library Path Discovery

**Choice:** Priority chain: `--library` flag > `GERMINATOR_LIBRARY` env > `~/.config/germinator/library/`

**Rationale:** Standard Go/Cobra pattern. Flag overrides everything, env for config, sensible default following XDG convention.

**Alternatives:**
- Config file support - rejected (out of scope, separate system per spec)
- XDG config home with full spec - rejected (over-engineering, `~/.config/` is conventional)

### D6: Resource Index Format

**Choice:** Nested by type: `resources: {type: {name: {path, description}}}`

**Rationale:** Namespaces resources by type, allowing identical names across skill/agent/command. Natural grouping for CLI output. Directory structure (skills/, agents/, commands/) maps directly to type.

**Example:**
```yaml
resources:
  skill:
    commit:
      path: skills/commit.yaml
      description: Git commit best practices
  agent:
    commit:
      path: agents/commit.yaml
      description: Commit automation agent
```

**CLI reference format:** `type/name` (e.g., `skill/commit`, `agent/commit`)

**Alternatives:**
- Flat with composite key (`skill/commit: {type, path, description}`) - rejected (redundant type in key and value)
- Flat with user-provided namespacing - rejected (index generates names, not user; collision-prone)

### D7: Output Path Derivation

**Choice:** Derive from resource name and type using platform conventions

| Type | OpenCode | Claude Code |
|------|----------|-------------|
| skill | `.opencode/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` |
| agent | `.opencode/agents/<name>.md` | `.claude/agents/<name>.md` |
| command | `.opencode/commands/<name>.md` | `.claude/commands/<name>.md` |
| memory | `.opencode/memory/<name>.md` | `.claude/memory/<name>.md` |

**Rationale:**
- Matches platform conventions (OpenCode expects `SKILL.md` in subdirectory)
- Predictable output locations
- No additional configuration needed per resource

**Alternatives considered:**
- Explicit output path in library.yaml: More flexible but more maintenance
- Flat structure: Doesn't match platform conventions

### D8: Architecture

**Choice:** Three-layer architecture with service orchestration

```
cmd/library.go ──────┐
cmd/init.go ─────────┼──▶ services/initializer.go ──▶ core/loader.go
                     │                                 core/serializer.go
                     ▼
               library/loader.go
               library/resolver.go
               library/lister.go
```

**Components:**
- `internal/library/` - Library loading, parsing, resource resolution, listing
- `internal/services/initializer.go` - Orchestrates init workflow
- `cmd/library.go` - Library CLI commands (resources, presets, show)
- `cmd/init.go` - Init CLI command

**Rationale:**
- Separates concerns (library vs transformation vs CLI)
- Reuses existing `core` package without modification
- Testable in isolation

### D9: Error Handling Strategy

**Choice:** Fail-fast - stop on first error

**Rationale:** Partial installation is confusing and leaves project in inconsistent state. User can fix the issue and retry the entire operation.

**Alternatives:**
- Continue-on-error with summary - rejected (harder to reason about partial state)
- Rollback on error - rejected (complexity, may not always be possible)

### D10: List Validation Strategy

**Choice:** Validate resources during list operations

**Rationale:** Catches issues early before installation attempt. Better user experience to discover invalid resources when browsing.

**Trade-off:** Slower list operations but better error detection.

**Alternatives:**
- Validate only during init - rejected (delays error discovery)
- Skip validation entirely - rejected (allows invalid library state)

### D11: Preset Extensibility

**Choice:** No preset extension in v1 - `--preset foo --resources bar` is mutually exclusive

**Rationale:** Clearer mental model. User either installs a preset or selects specific resources. Mixing is ambiguous.

**Future:** Consider `--resources +bar` syntax for additive mode if needed.

**Alternatives:**
- Allow `--preset foo --resources bar` to add to preset - rejected (ambiguous, v1 simplicity)

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| library.yaml format evolves | Version field (`version: "1"`) allows future schema changes |
| Resource file missing from disk | Resolver returns error with expected path |
| Invalid library.yaml | Loader returns parse error with details |
| Empty library directory | Commands succeed with empty output (valid state) |
| Resource name collisions | Index is flat map by type, names are unique within type |
| Library Index Maintenance Burden | Document the format clearly; note future tooling need |
| Output Path Conflicts | Default to error on existing files; `--force` flag to overwrite |
| Invalid Library Resources | Validate each resource during load; report errors clearly with file paths |

## Open Questions

(none - all questions resolved as D9, D10, D11)

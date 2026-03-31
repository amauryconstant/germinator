## Context

The library system manages canonical resources (skills, agents, commands, memory) that can be installed into user projects via `germinator init`. Currently, adding new resources to the library requires:
1. Manually copying the canonical file to the appropriate subdirectory (`skills/`, `agents/`, `commands/`, `memory/`)
2. Manually editing `library.yaml` to register the resource with its path and description

This manual workflow is error-prone and not discoverable through the CLI.

## Goals / Non-Goals

**Goals:**
- Enable importing canonical or platform documents into the library via `germinator library add <source>`
- Support auto-detection of type, name, and description from file content or flags
- Canonicalize platform documents (OpenCode, Claude Code) to canonical format on import
- Validate canonical documents before adding to the library
- Maintain library integrity by validating after updates

**Non-Goals:**
- Editing or modifying existing resources (that can be done by re-adding with --force)
- Creating new resources from scratch (scaffolding is separate concern)
- Importing resources that are already in canonical format without validation

## Decisions

### Decision 1: AddResource function in library infrastructure

**Choice:** Create `AddResource(opts AddOptions) error` in `internal/infrastructure/library/adder.go`

**Rationale:** The library package already has `LoadLibrary`, `CreateLibrary`, `ResolveResource`, etc. Adding a new `AddResource` function keeps all library-modifying operations in the same package. The `AddOptions` struct encapsulates all inputs.

**Alternatives considered:**
- Put logic in command layer (cmd/library_add.go): Rejected - library operations belong in infrastructure
- Create new package `library/add.go`: Rejected - over-engineering for single function
- Add to existing loader.go: Rejected - loader is already complex with LoadLibrary

### Decision 2: Auto-detection priority for type

**Choice:** `--type` flag > frontmatter `type:` field > filename pattern

**Rationale:** Explicit flag gives user control when needed. Frontmatter is natural since documents already contain metadata. Filename pattern is fallback for unlabeled files.

**Alternatives considered:**
- Filename pattern first: Rejected - less reliable than explicit flags or metadata
- Skip type detection entirely: Rejected - adds friction to common case

### Decision 3: Canonicalization on import

**Choice:** If source is platform format, canonicalize before storing; if already canonical, validate only

**Rationale:** Users may have platform documents they want to contribute. Canonicalizing ensures library stores canonical format (single source of truth). Platform detection via frontmatter fields (`name:`, `description:`, `allowed-tools:` indicate canonical).

**Alternatives considered:**
- Require canonical format only: Rejected - adds friction for users with platform docs
- Always try to canonicalize: Rejected - could produce invalid canonical from ambiguous input

### Decision 4: Target path derivation

**Choice:** `{library}/{type}s/{name}.md` (e.g., `library/agents/reviewer.md`)

**Rationale:** Mirrors existing canonical format in `test/fixtures/library/skills/skill-commit.md` which uses `skills/skill-commit.md` path. Simple, predictable, matches library conventions.

**Alternatives considered:**
- `{type}-{name}.md` (e.g., `agent-reviewer.md`): Rejected - doesn't match existing canonical files

### Decision 5: library.yaml modification

**Choice:** Re-serialize full library struct after adding resource

**Rationale:** Simple implementation. Current library.yaml is not hand-edited, so losing formatting is acceptable. Validation after load ensures consistency.

**Alternatives considered:**
- Manual YAML editing: Rejected - complex, error-prone
- Preserve formatting: Rejected - adds significant complexity for no practical benefit

## Risks / Trade-offs

- [Risk] Name collision: If user adds resource with name that exists
  - **Mitigation:** Error with clear message. Use `--force` to replace existing.
- [Risk] Invalid canonical document: User imports malformed resource
  - **Mitigation:** Validate using existing `ValidateAgent`, `ValidateCommand`, etc. before adding.
- [Risk] Library corruption: Failed update leaves library.yaml invalid
  - **Mitigation:** Write to temp file first, validate, then rename atomically.
- [Risk] Source file disappears between validation and copy
  - **Mitigation:** Copy before validation (fail after add is acceptable - user can fix).

## Open Questions

- Should we support `--description` flag to override extracted description? (Yes, flag > extracted)
- What exit code for validation failure? (5 - Validation, per CLI conventions)

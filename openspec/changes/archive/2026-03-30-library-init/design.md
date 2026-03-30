## Context

The Germinator CLI manages a library of canonical resources (skills, agents, commands, memory) organized in a directory with a `library.yaml` manifest. Users can:
- List resources via `germinator library resources`
- Show resource details via `germinator library show <ref>`
- Install resources to projects via `germinator init`

However, there is no way to scaffold a new library directory. Users must manually create the structure and `library.yaml`, which is error-prone and not discoverable through the CLI.

This design addresses the creation of library directory structures, not the installation of resources (handled by `germinator init`).

## Goals / Non-Goals

**Goals:**
- Allow users to create a new library directory at a specified path via `germinator library init`
- Create a valid, loadable library structure that passes `LoadLibrary` validation
- Provide flags for path specification, dry-run preview, and force overwrite
- Create empty resource directories (`skills/`, `agents/`, `commands/`, `memory/`)
- Follow existing CLI patterns and conventions

**Non-Goals:**
- Interactive scaffolding (questions, prompts)
- Template-based creation with example resources
- Moving or migrating existing libraries
- Modifying existing library loading or resource installation behavior

## Decisions

### Decision 1: Default path behavior

**Choice:** Default to `~/.config/germinator/library/` when `--path` is not specified.

**Rationale:** This aligns with `FindLibrary`'s default path. A user running `germinator library init` without flags expects to create a library at the standard location.

**Alternatives considered:**
- Require `--path` explicitly (safer but less convenient)
- Default to current directory (`.`) (unconventional, conflicts with typical library expectations)

### Decision 2: Existence check and force behavior

**Choice:** Error if library already exists at target path unless `--force` is specified.

**Rationale:** Prevents accidental overwrites. The `--force` flag is explicit acknowledgment of destructive behavior.

**Alternatives considered:**
- Skip if exists (idempotent) (could hide mistakes)
- Always overwrite (too dangerous)

### Decision 3: Post-creation validation

**Choice:** Validate created library by calling `LoadLibrary` after creation.

**Rationale:** Ensures the created library is well-formed and loadable. If validation fails, the partial structure is left in place for debugging.

**Alternatives considered:**
- Trust creation (faster but could produce invalid libraries)
- Cleanup on failure (hides debugging information)

### Decision 4: Create empty resource directories

**Choice:** Create empty `skills/`, `agents/`, `commands/`, `memory/` directories.

**Rationale:** Follows the observed fixture structure and makes the library ready for resources without additional scaffolding.

**Alternatives considered:**
- Only create `library.yaml` (minimal, but directories would need to be created later)
- Create with `.gitkeep` files (keeps empty dirs in git, more complex)

### Decision 5: No new service interface

**Choice:** Implement creation logic as a function in `internal/infrastructure/library/creator.go` rather than adding a new application service interface.

**Rationale:** This is a simple file-creation operation, not a multi-step pipeline requiring the service pattern. The existing `Initializer` service handles resource installation; this is orthogonal.

**Alternatives considered:**
- Add `application.LibraryCreator` interface (overkill for simple file operations)
- Put logic in CLI layer only (less reusable)

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| User runs `library init` without `--path` and overwrites existing default library | Require `--force` if library exists at target path |
| Race condition: library created between check and creation | Document as timing limitation; check-then-create is inherently racy |
| Created library structure becomes out of sync with `LoadLibrary` expectations | Post-creation validation via `LoadLibrary` catches drift |
| Users expect `library init` to work like `git init` (initialize in current directory) | Document default path behavior clearly in help text |
| Validation failure leaves partial structure that confuses users | Leave partial structure for debugging; document in error message |

## Resolved Questions

1. **Cleanup on validation failure**: Partial structure is left in place for debugging. Error message indicates what was created and that validation failed. User can manually remove if needed.

2. **Output verbosity**: `--verbose` (single) shows basic progress. `--dry-run` shows preview of what would be created. No additional verbosity needed.

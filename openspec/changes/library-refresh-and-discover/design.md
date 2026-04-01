## Context

The library system stores resource metadata (path, description) in `library.yaml` while actual resource files live in `skills/`, `agents/`, `commands/`, `memory/` directories. Currently, `library add` and `library remove` keep the index synchronized, but manual edits to resource files or filesystem operations (git mv, file manager) can desynchronize the index.

## Goals / Non-Goals

**Goals:**
- Allow users to sync metadata from resource files back into `library.yaml`
- Allow discovery of orphaned resource files not registered in the index
- Align output format with existing CLI conventions (checkmarks, grouped output, --dry-run)
- Collect all errors and report at end (don't fail on first error)

**Non-Goals:**
- Auto-fixing preset refs when resources are renamed (manual operation)
- Auto-removing entries for missing files (use `validate --fix`)
- Auto-renaming resource keys (too risky)
- Bidirectional sync (files are source of truth for content, index is source for path/name)

## Decisions

### Decision 1: `library refresh` as new command vs extension of `library validate`

**Choice:** New `library refresh` command

**Rationale:** `validate --fix` handles missing files. `refresh` handles stale metadata. Splitting them keeps commands focused and predictable.

**Alternatives considered:**
- Extend `validate --fix` to also update descriptions: Would blur command purpose and make behavior less predictable

### Decision 2: Error handling - fail fast vs collect and continue

**Choice:** Collect all errors, report at end, exit code 1 if any errors

**Rationale:** Matches user decision: "fallback to error and manual resolution for any issues that cannot be easily determined by internal logic. Any such error shouldn't prevent the rest of the function to execute"

**Alternatives considered:**
- Fail on first error: Would require multiple runs to resolve all issues

### Decision 3: Path update detection - safe only vs aggressive

**Choice:** Only update path when frontmatter `name` matches entry key (Scenario A: file renamed, frontmatter unchanged)

**Rationale:** Safe default. If frontmatter name doesn't match key, it's a conflict that requires manual resolution.

**Alternatives considered:**
- Always update path when file found elsewhere: Could silently break preset refs
- Never update path: User would have to manually fix every rename

### Decision 4: Orphan discovery - separate command vs flag

**Choice:** `library add --discover` flag

**Rationale:** Keeps "add" as registration command, just with a discovery mode. Cleaner UX than a separate command.

**Alternatives considered:**
- `library discover` subcommand: New verb that doesn't match existing patterns
- `library add` without flag: Would change existing behavior unexpectedly

### Decision 5: Discover behavior - force required vs automatic

**Choice:** Report-only by default, require `--force` to actually add

**Rationale:** Avoids accidentally registering unwanted files. User must opt-in to modifications.

**Alternatives considered:**
- Interactive mode: Adds complexity, harder to script
- Batch add by default: Could register unexpected files

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| User renames file but frontmatter still has old name | Conflict error, manual resolution required |
| Malformed frontmatter during refresh | Skip and error, user must fix manually |
| Orphan file with same name as registered resource | Treated as conflict, manual resolution required |
| Missing files during refresh | Skip silently (left to `validate --fix`) |

## Technical Approach

### Implementation References

Detailed algorithms are specified in:
- `specs/library/library-refresh/spec.md` - Refresh capability requirements
- `specs/library/library-orphan-discovery/spec.md` - Orphan discovery requirements

### Key Implementation Notes

**Refresh**: Search `{type}s/` directories when file not at registered path. Match by frontmatter `name` field.

**Discover**: Type from directory (always authoritative), name/description from frontmatter with filename fallback.

**Output**: Align with existing CLI conventions (checkmarks, grouped issues, hints, JSON structured output).

## Open Questions

None at this time. All decisions made in exploration phase.

## Context

The library system (`internal/infrastructure/library/`) manages canonical resources stored on disk. Metadata is tracked in `library.yaml` which maps resource refs (`type/name`) to file paths. Over time this metadata can drift from reality:

- Entries in `library.yaml` may point to files that were deleted
- Presets may reference resources that were removed
- Files may be added directly without updating `library.yaml`

Currently there's no way to detect these inconsistencies. The only validation is `LoadLibrary()` which checks `library.yaml` structure but not whether referenced files actually exist.

## Goals / Non-Goals

**Goals:**
- Provide a `library validate` command that audits library health
- Detect four issue types: missing files, ghost preset refs, orphaned files, malformed frontmatter
- Support `--fix` to auto-repair `library.yaml` metadata
- Provide human-readable output by default, JSON for scripting
- Follow existing CLI conventions for flags, exit codes, and error handling

**Non-Goals:**
- Fixing malformed frontmatter (too ambiguous to auto-repair)
- Deleting orphaned files (conservative fix only touches `library.yaml`)
- Validating resource content (structural validation is separate)
- Remote library support (filesystem only)

## Decisions

### 1. Validation Logic Location

**Choice:** Place validation logic in `internal/infrastructure/library/validator.go`

**Rationale:** Keeps validation close to where library state lives. The `Loader`, `Resolver`, and `Saver` all live here. Validation reads the same structures.

**Alternatives considered:**
- Application layer: would require new service interface and DI wiring for a single command
- Command layer: would duplicate loading/parsing logic

### 2. Issue Type Design

**Choice:** Four distinct issue types with severity gradation

**Rationale:** Different issue types require different handling:
- `missing-file` (error): entry exists but file gone → fix removes entry
- `ghost-resource` (error): preset refs non-existent resource → fix strips ref
- `orphan` (warning): file exists but no entry → informational, fix ignores
- `malformed` (error): file exists but frontmatter invalid → fix skips (ambiguous)

**Alternatives considered:**
- Single "inconsistency" type: loses ability to distinguish fixable vs informational issues
- More granular types (e.g., "duplicate-entry"): premature complexity

### 3. Fix Behavior

**Choice:** Conservative fix (only modifies `library.yaml`, never deletes files)

**Rationale:** Users may have orphaned files they intend to import later with `library add`. We should not delete user content.

**Alternatives considered:**
- Aggressive fix (delete orphans): Risk of data loss; user can manually delete
- Interactive fix: Adds complexity; `--fix` should be non-interactive

### 4. Exit Codes

**Choice:** Follow existing CLI pattern — `0` clean, `5` validation errors, `1` unexpected

**Rationale:** Consistent with `validate` command (exit 5 for validation failures). "Unexpected errors" (exit 1) for filesystem issues keeps semantic distinction.

### 5. Output Format

**Choice:** Human-readable default, `--json` opt-in

**Rationale:** CLI tools should be usable by humans by default. JSON is secondary for scripting integration.

**Alternatives considered:**
- `--format` flag: over-engineered; only two formats exist
- TTY detection: automatic JSON when piped: adds complexity and edge cases

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Large library with many orphans may produce noisy output | Issues are grouped by type; `--json` allows filtering |
| Malformed frontmatter detection may be too strict | Malformed is severity=error but fix skips it; user must fix manually |
| `--fix` removes entries that user might want to recover | `--fix` only removes from `library.yaml`; files remain on disk |
| Library path discovery differs from other commands | Uses same `library.FindLibrary()` as other library commands |

## Open Questions

None at this time.

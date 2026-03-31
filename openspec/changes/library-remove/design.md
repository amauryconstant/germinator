## Context

Users can add resources and presets via `library add` and `library create preset`, but there is no inverse operation. Currently, removing a resource or preset requires manually editing `library.yaml` and deleting physical files. This is error-prone and creates an asymmetric workflow.

The library system is structured as:
- `library.yaml` - index file with resources and presets
- `{type}s/` directories - physical resource files (e.g., `skills/commit.md`)

## Goals / Non-Goals

**Goals:**
- Provide `library remove resource <ref>` to remove a resource
- Provide `library remove preset <name>` to remove a preset
- Delete physical file AND YAML entry when removing resources
- Delete only YAML entry when removing presets (no physical file)
- Error if a resource is referenced by any preset (explicit decision required)
- Add `--json` flag for scripted/structured output

**Non-Goals:**
- `--dry-run` or `--force` flags (keep it simple)
- Auto-detect type when removing (explicit subcommands)
- Soft-delete or preservation of physical files
- Removing multiple resources at once

## Decisions

### 1. Explicit Subcommands Over Auto-Detect

**Choice:** `library remove resource <ref>` and `library remove preset <name>`

**Rationale:** Being explicit about destructive operations reduces errors. The user must know what they're deleting. Auto-detect could cause accidental deletion if a name conflicts between a resource and preset.

**Alternative Considered:** Auto-detect with `--type resource|preset` flag. More flexible but less safe for destructive operations.

### 2. Delete Physical File for Resources

**Choice:** Delete both the physical file and YAML entry when removing a resource.

**Rationale:** `library add` copies the file into the library, so `library remove` should delete it. Keeping orphaned files creates confusion.

**Alternative Considered:** Only remove from YAML (soft delete). Would require users to manually clean up files, which is the current problem.

### 3. Check Preset References Before Resource Deletion

**Choice:** Error if any preset references the resource being deleted.

**Rationale:** Silent removal from presets could leave broken references. The user must explicitly decide what to do with affected presets.

**Alternative Considered:** Auto-remove from presets. Risky - user might not realize the preset is now incomplete.

### 4. JSON Output Structure

**Choice:** Different JSON schemas for resource vs preset removal.

**Rationale:** Resource removal returns file deleted info; preset removal returns former contents (useful for undo scenarios).

**Resource removal:**
```json
{
  "type": "resource",
  "resourceType": "skill",
  "name": "commit",
  "fileDeleted": "skills/commit.md",
  "libraryPath": "/path/to/library"
}
```

**Preset removal:**
```json
{
  "type": "preset",
  "name": "git-workflow",
  "resourcesRemoved": ["skill/commit", "skill/pr"]
}
```

### 5. Errors Always Structured with `--json`

**Choice:** When `--json` is used, errors are also returned as JSON.

**Rationale:** Consistent scripting experience - no need to parse human-readable error messages.

**Error format:**
```json
{
  "error": "resource not found",
  "type": "resource",
  "resourceType": "skill",
  "name": "nonexistent"
}
```

## Implementation

### New Files

| File | Purpose |
|------|---------|
| `cmd/library_remove.go` | Command with `resource` and `preset` subcommands |
| `internal/infrastructure/library/remover.go` | `RemoveResource()` and `RemovePreset()` |

### Remove Resource Flow

```
1. Parse ref (e.g., "skill/commit") → type, name
2. Load library from {libraryPath}/library.yaml
3. Verify resource exists → error if not
4. Check no presets reference this resource → error if in use
5. Delete physical file: {libraryPath}/{type}s/{name}.md
6. Remove from library.yaml: Resources[type][name]
7. Save library.yaml
8. Output (--json or human)
```

### Remove Preset Flow

```
1. Parse name (e.g., "git-workflow")
2. Load library from {libraryPath}/library.yaml
3. Verify preset exists → error if not
4. Capture resources list for output
5. Remove from library.yaml: Presets[name]
6. Save library.yaml
7. Output (--json or human)
```

### Key Functions (remover.go)

```go
type RemoveResourceOptions struct {
    Ref        string // e.g., "skill/commit"
    LibraryPath string
    JSON       bool
}

type RemovePresetOptions struct {
    Name        string
    LibraryPath string
    JSON       bool
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Accidental deletion of wrong resource | Explicit subcommands force user to be intentional |
| Preset contains last reference to resource | Error forces user to handle preset first |
| File deletion leaves no trace | JSON output includes what was deleted |
| YAML update corrupts library.yaml | Validation after update ensures well-formedness |

## Open Questions

None at this time. Decisions have been made based on the proposal requirements.

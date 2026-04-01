**Location**: `internal/infrastructure/library/`
**Parent**: See `/internal/AGENTS.md` for infrastructure layer overview

---

# Library Infrastructure

Library management for canonical resources (skills, agents, commands, memory).

## Files

| File | Purpose |
|------|---------|
| `library.go` | `Library` struct and `Exists()` function |
| `loader.go` | `LoadLibrary()` - loads and parses library.yaml |
| `creator.go` | `CreateLibrary()` - scaffolds new library directories |
| `discovery.go` | `FindLibrary()`, `DefaultLibraryPath()` |
| `lister.go` | `ListResources()` - groups resources by type |
| `resolver.go` | `ResolveResource()` - resolves refs to full paths |
| `adder.go` | `AddResource()` - imports resources into library; also `DiscoverOrphans()` for orphan discovery |
| `refresher.go` | `RefreshLibrary()` - syncs metadata from resource files into library.yaml |
| `saver.go` | `SaveLibrary()`, `AddPreset()`, `PresetExists()` |
| `remover.go` | `RemoveResource()`, `RemovePreset()` - remove resources/presets |
| `validator.go` | `ValidateLibrary()` - checks library integrity (missing files, ghosts, orphans, malformed) |
| `library_test.go` | Tests for Library struct and Exists |
| `loader_test.go` | Tests for LoadLibrary |
| `lister_test.go` | Tests for ListResources |
| `resolver_test.go` | Tests for ResolveResource |
| `discovery_test.go` | Tests for FindLibrary |
| `adder_test.go` | Tests for AddResource and DiscoverOrphans |
| `refresher_test.go` | Tests for RefreshLibrary |
| `saver_test.go` | Tests for SaveLibrary and AddPreset |
| `remover_test.go` | Tests for RemoveResource and RemovePreset |
| `validator_test.go` | Tests for ValidateLibrary and fix operations |

## Core Types

### Library

```go
type Library struct {
    Version   string
    Resources Resources
    Presets   map[string]Preset
}
```

### Resources

```go
type Resources struct {
    Skill    map[string]SkillRef
    Agent    map[string]AgentRef
    Command  map[string]CommandRef
    Memory   map[string]MemoryRef
}
```

## Library Discovery

Priority: explicit path > `GERMINATOR_LIBRARY` env > `$XDG_DATA_HOME/germinator/library/` or `~/.local/share/germinator/library/`

```go
// Get default library path
DefaultLibraryPath() string

// Find library with priority resolution
FindLibrary(explicitPath string) (string, error)
```

## Loading

```go
// Load library from path
LoadLibrary(path string) (*Library, error)
```

Validation: Parses `library.yaml`, validates structure, returns error if invalid.

## Creation

```go
type CreateOptions struct {
    Path   string  // Library directory path
    DryRun bool    // Preview without creating
    Force  bool     // Overwrite existing
}

// Create new library directory structure
CreateLibrary(opts CreateOptions) error
```

Creates:
- `library.yaml` with version "1" and empty resources/presets
- `skills/`, `agents/`, `commands/`, `memory/` directories

Post-creation: Validates by calling `LoadLibrary()` to ensure well-formed structure.

## Listing

```go
// List all resources grouped by type
ListResources(lib *Library) ResourceList

type ResourceList struct {
    Skills    []Resource
    Agents    []Resource
    Commands  []Resource
    Memories  []Resource
}
```

## Resolution

```go
// Resolve ref to full path (e.g., "skill/commit" -> "/path/to/skills/commit")
ResolveResource(libPath, ref string) (string, error)
```

## Adding Resources

```go
type AddOptions struct {
    SourcePath   string  // Source file to import
    Name         string  // Resource name (auto-detected if empty)
    Description  string  // Resource description (auto-detected if empty)
    Type         string  // Resource type: agent, command, skill, memory (auto-detected)
    Platform     string  // Source platform: opencode, claude-code (auto-detected)
    LibraryPath  string  // Target library path
    DryRun       bool    // Preview without modifying
    Force        bool    // Overwrite existing
}

// Add resource to library (imports, canonicalizes if needed, validates, registers)
AddResource(opts AddOptions) error
```

Type detection priority: `--type` flag > frontmatter `type:` > filename pattern
Platform detection: `--platform` flag > frontmatter `platform:` > auto-detect from content
Target path: `{library}/{type}s/{name}.md` (e.g., `library/agents/reviewer.md`)

Validation: Validates canonical document before adding; validates library.yaml after update.

## Preset Management

```go
// SaveLibrary persists the library to library.yaml
SaveLibrary(lib *Library) error

// AddPreset adds a preset to the library in-memory
AddPreset(lib *Library, preset Preset) error

// PresetExists checks if a preset with the given name exists
PresetExists(lib *Library, name string) bool
```

**SaveLibrary**: Marshals the entire library struct and writes to `{RootPath}/library.yaml`. Creates directory if needed.

**AddPreset**: Validates preset before adding; initializes Presets map if nil.

**PresetExists**: Returns false if library or Presets map is nil.

## Removing Resources

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

// RemoveResource removes a resource from the library (deletes file + YAML entry)
RemoveResource(opts RemoveResourceOptions) error

// RemovePreset removes a preset from the library YAML
RemovePreset(opts RemovePresetOptions) error
```

### RemoveResource Flow

1. Parse ref (e.g., "skill/commit") → type, name
2. Load library from `{libraryPath}/library.yaml`
3. Verify resource exists → error if not
4. Check no presets reference this resource → error if in use
5. Delete physical file: `{libraryPath}/{type}s/{name}.md`
6. Remove from library.yaml: `Resources[type][name]`
7. Save library.yaml
8. Output (--json or human)

### RemovePreset Flow

1. Parse name (e.g., "git-workflow")
2. Load library from `{libraryPath}/library.yaml`
3. Verify preset exists → error if not
4. Capture resources list for output
5. Remove from library.yaml: `Presets[name]`
6. Save library.yaml
7. Output (--json or human)

### JSON Output

**Resource removal:**
```json
{
  "type": "resource",
  "resourceType": "skill",
  "name": "commit",
  "fileDeleted": "/path/to/library/skills/commit.md",
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

## Validation

```go
// ValidateLibrary checks library integrity and returns issues
ValidateLibrary(libPath string) ([]Issue, error)

// FixLibrary auto-repairs library.yaml (removes missing entries, strips ghost refs)
FixLibrary(libPath string) ([]Issue, error)
```

### Issue Types

| Type | Severity | Description | Fix Action |
|------|----------|-------------|------------|
| `missing-file` | error | Entry in library.yaml but file doesn't exist | Remove entry |
| `ghost-resource` | error | Preset references non-existent resource | Strip ref from preset |
| `orphan` | warning | File exists on disk but not in library.yaml | None (informational) |
| `malformed` | error | Resource file has invalid YAML frontmatter | None (manual fix required) |

### Fix Behavior

- **Conservative**: Only modifies `library.yaml`, never deletes actual files
- `--fix` removes missing file entries and strips ghost preset references
- Orphaned files are informational only (not auto-deleted)
- Malformed frontmatter cannot be auto-repaired

## Refresh

```go
type RefreshOptions struct {
    LibraryPath string
    DryRun      bool
    Force       bool
}

type RefreshResult struct {
    Refreshed []RefreshChange
    Skipped   []RefreshSkipped
    Errors    []RefreshError
}

type RefreshChange struct {
    Ref   string
    Field string // "description" or "path"
    Old   string
    New   string
}

type RefreshSkipped struct {
    Ref    string
    Reason string // "missing_file"
}

type RefreshError struct {
    Ref    string
    Reason string // "name_mismatch", "malformed_frontmatter"
}

// RefreshLibrary syncs metadata from registered resource files into library.yaml
RefreshLibrary(opts RefreshOptions) (*RefreshResult, error)
```

### Refresh Behavior

| Scenario | Action |
|----------|--------|
| Description stale | Update from frontmatter |
| File renamed, name matches key | Update path |
| File renamed, name mismatch | Error (use `--force` to skip) |
| File missing | Skip silently |
| Malformed frontmatter | Error (use `--force` to skip) |

- **Collects all errors**: Does not fail on first error; reports all at end
- **Exit code 1** if any errors occurred
- **--force** skips conflicting resources instead of erroring

## Orphan Discovery

```go
type DiscoverOptions struct {
    LibraryPath string
    DryRun      bool
    Force       bool
}

type Orphan struct {
    Name        string
    Type        string // "skill", "agent", "command", "memory"
    Description string
    Path        string
    HasConflict bool // true if name matches existing resource
}

// DiscoverOrphans finds resource files not registered in library.yaml
DiscoverOrphans(opts DiscoverOptions) ([]Orphan, error)
```

### Discover Behavior

- Scans `skills/`, `agents/`, `commands/`, `memory/` directories
- Type detected from directory (authoritative)
- Name from frontmatter `name` field or filename fallback
- Description from frontmatter `description` field
- **Report-only by default**: Use `--force` to actually register orphans
- **Conflict detection**: Orphan name matching existing resource is flagged

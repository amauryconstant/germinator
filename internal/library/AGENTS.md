**Location**: `internal/library/`
**Parent**: See `/internal/AGENTS.md` for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Library Package

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
| `adder.go` | `AddResource()` - imports resources; `BatchAddResources()` for batch operations; `DiscoverOrphans()` for orphan discovery |
| `refresher.go` | `RefreshLibrary()` - syncs metadata from resource files into library.yaml |
| `saver.go` | `SaveLibrary()`, `AddPreset()`, `PresetExists()` |
| `remover.go` | `RemoveResource()`, `RemovePreset()` - remove resources/presets |
| `validator.go` | `ValidateLibrary()` - checks library integrity (missing files, ghosts, orphans, malformed) |
| `requests.go` | Request/result types for `(*Library) X` methods (`InitRequest`, `RefreshRequest`, `RemoveResourceRequest`, `RemovePresetRequest`, `ValidateRequest`, `FixRequest`, `RefreshUnchanged`, `FixResult`) |
| `library_test.go` | Tests for Library struct and Exists |
| `methods_test.go` | Table-driven tests for each `(*Library) X` method (success + each error path + ctx cancellation) |
| `loader_test.go` | Tests for LoadLibrary |
| `lister_test.go` | Tests for ListResources |
| `resolver_test.go` | Tests for ResolveResource |
| `discovery_test.go` | Tests for FindLibrary |
| `adder_test.go` | Tests for AddResource, BatchAddResources, and DiscoverOrphans |
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
FindLibrary(flagPath, envPath string) string
```

## Loading

```go
// Load library from path. ctx is checked between I/O operations; on
// cancellation ctx.Err() is returned, wrapped.
LoadLibrary(ctx context.Context, path string) (*Library, error)
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
type AddRequest struct {
    Source      string  // Source file to import
    Name        string  // Resource name (auto-detected if empty)
    Description string  // Resource description (auto-detected if empty)
    Type        string  // Resource type: agent, command, skill, memory (auto-detected)
    LibraryPath string  // Target library path
    DryRun      bool    // Preview without modifying
    Force       bool    // Overwrite existing
}

// Add resource to library (imports, canonicalizes if needed, validates, registers).
// ctx is checked before each I/O step; on cancellation it returns wrapped ctx.Err().
AddResource(ctx context.Context, opts AddRequest) error
```

Type detection priority: `--type` flag > frontmatter `type:` > filename pattern
Platform detection: `--platform` flag > frontmatter `platform:` > auto-detect from content
Target path: `{library}/{type}s/{name}.md` (e.g., `library/agents/reviewer.md`)

Validation: Validates canonical document before adding; validates library.yaml after update.

## Batch Adding Resources

```go
type BatchAddResult struct {
    Added   []BatchAddSuccess  `json:"added"`
    Skipped []BatchSkipInfo   `json:"skipped"`
    Failed  []BatchFailureInfo `json:"failed"`
    Summary BatchSummary       `json:"summary"`
}

type BatchAddSuccess struct {
    Ref  string `json:"ref"`  // Resource reference (e.g., "skill/commit")
    Path string `json:"path"` // Path in library
}

type BatchSkipInfo struct {
    Source string `json:"source"` // Original source path
    Issue  string `json:"issue"`  // "already_exists" or "conflict"
}

type BatchFailureInfo struct {
    Source string `json:"source"` // Original source path
    Error  string `json:"error"`  // Error message
}

type BatchSummary struct {
    Total   int `json:"total"`
    Added   int `json:"added"`
    Skipped int `json:"skipped"`
    Failed  int `json:"failed"`
}

type BatchAddOptions struct {
    Sources     []string     // Source files/directories to add
    LibraryPath string       // Path to the library
    DryRun      bool         // Preview without modifying
    Force       bool         // Overwrite existing resources
    Name        string       // Optional resource name override
    Description string       // Optional resource description override
    Type        string       // Optional resource type override
    Platform    string       // Optional platform override
    Orphans     []Orphan     // Orphan info for discovered resources (provides type/name)
}

// BatchAddResources adds multiple resources to the library in batch mode.
// ctx is checked between files; on cancellation partial results are returned
// alongside wrapped ctx.Err().
BatchAddResources(ctx context.Context, opts BatchAddOptions) (*BatchAddResult, error)
```

### Batch Behavior

| Scenario | Behavior |
|----------|----------|
| Directory in sources | Recursively scanned for `*.md` files |
| Already exists (without `--force`) | Skipped with issue "already_exists" |
| Name conflict with existing resource | Skipped with issue "conflict" |
| File read/parse error | Failure recorded, continues processing |
| `--dry-run` | Preview what would be added/skipped/failed |

**Result categories:**
- **Added**: Successfully imported and registered
- **Skipped**: File processed but not added (already exists or conflict)
- **Failed**: Unexpected error during processing

**Exit code**: Batch mode always returns `nil` error (exit 0), even if some failed. Scripts should inspect `BatchAddResult` for details.

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
    Unchanged []RefreshUnchanged   // always populated when present in the library
    Skipped   []SkipInfo
    Errors    []RefreshError
}

type RefreshUnchanged struct {
    Ref        string
    LastSynced string  // RFC3339 mtime, or "" when not determinable
}

type RefreshChange struct {
    Ref   string
    Field string // "description" or "path"
    Old   string
    New   string
}

type SkipInfo struct {
    Ref    string
    Reason string // "missing_file"
}

type RefreshError struct {
    Ref   string
    Field string
    Type  string
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

## Methods on `*Library`

Mutating operations live as methods on `*Library` so the cmd-side can declare a minimal interface per command (no `*LibraryService` wrapper). Mirrors the `(*Library).CreatePreset` precedent at `internal/library/creator.go`.

```go
func (lib *Library) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResult, error)
func (lib *Library) RemoveResource(ctx context.Context, req *RemoveResourceRequest) error
func (lib *Library) RemovePreset(ctx context.Context, req *RemovePresetRequest) error
func (lib *Library) Validate(ctx context.Context, req *ValidateRequest) (*ValidationResult, error)
func (lib *Library) Fix(ctx context.Context, _ *FixRequest) (*FixResult, error)

// Init is a package function (no pre-existing *Library to receive a method)
func Init(ctx context.Context, req *InitRequest) error
```

All methods assert `lib != nil && lib.RootPath != ""` at entry. See `methods_test.go` for table-driven coverage. The package-level functions (`RefreshLibrary`, `RemoveResource`, `RemovePreset`, `ValidateLibrary`, `FixLibrary`) preserve their existing public signatures and delegate internally to the method form.

## Orphan Discovery

```go
// ErrNameConflict is returned by checkNameConflict when an orphan name
// collides with an existing resource of a different type. Callers use
// errors.Is(err, ErrNameConflict) to detect a typed name-conflict error
// (this is the producer-side half of the contract; the consumer-side
// half is verified in task 6.4's runAdd tests via *core.OperationError's
// Unwrap chain).
var ErrNameConflict = errors.New("name conflict with existing resource")

type DiscoverOptions struct {
    LibraryPath string
    DryRun      bool
    Force       bool
    Batch       bool // Process all orphans continuously
}

type Orphan struct {
    Path  string `json:"path"`
    Type  string `json:"type"` // "skill", "agent", "command", "memory"
    Name  string `json:"name"`
    Issue string `json:"issue,omitempty"` // "name_conflict" or empty
}

type ConflictInfo struct {
    Orphan Orphan `json:"orphan"`
    Issue  string `json:"issue"` // "<type>/<name>: <wrapped ErrNameConflict>"
}

type AddSuccess struct {
    Type string `json:"type"`
    Name string `json:"name"`
    Path string `json:"path"`
}

type DiscoverSummary struct {
    TotalScanned  int `json:"totalScanned"`
    TotalOrphans  int `json:"totalOrphans"`
    TotalAdded    int `json:"totalAdded"`
    TotalSkipped  int `json:"totalSkipped"`
    TotalFailed   int `json:"totalFailed"`
}

type DiscoverResult struct {
    Orphans   []Orphan   `json:"orphans"`
    Added     []AddSuccess   `json:"added"`
    Conflicts []ConflictInfo `json:"conflicts"`
    Summary   DiscoverSummary `json:"summary"`
}

// DiscoverOrphans finds resource files not registered in library.yaml.
// ctx is checked between files; on cancellation a partial DiscoverResult
// is returned alongside wrapped ctx.Err().
DiscoverOrphans(ctx context.Context, opts DiscoverOptions) (*DiscoverResult, error)

// checkNameConflict returns ErrNameConflict wrapped with the offending
// "<type>/<name>" ref when the orphan collides with an existing
// resource of a different type; nil otherwise. Same-type duplicates
// are handled elsewhere (isRegistered + AddResource).
checkNameConflict(lib *Library, orphan *Orphan) error
```

### Discover Behavior

- Scans `skills/`, `agents/`, `commands/`, `memory/` directories **recursively**
- Type detected from top-level directory (authoritative)
- Name from frontmatter `name` field or filename fallback
- **Report-only by default**: Use `--force` to actually register orphans
- **Conflict detection**: Orphan name matching existing resource is flagged with `Issue: "name_conflict"`

### Batch Mode

When `--batch` is enabled with `--force`:
- Processes all discovered orphans continuously
- Skips individual registration errors without stopping
- Reports TotalAdded, TotalSkipped, TotalFailed in summary
- Use `--dry-run` to preview without modifying

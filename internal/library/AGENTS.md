**Location**: `internal/library/`
**Parent**: See `/internal/AGENTS.md` for package overview
**Skill references**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`, `@.opencode/skills/golang-testing/SKILL.md` (for the `goleak` test-main / `t.Parallel` patterns)

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
| `resolver.go` | `ResolveResource()`, `ResolveResourceEntry()`, `ResolvePresetEntry()`; `(*Library).ResolvePreset` — resolve refs/names to typed results with `*core.NotFoundError` on miss |
| `adder.go` | `AddResource()` - imports resources; `BatchAddResources()` for batch operations; `DiscoverOrphans()` for orphan discovery |
| `refresher.go` | `RefreshLibrary()` - syncs metadata from resource files into library.yaml |
| `saver.go` | `SaveLibrary()`, `AddPreset()`, `PresetExists()` |
| `remover.go` | `RemoveResource()`, `RemovePreset()` - remove resources/presets |
| `validator.go` | `ValidateLibrary()` - checks library integrity (missing files, ghosts, orphans, malformed) |
| `requests.go` | Request/result types for `(*Library) X` methods (`InitRequest`, `RefreshRequest`, `RemoveResourceRequest`, `RemovePresetRequest`, `ValidateRequest`, `FixRequest`, `RefreshUnchanged`, `FixResult`) |
| `library_test.go` | Tests for Library struct and Exists; declares `TestMain` for `goleak.VerifyTestMain` (goroutine leak detection) |
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

Priority: explicit path > `GERMINATOR_LIBRARY` env > config file (`Config.Library`) > `$XDG_DATA_HOME/germinator/library/` or `~/.local/share/germinator/library/`

```go
// Get default library path (XDG via adrg/xdg, with CWD fallback)
DefaultLibraryPath() string

// Find library with 4-tier priority resolution (flag > env > cfg > default)
FindLibrary(flagPath, envPath, cfgPath string) string
```

The 3-arg `FindLibrary` encodes the spec-mandated precedence directly:
1. `flagPath` — `--library` flag (highest)
2. `envPath` — `GERMINATOR_LIBRARY` env var
3. `cfgPath` — `Config.Library` (config-file override)
4. `DefaultLibraryPath()` — XDG via `adrg/xdg.DataHome`, falling back to `./germinator/library/` for project-local libraries

`DefaultLibraryPath()` reads `xdg.DataHome` directly (not `xdg.DataFile`) to avoid creating the directory on disk; mutex-protected via `xdgReload()` so parallel tests see updated values after `t.Setenv`.

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
CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error
```

Creates:
- `library.yaml` with version "1" and empty resources/presets
- `skills/`, `agents/`, `commands/`, `memory/` directories

Post-creation: Validates by calling `LoadLibrary()` to ensure well-formed structure. The `ctx` parameter is forwarded to the validation load; `stdout` receives the DryRun summary when non-nil (or is suppressed entirely when nil).

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
    Source, Name, Description, Type, LibraryPath string
    DryRun, Force                                bool
}

// AddResource imports, canonicalizes, validates, and registers a resource.
// ctx is checked before each I/O step; on cancellation returns wrapped ctx.Err().
AddResource(ctx context.Context, opts AddRequest) error
```

| Detection | Priority order |
|---|---|
| Type | `--type` flag > frontmatter `type:` > filename pattern |
| Platform | `--platform` flag > frontmatter `platform:` > auto-detect from content |

Target path: `{library}/{type}s/{name}.md` (e.g., `library/agents/reviewer.md`). Validates canonical document before adding; validates `library.yaml` after update.

## Batch Adding Resources

```go
type BatchAddResult struct {
    Added   []BatchAddSuccess        `json:"added"`
    Skipped []BatchSkipInfo          `json:"skipped"`
    Failed  []BatchFailureInfo       `json:"failed"`
    Summary BatchSummary             `json:"summary"`
}

type BatchAddOptions struct {
    Sources                                 []string // files/dirs to add
    LibraryPath, Name, Description          string
    Type, Platform                          string
    Orphans                                 []Orphan  // from DiscoverOrphans
    DryRun, Force                           bool
}

// BatchAddResources adds multiple resources in batch mode.
// ctx is checked between files; on cancellation returns partial results + wrapped ctx.Err().
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
RemoveResource(ctx context.Context, opts RemoveResourceOptions) error

// RemovePreset removes a preset from the library YAML
RemovePreset(ctx context.Context, opts RemovePresetOptions) error
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

`{type, resourceType, name, fileDeleted, libraryPath}` for resources; `{type, name, resourcesRemoved}` for presets.

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
type RefreshOptions struct { LibraryPath string; DryRun, Force bool }
type RefreshResult   struct { Refreshed []RefreshChange; Unchanged []RefreshUnchanged; Skipped []SkipInfo; Errors []RefreshError }
type RefreshChange   struct { Ref, Field, Old, New string }                                       // Field: "description" | "path"
type RefreshUnchanged struct { Ref, LastSynced string }                                           // LastSynced: RFC3339 mtime or ""
type SkipInfo         struct { Ref, Reason string }                                               // Reason: "missing_file"
type RefreshError     struct { Ref, Field, Type string }

RefreshLibrary(ctx context.Context, opts RefreshOptions) (*RefreshResult, error)
```

### Refresh Behavior

| Scenario | Action |
|----------|--------|
| Description stale | Update from frontmatter |
| File renamed, name matches key | Update path |
| File renamed, name mismatch | Error (use `--force` to skip) |
| File missing | Skip silently |
| Malformed frontmatter | Error (use `--force` to skip) |

**--force** skips conflicting resources instead of erroring. **Exit code 1** if any errors occurred (all errors collected; no fail-fast).

## Methods on `*Library`

Mutating operations live as methods on `*Library` so the cmd-side can declare a minimal interface per command (no `*LibraryService` wrapper). Mirrors the `(*Library).CreatePreset` precedent at `internal/library/creator.go`.

```go
func (lib *Library) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResult, error)
func (lib *Library) RemoveResource(ctx context.Context, req *RemoveResourceRequest) error
func (lib *Library) RemovePreset(ctx context.Context, req *RemovePresetRequest) error
func (lib *Library) Validate(ctx context.Context, req *ValidateRequest) (*ValidationResult, error)
func (lib *Library) Fix(ctx context.Context, _ *FixRequest) (*FixResult, error)
func (lib *Library) ResolvePreset(ctx context.Context, name string) ([]string, error) // ctx accept-and-may-ignore (pure in-memory lookup)
func (lib *Library) Add(ctx context.Context, req *AddRequest) error
func (lib *Library) BatchAddResources(ctx context.Context, opts *BatchAddOptions) (*BatchAddResult, error)
func (lib *Library) DiscoverOrphans(ctx context.Context, opts *DiscoverOptions) (*DiscoverResult, error)

// Init is a package function (no pre-existing *Library to receive a method)
func Init(ctx context.Context, req *InitRequest) error
```

All methods assert `lib != nil && lib.RootPath != ""` at entry (see `methods_test.go` for table-driven coverage). The methods delegate to the package-level functions (`RefreshLibrary`, `RemoveResource`, `RemovePreset`, `AddResource`, `BatchAddResources`, `DiscoverOrphans`, `CreateLibrary`), forwarding `ctx` through. `ValidateLibrary(lib)` and `FixLibrary(lib)` are pure in-memory operations and do not take ctx. The `*Library` methods satisfy the cmd-side `adderLibrary` interface in `cmd/library_add.go` directly — no adapter shim.

## Orphan Discovery

```go
var ErrNameConflict = errors.New("name conflict with existing resource")

type DiscoverOptions struct {
    LibraryPath string
    DryRun, Force, Batch bool // Batch: process all orphans continuously
}

type Orphan       struct { Path, Type, Name string; Issue string `json:"issue,omitempty"` }
type ConflictInfo struct { Orphan Orphan; Issue string }
type AddSuccess   struct { Type, Name, Path string }
type DiscoverSummary struct {
    TotalScanned, TotalOrphans, TotalAdded, TotalSkipped, TotalFailed int
}

type DiscoverResult struct {
    Orphans   []Orphan
    Added     []AddSuccess
    Conflicts []ConflictInfo
    Summary   DiscoverSummary
}

DiscoverOrphans(ctx context.Context, opts DiscoverOptions) (*DiscoverResult, error)
checkNameConflict(lib *Library, orphan *Orphan) error // wraps ErrNameConflict on type-conflict
```

`ErrNameConflict` is the typed name-conflict sentinel — callers detect it via `errors.Is`. `DiscoverOrphans` checks `ctx` between files; on cancellation a partial `DiscoverResult` is returned alongside wrapped `ctx.Err()`.

### Discover Behavior

- Scans `skills/`, `agents/`, `commands/`, `memory/` directories **recursively**
- Type detected from top-level directory (authoritative)
- Name from frontmatter `name` field or filename fallback
- **Report-only by default**: Use `--force` to actually register orphans
- **Conflict detection**: Orphan name matching existing resource is flagged with `Issue: "name_conflict"`

### Batch Mode (`--batch --force`)

Processes all discovered orphans continuously; skips individual registration errors. Reports `TotalAdded`, `TotalSkipped`, `TotalFailed` in `Summary`. Use `--dry-run` to preview.

## Goroutine Leak Detection

The package uses `go.uber.org/goleak` via a package-level `TestMain` in `library_test.go`:

```go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

`goleak.VerifyTestMain` wraps `m.Run()` and verifies no goroutines remain after the suite completes — catching leaks from `adder.go:scanDirectory` / `scanLevel` (which use `errgroup.SetLimit` for concurrent orphan scans).

> **Note**: `VerifyTestMain` already wraps `m.Run()` and exits the process. Do **NOT** append `os.Exit(m.Run())` after it — that runs the suite twice. If a legitimate test must spawn a long-lived goroutine (e.g., a watcher), use `goleak.IgnoreTopFunction` for the relevant test or exclude the goroutine's function name via `goleak.VerifyTestMain`'s `opts...Ignore*` parameters.

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
| `adder.go` | `AddResource()` - imports resources into library |
| `library_test.go` | Tests for Library struct and Exists |
| `loader_test.go` | Tests for LoadLibrary |
| `lister_test.go` | Tests for ListResources |
| `resolver_test.go` | Tests for ResolveResource |
| `discovery_test.go` | Tests for FindLibrary |
| `adder_test.go` | Tests for AddResource |

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

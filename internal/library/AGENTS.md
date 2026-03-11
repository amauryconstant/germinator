**Location**: `internal/library/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Library Package

Manage canonical resource libraries with indexed storage and preset grouping.

## Files

| File | Purpose |
|------|---------|
| `library.go` | Types (Library, Resource, Preset, ResourceType) |
| `loader.go` | LoadLibrary, library.yaml parsing |
| `resolver.go` | ResolveResource, ResolvePreset, GetOutputPath |
| `lister.go` | ListResources, ListPresets, ResourceInfo |
| `discovery.go` | FindLibrary, DefaultLibraryPath |

---

# Types

## Library

```go
type Library struct {
    Version   string                         // Library format version ("1")
    RootPath  string                         // Absolute path to library directory
    Resources map[string]map[string]Resource // type → name → resource
    Presets   map[string]Preset              // name → preset
}
```

## Resource

```go
type Resource struct {
    Path        string // Relative path from library root
    Description string // Human-readable description
}
```

## Preset

```go
type Preset struct {
    Name        string   // Preset identifier
    Description string   // Human-readable description
    Resources   []string // Resource refs in "type/name" format
}
```

## ResourceType

```go
type ResourceType string

const (
    ResourceTypeSkill   ResourceType = "skill"
    ResourceTypeAgent   ResourceType = "agent"
    ResourceTypeCommand ResourceType = "command"
    ResourceTypeMemory  ResourceType = "memory"
)
```

---

# Library Loading

## LoadLibrary

```go
lib, err := LoadLibrary("/path/to/library")
```

- Expects `library.yaml` in directory
- Validates version (only "1" supported)
- Validates resource types and paths
- Validates preset resource references

## library.yaml Format

```yaml
version: "1"
resources:
  skill:
    commit:
      path: skills/skill-commit.md
      description: Git commit best practices
  agent:
    reviewer:
      path: agents/agent-reviewer.md
      description: Code review agent
presets:
  git-workflow:
    description: Git workflow tools
    resources:
      - skill/commit
      - skill/merge-request
```

---

# Resource Resolution

## ResolveResource

```go
path, err := ResolveResource(lib, "skill/commit")
// Returns absolute path to resource file
```

## ResolvePreset

```go
refs, err := ResolvePreset(lib, "git-workflow")
// Returns []string{"skill/commit", "skill/merge-request"}
```

## ParseRef

```go
typ, name, err := ParseRef("skill/commit")
// Returns "skill", "commit", nil
```

## GetOutputPath

```go
path, err := GetOutputPath("skill", "commit", "opencode", ".")
// Returns ".opencode/skills/commit/SKILL.md"
```

Platform-specific paths:

| Type | OpenCode | Claude Code |
|------|----------|-------------|
| skill | `.opencode/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` |
| agent | `.opencode/agents/<name>.md` | `.claude/agents/<name>.md` |
| command | `.opencode/commands/<name>.md` | `.claude/commands/<name>.md` |
| memory | `.opencode/memory/<name>.md` | `.claude/memory/<name>.md` |

---

# Library Discovery

## FindLibrary

Priority chain: flag > env > default

```go
path := FindLibrary(flagPath, envPath)
```

## DefaultLibraryPath

Uses `os.UserConfigDir` for XDG compliance:
- Linux: `~/.config/germinator/library/`
- macOS: `~/Library/Application Support/germinator/library/`
- Windows: `%APPDATA%/germinator/library/`

---

# Listing

## ListResources

```go
resources := ListResources(lib)
// Returns map[string][]ResourceInfo grouped by type
```

## ListPresets

```go
presets := ListPresets(lib)
// Returns []Preset sorted by name
```

---

# Error Messages

Terse errors (no suggestions):
- `resource not found: skill/nonexistent`
- `preset not found: nonexistent`
- `invalid resource reference format: invalidformat`
- `library not found: /path/to/library`
- `library.yaml not found: /path/to/library/library.yaml`
- `unsupported library version: 2 (expected 1)`

---

# Testing

Table-driven tests in `*_test.go` files.

Test fixtures in `test/fixtures/library/`:
- `library.yaml` - Sample library index
- `skills/`, `agents/`, `commands/`, `memory/` - Sample resources

Run tests:
```bash
go test ./internal/library/... -v
```

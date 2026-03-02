## Why

Germinator transforms AI coding assistant documents between platforms but lacks a way to store and manage reusable canonical resources. Users managing multiple skills, agents, and commands across projects need:

1. **A central library** to organize, discover, and reference reusable resources
2. **An init command** to install pre-defined configurations to new projects, accelerating setup and promoting best practices

The Library System provides indexed storage with preset grouping, and the Init command provides batch transformation of library resources to platform-specific output files.

## What Changes

### Library System
- Add `internal/library/` package with types, loader, resolver, lister, and discovery
- Define `library.yaml` index format with resources nested by type (`resources: {type: {name: {...}}}`) supporting skill, agent, command, and memory resource types
- Support library path discovery via flag, environment variable, and default

### CLI Commands
- `germinator library` command with subcommands:
  - `resources` - List all resources in library (grouped by type)
  - `presets` - List all presets in library
  - `show <ref>` - Display resource or preset details (ref format: `type/name`)
- `germinator init` command:
  - Install resources from library to target project
  - Support explicit resource selection (`--resources commit,merge-request`)
  - Support presets for bundled resources (`--preset git-workflow`)
  - Derive platform-specific output paths from resource name and type

### Service Layer
- Add `internal/services/initializer.go` - orchestrates init workflow, reusing existing transform pipeline

## Capabilities

### New Capabilities

- `library-system`: Load, resolve, and discover indexed canonical resources (skills, agents, commands) with preset grouping
- `resource-installation`: Batch transformation of canonical resources to platform-specific output files
- `init-command`: CLI command for installing resources from a library to a target project

### Modified Capabilities

(none)

## Impact

**New code:**
- `internal/library/library.go` - Library, Resource, Preset types (Resources nested by type, supports skill/agent/command/memory)
- `internal/library/loader.go` - LoadLibrary, library.yaml parsing
- `internal/library/resolver.go` - ResolveResource(ref), ResolvePreset, ParseRef, GetOutputPath
- `internal/library/lister.go` - ListResources, ListPresets, FormatList
- `internal/library/discovery.go` - FindLibrary (flag/env/default)
- `internal/services/initializer.go` - Init orchestration service
- `cmd/library.go` - CLI commands (resources, presets, show)
- `cmd/init.go` - CLI command for resource installation

**Reused code:**
- `internal/core/loader.go` - LoadDocument (unchanged)
- `internal/core/serializer.go` - RenderDocument (unchanged)

**No changes to:**
- Existing transformation pipeline
- Existing adapters or models
- Existing CLI commands (except root command registration)

**User impact:**
- Users must create their own library at `~/.config/germinator/library/` (no default shipped)
- New CLI commands available: `germinator library` and `germinator init`

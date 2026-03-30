## Why

The library system currently supports listing, showing, and installing resources from a library to projects (via `init`), but lacks the ability to add new resources to the library. Users who create new skills, agents, commands, or memory documents must manually copy files and edit `library.yaml` to register them. The `add` command automates this workflow.

## What Changes

- New `germinator library add <source>` command that imports a resource into the library
- Auto-detection of resource type from filename pattern, frontmatter, or `--type` flag
- Auto-detection of resource name and description from frontmatter or `--name`/`--description` flags
- Platform document canonicalization on import (e.g., OpenCode agent → canonical format)
- Canonical document validation before adding to library
- Update `library.yaml` with new resource entry
- Support for `--dry-run` preview and `--force` overwrite
- Library path discovery via `--library` flag, `GERMINATOR_LIBRARY` env, or default path

## Capabilities

### New Capabilities

- `library-resource-import`: Import existing canonical or platform documents into the library with automatic type detection, canonicalization, validation, and library.yaml registration

### Modified Capabilities

- `library-system`: Add capability to update library.yaml with new resource entries (new `AddResource` function)

## Impact

- New command: `cmd/library_add.go`
- New library infrastructure: `internal/infrastructure/library/adder.go`
- Dependencies: `application.Canonicalizer` (for platform→canonical), `serialization.MarshalCanonical` (for canonical output), `domain.Validate*` functions (for document validation)

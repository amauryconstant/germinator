## Why

The current `--discover` functionality in `germinator library add` lacks recursive directory scanning, a proper result structure for batch operations, and integration with `--batch` mode. This limits its usefulness for managing large libraries with nested directory structures.

## What Changes

- **Recursive directory scanning**: Discover scans directories recursively by default to find all `.md` files not registered in library.yaml
- **Enhanced result structure**: `DiscoverResult` with `OrphanInfo`, `AddSuccess`, `ConflictInfo`, and `DiscoverSummary` fields
- **Batch mode integration**: `--batch --force` discovers all orphans and adds them continuously (skipping errors)
- **Improved output**: Human-readable summary and JSON output via `--json` flag

## Capabilities

### New Capabilities
- `discover-orphans-batch`: Batch orphan discovery and registration with proper result aggregation. Creates `specs/discover-orphans-batch/spec.md`.

### Modified Capabilities
- `library-orphan-discovery`: Add recursive scanning requirement and enhanced result structure. Creates delta spec `specs/library/library-orphan-discovery/discover-orphans-delta.md`.

## Impact

- **Code**: `internal/infrastructure/library/` - DiscoverOrphans method and batch integration
- **CLI**: `cmd/` - New `--batch` and `--json` flags for `library add` command
- **Types**: `internal/domain/` - New result types for batch discover operations

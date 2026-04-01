## Why

When resources in the library are edited manually (e.g., changing a description in a text editor), the `library.yaml` index becomes stale. Users need a way to sync metadata from the actual resource files back into the library index. Additionally, when resource files are added directly to the library directory (e.g., via git mv or file manager), they become orphaned from the index.

## What Changes

- **New command**: `library refresh` - Syncs metadata from registered resource files into `library.yaml`
  - Updates `description` field from frontmatter
  - Updates `path` field when file is renamed (only if frontmatter name confirms same resource)
  - Skips files with conflicts (name mismatch) and continues
  - Does NOT remove entries for missing files (use `validate --fix`)

- **New flag**: `--discover` on `library add` - Finds orphaned resource files not in `library.yaml`
  - Scans `skills/`, `agents/`, `commands/`, `memory/` directories
  - Reports orphaned files with detected metadata
  - `--force` actually registers them; without `--force` only reports

## Capabilities

### New Capabilities

- `library-refresh`: Sync metadata from registered resource files into library.yaml
- `library-orphan-discovery`: Discover and optionally register orphaned resource files

### Modified Capabilities

- `library-resource-import`: Add `--discover` flag to enable orphan discovery mode

## Impact

- New files: `internal/infrastructure/library/refresher.go`, `cmd/library_refresh.go`
- Modified files: `cmd/library_add.go` (add --discover flag), `cmd/library_formatters.go` (output formatting)
- New specs: `specs/library/library-refresh/spec.md`, `specs/library/library-orphan-discovery/spec.md`

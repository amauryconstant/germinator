## Why

The library system stores resource metadata in `library.yaml` but provides no way to detect when that metadata becomes stale or inconsistent with the actual filesystem. Over time, entries can point to missing files, presets can reference non-existent resources, and orphaned files can accumulate. Users need a way to audit library health and repair it automatically.

## What Changes

- New `library validate` command that checks library integrity across four issue types:
  - **Missing file**: entry in `library.yaml` but file doesn't exist on disk
  - **Ghost resource**: preset references a resource that doesn't exist
  - **Orphaned file**: file exists on disk but isn't registered in `library.yaml`
  - **Malformed frontmatter**: resource file has invalid YAML frontmatter
- Severity tiers: `error` (missing, ghost, malformed) and `warning` (orphan)
- `--fix` flag to auto-cleanup `library.yaml` (removes missing entries, strips ghost preset refs)
- `--json` flag for machine-readable output
- Exit codes: `0` clean, `5` validation errors, `1` unexpected errors
- Conservative fix: only modifies `library.yaml`, never deletes actual files

## Capabilities

### New Capabilities
- `library-validation`: Detect and optionally fix inconsistencies between `library.yaml` metadata and the actual filesystem state

## Impact

- New command: `germinator library validate`
- New infrastructure: `internal/infrastructure/library/validator.go` for validation logic
- New command file: `cmd/library_validate.go`
- No changes to existing commands, services, or adapters

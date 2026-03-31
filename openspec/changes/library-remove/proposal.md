## Why

Users can add resources and presets to the library via `library add` and `library create preset`, but there is no inverse operation. Removing a resource or preset currently requires manually editing `library.yaml` and deleting the physical file. This creates an asymmetric, error-prone workflow.

## What Changes

- Add `library remove resource <ref>` command to remove a resource from the library
- Add `library remove preset <name>` command to remove a preset from the library
- Add `--json` flag for structured output (only on remove commands)
- When removing a resource:
  - Delete the physical file from the library directory
  - Remove the entry from `library.yaml`
  - Error if any preset references the resource (user must explicitly remove preset first)
- When removing a preset:
  - Remove the entry from `library.yaml`
  - No physical file to delete (presets exist only in YAML)
- No `--dry-run` or `--force` flags for simplicity

## Capabilities

### New Capabilities
- `library-remove-resource`: Remove a resource (skill, agent, command, memory) from the library, deleting both the physical file and YAML entry
- `library-remove-preset`: Remove a preset definition from the library YAML

### Modified Capabilities
<!-- No existing capability requirements are changing -->

## Impact

- New command file: `cmd/library_remove.go` (implements `library remove` command)
- New infrastructure file: `internal/infrastructure/library/remover.go`
- No changes to existing commands, services, or domain models
- No breaking changes to existing library structure or workflows

## Why

The library system currently supports reading and listing presets, but users cannot create new presets through the CLI. They must manually edit `library.yaml` to add presets. Adding a dedicated command for preset creation provides a user-friendly, validated workflow for managing presets.

## What Changes

- New CLI subcommand: `germinator library create preset <name>`
- New library infrastructure function: `SaveLibrary()` to persist library changes
- Strict validation: referenced resources must exist in the library before preset creation succeeds
- Overwrite protection with `--force` flag to replace existing presets

## Capabilities

### New Capabilities

- `library-preset-creation`: CLI command and infrastructure for creating presets in the library

### Modified Capabilities

<!-- No requirement changes to existing capabilities -->

## Impact

- New file: `internal/infrastructure/library/saver.go` - Library persistence
- New file: `cmd/library_create.go` - CLI command implementation
- Modified: `cmd/library.go` - Add `create` subcommand
- Modified: `cmd/completions.go` - Add completions for new command

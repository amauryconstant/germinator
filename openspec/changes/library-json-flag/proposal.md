## Why

The `germinator library` command lacks a consistent JSON output mode across its subcommands. Users integrating Germinator into scripts or automated workflows need machine-readable output. Currently, each subcommand handles output differently, making programmatic consumption inconsistent and error-prone.

## What Changes

- Add `--json` flag to the parent `germinator library` command in `cmd/library.go`
- Flag inherits to all library subcommands via Cobra's flag inheritance mechanism
- Subcommands check the flag and output JSON when enabled
- Consistent JSON output structures for all subcommands

## Capabilities

### New Capabilities

- `library-json-output`: Add JSON output flag to library parent command and implement JSON formatting for all library subcommands (resources, presets, add, init, remove, validate, refresh, show)

### Modified Capabilities

- `library-system`: Extend to support JSON-formatted output for list operations (resources, presets)
- `library-resource-import`: Extend BatchAddResult to support JSON serialization
- `library-scaffolding`: Extend init command to support JSON output
- `library-remove-resource`: Already has JSON support, ensure consistent behavior
- `library-validation`: Already has JSON support, ensure consistent behavior
- `library-refresh`: Already has JSON support, ensure consistent behavior

## Impact

- **Files Modified:**
  - `cmd/library.go` - Add `--json` flag to parent command
  - `cmd/library/*.go` - Subcommands check flag and output JSON appropriately
- **New Dependencies:** None
- **CLI Impact:** New `--json` flag available on all `germinator library` subcommands

## Why

Users need to import multiple resources to the library one at a time, which is inefficient when migrating existing files or adding entire directories of skills/agents. A batch mode would allow processing multiple files and directories in a single command with error resilience.

## What Changes

- New `--batch` flag on `germinator library add` command
- Accepts multiple positional arguments: files AND/OR directories
- Directories are scanned recursively for `.md` files
- Continues on error by default (processes all inputs, collects failures)
- Exit code always 0 for batch (partial success = success)
- Summary output at end: "Added N resources, skipped M, failed K"
- JSON output via `--json` flag on parent `library` command
- `--discover --batch` integration to discover orphans and add all of them

## Capabilities

### New Capabilities

- `library-batch-add`: Batch processing for library resource imports. Each batch operation processes multiple files/directories, collects results by category (added/skipped/failed), and returns a structured result enabling rich CLI output and scripting.

### Modified Capabilities

<!-- No existing capability requirements change - this is purely additive -->

## Impact

- **cmd/library/**: Add command modified to accept batch flag and multiple positional args
- **internal/application/library.go**: New BatchAddResult type and batch service method
- **internal/infrastructure/library/**: Library service updated with batch processing logic
- **CLI output**: Human-readable summary format and JSON serialization for batch results

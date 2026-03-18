## Why

Centralized error handling improves testability and provides consistent error handling across all CLI commands. The current pattern with per-command `HandleError()` calls makes testing difficult due to `os.Exit()` calls within command handlers. Migrating to the `RunE` pattern aligns germinator with the unified CLI standard established through investigation.

## What Changes

- Migrate CLI commands from `Run` + `HandleError()` to `RunE` pattern with centralized error handling
- Expand exit codes to cover all error categories (config, git, validation, not found)
- Update `HandleError()` to `HandleCLIError()` with updated signature for centralized handling
- All errors bubble up to `main.go` for consistent handling

## Capabilities

### New Capabilities

- `cli-error-handling`: Centralized RunE pattern with expanded exit codes and error categorization

## Impact

**Affected Code**:
- All 16 cmd Go files (Run → RunE pattern)
- `cmd/error_handler.go` (exit codes, categories, HandleCLIError)
- `main.go` (centralized error handling)

**Affected Directories**:
- `cmd/` - Error handling pattern change

**No Public API Impact**: CLI flags, arguments, and user-facing behavior remain unchanged.

## Why

Germinator uses basic `fmt.Errorf` wrapping with a single exit code (1) for all errors and has no verbosity control, making debugging difficult for users and scripts. This change establishes foundational CLI infrastructure—typed errors, semantic exit codes, composable error formatting, and verbosity flags—to support Germinator's evolution into a one-stop CLI for AI coding agent configuration management.

## What Changes

**Typed Error System (from typed-errors):**

- New domain error types in `internal/errors/types.go`: ParseError, ValidationError, TransformError, FileError, ConfigError
- Semantic exit codes: 0=success, 1=error, 2=usage, 3=parse
- Error categorization function mapping error types to exit codes
- Composable error formatter with type-specific hints in `cmd/error_formatter.go`
- Central error handler in `cmd/error_handler.go`
**Verbosity System (from verbose-output):**
- Verbosity type with helper methods (IsVerbose, IsVeryVerbose) in `cmd/verbose.go`
- Persistent `-v`/`-vv` flag on root command using Cobra's CountP
- Level 0 (default), Level 1 (-v), Level 2 (-vv)
- Verbose output to stderr with 2-space indentation for level 2 details

**New Integration:**
- `CommandConfig` struct in `cmd/config.go` holding ErrorFormatter and Verbosity
- Commands receive CommandConfig instead of using globals
- Foundation for future dependency injection expansion

## Capabilities

### New Capabilities

- `typed-errors`: Domain-specific error types with structured fields (path, message, cause) for parse, validation, transform, file, and config errors
- `exit-codes`: Semantic exit codes (0=success, 1=error, 2=usage, 3=parse) with error-to-exit-code mapping
- `error-formatting`: Composable error formatting with type-specific output and contextual hints
- `verbose-output`: Multi-level verbosity control for CLI commands with structured output formatting

### Modified Capabilities

- `cli-framework`: Commands will use typed errors, semantic exit codes, verbose output, and CommandConfig pattern instead of `fmt.Errorf`, `os.Exit(1)`, and globals
- `document-loading`: Will return typed `ParseError` and `FileError` instead of generic errors
- `document-transformation`: Will return typed `TransformError` and `ValidationError`

## Impact

**New Files:**

- `internal/errors/types.go` - Domain error types
- `internal/errors/types_test.go` - Error type tests
- `cmd/error_handler.go` - Exit codes and categorization
- `cmd/error_formatter.go` - Error formatting with hints
- `cmd/error_formatter_test.go` - Formatter tests
- `cmd/verbose.go` - Verbosity helper type and functions
- `cmd/config.go` - CommandConfig struct for dependency injection

**Modified Files:**
- `cmd/root.go` - Add persistent verbose flag, initialize CommandConfig
- `cmd/validate.go` - Use CommandConfig, HandleError, verbose output
- `cmd/adapt.go` - Use CommandConfig, HandleError, verbose output
- `cmd/canonicalize.go` - Use CommandConfig, HandleError, verbose output
- `internal/core/loader.go` - Return typed errors
- `internal/services/transformer.go` - Return typed errors
- `internal/services/canonicalizer.go` - Return typed errors
- `cmd/AGENTS.md` - Document verbose flag and CommandConfig patterns

**Backward Compatibility:**
- Error messages remain human-readable and helpful
- Scripts checking for non-zero exit will still work
- New exit codes provide more granularity for scripts that want it
- Default behavior (level 0 verbosity) unchanged

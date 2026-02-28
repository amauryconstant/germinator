## Context

Germinator currently uses basic `fmt.Errorf` wrapping with a single exit code (1) for all errors and has no verbosity control. This makes debugging difficult for:

- **Users** who need visibility into what the tool is doing
- **Scripts** that need to programmatically handle different error types
- **Debugging** to quickly identify the root cause

The reference implementation from twiggit demonstrates a superior pattern:

1. Domain-specific error types with structured fields
2. Semantic exit codes for different error categories
3. Composable error formatting with contextual hints

Additionally, Germinator is evolving into a one-stop CLI for AI coding agent configuration management. This requires a foundation for dependency injection to support future growth.

### Current State

```go
// cmd/validate.go - current approach
if validatePlatform == "" {
    fmt.Fprintf(os.Stderr, "Error: --platform flag is required\n")
    os.Exit(1)  // All errors exit 1
}
// internal/core/loader.go - current approach
return nil, fmt.Errorf("unrecognizable filename: %s", filepath)
```

### Constraints

- Must maintain backward compatibility for error messages (human-readable)
- Must not break existing scripts that check for non-zero exit
- Must integrate with existing Cobra CLI structure
- Must follow Go idioms (error wrapping, `errors.As`/`errors.Is`)
- Verbose output must go to stderr (stdout stays clean for piping)
- CommandConfig pattern must be extensible for future additions

## Goals / Non-Goals

**Goals:**

- Implement typed error types for germinator's domain (parse, validation, transform, file, config)
- Define semantic exit codes (0=success, 1=error, 2=usage, 3=parse)
- Create error categorization function for exit code mapping
- Build composable error formatter with type-specific hints
- Add persistent `-v`/`-vv` flags for verbosity control
- Introduce CommandConfig struct for dependency injection foundation
- Integrate with existing cmd/ commands
- Provide comprehensive tests

**Non-Goals:**
- Changing error message content significantly (maintain helpful messages)
- Adding new validation rules or error conditions
- Creating error codes for every possible error variant
- Implementing error telemetry or logging
- Log file output
- JSON/structured logging format
- More than 2 verbosity levels
- Full dependency injection framework (future change)

## Decisions

### Decision 1: Error Type Location

**Choice:** Place error types in `internal/errors/types.go`
**Rationale:**

- Follows Go convention of domain-specific packages
- Keeps errors close to where they're used (internal packages)
- Avoids circular dependencies with cmd/ and services/
- Allows reuse across core/, services/, and cmd/

**Alternatives considered:**
- `internal/domain/errors.go` - domain package doesn't exist; would require new package
- `cmd/errors.go` - would create cmd→internal dependency cycle
- `pkg/errors/` - germinator doesn't use pkg/ pattern

### Decision 2: Exit Code Values

**Choice:**

- `0` - Success
- `1` - General error (transform, file, unexpected)
- `2` - Usage/validation error (invalid flags, missing args, validation failures)
- `3` - Parse error (malformed YAML, unrecognized document type)

**Rationale:**

- Follows Unix conventions (0=success, 1=error, 2=usage)
- `3` for parse errors provides clear distinction for input problems
- Aligns with Cobra's built-in usage error handling (exit 2)
- Mirrors twiggit's approach for consistency

**Alternatives considered:**

- Only 0/1 - too coarse, no distinction for scripts
- 0/1/2 only - doesn't distinguish parse from validation errors
- HTTP-style codes (4xx, 5xx) - not idiomatic for CLI tools

### Decision 3: Error Type Structure

**Choice:** Struct types with `Error()`, `Unwrap()`, and helper methods

```go
type ParseError struct {
    Path    string
    Message string
    Cause   error
}
func (e *ParseError) Error() string { ... }
func (e *ParseError) Unwrap() error { return e.Cause }
```

**Rationale:**

- Structured fields enable rich formatting
- `Unwrap()` supports `errors.As`/`errors.Is` chains
- Helper methods (e.g., `IsFileNotFound()`) enable categorization
- Matches twiggit pattern for proven design

**Alternatives considered:**

- Simple string errors - no structured data for formatting
- Interface-based errors - more complex, less idiomatic for Go
- Error codes with maps - loses type safety

### Decision 4: Error Formatter Pattern

**Choice:** Registry-based formatter with type-specific functions

```go
type ErrorFormatter struct {
    formatters map[reflect.Type]func(error) string
}
func (f *ErrorFormatter) Format(err error) string {
    for errType, formatter := range f.formatters {
        if errors.As(err, &target) {
            return formatter(err)
        }
    }
    return defaultFormat(err)
}
```

**Rationale:**

- Composable and extensible
- Type-safe formatting per error type
- Easy to add new formatters without modifying existing code
- Supports hints and contextual suggestions

**Alternatives considered:**
- Switch on error type in one function - harder to extend
- Interface method on error types - couples formatting to domain
- Template-based formatting - overkill for this use case

### Decision 5: Verbosity Type Design

**Choice:** Create `Verbosity` type in `cmd/verbose.go` with helper methods.

```go
type Verbosity int
func (v Verbosity) IsVerbose() bool     { return v >= 1 }
func (v Verbosity) IsVeryVerbose() bool { return v >= 2 }
```

**Rationale:**

- Type-safe verbosity handling
- Semantic methods are more readable than integer comparisons
- Easy to extend with additional methods if needed

**Alternatives considered:**
- Plain `int` - Simpler but less expressive, error-prone
- External logging library - Overkill for CLI use case

### Decision 6: CommandConfig Pattern

**Choice:** Create `CommandConfig` struct holding shared command dependencies.

```go
type CommandConfig struct {
    ErrorFormatter *ErrorFormatter
    Verbosity      Verbosity
}
```

**Rationale:**

- Foundation for dependency injection
- Commands receive config instead of accessing globals
- Extensible for future additions (e.g., OutputWriter, Logger)
- Makes commands more testable

**Alternatives:**
- Global variables - Harder to test, implicit dependencies
- Full DI framework - Overkill for current needs
- Pass each dependency separately - Verbose, harder to extend

### Decision 7: Integration Strategy

**Choice:** Central error handler in cmd package with CommandConfig

```go
func HandleError(cfg *CommandConfig, err error) {
    fmt.Fprintln(os.Stderr, cfg.ErrorFormatter.Format(err))
    os.Exit(int(GetExitCodeForError(err)))
}
```

**Rationale:**

- Single point of control for all error handling
- Consistent formatting across all commands
- Easy to modify behavior globally
- Matches Cobra's Run function pattern

**Alternatives considered:**
- Each command handles errors independently - inconsistent
- Middleware pattern - Cobra doesn't support this well
- Panic/recover - non-idiomatic for expected errors

## Risks / Trade-offs

### Risk: Breaking existing scripts

**Mitigation:** Scripts checking `!= 0` will still work. Only scripts checking specific exit codes (rare) may break.

### Risk: Error types proliferate

**Mitigation:** Start with 5 types (Parse, Validation, Transform, File, Config). Add more only when clearly needed.

### Risk: Formatter becomes complex

**Mitigation:** Keep formatters simple with single responsibility. Use helper functions for shared logic.

### Risk: Verbose output could get stale

**Mitigation:** E2E tests verify output matches behavior.

### Trade-off: More boilerplate for error creation

**Acceptance:** Small price for better error handling. Helper constructors (`NewParseError(...)`) reduce verbosity.

### Trade-off: Learning curve for new pattern

**Acceptance:** Well-documented with examples. Follows Go and twiggit conventions.

### Trade-off: CommandConfig adds indirection

**Acceptance:** Foundation for future DI expansion. Benefits outweigh minimal complexity.

## Open Questions

1. **Should we add error codes within types?** (e.g., `ParseError{Code: "E001"}`)
   - **Resolution:** No. Exit codes provide enough categorization. Error codes add complexity without clear benefit.
2. **Should errors include stack traces?**
   - **Resolution:** No. Not needed for CLI tool. File/line context from error messages is sufficient.
3. **Should we translate error messages?**
   - **Resolution:** No. English-only is acceptable for developer tool.
4. **Should verbosity affect error output detail?**
   - **Resolution:** Not in this change. Error formatting is independent of verbosity. Future enhancement if needed.

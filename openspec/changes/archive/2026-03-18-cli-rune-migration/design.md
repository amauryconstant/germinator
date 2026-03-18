## Context

**Current State:**
- CLI uses `Run` + `HandleError()` pattern with `os.Exit()` in each command
- Each command calls `HandleError(cfg, err)` which terminates the process
- Testing is difficult due to `os.Exit()` calls in command handlers
- Limited exit codes (Success, Error, Usage)

**Target State:**
- CLI uses `RunE` pattern with centralized error handling in `main.go`
- Commands return errors instead of calling `os.Exit()`
- Errors bubble up to a central handler in `main.go`
- Expanded exit codes for all error categories

**Constraints:**
- All changes are to cmd package - no business logic changes
- Existing E2E tests must continue to pass with correct exit codes
- User-facing behavior must remain unchanged

## Goals / Non-Goals

**Goals:**
- Migrate CLI to RunE pattern with centralized error handling
- Expand exit codes to cover all error categories
- Improve testability by removing `os.Exit()` from command handlers
- Maintain all existing functionality and E2E test behavior

**Non-Goals:**
- No changes to business logic or service implementations
- No changes to CLI flags, arguments, or user-facing behavior
- No changes to error formatting (ErrorFormatter)

## Decisions

### DEC-001: Exit Code Mapping

**Choice:** Expanded exit codes with clear category mapping

**Current → Target Mapping:**
| Category | Old Exit Code | New Exit Code |
|----------|---------------|---------------|
| Success | 0 | 0 (unchanged) |
| Generic Error | 1 | 1 (unchanged) |
| Usage/Cobra | 2 | 2 (unchanged) |
| Config | 2 (Usage) | 3 (Config) |
| Git | N/A | 4 (Git) - NEW |
| Validation | 2 (Usage) | 5 (Validation) |
| NotFound | N/A | 6 (NotFound) - NEW |
| Parse | 3 | 3 (merged into Config) |

**Note:** `ExitCodeParse` (3) is being renamed to `ExitCodeConfig`. Parse errors will map to Config exit code.

### DEC-002: CLI Error Handling Pattern

**Choice:** RunE with centralized error handling in `main.go`

**Note:** `HandleValidationErrors()` function will be removed - validation errors will be wrapped and returned through the standard error path to `HandleCLIError()`.

**Current Pattern:**
```go
Run: func(c *cobra.Command, args []string) {
    result, err := service.Do(...)
    if err != nil {
        HandleError(cfg, err)  // Calls os.Exit()
    }
}
```

**Target Pattern:**
```go
RunE: func(c *cobra.Command, args []string) error {
    result, err := service.Do(...)
    if err != nil {
        return err  // Bubbles to main.go
    }
    return nil
}

// main.go
if err := rootCmd.Execute(); err != nil {
    exitCode := HandleCLIError(rootCmd, err)
    os.Exit(int(exitCode))
}
```

**Rationale:** RunE pattern allows errors to bubble up to a central handler. More testable (no os.Exit in command handlers) and provides consistent error handling.

**Alternatives Considered:**
- Keep current Run + HandleError pattern → Rejected: harder to test, inconsistent with standard

### DEC-003: HandleValidationErrors Removal

**Choice:** Remove `HandleValidationErrors()` and route all errors through `HandleCLIError()`

**Rationale:** Having two error handling paths creates inconsistency. All errors should flow through a single centralized handler for consistent formatting and exit code mapping.

**Migration:**
- Code calling `HandleValidationErrors(cfg, errs)` → wrap errors and return through RunE
- main.go handles all errors via `HandleCLIError(rootCmd, err)`

## Risks / Trade-offs

### Risk: RunE Migration Introduces Error Handling Bugs
**Impact:** Commands might not exit with correct codes
**Mitigation:** E2E tests verify exit codes; test each command manually after migration

### Risk: Exit Code Correctness
**Impact:** New exit codes might not map correctly to error types
**Mitigation:** Comprehensive error categorization function with unit tests

### Trade-off: All Commands Need Updates
**Impact:** All 16 cmd Go files need changes
**Mitigation:** Systematic migration with verification after each command

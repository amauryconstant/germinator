## Why

Error types in `internal/errors/` have inconsistent APIs. Change 2 (`add-validation-pipeline`) introduces immutable builders (`WithSuggestions()`, `WithContext()`) and private fields for `ValidationError`, but other error types (`ParseError`, `TransformError`, `FileError`, `ConfigError`) remain with public fields and no builders. This creates an inconsistent developer experience and prevents rich error context across the codebase.

## What Changes

- Enhance all 4 error types (`ParseError`, `TransformError`, `FileError`, `ConfigError`) with the same immutable builder pattern as `ValidationError` (Change 2)
- Make all fields private across all error types
- Add getter methods for all fields (e.g., `Path()`, `Message()`, `Cause()`, `Suggestions()`, `Context()`)
- Add `WithSuggestions([]string) *ErrorType` builder to all error types (returns new instance, immutable)
- Add `WithContext(string) *ErrorType` builder to all error types (returns new instance, immutable)
- Add new `suggestions` and `context` fields to all error types
- **BREAKING**: Change `ConfigError` constructor signature from `NewConfigError(field, value string, available []string, message string)` to `NewConfigError(field, value, message string)` (move `available` to `WithSuggestions()` builder)
- Rename `ConfigError.Available` field to `suggestions` for API consistency
- Update `error_formatter.go` to use getters instead of direct field access
- Update Error() methods to include context and suggestions in output

## Capabilities

### New Capabilities

- `enhanced-errors`: Apply immutable builder pattern (private fields, getters, `WithSuggestions()`, `WithContext()`) to all error types (ParseError, TransformError, FileError, ConfigError) for complete API alignment with ValidationError

### Modified Capabilities

None - this is a new capability extending the pattern established in Change 2.

## Impact

**New Fields:**
- All error types: `suggestions []string`, `context string`

**Modified Files:**
- `internal/errors/types.go` - Private fields, getters, builders for all 4 error types
- `cmd/error_formatter.go` - Use getters instead of direct field access, add context/suggestions to output
- `internal/core/loader.go` - Update ParseError (2), ConfigError (2) usage
- `internal/core/parser.go` - Update FileError usage (1 call site)
- `internal/services/transformer.go` - Update ParseError (3), TransformError (1), FileError (1), ConfigError (2) usage
- `internal/services/canonicalizer.go` - Update ParseError (2), TransformError (1), FileError (1) usage
- `internal/services/initializer.go` - Update FileError usage (2 call sites)
- `internal/config/config.go` - Update TransformError (1), ConfigError (2) usage
- `internal/config/manager.go` - Update ParseError (1), FileError (1) usage
- `internal/library/library.go` - Update ParseError usage (1 call site)
- `cmd/validate.go` - Update ConfigError usage (1 call site)
- `cmd/adapt.go` - Update ConfigError usage (1 call site)
- `cmd/canonicalize.go` - Update ConfigError usage (4 call sites)
- `cmd/init.go` - Update ConfigError usage (4 call sites)
- Test files - Update all error type usage (56 call sites)

**Total Call Sites:** ~90 across production (34) and test (56) code

**Breaking Changes:**
- `ConfigError` constructor signature change (26 call sites: 16 production, 10 test)

**Dependencies:**
- Must run AFTER Change 2 (`add-validation-pipeline`) - references ValidationError pattern
- No other dependencies

## Context

After Change 2 (`add-validation-pipeline`), `ValidationError` has an immutable builder pattern:
- Private fields with getter methods
- `WithSuggestions()` and `WithContext()` builders that return new instances
- Constructor takes only core fields

However, 4 other error types still use public fields and lack builders:
- `ParseError` - public Path, Message, Cause fields
- `TransformError` - public Operation, Platform, Message, Cause fields
- `FileError` - public Path, Operation, Message, Cause fields
- `ConfigError` - public Field, Value, Available, Message fields

This creates API inconsistency and prevents rich error context across the codebase. Developers using `ValidationError` get fluent construction and immutability, but other error types don't.

Current usage patterns:
- ~90 call sites across production (34) and test (56) code
- Most call sites use constructors, not direct field access
- `error_formatter.go` accesses fields directly (needs update to use getters)

Reference implementation: twiggit's `ValidationError` pattern (Change 2 already adopts this).

## Goals / Non-Goals

**Goals:**
- Apply identical immutable builder pattern to all 4 error types (ParseError, TransformError, FileError, ConfigError)
- Make all error fields private with getter methods
- Add `WithSuggestions()` and `WithContext()` builders to all types
- Add `suggestions` and `context` fields to all types
- Update all call sites to use getters
- Achieve 100% API consistency across all error types

**Non-Goals:**
- Changing error semantics or error types themselves
- Adding new validation rules or error types
- Changing CLI behavior or error output format (beyond adding context/suggestions)
- Supporting both old and new APIs (clean migration)

## Decisions

### Decision 1: Constructor signature changes

**Choice:** Change `ConfigError` constructor signature to remove `available` parameter. Keep other constructors unchanged.

```go
// ConfigError - BREAKING CHANGE
// Before: NewConfigError(field, value string, available []string, message string)
// After:  NewConfigError(field, value, message string) *ConfigError

// Other types - no change
NewParseError(path, message string, cause error) *ParseError
NewTransformError(operation, platform, message string, cause error) *TransformError
NewFileError(path, operation, message string, cause error) *FileError
```

**Rationale:**
- Aligns with ValidationError pattern: constructor takes core identity fields, optional fields via builders
- "Available" is enrichment (valid options), not core error identity
- Other error types already have minimal constructors (no changes needed)

**Alternatives considered:**
- Keep `available` in ConfigError constructor - breaks pattern consistency
- Change all constructors to add suggestions/context params - unnecessary, not all errors have optional fields

### Decision 2: All error types get both builders

**Choice:** All 4 error types get `WithSuggestions()` and `WithContext()` builders, even TransformError (where suggestions are rarely used).

**Rationale:**
- Complete API alignment means identical builder methods everywhere
- Future-proof: even if TransformError rarely uses suggestions, API is consistent
- Developer experience: same pattern for all error types

**Alternatives considered:**
- Only add WithSuggestions() to ParseError, FileError, ConfigError - breaks consistency
- TransformError only gets WithContext() - creates API divergence

### Decision 3: Rename ConfigError.Available to suggestions

**Choice:** Rename `ConfigError.Available` field to `suggestions` for API consistency.

**Rationale:**
- "Complete alignment" prioritizes consistency over semantic nuance
- All error types use `Suggestions()` getter and `WithSuggestions()` builder
- Loses semantic distinction (valid options vs hints), but gains API uniformity

**Alternatives considered:**
- Keep field name as `available`, add separate `suggestions` field - confusing, two similar fields
- Use `WithAvailable()` builder instead of `WithSuggestions()` - breaks pattern consistency

### Decision 4: Implementation order by complexity

**Choice:** Implement error types in order: ParseError → TransformError → FileError → ConfigError

**Rationale:**
- ParseError, TransformError, FileError: No breaking changes, moderate call sites (29, 18, 17)
- ConfigError: Breaking constructor change, most call sites (26)
- Build confidence with simple changes before complex breaking change

**Alternatives considered:**
- ConfigError first (largest impact) - high risk if pattern needs adjustment
- Random order - no clear migration path

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| ConfigError constructor breaking change affects 26 call sites | Grep for all `NewConfigError` calls, update systematically. No compatibility constraints = clean migration. |
| Direct field access in error_formatter.go breaks | Update all formatters to use getters. Test each formatter after change. |
| Tests rely on direct field access | Most tests use constructors, not fields. Minimal test changes needed. |
| Immutability adds allocation overhead | Negligible for error types (created rarely, not in hot paths). Benefits outweigh cost. |
| Suggestions rarely used on TransformError | Accept as consistency tax. API uniformity more valuable than micro-optimization. |

## Migration Plan

**Prerequisites:**
- Change 2 (`add-validation-pipeline`) must be complete
- ValidationError pattern established and tested

**Phase 1: ParseError (29 call sites: 9 production, 20 test)**
1. Make fields private: `path`, `message`, `cause`
2. Add new fields: `suggestions []string`, `context string`
3. Add getters: `Path()`, `Message()`, `Cause()`, `Suggestions()`, `Context()`
4. Add builders: `WithSuggestions()`, `WithContext()`
5. Update `Error()` method to use getters
6. Update production call sites:
   - `internal/core/loader.go` (2 sites)
   - `internal/services/transformer.go` (3 sites)
   - `internal/services/canonicalizer.go` (2 sites)
   - `internal/config/manager.go` (1 site)
   - `internal/library/library.go` (1 site)
7. Update test call sites (20 sites)
8. Update `error_formatter.go`
9. Test

**Phase 2: TransformError (18 call sites: 3 production, 15 test)**
1. Make fields private: `operation`, `platform`, `message`, `cause`
2. Add new fields: `suggestions []string`, `context string`
3. Add getters: `Operation()`, `Platform()`, `Message()`, `Cause()`, `Suggestions()`, `Context()`
4. Add builders: `WithSuggestions()`, `WithContext()`
5. Update `Error()` method to use getters
6. Update production call sites:
   - `internal/services/transformer.go` (1 site)
   - `internal/services/canonicalizer.go` (1 site)
   - `internal/config/config.go` (1 site)
7. Update test call sites (15 sites)
8. Update `error_formatter.go`
9. Test

**Phase 3: FileError (17 call sites: 6 production, 11 test)**
1. Make fields private: `path`, `operation`, `message`, `cause`
2. Add new fields: `suggestions []string`, `context string`
3. Add getters: `Path()`, `Operation()`, `Message()`, `Cause()`, `Suggestions()`, `Context()`
4. Add builders: `WithSuggestions()`, `WithContext()`
5. Update `Error()` method to use getters
6. Update production call sites:
   - `internal/config/manager.go` (1 site)
   - `internal/core/parser.go` (1 site)
   - `internal/services/initializer.go` (2 sites)
   - `internal/services/transformer.go` (1 site)
   - `internal/services/canonicalizer.go` (1 site)
7. Update test call sites (11 sites)
8. Update `error_formatter.go`
9. Test

**Phase 4: ConfigError (26 call sites: 16 production, 10 test) - BREAKING**
1. Make fields private: `field`, `value`, `message`
2. Rename `Available` → `suggestions` (private)
3. Add new field: `context string`
4. Change constructor: `NewConfigError(field, value, message string)` (remove `available`)
5. Add getters: `Field()`, `Value()`, `Message()`, `Suggestions()`, `Context()`
6. Add builders: `WithSuggestions()`, `WithContext()`
7. Update `Error()` method to use getters
8. Update production call sites:
   - `cmd/init.go` (4 sites)
   - `cmd/canonicalize.go` (4 sites)
   - `cmd/adapt.go` (1 site)
   - `cmd/validate.go` (1 site)
   - `internal/config/config.go` (2 sites)
   - `internal/services/transformer.go` (2 sites)
   - `internal/core/loader.go` (2 sites)
9. Update test call sites (10 sites)
10. Update `error_formatter.go`
11. Test

**Rollback:** Revert commits. No data migration, pure refactoring.

## Open Questions

None - all decisions resolved during exploration.

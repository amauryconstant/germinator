## Context

Germinator's validation uses `[]error` returns from model methods (`agent.Validate() []error`). This approach:
- Cannot short-circuit on first error
- Scatters validation logic across models (`internal/models/canonical/`) and services (`internal/services/`)
- Makes composition difficult (can't combine validators)
- Doesn't integrate well with the Result[T] pattern used in twiggit

After Change 1 (`introduce-service-interfaces`), the `Validator` interface exists with:
```go
Validate(ctx context.Context, req *ValidateRequest) (*ValidateResult, error)
```

This change introduces a functional validation pipeline behind that interface.

## Goals / Non-Goals

**Goals:**
- Introduce `Result[T]` type for functional error handling (success/failure without exceptions)
- Create `ValidationPipeline[T]` that chains validators with early exit
- Implement standalone validator functions (composable, pure, testable)
- Enhance `ValidationError` with immutable builders (`WithSuggestions()`, `WithContext()`)
- Organize validators by bounded context (generic vs platform-specific)
- Replace model `Validate()` methods with standalone functions

**Non-Goals:**
- Changing `Validator` interface or `ValidateResult` struct (done in Change 1)
- Adding new validation rules (migrate existing rules only)
- Changing CLI behavior or error output format
- Supporting both `[]error` and `Result[T]` styles long-term (clean migration)

## Decisions

### Decision 1: Package structure for Result[T] and ValidationPipeline

**Choice:** Create `internal/validation/` package containing:
- `result.go` - `Result[T]`, `NewResult[T]()`, `NewErrorResult[T]()`
- `pipeline.go` - `ValidationFunc[T]`, `ValidationPipeline[T]`
- `validators.go` - Generic validators (ValidateAgentName, etc.)
- `opencode/validators.go` - OpenCode-specific validators

**Rationale:**
- Clean separation from error types (`internal/errors/`)
- No circular dependencies (services imports validation, not vice versa)
- Matches twiggit's pattern (domain/validation.go)
- Single import for validation utilities

**Alternatives considered:**
- `internal/errors/` - Conflates error representation with validation logic
- `internal/application/` - Mixes service contracts with implementation utilities
- `internal/domain/` - Over-engineering for current codebase size

### Decision 2: ValidationError replacement with immutable builders

**Choice:** Replace `internal/errors/types.go` ValidationError with twiggit's pattern:
```go
type ValidationError struct {
    request, field, value, message string  // private
    suggestions []string
    context     string
}

func NewValidationError(request, field, value, message string) *ValidationError
func (e *ValidationError) WithSuggestions(s []string) *ValidationError  // returns copy
func (e *ValidationError) WithContext(c string) *ValidationError        // returns copy
func (e *ValidationError) Field() string      // getter
func (e *ValidationError) Value() string      // getter
func (e *ValidationError) Message() string    // getter
```

**Rationale:**
- Immutable builders enable fluent construction
- Private fields ensure validation errors are constructed via builders
- Getters allow reading without mutation
- Matches twiggit's proven pattern

**Alternatives considered:**
- Keep current ValidationError - lacks builders, can't express context
- Create new type in validation/ - two error types is confusing

### Decision 3: Standalone validator functions vs model methods

**Choice:** Remove `Validate() []error` from models. Create standalone functions:
```go
// internal/validation/validators.go
func ValidateAgentName(a *canonical.Agent) Result[bool]
func ValidateAgentDescription(a *canonical.Agent) Result[bool]
func ValidateAgent(a *canonical.Agent) Result[bool] {
    return NewValidationPipeline(
        ValidateAgentName,
        ValidateAgentDescription,
        ValidateAgentPermissionPolicy,
    ).Validate(a)
}
```

**Rationale:**
- Pure functions are composable
- Each validator testable in isolation
- Pipeline composition is explicit
- Model doesn't know about validation strategy

**Alternatives considered:**
- Keep model methods calling pipelines - indirection, two places to look
- Keep current []error approach - no composition, no early exit

### Decision 4: Platform-specific validators in subpackages

**Choice:** Platform validators in `internal/validation/opencode/`:
```go
// internal/validation/opencode/validators.go
func ValidateAgentMode(a *canonical.Agent) Result[bool]
func ValidateAgentTemperature(a *canonical.Agent) Result[bool]
func ValidateAgentOpenCode(a *canonical.Agent) Result[bool] {
    return NewValidationPipeline(
        ValidateAgentMode,
        ValidateAgentTemperature,
    ).Validate(a)
}
```

Service composes generic + platform:
```go
// internal/services/validator.go
func (v *validator) Validate(ctx, req) (*ValidateResult, error) {
    // Generic invariants
    if r := validation.ValidateAgent(doc); r.IsError() {
        return &ValidateResult{Errors: []error{r.Error}}, nil
    }
    // Platform context
    if req.Platform == "opencode" {
        if r := opencode.ValidateAgentOpenCode(doc); r.IsError() {
            return &ValidateResult{Errors: []error{r.Error}}, nil
        }
    }
    return &ValidateResult{}, nil
}
```

**Rationale:**
- DDD bounded contexts: generic invariants vs platform rules
- Validators remain pure (no platform parameter)
- Service is composition root
- Easy to add new platforms

**Alternatives considered:**
- Closure factories - adds factory layer, less clear
- Request structs as input - mixes validation with request types

### Decision 5: Result[T] type design

**Choice:** Simple struct with helper constructors:
```go
type Result[T any] struct {
    Value T
    Error error
}

func NewResult[T any](value T) Result[T]
func NewErrorResult[T any](err error) Result[T]
func (r Result[T]) IsSuccess() bool
func (r Result[T]) IsError() bool
```

**Rationale:**
- Minimal API surface
- Type-safe (can't access Value when Error != nil if caller checks)
- Matches twiggit exactly
- Works with ValidationFunc[T]

**Alternatives considered:**
- More methods (Map, FlatMap) - YAGNI for validation use case

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking change to ValidationError | All usages in tests/services need update; grep for `NewValidationError` |
| Removing model Validate() methods | All callers now use validation.ValidateAgent() etc. |
| Platform-specific validators duplicated | Share validators where rules overlap; document differences |
| Result[T] unused outside validation | Accept as hygiene; pattern can spread if useful |
| Migration complexity | Run AFTER Change 1 lands; validator interface already exists |

## Migration Plan

**Prerequisites:**
- Change 1 (`introduce-service-interfaces`) must be complete and merged
- `Validator` interface exists in `internal/application/`
- ServiceContainer has `Validator` field wired

**Phase 1: Add validation package (no behavior change)**
1. Create `internal/validation/result.go` with Result[T]
2. Create `internal/validation/pipeline.go` with ValidationFunc, ValidationPipeline
3. Create `internal/validation/validators.go` with generic validators
4. Create `internal/validation/opencode/validators.go` with platform validators
5. Replace `internal/errors/types.go` ValidationError with enhanced version
6. Add tests for all new types

**Phase 2: Wire validators into service**
1. Update `internal/services/validator.go` to use pipelines
2. Remove platform-specific functions (`validateOpenCodeAgent`, etc.)
3. Keep model `Validate()` methods for now (call pipelines internally)

**Phase 3: Remove model methods**
1. Remove `Validate()` methods from all canonical models
2. Update any remaining callers to use `validation.ValidateAgent()` etc.

**Phase 4: Cleanup**
1. Remove old validation helper functions from services
2. Verify all tests pass

**Rollback:** Revert to Phase 1 state if issues—old code still works.

## Open Questions

None - all decisions resolved during exploration.

## Why

Current validation uses `[]error` returns that don't compose, can't short-circuit, and scatter validation logic across models and services. The `internal/errors/validation.go` ValidationError lacks immutable builders for fluent error construction. This change introduces a functional validation pipeline with `Result[T]` pattern (from twiggit) to enable composable, early-exit validation with clean error handling.

## What Changes

- **BREAKING**: Add `internal/validation/` package with `Result[T]` type and `ValidationPipeline[T]`
- **BREAKING**: Replace `internal/errors/types.go` ValidationError with twiggit's pattern (immutable `WithSuggestions()`, `WithContext()` builders, private fields, getter methods)
- **BREAKING**: Remove `Validate() []error` methods from canonical models (`internal/models/canonical/models.go`)
- Add standalone validator functions in `internal/validation/validators.go` (e.g., `ValidateAgentName`, `ValidateAgentDescription`)
- Add platform-specific validators in `internal/validation/opencode/validators.go`
- Update `internal/services/validator.go` to compose pipelines instead of calling model methods

## Capabilities

### New Capabilities

- `result-type`: Generic `Result[T]` type for functional error handling (success/failure without exceptions)
- `validation-pipeline`: Composable `ValidationPipeline[T]` that chains `ValidationFunc[T]` with early exit
- `composable-validators`: Standalone validator functions for Agent, Command, Skill, Memory (generic) + OpenCode-specific (platform)
- `enhanced-validation-errors`: ValidationError with immutable builders, private fields, getter methods

### Modified Capabilities

- `service-contracts`: ValidateResult.Errors type changes from `[]error` to use Result[T] internally (interface unchanged)

## Impact

**New Package**:
- `internal/validation/` - Result[T], ValidationPipeline, validators

**Modified Packages**:
- `internal/errors/types.go` - Replace ValidationError with enhanced version
- `internal/models/canonical/models.go` - Remove all Validate() methods
- `internal/services/transformer.go` - Use validation pipelines instead of model.Validate()
- `internal/services/canonicalizer.go` - Use validation pipelines

**Dependent Changes**:
- Must run AFTER `introduce-service-interfaces` (Change 1) - validator interface already exists
- Service interface unchanged (ValidateResult still has `Errors []error`, `Valid() bool`)

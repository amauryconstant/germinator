## 1. Create Validation Package (Phase 1a)

- [x] 1.1 Create `internal/validation/` directory
- [x] 1.2 Create `internal/validation/result.go` with `Result[T any]` struct, `NewResult[T]()`, `NewErrorResult[T]()`, `IsSuccess()`, `IsError()` methods
- [x] 1.3 Create `internal/validation/result_test.go` with tests for all Result[T] functionality
- [x] 1.4 Create `internal/validation/pipeline.go` with `ValidationFunc[T any]` type, `ValidationPipeline[T]` struct, `NewValidationPipeline[T]()`, `Validate()` method
- [x] 1.5 Create `internal/validation/pipeline_test.go` with tests for pipeline composition and early exit
- [x] 1.6 Create `internal/validation/validators.go` with generic validators:
  - `ValidateAgentName()`, `ValidateAgentDescription()`, `ValidateAgentPermissionPolicy()`, `ValidateAgent()`
  - `ValidateCommandName()`, `ValidateCommandDescription()`, `ValidateCommandExecution()`, `ValidateCommand()`
  - `ValidateSkillName()`, `ValidateSkillDescription()`, `ValidateSkillExecution()`, `ValidateSkill()`
  - `ValidateMemory()`
- [x] 1.7 Create `internal/validation/validators_test.go` with tests for all validators
- [x] 1.8 Create `internal/validation/opencode/` directory
- [x] 1.9 Create `internal/validation/opencode/validators.go` with OpenCode-specific validators:
  - `ValidateAgentMode()`, `ValidateAgentTemperature()`, `ValidateAgentOpenCode()`
  - `ValidateCommandOpenCode()`, `ValidateSkillOpenCode()`
- [x] 1.10 Create `internal/validation/opencode/validators_test.go` with tests for OpenCode validators

## 2. Replace ValidationError (Phase 1b)

- [x] 2.1 Replace `internal/errors/types.go` ValidationError with new implementation (private fields: request, field, value, message, suggestions, context)
- [x] 2.2 Update `NewValidationError()` signature to `(request, field, value, message string)`
- [x] 2.3 Add `WithSuggestions()`, `WithContext()` immutable builder methods
- [x] 2.4 Add getter methods: `Field()`, `Value()`, `Message()`, `Request()`, `Suggestions()`, `Context()`
- [x] 2.5 Update `Error()` method to include suggestions with 💡 prefix
- [x] 2.6 Update `internal/errors/types_test.go` with tests for new ValidationError
- [x] 2.7 Update all callers of old `NewValidationError(message, field, suggestions)` signature:
  - `cmd/cmd_test.go`: 4 usages - add request context (e.g., "Agent", "Command")
  - `cmd/error_formatter_test.go`: 6 usages - add request context and value
  - `internal/services/canonicalizer.go`: 1 usage - add request context
  - `internal/core/loader.go`: 1 usage - add request context

## 3. Wire Validators into Service (Phase 2)

- [x] 3.1 Update `internal/services/validator.go` to import `internal/validation` and `internal/validation/opencode`
- [x] 3.2 Update `validator.Validate()` method to call `validation.ValidateAgent()` etc. instead of `doc.Validate()`
- [x] 3.3 Add platform-specific validation: if platform == "opencode", call `opencode.ValidateAgentOpenCode()` etc.
- [x] 3.4 Convert `Result[bool]` to `[]error` for `ValidateResult.Errors` field
- [x] 3.5 Remove `validateOpenCodeAgent()` function from transformer.go
- [ ] 3.6 Remove `validatePlatform()` function from transformer.go (move to validation package if needed)
- [ ] 3.7 Update `internal/services/canonicalizer.go` to use validation pipelines in `validateCanonicalDoc()`

## 4. Remove Model Validate Methods (Phase 3)

- [ ] 4.1 Remove `Validate() []error` method from `internal/models/canonical/models.go` Agent struct
- [ ] 4.2 Remove `Validate() []error` method from `internal/models/canonical/models.go` AgentBehavior struct
- [ ] 4.3 Remove `Validate() []error` method from `internal/models/canonical/models.go` Command struct
- [ ] 4.4 Remove `Validate() []error` method from `internal/models/canonical/models.go` CommandExecution struct
- [ ] 4.5 Remove `Validate() []error` method from `internal/models/canonical/models.go` Memory struct
- [ ] 4.6 Remove `Validate() []error` method from `internal/models/canonical/models.go` Skill struct
- [ ] 4.7 Remove `Validate() []error` method from `internal/models/canonical/models.go` SkillExecution struct
- [ ] 4.8 Remove `Validate() []error` method from `internal/models/canonical/models.go` SkillExtensions struct
- [ ] 4.9 Remove `Validate() []error` method from `internal/models/canonical/models.go` AgentExtensions struct
- [ ] 4.10 Update `internal/models/canonical/models_test.go` to use `validation.ValidateAgent()` etc. instead of model methods

## 5. Cleanup (Phase 4)

- [ ] 5.1 Remove any remaining validation helper functions from services package
- [ ] 5.2 Ensure all imports are correct after removing model methods

## 6. Verification

- [ ] 6.1 Run `mise run check` (lint, format, test, build)
- [ ] 6.2 Verify all existing tests still pass
- [ ] 6.3 Verify CLI commands work as expected (manual smoke test: `germinator validate`, `germinator adapt`)
- [ ] 6.4 Verify error messages still display correctly with suggestions

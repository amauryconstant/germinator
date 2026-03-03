## Context

Germinator's `internal/services/` package exposes four operations as package-level functions:
- `TransformDocument(inputPath, outputPath, platform string) error`
- `ValidateDocument(inputPath, platform string) ([]error, error)`
- `CanonicalizeDocument(inputPath, outputPath, platform, docType string) error`
- `InitializeResources(opts InitOptions, refs []string) ([]InitResult, error)`

CLI commands call these directly (e.g., `services.TransformDocument(...)`), creating tight coupling. The `cmd/container.go` ServiceContainer is empty—no actual wiring occurs.

## Goals / Non-Goals

**Goals:**
- Introduce service interfaces enabling dependency injection and testability
- Create request/result types separating service contracts from domain models
- Wire ServiceContainer in main.go with concrete implementations
- Maintain backward compatibility during migration (tests continue passing)

**Non-Goals:**
- Refactoring core package (loader, parser, serializer)
- Adding new service operations
- Changing CLI command signatures or user-facing behavior
- Introducing a DI framework (wire manually)

## Decisions

### Decision 1: Package structure for interfaces and DTOs

**Choice:** Create `internal/application/` containing interfaces, request types, and result types together.

**Rationale:**
- Avoids circular imports: `services/` can import `application/`, but not vice versa
- Keeps related contracts in one place (interfaces + their request/result types)
- Single import for service users

**Alternative considered:** Some codebases place DTOs in `internal/domain/` separate from interfaces. We place them in `application/` alongside interfaces for tighter cohesion (one import, related types together).

**Alternatives considered:**
- `internal/contracts/` or `internal/dto/` - over-engineering for project size
- `internal/services/` for DTOs - creates circular dependency with interfaces in `application/`

### Decision 2: Include `context.Context` in all service methods

**Choice:** Add `context.Context` as first parameter to all interface methods.

**Rationale:**
- Idiomatic Go for service-layer interfaces
- Low cost to add now, expensive to retrofit later
- Enables future: cancellation, timeouts, tracing, request-scoped values
- Consistent with idiomatic Go service-layer patterns

**Alternatives considered:**
- Skip context for now - saves nothing, creates inconsistency

### Decision 3: Include `Refs` in `InitializeRequest`

**Choice:** Add `Refs []string` field to `InitializeRequest`, deprecate separate parameter.

**Rationale:**
- `Refs` is input data (what to process), not configuration (how to process)
- Consistent with other requests having all input in the struct
- Preset resolution stays in command layer (not service concern)

**Alternatives considered:**
- Keep `Refs` as separate parameter - inconsistent with other services
- Add `Preset` field to request - service shouldn't know about presets

### Decision 4: Constructors in services package

**Choice:** `services.NewTransformer()`, `services.NewValidator()`, etc. in `internal/services/`.

**Rationale:**
- Must be in services/ to avoid circular imports
- Constructor returns concrete type, caller assigns to interface

### Decision 5: ValidateResult with `Valid()` method

**Choice:** `ValidateResult{ Errors []error }` with `Valid() bool` method.

**Rationale:**
- Dual-return pattern preserved: `(*ValidateResult, error)`
- `error` return = fatal (couldn't validate)
- `result.Errors` = validation issues (business outcome)
- Computed `Valid()` is more Go-idiomatic than stored boolean

### Decision 6: Remove `InitializeFromPreset` as service method

**Choice:** Preset resolution happens in command layer, not service.

**Rationale:**
- Preset resolution is trivial map lookup
- Command already handles "refs vs preset" CLI logic
- Service stays focused on initialization, not preset resolution
- Simpler interface (one method instead of two)

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking existing tests during migration | Keep old functions as wrappers calling new methods during transition |
| Circular import between application and services | DTOs live in application/ with interfaces; services imports application |
| Over-engineering for small codebase | Interfaces are minimal (4 services, ~8 types); manual wiring is simple |
| Context unused in current operations | Accept as hygiene; enables future use without retrofit |

## Migration Plan

**Phase 1: Add interfaces (no behavior change)**
1. Create `internal/application/` with interfaces, requests, results
2. Add struct types to services (`type transformer struct{}`)
3. Implement interface methods on structs
4. Keep old functions as wrappers calling methods
5. Update ServiceContainer with interface fields
6. Wire in main.go

**Phase 2: Migrate commands**
1. Update each command to call through `cfg.Services.XXX`
2. Update tests to use interfaces (can now mock)

**Phase 3: Cleanup**
1. Remove old package-level functions
2. Remove `InitializeFromPreset` function (logic moves to command)

**Rollback:** Revert to Phase 1 state if issues—old functions still work.

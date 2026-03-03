## Why

Services in `internal/services/` are currently called as package-level functions (`services.TransformDocument()`), creating tight coupling between CLI commands and implementations. This prevents mocking in tests, complicates dependency management, and diverges from idiomatic Go service patterns. Introducing interfaces with request/result types enables testability and cleaner architecture.

## What Changes

- Create `internal/application/` package with service interfaces (Transformer, Validator, Canonicalizer, Initializer)
- Add request/result types for each service operation alongside interfaces
- Convert package functions to struct methods implementing interfaces
- Populate `ServiceContainer` with wired service implementations
- Update CLI commands to call services through interfaces via `cfg.Services`
- Add `context.Context` to all service methods for idiomatic Go

## Capabilities

### New Capabilities

- `service-contracts`: Defines service interfaces (Transformer, Validator, Canonicalizer, Initializer) with request/result types in `internal/application/`

### Modified Capabilities

- `dependency-injection`: Extends existing ServiceContainer pattern to hold populated interface implementations, not just empty struct. Commands now access services through interfaces.

## Impact

**New Files:**
- `internal/application/interfaces.go` - Service interface definitions
- `internal/application/requests.go` - Request types for service operations
- `internal/application/results.go` - Result types for service operations

**Modified Files:**
- `internal/services/transformer.go` - Add struct type, implement interface
- `internal/services/canonicalizer.go` - Add struct type, implement interface
- `internal/services/initializer.go` - Rename InitOptions → InitializeRequest, implement interface
- `cmd/container.go` - Add interface fields to ServiceContainer
- `cmd/adapt.go` - Call `cfg.Services.Transformer.Transform()`
- `cmd/validate.go` - Call `cfg.Services.Validator.Validate()`
- `cmd/canonicalize.go` - Call `cfg.Services.Canonicalizer.Canonicalize()`
- `cmd/init.go` - Call `cfg.Services.Initializer.Initialize()`
- `main.go` - Wire implementations into ServiceContainer

**Tests:** Existing tests continue to work during migration (old functions kept as wrappers initially), then updated to use interfaces.

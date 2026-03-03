## 1. Create Application Package (Interfaces & DTOs)

- [x] 1.1 Create `internal/application/` directory
- [x] 1.2 Create `internal/application/requests.go` with TransformRequest, ValidateRequest, CanonicalizeRequest, InitializeRequest
- [x] 1.3 Create `internal/application/results.go` with TransformResult, ValidateResult (including Valid() method), CanonicalizeResult, InitializeResult
- [x] 1.4 Create `internal/application/interfaces.go` with Transformer, Validator, Canonicalizer, Initializer interfaces

## 2. Implement Service Structs (Phase 1 - Backward Compatible)

- [x] 2.1 Create `internal/services/validator.go` with `type validator struct{}` and `NewValidator()` constructor
- [x] 2.2 Implement `Validate(ctx, req)` method on validator struct
- [x] 2.3 Keep existing `ValidateDocument()` function in `internal/services/transformer.go` as wrapper calling method
- [x] 2.4 Update `internal/services/transformer.go` with `type transformer struct{}` and `NewTransformer()` constructor
- [x] 2.5 Implement `Transform(ctx, req)` method on transformer struct
- [x] 2.6 Keep existing `TransformDocument()` function as wrapper calling method
- [x] 2.7 Create `internal/services/canonicalizer.go` with `type canonicalizer struct{}` and `NewCanonicalizer()` constructor
- [x] 2.8 Implement `Canonicalize(ctx, req)` method on canonicalizer struct
- [x] 2.9 Keep existing `CanonicalizeDocument()` function as wrapper calling method
- [x] 2.10 Update `internal/services/initializer.go`: rename `InitOptions` to `InitializeRequest` (add `Refs` field), rename `InitResult` to `InitializeResult` (moves to `internal/application/results.go`)
- [x] 2.11 Create `type initializer struct{}` and `NewInitializer()` constructor
- [x] 2.12 Implement `Initialize(ctx, req)` method on initializer struct
- [x] 2.13 Keep existing `InitializeResources()` function as wrapper calling method
- [x] 2.14 Remove `InitializeFromPreset()` function (logic moves to command)
- [x] 2.15 Add compile-time interface satisfaction checks in each service file (e.g., `var _ application.Transformer = (*transformer)(nil)`)

## 3. Wire ServiceContainer

- [x] 3.1 Update `cmd/container.go` to add interface fields: Transformer, Validator, Canonicalizer, Initializer
- [x] 3.2 Update `cmd.NewServiceContainer()` to populate all interface fields with implementations (constructor does wiring)
- [x] 3.3 Verify `main.go` uses `NewServiceContainer()` (no additional wiring needed in main)

## 4. Migrate Commands to Use Interfaces (Phase 2)

- [x] 4.1 Update `cmd/adapt.go` to call `cfg.Services.Transformer.Transform(ctx, req)` instead of `services.TransformDocument()`
- [x] 4.2 Update `cmd/validate.go` to call `cfg.Services.Validator.Validate(ctx, req)` instead of `services.ValidateDocument()`
- [x] 4.3 Update `cmd/canonicalize.go` to call `cfg.Services.Canonicalizer.Canonicalize(ctx, req)` instead of `services.CanonicalizeDocument()`
- [x] 4.4 Update `cmd/init.go` to call `cfg.Services.Initializer.Initialize(ctx, req)` instead of `services.InitializeResources()`
- [x] 4.5 Move preset resolution logic to `cmd/init.go` (was in `InitializeFromPreset`)

## 5. Cleanup (Phase 3)

- [x] 5.1 Remove `TransformDocument()` wrapper function from `internal/services/transformer.go`
- [x] 5.2 Remove `ValidateDocument()` wrapper function from `internal/services/transformer.go`
- [x] 5.3 Remove `CanonicalizeDocument()` wrapper function from `internal/services/canonicalizer.go`
- [x] 5.4 Remove `InitializeResources()` wrapper function from `internal/services/initializer.go`

## 6. Verification

- [x] 6.1 Run `mise run check` (lint, format, test, build)
- [x] 6.2 Verify all existing tests still pass
- [x] 6.3 Verify CLI commands work as expected (manual smoke test)

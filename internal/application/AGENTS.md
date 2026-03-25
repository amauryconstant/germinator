**Location**: `internal/application/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Application Package

Service interfaces and data transfer objects for dependency injection and testability.

## Files

| File | Purpose |
|------|---------|
| `interfaces.go` | Service interfaces (Transformer, Validator, Canonicalizer, Initializer) and infrastructure interfaces (Parser, Serializer) |
| `requests.go` | Request types for service operations |

**Note**: Result types have been moved to `internal/domain/results.go` as part of the domain consolidation.

---

# Service Interfaces

All service methods take `context.Context` as first parameter (idiomatic Go).

| Interface | Method | Purpose |
|-----------|--------|---------|
| Transformer | `Transform(ctx, *TransformRequest) (*domain.TransformResult, error)` | Transform canonical → platform |
| Validator | `Validate(ctx, *ValidateRequest) (*domain.ValidateResult, error)` | Validate document |
| Canonicalizer | `Canonicalize(ctx, *CanonicalizeRequest) (*domain.CanonicalizeResult, error)` | Convert platform → canonical |
| Initializer | `Initialize(ctx, *InitializeRequest) ([]domain.InitializeResult, error)` | Install library resources |

---

# Infrastructure Interfaces

Interfaces for dependency injection of infrastructure services, enabling unit testing without real I/O.

| Interface | Method | Purpose |
|-----------|--------|---------|
| Parser | `LoadDocument(path, platform) (*domain.Document, error)` | Document loading abstraction |
| Serializer | `RenderDocument(doc *domain.Document, platform string) (string, error)` | Document rendering abstraction |

**Location**: `internal/application/interfaces.go`

**Implementation**: Concrete adapters in `internal/infrastructure/parsing/` and `internal/infrastructure/serialization/`

**Usage**: Injected into `Transformer` and `Initializer` services via constructors.

---

# Request Types

| Type | Fields |
|------|--------|
| TransformRequest | InputPath, OutputPath, Platform |
| ValidateRequest | InputPath, Platform |
| CanonicalizeRequest | InputPath, OutputPath, Platform, DocType |
| InitializeRequest | Library, Platform, OutputDir, Refs, DryRun, Force |

---

# Result Types

Result types are defined in `internal/domain/results.go`:

| Type | Fields | Notes |
|------|--------|-------|
| TransformResult | OutputPath | |
| ValidateResult | Errors []error | Has `Valid() bool` method |
| CanonicalizeResult | OutputPath | |
| InitializeResult | Ref, InputPath, OutputPath, Error | Per-resource result |

---

# Usage Pattern

## Commands

Commands call services through interfaces via `cfg.Services`:

```go
result, err := cfg.Services.Transformer.Transform(ctx, &application.TransformRequest{
    InputPath:  inputPath,
    OutputPath: outputPath,
    Platform:   platform,
})
```

## ServiceContainer

`cmd/container.go` wires implementations with infrastructure injection:

```go
type ServiceContainer struct {
    Transformer   application.Transformer
    Validator     application.Validator
    Canonicalizer application.Canonicalizer
    Initializer   application.Initializer
}

func NewServiceContainer() *ServiceContainer {
    // Create infrastructure adapters
    parser := &parsingParser{}      // implements Parser interface
    serializer := &serializationSerializer{}  // implements Serializer interface

    return &ServiceContainer{
        Transformer:   services.NewTransformer(parser, serializer),
        Validator:     services.NewValidator(),
        Canonicalizer: services.NewCanonicalizer(),  // uses ParsePlatformDocument/MarshalCanonical
        Initializer:   services.NewInitializer(parser, serializer),
    }
}
```

**Note**: `Canonicalizer` is excluded from infrastructure injection because it uses `ParsePlatformDocument` and `MarshalCanonical` (different functions requiring separate interfaces).

---

# Implementations

Concrete implementations in `internal/service/`:
- `service.NewTransformer(parser Parser, serializer Serializer)` → implements `application.Transformer`
- `service.NewValidator()` → implements `application.Validator`
- `service.NewCanonicalizer()` → implements `application.Canonicalizer`
- `service.NewInitializer(parser Parser, serializer Serializer)` → implements `application.Initializer`

**Note**: `Transformer` and `Initializer` require `Parser` and `Serializer` for testability. `Canonicalizer` and `Validator` use different infrastructure patterns.

---

# Why This Package Exists

**Problem**: Package-level functions (`services.TransformDocument()`) created tight coupling, preventing mocking.

**Solution**: Interfaces with request/result types enable:
- Dependency injection via ServiceContainer
- Mock implementations in tests
- Clean separation of contract and implementation
- Context propagation for future use (cancellation, tracing)

**Location rationale**: `application/` avoids circular imports—`services/` imports `application/`, not vice versa.

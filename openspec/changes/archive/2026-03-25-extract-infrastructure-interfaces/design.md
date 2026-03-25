## Context

The current service layer directly instantiates infrastructure dependencies:

```go
// internal/service/transformer.go
func (t *transformer) Transform(_ context.Context, req *application.TransformRequest) (*domain.TransformResult, error) {
    doc, err := parsing.LoadDocument(req.InputPath, req.Platform)  // direct call
    rendered, err := serialization.RenderDocument(doc, req.Platform)  // direct call
}

// internal/service/initializer.go
func (i *initializer) Initialize(_ context.Context, req *application.InitializeRequest) ([]domain.InitializeResult, error) {
    doc, err := parsing.LoadDocument(inputPath, req.Platform)  // direct call
    rendered, err := serialization.RenderDocument(doc, req.Platform)  // direct call
}
```

This pattern:
- Prevents unit testing services without full parsing/serialization stack
- Violates dependency inversion principle (services depend on concrete implementations)
- Makes it impossible to inject mock implementations for test isolation

**Note**: `Canonicalizer` uses different infrastructure functions (`ParsePlatformDocument`, `MarshalCanonical`) and is excluded from this change. It will require separate interfaces in a future iteration.

Current architecture:
- `parsing.LoadDocument(path, platform) (*domain.Document, error)` - concrete function
- `serialization.RenderDocument(doc, platform) (string, error)` - concrete function
- `parsing.ParsePlatformDocument(path, platform, docType) (interface{}, error)` - concrete function (different signature)
- `serialization.MarshalCanonical(doc interface{}) (string, error)` - concrete function (different signature)
- Services call these directly in their methods

## Goals / Non-Goals

**Goals:**
- Enable unit testing of `Transformer` and `Initializer` services without filesystem or YAML parsing
- Extract `Parser` and `Serializer` interfaces in `internal/application/`
- Inject infrastructure dependencies via service constructors

**Non-Goals:**
- Not changing the `Canonicalizer` service (it uses `ParsePlatformDocument` and `MarshalCanonical` which require different interfaces)
- Not changing the `Validator` service (it uses a different pattern via `validation.go`)
- Not changing platform adapters (they remain concrete implementations)
- Not changing CLI commands or user-facing behavior

## Decisions

### Decision 1: Interface Location

**Choice:** Define `Parser` and `Serializer` interfaces in `internal/application/interfaces.go`

**Rationale:** Application layer already holds service interfaces (Transformer, Validator, etc.). Adding infrastructure interfaces there keeps dependency direction consistent: application layer defines contracts, infrastructure implements them.

**Alternatives Considered:**
- Define in `internal/domain/` - rejected because domain should remain pure and domain already imports parsing types
- Define in `internal/infrastructure/` - rejected because that inverts dependency (infrastructure would define interfaces consumed by application)

### Decision 2: Interface Methods

**Choice:** Parser interface matches `LoadDocument` signature; Serializer interface matches `RenderDocument` signature

```go
type Parser interface {
    LoadDocument(path string, platform string) (*domain.Document, error)
}

type Serializer interface {
    RenderDocument(doc *domain.Document, platform string) (string, error)
}
```

**Rationale:** Minimal interface that covers Transformer and Initializer needs. Implementation can delegate to existing functions.

**Alternatives Considered:**
- Add context to interface methods - deferred to future iteration for context propagation
- Split into smaller interfaces (e.g., `Loader`, `Renderer`) - premature complexity

### Decision 3: Constructor Signature Changes

**Choice:** Change service constructors to accept infrastructure interfaces

```go
// Before
func NewTransformer() application.Transformer
func NewInitializer() application.Initializer

// After
func NewTransformer(parser Parser, serializer Serializer) application.Transformer
func NewInitializer(parser Parser, serializer Serializer) application.Initializer
```

**Rationale:** Standard constructor injection pattern. Makes dependencies explicit and testable.

### Decision 4: ServiceContainer Updates

**Choice:** ServiceContainer creates infrastructure instances and passes to service constructors

```go
func NewServiceContainer() *ServiceContainer {
    parser := &parsingParser{}  // concrete implementation
    serializer := &serializationSerializer{}  // concrete implementation

    return &ServiceContainer{
        Transformer:   service.NewTransformer(parser, serializer),
        Validator:     service.NewValidator(),
        Canonicalizer: service.NewCanonicalizer(),  // No infrastructure injection - uses different functions
        Initializer:   service.NewInitializer(parser, serializer),
    }
}
```

**Rationale:** Keeps ServiceContainer as composition root. Infrastructure implementations remain internal to infrastructure package. Canonicalizer is unchanged because it uses ParsePlatformDocument/MarshalCanonical which need different interfaces.

### Decision 5: Mock Implementations

**Choice:** Create mocks for Parser and Serializer in `test/mocks/`

```go
type MockParser struct {
    LoadDocumentFunc func(path string, platform string) (*domain.Document, error)
}

type MockSerializer struct {
    RenderDocumentFunc func(doc *domain.Document, platform string) (string, error)
}
```

**Rationale:** Follows existing mock pattern (see `test/mocks/*.go`). Enables service unit tests with mock infrastructure.

## Risks / Trade-offs

[Risk] Breaking change to service constructors
→ [Mitigation] All callers updated in same change. ServiceContainer remains single instantiation point.

[Risk] Context not propagated through infrastructure
→ [Mitigation] Deferred to future iteration. Current interfaces don't include context; services pass `_` (ignored).

[Risk] Canonicalizer not included in interface extraction
→ [Mitigation] Future iteration can add `PlatformParser` and `Marshaler` interfaces for Canonicalizer if needed.

## Migration Plan

1. Add `Parser` and `Serializer` interfaces to `internal/application/interfaces.go`
2. Create concrete adapters in `internal/infrastructure/` that implement new interfaces
3. Update `Transformer` and `Initializer` constructors to accept new interfaces
4. Update `ServiceContainer` to instantiate and inject infrastructure
5. Create mocks in `test/mocks/`
6. Add unit tests for `Transformer` and `Initializer` using mocks
7. Run full test suite to verify no regressions

**Rollback:** Revert service constructor changes and ServiceContainer wiring. Interfaces can remain (no harm).

## Open Questions

1. ~~Should Validator also receive Parser and Serializer for future flexibility, or keep as-is?~~ **Closed**: Validator is not changed.
2. ~~Do we need `Parser` and `Serializer` interfaces to include `context.Context` in this iteration, or defer?~~ **Closed**: Deferred to future iteration. Context propagation out of scope for this change.
3. Should Canonicalizer get its own `PlatformParser` and `Marshaler` interfaces in a future iteration?

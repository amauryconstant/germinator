# Mock Infrastructure

Mock implementations of application interfaces for isolated unit testing.

## Purpose

This package provides testify/mock implementations:
- **Service interfaces**: `Transformer`, `Validator`, `Canonicalizer`, `Initializer`
- **Infrastructure interfaces**: `Parser`, `Serializer`

Mocks enable tests to isolate business logic from real implementations. Tests choose mocks or real implementations based on testing strategy.

## Available Mocks

### MockParser

**Interface**: `application.Parser`

**Methods**:
- `LoadDocument(path string, platform string) (*domain.Document, error)`

**Use Cases**:
- Unit test `Transformer` and `Initializer` without filesystem I/O
- Simulate document loading errors
- Test service behavior with specific document content

**Example**:
```go
mockParser := new(mocks.MockParser)
mockParser.On("LoadDocument", "/path/doc.md", "opencode").
    Return(&domain.Document{Content: "test"}, nil)

doc, err := mockParser.LoadDocument("/path/doc.md", "opencode")
```

---

### MockSerializer

**Interface**: `application.Serializer`

**Methods**:
- `RenderDocument(doc *domain.Document, platform string) (string, error)`

**Use Cases**:
- Unit test `Transformer` and `Initializer` without serialization
- Simulate render failures
- Verify rendered output

**Example**:
```go
mockSerializer := new(mocks.MockSerializer)
mockSerializer.On("RenderDocument", mock.Anything, "opencode").
    Return("---\ncontent\n---", nil)

output, err := mockSerializer.RenderDocument(doc, "opencode")
```

---

### MockTransformer

**Interface**: `application.Transformer`

**Methods**:
- `Transform(ctx context.Context, req *application.TransformRequest) (*application.TransformResult, error)`

**Use Cases**:
- Test command handlers that transform documents without real I/O
- Verify transformation logic independently of serialization
- Simulate transformation failures for error handling tests

**Example**:
```go
mockTransformer := new(mocks.MockTransformer)
mockTransformer.On("Transform", ctx, mock.AnythingOfType("*application.TransformRequest")).
    Return(&application.TransformResult{OutputPath: "/output.md"}, nil)

result, err := mockTransformer.Transform(ctx, &application.TransformRequest{...})
```

---

### MockValidator

**Interface**: `application.Validator`

**Methods**:
- `Validate(ctx context.Context, req *application.ValidateRequest) (*application.ValidateResult, error)`

**Use Cases**:
- Test validation logic independently of real validation rules
- Simulate validation errors for error handling tests
- Verify command handler behavior based on validation results

**Example**:
```go
mockValidator := new(mocks.MockValidator)
mockValidator.On("Validate", ctx, mock.AnythingOfType("*application.ValidateRequest")).
    Return(&application.ValidateResult{Errors: []error{errors.New("invalid")}}, nil)

result, err := mockValidator.Validate(ctx, &application.ValidateRequest{...})
```

---

### MockCanonicalizer

**Interface**: `application.Canonicalizer`

**Methods**:
- `Canonicalize(ctx context.Context, req *application.CanonicalizeRequest) (*application.CanonicalizeResult, error)`

**Use Cases**:
- Test canonicalization logic independently of real parsing
- Simulate platform-specific conversion errors
- Verify command handler behavior with different document types

**Example**:
```go
mockCanonicalizer := new(mocks.MockCanonicalizer)
mockCanonicalizer.On("Canonicalize", ctx, mock.AnythingOfType("*application.CanonicalizeRequest")).
    Return(&application.CanonicalizeResult{OutputPath: "/canonical.yaml"}, nil)

result, err := mockCanonicalizer.Canonicalize(ctx, &application.CanonicalizeRequest{...})
```

---

### MockInitializer

**Interface**: `application.Initializer`

**Methods**:
- `Initialize(ctx context.Context, req *application.InitializeRequest) ([]application.InitializeResult, error)`

**Use Cases**:
- Test initialization logic independently of real library resources
- Simulate installation failures for error handling tests
- Verify command handler behavior with multiple resources

**Example**:
```go
mockInitializer := new(mocks.MockInitializer)
mockInitializer.On("Initialize", ctx, mock.AnythingOfType("*application.InitializeRequest")).
    Return([]application.InitializeResult{
        {Ref: "skill/commit", InputPath: "/lib/skill/commit.md", OutputPath: "/output/commit.md"},
    }, nil)

results, err := mockInitializer.Initialize(ctx, &application.InitializeRequest{...})
```

---

## Mock Lifecycle

### 1. Create Mock Instance

```go
mockValidator := new(mocks.MockValidator)
```

### 2. Set Up Expected Calls

Use `On()` to define expected method calls and return values:

```go
mockValidator.On("Validate", ctx, expectedReq).
    Return(&application.ValidateResult{Errors: []error{}}, nil)
```

**Argument Matching Options**:
- Exact match: `On("Method", exactArg1, exactArg2)`
- Type match: `On("Method", mock.AnythingOfType("*application.ValidateRequest"))`
- Any value: `On("Method", mock.Anything)`

**Return Values**:
- Single return: `Return(result, nil)`
- Multiple return: `Return(result, error)`

### 3. Call the Method

```go
result, err := mockValidator.Validate(ctx, &application.ValidateRequest{
    InputPath: "/path/to/doc.md",
    Platform:  "opencode",
})
```

### 4. Verify Behavior

Use assertions to verify the method was called:

```go
mockValidator.AssertCalled(t, "Validate", ctx, req)
mockValidator.AssertNumberOfCalls(t, "Validate", 1)
mockValidator.AssertExpectations(t)  // Verify all expectations were met
```

### 5. Reset Mock (if needed)

```go
mockValidator.ExpectedCalls = nil  // Clear all expectations
```

---

## Best Practices

### DO:

- Use mocks for unit tests that need to isolate from real implementations
- Use specific argument matching when possible for better test precision
- Always call `AssertExpectations(t)` at the end of each test
- Reset mocks between test cases when reusing the same mock instance
- Document the expected behavior in test comments

### DON'T:

- Mock everything - use real implementations when they're fast and reliable
- Over-specify expectations - only assert what's important for the test
- Forget to verify expectations - tests may pass without calling the mocked method
- Mix mocks and real implementations in the same test without clear intent
- Use mocks for integration tests - they're for unit tests

---

## Mock vs. Real Implementation

| Scenario | Use Mock | Use Real Implementation |
|----------|----------|-------------------------|
| Fast unit tests of business logic | ✓ | |
| Integration tests with I/O | | ✓ |
| Testing error handling | ✓ | |
| Testing with real data | | ✓ |
| Test isolation from external dependencies | ✓ | |
| Golden file tests | | ✓ |

---

## See Also

- `test/AGENTS.md` - Comprehensive mock usage patterns and examples
- `cmd/validate_test.go` - Example test demonstrating MockValidator usage
- `internal/application/interfaces.go` - Interface definitions
- `internal/application/AGENTS.md` - Application package documentation

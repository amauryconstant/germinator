**Location**: `test/`
**Parent**: See `/AGENTS.md` for project overview

---

# Testing Infrastructure

## Structure

- `fixtures/` - Test fixtures (Germinator format inputs)
- `golden/` - Golden files (expected outputs)
- `e2e/` - E2E tests (CLI behavior validation)
- `mocks/` - Mock implementations of application service interfaces
- `helpers/` - Shared test utilities (future)

---

## E2E Testing

**Stack**: Ginkgo v2, Gomega, gexec
**Build tag**: `//go:build e2e` (excluded from `go test ./...`)

### Running E2E Tests

```bash
mise run test:e2e          # E2E tests only
mise run test              # All tests (unit + golden + integration + e2e)
go test -tags=e2e ./test/e2e/... -run TestValidate -v  # Specific test
```

### E2E Test Structure

| File | Purpose |
|------|---------|
| `e2e_suite_test.go` | Suite setup, builds binary in BeforeSuite |
| `helpers/cli_helper.go` | CLI runner utilities |
| `fixtures/` | E2E-specific test fixtures |
| `*_test.go` | Test cases by command |

### CLI Helper Pattern

```go
cli := helpers.NewGerminatorCLI(e2e.BinaryPath())
session := cli.Run("validate", fixturePath, "--platform", "opencode")
cli.ShouldSucceed(session)
cli.ShouldOutput(session, "Document is valid")
```

### Adding E2E Tests

1. Create test file in `test/e2e/` with `//go:build e2e` tag
2. Use `e2e.BinaryPath()` for built binary path
3. Use `helpers.NewGerminatorCLI()` for CLI operations
4. Add fixtures to `test/e2e/fixtures/` if needed

---

## Integration Testing

**Build tag**: `//go:build integration` (excluded from `go test ./...`)

### Running Integration Tests

```bash
mise run test:integration              # All integration tests
mise run test:coverage:integration     # Integration tests with coverage
go test -tags=integration ./internal/core -v
```

### Integration Test Structure

Located in `internal/infrastructure/parsing/integration_test.go`:

- Tests end-to-end document loading pipeline: `LoadDocument → Validate → RenderDocument`
- Uses fixtures from `test/fixtures/` for representative document types
- Verifies document structure, file paths, and content extraction

### Adding Integration Tests

1. Create test functions in `internal/infrastructure/parsing/integration_test.go` with `//go:build integration` tag
2. Use fixtures from `test/fixtures/` for test data
3. Test complete pipelines, not individual functions
4. Keep integration tests focused on critical workflows

---

## Golden File Testing

### Usage

- **Test fixtures**: `test/fixtures/opencode/` contains Germinator format YAML files (inputs)
- **Golden files**: `test/golden/opencode/` contains expected OpenCode format outputs

### Running Tests

```bash
mise run test:golden              # All golden file tests
mise run test:coverage:golden      # Golden tests with coverage
go test -tags=golden ./internal/services -v
go test -tags=golden ./internal/services -run TestGoldenFiles/agent-full -v  # Specific test
```

### Updating Golden Files

```bash
UPDATE_GOLDEN=true mise run test:golden
# or
UPDATE_GOLDEN=true go test -tags=golden ./internal/services -v
```

Commit updated golden files. Verify without flag: `mise run test:golden`

### Adding New Golden File Tests

1. Create fixture in `test/fixtures/opencode/`
2. Generate output: `./bin/germinator adapt test/fixtures/opencode/new-fixture.md /tmp/output.md --platform opencode`
3. Copy to `test/golden/opencode/new-fixture.md.golden`
4. Add test case to `transformer_golden_test.go`

### Test Structure

**Build tag**: `//go:build golden` (excluded from `go test ./...`)

Table-driven in `internal/service/transformer_golden_test.go`:
```go
tests := []struct {
    name     string
    fixture  string // Germinator format fixture
    golden   string // Golden file path
    platform string // Platform to test
}{
    {"agent-full", "../../test/fixtures/opencode/agent-full.md", "../../test/golden/opencode/agent-full.md.golden", "opencode"},
}
```

---

## Mock Testing

### Overview

Mock infrastructure provides testify/mock implementations of all application service interfaces, enabling isolated unit testing without real implementations. Mocks are optional - tests can choose whether to use mocks or real implementations based on their testing strategy.

### Available Mocks

| Mock | Interface | Purpose |
|------|-----------|---------|
| `MockTransformer` | `application.Transformer` | Document transformation |
| `MockValidator` | `application.Validator` | Document validation |
| `MockCanonicalizer` | `application.Canonicalizer` | Platform → canonical conversion |
| `MockInitializer` | `application.Initializer` | Library resource installation |

**Location**: `test/mocks/` - See `test/mocks/AGENTS.md` for complete mock inventory and documentation

### Mock Usage Pattern

#### 1. Create Mock Instance

```go
import "gitlab.com/amoconst/germinator/test/mocks"

mockValidator := new(mocks.MockValidator)
```

#### 2. Set Up Expected Calls

Use `On()` to define expected method calls and return values:

```go
import (
    "context"
    "github.com/stretchr/testify/mock"
)

ctx := context.Background()
expectedReq := &application.ValidateRequest{
    InputPath: "/path/to/doc.md",
    Platform:  "opencode",
}

// Exact argument matching
mockValidator.On("Validate", ctx, expectedReq).
    Return(&application.ValidateResult{Errors: []error{}}, nil)

// Type-based matching (flexible)
mockValidator.On("Validate", ctx, mock.AnythingOfType("*application.ValidateRequest")).
    Return(&application.ValidateResult{Errors: []error{}}, nil)

// Match anything
mockValidator.On("Validate", ctx, mock.Anything).
    Return(&application.ValidateResult{Errors: []error{}}, nil)
```

**Return Value Options**:
- Success: `Return(&ValidateResult{Errors: []error{}}, nil)`
- Validation errors: `Return(&ValidateResult{Errors: []error{err1, err2}}, nil)`
- Fatal error: `Return(nil, errors.New("file not found"))`

#### 3. Call the Method

```go
result, err := mockValidator.Validate(ctx, &application.ValidateRequest{
    InputPath: "/path/to/doc.md",
    Platform:  "opencode",
})
```

#### 4. Verify Behavior

Use assertions to verify the method was called:

```go
import "github.com/stretchr/testify/assert"

// Verify method was called with specific arguments
mockValidator.AssertCalled(t, "Validate", ctx, req)

// Verify method was called a specific number of times
mockValidator.AssertNumberOfCalls(t, "Validate", 1)

// Verify all expectations were met
mockValidator.AssertExpectations(t)

// Verify method was NOT called
mockValidator.AssertNotCalled(t, "Validate")
```

#### 5. Reset Mock (if needed)

```go
mockValidator.ExpectedCalls = nil  // Clear all expectations
```

### Complete Example

```go
package cmd_test

import (
    "context"
    "testing"
    "errors"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gitlab.com/amoconst/germinator/internal/application"
    "gitlab.com/amoconst/germinator/test/mocks"
)

func TestCommandWithMockValidator(t *testing.T) {
    // Setup: Create mock
    mockValidator := new(mocks.MockValidator)
    ctx := context.Background()

    // Arrange: Set up expected call
    expectedReq := &application.ValidateRequest{
        InputPath: "/path/to/doc.md",
        Platform:  "opencode",
    }
    mockValidator.On("Validate", ctx, expectedReq).
        Return(&application.ValidateResult{
            Errors: []error{errors.New("missing required field")},
        }, nil)

    // Act: Call the method being tested
    result, err := mockValidator.Validate(ctx, expectedReq)

    // Assert: Verify results
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.False(t, result.Valid())
    assert.Len(t, result.Errors, 1)

    // Verify: Ensure method was called as expected
    mockValidator.AssertCalled(t, "Validate", ctx, expectedReq)
    mockValidator.AssertExpectations(t)
}
```

### Argument Matching Strategies

| Matcher | Usage | When to Use |
|---------|-------|--------------|
| Exact value | `On("Method", ctx, exactReq)` | When you know exact input |
| Type match | `On("Method", ctx, mock.AnythingOfType("*Req"))` | When input varies but type matters |
| Anything | `On("Method", ctx, mock.Anything)` | When input doesn't matter |
| Custom function | `On("Method", ctx, mock.MatchedBy(func(req *Req) bool { ... }))` | Complex validation logic |

### Multiple Calls Setup

```go
// Return different values on different calls
mockValidator.On("Validate", ctx, req1).Return(result1, nil)
mockValidator.On("Validate", ctx, req2).Return(result2, errors.New("error"))

// Or use After() for call ordering
mockValidator.On("Validate", ctx, req1).Return(result1, nil)
mockValidator.On("Validate", ctx, req2).Return(result2, nil).After(mockValidator.On("Validate", ctx, req1))
```

### Mock vs. Real Implementation

| Scenario | Use Mock | Use Real Implementation |
|----------|----------|-------------------------|
| Fast unit tests of business logic | ✓ | |
| Integration tests with I/O | | ✓ |
| Testing error handling | ✓ | |
| Testing with real data | | ✓ |
| Test isolation from external dependencies | ✓ | |
| Golden file tests | | ✓ |
| E2E tests | | ✓ |

### Best Practices

#### DO:

- Use mocks for unit tests that need to isolate from real implementations
- Use specific argument matching when possible for better test precision
- Always call `AssertExpectations(t)` at the end of each test
- Reset mocks between test cases when reusing the same mock instance
- Document the expected behavior in test comments
- Keep mocks focused on a single behavior per test

#### DON'T:

- Mock everything - use real implementations when they're fast and reliable
- Over-specify expectations - only assert what's important for the test
- Forget to verify expectations - tests may pass without calling the mocked method
- Mix mocks and real implementations in the same test without clear intent
- Use mocks for integration tests - they're for unit tests only
- Create overly complex mock setups - simplify test logic

### Common Patterns

#### Pattern: Test Error Handling

```go
mockValidator.On("Validate", ctx, mock.AnythingOfType("*application.ValidateRequest")).
    Return(nil, errors.New("file not found"))

result, err := mockValidator.Validate(ctx, req)
assert.Error(t, err)
assert.Nil(t, result)
```

#### Pattern: Test with Multiple Errors

```go
mockValidator.On("Validate", ctx, req).
    Return(&application.ValidateResult{
        Errors: []error{
            errors.New("missing name"),
            errors.New("invalid platform"),
        },
    }, nil)

result, err := mockValidator.Validate(ctx, req)
assert.NoError(t, err)
assert.Len(t, result.Errors, 2)
```

#### Pattern: Test Success Path

```go
mockValidator.On("Validate", ctx, req).
    Return(&application.ValidateResult{Errors: []error{}}, nil)

result, err := mockValidator.Validate(ctx, req)
assert.NoError(t, err)
assert.True(t, result.Valid())
```

### See Also

- `test/mocks/AGENTS.md` - Complete mock inventory and detailed documentation
- `cmd/validate_test.go` - Example test demonstrating MockValidator usage with multiple scenarios
- `internal/application/interfaces.go` - Interface definitions being mocked
- `internal/application/AGENTS.md` - Application package documentation

---

## Test Data Setup Patterns

**t.TempDir()** (dynamic): Use for test-specific data, file modifications, isolated scenarios
```go
tmpDir := t.TempDir()
testFile := filepath.Join(tmpDir, "test-agent.md")
```

**Static fixtures** (shared): Use for representative examples, golden file inputs
```go
fixturesDir := filepath.Join("..", "..", "test", "fixtures")
testFile := filepath.Join(fixturesDir, "agent-test.md")
```

**Embedded fixtures** (cross-platform): Use for CI/CD, tests from any directory
```go
//go:embed testdata/agent.md
var agentFixture []byte
```

---

## Test Naming Conventions

### Test Function Names

Pattern: `Test<FunctionOrComponent><Scenario>`

**Examples**:
- `TestValidateDocumentWithValidInput` - good
- `TestAgentModel` - bad (too vague)
- `TestLoadDocumentIntegration` - good

### Test Case Names (Table-Driven)

Descriptive, short, kebab-case. Avoid generic names like "test1", "test2"

**Good examples**:
```go
{name: "valid agent with all fields"},
{name: "missing name returns error"},
```

### Fixture File Names

Pattern: `<doc-type>-<scenario>.md`

**Examples**:
- `agent-full.md` - agent with all fields
- `command-test.md` - test command example
- `memory-paths-only.md` - memory with only paths

---

## Platform Testing Expectations

### Required Platforms

All tests with `platform` parameter MUST test both `claude-code` and `opencode`

### Test Structure

```go
tests := []struct {
    name     string
    platform  string
    input     string
    wantError bool
}{
    {name: "valid agent (claude-code)", platform: "claude-code", input: validAgent, wantError: false},
    {name: "valid agent (opencode)", platform: "opencode", input: validAgent, wantError: false},
}
```

### Platform-Specific Validation

- **Claude Code**: Short model names (sonnet, opus, haiku), permissionMode enum
- **OpenCode**: Mode values (primary/subagent/all), temperature range (0.0-1.0), maxSteps (> 0)

---

## Table-Driven Test Pattern

See Go testing documentation for standard table-driven test pattern.

## Error Counting Patterns

Use explicit `errorCount` field in table tests when validation functions return `[]error` and exact error count matters.

```go
tests := []struct {
    name       string
    input      string
    errorCount int
}{
    {name: "valid input", input: validData, errorCount: 0},
    {name: "missing name", input: invalidData, errorCount: 1},
}
```

Use `len(errs) > 0` for simple pass/fail checks.

---

## CI Integration

Golden file tests run automatically in CI (`go test ./... -v`). All must pass before merging.

**Notes**:
- Golden files are byte-exact comparisons (line endings, whitespace matter)
- Use `UPDATE_GOLDEN=true` when updating multiple golden files
- Verify golden files match expected transformation behavior

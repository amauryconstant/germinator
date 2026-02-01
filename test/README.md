# Test Support

This directory stores test fixtures for testing.

## Structure

- `fixtures/` - Test fixtures and test data (Germinator format inputs)
- `golden/` - Golden files (expected outputs for transformation tests)

## When to Add Files

Add files to these directories when writing tests:
- Place test data in `fixtures/` (Germinator format YAML files)
- Place expected outputs in `golden/` (for transformation tests)

## Test Data Setup Patterns

### Pattern 1: t.TempDir() (Dynamic Test Files)

**Use when:**
- Test-specific test data that doesn't need to be shared
- Tests that need to create/modify files
- Isolated test scenarios

**Example:**
```go
func TestValidation(t *testing.T) {
    tmpDir := t.TempDir()
    testFile := tmpDir + "/test-agent.md"

    content := `---
name: test-agent
description: Test
---
Content`

    if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }

    // Test with testFile...
}
```

**Benefits:**
- Automatic cleanup after test
- No stale test files
- Isolated test environments

### Pattern 2: Static Fixtures (Shared Test Data)

**Use when:**
- Test data used across multiple test files
- Representative examples of document formats
- Golden file test inputs

**Example:**
```go
func TestLoadDocument(t *testing.T) {
    fixturesDir := filepath.Join("..", "..", "test", "fixtures")
    testFile := filepath.Join(fixturesDir, "agent-test.md")

    doc, err := LoadDocument(testFile, "claude-code")
    // Test with doc...
}
```

**Benefits:**
- Reusable across tests
- Version-controlled test data
- Consistent test scenarios

### Pattern 3: Embedded Fixtures (for Cross-Platform Tests)

**Use when:**
- Tests must work from any working directory
- CI/CD environments with different path structures
- Integration tests

**Example:**
```go
import _ "embed"

//go:embed testdata/agent.md
var agentFixture []byte

func TestWithEmbeddedFixture(t *testing.T) {
    content := string(agentFixture)
    // Test with content...
}
```

**Benefits:**
- No external file dependencies
- Tests work from any directory
- Self-contained test suites

### Decision Guidelines

Choose **t.TempDir()** when:
- Test creates files during execution
- Test needs isolated file system
- Files are test-specific and not shared

Choose **static fixtures** when:
- Multiple tests need same input data
- Testing representative real-world examples
- Golden file test inputs

Choose **embedded fixtures** when:
- Tests run from unpredictable directories
- Need complete test isolation
- Want self-contained test package

## Golden File Testing

### Purpose

Golden files store expected output from document transformations. They provide a way to verify that serialization logic produces correct output and catch regressions when code changes.

### Usage

- **Test fixtures**: `test/fixtures/opencode/` contains Germinator format YAML files (inputs)
- **Golden files**: `test/golden/opencode/` contains expected OpenCode format outputs

### Running Golden File Tests

Run all golden file tests:
```bash
go test ./internal/services -run TestGoldenFiles -v
```

Run a specific golden file test:
```bash
go test ./internal/services -run TestGoldenFiles/agent-full -v
```

### Updating Golden Files

When transformation logic changes and golden files need updating:

1. Set the `UPDATE_GOLDEN` environment variable:
```bash
UPDATE_GOLDEN=true go test ./internal/services -run TestGoldenFiles -v
```

2. Commit the updated golden files along with your changes

3. Verify tests pass without the flag:
```bash
go test ./internal/services -run TestGoldenFiles -v
```

### Golden File Test Structure

Golden file tests are table-driven and defined in `internal/services/transformer_golden_test.go`:

```go
tests := []struct {
    name     string
    fixture  string // Germinator format fixture
    golden   string // Golden file path
    platform string // Platform to test
}{
    {"agent-full", "../../test/fixtures/opencode/agent-full.md", "../../test/golden/opencode/agent-full.md.golden", "opencode"},
    // ... more test cases
}
```

Each test:
1. Loads a Germinator format fixture
2. Runs `TransformDocument` with the specified platform
3. Reads the expected output from the golden file
4. Compares actual vs expected (byte-by-byte)
5. Shows detailed diff if mismatch

### Adding New Golden File Tests

To add a new golden file test:

1. Create a fixture file in `test/fixtures/opencode/` (Germinator format)
2. Run the transformation to generate expected output:
```bash
./bin/germinator adapt test/fixtures/opencode/new-fixure.md /tmp/output.md --platform opencode
```
3. Copy the output to `test/golden/opencode/new-fixture.md.golden`
4. Add test case to `transformer_golden_test.go`:
```go
{
    name:     "new-fixture",
    fixture:  "../../test/fixtures/opencode/new-fixture.md",
    golden:   "../../test/golden/opencode/new-fixture.md.golden",
    platform: "opencode",
},
```
5. Run tests to verify

### CI Integration

Golden file tests run automatically in CI as part of the `test` stage (`go test ./... -v`). All golden file tests must pass before merging.

### Notes

- Golden files are byte-exact comparisons (line endings, whitespace matter)
- Use `UPDATE_GOLDEN=true` when updating multiple golden files at once
- Always verify golden files match expected transformation behavior
- Golden files should be committed with the code that produces them
- New subdirectories may be added as needed for test organization

## Test Naming Conventions

### Test Function Names

Use descriptive names following the pattern:
- `Test<FunctionOrComponent><Scenario>` - unit tests
- `Test<Feature><Integration>` - integration tests
- `TestGoldenFiles` - golden file tests

**Examples:**
- `TestValidateDocumentWithValidInput` - good
- `TestAgentModel` - bad (too vague)
- `TestLoadDocumentIntegration` - good

### Test Case Names (Table-Driven)

Use descriptive, short names:
- Describe the scenario being tested
- Use kebab-case
- Avoid generic names like "test1", "test2"

**Good examples:**
```go
{
    name: "valid agent with all fields",
    // ...
},
{
    name: "missing name returns error",
    // ...
},
{
    name: "invalid temperature range",
    // ...
},
```

**Bad examples:**
```go
{
    name: "test1",  // Not descriptive
    // ...
},
{
    name: "good test",  // Too vague
    // ...
},
```

### Fixture File Names

Use clear, descriptive names:
- `<doc-type>-<scenario>.md` - for document fixtures
- `<doc-type>-full.md` - complete example with all fields
- `<doc-type>-minimal.md` - minimal valid example

**Examples:**
- `agent-full.md` - agent with all fields
- `command-test.md` - test command example
- `memory-paths-only.md` - memory with only paths
- `skill-invalid.md` - intentionally invalid skill

## Platform Testing Expectations

### Required Platforms

All tests that accept a `platform` parameter MUST test both:
- `claude-code` - Claude Code platform
- `opencode` - OpenCode platform

### Test Structure

**Example:**
```go
tests := []struct {
    name     string
    platform  string
    input     string
    wantError bool
}{
    {
        name:    "valid agent (claude-code)",
        platform: "claude-code",
        input:    validAgent,
        wantError: false,
    },
    {
        name:    "valid agent (opencode)",
        platform: "opencode",
        input:    validAgent,
        wantError: false,
    },
    {
        name:    "invalid platform",
        platform: "invalid-platform",
        input:    validAgent,
        wantError: true,
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test with tt.platform...
    })
}
```

### Platform-Specific Validation

Tests should verify platform-specific constraints:
- **Claude Code**: Short model names (sonnet, opus, haiku), permissionMode enum
- **OpenCode**: Mode values (primary/subagent/all), temperature range (0.0-1.0), maxSteps (> 0)

**Example:**
```go
func TestOpenCodeTemperatureRange(t *testing.T) {
    tests := []struct {
        name    string
        value   float64
        wantErr bool
    }{
        {"valid 0.0", 0.0, false},
        {"valid 0.5", 0.5, false},
        {"valid 1.0", 1.0, false},
        {"invalid -0.5", -0.5, true},
        {"invalid 1.5", 1.5, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := validateTemperature(tt.value)
            if (len(errs) > 0) != tt.wantErr {
                t.Errorf("validateTemperature(%v) error = %v, wantErr %v", tt.value, errs, tt.wantErr)
            }
        })
    }
}
```

## Examples for Adding New Tests

### Adding a Unit Test

1. Create test file if doesn't exist: `<package>_test.go`
2. Use table-driven pattern for multiple scenarios
3. Test both success and error cases
4. Follow naming conventions

```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "result",
            wantErr: false,
        },
        {
            name:  "invalid input",
            input: "",
            want:  "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewFeature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("NewFeature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Adding a Golden File Test

See "Golden File Testing" section above for detailed steps.

### Adding an Integration Test

1. Create file in `internal/core/integration_test.go` or appropriate package
2. Use table-driven structure
3. Test end-to-end workflows
4. Verify multiple components work together

```go
func TestEndToEndWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        fixture string
        platform string
    }{
        {
            name:    "agent workflow",
            fixture: "../../test/fixtures/agent-test.md",
            platform: "claude-code",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Load document
            doc, err := LoadDocument(tt.fixture, tt.platform)
            if err != nil {
                t.Fatalf("LoadDocument failed: %v", err)
            }

            // Validate document
            errs := doc.Validate(tt.platform)
            if len(errs) > 0 {
                t.Errorf("Validation failed: %v", errs)
            }

            // Transform document
            output, err := RenderDocument(doc, tt.platform)
            if err != nil {
                t.Errorf("RenderDocument failed: %v", err)
            }

            // Verify output is not empty
            if output == "" {
                t.Error("Expected non-empty output")
            }
        })
    }
}
```

## Table-Driven Test Pattern

### Structure

Table-driven tests are the preferred pattern for testing multiple scenarios:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string  // Test case name (required)
        input    string  // Input parameters (vary by function)
        expected string  // Expected output (vary by function)
        wantErr  bool    // Whether error is expected
    }{
        {
            name:     "descriptive test case name",
            input:    "test input",
            expected: "expected output",
            wantErr:  false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionUnderTest(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
                return
            }

            if got != tt.expected {
                t.Errorf("FunctionUnderTest(%q) = %q, want %q", tt.input, got, tt.expected)
            }
        })
    }
}
```

### Benefits

- **Consistent structure**: All tests follow same pattern
- **Easy to add cases**: Just add to struct slice
- **Clear failure output**: `t.Run()` provides context
- **Parallel execution**: Add `t.Parallel()` if tests are independent

### When to Use

- Multiple scenarios for same function
- Testing edge cases and boundary conditions
- Validation tests with various inputs
- Platform testing (different platforms as test cases)

### When NOT to Use

- Single, simple test case (overhead not worth it)
- Complex setup that differs per test case
- Tests that require different fixtures per case

## Error Counting Patterns

### Standard Pattern

When testing validation functions that return `[]error`, use explicit error count in test struct:

```go
tests := []struct {
    name       string
    input      string
    errorCount int  // Explicit count of expected errors
}{
    {
        name:       "valid input",
        input:      validData,
        errorCount: 0,  // No errors expected
    },
    {
        name:       "missing name",
        input:      invalidData,
        errorCount: 1,  // Exactly 1 error expected
    },
    {
        name:       "multiple errors",
        input:      multipleErrorsData,
        errorCount: 2,  // Exactly 2 errors expected
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        errs := FunctionUnderTest(tt.input)
        if len(errs) != tt.errorCount {
            t.Errorf("FunctionUnderTest() error count = %d, want %d", len(errs), tt.errorCount)
        }
    })
}
```

### When to Use Error Count

**Use explicit errorCount when:**
- Testing validation functions that return multiple errors
- Need to verify exact number of validation issues
- Testing complex validation rules with multiple constraints
- Testing OpenCode-specific validation (mode, temperature, maxSteps)

**Use simple len(errs) > 0 when:**
- Testing that errors occur, not exact count
- Error indicates binary pass/fail
- Checking for any errors vs no errors
- Simple validation scenarios

### Examples

**Explicit errorCount:**
```go
// Test OpenCode temperature validation
{
    name:       "temperature below minimum",
    input:      &models.Agent{Temperature: -0.5},
    errorCount: 1,
}
```

**Simple error check:**
```go
// Test basic validation
{
    name:    "missing required field",
    input:    invalidData,
    wantErr: true,  // Just check if error exists
}

// In test:
errs := Validate(input)
if (len(errs) > 0) != tt.wantErr {
    t.Errorf("Validate() error = %v, wantErr %v", errs, tt.wantErr)
}
```

**Location**: `test/`
**Parent**: See `/AGENTS.md` for project overview

---

# Testing Infrastructure

## Structure

- `fixtures/` - Test fixtures (Germinator format inputs)
- `golden/` - Golden files (expected outputs)

## Golden File Testing

### Usage

- **Test fixtures**: `test/fixtures/opencode/` contains Germinator format YAML files (inputs)
- **Golden files**: `test/golden/opencode/` contains expected OpenCode format outputs

### Running Tests

```bash
# All golden file tests
go test ./internal/services -run TestGoldenFiles -v

# Specific test
go test ./internal/services -run TestGoldenFiles/agent-full -v
```

### Updating Golden Files

```bash
UPDATE_GOLDEN=true go test ./internal/services -run TestGoldenFiles -v
```

Commit updated golden files. Verify without flag: `go test ./internal/services -run TestGoldenFiles -v`

### Adding New Golden File Tests

1. Create fixture in `test/fixtures/opencode/`
2. Generate output: `./bin/germinator adapt test/fixtures/opencode/new-fixture.md /tmp/output.md --platform opencode`
3. Copy to `test/golden/opencode/new-fixture.md.golden`
4. Add test case to `transformer_golden_test.go`

### Test Structure

Table-driven in `internal/services/transformer_golden_test.go`:
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

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {name: "descriptive test case name", input: "test input", expected: "expected output", wantErr: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got = %q, want %q", got, tt.expected)
            }
        })
    }
}
```

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

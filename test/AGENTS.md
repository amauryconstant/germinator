**Location**: `test/`
**Parent**: See `/AGENTS.md` for project overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/07-testing.md`

---

# Testing Infrastructure

## Structure

- `fixtures/` - Test fixtures (Germinator format inputs)
- `golden/` - Golden files (expected outputs)
- `e2e/` - E2E tests (CLI behavior validation)
- `helpers/` - Shared test utilities (future)

> The `test/mocks/` package was deprecated when the project migrated to `runF` injection with `iostreams.Test()` buffers (see `cmd/AGENTS.md` → Testing). New tests do not use mocks.

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

Located in `internal/parser/integration_test.go`:

- Tests end-to-end document loading pipeline: `LoadDocument → Validate → RenderDocument`
- Uses fixtures from `test/fixtures/` for representative document types
- Verifies document structure, file paths, and content extraction

### Adding Integration Tests

1. Create test functions in `internal/parser/integration_test.go` with `//go:build integration` tag
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
go test -tags=golden ./cmd -v                 # CLI golden tests
go test -tags=golden ./internal/opencode -v   # Adapter golden tests
```

### Updating Golden Files

```bash
UPDATE_GOLDEN=true mise run test:golden
```

Commit updated golden files. Verify without flag: `mise run test:golden`

### Adding New Golden File Tests

1. Create fixture in `test/fixtures/opencode/`
2. Generate output: `./bin/germinator adapt test/fixtures/opencode/new-fixture.md /tmp/output.md --platform opencode`
3. Copy to `test/golden/opencode/new-fixture.md.golden`
4. Add test case to the relevant `*_golden_test.go` (e.g. `cmd/canonicalize_golden_test.go`)

### Test Structure

**Build tag**: `//go:build golden` (excluded from `go test ./...`)

Table-driven in `cmd/canonicalize_golden_test.go` and adapter-level `*_golden_test.go` files:
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

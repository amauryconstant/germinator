# Test Support

This directory stores test fixtures for testing.

## Structure

- `fixtures/` - Test fixtures and test data (Germinator format inputs)
- `golden/` - Golden files (expected outputs for transformation tests)

## When to Add Files

Add files to these directories when writing tests:
- Place test data in `fixtures/` (Germinator format YAML files)
- Place expected outputs in `golden/` (for transformation tests)

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

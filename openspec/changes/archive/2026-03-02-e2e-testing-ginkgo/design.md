## Context

Germinator currently has unit tests and golden file tests but lacks E2E tests that validate actual CLI behavior. We're adopting Ginkgo v2 + Gomega + gexec patterns from twiggit, a proven approach for CLI testing that:
- Uses build tags to separate E2E tests from unit tests
- Builds a test binary before running tests
- Provides fluent assertions for CLI output and exit codes

Current state:
- Unit tests: `internal/**/*_test.go` using standard `testing` package
- Golden file tests: `internal/services/transformer_golden_test.go`
- Test fixtures: `test/fixtures/` with valid/invalid documents
- No E2E tests for CLI commands

## Goals / Non-Goals

**Goals:**
- Establish E2E testing infrastructure with Ginkgo v2, Gomega, and gexec
- Create test coverage for all CLI commands (validate, adapt, version, root)
- Enable isolated test runs via build tags (`//go:build e2e`)
- Provide CLI helper utilities for running germinator binary in tests
- Add mise tasks for running E2E tests

**Non-Goals:**
- Replacing existing unit tests or golden file tests
- Testing internal packages (E2E tests only test CLI surface)
- Performance/load testing
- Testing on Windows (focus on Linux/macOS first)

## Decisions

### 1. Use Ginkgo v2 + Gomega + gexec

**Choice**: Adopt the same testing stack as twiggit

**Rationale**: 
- Ginkgo provides BDD-style tests with BeforeEach/AfterEach setup
- Gomega offers fluent assertions (`Expect`, `Eventually`)
- gexec handles process lifecycle (building, running, cleanup)
- Proven pattern in twiggit codebase

**Alternatives considered**:
- Standard `testing` + `os/exec`: More verbose, less readable
- Testify: Good for unit tests, lacks CLI-specific helpers

### 2. Use `//go:build e2e` Build Tag

**Choice**: Separate E2E tests from unit tests via build tags

**Rationale**:
- E2E tests are slower and require building the binary
- CI can run unit tests on every commit, E2E tests on schedule/merge
- Matches twiggit pattern

### 3. Test Directory Structure

**Choice**: `test/e2e/` with subdirectories for helpers and fixtures

**Structure**:
```
test/e2e/
├── e2e_suite_test.go      # Suite setup, binary build
├── helpers/
│   └── cli_helper.go      # CLI runner utilities
├── fixtures/
│   └── document_fixtures.go  # Test fixture management
├── validate_test.go       # Validate command tests
├── adapt_test.go          # Adapt command tests
├── version_test.go        # Version command tests
└── root_test.go           # Root command tests
```

**Rationale**:
- Mirrors twiggit structure for familiarity
- Separates concerns (helpers, fixtures, test cases)
- Clear organization by command

### 4. Binary Naming

**Choice**: Build to `bin/germinator-e2e`

**Rationale**:
- Distinguishes from development binary (`bin/germinator`)
- gexec cleanup handles artifact removal
- Matches twiggit pattern (`bin/twiggit-e2e`)

### 5. Reuse Existing Fixtures

**Choice**: Use existing `test/fixtures/` files where possible

**Rationale**:
- Avoid duplication
- Existing fixtures already cover valid/invalid cases
- E2E fixtures can reference or copy from existing fixtures

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| E2E tests are slower than unit tests | Use build tags to run separately; keep tests focused |
| Binary build adds overhead | Build once in BeforeSuite; reuse across tests |
| Flaky tests due to timing | Use Gomega's `Eventually` for async assertions |
| Test fixtures drift from real usage | Reuse existing fixtures; add E2E-specific fixtures only when needed |

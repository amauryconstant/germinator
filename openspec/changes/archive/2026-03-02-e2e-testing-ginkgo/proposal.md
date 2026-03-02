## Why

Germinator has unit tests and golden file tests but no E2E tests that validate actual CLI behavior. Without E2E tests, we cannot verify that commands work correctly when invoked through the actual binary, including flag parsing, exit codes, and output formatting. Adopting Ginkgo/Gomega from twiggit provides battle-tested patterns for CLI testing.

## What Changes

- Add Ginkgo v2, Gomega, and gexec dependencies to `go.mod`
- Create `test/e2e/` directory with suite setup using `//go:build e2e` build tag
- Implement CLI helper for building and running germinator binary in tests
- Add fixture management for test documents (valid/invalid YAML files)
- Create E2E test cases for all commands (validate, adapt, version, root)
- Add mise tasks for running E2E tests (`test:e2e`, `test:full`)

## Capabilities

### New Capabilities

- `e2e-testing`: End-to-end CLI testing infrastructure with Ginkgo v2, Gomega, and gexec. Covers test suite setup, CLI helper utilities, fixture management, and test cases for validate, adapt, version, and root commands.

### Modified Capabilities

## Impact

- **New files**: `test/e2e/` directory with suite, helpers, fixtures, and test files
- **Dependencies**: ginkgo/v2, gomega, gexec added to `go.mod`
- **Build**: E2E tests excluded from default `go test ./...` via build tag
- **Tasks**: `.mise.toml` updated with `test:e2e` and `test:full` tasks
- **No changes** to production code

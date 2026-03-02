## Why

The `canonicalize` command has no E2E test coverage, leaving a gap in CLI validation. Additionally, existing E2E tests only exercise the `opencode` platform, violating the project convention that all platform tests must cover both `opencode` and `claude-code`.

## What Changes

- Add E2E test file for `canonicalize` command covering success, missing flags, invalid inputs, and error cases
- Add `claude-code` platform variants to existing `validate` and `adapt` E2E tests
- No changes to production code - test files only

## Capabilities

### New Capabilities

- `e2e-canonicalize-tests`: E2E test coverage for the canonicalize command, including valid document conversion, missing required flags (`--platform`, `--type`), invalid platform/type values, and nonexistent input files.

### Modified Capabilities

- `e2e-testing`: Extend validate and adapt command test requirements to include `claude-code` platform scenarios, ensuring both supported platforms are tested per project conventions.

## Impact

- **New files**: `test/e2e/canonicalize_test.go`
- **Modified files**: `test/e2e/validate_test.go`, `test/e2e/adapt_test.go` (add claude-code scenarios)
- **Dependencies**: Uses existing Ginkgo v2, Gomega, gexec infrastructure
- **No production code changes**

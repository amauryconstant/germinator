## Why

Comprehensive linting catches more issues earlier in the development cycle, reducing bugs and maintaining code quality. The current configuration uses only 8 linters, missing many potential issues that 25 linters would catch. Domain purity enforcement via depguard ensures architectural integrity by preventing external dependencies from leaking into the domain layer. This aligns with CLI tooling standards established in cross-project investigations.

**Prerequisite:** This change depends on `domain-restructure` which creates the `internal/domain/` package. Apply this change after domain-restructure is complete.

## What Changes

- Expand golangci-lint from 8 to 25 linters organized by category:
  - Essential: `staticcheck`, `unused`
  - Code Quality: `gocyclo`, `gocognit`, `funlen`
  - Style: `misspell`, `whitespace`, `revive`
  - Error Handling: `errorlint`, `wrapcheck`, `errname`
  - Performance: `prealloc`, `perfsprint`
  - Security: `gosec`
  - Tests: `testifylint`, `tparallel`, `thelper`
  - Best Practices: `nakedret`, `unconvert`, `unparam`, `wastedassign`
  - Architecture: `depguard`
- Add domain purity enforcement via depguard (no external deps in `internal/domain/`)
- Configure appropriate thresholds for complexity linters
- Add exclusions for test files and common false positives

## Capabilities

### New Capabilities

- `comprehensive-linting`: 25+ linter configuration with depguard domain purity enforcement

## Impact

**Affected Files**:
- `.golangci.yml` - Major expansion from 8 to 25+ linters
- Potentially all Go files - may require fixes for newly enabled linter errors

**Affected Directories**:
- All `internal/` packages - may surface existing issues that need fixing

**No Public API Impact**: All changes are to linting configuration; no runtime behavior changes.

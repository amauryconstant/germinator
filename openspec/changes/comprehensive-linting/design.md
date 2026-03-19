## Context

**Current State:**
- 8 linters in `.golangci.yml`: `errcheck`, `govet`, `ineffassign`, `unused`, `misspell`, `whitespace`, `revive`, `depguard`
- Domain purity enforcement via depguard (for `internal/domain/`)
- No complexity thresholds
- No organized linter categories

**Prerequisite:** The `domain-restructure` change created the `internal/domain/` package and is complete. This change expands the linting configuration but does not depend on any other pending changes.

**Target State:**
- 25 linters organized by category
- Enhanced depguard rules for comprehensive architecture enforcement
- Appropriate complexity thresholds (gocyclo: 25, gocognit: 30, funlen: 150 lines/100 statements)
- Test file exclusions for complexity linters
- GoSec exclusions for CLI tool false positives

**Constraints:**
- All linter fixes must maintain existing functionality
- No changes to business logic or transformation behavior
- Test files should have reasonable exclusions

## Goals / Non-Goals

**Goals:**
- Expand linting to 25 linters matching CLI standard
- Enhance domain purity enforcement with additional checks
- Configure appropriate thresholds and exclusions
- Fix all surfaced linting errors

**Non-Goals:**
- No changes to business logic or transformation behavior
- No refactoring beyond what linting errors require
- No changes to CLI flags or user-facing behavior

## Decisions

### DEC-001: Domain Purity Enforcement

**Choice:** Use depguard linter to enforce no external dependencies in `internal/domain/`

**Rationale:** Prevents architectural erosion over time. The domain layer is the foundation - protecting it pays dividends. Automated enforcement is more reliable than code review alone.

**Configuration:**
```yaml
depguard:
  rules:
    domain:
      files:
        - internal/domain/**
      allow:
        - $gostd
        - internal/domain
```

**Alternatives Considered:**
- Convention only (document in AGENTS.md) → Rejected: can drift over time
- Manual code review → Rejected: unreliable, easy to miss

### DEC-002: Linter Expansion

**Choice:** 25+ linters matching CLI standard

**Linter Categories and Additions (17 new linters, 25 total):**

| Category | Linters |
|----------|---------|
| Essential | `staticcheck` (NEW), `typecheck` (NEW), `unused` (already enabled) |
| Code Quality | `gocyclo`, `gocognit`, `funlen` |
| Style | `misspell` (already enabled), `whitespace` (already enabled), `revive` (already enabled) |
| Error Handling | `errorlint`, `wrapcheck`, `errname` |
| Performance | `prealloc`, `perfsprint` |
| Security | `gosec` |
| Tests | `testifylint`, `tparallel`, `thelper` |
| Best Practices | `nakedret`, `unconvert`, `unparam`, `wastedassign` |
| Architecture | `depguard` (already enabled) |

**Thresholds:**
- `gocyclo`: min-complexity: 25
- `gocognit`: min-complexity: 30
- `funlen`: lines: 150, statements: 100

**Rationale:** Comprehensive linting catches more issues earlier. The linters are well-organized by category with appropriate thresholds.

**Alternatives Considered:**
- Keep minimal 8 linters → Rejected: misses many code quality issues
- Incremental addition → Rejected: prefer clean cutover

## Risks / Trade-offs

### Risk: Linter Errors Block Progress
**Impact:** 25+ linters may surface many issues, delaying completion
**Mitigation:** Run `golangci-lint run` early to assess scope; prioritize fixing over `//nolint`

### Risk: False Positives from GoSec
**Impact:** GoSec may flag valid CLI tool patterns as security issues
**Mitigation:** Configure exclusions for G104, G204, G301, G304, G306, G307, G401, G501

### Risk: Complexity Thresholds Too Strict
**Impact:** Functions may exceed thresholds and require refactoring
**Mitigation:** Set thresholds at reasonable levels (25/30/150) that allow complex but readable code

## Implementation Plan

**Sequence:**
1. Update `.golangci.yml` with new linters and configuration
2. Run `golangci-lint run` and categorize all errors
3. Fix errors by category (essential, code quality, error handling, etc.)
4. Verify `golangci-lint run` passes cleanly
5. Update documentation

**Rollback:** Revert `.golangci.yml` changes if linters cause blocking issues.

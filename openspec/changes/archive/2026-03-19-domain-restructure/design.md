## Context

**Current State:**
- Domain types scattered across 4 packages: `models/canonical/`, `errors/`, `validation/`, `application/` (requests/results)
- No enforcement of domain purity
- 10 internal packages total

**Target State:**
- Domain types consolidated into single `internal/domain/` package
- Domain purity enforced via depguard (no external dependencies)
- Foundation for future 4-package DDD-light structure

**Constraints:**
- All changes are internal packages - no public API breakage concerns
- Existing tests must continue to pass (with updated import paths)
- Single change (all-at-once), not phased migration

## Goals / Non-Goals

**Goals:**
- Consolidate domain types into `internal/domain/` package
- Establish domain purity via depguard enforcement
- Maintain all existing functionality and test coverage
- Reduce cognitive overhead by grouping related domain types

**Non-Goals:**
- Infrastructure layer restructuring (separate change)
- Service renaming (separate change)
- CLI error handling pattern changes (separate change)
- Mock infrastructure (separate change)
- Linter expansion beyond domain purity (separate change)

## Decisions

### DEC-001: Package Structure (Domain Portion)

**Choice:** Consolidate domain types into single `internal/domain/` package

**Rationale:** Grouping all domain types (models, errors, validation, requests, results) provides a clear architectural boundary. The domain layer becomes the single source of truth for business types.

**Mapping:**
```
Current                         →  Target
───────────────────────────────────────────
internal/models/canonical/      →  internal/domain/
internal/errors/                →  internal/domain/errors.go
internal/validation/validators.go →  internal/domain/validation.go
internal/validation/result.go   →  internal/domain/result.go
internal/validation/opencode/   →  internal/domain/opencode/ (subdirectory preserved)
internal/application/results.go →  internal/domain/results.go
```

**Exception:** `internal/application/requests.go` stays in application layer because `InitializeRequest` depends on `*library.Library` (infrastructure concern). Only pure DTOs move to domain.

**File Organization in domain/:**
```
internal/domain/
├── doc.go           # Package documentation
├── agent.go         # Agent type
├── command.go       # Command type
├── skill.go         # Skill type
├── memory.go        # Memory type
├── platform.go      # Platform types
├── errors.go        # Domain errors
├── validation.go    # Validation types (validators)
├── result.go        # Result[T] type
├── results.go       # Result types (service results)
└── opencode/        # OpenCode-specific validators (subdirectory)
    ├── validators.go
    └── validators_test.go
```

### DEC-002: Domain Purity Enforcement

**Choice:** Use depguard linter to enforce no external dependencies in `domain/`

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

## Risks / Trade-offs

### Risk: Import Path Changes Break Tests
**Impact:** All test files with domain imports need import path updates
**Mitigation:** Single find/replace pass after package moves; verify with `go build ./... && go test ./...`

### Risk: Domain Purity Requires Refactoring
**Impact:** Moving validation/errors to domain might reveal external dependencies
**Mitigation:** Both packages currently use only stdlib; verify before moving

### Trade-off: Smaller Scope Than Full Alignment
**Impact:** This change only addresses domain layer, leaving other restructuring for later
**Mitigation:** Clean separation allows future changes to build on this foundation

## Migration Plan

**Sequence:**
1. Create `internal/domain/` directory
2. Move domain files to new locations (git mv for history)
3. Split models.go into type-specific files
4. Update import paths (find/replace)
5. Verify compilation (`go build ./...`)
6. Move test files
7. Update test import paths
8. Verify tests pass (`go test ./...`)
9. Add depguard rule to `.golangci.yml`
10. Verify linting passes
11. Remove old empty directories
12. Update documentation

**Rollback:** Not applicable - single atomic change. Git revert if needed.

## Open Questions

None - scope is well-defined and limited to domain layer.

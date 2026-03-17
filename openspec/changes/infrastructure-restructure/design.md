## Context

**Current State:**
- Infrastructure code scattered across 4 packages: `core/`, `adapters/`, `config/`, `library/`
- Services package uses plural naming (`services/`)
- No clear infrastructure layer organization

**Target State:**
- Unified `infrastructure/` package with subdirectories for each concern
- Singular service package naming (`service/`)
- Clear layer separation following DDD-light pattern

**Constraints:**
- All changes are internal packages - no public API breakage concerns
- Existing tests must continue to pass (with updated import paths)
- Single change (all-at-once), not phased migration

## Goals / Non-Goals

**Goals:**
- Consolidate infrastructure packages under unified `infrastructure/` layer
- Adopt singular naming convention for service package
- Maintain all existing functionality and test coverage

**Non-Goals:**
- No changes to business logic or transformation behavior
- No changes to domain types or validation logic
- No changes to application layer interfaces

## Decisions

### DEC-001: Package Structure (Infrastructure Portion)

**Choice:** Move infrastructure packages under `internal/infrastructure/`

**Rationale:** The `infrastructure/` naming is clearer than scattered packages for external integrations, configuration, and I/O operations. Provides better organization and discoverability.

**Mapping (Infrastructure Only):**
```
Current                         →  Target
────────────────────────────────────────────────
internal/core/                  →  internal/infrastructure/parsing/ + serialization/
internal/adapters/              →  internal/infrastructure/adapters/
internal/config/                →  internal/infrastructure/config/
internal/library/               →  internal/infrastructure/library/
internal/services/              →  internal/service/
```

**Adapter Files:**
- `internal/adapters/adapter.go` → `internal/infrastructure/adapters/adapter.go`
- `internal/adapters/helpers.go` → `internal/infrastructure/adapters/helpers.go`
- `internal/adapters/claude-code/` → `internal/infrastructure/adapters/claude-code/`
- `internal/adapters/opencode/` → `internal/infrastructure/adapters/opencode/`

### DEC-002: Interface Location

**Choice:** Mixed location pattern - service contracts in `application/`, adapter interface in `infrastructure/adapters/`

**Rationale:** Interfaces belong to the layer that defines the contract. Application service interfaces (Transformer, Validator, Canonicalizer, Initializer) stay in `application/`. The Adapter interface moves with its implementations to `infrastructure/adapters/`.

**Alternatives Considered:**
- All interfaces in `application/` → Rejected: adapter interface is infrastructure concern
- All interfaces in `domain/` → Rejected: service contracts are application-level

## Risks / Trade-offs

### Risk: Import Path Changes Break Tests
**Impact:** All test files importing infrastructure packages need import path updates
**Mitigation:** Single find/replace pass after package moves; verify with `go build ./... && go test ./...`

### Risk: Large Number of Files Affected
**Impact:** Import path changes touch most files in the codebase
**Mitigation:** Use git mv for history preservation; systematic verification at each step

### Trade-off: Larger Change Size
**Impact:** Single change touches many files, harder to review incrementally
**Mitigation:** Logical grouping in commits; all changes are structural not behavioral

## Migration Plan

**Sequence:**
1. Create new infrastructure package structure (empty directories)
2. Move files to new locations (git mv for history)
3. Update import paths (find/replace)
4. Rename services/ to service/
5. Verify compilation (`go build ./...`)
6. Move and update test files
7. Cleanup old directories
8. Verify tests pass (`go test ./...`)

**Rollback:** Not applicable - single atomic change. Git revert if needed.

## Open Questions

None - scope is well-defined to infrastructure reorganization only.

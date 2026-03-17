## Why

An investigation comparing germinator and twiggit established unified Go CLI standards. Germinator currently diverges from these standards in architecture (10 packages vs 4). This change focuses specifically on establishing a clean domain layer with DDD-light architecture, domain purity, and reduced cognitive overhead.

Adopting a dedicated `internal/domain/` package isolates core business types (canonical models, errors, validation, requests, results) from infrastructure concerns, providing a clear separation that prevents architectural drift.

## What Changes

- **BREAKING**: Move `models/canonical/` into `domain/` package with files split by type
- **BREAKING**: Move `errors/` into `domain/errors.go`
- **BREAKING**: Move `validation/` into `domain/validation.go`
- **BREAKING**: Move `application/requests.go` and `application/results.go` into `domain/`
- Add domain purity enforcement via depguard (no external deps in domain/)

## Capabilities

### New Capabilities

- `domain-purity`: depguard enforcement ensuring `internal/domain/` has no external dependencies (stdlib and internal/domain only)

## Impact

**Affected Code**:
- All files with imports to `internal/models/canonical`, `internal/errors`, `internal/validation`
- All files with imports to `internal/application` for requests/results types
- Corresponding test files

**Affected Directories**:
- `internal/domain/` - New package (created)
- `internal/models/canonical/` - Removed (moved to domain/)
- `internal/errors/` - Removed (moved to domain/)
- `internal/validation/` - Removed (moved to domain/)
- `internal/application/` - Reduced (requests/results moved to domain/)
- `.golangci.yml` - Add depguard rule for domain purity

**No Public API Impact**: All changes are to internal packages; no external consumers affected.

**No New Dependencies**: Uses existing depguard linter.

## Why

Adopt DDD-light 4-package structure with clearer infrastructure organization. The current scattered infrastructure packages (core/, adapters/, config/, library/) are better organized under a unified infrastructure/ layer, improving navigation and architectural clarity.

## What Changes

- **BREAKING**: Move `core/` (parsing, serialization) into `infrastructure/parsing/` and `infrastructure/serialization/`
- **BREAKING**: Move `adapters/` into `infrastructure/adapters/`
- **BREAKING**: Move `config/` into `infrastructure/config/`
- **BREAKING**: Move `library/` into `infrastructure/library/`
- **BREAKING**: Rename `services/` → `service/` to match standard naming convention

## Impact

**Affected Code**:
- All infrastructure-related Go files (import path changes)
- All files importing infrastructure packages (import path updates)
- All infrastructure test files

**Affected Directories**:
- `internal/core/` → `internal/infrastructure/parsing/` + `internal/infrastructure/serialization/`
- `internal/adapters/` → `internal/infrastructure/adapters/`
- `internal/config/` → `internal/infrastructure/config/`
- `internal/library/` → `internal/infrastructure/library/`
- `internal/services/` → `internal/service/`

**No Public API Impact**: All changes are to internal packages; no external consumers affected.

## Capabilities

### New Capabilities

- `infrastructure-structure`: Unified infrastructure package organization with DDD-light naming

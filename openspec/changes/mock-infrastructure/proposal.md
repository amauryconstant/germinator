## Why

Unit tests currently lack isolation, making it difficult to test command handlers and application logic independently of real implementations. The testing standard established in the investigation requires testify/mock implementations for all application interfaces. Adding mock infrastructure enables better unit test isolation and follows proven patterns.

## What Changes

- Add testify/mock implementations for all application interfaces
- Create `test/mocks/` directory with mock files for Transformer, Validator, Canonicalizer, and Initializer
- Create `test/helpers/` directory for shared test utilities

## Capabilities

### New Capabilities

- `mock-infrastructure`: testify/mock implementations for Transformer, Validator, Canonicalizer, Initializer interfaces

## Impact

**Affected Code**:
- All test files may optionally use new mocks
- New `test/mocks/` directory with 5 files (4 mocks + doc.go)
- New `test/helpers/` directory with shared utilities

**Affected Directories**:
- `test/` - New mocks/ and helpers/ directories

**No Public API Impact**: All changes are to test infrastructure; no production code affected.

**Dependencies**: Add `github.com/stretchr/testify/mock` for mock infrastructure

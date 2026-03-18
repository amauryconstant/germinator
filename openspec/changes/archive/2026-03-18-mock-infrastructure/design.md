## Context

**Current State:**
- No mock infrastructure - tests use real implementations with fixtures
- Unit tests cannot easily isolate command handlers from service implementations
- Integration tests and golden file tests work well but lack unit test isolation

**Target State:**
- testify/mock implementations for all application interfaces
- Unit tests can use mocks for interface isolation
- Mocks coexist with existing golden file and integration tests

**Constraints:**
- Existing tests must continue to pass unchanged
- Mocks are optional - tests choose whether to use mocks or real implementations
- Minimal production code changes (test-only infrastructure)

**Reference Documents:**
- `/home/amaury/Projects/go-cli-toolkit/investigation/phase3-alignment/testing-standard.md`

## Goals / Non-Goals

**Goals:**
- Add testify/mock infrastructure for all application interfaces
- Enable isolated unit testing without real implementations
- Maintain all existing test coverage and patterns

**Non-Goals:**
- No refactoring of existing tests to use mocks (optional migration)
- No changes to golden file tests
- No changes to business logic or production code

## Decisions

### DEC-005: Mock Infrastructure

**Choice:** testify/mock with hand-written mocks in `test/mocks/`

**Rationale:** Hand-written mocks provide full control and type safety. The testify/mock pattern is well-understood and provides good assertion capabilities. For 4 interfaces, maintenance burden is manageable.

**Mocks Required:**
- `MockTransformer` - implements `application.Transformer`
- `MockValidator` - implements `application.Validator`
- `MockCanonicalizer` - implements `application.Canonicalizer`
- `MockInitializer` - implements `application.Initializer`

**Alternatives Considered:**
- No mocks (real implementations only) → Rejected: limits unit test isolation
- Generated mocks (mockery/mockgen) → Rejected: additional tooling, less control
- Keep current approach + add mocks → Chosen: both coexist for different test types

## Risks / Trade-offs

### Risk: Mock Maintenance Burden
**Impact:** Mocks must be updated when interfaces change
**Mitigation:** Only 4 interfaces to maintain; interface changes are infrequent; testify/mock pattern is straightforward to update

### Trade-off: Optional Mock Usage
**Impact:** Tests won't automatically benefit from mocks
**Mitigation:** Document mock usage patterns in test/AGENTS.md including: mock setup with `On()`, assertions with `AssertCalled()` and `AssertNumberOfCalls()`; provide example test in cmd/validate_test.go

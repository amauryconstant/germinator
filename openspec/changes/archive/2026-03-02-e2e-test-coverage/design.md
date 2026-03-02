## Context

The E2E testing infrastructure is established (Ginkgo v2, Gomega, gexec) but has coverage gaps:
- The `canonicalize` command has no E2E tests
- Existing tests only exercise `opencode` platform, violating project convention

Current state:
- `test/e2e/validate_test.go` - 5 scenarios, all `opencode`
- `test/e2e/adapt_test.go` - 3 scenarios, all `opencode`
- `test/e2e/canonicalize_test.go` - does not exist

## Goals / Non-Goals

**Goals:**
- Add E2E test coverage for `canonicalize` command
- Add `claude-code` platform variants for validate and adapt tests
- Follow existing test patterns and conventions

**Non-Goals:**
- Adding new test infrastructure (uses existing helpers/fixtures)
- Testing internal packages
- Performance testing

## Decisions

### 1. Canonicalize Test Structure

**Choice**: Mirror existing command test patterns

**Structure**:
```
test/e2e/canonicalize_test.go
├── "canonicalizing a valid document" (success case)
├── "canonicalizing without platform flag" (error case)
├── "canonicalizing without type flag" (error case)
├── "canonicalizing with invalid platform" (error case)
├── "canonicalizing with invalid type" (error case)
└── "canonicalizing a nonexistent file" (error case)
```

**Rationale**: Consistency with validate_test.go and adapt_test.go patterns

### 2. Platform Parity Approach

**Choice**: Add parallel test cases for `claude-code` in existing test files

**Rationale**:
- Minimal diff - add new Describe blocks alongside existing ones
- Easy to compare platform behaviors
- Follows project convention in test/AGENTS.md

**Alternatives considered**:
- Table-driven with platform as parameter: Would require restructuring existing tests
- Separate files per platform: Fragmented, harder to maintain parity

### 3. Fixture Reuse

**Choice**: Use existing `test/fixtures/` documents for canonicalize tests

**Rationale**:
- Existing fixtures cover valid/invalid cases
- Canonicalize can consume same fixtures as validate/adapt
- No new fixture files needed

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Duplicate test code between platforms | Acceptable - platform-specific failures need clear diagnosis |
| Test runtime increases with more scenarios | Tests are fast (~0.3s for 11 tests); acceptable overhead |
| Canonicalize may have unique edge cases | Add scenarios as discovered during implementation |

# testing-iostreams-injection Specification (delta)

## MODIFIED Requirements

### Requirement: t.Context() pattern for new tests

New tests added in `cmd/`, `internal/`, or `test/` SHALL use `t.Context()` (Go 1.24+) for tests that need a `context.Context`. The `t.Context()` is auto-cancelled when the test ends, eliminating the need for explicit `defer cancel()` calls and preventing goroutine leaks.

**Change**: NEW requirement. Pre-change tests use `context.Background()` as the minimum-churn approach; new tests adopt the `t.Context()` pattern. The migration of existing tests is a follow-up.

#### Scenario: New test uses t.Context()

- **WHEN** a new test is added to any package
- **AND** the test needs a `context.Context` for I/O calls
- **THEN** the test SHALL use `t.Context()` instead of `context.Background()`
- **AND** the test SHALL NOT call `defer cancel()` (the context is auto-cancelled)

#### Scenario: Existing tests may use context.Background()

- **WHEN** an existing test is migrated to the new pattern
- **THEN** the test MAY continue to use `context.Background()` until the next refactor
- **AND** new tests added in the same file SHALL use `t.Context()` to establish the pattern

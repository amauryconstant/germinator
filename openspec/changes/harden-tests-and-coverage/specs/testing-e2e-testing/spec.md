# testing-e2e-testing Specification (delta)

## ADDED Requirements

### Requirement: t.Parallel() safety in cmd tests

The `t.Parallel()` annotation in cmd-side tests SHALL only be used when the test does not call `cmd.Execute()` (which mutates Cobra's package-level `OnInitialize` slice). Tests that call `cmd.Execute()` SHALL NOT be marked `t.Parallel()`; the race detector flags concurrent reads/writes on Cobra globals.

**Change**: NEW requirement. The pre-change `cmd/lint_test.go:19` used `t.Parallel()` while calling `exec.Command("mise", "run", "lint")` 8 times, racing on Cobra globals. The hotfix in change `harden-tests-and-coverage` removes the `t.Parallel()` from `TestLintBaseline`.

#### Scenario: cmd.Execute() tests are sequential

- **WHEN** a test calls `cmd.Execute()` (or any function that calls Cobra's `AddCommand` / `OnInitialize` machinery)
- **THEN** the test SHALL NOT use `t.Parallel()`
- **AND** sibling tests in the same package SHALL run sequentially with respect to it

> **Note (non-normative):** tests that shell out via `exec.Command(...)` to
> subprocesses (e.g., `cmd/lint_test.go:TestLintBaseline` shells out to
> `mise run lint`; `cmd/lint_test.go:TestNoNewForbidigoPatterns` shells out
> to `go list ...`) are NOT in-process callers of Cobra's `AddCommand` /
> `OnInitialize` machinery, and therefore remain eligible for `t.Parallel()`.
> The pre-change `tasks.md:3` D-001 hotfix claim is reconciled by this note
> (Phase 0 was a documentation error, not a code change).

#### Scenario: Race detector passes

- **WHEN** `go test -race -count=1 ./...` is run
- **THEN** zero race conditions SHALL be reported on Cobra globals
- **AND** zero race conditions SHALL be reported on `t.Setenv` / `os.Chdir` / `t.TempDir` usage

#### Scenario: Non-Cmd tests use t.Parallel() safely

- **WHEN** a test does not call `cmd.Execute()` and does not share `t.Setenv` / `os.Chdir` with sibling tests
- **THEN** the test SHALL be safe to run concurrently with other `t.Parallel()` tests
- **AND** the test SHALL be marked `t.Parallel()` in subtests inside `t.Run(...)` blocks to opt into parallel execution

### Requirement: t.Context() adoption for new tests

New tests added in `cmd/`, `internal/`, or `test/` SHALL use `t.Context()` (Go 1.24+) for tests that need a `context.Context`. Every new test function (regardless of whether it does I/O) SHALL adopt the `t.Context()` pattern. The `t.Context()` is auto-cancelled when the test ends, eliminating explicit `defer cancel()` calls and preventing goroutine leaks.

**Change**: NEW requirement. Pre-change tests use `context.Background()` as the minimum-churn approach; new tests adopt the `t.Context()` pattern universally. The migration of existing tests is a follow-up tracked separately.

#### Scenario: New test uses t.Context()

- **WHEN** a new test is added to any package
- **THEN** the test SHALL use `t.Context()` instead of `context.Background()`
- **AND** the test SHALL NOT call `defer cancel()` (the context is auto-cancelled)
- **AND** the `ctx` returned by `t.Context()` SHALL be forwarded to every function under test that accepts `context.Context`

#### Scenario: Existing tests transition over time

- **WHEN** an existing test is migrated to the new pattern
- **THEN** the test MAY continue to use `context.Background()` until the next refactor
- **AND** new tests added in the same file SHALL use `t.Context()` to establish the pattern
- **WHEN** a new test is added to a file that calls `LoadLibrary(ctx, ...)` or similar context-taking function
- **THEN** the new test MUST use `t.Context()` if the function is called from the new test's body
- **AND** if the new test exercises an existing path that uses `context.Background()`, the existing path remains unchanged until refactored

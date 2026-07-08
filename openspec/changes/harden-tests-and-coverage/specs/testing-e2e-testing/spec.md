# testing-e2e-testing Specification (delta)

## ADDED Requirements

### Requirement: t.Parallel() safety in cmd tests

The `t.Parallel()` annotation in cmd-side tests SHALL only be used when the test does not call `cmd.Execute()` (which mutates Cobra's package-level `OnInitialize` slice). Tests that call `cmd.Execute()` SHALL NOT be marked `t.Parallel()`; the race detector flags concurrent reads/writes on Cobra globals.

**Change**: NEW requirement. The pre-change `cmd/lint_test.go:19` used `t.Parallel()` while calling `exec.Command("mise", "run", "lint")` 8 times, racing on Cobra globals. The hotfix in change `harden-tests-and-coverage` removes the `t.Parallel()` from `TestLintBaseline`.

#### Scenario: cmd.Execute() tests are sequential

- **WHEN** a test calls `cmd.Execute()` (or any function that calls Cobra's `AddCommand` / `OnInitialize` machinery)
- **THEN** the test SHALL NOT use `t.Parallel()`
- **AND** sibling tests in the same package SHALL run sequentially with respect to it

#### Scenario: Race detector passes

- **WHEN** `go test -race -count=1 ./...` is run
- **THEN** zero race conditions SHALL be reported on Cobra globals
- **AND** zero race conditions SHALL be reported on `t.Setenv` / `os.Chdir` / `t.TempDir` usage

#### Scenario: Non-Cmd tests may use t.Parallel()

- **WHEN** a test does not call `cmd.Execute()` and does not share `t.Setenv` / `os.Chdir` with sibling tests
- **THEN** the test MAY use `t.Parallel()`
- **AND** the test SHALL be safe to run concurrently with other `t.Parallel()` tests

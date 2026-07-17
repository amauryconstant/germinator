# E2E Testing Specification

## Purpose

End-to-end testing infrastructure using Ginkgo v2, Gomega, and gexec to validate the germinator CLI commands through actual binary execution.
## Requirements
### Requirement: E2E Test Suite Setup

The E2E test suite SHALL be configured with Ginkgo v2, Gomega, and gexec using the `//go:build e2e` build tag.
#### Scenario: Suite initializes successfully

- **WHEN** the E2E test suite runs
- **THEN** the germinator test binary SHALL be built to a per-suite temp directory via `gexec.Build` (e.g., `$TMPDIR/ginkgo-germinator-<random>/germinator`)
- **AND** the binary SHALL be available for all test cases
- **AND** the binary SHALL be cleaned up after all tests complete (gexec owns the lifecycle)

#### Scenario: Build tag excludes E2E tests from default test run
- **WHEN** `go test ./...` is executed
- **THEN** E2E tests SHALL NOT run
- **AND** only unit tests SHALL execute

#### Scenario: Build tag includes E2E tests when specified
- **WHEN** `go test -tags=e2e ./test/e2e/...` is executed
- **THEN** all E2E tests SHALL run

---

### Requirement: CLI Helper for Running Germinator

A CLI helper SHALL provide utilities for running the germinator binary in tests.

#### Scenario: Run command returns session
- **WHEN** `cli.Run(args...)` is called with command arguments
- **THEN** a gexec.Session SHALL be returned
- **AND** the session SHALL capture stdout and stderr

#### Scenario: Assert successful execution
- **WHEN** `cli.ShouldSucceed(session)` is called after a successful command
- **THEN** the assertion SHALL pass if exit code is 0

#### Scenario: Assert failed execution with exit code
- **WHEN** `cli.ShouldFailWithExit(session, code)` is called
- **THEN** the assertion SHALL pass if exit code matches

#### Scenario: Assert stdout output
- **WHEN** `cli.ShouldOutput(session, expected)` is called
- **THEN** the assertion SHALL pass if stdout contains the expected string

#### Scenario: Assert stderr output
- **WHEN** `cli.ShouldErrorOutput(session, expected)` is called
- **THEN** the assertion SHALL pass if stderr contains the expected string

---

### Requirement: Test Fixture Management

Test fixtures SHALL provide valid and invalid document files for testing.

#### Scenario: Valid document fixture exists
- **WHEN** a test needs a valid document
- **THEN** a valid canonical YAML fixture SHALL be available

#### Scenario: Invalid document fixture exists
- **WHEN** a test needs an invalid document
- **THEN** an invalid document fixture SHALL be available

#### Scenario: Nonexistent file path
- **WHEN** a test needs to test file-not-found errors
- **THEN** a nonexistent file path SHALL be available

---

### Requirement: Validate Command E2E Tests

The validate command SHALL be tested for all expected behaviors.

#### Scenario: Validate valid document succeeds
- **WHEN** `germinator validate <valid-doc> --platform opencode` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "Document is valid"

#### Scenario: Validate with missing platform flag fails
- **WHEN** `germinator validate <doc>` is executed without `--platform`
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain "required" or "platform"

#### Scenario: Validate nonexistent file fails
- **WHEN** `germinator validate nonexistent.yaml --platform opencode` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

#### Scenario: Validate with invalid platform fails
- **WHEN** `germinator validate <doc> --platform invalid` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL indicate the platform is invalid or unknown

#### Scenario: Validate invalid document fails
- **WHEN** `germinator validate <invalid-doc> --platform opencode` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain validation errors

---

### Requirement: Adapt Command E2E Tests

The adapt command SHALL be tested for all expected behaviors.

#### Scenario: Adapt document succeeds
- **WHEN** `germinator adapt <valid-doc> <output> --platform opencode` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "transformed successfully"
- **AND** output file SHALL be created

#### Scenario: Adapt with missing platform flag fails
- **WHEN** `germinator adapt <doc> <output>` is executed without `--platform`
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain "required" or "platform"

#### Scenario: Adapt nonexistent file fails
- **WHEN** `germinator adapt nonexistent.yaml <output> --platform opencode` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

---

### Requirement: Validate Command Platform Parity

The validate command SHALL be tested for both supported platforms.

#### Scenario: Validate valid document succeeds with claude-code platform
- **WHEN** `germinator validate <valid-doc> --platform claude-code` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "Document is valid"

#### Scenario: Validate nonexistent file fails with claude-code platform
- **WHEN** `germinator validate nonexistent.yaml --platform claude-code` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

#### Scenario: Validate invalid document fails with claude-code platform
- **WHEN** `germinator validate <invalid-doc> --platform claude-code` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain validation errors

---

### Requirement: Adapt Command Platform Parity

The adapt command SHALL be tested for both supported platforms.

#### Scenario: Adapt document succeeds with claude-code platform
- **WHEN** `germinator adapt <valid-doc> <output> --platform claude-code` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "transformed successfully"
- **AND** output file SHALL be created

#### Scenario: Adapt nonexistent file fails with claude-code platform
- **WHEN** `germinator adapt nonexistent.yaml <output> --platform claude-code` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

---

### Requirement: Version Command E2E Tests

The version command SHALL be tested for expected output.

#### Scenario: Version displays version info
- **WHEN** `germinator version` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL match pattern `germinator <version> (<commit>) <date>`

---

### Requirement: Root Command E2E Tests

The root command SHALL be tested for help display.

#### Scenario: Root command shows help
- **WHEN** `germinator` is executed without arguments
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain usage information

#### Scenario: Help flag shows help
- **WHEN** `germinator --help` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain usage information

---

### Requirement: Init Command E2E Tests

The init command SHALL be tested for all expected behaviors.

#### Scenario: Init with dry-run preview
- **WHEN** `germinator init --platform opencode --resources skill/commit --dry-run` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL show preview of changes
- **AND** no files SHALL be created

#### Scenario: Init with force overwrite
- **GIVEN** an existing output file
- **WHEN** `germinator init --platform opencode --resources skill/commit --force` is executed
- **THEN** exit code SHALL be 0
- **AND** existing file SHALL be overwritten

#### Scenario: Init fails without force when file exists
- **GIVEN** an existing output file
- **WHEN** `germinator init --platform opencode --resources skill/commit` is executed without `--force`
- **THEN** exit code SHALL be > 0
- **AND** stderr SHALL indicate file exists

#### Scenario: Init with preset expands resources
- **WHEN** `germinator init --platform opencode --preset git-workflow --dry-run` is executed
- **THEN** exit code SHALL be 0
- **AND** all preset resources SHALL be shown in preview

#### Scenario: Init fails for nonexistent resource
- **WHEN** `germinator init --platform opencode --resources skill/nonexistent` is executed
- **THEN** exit code SHALL be > 0
- **AND** stderr SHALL indicate resource not found

#### Scenario: Init fails for nonexistent preset
- **WHEN** `germinator init --platform opencode --preset nonexistent` is executed
- **THEN** exit code SHALL be > 0
- **AND** stderr SHALL indicate preset not found

#### Scenario: Init requires platform flag
- **WHEN** `germinator init --resources skill/commit` is executed without `--platform`
- **THEN** exit code SHALL be 2
- **AND** stderr SHALL indicate platform is required

#### Scenario: Init requires resources or preset
- **WHEN** `germinator init --platform opencode` is executed without `--resources` or `--preset`
- **THEN** exit code SHALL be 2
- **AND** stderr SHALL indicate resources or preset is required

#### Scenario: Init rejects mutually exclusive flags
- **WHEN** `germinator init --platform opencode --resources skill/commit --preset git-workflow` is executed
- **THEN** exit code SHALL be 2
- **AND** stderr SHALL indicate flags are mutually exclusive

#### Scenario: Init fails for invalid platform
- **WHEN** `germinator init --platform invalid --resources skill/commit` is executed
- **THEN** exit code SHALL be 2
- **AND** stderr SHALL indicate unknown platform

#### Scenario: Init succeeds with claude-code platform
- **WHEN** `germinator init --platform claude-code --resources skill/commit --dry-run` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL show claude-code output paths

---

### Requirement: Library Command E2E Tests

The library command SHALL be tested for all expected behaviors.

#### Scenario: Library resources lists resources
- **WHEN** `germinator library resources` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL show resources grouped by type (Skills, Agents, Commands, Memory)

#### Scenario: Library presets lists presets
- **WHEN** `germinator library presets` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL show presets with descriptions and resource lists

#### Scenario: Library show displays resource details
- **WHEN** `germinator library show skill/commit` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL show resource details

#### Scenario: Library show displays preset details
- **WHEN** `germinator library show preset/git-workflow` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL show preset details with resource list

#### Scenario: Library show fails for invalid reference format
- **WHEN** `germinator library show invalidformat` is executed
- **THEN** exit code SHALL be > 0
- **AND** stderr SHALL indicate invalid reference format

#### Scenario: Library uses custom path via flag
- **WHEN** `germinator library resources --library /custom/path` is executed
- **THEN** exit code SHALL be 0 or indicate library not found
- **AND** library SHALL be loaded from specified path

#### Scenario: Library uses custom path via environment
- **GIVEN** environment variable `GERMINATOR_LIBRARY=/custom/path`
- **WHEN** `germinator library resources` is executed
- **THEN** exit code SHALL be 0 or indicate library not found
- **AND** library SHALL be loaded from environment path

---

### Requirement: Mise Tasks for E2E Testing

Mise tasks SHALL be provided for running E2E tests.

#### Scenario: test:e2e task runs E2E tests

- **WHEN** `mise run test:e2e` is executed
- **THEN** all E2E tests SHALL run with verbose output

#### Scenario: test:full task runs all tests

- **WHEN** `mise run test:full` is executed
- **THEN** unit tests SHALL run first
- **AND** E2E tests SHALL run after unit tests pass

### Requirement: Coverage-instrumented E2E binary

The E2E binary SHALL be built with `-cover` instrumentation (Go 1.20+) so coverage data can be extracted from subprocess runs.

#### Scenario: E2E binary is cover-instrumented

- **WHEN** `gexec.Build` constructs the E2E binary
- **THEN** the build command SHALL include `-cover`
- **AND** the produced binary SHALL write coverage profiles when `GOCOVERDIR` is set
- **AND** tests may aggregate per-package coverage from the subprocess runs

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


# E2E Testing Specification

## Purpose

End-to-end testing infrastructure using Ginkgo v2, Gomega, and gexec to validate the germinator CLI commands through actual binary execution.

## Requirements

### Requirement: E2E Test Suite Setup

The E2E test suite SHALL be configured with Ginkgo v2, Gomega, and gexec using the `//go:build e2e` build tag.

#### Scenario: Suite initializes successfully
- **WHEN** the E2E test suite runs
- **THEN** the germinator-e2e binary SHALL be built to `bin/germinator-e2e`
- **AND** the binary SHALL be available for all test cases
- **AND** the binary SHALL be cleaned up after all tests complete

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

### Requirement: Mise Tasks for E2E Testing

Mise tasks SHALL be provided for running E2E tests.

#### Scenario: test:e2e task runs E2E tests
- **WHEN** `mise run test:e2e` is executed
- **THEN** all E2E tests SHALL run with verbose output

#### Scenario: test:full task runs all tests
- **WHEN** `mise run test:full` is executed
- **THEN** unit tests SHALL run first
- **AND** E2E tests SHALL run after unit tests pass

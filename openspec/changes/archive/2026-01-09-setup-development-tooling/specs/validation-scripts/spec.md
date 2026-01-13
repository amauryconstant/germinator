# Validation Scripts Specification

## ADDED Requirements

### Requirement: mise Validation Task

The project SHALL provide a mise task that runs comprehensive quality checks on codebase.

#### Scenario: mise validation task exists
**Given** development tooling is set up
**When** a developer checks for mise tasks
**Then** mise.toml SHALL exist with [tasks.validate]
**And** task SHALL be discoverable via `mise run --help`

#### Scenario: Validation task runs all checks
**Given** mise.toml exists with [tasks.validate]
**When** a developer runs `mise run validate`
**Then** it SHALL run `go build ./...`
**And** it SHALL run `go mod tidy`
**And** it SHALL run `go vet ./...`
**And** it SHALL run `golangci-lint run`

#### Scenario: Validation task reports failures
**Given** mise validate task is running
**When** any check fails
**Then** task SHALL exit with non-zero status
**And** it SHALL report which check failed
**And** it SHALL provide actionable error information

---

### Requirement: mise Smoke Test Task

The project SHALL provide a mise task for quick validation of basic project health.

#### Scenario: mise smoke-test task exists
**Given** development tooling is set up
**When** a developer checks for mise tasks
**Then** mise.toml SHALL exist with [tasks.smoke-test]
**And** task SHALL be discoverable via `mise run --help`

#### Scenario: Smoke test runs basic checks
**Given** mise.toml exists with [tasks.smoke-test]
**When** a developer runs `mise run smoke-test`
**Then** it SHALL verify the project builds
**And** it SHALL complete quickly

#### Scenario: Smoke test uses incremental builds
**Given** mise.toml exists with [tasks.smoke-test]
**When** a developer inspects task configuration
**Then** sources field SHALL be defined (e.g., "cmd/**/*.go")
**And** outputs field SHALL be defined (e.g., "germinator")
**And** task SHALL skip unchanged files on subsequent runs

---

### Requirement: File-Based Task Scripts

The project SHALL provide file-based task scripts in `.mise/tasks/` directory.

#### Scenario: .mise/tasks/ directory exists
**Given** development tooling is set up
**When** a developer inspects .mise directory
**Then** .mise/tasks/ directory SHALL exist

#### Scenario: File-based scripts are executable
**Given** .mise/tasks/ directory exists
**When** a developer inspects task scripts
**Then** .mise/tasks/validate.sh SHALL exist and be executable
**And** .mise/tasks/smoke-test.sh SHALL exist and be executable
**And** each script SHALL have shebang `#!/usr/bin/env bash`

#### Scenario: File-based scripts execute correctly
**Given** file-based scripts exist
**When** a developer executes a script directly
**Then** script SHALL run correctly with same behavior as mise task

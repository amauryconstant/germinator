# cli-framework Specification (Delta)

## ADDED Requirements

### Requirement: Enhanced Version Display

The version command SHALL display version, commit SHA, and build date for better debugging.

#### Scenario: Version command shows full info
**Given** germinator is built with version information
**When** a user runs `germinator version`
**Then** it SHALL display format: `germinator {version} ({commit}) {date}`
**And** version SHALL be the semantic version (e.g., v0.3.0)
**And** commit SHALL be 7-character commit SHA (e.g., abc1234)
**And** date SHALL be YYYY-MM-DD format (e.g., 2025-01-13)

#### Scenario: Version with tag
**Given** germinator is built from a Git tag (e.g., v0.3.0)
**When** version command runs
**Then** it SHALL display: `germinator v0.3.0 (abc1234) 2025-01-13`
**And** version SHALL match Git tag
**And** commit SHALL be tag's commit SHA
**And** date SHALL be commit date

#### Scenario: Version without tag
**Given** germinator is built from non-tagged commit
**When** version command runs
**Then** it SHALL display: `germinator v0.3.0-1-gabc1234 (abc1234) 2025-01-13`
**And** version SHALL include git describe output
**And** commit SHALL be current HEAD SHA
**And** date SHALL be current date

---

### Requirement: Version Package Variables

The version package SHALL use variables instead of constants for build-time injection.

#### Scenario: Version is variable
**Given** `internal/version/version.go` is inspected
**When** version variable is declared
**Then** it SHALL use `var` instead of `const`
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "0.2.0"

#### Scenario: Commit is variable
**Given** `internal/version/version.go` is inspected
**When** commit variable is declared
**Then** it SHALL use `var` for commit SHA
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "dev"

#### Scenario: Date is variable
**Given** `internal/version/version.go` is inspected
**When** date variable is declared
**Then** it SHALL use `var` for build date
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "unknown"

#### Scenario: Variables exported
**Given** version package is inspected
**When** exports are checked
**Then** `Version` variable SHALL be exported
**And** `Commit` variable SHALL be exported
**And** `Date` variable SHALL be exported

---

## MODIFIED Requirements

### Requirement: Version Command

The version command SHALL display version information for debugging and release tracking.

#### Scenario: Version command works
**Given** germinator is installed
**When** a user runs `germinator version`
**Then** it SHALL execute successfully
**And** it SHALL display version in format: `germinator {version} ({commit}) {date}`
**And** it SHALL exit with code 0

#### Scenario: Version help is available
**Given** a user runs `germinator version --help`
**Then** it SHALL display command help
**And** it SHALL show description: "Show version of germinator"

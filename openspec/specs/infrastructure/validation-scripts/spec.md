# validation-scripts Specification

## Purpose
Define validation scripts for checking git state and GoReleaser configuration.

## Requirements
### Requirement: Release Validation Script

The project SHALL provide a mise task that validates release prerequisites.

#### Scenario: Release validate task exists
**Given** development tooling is set up
**When** a developer checks for mise tasks
**Then** .mise/config.toml SHALL have [tasks."release:validate"]
**And** task SHALL be discoverable via `mise run --help`

#### Scenario: Release validation script exists
**Given** tool management is set up
**When** a developer inspects `.mise/tasks/` directory
**Then** `.mise/tasks/release/validate.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Release validation checks git state
**Given** .mise/tasks/release/validate.sh is inspected
**When** script executes
**Then** it SHALL check git working directory is clean
**And** it SHALL verify current branch is main
**And** it SHALL validate GoReleaser configuration if goreleaser is installed
**And** it SHALL report all issues found with clear error messages

#### Scenario: Release validation script is executable
**Given** .mise/tasks/release/validate.sh exists
**When** file permissions are inspected
**Then** file SHALL be executable
**And** script SHALL have shebang `#!/usr/bin/env bash`

#### Scenario: Release validation task reports failures
**Given** .mise/run release:validate is running
**When** any check fails
**Then** task SHALL exit with non-zero status
**And** it SHALL report which check failed
**And** it SHALL provide actionable error information

---

### Requirement: Tool Management Scripts

The project SHALL provide scripts for checking and updating tool versions.

#### Scenario: Tool check script exists
**Given** tool management is set up
**When** a developer inspects `.mise/tasks/tools/` directory
**Then** `.mise/tasks/tools/check.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Tool update script exists
**Given** tool management is set up
**When** a developer inspects `.mise/tasks/tools/` directory
**Then** `.mise/tasks/tools/update.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Release tag script exists
**Given** tool management is set up
**When** a developer inspects `.mise/tasks/release/` directory
**Then** `.mise/tasks/release/tag.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: File-based scripts are executable
**Given** .mise/tasks/ directory exists
**When** a developer inspects task scripts
**Then** all .sh scripts SHALL be executable
**And** each script SHALL have shebang `#!/usr/bin/env bash`

#### Scenario: File-based scripts execute correctly
**Given** file-based scripts exist
**When** a developer executes a script directly
**Then** script SHALL run correctly with same behavior as mise task

---

### Requirement: Comprehensive Validation via Check Task

The project SHALL provide a mise task that runs all validation checks (lint, format, test, build).

#### Scenario: Check task exists
**Given** .mise/config.toml exists with tasks defined
**When** [tasks] section is inspected
**Then** [tasks.check] SHALL exist
**And** task SHALL be discoverable via `mise run --help`
**And** task SHALL have description "Run all validation checks"

#### Scenario: Check task runs all validations
**Given** .mise/config.toml exists with [tasks.check]
**When** a developer runs `mise run check`
**Then** it SHALL run lint task
**And** it SHALL run format task
**And** it SHALL run test task
**And** it SHALL run build task
**And** tasks SHALL run in dependency order

#### Scenario: Check task reports failures
**Given** mise check task is running
**When** any dependent task fails
**Then** task SHALL exit with non-zero status
**And** it SHALL report which task failed
**And** it SHALL provide actionable error information

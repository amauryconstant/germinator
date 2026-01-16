# mise-task-runner Specification

## Purpose
TBD - created by archiving change setup-development-tooling. Update Purpose after archive.
## Requirements
### Requirement: mise Configuration File

The project SHALL have a mise.toml configuration file for task definitions and tool installation.

#### Scenario: mise.toml exists
**Given** development tooling is set up
**When** a developer checks for mise configuration
**Then** mise.toml SHALL exist in project root
**And** it SHALL be valid TOML syntax

#### Scenario: Tools section exists
**Given** mise.toml exists
**When** the configuration is inspected
**Then** [tools] section SHALL exist
**And** golangci-lint SHALL be configured
**And** version SHALL be specified (e.g., "latest")

#### Scenario: Tasks section exists
**Given** mise.toml exists
**When** the configuration is inspected
**Then** [tasks] section SHALL exist
**And** at least 2 tasks SHALL be defined (validate, smoke-test)

---

### Requirement: Tool Auto-Installation

The project SHALL leverage mise's automatic tool installation for golangci-lint.

#### Scenario: golangci-lint installs automatically
**Given** mise.toml exists with golangci-lint configured
**When** a developer runs `mise use golangci-lint@latest`
**Then** mise SHALL download and install golangci-lint
**And** tool SHALL be available for use

#### Scenario: Tool is discoverable
**Given** mise is installed
**When** a developer runs `mise list`
**Then** installed tools SHALL be listed
**And** golangci-lint SHALL appear in list

---

### Requirement: Task Discovery

The project SHALL provide task discovery through mise help system.

#### Scenario: Tasks are discoverable
**Given** mise.toml exists with tasks defined
**When** a developer runs `mise run --help`
**Then** all defined tasks SHALL be listed
**And** each task SHALL show its description
**And** tasks SHALL be in alphabetical order

#### Scenario: Task usage is documented
**Given** task list is displayed
**When** a developer inspects a task
**Then** task name SHALL be shown
**And** task description SHALL be shown

---

### Requirement: Parallel Task Execution

The project SHALL leverage mise's parallel task execution capabilities.

#### Scenario: Tasks run in parallel
**Given** multiple tasks are defined without dependencies
**When** a developer runs tasks in parallel (e.g., `mise run task1 task2`)
**Then** tasks SHALL execute concurrently
**And** execution time SHALL be reduced

---

### Requirement: Incremental Builds

The project SHALL leverage mise's incremental build capabilities for performance.

#### Scenario: Task has sources defined
**Given** mise.toml exists with smoke-test task
**When** a developer inspects task configuration
**Then** sources field SHALL be defined
**And** pattern SHALL match input files (e.g., "cmd/**/*.go")

#### Scenario: Task has outputs defined
**Given** mise.toml exists with smoke-test task
**When** a developer inspects task configuration
**Then** outputs field SHALL be defined
**And** output path SHALL be specified (e.g., "germinator")

#### Scenario: Task skips unchanged files
**Given** task has sources and outputs defined
**When** task is run multiple times
**Then** task SHALL skip re-execution if sources are unchanged
**And** outputs remain valid

### Requirement: GoReleaser Tool Management

The project SHALL manage GoReleaser via mise for release automation.

#### Scenario: GoReleaser in tools section
**Given** `.mise/config.toml` exists
**When** [tools] section is inspected
**Then** goreleaser SHALL be configured
**And** version SHALL be specified (e.g., "2.4.0")

#### Scenario: GoReleaser installs automatically
**Given** `.mise/config.toml` has goreleaser configured
**When** a developer runs `mise install --yes`
**Then** mise SHALL download and install GoReleaser
**And** tool SHALL be available as `goreleaser`

#### Scenario: GoReleaser is discoverable
**Given** mise is installed
**When** a developer runs `mise list`
**Then** goreleaser SHALL appear in installed tools list

---

### Requirement: Tool Version Check Script

The project SHALL provide script to check for available tool updates.

#### Scenario: Check script exists
**Given** tool management is set up
**When** a developer inspects `.mise/tasks/` directory
**Then** `.mise/tasks/tools/check.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Check script queries GitHub API
**Given** a developer runs `mise run tools:check`
**When** script executes
**Then** it SHALL query GitHub API for golangci-lint latest release
**And** it SHALL query GitHub API for GoReleaser latest release
**And** it SHALL compare with current versions in `.mise/config.toml`
**And** it SHALL display current and latest versions for each tool
**And** it SHALL indicate if updates are available

---

### Requirement: Tool Update Script

The project SHALL provide script to update tool versions in mise configuration.

#### Scenario: Update script exists
**Given** tool management is set up
**When** a developer inspects `.mise/tasks/` directory
**Then** `.mise/tasks/tools/update.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Update script fetches latest versions
**Given** a developer runs `mise run tools:update`
**When** script executes
**Then** it SHALL fetch latest golangci-lint version from GitHub API
**And** it SHALL fetch latest GoReleaser version from GitHub API
**And** it SHALL handle API failures gracefully with error message
**And** it SHALL update `.mise/config.toml` with latest versions

#### Scenario: Update script modifies config
**Given** latest versions are fetched
**When** `.mise/config.toml` is updated
**Then** golangci-lint version SHALL be updated
**And** GoReleaser version SHALL be updated
**And** file SHALL remain valid TOML syntax

#### Scenario: Update script uses Python for cross-platform compatibility
**Given** `.mise/tasks/update-tools.sh` is executed
**When** script modifies TOML files
**Then** it SHALL use Python for cross-platform TOML manipulation
**And** it SHALL work on Linux (sed/gsed differences)
**And** it SHALL work on macOS (sed/gsed differences)
**And** it SHALL handle Python installation failure gracefully

#### Scenario: Update script documents next steps
**Given** tool versions are updated
**When** script completes
**Then** it SHALL display message showing updated versions
**And** it SHALL instruct to review changes with `git diff`
**And** it SHALL instruct to install updated tools with `mise install --yes`
**And** it SHALL instruct to rebuild CI image
**And** it SHALL instruct to commit and push changes

---

### Requirement: Release Tasks

The project SHALL provide mise tasks for release-related operations.

#### Scenario: Dry-run release task exists
**Given** `.mise/config.toml` exists
**When** [tasks] section is inspected
**Then** `release:dry-run` task SHALL exist
**And** it SHALL run `goreleaser release --skip=publish --clean`
**And** it SHALL have description "Test GoReleaser without creating release"

#### Scenario: Dry-run builds artifacts locally
**Given** a developer runs `mise run release:dry-run`
**When** task executes
**Then** GoReleaser SHALL build all artifacts
**And** it SHALL skip publishing to GitLab
**And** it SHALL create artifacts in `dist/` directory
**And** it SHALL display what would be released

#### Scenario: Release validate task exists
**Given** `.mise/config.toml` exists
**When** [tasks] section is inspected
**Then** `release:validate` task SHALL exist
**And** it SHALL check git state and branch
**And** it SHALL have description "Validate release prerequisites before tagging"

#### Scenario: Release validate task checks prerequisites
**Given** a developer runs `mise run release:validate`
**When** task executes
**Then** task SHALL check git working directory is clean
**And** task SHALL verify current branch is main
**And** task SHALL validate GoReleaser configuration if installed
**And** task SHALL report all issues found

#### Scenario: Release tag task exists
**Given** `.mise/config.toml` exists
**When** [tasks] section is inspected
**Then** `release:tag` task SHALL exist
**And** it SHALL accept patch|minor|major argument
**And** it SHALL create and push git tag
**And** it SHALL enforce vX.Y.Z format

#### Scenario: Release tag task creates version tag
**Given** a developer runs `mise run release:tag patch`
**When** latest tag is v0.3.20
**Then** task SHALL create tag v0.3.21
**And** task SHALL push tag to origin
**And** task SHALL validate format before creating

---

### Requirement: Tool Version Pinning

The project SHALL pin golangci-lint and GoReleaser to specific versions for reproducible builds.

#### Scenario: Pinned versions in config
**Given** `.mise/config.toml` exists
**When** [tools] section is inspected
**Then** golangci-lint SHALL be pinned to specific version (e.g., "1.60.1")
**And** GoReleaser SHALL be pinned to specific version (e.g., "2.4.0")
**And** versions SHALL not use "latest" for production

#### Scenario: Reproducible builds
**Given** pinned versions are set
**When** build runs multiple times
**Then** same tool versions SHALL be used
**And** build results SHALL be consistent


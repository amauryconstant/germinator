# project-layout Specification

## Purpose
Establish standard Go directory structure following conventions, with internal packages for private code.

## Requirements
### Requirement: Standard Go Directory Structure

The project SHALL provide a standard Go directory layout following [Standard Go Project Layout](https://github.com/golang-standards/project-layout) conventions.

#### Scenario: Developer navigates project structure
**Given** the project has been initialized
**When** a developer runs `tree -L 2` or `ls -R`
**Then** the following directories SHALL exist:
- cmd/
- internal/
- internal/core/
- internal/services/
- internal/models/
- internal/version/
- config/
- config/templates/
- config/templates/claude-code/
- test/
- test/fixtures/

#### Scenario: Go packages compile successfully
**Given** the project structure has been created
**When** a developer runs `go build ./...`
**Then** the build SHALL succeed without errors
**And** all packages SHALL compile

---

### Requirement: Go Module Initialization

The project SHALL be initialized as a Go module with a valid module path.

#### Scenario: Go module file exists
**Given** the project has been initialized
**When** a developer checks for go.mod file
**Then** go.mod SHALL exist in the project root
**And** it SHALL contain a valid `module` declaration

#### Scenario: Go module has correct version
**Given** go.mod exists
**When** a developer reads the go.mod file
**Then** it SHALL specify a Go version
**And** it SHALL use a valid module path format (e.g., gitlab.com/username/germinator)

#### Scenario: Dependencies are managed
**Given** the Go module is initialized
**When** dependencies are added with `go get`
**Then** go.sum SHALL be updated automatically
**And** `go mod tidy` SHALL resolve all dependencies

---

### Requirement: Package Documentation

Each package SHALL have minimal documentation explaining its purpose.

#### Scenario: Package doc.go files exist
**Given** packages have been created
**When** a developer inspects package directories
**Then** internal/core/doc.go SHALL exist
**And** internal/services/doc.go SHALL exist
**And** internal/models/doc.go SHALL exist
**And** each doc.go SHALL describe the package's purpose

---

### Requirement: Configuration Structure

The project SHALL have a configuration structure for templates.

#### Scenario: Configuration directories exist
**Given** the project has been initialized
**When** a developer inspects config/ directory
**Then** config/templates/ SHALL exist
**And** config/templates/claude-code/ SHALL exist

---

### Requirement: Test Structure

The project SHALL have a test structure for fixtures.

#### Scenario: Test directories exist
**Given** the project has been initialized
**When** a developer inspects test/ directory
**Then** test/fixtures/ SHALL exist

---

### Requirement: Utility Scripts Location

The project SHALL have utility scripts in .mise/tasks/ directory following mise conventions.

#### Scenario: Scripts directory exists
**Given** the project has been initialized
**When** a developer checks for scripts directory
**Then** .mise/tasks/ SHALL exist
**And** .mise/tasks/release/ SHALL exist for release scripts
**And** .mise/tasks/tools/ SHALL exist for tool management scripts

---

### Requirement: Project Documentation

The project SHALL have documentation explaining its structure and how to build it.

#### Scenario: Root README exists
**Given** the project has been initialized
**When** a developer reads README.md
**Then** it SHALL describe the project's purpose
**And** it SHALL explain the directory structure
**And** it SHALL provide build instructions

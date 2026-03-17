# comprehensive-linting Specification

## Purpose

Configure golangci-lint with 25+ linters including depguard for domain purity enforcement, providing comprehensive code quality coverage.

## ADDED Requirements

### Requirement: Linter Count and Categories

The project SHALL use a comprehensive linter configuration with 25+ linters organized by category.

#### Scenario: Essential linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following essential linters SHALL be enabled:
  - `errcheck`
  - `govet`
  - `ineffassign`
  - `staticcheck`
  - `unused`
  - `typecheck`

#### Scenario: Code quality linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following code quality linters SHALL be enabled:
  - `gocyclo` (min-complexity: 25)
  - `gocognit` (min-complexity: 30)
  - `funlen` (lines: 150, statements: 100)

#### Scenario: Style linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following style linters SHALL be enabled:
  - `misspell`
  - `whitespace`
  - `revive`

#### Scenario: Error handling linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following error handling linters SHALL be enabled:
  - `errorlint`
  - `wrapcheck`
  - `errname`

#### Scenario: Performance linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following performance linters SHALL be enabled:
  - `prealloc`
  - `perfsprint`

#### Scenario: Security linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** `gosec` SHALL be enabled with appropriate exclusions

#### Scenario: Test linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following test linters SHALL be enabled:
  - `testifylint`
  - `tparallel`
  - `thelper`

#### Scenario: Best practices linters enabled

- **WHEN** `.golangci.yml` is inspected
- **THEN** the following best practices linters SHALL be enabled:
  - `nakedret`
  - `unconvert`
  - `unparam`
  - `wastedassign`

---

### Requirement: Domain Purity Enforcement

The linter configuration SHALL enforce that the `internal/domain/` package has no external dependencies.

#### Scenario: Depguard rule for domain package

- **WHEN** `.golangci.yml` is inspected
- **THEN** a `depguard` rule named `domain` SHALL exist
- **AND** it SHALL apply to files matching `internal/domain/**`

#### Scenario: Domain allows only stdlib and internal/domain

- **GIVEN** a Go file in `internal/domain/`
- **WHEN** the file imports an external package (e.g., `github.com/...`)
- **THEN** `golangci-lint run` SHALL fail with a depguard error

#### Scenario: Domain allows standard library

- **GIVEN** a Go file in `internal/domain/`
- **WHEN** the file imports only standard library packages
- **THEN** `golangci-lint run` SHALL pass the depguard check

#### Scenario: Domain allows internal/domain imports

- **GIVEN** a Go file in `internal/domain/`
- **WHEN** the file imports another file from `internal/domain/`
- **THEN** `golangci-lint run` SHALL pass the depguard check

#### Scenario: Domain does NOT allow application layer imports

- **GIVEN** a Go file in `internal/domain/`
- **WHEN** the file imports from `internal/application`
- **THEN** `golangci-lint run` SHALL fail with a depguard error
- **AND** the error SHALL indicate domain layer cannot depend on application layer

---

### Requirement: Linter Thresholds

The linter configuration SHALL set appropriate complexity thresholds.

#### Scenario: Cyclomatic complexity threshold

- **WHEN** a function has cyclomatic complexity greater than 25
- **THEN** `gocyclo` SHALL report an error

#### Scenario: Cognitive complexity threshold

- **WHEN** a function has cognitive complexity greater than 30
- **THEN** `gocognit` SHALL report an error

#### Scenario: Function length threshold

- **WHEN** a function exceeds 150 lines
- **THEN** `funlen` SHALL report an error

#### Scenario: Function statements threshold

- **WHEN** a function exceeds 100 statements
- **THEN** `funlen` SHALL report an error

---

### Requirement: Test File Exclusions

The linter configuration SHALL exclude certain linters for test files.

#### Scenario: Test files excluded from complexity linters

- **GIVEN** a file matching `*_test.go`
- **WHEN** `golangci-lint run` is executed
- **THEN** the following linters SHALL NOT apply:
  - `funlen`
  - `gocyclo`
  - `gocognit`

#### Scenario: Test files excluded from security linter

- **GIVEN** a file matching `*_test.go`
- **WHEN** `golangci-lint run` is executed
- **THEN** `gosec` SHALL NOT apply

#### Scenario: Test files excluded from wrapcheck

- **GIVEN** a file matching `*_test.go`
- **WHEN** `golangci-lint run` is executed
- **THEN** `wrapcheck` SHALL NOT apply

#### Scenario: E2E test files have broad exclusions

- **GIVEN** a file matching `test/e2e/**/*.go`
- **WHEN** `golangci-lint run` is executed
- **THEN** the following linters SHALL NOT apply:
  - `funlen`
  - `gocyclo`
  - `gocognit`
  - `gosec`
  - `unparam`
  - `wastedassign`
  - `wrapcheck`
  - `errcheck`
  - `prealloc`

---

### Requirement: GoSec Exclusions

The GoSec linter SHALL exclude specific rules that generate false positives for CLI tools.

#### Scenario: GoSec excludes G104

- **WHEN** `.golangci.yml` is inspected
- **THEN** GoSec SHALL exclude G104 (unchecked errors in defer/test contexts)

#### Scenario: GoSec excludes file permission rules

- **WHEN** `.golangci.yml` is inspected
- **THEN** GoSec SHALL exclude G301, G304, G306, G307 (file/directory permissions)

#### Scenario: GoSec excludes subprocess rule

- **WHEN** `.golangci.yml` is inspected
- **THEN** GoSec SHALL exclude G204 (subprocess input)

---

### Requirement: Wrapcheck Configuration

The wrapcheck linter SHALL be configured to ignore standard error patterns.

#### Scenario: Wrapcheck ignores standard error constructors

- **WHEN** `.golangci.yml` is inspected
- **THEN** wrapcheck SHALL ignore signatures matching:
  - `.Error\(`
  - `.Errorf\(`
  - `errors\.New\(`
  - `errors\.As\(`
  - `errors\.Is\(`
  - `fmt\.Errorf\(`

#### Scenario: Wrapcheck ignores domain package

- **WHEN** `.golangci.yml` is inspected
- **THEN** wrapcheck SHALL ignore packages matching `internal/domain`

---

### Requirement: Linting Task

The mise configuration SHALL provide lint tasks.

#### Scenario: lint task runs golangci-lint

- **WHEN** `mise run lint` is executed
- **THEN** `golangci-lint run` SHALL be executed
- **AND** all enabled linters SHALL check the codebase

#### Scenario: lint:fix task auto-fixes issues

- **WHEN** `mise run lint:fix` is executed
- **THEN** `golangci-lint run --fix` SHALL be executed
- **AND** auto-fixable issues SHALL be resolved

#### Scenario: check task includes lint

- **WHEN** `mise run check` is executed
- **THEN** linting SHALL be included in the validation pipeline

---

### Requirement: Clean Linting

All code SHALL pass the comprehensive linter configuration without errors.

#### Scenario: No linting errors in production code

- **GIVEN** all non-test Go files in the project
- **WHEN** `golangci-lint run` is executed
- **THEN** no errors SHALL be reported

#### Scenario: No linting errors in test code

- **GIVEN** all test Go files in the project
- **WHEN** `golangci-lint run` is executed
- **THEN** no errors SHALL be reported (after exclusions applied)

# domain-structure Specification

## Purpose

Consolidate domain types (models, errors, validation, requests, results) into a single `internal/domain/` package with no external dependencies, establishing the foundation for DDD-light architecture.

## Requirements

### Requirement: Domain Package Structure

The system SHALL provide a unified `internal/domain/` package containing all domain types.

#### Scenario: Domain package exists

- **WHEN** the project structure is inspected
- **THEN** an `internal/domain/` directory SHALL exist
- **AND** it SHALL contain all domain types with no external dependencies

#### Scenario: Domain package files organized by type

- **WHEN** the domain package is inspected
- **THEN** it SHALL contain `agent.go`, `command.go`, `skill.go`, `memory.go`, `platform.go`
- **AND** it SHALL contain `errors.go`, `validation.go`, `result.go`, `results.go`
- **AND** it SHALL contain `opencode/` subdirectory with OpenCode-specific validators
- **AND** it SHALL contain `doc.go` with package documentation

---

### Requirement: Domain Types Migrated

All domain types from `internal/models/canonical/` SHALL be migrated to `internal/domain/`.

#### Scenario: Agent types in domain

- **WHEN** domain package is loaded
- **THEN** `Agent` struct SHALL be defined in `agent.go`
- **AND** `AgentMemory` struct SHALL be defined in `memory.go`

#### Scenario: Command types in domain

- **WHEN** domain package is loaded
- **THEN** `Command` struct SHALL be defined in `command.go`

#### Scenario: Skill types in domain

- **WHEN** domain package is loaded
- **THEN** `Skill` struct SHALL be defined in `skill.go`

#### Scenario: Platform types in domain

- **WHEN** domain package is loaded
- **THEN** `Platform` enum SHALL be defined in `platform.go`

---

### Requirement: Domain Errors Consolidated

All error types from `internal/errors/` SHALL be consolidated into `internal/domain/errors.go`.

#### Scenario: Error types in domain

- **WHEN** domain package is loaded
- **THEN** all error types from `internal/errors/` SHALL be available
- **AND** the `internal/errors/` package SHALL be removed

---

### Requirement: Domain Validation Consolidated

All validation types from `internal/validation/` SHALL be consolidated into `internal/domain/`.

#### Scenario: Result type in domain

- **WHEN** domain package is loaded
- **THEN** `Result[T]` type SHALL be available from `result.go`

#### Scenario: Validation types in domain

- **WHEN** domain package is loaded
- **THEN** all validators SHALL be available from `validation.go`
- **AND** the `internal/validation/` package SHALL be removed

---

### Requirement: Result Types in Domain

Result types SHALL be moved from `internal/application/` to `internal/domain/`.

#### Scenario: Result types in domain

- **WHEN** domain package is loaded
- **THEN** `TransformResult`, `ValidateResult`, `CanonicalizeResult`, `InitializeResult` SHALL be available
- **AND** they SHALL be imported from `internal/domain`

---

### Requirement: Request Types Stay in Application

Request types SHALL remain in `internal/application/` because `InitializeRequest` has external dependencies.

#### Scenario: Request types remain in application layer

- **GIVEN** `InitializeRequest` depends on `*library.Library` (infrastructure concern)
- **WHEN** domain purity is enforced
- **THEN** request types SHALL remain in `internal/application/`
- **AND** they SHALL NOT be moved to `internal/domain/`

#### Scenario: Pure request DTOs may move

- **GIVEN** `TransformRequest`, `ValidateRequest`, `CanonicalizeRequest` have no external dependencies
- **WHEN** domain migration is complete
- **THEN** these types MAY optionally move to `internal/domain/` in a future change
- **BUT** `InitializeRequest` SHALL stay in application layer

---

### Requirement: Domain Purity Enforcement

The `internal/domain/` package SHALL have no external dependencies beyond Go standard library.

#### Scenario: depguard enforces domain purity

- **WHEN** `golangci-lint run` is executed
- **THEN** depguard SHALL reject any import in `internal/domain/**` that is not:
  - Go standard library (`$gostd`)
  - `internal/domain` itself

#### Scenario: Domain package imports only stdlib

- **WHEN** any file in `internal/domain/` is inspected
- **THEN** all imports SHALL be from Go standard library only
- **AND** no third-party imports SHALL be present

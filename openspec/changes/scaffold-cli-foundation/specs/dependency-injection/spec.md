# dependency-injection Specification (delta)

## MODIFIED Requirements

### Requirement: ServiceContainer removed

The `ServiceContainer` type and its `NewServiceContainer()` constructor SHALL be **removed**. The eager service wiring pattern SHALL be replaced conceptually by `cmdutil.Factory` (introduced in the `cli/cli-factory` capability; see `openspec/changes/scaffold-cli-foundation/specs/cli-factory/spec.md`).

#### Scenario: ServiceContainer type removed

- **WHEN** the codebase is inspected
- **THEN** the `ServiceContainer` type SHALL NOT exist in any package
- **AND** `NewServiceContainer()` SHALL NOT exist
- **AND** `cmd/container.go` SHALL be deleted (deletion happens in change-2)

> **Status (slice 1 / foundation):** the `cmdutil.Factory` is introduced with tests; the `ServiceContainer` type still exists in `cmd/container.go` and is still wired in `main.go`. Deletion of `ServiceContainer` and `container.go` happens in change-2 (`wire-factory-and-pilots`); full removal of `internal/application/` happens in change-7 (`migrate-library-rest`).

### Requirement: Dependency injection via Factory

The system SHALL provide dependency injection exclusively through the `cmdutil.Factory` struct (see `cli/cli-factory` capability). No other DI mechanism SHALL be used.

#### Scenario: Single DI mechanism

- **WHEN** the codebase is inspected
- **THEN** there SHALL be exactly one DI mechanism: the `cmdutil.Factory` struct
- **AND** no DI container (e.g. `samber/do`) SHALL be added
- **AND** no service locator pattern SHALL be used

> **Status (slice 1 / foundation):** the Factory exists with tests; no command code uses it yet. Adoption happens command-by-command in changes 2-9.

# dependency-injection Specification (delta)

## MODIFIED Requirements

### Requirement: ServiceContainer replaced by Factory

The `cmd.ServiceContainer` struct and `cmd.NewServiceContainer()` constructor SHALL NOT be used for new code; commands SHALL obtain dependencies through the `cmdutil.Factory` introduced in `cli-factory`. The `ServiceContainer` type and `cmd/container.go` SHALL be deleted in change-7 (see `## REMOVED Requirements` below).

#### Scenario: No command imports ServiceContainer

- **WHEN** a command file under `cmd/` is inspected
- **THEN** it SHALL NOT import the `ServiceContainer` type or call `NewServiceContainer`
- **AND** it SHALL obtain its dependencies from a `*cmdutil.Factory`

> **Status (slice 1 / foundation):** `cmdutil.Factory` is introduced with full table-driven tests; the legacy `ServiceContainer` and `cmd/container.go` still exist and are removed in changes 2 and 7 as noted in the REMOVED section.

### Requirement: Dependency injection via Factory

The system SHALL provide dependency injection exclusively through the `cmdutil.Factory` struct (see `cli-factory` capability). No other DI mechanism SHALL be used.

#### Scenario: Single DI mechanism

- **WHEN** the codebase is inspected
- **THEN** there SHALL be exactly one DI mechanism: the `cmdutil.Factory` struct
- **AND** no DI container (e.g. `samber/do`) SHALL be added
- **AND** no service locator pattern SHALL be used

> **Status (slice 1 / foundation):** the Factory exists with full table-driven tests; no command code uses it yet. Adoption happens command-by-command in changes 2-9.

## REMOVED Requirements

### Requirement: ServiceContainer constructor and type

**Reason**: The eager `ServiceContainer` pattern is incompatible with the lazy `Factory` model; `ServiceContainer` instantiates all four services (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`) at `main.go` startup regardless of whether the invoked command needs them.

**Migration**: Use `*cmdutil.Factory` instead. The `Factory`'s lazy `func() (T, error)` fields populate the same services on first use. See `specs/cli-factory/spec.md` for the Factory contract.

#### Scenario: ServiceContainer type removed

- **WHEN** the codebase is inspected after change-7
- **THEN** the `ServiceContainer` type SHALL NOT exist in any package
- **AND** `NewServiceContainer()` SHALL NOT exist
- **AND** `cmd/container.go` SHALL be deleted

> **Status (slice 1 / foundation):** `ServiceContainer` and `cmd/container.go` still exist; full removal of `internal/application/` happens in change-7. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

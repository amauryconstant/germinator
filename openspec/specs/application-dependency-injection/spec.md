# Dependency Injection

## Purpose

This capability defines the dependency injection pattern for wiring services and commands in the CLI application. It establishes a clean architecture where commands receive their dependencies through a `cmdutil.Factory` (see `cli-factory`), enabling testability, lazy initialization, and maintainability.

## Requirements

### Requirement: ServiceContainer replaced by Factory

The `cmd.ServiceContainer` struct and `cmd.NewServiceContainer()` constructor SHALL NOT be used for new code; commands SHALL obtain dependencies through the `cmdutil.Factory` introduced in `cli-factory`. The `ServiceContainer` type and `cmd/container.go` SHALL be removed (see `cli-factory` for the replacement contract).

#### Scenario: No command imports ServiceContainer

- **WHEN** a command file under `cmd/` is inspected
- **THEN** it SHALL NOT import the `ServiceContainer` type or call `NewServiceContainer`
- **AND** it SHALL obtain its dependencies from a `*cmdutil.Factory`

### Requirement: Dependency injection via Factory

The system SHALL provide dependency injection exclusively through the `cmdutil.Factory` struct (see `cli-factory`). No other DI mechanism SHALL be used.

#### Scenario: Single DI mechanism

- **WHEN** the codebase is inspected
- **THEN** there SHALL be exactly one DI mechanism: the `cmdutil.Factory` struct
- **AND** no DI container (e.g. `samber/do`) SHALL be added
- **AND** no service locator pattern SHALL be used

### Requirement: No global command variables

Commands SHALL NOT use global variables for command definitions or flags.

#### Scenario: Flags bound to local variables

- **WHEN** a command constructor defines flags
- **THEN** flag values are bound to local variables within the constructor closure
- **AND** no package-level variables store flag state

#### Scenario: No init() functions for command registration

- **WHEN** the cmd package is imported
- **THEN** no commands are automatically registered via init()
- **AND** commands are only registered through NewRootCommand

## Fulfilled

**Change:** `migrate-library-rest` (slice 7 of 9)
**Date:** 2026-07-01

> The legacy `ServiceContainer` was fully removed in this change. The replacement `cmdutil.Factory` is the live contract.

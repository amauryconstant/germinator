# Dependency Injection

## Purpose

This capability defines the dependency injection pattern for wiring services and commands in the CLI application. It establishes a clean architecture where commands receive their dependencies through constructors rather than global variables, enabling testability and maintainability.

## Requirements

### Requirement: ServiceContainer holds service instances

The system SHALL provide a `ServiceContainer` struct that groups service instances for command use.

#### Scenario: ServiceContainer is embeddable in CommandConfig
- **WHEN** CommandConfig is created
- **THEN** ServiceContainer is available as a field for commands to access services

#### Scenario: ServiceContainer is initially sparse
- **WHEN** ServiceContainer is created
- **THEN** it MAY have zero or more service fields
- **AND** the pattern supports future service additions without breaking changes

### Requirement: Commands receive configuration via constructor

Each command SHALL have a constructor function `NewXCommand(config *CommandConfig)` that returns a configured `*cobra.Command`.

#### Scenario: Validate command constructor
- **WHEN** `NewValidateCommand(config)` is called
- **THEN** it returns a `*cobra.Command` configured for validation
- **AND** the command uses services from `config.Services`

#### Scenario: Adapt command constructor
- **WHEN** `NewAdaptCommand(config)` is called
- **THEN** it returns a `*cobra.Command` configured for transformation
- **AND** the command uses services from `config.Services`

#### Scenario: Canonicalize command constructor
- **WHEN** `NewCanonicalizeCommand(config)` is called
- **THEN** it returns a `*cobra.Command` configured for canonicalization
- **AND** the command uses services from `config.Services`

#### Scenario: Version command constructor
- **WHEN** `NewVersionCommand(config)` is called
- **THEN** it returns a `*cobra.Command` configured for version display

### Requirement: Root command aggregates subcommands

The system SHALL provide `NewRootCommand(config *CommandConfig)` that creates the root command with all subcommands attached.

#### Scenario: Root command includes all subcommands
- **WHEN** `NewRootCommand(config)` is called
- **THEN** the returned command has validate, adapt, canonicalize, and version subcommands

#### Scenario: Configuration flows to all commands
- **WHEN** `NewRootCommand(config)` creates subcommands
- **THEN** each subcommand receives the same `config` instance

### Requirement: main.go is composition root

The `main.go` file SHALL be the composition root where all services are wired and the command tree is constructed.

#### Scenario: ServiceContainer creation in main
- **WHEN** main() executes
- **THEN** a ServiceContainer is instantiated
- **AND** the container is embedded in CommandConfig

#### Scenario: Command tree construction in main
- **WHEN** main() executes  
- **THEN** `NewRootCommand(config)` is called
- **AND** the returned command is executed via `cmd.Execute()`

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

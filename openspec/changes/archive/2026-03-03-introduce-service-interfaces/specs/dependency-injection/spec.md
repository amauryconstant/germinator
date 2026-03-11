## Purpose

Extend the dependency injection pattern to include populated service interfaces in ServiceContainer, enabling commands to access services through interfaces rather than package-level functions.

## MODIFIED Requirements

### Requirement: ServiceContainer holds service instances

The system SHALL provide a `ServiceContainer` struct that groups service instances for command use.

#### Scenario: ServiceContainer is embeddable in CommandConfig

- **WHEN** CommandConfig is created
- **THEN** ServiceContainer is available as a field for commands to access services

#### Scenario: ServiceContainer is initially sparse

- **WHEN** ServiceContainer is created
- **THEN** it MAY have zero or more service fields
- **AND** the pattern supports future service additions without breaking changes

#### Scenario: ServiceContainer holds populated interface implementations

- **WHEN** ServiceContainer is created in main.go
- **THEN** it SHALL have `Transformer application.Transformer` field
- **AND** it SHALL have `Validator application.Validator` field
- **AND** it SHALL have `Canonicalizer application.Canonicalizer` field
- **AND** it SHALL have `Initializer application.Initializer` field
- **AND** each field SHALL be populated with concrete implementation

---

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

#### Scenario: Init command constructor

- **WHEN** `NewInitCommand(config)` is called
- **THEN** it returns a `*cobra.Command` configured for resource installation
- **AND** the command uses services from `config.Services`

---

### Requirement: Root command aggregates subcommands

The system SHALL provide `NewRootCommand(config *CommandConfig)` that creates the root command with all subcommands attached.

#### Scenario: Root command includes all subcommands

- **WHEN** `NewRootCommand(config)` is called
- **THEN** the returned command has validate, adapt, canonicalize, init, and version subcommands

#### Scenario: Configuration flows to all commands

- **WHEN** `NewRootCommand(config)` creates subcommands
- **THEN** each subcommand receives the same `config` instance

---

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

#### Scenario: Services are wired in main

- **WHEN** main() executes
- **THEN** `services.NewTransformer()` is called
- **AND** `services.NewValidator()` is called
- **AND** `services.NewCanonicalizer()` is called
- **AND** `services.NewInitializer()` is called
- **AND** each is assigned to ServiceContainer field

---

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

---

## ADDED Requirements

### Requirement: Commands access services through interfaces

Commands SHALL call service methods through the ServiceContainer interfaces, not package-level functions.

#### Scenario: Adapt command uses Transformer interface

- **WHEN** adapt command executes
- **THEN** it SHALL call `cfg.Services.Transformer.Transform(ctx, req)`
- **AND** it SHALL NOT call `services.TransformDocument()` directly

#### Scenario: Validate command uses Validator interface

- **WHEN** validate command executes
- **THEN** it SHALL call `cfg.Services.Validator.Validate(ctx, req)`
- **AND** it SHALL NOT call `services.ValidateDocument()` directly

#### Scenario: Canonicalize command uses Canonicalizer interface

- **WHEN** canonicalize command executes
- **THEN** it SHALL call `cfg.Services.Canonicalizer.Canonicalize(ctx, req)`
- **AND** it SHALL NOT call `services.CanonicalizeDocument()` directly

#### Scenario: Init command uses Initializer interface

- **WHEN** init command executes
- **THEN** it SHALL call `cfg.Services.Initializer.Initialize(ctx, req)`
- **AND** it SHALL NOT call `services.InitializeResources()` directly

---

### Requirement: ServiceContainer constructor populates all services

The system SHALL provide `NewServiceContainer()` that returns a fully populated ServiceContainer.

#### Scenario: NewServiceContainer creates all services

- **WHEN** `NewServiceContainer()` is called
- **THEN** it SHALL return a ServiceContainer with all interface fields populated
- **AND** `Transformer` field SHALL not be nil
- **AND** `Validator` field SHALL not be nil
- **AND** `Canonicalizer` field SHALL not be nil
- **AND** `Initializer` field SHALL not be nil

# cli-factory Specification

## Purpose

Replace the eager `ServiceContainer` and mutable `CommandConfig` with a `cmdutil.Factory` that exposes every shared dependency as a lazy `func() (T, error)` field. The Factory is the only DI mechanism in the new architecture.

## ADDED Requirements

### Requirement: Factory struct holds eager values

The `cmdutil.Factory` struct SHALL hold the always-needed values eagerly.

#### Scenario: Eager fields are populated at construction

- **WHEN** the Factory is constructed
- **THEN** its `IOStreams *iostreams.IOStreams` field SHALL be set to a non-nil value shared across the command tree
- **AND** its `AppVersion string` field SHALL be set to the build-time version
- **AND** its `Executable string` field SHALL be set to `"germinator"`
- **AND** its `RootContext context.Context` field SHALL be set via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`
- **AND** `RootContext` SHALL be cancelled when SIGINT or SIGTERM is received

### Requirement: Factory exposes lazy dependency functions

Every service, loader, or expensive resource SHALL be exposed on the Factory as a `func() (T, error)` field.

#### Scenario: Library is a lazy function

- **WHEN** the Factory is constructed
- **THEN** `Factory.Library func() (*library.Library, error)` SHALL be set
- **AND** the function SHALL NOT be called during Factory construction
- **AND** the function SHALL cache its result via `sync.OnceValues` (Go 1.21+)

#### Scenario: Config is a lazy function

- **WHEN** the Factory is constructed
- **THEN** `Factory.Config func() (*config.Config, error)` SHALL be set to a function that loads the application config from `$XDG_CONFIG_HOME/germinator/config.toml`
- **AND** the function SHALL NOT be called during Factory construction
- **AND** the function SHALL cache its result via `sync.OnceValues`

#### Scenario: Transformer, Validator, Canonicalizer, Initializer are lazy functions

- **WHEN** the Factory is constructed
- **THEN** `Factory.Transformer`, `Factory.Validator`, `Factory.Canonicalizer`, `Factory.Initializer` SHALL all be set to functions that construct their respective services
- **AND** none of the functions SHALL be called during Factory construction

### Requirement: Lazy functions cache via sync.OnceValues

The Factory SHALL use `sync.OnceValues` (Go 1.21+) to cache each lazy function's result so multiple callers in one command invocation receive the same instance.

#### Scenario: Two callers receive the same instance

- **GIVEN** a Factory instance `f` with a `Library` function that increments a counter
- **WHEN** `f.Library()` is called twice
- **THEN** the counter SHALL be `1` (the second call returns the cached value)
- **AND** both calls SHALL return the same `*library.Library` pointer

#### Scenario: Errors are cached

- **GIVEN** a Factory with a `Library` function that returns `(nil, errors.New("disk full"))`
- **WHEN** `f.Library()` is called once
- **THEN** the cache SHALL hold `(nil, errors.New("disk full"))`
- **AND** a subsequent `f.Library()` call SHALL return the same error without re-invoking the function

### Requirement: Lazy functions chain through cached dependencies

A lazy function that depends on another lazy function SHALL call through the Factory function rather than re-implementing the resolution, so the cache is shared.

#### Scenario: Initializer chains through Library and Transformer

- **GIVEN** the Factory's `Initializer` function needs the library and the transformer
- **WHEN** `Initializer()` is called
- **THEN** it SHALL call `f.Library()` and `f.Transformer()` to obtain its dependencies
- **AND** it SHALL NOT load the library or construct the transformer directly

#### Scenario: Transformer chains through Config

- **GIVEN** the Factory's `Transformer` function needs the application config
- **WHEN** `Transformer()` is called
- **THEN** it SHALL call `f.Config()` to obtain the config
- **AND** it SHALL NOT load the config file directly

### Requirement: Cache is per-Factory instance

Each Factory instance SHALL have its own cache. Constructing a new Factory SHALL start with a fresh cache.

#### Scenario: Two Factory instances have independent caches

- **GIVEN** two Factory instances `f1` and `f2` each populated with a `Library` function that increments a counter
- **WHEN** `f1.Library()` and `f2.Library()` are each called once
- **THEN** the counter SHALL be `2` (each Factory calls the function independently)

### Requirement: main.go is the only composition root

Only `main.go` SHALL construct the Factory and populate its fields. No command code, library code, or test file SHALL instantiate a Factory or call a constructor for a service implementation.

#### Scenario: Command code uses Factory, not constructors

- **GIVEN** a command file under `cmd/`
- **WHEN** the file is inspected
- **THEN** it SHALL NOT import any service implementation package for the purpose of calling a constructor
- **AND** it SHALL obtain its dependencies by calling the corresponding function field on the Factory

### Requirement: No global state

The `cmdutil` package SHALL NOT maintain any package-level global variables. All state lives on the Factory instance.

#### Scenario: No package-level Factory

- **WHEN** the `cmdutil` package is inspected
- **THEN** it SHALL NOT contain any package-level variable of type `*Factory` or any mutable package-level service instance

#### Scenario: No SetGlobal functions

- **WHEN** the `cmdutil` package is inspected
- **THEN** it SHALL NOT export any `SetGlobal*` function

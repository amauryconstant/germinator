# cli-factory Specification

## Purpose

Replace the eager `ServiceContainer` and mutable `CommandConfig` with a `cmdutil.Factory` that exposes every shared dependency as a lazy `func() (T, error)` field. The Factory is the only DI mechanism in the new architecture.

## Requirements

### Requirement: Factory struct holds eager values

The `cmdutil.Factory` struct SHALL hold the always-needed values eagerly.

#### Scenario: Eager fields are populated at construction

- **WHEN** the Factory is constructed via `NewFactory(ctx, io, appVersion, executable)`
- **THEN** its `IOStreams *iostreams.IOStreams` field SHALL be set to a non-nil value shared across the command tree
- **AND** its `AppVersion string` field SHALL be set to the build-time version
- **AND** its `Executable string` field SHALL be set to `"germinator"`
- **AND** its `RootContext context.Context` field SHALL be set via `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` (signal handling wrapped around the supplied context)
- **AND** `RootContext` SHALL be cancelled when SIGINT or SIGTERM is received, or when `Factory.Close()` is invoked
- **AND** its `CompletionCache *CompletionCache` field SHALL be assigned alongside the lazy fields in `main.go`

### Requirement: Factory exposes lazy dependency functions

Every service that performs expensive I/O or holds shared mutable state SHALL be exposed on the Factory as a `func() (T, error)` field. The Factory exposes exactly two such fields: `Config` and `Library`.

#### Scenario: Library is a lazy function

- **WHEN** the Factory is constructed
- **THEN** `Factory.Library func() (*library.Library, error)` SHALL be set
- **AND** the function SHALL NOT be called during Factory construction
- **AND** the function SHALL cache its result via `cmdutil.OnceValuesFunc[T]` (a generic helper at `internal/cmdutil/factory.go:71` wrapping `sync.Once`)

#### Scenario: Config is a lazy function

- **WHEN** the Factory is constructed
- **THEN** `Factory.Config func() (*config.Config, error)` SHALL be set to a function that loads the application config from `$XDG_CONFIG_HOME/germinator/config.toml` (falling back to `~/.config/germinator/config.toml`)
- **AND** the function SHALL NOT be called during Factory construction
- **AND** the function SHALL cache its result via `cmdutil.OnceValuesFunc[T]`

#### Scenario: No additional lazy service fields

- **WHEN** the `cmdutil.Factory` type is inspected
- **THEN** there SHALL be NO `Transformer`, `Validator`, `Canonicalizer`, or `Initializer` lazy function field
- **AND** commands that need a transformer/validator/canonicalizer SHALL declare the dependency in their own per-command options struct (e.g., `cmd/adapt.go:Transformer`) with a nil-safe fallback to a local constructor
- **AND** commands that need an initializer (e.g., `cmd/init.go`) SHALL call the package-local `NewInitializer()` constructor directly in their `runXxx` body

### Requirement: Lazy functions cache via cmdutil.OnceValuesFunc

The Factory SHALL use `cmdutil.OnceValuesFunc[T]` (a generic helper at `internal/cmdutil/factory.go:71`) to cache each lazy function's result so multiple callers in one command invocation receive the same instance. The helper wraps `sync.Once` and is the canonical wrapper for any lazy field assigned to the Factory.

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

#### Scenario: Config chains through XDG-aware loader

- **GIVEN** the Factory's `Config` function needs to read from disk
- **WHEN** `Config()` is called
- **THEN** it SHALL call `internal/config.Load()` (or a wrapper) which performs the XDG-aware file lookup
- **AND** the function SHALL NOT re-read the config file directly without going through the loader

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

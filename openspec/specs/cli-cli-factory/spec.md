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

- **GIVEN** the Factory is constructed via `cmdutil.BuildFactory(ctx, io, appVersion, executable)`
- **WHEN** the Factory struct is inspected
- **THEN** `Factory.Library func() (*library.Library, error)` SHALL be exposed with the correct signature
- **AND** the contract test `TestFactory_OnlyConfigAndLibraryAreLazyFields` SHALL continue to pass because the `Library` field exists with the correct signature; only the `BuildFactory` wiring is removed
- **AND** any cmd-side code MAY assign `f.Library` (e.g., for tests or alternative composition roots in `main.go`) without violating this spec — the field is exposed and assignable, not required-to-be-set-by-BuildFactory

#### Scenario: f.Library is nil after BuildFactory returns

- **WHEN** `cmdutil.BuildFactory(ctx, io, appVersion, executable)` returns
- **THEN** `f.Library` SHALL be `nil` (the eager `OnceValuesFunc` wiring that previously captured `f.RootContext` is removed per `propagate-context-through-shell` Decision 6)

#### Scenario: Per-RunE lazy closure captures c.Context()

- **GIVEN** a `RunE` for a command that needs the library
- **WHEN** the command is invoked
- **THEN** the `RunE` SHALL construct a lazy `opts.Library` closure capturing `c.Context()` at `RunE` entry
- **AND** the closure SHALL load the library once per `RunE` invocation and cache the result for subsequent `opts.Library()` calls within the same invocation

#### Scenario: Config is a lazy function

- **WHEN** the Factory is constructed in `main.go`
- **THEN** `Factory.Config func() (*config.Config, error)` SHALL be set to a function that calls `internal/config.Load()` (the top-level wrapper at `internal/config/load.go`) which loads the application config from `$XDG_CONFIG_HOME/germinator/config.toml` (falling back to `~/.config/germinator/config.toml`)
- **AND** the function SHALL NOT be called during Factory construction
- **AND** the function SHALL cache its result via `cmdutil.OnceValuesFunc[T]`
- **AND** `main.go` is the **only** location that assigns `f.Config` (per the `main.go is the only composition root` requirement)

#### Scenario: No additional lazy service fields

- **WHEN** the `cmdutil.Factory` type is inspected
- **THEN** there SHALL be NO `Transformer`, `Validator`, `Canonicalizer`, or `Initializer` lazy function field
- **AND** commands that need a transformer/validator/canonicalizer SHALL declare the dependency in their own per-command options struct (e.g., `cmd/adapt.go:Transformer`) with a nil-safe fallback to a local constructor
- **AND** commands that need an initializer (e.g., `cmd/init.go`) SHALL call the package-local `NewInitializer()` constructor directly in their `runXxx` body

### Requirement: Lazy functions cache via cmdutil.OnceValuesFunc

The Factory SHALL use `cmdutil.OnceValuesFunc[T]` (a generic helper wrapping `sync.Once`) to cache each lazy function's result so multiple callers in one command invocation receive the same instance. The helper is the canonical wrapper for any lazy field assigned to the Factory.

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

#### Scenario: Config chains through the top-level Load wrapper

- **GIVEN** the Factory's `Config` function needs to read from disk
- **WHEN** `Config()` is called
- **THEN** it SHALL call `internal/config.Load()` (the top-level wrapper at `internal/config/load.go`) which delegates to `internal/config.NewConfigManager().Load()`
- **AND** the loader SHALL merge defaults, config file, and `GERMINATOR_*` env vars (last wins)
- **AND** the function SHALL NOT re-read the config file directly without going through the loader

### Requirement: Internal/config.Load top-level wrapper

The `internal/config` package SHALL export a top-level `Load()` function (at `internal/config/load.go`) that callers can use without instantiating a Manager. The wrapper has the same precedence contract as `NewConfigManager().Load()` — defaults → file → env. The Cobra-flag tier is handled per-command and remains outside `Load()`'s scope (the library-path tier is encoded in `library.FindLibrary`, not in `Load()`).

#### Scenario: Load returns the resolved Config

- **WHEN** `config.Load()` is called from `main.go`
- **THEN** it SHALL return `(*Config, error)` where `Config` is the result of the koanf-managed merge (defaults → file → env); the Cobra-flag tier (`application-configuration/spec.md` precedence tier 4) is applied separately by each command and not by `Load()`
- **AND** it SHALL NOT return an error when the config file is missing (defaults apply)
- **AND** `*Config` SHALL be non-nil on both success and error paths; on error, `*Config` holds zero values or partial defaults — the error chain is the authoritative signal
- **AND** it SHALL return a `*core.FileError`, `*core.ParseError`, or `*core.ConfigError` matching the failure mode (file I/O, TOML parse, validation), all dispatched via `errors.As` by `output.FormatError`

#### Scenario: Load caches via Factory

- **GIVEN** `main.go` wires `f.Config = cmdutil.OnceValuesFunc(config.Load)`
- **WHEN** `f.Config()` is called twice in one command invocation
- **THEN** the `OnceValuesFunc` wrapper SHALL invoke the inner `config.Load()` exactly once
- **AND** both calls SHALL return the same `*Config` pointer

### Requirement: Composition root builds Factory

The CLI SHALL wire `Factory.Config` via `cmdutil.BuildFactory(ctx, io, appVersion, executable)` which:

1. Calls `config.Load()` exactly once across all `f.Config()` invocations within a process
2. Activates debug logging via `IOStreams.SetDebug(cfg.Debug)` when `cfg.Debug == true`
3. Returns a non-nil `*Factory` plus any error from the first config load
4. Wires `f.Library` with the priority chain `flag > env > cfg > XDG default` via `library.FindLibrary`
5. Populates `f.CompletionCache` so completion actions have a working cache immediately

`main.go` remains the only place that calls `os.Exit`; `BuildFactory` returns errors that `main.go` maps to exit codes via `cmdutil.ExitCodeFor`.

#### Scenario: Debug activation flows through Config

- **GIVEN** `config.toml` sets `debug = true` (or `GERMINATOR_DEBUG=1`)
- **WHEN** `cmdutil.BuildFactory` runs
- **THEN** `IOStreams.Logger` SHALL be a debug-level handler writing to `ErrOut`

#### Scenario: Config load errors propagate from BuildFactory

- **WHEN** `config.Load()` returns a non-nil error
- **THEN** `BuildFactory` SHALL return that error unchanged so `main.go` can map it to an exit code via `cmdutil.ExitCodeFor`
- **AND** `BuildFactory` SHALL return a non-nil `*Factory` even when config load errors (so the caller can defer `f.Close()`)

#### Scenario: Factory exposes only documented lazy fields

- **WHEN** the `cmdutil.Factory` struct is inspected via reflection
- **THEN** it SHALL expose exactly `Config` and `Library` as exported `func() (T, error)` fields
- **AND** adding `Transformer`, `Validator`, `Canonicalizer`, or `Initializer` lazy fields SHALL fail the contract test (`TestFactory_OnlyConfigAndLibraryAreLazyFields`)

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

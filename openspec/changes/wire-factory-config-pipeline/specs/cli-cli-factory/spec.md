# cli-factory Specification (delta)

## MODIFIED Requirements

### Requirement: Factory exposes lazy dependency functions

Every service that performs expensive I/O or holds shared mutable state SHALL be exposed on the Factory as a `func() (T, error)` field. The Factory exposes exactly two such fields: `Config` and `Library`.

**Change**: clarify that `Config` is wired in `main.go` (the composition root) via `cmdutil.OnceValuesFunc`. Previously, the spec required the field to be set but did not specify the wiring location.

#### Scenario: Library is a lazy function

- **WHEN** the Factory is constructed
- **THEN** `Factory.Library func() (*library.Library, error)` SHALL be set
- **AND** the function SHALL NOT be called during Factory construction
- **AND** the function SHALL cache its result via `cmdutil.OnceValuesFunc[T]` (a generic helper at `internal/cmdutil/factory.go:71` wrapping `sync.Once`)

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

### Requirement: Lazy functions chain through cached dependencies

A lazy function that depends on another lazy function SHALL call through the Factory function rather than re-implementing the resolution, so the cache is shared.

#### Scenario: Config chains through the top-level Load wrapper

- **GIVEN** the Factory's `Config` function needs to read from disk
- **WHEN** `Config()` is called
- **THEN** it SHALL call `internal/config.Load()` (the top-level wrapper at `internal/config/load.go`) which delegates to `internal/config.NewConfigManager().Load()`
- **AND** the loader SHALL merge defaults, config file, and `GERMINATOR_*` env vars (last wins)
- **AND** the function SHALL NOT re-read the config file directly without going through the loader

### Requirement: Internal/config.Load top-level wrapper

The `internal/config` package SHALL export a top-level `Load()` function that callers can use without instantiating a Manager.

**Change**: NEW requirement documenting the wrapper added in change `wire-factory-config-pipeline`. The wrapper exists to satisfy the Factory's `Config` lazy field wiring (`main.go` calls `config.Load()` directly via `cmdutil.OnceValuesFunc`).

#### Scenario: Load returns the resolved Config

- **WHEN** `config.Load()` is called from `main.go`
- **THEN** it SHALL return `(*Config, error)` where `Config` is the result of the four-tier merge (defaults → file → env)
- **AND** it SHALL NOT return an error when the config file is missing (defaults apply)
- **AND** it SHALL return a `*core.ConfigError` if the config file is malformed or contains invalid values

#### Scenario: Load caches via Factory

- **GIVEN** `main.go` wires `f.Config = cmdutil.OnceValuesFunc(config.Load)`
- **WHEN** `f.Config()` is called twice in one command invocation
- **THEN** `config.Load()` SHALL execute exactly once
- **AND** both calls SHALL return the same `*Config` pointer

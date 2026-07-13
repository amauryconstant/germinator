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

The `internal/config` package SHALL export a top-level `Load()` function (at `internal/config/load.go`) that callers can use without instantiating a Manager. The wrapper has the same precedence contract as `NewConfigManager().Load()` — defaults → file → env. The Cobra-flag tier is handled per-command and remains outside `Load()`'s scope (the library-path tier is encoded in `library.FindLibrary`, not in `Load()`).

**Change**: NEW requirement documenting the wrapper added in change `wire-factory-config-pipeline`. The wrapper exists to satisfy the Factory's `Config` lazy field wiring (`main.go` calls `config.Load()` directly via `cmdutil.OnceValuesFunc`).

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

**Change**: NEW requirement extracted from `main.go`'s inline wiring to make composition-root behavior testable without `main_test.go`.

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

# cli-cli-factory Specification (delta)

## MODIFIED Requirements

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

**Change**: MODIFIED the "SHALL be set" wording at `cli-cli-factory/spec.md:31` to relax the BuildFactory-assignment requirement while preserving the field-presence and signature-presence contract. The pre-change behavior was `f.Library = OnceValuesFunc(...)` in `BuildFactory`, capturing `f.RootContext` at construction time — defeating per-command cancellation. The post-change behavior is `f.Library = nil` after `BuildFactory`; each `RunE` builds its own per-RunE-cached lazy loader that captures `c.Context()` at construction (see `propagate-context-through-shell` Design Decision 4). The cache scope shrinks from per-Factory to per-RunE — the only lost behavior is sharing the library instance across multiple RunE invocations within one Factory instance, which never occurs because the Factory is constructed once per CLI invocation. This delta formalizes the new contract.

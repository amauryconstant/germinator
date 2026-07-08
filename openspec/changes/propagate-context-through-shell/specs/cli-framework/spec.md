# cli-framework Specification (delta)

## ADDED Requirements

### Requirement: I/O adapter ctx propagation

Service-style I/O adapters (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`, and the per-resource adders) — implemented in `internal/{transform,validate,canonicalize,install}/` shell packages and as `*Library` methods on `internal/library/library.go` — SHALL accept `ctx context.Context` as the first parameter of every public method. The `ctx` SHALL be forwarded to all downstream calls (`parser.LoadDocument`, `renderer.RenderDocument`, `LoadLibrary`, `SaveLibrary`, `*Library.Refresh`, `*Library.RemoveResource`, etc.).

The adapter SHALL NOT discard `ctx` (e.g., via `_ context.Context`). If a method does not need cancellation, the `ctx` SHALL still be accepted (and may be ignored), so the call site is uniform across the codebase.

**Change**: NEW requirement. The pre-change adapters in `cmd/{initializer,transformer,canonicalize,validate}.go` accepted `ctx` as a method parameter but discarded it via the `_` underscore binding. The `extract-io-adapters` change relocates the adapters to `internal/<x>/` shell packages; this change threads `ctx` through, fulfilling the spec promise at `openspec/changes/extract-io-adapters/specs/cli-framework/spec.md:42`.

#### Scenario: Service method accepts ctx as first parameter

- **WHEN** an adapter's public method is called
- **THEN** the method signature SHALL have `ctx context.Context` as the first parameter
- **AND** the method SHALL forward `ctx` to all downstream `parser.*` / `renderer.*` / `library.*` calls
- **AND** the method SHALL NOT use `context.Background()` or any other synthesized context in place of the caller's `ctx`

#### Scenario: Cancellation propagates through the adapter

- **GIVEN** a cmd-side `ctx` that is cancelled mid-call
- **WHEN** the adapter method is called with that `ctx`
- **THEN** the call SHALL return within bounded time (verified by `goleak` and `-race` tests)
- **AND** the returned error SHALL be `context.Canceled` or `context.DeadlineExceeded` (or wrap one of them)

#### Scenario: Adapter signature inspection

- **WHEN** the codebase is searched for `func \w+\(_ context\.Context` in `internal/{transform,validate,canonicalize,install}/` and in `*Library` methods
- **THEN** zero matches SHALL appear (no discarded `ctx` parameters)
- **AND** every public method SHALL have `ctx context.Context` as a named, non-underscore parameter

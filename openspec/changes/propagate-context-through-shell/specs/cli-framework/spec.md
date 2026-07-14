# cli-framework Specification (delta)

## ADDED Requirements

### Requirement: I/O adapter ctx propagation

Service-style I/O adapters (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`, and the per-resource adders) — implemented in `internal/{transform,validate,canonicalize,install}/` shell packages and as `*Library` methods on `internal/library/library.go` — SHALL accept `ctx context.Context` as the first parameter of every public method. The `ctx` SHALL be forwarded to all downstream calls (`parser.LoadDocument`, `renderer.RenderDocument`, `LoadLibrary`, `SaveLibrary`, `*Library.Refresh`, `*Library.RemoveResource`, `*Library.ResolvePreset`, etc.).

The adapter SHALL NOT discard `ctx` (e.g., via `_ context.Context`). If a method does not need cancellation, the `ctx` SHALL still be accepted (and may be ignored), so the call site is uniform across the codebase. The canonical example of the accept-and-may-ignore pattern is `(*Library).ResolvePreset` at `library-library-resolution/spec.md` — a pure in-memory map lookup that accepts `ctx` for spec symmetry with no I/O to forward to today.

**Change**: NEW requirement. The pre-change adapters in `cmd/{initializer,transformer,canonicalize,validate}.go` accepted `ctx` as a method parameter but discarded it via the `_` underscore binding. The `extract-io-adapters` change relocates the adapters to `internal/<x>/` shell packages; this change threads `ctx` through, fulfilling the spec promise at `openspec/changes/extract-io-adapters/specs/cli-framework/spec.md:42`. The `internal/library/resolver.go:67` `ResolvePreset` underscore binding is also replaced with `ctx context.Context` for spec symmetry.

#### Scenario: Service method accepts ctx as first parameter

- **WHEN** an adapter's public method is called
- **THEN** the method signature SHALL have `ctx context.Context` as the first parameter
- **AND** the method SHALL forward `ctx` to all downstream `parser.*` / `renderer.*` / `library.*` calls
- **AND** any `ctx.Err()` check site SHALL wrap the sentinel with `%w` (e.g., `fmt.Errorf("...: %w", ctx.Err())`) so callers can `errors.Is(err, context.Canceled)`
- **AND** the method SHALL NOT use `context.Background()` or `context.TODO()` in place of the caller's `ctx` — synthesizing any context inside a request path violates the `golang-context` best practice of "never create a new context in the middle of a request path"

#### Scenario: Cancellation propagates through the adapter

- **GIVEN** a cmd-side `ctx` that is cancelled mid-call
- **WHEN** the adapter method is called with that `ctx`
- **THEN** the call SHALL return within bounded time (verified by `goleak` and `-race` tests)
- **AND** the returned error SHALL be `context.Canceled` or `context.DeadlineExceeded` (or wrap one of them via `%w`)

#### Scenario: Adapter signature inspection

- **WHEN** `mise run lint` runs `staticcheck` (enabled via `.golangci.yml:14`) against `internal/{transform,validate,canonicalize,install}/` (once `extract-io-adapters` lands; otherwise the legacy `cmd/{transformer,validate,canonicalize,initializer}.go` paths) and the `*Library` methods in `internal/library/library.go` and `internal/library/resolver.go`
- **THEN** zero `lostcancel` (go vet) violations SHALL appear
- **AND** zero `_ context\.Context` underscore-binding matches SHALL appear in the production-code scope (verified by `rg --type=go -g '!*_test.go' "_ context\.Context" cmd/ internal/`)
- **AND** zero `context\.TODO\(\)` matches SHALL appear in the production-code scope (verified by `rg --type=go -g '!*_test.go' "context\.TODO\(\)" cmd/ internal/`) — the `internal/library/resolver.go:54-56` package-level `ResolvePreset` shim was deleted by `propagate-context-through-shell` task 3.4b; this rg ensures no replacement shim is introduced
- **AND** every public method on the adapters SHALL have `ctx context.Context` as a named, non-underscore parameter

> **Note:** The pre-change adapters in `cmd/{initializer,transformer,canonicalize,validate}.go` accepted `ctx` but discarded it via the `_` underscore binding. After `extract-io-adapters` relocates the adapters to `internal/<x>/` shell packages, this scenario verifies the future-proofed state. The scenario applies regardless of whether the adapter is currently in `cmd/` (legacy) or `internal/<x>/` (post-extraction). Test fakes in `cmd/*_test.go` are exempt from the underscore-binding check (the rg glob excludes `*_test.go`); test fakes renamed to `ctx context.Context` (zero behavior change) are encouraged but not required.

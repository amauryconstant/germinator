# library-library-resource-import Specification (delta)

## ADDED Requirements

### Requirement: AddResource forwards ctx to underlying method

The `library.AddResource` package-level function SHALL forward the caller's `ctx context.Context` to all I/O (file reads, `LoadLibrary`, `SaveLibrary`) and to the per-resource `libraryAdapter.AddResource(ctx, ...)` in `cmd/library_add.go:82` (and to the eventual `*Library.Add(ctx, *AddRequest) error` method introduced by `extract-io-adapters` Stage 2). The method/adapter SHALL NOT synthesize a `context.Background()` or `context.TODO()` in place of the caller's `ctx`.

**Change**: CLARIFY the ctx propagation contract. The pre-change `AddResource` already accepted `ctx` (per slice 7) and the package-level signature is unchanged. The change enforces that the `libraryAdapter` and the future `*Library.Add` method forward the same `ctx` end-to-end. Any `ctx.Err()` check site SHALL wrap the sentinel with `%w` so callers can `errors.Is(err, context.Canceled)`.

#### Scenario: AddResource forwards ctx

- **WHEN** `library.AddResource(ctx, opts, stdout)` is called
- **THEN** the function SHALL call `lib.Add(ctx, req)` with the same `ctx` (not a synthesized one)
- **AND** the `*Library.Add` method SHALL use that `ctx` for all I/O
- **AND** no `context.Background()` or `context.TODO()` synthesis SHALL appear in the call chain (verified by `rg --type=go -g '!*_test.go' "context\.(Background|TODO)\(\)" internal/library/adder.go`)

#### Scenario: Cancellation during add

- **GIVEN** a `ctx` that is cancelled mid-add
- **WHEN** `library.AddResource(ctx, opts, stdout)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant via `%w`) within bounded time

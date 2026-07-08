# library-library-resource-import Specification (delta)

## MODIFIED Requirements

### Requirement: AddResource forwards ctx to underlying method

The `library.AddResource` package-level function SHALL accept `ctx context.Context` as the first parameter and SHALL forward it to the `*Library.Add(ctx, *AddRequest) error` method. The method SHALL use the caller's `ctx` for any I/O (file reads, `LoadLibrary`, `SaveLibrary`).

**Change**: CLARIFY the ctx propagation contract. The pre-change `AddResource` accepted `ctx` (no signature change needed) but the underlying `*Library.Add` method (introduced by `extract-io-adapters` Stage 2) may have used `context.Background()` internally. The post-change method uses the caller's `ctx` throughout.

#### Scenario: AddResource forwards ctx

- **WHEN** `library.AddResource(ctx, opts, stdout)` is called
- **THEN** the function SHALL call `lib.Add(ctx, req)` with the same `ctx` (not a synthesized one)
- **AND** the `*Library.Add` method SHALL use that `ctx` for all I/O

#### Scenario: Cancellation during add

- **GIVEN** a `ctx` that is cancelled mid-add
- **WHEN** `library.AddResource(ctx, opts, stdout)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant) within bounded time

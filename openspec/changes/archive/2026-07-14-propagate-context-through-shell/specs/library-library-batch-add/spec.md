# library-library-batch-add Specification (delta)

## ADDED Requirements

### Requirement: BatchAddResources forwards ctx to underlying method

The `library.BatchAddResources` package-level function SHALL forward the caller's `ctx context.Context` to all I/O (file reads, `LoadLibrary`, `SaveLibrary`) and to the eventual `*Library.BatchAddResources(ctx, *BatchAddOptions, io.Writer) (*BatchAddResult, error)` method introduced by `extract-io-adapters` Stage 2. The method SHALL NOT synthesize `context.Background()` or `context.TODO()` in place of the caller's `ctx` — both violate the `golang-context` best practice of "never create a new context in the middle of a request path."

**Change**: CLARIFY the ctx propagation contract. The pre-change `BatchAddResources` already accepted `ctx` (per slice 7) and the package-level signature is unchanged. The change enforces forward propagation end-to-end. Any `ctx.Err()` check site SHALL wrap the sentinel with `%w` so callers can `errors.Is(err, context.Canceled)`.

#### Scenario: BatchAddResources forwards ctx

- **WHEN** `library.BatchAddResources(ctx, opts, stdout)` is called
- **THEN** the function SHALL call `lib.BatchAddResources(ctx, opts, stdout)` with the same `ctx`
- **AND** the `*Library.BatchAddResources` method SHALL use that `ctx` for all file reads, `LoadLibrary`, and `SaveLibrary` calls
- **AND** no `context.Background()` or `context.TODO()` synthesis SHALL appear in the call chain (verified by `rg --type=go -g '!*_test.go' "context\.(Background|TODO)\(\)" internal/library/`)

#### Scenario: Cancellation during batch add

- **GIVEN** a `ctx` that is cancelled mid-batch
- **WHEN** `library.BatchAddResources(ctx, opts, stdout)` is called
- **THEN** the function SHALL return within bounded time
- **AND** the returned error SHALL wrap `context.Canceled` via `%w` if cancellation was the cause

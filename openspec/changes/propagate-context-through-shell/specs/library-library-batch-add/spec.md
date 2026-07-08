# library-library-batch-add Specification (delta)

## MODIFIED Requirements

### Requirement: BatchAddResources forwards ctx to underlying method

The `library.BatchAddResources` package-level function SHALL accept `ctx context.Context` as the first parameter and SHALL forward it to the `*Library.BatchAddResources(ctx, *BatchAddOptions, io.Writer) (*BatchAddResult, error)` method. The method SHALL use the caller's `ctx` for all I/O.

**Change**: CLARIFY the ctx propagation contract.

#### Scenario: BatchAddResources forwards ctx

- **WHEN** `library.BatchAddResources(ctx, opts, stdout)` is called
- **THEN** the function SHALL call `lib.BatchAddResources(ctx, opts, stdout)` with the same `ctx`
- **AND** the `*Library.BatchAddResources` method SHALL use that `ctx` for all file reads, `LoadLibrary`, and `SaveLibrary` calls

#### Scenario: Cancellation during batch add

- **GIVEN** a `ctx` that is cancelled mid-batch
- **WHEN** `library.BatchAddResources(ctx, opts, stdout)` is called
- **THEN** the function SHALL return within bounded time
- **AND** the returned error SHALL wrap `context.Canceled` if cancellation was the cause

# library-library-orphan-discovery Specification (delta)

## MODIFIED Requirements

### Requirement: DiscoverOrphans forwards ctx to errgroup

The `library.DiscoverOrphans` function SHALL accept `ctx context.Context` as the first parameter and SHALL forward it to the `errgroup.WithContext(ctx)` (per change `fix-library-io-discipline`). The errgroup-derived child `ctx` SHALL be checked at every directory scan. Cancellation of the parent `ctx` SHALL propagate to all goroutines within bounded time.

**Change**: CLARIFY the ctx propagation contract. The pre-change `DiscoverOrphans` accepted `ctx` but only checked it at the top of the directory loop (BCD-009 / C-018). The post-change errgroup path checks it per-directory.

#### Scenario: DiscoverOrphans uses errgroup-derived ctx

- **WHEN** `library.DiscoverOrphans(ctx, opts)` is called
- **THEN** the function SHALL use `errgroup.WithContext(ctx)` to derive a child context
- **AND** the function SHALL check the child `ctx.Err()` before processing each directory
- **AND** the function SHALL return the errgroup's `Wait()` error (which is `context.Canceled` or `context.DeadlineExceeded` if the parent `ctx` is cancelled)

#### Scenario: Cancellation during directory scan

- **GIVEN** a `ctx` that is cancelled mid-scan
- **WHEN** `library.DiscoverOrphans(ctx, opts)` is called
- **THEN** the function SHALL return within bounded time (one directory's processing time, not the full scan)
- **AND** the returned error SHALL wrap `context.Canceled`

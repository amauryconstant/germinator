# library-library-orphan-discovery Specification (delta)

## ADDED Requirements

### Requirement: Per-directory cancellation via errgroup

The `DiscoverOrphans` function SHALL wrap its directory scan in `errgroup.WithContext` so that `ctx` cancellation is respected at the directory level, not just at the top of the loop. Each goroutine SHALL check `ctx.Err()` before processing its directory. Concurrency SHALL be capped at 8 workers via a semaphore channel to bound memory usage on libraries with many nested directories.

**Change**: NEW requirement. The pre-change implementation checked `ctx.Err()` only at the top of the directory loop; per-directory cancellation was not respected, leading to multi-second delays when a large directory was being scanned.

#### Scenario: Cancellation during directory scan

- **GIVEN** a library with nested directories (e.g., `skills/sub1/`, `skills/sub1/sub2/`, `skills/sub1/sub2/sub3/`) and a `ctx` that is cancelled mid-scan
- **WHEN** `DiscoverOrphans(ctx, opts)` is called
- **THEN** the function SHALL return as soon as the cancellation is observed (within one directory's processing time, not waiting for the full scan to complete)
- **AND** the returned error SHALL wrap `context.Canceled` or `context.DeadlineExceeded` as appropriate

#### Scenario: Errgroup concurrency cap

- **WHEN** the library has more than 8 directories at the top level
- **THEN** the errgroup SHALL process at most 8 directories concurrently
- **AND** the remaining directories SHALL queue until a worker is free
- **AND** the cap SHALL be implemented via a buffered channel of size 8 acquired before the goroutine and released after

#### Scenario: No goroutine leak on cancellation

- **GIVEN** a `ctx` that is cancelled mid-scan
- **WHEN** `DiscoverOrphans` returns
- **THEN** no goroutines SHALL remain blocked on the errgroup's wait group
- **AND** the function SHALL return within a bounded time (verified by `goleak` or race detector in tests)

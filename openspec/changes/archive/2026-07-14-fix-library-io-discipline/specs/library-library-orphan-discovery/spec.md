# library-library-orphan-discovery Specification (delta)

## ADDED Requirements

### Requirement: Recursive cancellation via errgroup in scanDirectory

The `scanDirectory` function SHALL wrap its recursive subtree descent in `errgroup.SetLimit(scanConcurrencyLimit)` so that `ctx` cancellation is observed at the next sibling-subtree yield. Each goroutine SHALL check `ctx.Err()` before processing its subtree. Concurrency SHALL be capped at `scanConcurrencyLimit` (declared as `const scanConcurrencyLimit = 8` at file scope in `internal/library/adder.go`) concurrent subtree workers via `errgroup.SetLimit(scanConcurrencyLimit)`.

The outer `DiscoverOrphans` 4-directory loop (`skills` / `agents` / `commands` / `memory`) is **not** wrapped in errgroup — N=4 fails the `golang-cli-architecture` skill's `N>10` errgroup trigger; the per-iteration ctx check at the top of the outer loop is sufficient.

Shared `*DiscoverResult` writes — slice appends to `Orphans` and `Conflicts`, AND the `Summary.TotalScanned++` integer increment at `adder.go:868` — SHALL all be guarded by a single `sync.Mutex`. The mutex is the cheapest option and matches the codebase's existing `sync.Once` precedent at `internal/cmdutil/factory.go`. Per `golang-safety`, concurrent writes to shared slice backing arrays and integer fields are unsafe without serial access.

**Change**: NEW requirement. The pre-change implementation used sequential `filepath.WalkDir` inside `scanDirectory`, so sibling subtrees were walked one-at-a-time and `ctx.Err()` checks happened per-file. The refactor parallelizes sibling-subtree descent so a deeply nested library (`skills/sub1/sub2/sub3/...`) returns control to the caller as soon as the next goroutine observes `ctx.Err()`.

#### Scenario: Cancellation during deep subtree scan

- **GIVEN** a library with deeply nested directories (e.g., `skills/sub1/.../sub10/`, at minimum 10 levels) and a `ctx` that is cancelled mid-scan
- **WHEN** `DiscoverOrphans(ctx, opts)` is called
- **THEN** the function SHALL return as soon as the errgroup's `WithContext` propagation surfaces the cancellation (typically within one subtree's processing time, not after the full scan completes)
- **AND** the returned error SHALL wrap `context.Canceled` or `context.DeadlineExceeded` as appropriate
- **AND** the partial `*DiscoverResult` accumulated before cancellation SHALL remain available for inspection by the caller

#### Scenario: Errgroup concurrency cap via SetLimit(scanConcurrencyLimit)

- **WHEN** `scanDirectory` encounters more than 8 sibling subdirectories at any level
- **THEN** the errgroup SHALL process at most 8 sibling subtrees concurrently
- **AND** the remaining subtrees SHALL queue until a worker is free
- **AND** the cap SHALL be implemented via `errgroup.SetLimit(scanConcurrencyLimit)` (the built-in worker pool, idiomatic per `golang-cli-architecture`), where `scanConcurrencyLimit` is a file-scope `const` in `internal/library/adder.go`

#### Scenario: No goroutine leak on cancellation

- **GIVEN** a `ctx` that is cancelled mid-scan
- **WHEN** `DiscoverOrphans` returns
- **THEN** no goroutines SHALL remain blocked on the errgroup's wait group
- **AND** the function SHALL return within a bounded time (verified by `goleak` or race detector in tests)

#### Scenario: Order of result.Orphans is unordered

- **WHEN** `DiscoverOrphans` returns a non-empty `*DiscoverResult`
- **THEN** the order of entries in `result.Orphans` is implementation-defined (driven by goroutine scheduling)
- **AND** consumers SHALL NOT assume directory-then-file order
- **AND** summary statistics (TotalScanned, TotalOrphans) are unaffected by ordering

#### Scenario: Summary counter is thread-safe under parallel scan

- **WHEN** `scanDirectory` runs sibling subtrees in parallel via `errgroup.SetLimit(scanConcurrencyLimit)`
- **AND** each goroutine increments `result.Summary.TotalScanned` per `.md` file processed
- **THEN** the final `TotalScanned` SHALL equal the count of `.md` files processed (no lost increments due to concurrent writes)
- **AND** the increment SHALL be covered by the same `sync.Mutex` as the slice appends

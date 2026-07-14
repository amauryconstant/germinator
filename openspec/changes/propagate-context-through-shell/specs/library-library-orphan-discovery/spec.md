# library-library-orphan-discovery Specification (delta)

## ADDED Requirements

### Requirement: DiscoverOrphans respects ctx at every directory scan

The `library.DiscoverOrphans` function SHALL honor caller-supplied cancellation by checking `ctx.Err()` at every directory scan entry and between every per-file walker entry. The function SHALL return wrapped `ctx.Err()` (using `%w`) on cancellation. The function SHALL NOT synthesize `context.Background()` or `context.TODO()` in place of the caller's `ctx`.

**Change**: CLARIFY the ctx propagation contract. The pre-change `DiscoverOrphans` already accepted `ctx` (per slice 7, see `internal/library/adder.go:785`) and performed sequential `ctx.Err()` checks per-directory and per-file (see `adder.go:803,819,834,863`). This delta codifies that pattern as the contract. Earlier wording referenced `errgroup.WithContext(ctx)`; that refactor is deferred because the sequential pattern is sufficient for the 4-directory scan (latency delta is sub-millisecond, not user-perceptible per `golang-cli-architecture/references/06-concurrency.md`).

#### Scenario: DiscoverOrphans checks ctx before each directory

- **WHEN** `library.DiscoverOrphans(ctx, opts)` is called
- **THEN** the function SHALL check `ctx.Err()` before processing each top-level directory (`skills`, `agents`, `commands`, `memory`)
- **AND** the per-file recursive walker SHALL check `ctx.Err()` between file entries
- **AND** the function SHALL return wrapped `ctx.Err()` (via `%w`) on cancellation
- **AND** no `context.Background()` or `context.TODO()` synthesis SHALL appear in the call chain (verified by `rg --type=go -g '!*_test.go' "context\.(Background|TODO)\(\)" internal/library/adder.go`)

#### Scenario: Cancellation during directory scan

- **GIVEN** a `ctx` that is cancelled mid-scan
- **WHEN** `library.DiscoverOrphans(ctx, opts)` is called
- **THEN** the function SHALL return within bounded time (one directory's processing time, not the full scan)
- **AND** the returned error SHALL wrap `context.Canceled` via `%w`

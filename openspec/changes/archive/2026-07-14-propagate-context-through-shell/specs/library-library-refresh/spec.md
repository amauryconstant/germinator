# library-library-refresh Specification (delta)

## ADDED Requirements

### Requirement: RefreshLibrary accepts ctx

The `library.RefreshLibrary` package-level function SHALL accept `ctx context.Context` as the first parameter. The function delegates to a `*Library.Refresh(ctx, *RefreshRequest) (*RefreshResult, error)` method, and the method SHALL use the caller's `ctx` for any I/O (file reads, `SaveLibrary`). The method SHALL NOT synthesize `context.Background()` or `context.TODO()` in place of the caller's `ctx`.

**Change**: ADDED the `ctx` parameter requirement. The pre-change `RefreshLibrary` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/refresher.go:60` (line 59 marker, function at line 58). The marker is removed in change `propagate-context-through-shell`. Any `ctx.Err()` check site SHALL wrap the sentinel with `%w`.

#### Scenario: RefreshLibrary signature

- **WHEN** `library.RefreshLibrary(ctx, opts)` is called
- **THEN** the function SHALL return `(*RefreshResult, error)`
- **AND** the function SHALL forward `ctx` to the `*Library.Refresh` method
- **AND** the function SHALL NOT use `context.Background()` or `context.TODO()` in place of the caller's `ctx`

#### Scenario: Cancellation during refresh

- **GIVEN** a `ctx` that is cancelled mid-refresh
- **WHEN** `library.RefreshLibrary(ctx, opts)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant via `%w`) within bounded time

#### Scenario: No TODO(slice-7) markers in library

- **WHEN** the codebase is searched for `TODO(slice-7)` in `internal/library/`
- **THEN** zero matches SHALL appear (the slice-7 debt is fully retired)

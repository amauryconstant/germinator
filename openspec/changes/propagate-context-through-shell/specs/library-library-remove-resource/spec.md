# library-library-remove-resource Specification (delta)

## ADDED Requirements

### Requirement: RemoveResource accepts ctx

The `library.RemoveResource` package-level function SHALL accept `ctx context.Context` as the first parameter. The function delegates to a `*Library.RemoveResource(ctx, *RemoveResourceRequest) error` method, and the method SHALL use the caller's `ctx` for any I/O (file removals, `SaveLibrary`). The method SHALL NOT synthesize `context.Background()` or `context.TODO()` in place of the caller's `ctx`.

**Change**: ADDED the `ctx` parameter requirement. The pre-change `RemoveResource` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/remover.go:64` (line 63 marker, function at line 62). The marker is removed in change `propagate-context-through-shell`. Any `ctx.Err()` check site SHALL wrap the sentinel with `%w`.

#### Scenario: RemoveResource signature

- **WHEN** `library.RemoveResource(ctx, opts)` is called
- **THEN** the function SHALL return `(*RemoveResourceOutput, error)`
- **AND** the function SHALL forward `ctx` to the `*Library.RemoveResource` method
- **AND** the function SHALL NOT use `context.Background()` or `context.TODO()` in place of the caller's `ctx`

#### Scenario: Cancellation during remove

- **GIVEN** a `ctx` that is cancelled mid-remove
- **WHEN** `library.RemoveResource(ctx, opts)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant via `%w`) within bounded time

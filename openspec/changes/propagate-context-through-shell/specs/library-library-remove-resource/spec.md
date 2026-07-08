# library-library-remove-resource Specification (delta)

## MODIFIED Requirements

### Requirement: RemoveResource accepts ctx

The `library.RemoveResource` package-level function SHALL accept `ctx context.Context` as the first parameter. The function delegates to a `*Library.RemoveResource(ctx, *RemoveResourceRequest) error` method, and the method SHALL use the caller's `ctx` for any I/O (file removals, `SaveLibrary`).

**Change**: ADDED the `ctx` parameter requirement. The pre-change `RemoveResource` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/remover.go:64`. The marker is removed in change `propagate-context-through-shell`.

#### Scenario: RemoveResource signature

- **WHEN** `library.RemoveResource(ctx, opts)` is called
- **THEN** the function SHALL return `error`
- **AND** the function SHALL forward `ctx` to the `*Library.RemoveResource` method
- **AND** the function SHALL NOT use `context.Background()` in place of the caller's `ctx`

#### Scenario: Cancellation during remove

- **GIVEN** a `ctx` that is cancelled mid-remove
- **WHEN** `library.RemoveResource(ctx, opts)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant) within bounded time

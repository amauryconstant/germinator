# library-library-remove-preset Specification (delta)

## MODIFIED Requirements

### Requirement: RemovePreset accepts ctx

The `library.RemovePreset` package-level function SHALL accept `ctx context.Context` as the first parameter. The function delegates to a `*Library.RemovePreset(ctx, *RemovePresetRequest) error` method, and the method SHALL use the caller's `ctx` for any I/O (`SaveLibrary`).

**Change**: ADDED the `ctx` parameter requirement. The pre-change `RemovePreset` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/remover.go:129`. The marker is removed in change `propagate-context-through-shell`.

#### Scenario: RemovePreset signature

- **WHEN** `library.RemovePreset(ctx, opts)` is called
- **THEN** the function SHALL return `error`
- **AND** the function SHALL forward `ctx` to the `*Library.RemovePreset` method
- **AND** the function SHALL NOT use `context.Background()` in place of the caller's `ctx`

#### Scenario: Cancellation during remove-preset

- **GIVEN** a `ctx` that is cancelled mid-remove
- **WHEN** `library.RemovePreset(ctx, opts)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant) within bounded time

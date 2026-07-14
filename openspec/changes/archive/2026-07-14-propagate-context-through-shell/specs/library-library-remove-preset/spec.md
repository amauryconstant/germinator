# library-library-remove-preset Specification (delta)

## ADDED Requirements

### Requirement: RemovePreset accepts ctx

The `library.RemovePreset` package-level function SHALL accept `ctx context.Context` as the first parameter. The function delegates to a `*Library.RemovePreset(ctx, *RemovePresetRequest) error` method, and the method SHALL use the caller's `ctx` for any I/O (`SaveLibrary`). The method SHALL NOT synthesize `context.Background()` or `context.TODO()` in place of the caller's `ctx`.

**Change**: ADDED the `ctx` parameter requirement. The pre-change `RemovePreset` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/remover.go:129` (line 128 marker, function at line 127). The marker is removed in change `propagate-context-through-shell`. Any `ctx.Err()` check site SHALL wrap the sentinel with `%w`.

#### Scenario: RemovePreset signature

- **WHEN** `library.RemovePreset(ctx, opts)` is called
- **THEN** the function SHALL return `(*RemovePresetOutput, error)`
- **AND** the function SHALL forward `ctx` to the `*Library.RemovePreset` method
- **AND** the function SHALL NOT use `context.Background()` or `context.TODO()` in place of the caller's `ctx`

#### Scenario: Cancellation during remove-preset

- **GIVEN** a `ctx` that is cancelled mid-remove
- **WHEN** `library.RemovePreset(ctx, opts)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant via `%w`) within bounded time

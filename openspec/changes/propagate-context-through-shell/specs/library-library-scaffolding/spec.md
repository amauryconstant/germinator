# library-library-scaffolding Specification (delta)

## ADDED Requirements

### Requirement: CreateLibrary accepts ctx

The `library.CreateLibrary` package-level function SHALL accept `ctx context.Context` as the first parameter. The function SHALL use the caller's `ctx` for any I/O (`LoadLibrary` for pre-flight validation, `os.Stat` for existence checks). The function SHALL NOT synthesize `context.Background()` or `context.TODO()` in place of the caller's `ctx`.

**Change**: ADDED the `ctx` parameter requirement (per Go convention: ctx first). The pre-change `CreateLibrary` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/creator.go:71` (line 70 marker, function at line 29). The marker is removed in change `propagate-context-through-shell`. The function also has an `io.Writer` parameter for dry-run output (per change `fix-library-io-discipline`, already applied). This change reorders the signature to ctx-first: `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error` per Go convention (design Decision 1). Any `ctx.Err()` check site SHALL wrap the sentinel with `%w` so callers can `errors.Is(err, context.Canceled)`.

#### Scenario: CreateLibrary signature

- **WHEN** `library.CreateLibrary(ctx, opts, stdout)` is called
- **THEN** the function SHALL return `error`
- **AND** the function SHALL forward `ctx` to `LoadLibrary` and any other I/O calls
- **AND** the function SHALL NOT use `context.Background()` or `context.TODO()` in place of the caller's `ctx`
- **AND** the function SHALL write dry-run output to `stdout` (not `os.Stdout` directly)

#### Scenario: Cancellation during init

- **GIVEN** a `ctx` that is cancelled mid-init
- **WHEN** `library.CreateLibrary(ctx, opts, stdout)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant via `%w`) within bounded time

# library-library-scaffolding Specification (delta)

## MODIFIED Requirements

### Requirement: CreateLibrary accepts ctx

The `library.CreateLibrary` package-level function SHALL accept `ctx context.Context` as the first parameter. The function SHALL use the caller's `ctx` for any I/O (`LoadLibrary` for pre-flight validation, `os.Stat` for existence checks).

**Change**: ADDED the `ctx` parameter requirement. The pre-change `CreateLibrary` hard-coded `context.Background()` behind a `// TODO(slice-7)` marker at `internal/library/creator.go:71`. The marker is removed in change `propagate-context-through-shell`. The function also gains an `io.Writer` parameter for dry-run output (per change `fix-library-io-discipline`); the combined signature is `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error`.

#### Scenario: CreateLibrary signature

- **WHEN** `library.CreateLibrary(ctx, opts, stdout)` is called
- **THEN** the function SHALL return `error`
- **AND** the function SHALL forward `ctx` to `LoadLibrary` and any other I/O calls
- **AND** the function SHALL NOT use `context.Background()` in place of the caller's `ctx`
- **AND** the function SHALL write dry-run output to `stdout` (not `os.Stdout` directly)

#### Scenario: Cancellation during init

- **GIVEN** a `ctx` that is cancelled mid-init
- **WHEN** `library.CreateLibrary(ctx, opts, stdout)` is called
- **THEN** the function SHALL return `context.Canceled` (or a wrapped variant) within bounded time

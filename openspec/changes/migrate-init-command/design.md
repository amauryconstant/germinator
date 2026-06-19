# Design — Migrate init command

## Context

The `init` command is the first germinator command to use the `core.PartialSuccessError` sentinel (defined in change-1 and recognized by `cmdutil.ExitCodeFor`). It is also the first command that needs a per-resource error list as part of its result type.

## Goals / Non-Goals

**Goals:**

- `cmd/init.go` follows the `NewCmdInit(f, runF) + runInit(opts)` pattern.
- Partial success returns `*core.PartialSuccessError`; `cmdutil.ExitCodeFor` returns 0 for `Succeeded > 0`.
- Preset expansion (e.g. `--preset git-workflow` → list of refs) is preserved.
- Dry-run and `--force` flags are preserved.
- Output format is human-readable per-resource status; JSON output via `--output json` is NOT added (init doesn't produce structured output suitable for JSON).

**Non-Goals:**

- Migrating library commands — changes 4, 6, 7.
- Adding `--output` flag to init — it produces per-resource status text, not structured data.
- Restructuring preset expansion — uses the existing library `ResolvePreset(ref)` method.

## Decisions

### 1. `runInit` returns `*core.PartialSuccessError` on partial success

**Choice**: After processing all refs, `runInit` returns:
- `nil` if all succeeded
- `*core.PartialSuccessError{Succeeded: n, Failed: 0}` if all succeeded (NOT USED — return nil instead; this case is impossible)
- `*core.PartialSuccessError{Succeeded: m, Failed: n}` if some succeeded and some failed (exit 0 via `cmdutil.ExitCodeFor`)
- The first error (or `*core.PartialSuccessError{Succeeded: 0, Failed: n}`) if all failed (exit 1)

**Rationale**: matches the foundation's `core/errors.go` design; gives `cmdutil.ExitCodeFor` enough information to map partial success to exit 0; preserves the legacy behavior.

**Alternatives considered**:

- Return a custom `InitResult` struct → defeats the new pattern (commands return `error`, not result objects).
- Accumulate errors in a global → violates the no-mutable-shared-state rule.

### 2. Preset expansion happens in `runInit`, not in the Library

**Choice**: `runInit` calls `lib.ResolvePreset(ctx, presetName)` to get the list of refs, then processes each ref in a loop.

**Rationale**: keeps the Library interface minimal; preset expansion is an `init`-specific concern.

### 3. The `Initializer` interface returns `[]core.InitializeResult` plus `error`

**Choice**: The `Initializer.Initialize` signature is:
```go
Initialize(ctx context.Context, req *InitializeRequest) ([]core.InitializeResult, error)
```

Where `core.InitializeResult` is `{Ref string, InputPath string, OutputPath string, Succeeded bool, Error error}`.

**Rationale**: the slice of results allows the caller to inspect per-resource outcomes; the error is reserved for transport-level failures (e.g. "library not found"). Per-resource failures are encoded in `core.InitializeResult.Error`.

**Alternatives considered**:

- Return only errors via `error` (no slice) → loses per-resource status info; caller can't report partial success.
- Return only `[]InitializeResult` (no error) → can't signal transport-level failures.

### 4. `core.InitializeError` wraps the cause for `errors.As`

**Choice**: `core.InitializeError` (defined in change-1 alongside `core.PartialSuccessError`) has a `Cause error` field and an `Unwrap() error` method that returns the cause. This lets `errors.As(err, &typedErr)` reach the underlying typed error.

**Rationale**: matches the foundation's `core.PartialSuccessError` design; preserves error chain for `cmdutil.ExitCodeFor` and `output.FormatError` dispatch.

## Risks / Trade-offs

- **`Initializer.Initialize` signature change** — the new signature is a breaking change for any caller that expected `[]InitializeResult` and `nil`/`error`. **Mitigation:** the only caller in this codebase is `cmd/init.go` (now migrated); the `library add` command in change-6 will be updated to use the same signature.
- **`core.InitializeResult` is a new type** — existing tests in `internal/service/initializer_test.go` use the legacy shape. **Mitigation:** tests are converted in this change; the type is defined in `internal/core/` to avoid import cycles.
- **Partial success edge case** — exactly one resource failing AND one succeeding should map to exit 0. **Mitigation:** explicit test case in task 5.3.

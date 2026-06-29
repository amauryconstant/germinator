# Design — Migrate init command

## Context

The `init` command is the first germinator command to use the `core.PartialSuccessError` sentinel (defined in the `scaffold-cli-foundation` change and recognized by `cmdutil.ExitCodeFor`). It is also the first command that needs a per-resource error list as part of its result type.

This change depends on three preliminary code extensions (tracked as tasks in `tasks.md` §5.0):

1. **`cmdutil.ExitCodeFor`** extended to map `*core.NotFoundError` → `ExitCodeUsage` (2), so preset-not-found propagates as a usage error.
2. **`(*library.Library).ResolvePreset(ctx, name)`** method introduced; the legacy package function `library.ResolvePreset(lib, preset)` becomes a thin shim.
3. **`cmdutil.Factory.Initializer`** lazy field added so `cmd/init.go` can resolve the initializer via the factory without an import cycle.

## Goals / Non-Goals

**Goals:**

- `cmd/init.go` follows the `NewCmdInit(f, runF) + runInit(opts)` pattern.
- Partial success returns `*core.PartialSuccessError`; `cmdutil.ExitCodeFor` returns 0 for `Succeeded > 0`.
- Preset expansion (e.g. `--preset git-workflow` → list of refs) is preserved; `--preset <nonexistent>` returns exit 2 via `*core.NotFoundError`.
- Dry-run and `--force` flags are preserved.
- Output format is human-readable per-resource status; JSON output via `--output json` is NOT added (init doesn't produce structured output suitable for JSON).
- The `--output`/`-o` flag is renamed to `--output-dir` (breaking change documented in `proposal.md` → Risks).

**Non-Goals:**

- Migrating library commands — separate follow-up changes (`migrate-library-add`, `migrate-library-create`, etc.).
- Adding `--output` flag to init — it produces per-resource status text, not structured data.
- Preset resolution semantics change — uses the (newly methodized) `(*library.Library).ResolvePreset` directly.

## Decisions

### 1. `runInit` returns `*core.PartialSuccessError` on partial or full failure

**Choice:** After processing all refs, `runInit` returns one of:

- `nil` if every result has `Succeeded == true` (exit 0).
- `*core.PartialSuccessError{Succeeded: m, Failed: n}` if some succeeded and some failed (exit 0 via `cmdutil.ExitCodeFor`).
- `*core.PartialSuccessError{Succeeded: 0, Failed: n}` if all failed (exit 1).
- `*core.NotFoundError{Entity: "preset", Name: opts.Preset}` if `(*Library).ResolvePreset` returned an error — exit 2 (mapped by the §5.0.1 extension).

**Rationale:** Matches the foundation's `core/errors.go` design; gives `cmdutil.ExitCodeFor` enough information to map partial success to exit 0 and preset-not-found to exit 2; preserves legacy behavior for the success/partial/all-failed triad.

**Alternatives considered:**

- Return a custom `InitResult` struct → defeats the new pattern (commands return `error`, not result objects).
- Accumulate errors in a global → violates the no-mutable-shared-state rule.

### 2. Preset expansion happens in `runInit` via the `(*Library).ResolvePreset` method

**Choice:** `runInit` calls `f.Library().ResolvePreset(opts.Ctx, opts.Preset)` to get the list of refs, then processes each ref in a loop. Any error from `ResolvePreset` becomes a `*core.NotFoundError` returned from `runInit`.

**Rationale:** Keeps the `*Library` interface minimal and contextual; preset expansion is an `init`-specific concern; method-on-pointer aligns with `interface-where-consumed`. Moving the package function to a method (task §5.0.2) also unblocks future callers (e.g., `library add` in `migrate-library-add`). The legacy package function is retained as a one-line shim that delegates to the method so non-migrated callers keep working without parallel maintenance; the shim is removed once all callers migrate (tracked as out-of-scope in `proposal.md` → "Out of scope").

### 3. The `Initializer` interface returns `[]core.InitializeResult` plus `error`

**Choice:** The `Initializer.Initialize` signature is:

```go
Initialize(ctx context.Context, req *InitializeRequest) ([]core.InitializeResult, error)
```

Where `core.InitializeResult` is `{Ref string, InputPath string, OutputPath string, Succeeded bool, Error error}`.

**Rationale:** The slice of results allows the caller to inspect per-resource outcomes; the `error` return is reserved for transport-level failures (e.g. "library not found"). Per-resource failures are encoded in `core.InitializeResult.Error`.

**Alternatives considered:**

- Return only errors via `error` (no slice) → loses per-resource status info; caller can't report partial success.
- Return only `[]InitializeResult` (no error) → can't signal transport-level failures.

The `Initializer` interface is consumed inline in `cmd/init.go` but is **declared** in `internal/application/` (one of two consumers — `cmd/init.go` and `*cmdutil.Factory.Initializer` — so it must be exported from a shared location, not declared per-consumer). The factory's `Initializer` lazy field has this type.

### 4. `--resources` and `--preset` are mutex, not merge

**Choice:** `runInit` validates that exactly one of `opts.Refs` or `opts.Preset` is set (preserved from base `cli-init-command` spec); the task field is treated strictly exclusive, not merged.

**Rationale:** Preserves existing base-spec behavior; avoids the ambiguity of "which refs come from where" when both are set.

## Foundation Dependencies

- `core.PartialSuccessError` — sentinel for partial success / all-failed aggregation. Maps to exit 0 (`Succeeded > 0`) or exit 1 (`Succeeded == 0`).
- `core.InitializeError` — per-resource failure with `Ref`, `InputPath`, `OutputPath`, `Cause`. `Unwrap()` returns the cause.
- `core.InitializeResult` — per-resource outcome carrying `Succeeded` + `Error`.
- `cmdutil.ExitCodeFor` — maps errors to exit codes. Extended in §5.0.1 to also map `*core.NotFoundError` → 2.
- `output.FormatError` — renders typed errors to `opts.IO.ErrOut`, including the partial-success aggregate.
- `legacyBridge` — runtime shim that lets non-migrated commands (those still using the legacy `cmd.ServiceContainer` DI pattern) coexist with migrated commands using the new `*cmdutil.Factory` pattern. Slice 5 adds the first migrated command (`init`) into a tree that also hosts non-migrated siblings, so verification (§5.5.8 in `tasks.md`) confirms the shim still routes services correctly.

## Risks / Trade-offs

- **`Initializer.Initialize` signature change** — the new signature is already a breaking change relative to anything earlier than `scaffold-cli-foundation`. **Mitigation:** the only caller in this codebase is `cmd/init.go` (now migrated); `library add` (`migrate-library-add`) and `library init` are migrated in follow-up changes.
- **`core.InitializeResult` is a new type** — existing tests in `internal/service/initializer_test.go` use the legacy shape. **Mitigation:** tests are converted in this change; the type is defined in `internal/core/` to avoid import cycles.
- **Partial success edge case** — exactly one resource failing AND one succeeding should map to exit 0. **Mitigation:** explicit test case in task §5.3.2.
- **`--output-dir` is a breaking rename** — legacy `--output`/`-o` is removed. **Mitigation:** documented in `proposal.md` Risks; users must update scripts and shell aliases.
- **Three preliminary code-change tasks** (§5.0.1, 5.0.2, 5.0.3) gate the migration. **Mitigation:** they are listed as discrete prerequisite tasks and are testable in isolation before any `cmd/init.go` rewrite.

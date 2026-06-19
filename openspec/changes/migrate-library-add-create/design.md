# Design — Migrate library add and library create

## Context

`library add` is the most behaviorally complex command in germinator: it supports three input modes (explicit files, `--discover` scan, `--discover --batch --force` continuous) and may produce partial successes. `library create` is simpler: it builds a preset from a list of refs. Both commands touch the library package directly, so the migration also adds `ctx` to the library's I/O functions for cancellation safety.

## Goals / Non-Goals

**Goals:**

- `cmd/library/add.go` and `cmd/library/create.go` follow the `NewCmdXxx(f, runF) + runXxx(opts)` pattern.
- `core.CanInstallResource` provides string-only ref validation (depguard-compatible).
- All three modes of `library add` work end-to-end.
- Batch mode handles context cancellation: partial successes are reported, then the function returns wrapped `ctx.Err()` so the exit code is 1.
- `library.DiscoverOrphans`, `library.BatchAddResources`, `library.LoadLibrary` accept `ctx context.Context` as the first parameter and respect cancellation.

**Non-Goals:**

- Migrating `library init`, `library refresh`, `library remove`, `library validate` — change-7.
- Restructuring `library` package internals (e.g. consolidating loader + saver into a `store` type) — deferred to a follow-up change.
- Adding `--output` flag to `library create` — it didn't previously have `--json`.

## Decisions

### 1. `core.CanInstallResource` is string-only

**Choice**: The function signature is `func CanInstallResource(ref string) error`. It parses `ref` using `strings.Cut(ref, "/")` and validates the `type` and `name` parts using only stdlib functions. It does NOT look up the resource in the library.

**Rationale**: depguard (`internal/core/**` allows only stdlib + `samber/lo`) prevents `core` from importing `internal/library/`. The function exists to do syntactic validation before the actual library lookup, which happens in `runAdd` after this validation passes.

### 2. Context flows through library I/O

**Choice**: Add `ctx context.Context` as the first parameter to `library.DiscoverOrphans`, `library.BatchAddResources`, `library.LoadLibrary`. Each function checks `ctx.Err()` after each I/O operation and returns the partial result + `ctx.Err()` if cancelled.

**Rationale**: cancellation safety is required for the batch mode's continuous behavior. Adding `ctx` to all three functions (not just `BatchAddResources`) is consistent and surfaces any hidden I/O blocking.

### 3. `--discover` returns partial-success results

**Choice**: When `--discover` finds orphans and some fail to add (e.g. invalid ref format), the function returns a `*core.PartialSuccessError` with the per-file errors.

**Rationale**: matches `init`'s partial-success semantics; gives the user per-resource feedback instead of aborting on the first failure.

### 4. Batch mode continues on failure

**Choice**: `--discover --batch --force` continues processing orphans after individual failures, collecting all results into a `*core.PartialSuccessError`. On context cancellation, the function returns wrapped `ctx.Err()` after collecting partial results.

**Rationale**: the "batch" semantic implies "process as many as possible"; stopping at the first failure would defeat the purpose.

### 5. `library create` does not get `--output`

**Choice**: `library create` does not call `cmdutil.AddOutputFlags`. The legacy implementation did not support `--json`.

**Rationale**: matches the `output-formats` capability's "only commands that previously supported `--json` get `--output`" rule.

## Risks / Trade-offs

- **Three modes × partial success × cancellation** — many code paths. **Mitigation:** each combination has explicit test coverage (tasks 6.1.5, 6.1.6).
- **Adding `ctx` to library functions** — breaks any caller not updated. **Mitigation:** `mise run check` catches all missed callers; only `cmd/library_add.go` (now `cmd/library/add.go`) and `cmd/library_init.go` (now `cmd/library/init.go` in change-7) are callers.
- **`core.CanInstallResource` may be redundant with library validation** — the library does its own ref validation. **Mitigation:** the function is a fast-fail check before I/O; the library does authoritative validation.
- **`library create` with empty resources** — what happens if `--resources` is empty? **Mitigation:** task 6.2.2 validates that `opts.Resources` is non-empty and returns a usage error (exit 2) if empty.

# Design — Migrate library add and library create

## Context

`library add` is the most behaviorally complex command in germinator: it supports three input modes (explicit files, `--discover` scan, `--discover --batch --force` continuous) and may produce partial successes. `library create` is simpler: it builds a preset from a list of refs. Both commands touch the library package directly, so the migration also adds `ctx` to the library's I/O functions for cancellation safety.

## Goals / Non-Goals

**Goals:**

- `cmd/library_add.go` and `cmd/library_create.go` follow the `NewCmdXxx(f, runF) + runXxx(opts)` pattern (in place, flat layout).
- `core.CanInstallResource` provides string-only ref validation (depguard-compatible).
- All three modes of `library add` work end-to-end.
- Batch mode handles context cancellation: partial successes are reported, then the function returns wrapped `ctx.Err()` so the exit code is 1.
- `library.DiscoverOrphans`, `library.BatchAddResources`, `library.LoadLibrary` accept `ctx context.Context` as the first parameter and respect cancellation.
- The `library create` Cobra group is collapsed to a leaf (`NewCmdCreatePreset`).
- Plain output for `library add` is byte-identical to the pre-change output (matches the spec delta's contract).

**Non-Goals:**

- Migrating `library init`, `library refresh`, `library remove`, `library validate` — change-7.
- Restructuring `library` package internals (e.g. consolidating loader + saver into a `store` type) — deferred to a follow-up change.
- Adding `--output` flag to `library create` — it didn't previously have `--json`.

## Decisions

### 0. `core.OperationError` is a new foundation unit (parallel to `core.NotFoundError` from slice-4)

**Choice**: Add `type OperationError struct { Op, Resource string; Cause error }` with constructor `core.NewOperationError(op, resource string, cause error) *OperationError` and `func (e *OperationError) Error() string` rendering `<op>: <resource>`. `Unwrap()` returns `Cause`. The `output.FormatError` dispatcher gains an `errors.As` branch that renders `Error: <op>: <resource>\n` to stderr via `Styles.Error`.

**Rationale**: Today, `library.DiscoverOrphans` returns `*library.ConflictInfo{Issue: "name_conflict"}` as a string field on the result; the typed-error migration replaces this with `*core.OperationError` so that:
1. The error is wrapped through `*core.PartialSuccessError` and rendered per-file by `output.FormatError`.
2. Callers can `errors.As` on the typed error to distinguish "already registered" (silent skip), "name conflict" (typed error, counted as failure), and "other error" (typed error, counted as failure).
3. The `Error()` string is owned by the type, not by the renderer (matches `*core.NotFoundError`, `*core.ValidationError`).

**Alternatives considered**: (a) keep `string Issue` and have the renderer pattern-match — rejected; loses type information. (b) reuse `*core.ValidationError` — rejected; `OperationError` is not a validation outcome, it's an operation outcome with a `Cause` chain.

### 1. `core.CanInstallResource` is string-only

**Choice**: The function signature is `func CanInstallResource(ref string) error`. It parses `ref` using `strings.Cut(ref, "/")` and validates the `type` and `name` parts using only stdlib functions. It does NOT look up the resource in the library.

**Rationale**: depguard (`internal/core/**` allows only stdlib + `samber/lo`) prevents `core` from importing `internal/library/`. The function exists to do syntactic validation before the actual library lookup, which happens in `runAdd` after this validation passes.

### 2. Context flows through library I/O

**Choice**: Add `ctx context.Context` as the first parameter to `library.DiscoverOrphans`, `library.BatchAddResources`, `library.LoadLibrary`. Each function checks `ctx.Err()` after each I/O operation and returns the partial result + `ctx.Err()` if cancelled.

**Rationale**: cancellation safety is required for the batch mode's continuous behavior. Adding `ctx` to all three functions (not just `BatchAddResources`) is consistent and surfaces any hidden I/O blocking.

### 3. `--discover` returns partial-success results with conflict distinct from skip

**Choice**: When `--discover` finds orphans:
- **Successfully added** entries populate the partial-success `Succeeded` count and the per-file list as successes.
- **`name_conflict`** entries (orphan name already registered under a different type) produce a `*core.OperationError{Op: "register", Resource: <ref>, Cause: <origErr>}` per file and increment the `Failed` count — they are *not* treated as successes.
- **Other errors** (e.g., file read error, type detection failure) produce a `*core.OperationError{Op: <callerOp>, Resource: <ref>, Cause: <origErr>}` per file and increment the `Failed` count.

**Rationale**: The pre-change code distinguished `name_conflict` from "already registered" and from "error". The migration preserves that distinction at the typed-error layer (the human formatter `output.FormatError` renders the per-file error messages to stderr) and aggregates them in `*core.PartialSuccessError` so the exit code is 0 if any succeed and 1 otherwise. Matches `init`'s partial-success semantics; gives the user per-resource feedback instead of aborting on the first failure.

### 4. Batch mode continues on failure

**Choice**: `--discover --batch --force` continues processing orphans after individual failures, collecting all results into a `*core.PartialSuccessError`. On context cancellation, the function returns wrapped `ctx.Err()` after collecting partial results.

**Rationale**: the "batch" semantic implies "process as many as possible"; stopping at the first failure would defeat the purpose.

### 5. `library create` does not get `--output`

**Choice**: `library create` does not call `cmdutil.AddOutputFlags`. The legacy implementation did not support `--json`.

**Rationale**: matches the `output-formats` capability's "only commands that previously supported `--json` get `--output`" rule.

### 6. Library type renames for interface alignment

**Choice**: Rename `library.AddOptions` → `library.AddRequest`; `library.OrphanInfo` → `library.Orphan`. `library.BatchAddResult`, `library.AddSuccess`, `library.DiscoverResult`, `library.DiscoverSummary`, `library.ConflictInfo` are kept as-is.

**Rationale**: Task 6.4.2's `resourceAdder` interface uses `*AddRequest`, `[]Orphan`, `*BatchAddResult`. Renaming the public types avoids two parallel naming conventions inside the same package and aligns with the `request/result` convention used elsewhere (`internal/application/requests.go`).

**Migration cost**: All intra-package callers in `internal/library/adder.go` and `adder_test.go`; external caller `cmd/library_add.go:runLibraryAdd`. Mechanical rename; `mise run check` will surface any miss.

**Test impact**: `internal/library/adder_test.go` (1 file) is updated mechanically for the rename; no behavior tests change because the renamed types carry the same fields and same semantic.

### 7. Command files stay flat in `cmd/`

**Choice**: The migrated files are `cmd/library_add.go` and `cmd/library_create.go` (top-level). No `cmd/library/` subdirectory is created.

**Rationale**: Every other command in this codebase lives flat in `cmd/`, including the slice-2/4/5 migrations (`adapt`, `init`, `resources`, `presets`, `show`). Introducing the first subdirectory would create an inconsistent pattern for no architectural gain (the `cmd/commands/AGENTS.md` per-command reference already groups by command name in prose).

### 8. `library create` collapses to a leaf

**Choice**: Delete the `NewLibraryCreateCommand` Cobra group. Register `NewCmdCreatePreset` directly under the `library` command. The user-facing command path `germinator library create preset <name> --resources ...` is unchanged.

**Rationale**: `preset` is the only subcommand under `library create`; the group adds an indirection that produces an extra help-screen and matches no other command in the CLI.

**Spec impact**: this is a non-user-visible restructuring — the user-facing command path is preserved. The transition is speced in the `library-library-json-output` delta as an `ADDED Requirement` ("library create preset is a leaf under library") and verified by `cmd/library.go` rewiring.

### 9. `--output plain` is byte-identical to the pre-change output

**Choice**: When `library add` is invoked without `--output` or with `--output plain`, the human-readable output is byte-identical to the pre-change `library_add.go` output (per the spec delta's "byte-identical to pre-change" scenario).

**Rationale**: Preserves the contract from the existing main spec (`library-library-json-output`). The new `--output json` and `--output table` formats are net-new outputs that do not affect the default path.

**Coverage requirement**: Golden-file test `cmd/library_add_test.go` compares the plain output against the recorded pre-change baseline (or against a regenerated baseline, depending on what the implementation phase records).

## Risks / Trade-offs

- **Three modes × partial success × cancellation** — many code paths. **Mitigation:** each combination has explicit test coverage (tasks 6.4.5, 6.4.6, 6.4.7).
- **Adding `ctx` to library functions** — breaks any caller not updated. **Mitigation:** `mise run check` catches all missed callers; `cmd/library_add.go` (in place, flat) is updated; `cmd/library_init.go`, `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_validate.go` (all deferred to change-7) get `context.Background()` threaded through today as a mechanical no-op (no behavior change for legacy commands; ctx will be wired properly when those commands migrate in slice-7).
- **`core.CanInstallResource` may be redundant with library validation** — the library does its own ref validation. **Mitigation:** the function is a fast-fail check before I/O; the library does authoritative validation.
- **`library create` with empty resources** — what happens if `--resources` is empty? **Mitigation:** task 6.5.4 validates that `opts.Resources` is non-empty and returns a usage error (exit 2) if empty.
- **Type renames (`AddOptions` → `AddRequest`, `OrphanInfo` → `Orphan`)** — touches intra-package and external callers. **Mitigation:** `mise run check` will surface any miss; the rename is mechanical (`goimports` / `gofmt` then compile).
- **Plain-output byte-identical guarantee** — coupling between the new code and the pre-change human format. **Mitigation:** golden-file test in `cmd/library_add_test.go` pins the format; the `--output json|table` formats are decoupled. A second golden-file test pins the partial-success per-resource status format (see task 6.4.7) so the new failure path is also pinned, not just the success path.
- **`library create` group collapse** — removes one indirection in the help tree. **Mitigation:** the user-facing command path (`germinator library create preset ...`) is unchanged; `cmd/library.go` registration is updated.

# Migrate library add and library create

## Why

The mutating library commands (`library add` and `library create`) have more complex behavior than the read-only commands: `library add` supports three modes (explicit files, `--discover` scan, `--discover --batch --force` continuous), and `library create` builds a preset from a list of refs. Migrating them after `init` (change-5) lets us reuse the `core.PartialSuccessError` pattern and adds `core.CanInstallResource` (a pure rule function in `internal/core/rules.go`).

## What Changes

### Add foundation unit `core.OperationError`

- **ADD** `*core.OperationError{Op, Resource string; Cause error}` to `internal/core/errors.go` with constructor `core.NewOperationError(op, resource string, cause error) *OperationError` and `func (e *OperationError) Error() string` rendering `<op>: <resource>`. The `Cause` is preserved via `Unwrap()`. This formalizes the per-file error reported during `--discover` so it can be typed (and not just a `string`), and so `output.FormatError` can render it through the existing typed-error dispatcher chain.
- **ADD** a dispatch branch in `output.FormatError` (`internal/output/errors.go`) that renders `Error: <op>: <resource>\n` to stderr via the existing `Styles.Error` channel. Uses `errors.As(err, &opErr)` for dispatch.
- **UPDATE** `internal/AGENTS.md` line 31 ("typed domain errors" bullet) to confirm `OperationError` exists.
- **ADD** unit tests in `internal/core/errors_test.go` and `internal/output/output_test.go` covering constructor, `Error()` string, `errors.As` dispatch, `Unwrap()` chain, and stderr rendering.

### Add `core.CanInstallResource` to `internal/core/rules.go`

- **ADD** `core.CanInstallResource(ref string) error` to `internal/core/rules.go`:
  - Parses `ref` using `strings.Cut(ref, "/")` (Go 1.18+)
  - Validates that `type` is one of `skill`, `agent`, `command`, `memory` (using `slices.Contains`)
  - Validates that `name` is a non-empty identifier
  - Returns `*core.ValidationError` on failure
  - **String-only** (does NOT import `internal/library/` — depguard enforces this)
  - **Rationale for not speccing as a new capability**: `CanInstallResource` is a private helper called only by `runAdd` (Mode 1) and `runCreatePreset`. Its error contract is covered by the existing `cli-error-formatting` capability (via `*core.ValidationError` rendering). Its call sites are speced in the modified `library-library-resource-import` and `library-library-preset-creation` deltas (below). No new top-level spec is needed.

### Migrate `cmd/library_add.go` (in place; flat layout, matching every other command)

- **MIGRATE** `cmd/library_add.go`:
  - Declare `addOptions`: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `InputPaths []string`, `Name string`, `Description string`, `Type string`, `Platform string`, `Discover bool`, `Batch bool`, `Force bool`, `DryRun bool`, `Output string`
  - Declare the **`resourceAdder`** interface (NOT named `Library`, to avoid shadowing the `library.Library` struct returned by `f.Library()`) with the methods called by all three modes
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` (legacy `--json` is replaced by `--output json`)
  - Implement `NewCmdAdd(f *cmdutil.Factory, libraryPath *string, runF func(*addOptions) error) *cobra.Command` and `runAdd(opts *addOptions) error`:
    - Mode 1 (explicit files): validate each `InputPath` with `core.CanInstallResource(name)`, add to library
    - Mode 2 (`--discover`): scan directories for orphan files; for each, validate ref; collect successes/failures
    - Mode 3 (`--discover --batch --force`): continuous processing; on per-file failure, skip and continue
  - On `name_conflict` (orphan name already registered under a different type): record a `*core.OperationError{Op: "register", Resource: <ref>, Cause: <origErr>}` for the file in the per-file list and increment `PartialSuccessError.Failed`; do not stop processing other orphans. Preserves the pre-change distinction between "already registered" (silently skipped), "name conflict" (warned), and "error" (failed).
  - On partial success (some added, some failed): return `*core.PartialSuccessError` (exit 0 via `cmdutil.ExitCodeFor`)
  - On context cancellation during batch processing: return wrapped `ctx.Err()` after collecting partial results
  - **Thread `opts.Ctx` into every call** to `library.DiscoverOrphans`, `library.BatchAddResources`, `library.LoadLibrary`

### Migrate `cmd/library_create.go` (in place; flat layout; `library create` Cobra group collapses to a leaf)

- **MIGRATE** `cmd/library_create.go`:
  - Declare `createPresetOptions`: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Resources []string`, `Description string`, `Force bool`
  - Declare the **`presetWriter`** interface (NOT named `Library`) with the methods called by `runCreatePreset`
  - Implement `NewCmdCreatePreset(f *cmdutil.Factory, libraryPath *string, runF func(*createPresetOptions) error) *cobra.Command` and `runCreatePreset(opts *createPresetOptions) error`
  - Validate each ref in `opts.Resources` via `core.CanInstallResource(ref)` (fast-fail before I/O)
  - **No `--output` flag** (legacy implementation didn't support `--json`; matches the `output-formats` capability's "only commands that previously had `--json` get `--output`")

### Collapse the `library create` Cobra group

- **DELETE** the `NewLibraryCreateCommand` Cobra group wrapper from `cmd/library_create.go` (it has only one subcommand, `preset`; the group adds an indirection that matches no other command in the CLI).
- **UPDATE** `cmd/library.go` to register `NewCmdCreatePreset(f, &libraryPath, nil)` directly under `library` (replacing `cmd.AddCommand(NewLibraryCreateCommand(bridge, libraryPath))`).
- The user-facing command is unchanged: `germinator library create preset <name> --resources ...`.

### Delete legacy files

- **DELETE** `internal/service/adder.go` if present
- **DELETE** `internal/service/creator.go` if present

## Capabilities

### New Capabilities

- **`errors-operation-error`** — `*core.OperationError{Op, Resource string; Cause error}` formalizes per-operation errors so that `output.FormatError` can render them through the typed-error dispatcher (rendering `Error: <op>: <resource>` to stderr) and so that callers (notably `runAdd` Mode 2's `--discover` aggregation) can carry them in typed error chains rather than ad-hoc strings. Foundation unit; parallels `core.NotFoundError` introduced in slice-4.

### Modified Capabilities

- **`library-library-json-output`** — `--output` flag is now available on `library add` (the legacy `--json` flag is replaced). `library create` does NOT get `--output` (it didn't previously have `--json`). Additionally, `library create preset` is registered directly under `library` as a leaf — the `library create` Cobra group wrapper is removed.
- **`library-library-orphan-discovery`** — `name_conflict` outcomes are now recorded as `*core.OperationError{Op: "register", Resource: <ref>}` and aggregated into `*core.PartialSuccessError.Failed` instead of being carried as a string `Issue` field. This preserves the pre-change distinction between "already registered" (silently skipped), "name conflict" (typed error, counted as failure), and "other error" (typed error, counted as failure) while moving to typed errors throughout. Also introduces the `library.ErrNameConflict` sentinel (task 6.2.8) so the cause chain is traversable via `errors.Is`.
- **`library-library-resource-import`** — adds a "Pre-flight ref validation" scenario covering `core.CanInstallResource(name)` rejecting malformed refs with `*core.ValidationError` before any I/O is performed.
- **`library-library-preset-creation`** — adds a "Pre-flight ref validation" scenario covering `core.CanInstallResource(ref)` rejecting malformed refs in `opts.Resources` with `*core.ValidationError` before `CreatePreset` is called.

## Out of scope (deferred)

- Migrating remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) — change-7
- Migrating config / completion / version — changes 8, 9

## Impact

### Affected code

- **Modified (1 file):** `internal/core/errors.go` (add `OperationError` type + `NewOperationError` constructor)
- **Modified (1 file):** `internal/core/errors_test.go` (add test cases for `OperationError`)
- **Modified (1 file):** `internal/output/errors.go` (add `errors.As` dispatch branch for `*core.OperationError`)
- **Modified (1 file):** `internal/output/output_test.go` (add stderr-rendering test for `OperationError`)
- **Modified (1 file):** `internal/AGENTS.md` (line 31: confirm `OperationError` in typed-error list)
- **Modified (1 file):** `internal/core/rules.go` (add `CanInstallResource`)
- **Modified (1 file):** `internal/core/rules_test.go` (add test cases for `CanInstallResource`)
- **Rewritten (1 file):** `cmd/library_add.go` (in place; flat layout)
- **Rewritten (1 file):** `cmd/library_create.go` (in place; group wrapper deleted, leaf body only)
- **Modified (1 file):** `cmd/library.go` (rewire registration; `NewLibraryAddCommand` → `NewCmdAdd`, `NewLibraryCreateCommand` → `NewCmdCreatePreset`)
- **Modified (1 file):** `internal/library/adder.go` (adds `ctx context.Context` as the first parameter to `AddResource`, `BatchAddResources`, and `DiscoverOrphans`; each function checks `ctx.Err()` after I/O and returns wrapped `ctx.Err()` on cancellation; renames `AddOptions` → `AddRequest` and `OrphanInfo` → `Orphan` to align with the request/result convention)
- **Modified (1 file):** `internal/library/loader.go` (adds `ctx context.Context` as the first parameter to `LoadLibrary`)
- **Modified (1 file):** `internal/library/adder_test.go` (update for the type renames and `ctx` parameter)
- **Deleted (0-2 files):** `internal/service/adder.go`, `internal/service/creator.go` if present (currently absent — see task 7.1)
- **Modified (1 file):** `cmd/library_add_test.go` (converted to `iostreams.Test()` + `runF` injection)
- **Modified (1 file):** `cmd/library_create_test.go` (converted similarly)

### Affected systems

- **CLI behavior:** `--output` flag added to `library add`; `library add` and `library create` output format may shift slightly (per-resource status lines instead of flat list, when partial success)

## Risks

- **Three modes in `library add`** — explicit, discover, batch — increase the migration surface. **Mitigation:** each mode is tested independently in tasks 6.4.5, 6.4.6, 6.4.7; batch mode's continuous behavior is the most complex and gets explicit attention.
- **Context cancellation handling in batch mode** — when the user hits Ctrl-C during a long batch, partial successes must be reported. **Mitigation:** task 6.4.4 explicitly handles this; `*core.PartialSuccessError` collects partial successes; `cmdutil.ExitCodeFor` returns 0 if any succeeded.
- **`CanInstallResource` must be string-only** — it can't import `internal/library/` (depguard). **Mitigation:** the function parses and validates the ref syntactically; the actual library lookup happens in `runAdd` after validation passes.
- **`core.OperationError` is a new foundational type** — added in this slice (parallel to `core.NotFoundError` in slice-4) and consumed by `runAdd` Mode 2 and the per-file error rendering. **Mitigation:** task group 6.0 introduces the type, constructor, dispatcher branch, and unit tests before any consumer task begins (ordering note in tasks.md header).
- **`internal/library/` API changes** — adding `ctx` to `DiscoverOrphans`, `BatchAddResources`, `LoadLibrary` is a breaking change to a package that has other callers (notably `cmd/library_init.go`, `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_validate.go`, which are migrated in slice-7 but still call `LoadLibrary` today). **Mitigation:** task 6.2.6 threads `context.Background()` through the legacy call sites in slice-6 (mechanical change; no behavior change for legacy commands). `mise run check` catches any missed caller.

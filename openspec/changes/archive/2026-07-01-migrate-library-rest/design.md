# Design — Migrate remaining library commands and delete legacy shell

## Context

After change-6 (`migrate-library-add-create`), the only remaining consumers of `internal/service/` and `internal/application/` are the four lifecycle library commands (`init`, `refresh`, `remove`, `validate`) plus the `legacyBridge` shim in `main.go`. This change migrates the last four commands and deletes the entire legacy shell in one go.

## Goals / Non-Goals

**Goals:**

- All four remaining library commands are migrated to `NewCmdXxx(f, runF) + runXxx(opts)`.
- All four commands gain `--output json|table|plain` via `cmdutil.AddOutputFlags`.
- `(*library.Library)` gains new methods `Refresh`, `RemoveResource`, `RemovePreset`, `Validate`, `Fix` (the slice-6 forward path that slice-7 was meant to deliver); `cmdutil.Factory.Library` stays as `func() (*library.Library, error)`.
- `internal/service/`, `internal/application/`, `cmd/legacy_bridge.go`, `cmd/error_formatter.go`, `cmd/verbose.go`, and `legacyBridge` (in `main.go`) are deleted.
- `mise run check` is green; `mise run build` succeeds.

**Non-Goals:**

- Migrating `config init`, `config validate` — change-8.
- Migrating `completion`, `version` — change-9.
- Restructuring the `library` package internals — deferred to a follow-up refactor change.

## Decisions

### 1. Each library command declares its minimal `Library` interface

**Choice**: Each of the four commands declares a `Library` interface with only the methods it calls (e.g. `library init` declares `Init(ctx, *InitRequest) error`; `library refresh` declares `Refresh(ctx, *RefreshRequest) (*RefreshResult, error)`; etc.).

**Rationale**: matches the `application/command-options-pattern` capability; lets each command depend only on what it needs.

### 2. `library remove` is a single command with sub-command dispatch

**Choice**: `library remove` is one Cobra command (`cmd/library_remove.go`) with sub-command dispatch between `resource` and `preset`. Both sub-commands share the same `removeOptions` struct but populate different fields — `resource` populates `Ref string` (from the positional `args[0]`, e.g. `"skill/commit"`, parsed via `library.ParseRef`); `preset` populates `PresetName string` (from the positional `args[0]`). `(*Library).RemoveResource` and `(*Library).RemovePreset` are both methods on `*library.Library` (per Decision 6); `runRemove` dispatches on `PresetName != ""`.

**Rationale**: matches the legacy command shape; the sub-command names are part of the public CLI surface.

### 3. `library validate --fix` is preserved

**Choice**: The `--fix` flag on `library validate` is preserved. It triggers auto-cleanup of the `library.yaml` (e.g. removing ghost preset refs).

**Rationale**: `library validate --fix` is a maintenance feature; removing it would break CI scripts that rely on it.

### 3a. `library validate --fix` with `--output json` emits a fix report

**Choice**: When `--fix` is combined with `--output json`, the JSON payload includes a `fix` field with `RemovedEntries []string` and `StrippedRefs []string` enumerating the changes applied. With `--output plain`, the existing "Fix applied to library.yaml" line is shown. With `--output table`, a two-column table (action, ref) is rendered.

**Rationale**: machine-readable fix output is needed for CI scripts that auto-fix and report.

### 4. Mocks deleted in this change

**Choice**: All `internal/service/*_mock_test.go` files are deleted in this change (not in earlier changes). Affected tests are converted to use `iostreams.Test()` + `runF` injection.

**Rationale**: until this change, the mocks were still needed by `cmd/cmd_test.go` sections for non-migrated commands. After this change, no command uses the mocks.

### 5. Deletion order is bottom-up

**Choice**: The deletion sequence is:
1. Migrate the four commands first (tasks 7.1-7.4)
2. Delete `internal/service/` and `internal/application/` (tasks 7.5.1, 7.5.2)
3. Delete `cmd/legacy_bridge.go` and remove the `bridge` arg from `NewRootCommand` + 6+ other constructors in `main.go` (tasks 7.5.4, 7.5.5, 7.5.6)
4. Delete `cmd/error_formatter.go` and `cmd/verbose.go` (tasks 7.5.7, 7.5.8)
5. Delete `cmd/legacy_test_helpers_test.go` (task 7.5.9)
6. Update `cmd/cmd_test.go` to drop `newTestBridge()` calls (task 7.5.10)
7. Verify no remaining legacy symbols (task 7.5.11)
8. Delete `cmd/library_formatters.go` after the formatter move (task 7.5.12)

**Rationale**: each step removes a consumer from the previous step; `mise run check` after each step catches any missed dependency.

### 6. Methods on `*library.Library` instead of a `*LibraryService` wrapper

**Choice**: Add four mutating methods directly on `*library.Library` (parallel to the slice-6 `(*Library).CreatePreset` precedent at `internal/library/creator.go:145`):

- `(lib *Library) Refresh(ctx, *RefreshRequest) (*RefreshResult, error)` — wraps the existing `library.RefreshLibrary` package function; uses `lib.RootPath` instead of taking a path arg; existing package-level `library.RefreshLibrary(opts RefreshOptions)` is refactored to delegate to the method form via a thin loader wrapper (preserves any external callers)
- `(lib *Library) RemoveResource(ctx, *RemoveResourceRequest) error` — wraps `library.RemoveResource`; uses `lib.RootPath`
- `(lib *Library) RemovePreset(ctx, *RemovePresetRequest) error` — wraps `library.RemovePreset`; uses `lib.RootPath`
- `(lib *Library) Validate(ctx, *ValidateRequest) (*ValidateResult, error)` — wraps `library.ValidateLibrary`
- `(lib *Library) Fix(ctx, *FixRequest) (*FixResult, error)` — wraps `library.FixLibrary`; required because validation with `req.Fix` needs to mutate and the spec calls for a separate `RemovedEntries`/`StrippedRefs` report

**For `library init`:** `init` creates a fresh library, so there is no pre-existing `*Library` to receive a method. `library.CreateLibrary(CreateOptions)` stays as a package-level function; a thin new package function `library.Init(ctx, *InitRequest) error` maps request fields to `CreateOptions` (parallels the `CreatePreset`/`lib.CreatePreset` dual-form at `internal/library/creator.go:127`). `runLibraryInit` calls `library.CreateLibrary` directly without an interface or adapter shim — the cmd layer's pattern still applies (single-file parse/execute/respond; runF injection; `iostreams.Test()`), but the I/O is one thin package call, not a method on `*Library`.

The `*Request` and `*Result` types are new, defined in `internal/library/requests.go` (a dedicated file to keep `library.go` focused on the data model):
- `library.InitRequest{Path, Force, DryRun string/bool}` → wraps `CreateOptions`
- `library.RefreshRequest{DryRun, Force bool}`
- `library.RemoveResourceRequest{Ref, Force string/bool}` (Ref is `"type/name"`)
- `library.RemovePresetRequest{Name, Force string/bool}`
- `library.ValidateRequest{Fix bool}` — `req.Fix=true` triggers `(*Library).Fix`
- `library.FixRequest{}` (no fields; uses `lib.RootPath`)
- `library.FixResult{RemovedEntries []string; StrippedRefs []string}` — captures the spec's `removedEntries`/`strippedRefs` JSON payload field shape

`cmdutil.Factory.Library` stays as `func() (*library.Library, error)` — **no signature change**. The four cmd-side interfaces (`refresherLibrary`, `removerLibrary`, `validatorLibrary`) are satisfied directly by `*library.Library`, eliminating the `*LibraryService` indirection.

**Rationale**: slice-6's canonical example in `cmd/canonical-examples/AGENTS.md:160` explicitly noted that "a future slice that converts the package functions to methods on `*library.Library` will allow the compile-time check against the concrete type instead of the adapter." This slice delivers that path. Adding a separate `*LibraryService` type would have reintroduced the same "parallel type" anti-pattern that slice-7 is otherwise deleting (mirroring the dying `application.Transformer`/`Validator`/`Canonicalizer`/`Initializer` interfaces). The compile-time check is preserved (`var _ refresherLibrary = (*library.Library)(nil)`). The `*Request` types match the slice-6 precedent of `library.CreatePresetRequest` (`internal/library/creator.go:99`).

### 7. `library refresh` adds an `Unchanged` field to `RefreshResult`

**Choice**: Extend `library.RefreshResult` with a new `Unchanged []RefreshUnchanged` field. `RefreshUnchanged` is `{Ref string, LastSynced string}`. The plain output renders an `Unchanged:` section listing resources that were inspected but matched `library.yaml` exactly.

**Rationale**: makes the success path explicit (users currently see only what changed; "no news" is ambiguous); requested in the delta spec; small additive change to `internal/library/refresher.go`.

### 8. `cmd/library_formatters.go` helpers move to `internal/output/`

**Choice**: After the four commands are migrated, the formatter helpers in `cmd/library_formatters.go` move to `internal/output/library.go` (new file). The four migrated command files import them from `internal/output` instead of the local `cmd` package.

**Rationale**: shared output belongs in `internal/output/` (per `internal/AGENTS.md` package rules).

## Risks / Trade-offs

- **Mass deletion** — 2 directories + 3 files in one change. **Mitigation:** `rg` checks at each step catch any missed reference; the deletion order ensures no transient broken state.
- **`library validate --fix` is auto-mutating** — it modifies `library.yaml` in place. **Mitigation:** the `--fix` flag is opt-in; without it, `library validate` is read-only.
- **`library remove` without `--force` on existing resources** — should it refuse? **Mitigation:** the existing behavior is preserved: without `--force`, `library remove` prompts (or refuses in non-TTY); with `--force`, it removes.
- **Methods-on-`*Library` make `cmdutil.Factory.Library`'s per-call load unavoidable** — every command that calls a `(*Library) X` method must first call `opts.Library()` to obtain the loaded library. If loading fails, the user sees `library not found` rather than the more specific "could not refresh / remove / validate". **Mitigation:** the loading failure is rendered via `output.FormatError(io, *core.NotFoundError)` so it carries `Error:` prefix and maps to exit 2 via `cmdutil.ExitCodeFor`.

# Migrate init command with partial-success semantics

## Why

The `init` command has unique semantics among the germinator commands: it processes N resources (a preset expansion or explicit refs), and partial success is "success" (exit 0) if at least one resource succeeded. The legacy `init` returns `[]InitializeResult` and exits 0 if any succeeded. The new pattern uses `error` returns, so partial success needs a sentinel error type (`core.PartialSuccessError`, defined in the `scaffold-cli-foundation` change) recognized by `cmdutil.ExitCodeFor` (returns 0 for `Succeeded > 0`).

This change also adds the first user-facing usage-error path: `--preset <nonexistent>` is now reported as `*core.NotFoundError` → exit 2, which requires `cmdutil.ExitCodeFor` to map that type (preliminary task §5.0.1).

## What Changes

### Preliminary code changes (gating tasks in `tasks.md` §5.0)

- **ADD** mapping in `cmdutil.ExitCodeFor` so `*core.NotFoundError` returns `ExitCodeUsage` (2) — task §5.0.1.
- **ADD** `(*library.Library).ResolvePreset(ctx, presetName) ([]string, error)` method in `internal/library/`; keep the legacy package function `library.ResolvePreset(lib, preset)` as a thin shim — task §5.0.2.
- **ADD** `Initializer func() (application.Initializer, error)` lazy field to `*cmdutil.Factory` with a corresponding wiring point — task §5.0.3.

### Migrate `cmd/init.go`

- **REWRITE** `cmd/init.go`:
  - Declare `initOptions`: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Initializer func() (application.Initializer, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`.
  - Declare the `application.Initializer` interface reference in `cmd/init.go` (consumed where used; one method: `Initialize(ctx, *InitializeRequest) ([]core.InitializeResult, error)`).
  - Implement `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`.
  - Implement `runInit(opts *initOptions) error`:
    - Validate that exactly one of `opts.Preset` and `opts.Refs` is set; reject both, error if neither (preserved from base spec).
    - Validate platform via `core.ValidatePlatform(opts.Platform)`.
    - If `opts.Preset != ""`: call `f.Library().ResolvePreset(opts.Ctx, opts.Preset)` to expand to refs; wrap any error as `*core.NotFoundError{Entity: "preset", Name: opts.Preset}`.
    - Construct `&application.InitializeRequest{Refs: refs, Platform: opts.Platform, OutputDir: opts.OutputDir, DryRun: opts.DryRun, Force: opts.Force}`.
    - Call `f.Initializer().Initialize(opts.Ctx, req)` to get `([]core.InitializeResult, error)`.
    - If the transport error is non-nil: wrap and return it.
    - Count successes and failures from the result slice:
      - All succeeded → return `nil` (exit 0).
      - Some succeeded, some failed → return `core.NewPartialSuccessError(succeeded, failed, errs)` (exit 0).
      - All failed → return `core.NewPartialSuccessError(0, failed, errs)` (exit 1).
    - Print per-resource status to `opts.IO.Out` (successes) or `opts.IO.ErrOut` via `output.FormatError` (failures).

### Update tests

- **CONVERT** `cmd/init_test.go` to `iostreams.Test()` + `runF` injection (single commit strategy).
- **ADD** explicit test cases for:
  - All-success: `germinator init --platform opencode --resources skill/commit` → exit 0.
  - Partial-success: `--resources skill/commit,skill/invalid` → exit 0 with formatted "partial success: 1 succeeded, 1 failed".
  - All-failed: `--resources skill/invalid1,skill/invalid2` → exit 1 with "partial success: 0 succeeded, 2 failed".
  - Preset expansion: `--preset git-workflow` expands via `(*Library).ResolvePreset`; partial-success logic applies to the expanded list.
  - Preset-not-found: `--preset ghost` → exit 2 via `*core.NotFoundError` → `cmdutil.ExitCodeFor` returns `ExitCodeUsage` (2).

## Capabilities

### Modified

- **`cli-init-command`** — The `init` command adopts the `command-options-pattern` shape (`NewCmdInit(f, runF)` + `initOptions` + `runInit`). Per-resource errors are rendered via `output.FormatError`. Partial-success is signalled by returning `*core.PartialSuccessError`. The output-directory flag is renamed from `--output`/`-o` to `--output-dir` (breaking change). `--resources` and `--preset` remain mutually exclusive, not merged. Preset-not-found surfaces as `*core.NotFoundError` → exit 2.
- **`library-partial-initialization`** — The `application.Initializer.Initialize` contract changes from "nil error on partial success, non-nil on full failure" to "always return the full `[]core.InitializeResult` slice; `error` reserved for transport-level failures". The caller (`runInit`) distinguishes partial vs full failure by inspecting `Succeeded == 0` and synthesizes `*core.PartialSuccessError`.

## Out of scope (deferred)

- Migrating `library add`, `library create` — `migrate-library-add`.
- Migrating remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) — separate change.
- Migrating `config`, `completion`, `version` — separate changes.
- Removing the legacy `library.ResolvePreset` package function shim — deferred until all callers are method-based.

## Impact

### Affected code

- **Rewritten (2 files):**
  - `cmd/init.go`
  - `cmd/init_test.go`
- **Extended (3 files via §5.0 preliminary tasks):**
  - `internal/cmdutil/exit.go` — `*core.NotFoundError` → exit 2 mapping (§5.0.1)
  - `internal/library/resolve.go` (new or extended) — `(*Library).ResolvePreset` method (§5.0.2)
  - `internal/cmdutil/factory.go` — `Initializer` lazy field (§5.0.3)
- **Documentation:**
  - `cmd/AGENTS.md` — update `init.go` entry: remove "Non-migrated command" comment; add `init.go` to the "Canonical examples" list.
- **No new dependencies.**

### Affected systems

- **CLI behavior:** `init` output format changes slightly (per-resource status lines); flag rename `--output`/`-o` → `--output-dir` (breaking); exit code semantics unchanged for the success triad; preset-not-found now maps to exit 2 (was `cmd.HandleCLIError` exit 1).

## Risks

- **Partial-success exit code preservation is critical** — consumers rely on exit 0 when at least one resource installs. **Mitigation:** task §5.3.2 explicitly tests this case; the canary mechanism from `scaffold-cli-foundation` doesn't apply here.
- **Preset-not-found as exit 2 (new behavior)** — callers that previously got exit 1 for an unknown preset will now get exit 2. **Mitigation:** documented in CHANGELOG; the mapping is opt-in via §5.0.1.
- **`--output-dir` is a breaking rename** — existing user scripts and shell aliases that pass `--output`/`-o` will fail. **Mitigation:** documented in CHANGELOG and in this change's Risks section; users must update to `--output-dir`.
- **Preset expansion via method, not package function** — internal callers still using `library.ResolvePreset(lib, preset)` will keep working via the shim during the migration window. **Mitigation:** the shim delegates to the method; legacy function is removed in a later change.
- **Three preliminary code-change tasks** (§5.0.1, §5.0.2, §5.0.3) must complete before `cmd/init.go` is rewritten. **Mitigation:** they are independently testable and tracked as discrete prerequisites.
- **`internal/service/initializer.go`** is not modified by this change — it already returns `[]core.InitializeResult` from the foundation work and is used by `library add`/`library init` migrated in later changes. It implements the `application.Initializer` interface (declared in `internal/application/`, consumed in `cmd/init.go`). **Mitigation:** the file stays as-is; its eventual deletion is owned by the library-migration follow-ups.

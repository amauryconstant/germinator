# Tasks — Migrate init command

**Slice 5 of 9.** Migrates `cmd/init.go` to the new pattern. Wires `core.PartialSuccessError` into the error pipeline (already done in `scaffold-cli-foundation`'s `cmdutil.ExitCodeFor` and `output.FormatError`; this change adds the integration tests). Preserves exit-code semantics (0 if any succeeded, 1 if all failed; new: 2 for preset-not-found).

Three preliminary code-change tasks (§5.0) gate the migration and must complete before §5.1.

Each task ends with `mise run check` passing.

## 5.0 Prerequisites (preliminary code changes)

These tasks edit code outside `cmd/`. They complete before any task in §5.1+.

- [x] 5.0.1 In `internal/cmdutil/exit.go`, add an `errors.As` branch so `*core.NotFoundError` returns `ExitCodeUsage` (2). Acceptance: `cmdutil.ExitCodeFor(core.NewNotFoundError("preset", "x"))` returns `ExitCodeUsage` (2); existing `NotFoundError` tests still pass; `output.FormatError` continues to render `NotFoundError` cleanly.
- [x] 5.0.2 Introduce `(*library.Library).ResolvePreset(ctx context.Context, presetName string) (refs []string, err error)` in `internal/library/` (new or extended file). Keep the legacy package function `library.ResolvePreset(lib, preset)` as a thin shim that delegates to the method. Add unit tests for resolution success, preset-not-found, malformed preset.
- [x] 5.0.3 Add `Initializer func() (application.Initializer, error)` field to `*cmdutil.Factory` in `internal/cmdutil/factory.go`. Lazy-init pattern via `sync.OnceValues` matching other Factory fields. Wire construction at the Factory's composition point. Acceptance: existing Factory tests still pass; `f.Initializer()` returns a working `application.Initializer` returning `[]core.InitializeResult`.

## 5.1 Verify foundation wiring intact

- [x] 5.1.1 Run `mise run test` to confirm `cmdutil.ExitCodeFor` still maps `*core.PartialSuccessError{Succeeded > 0}` → 0 and `{Succeeded == 0}` → 1; confirm `output.FormatError` still renders partial-success error lines. (Foundation work in `scaffold-cli-foundation`; this is a regression gate.)

## 5.2 Migrate `cmd/init.go`

- [x] 5.2.1 In `cmd/init.go`, define `initOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Initializer func() (application.Initializer, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`.
- [x] 5.2.2 In `cmd/init.go`, reference `application.Initializer` (consumed where used); the method signature is `Initialize(ctx, *InitializeRequest) ([]core.InitializeResult, error)`; import `InitializeRequest` from `internal/application/requests.go`.
- [x] 5.2.3 Implement `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`:
   - Add flags: `--platform`, `--output-dir` (replacing legacy `--output`/`-o`), `--library`, `--resources` ([]string), `--preset`, `--dry-run`, `--force`.
   - In `RunE`: construct `opts`, populate from `f.IOStreams`, `c.Context()`, `f.Library`, `f.Initializer` (§5.0.3), and parsed flags.
   - Call `runF(opts)` if non-nil, else `runInit(opts)`.
- [x] 5.2.4 Implement `runInit(opts *initOptions) error`:
   - Validate that exactly one of `opts.Refs` and `opts.Preset` is set (mutex per base spec); reject if both, error if neither.
   - Validate platform via `core.ValidatePlatform(opts.Platform)`.
   - If `opts.Preset != ""`: call `f.Library().ResolvePreset(opts.Ctx, opts.Preset)` (§5.0.2) to expand to refs; if it errors, wrap as `*core.NotFoundError{Entity: "preset", Name: opts.Preset}` and return it.
   - Construct `&application.InitializeRequest{Refs: refs, Platform: opts.Platform, OutputDir: opts.OutputDir, DryRun: opts.DryRun, Force: opts.Force}`.
   - Call `f.Initializer().Initialize(opts.Ctx, req)` to get `([]core.InitializeResult, error)`.
   - If the transport error is non-nil: wrap and return it.
   - Count successes and failures from the result slice.
   - If `Succeeded > 0 && Failed == 0`: return `nil` (exit 0).
   - If `Succeeded > 0 && Failed > 0`: build `[]core.InitializeError` from the failures and return `core.NewPartialSuccessError(succeeded, failed, errs)` (exit 0).
   - If `Succeeded == 0 && Failed > 0`: return `core.NewPartialSuccessError(0, failed, errs)` (exit 1).
   - Print per-resource status to `opts.IO.Out` (successes) or `opts.IO.ErrOut` via `output.FormatError` (failures).
- [x] 5.2.5 Convert `cmd/init_test.go` to `iostreams.Test()` + `runF` injection (single commit strategy, no parallel "add new tests" path).

## 5.3 Add partial-success and edge-case test cases

- [x] 5.3.1 All-success → `runInit(opts)` returns `nil`; `cmdutil.ExitCodeFor(nil) == 0`.
- [x] 5.3.2 Partial success (1 succeeded, 1 failed) → `runInit(opts)` returns `*core.PartialSuccessError{Succeeded: 1, Failed: 1}`; `cmdutil.ExitCodeFor(err) == 0`; `output.FormatError(io, err)` writes `partial success: 1 succeeded, 1 failed`.
- [x] 5.3.3 All-failed (0 succeeded, 2 failed) → `runInit(opts)` returns `*core.PartialSuccessError{Succeeded: 0, Failed: 2}`; `cmdutil.ExitCodeFor(err) == 1`.
- [x] 5.3.4 Preset expansion → `--preset git-workflow` expands via `(*Library).ResolvePreset`; each ref processed; partial-success logic applies.
- [x] 5.3.5 Preset-not-found → `--preset ghost`; `runInit` returns `*core.NotFoundError{Entity: "preset", Name: "ghost"}`; `cmdutil.ExitCodeFor(err) == 2` (gated by §5.0.1).

## 5.4 Update delta specs

- [x] 5.4.1 In `openspec/changes/migrate-init-command/specs/cli-init-command/spec.md`, ensure `MODIFIED Requirements` covers: command-options-pattern, validate required flags, `--library`, `--output-dir` rename, platform validation, dry-run, force; `ADDED Requirements` covers: partial-success semantics, per-resource error rendering, preset expansion via method, preset-not-found → exit 2, Initializer wiring through Factory.
- [x] 5.4.2 In `openspec/changes/migrate-init-command/specs/library-partial-initialization/spec.md`, ensure `REMOVED Requirements` lists the old "nil error on partial success" contract and `ADDED Requirements` cover the new `Initialize` contract, `core.InitializeError.Unwrap` reachability, caller distinguishes partial vs full failure, and preset-not-found → exit 2.

## 5.5 Verification

- [x] 5.5.1 Run `mise run lint` — confirm no new violations.
- [x] 5.5.2 Run `mise run test` — confirm all unit tests pass (including the new partial-success cases and preset-not-found → exit 2).
- [x] 5.5.3 Run `mise run build` — confirm `bin/germinator` builds.
- [x] 5.5.4 Run `mise run test:coverage` — confirm coverage for `cmd/init.go` ≥ 70%.
- [x] 5.5.5 Run `mise run test:e2e` — confirm E2E tests for init pass.
- [x] 5.5.6 Smoke-test:
   - `germinator init --platform opencode --resources skill/commit` → exit 0, success message.
   - `germinator init --platform opencode --resources skill/commit,skill/invalid` → exit 0 with `partial success: 1 succeeded, 1 failed`.
   - `germinator init --platform opencode --resources skill/invalid1,skill/invalid2` → exit 1 with `partial success: 0 succeeded, 2 failed`.
   - `germinator init --platform opencode --preset git-workflow` → expand and process.
   - `germinator init --platform opencode --preset git-workflow --dry-run` → preview only.
   - `germinator init --platform opencode --output-dir /tmp/x --resources skill/commit` → installs to `/tmp/x` (new flag name).
   - `germinator init --platform opencode --preset ghost` → exit 2 with `not found: preset "ghost"`.
- [x] 5.5.7 Update `cmd/AGENTS.md`:
   - Remove the `// Non-migrated command (slice 5 converts it to the NewCmdInit(f, runF) pattern).` comment line from the `init.go` row in the Files table.
   - Move/add `init.go` to the "Canonical examples (slice 5)" list with a one-line note about the partial-success pattern.
   - Update the "slice 7 deletes" notes if any reference `init.go`.
- [x] 5.5.8 Confirm `legacyBridge` still works for non-migrated commands.

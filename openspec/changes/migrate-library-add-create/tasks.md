# Tasks — Migrate library add and library create

**Slice 6 of 9.** Migrates `cmd/library_add.go` (in place; flat layout) with three modes (explicit, discover, batch) and `cmd/library_create.go` (in place; the `library create` Cobra group collapses to a leaf). Adds `core.CanInstallResource` to `internal/core/rules.go` and `core.OperationError` to `internal/core/errors.go` (foundation unit). Adds `ctx` to library I/O functions for cancellation safety. Renames `library.AddOptions` → `library.AddRequest` and `library.OrphanInfo` → `library.Orphan` to align with the request/result convention.

## Task ordering

Tasks execute in numeric order with the following critical paths:

- **6.0 (foundation: OperationError)** blocks **6.4 (library_add migration)** because `runAdd` Mode 2 records per-file `*core.OperationError` instances in the partial-success aggregate.
- **6.1 (CanInstallResource)** blocks **6.4 (library_add Mode 1, Mode 2)** and **6.5 (library_create runCreatePreset)** because both consumers validate refs before I/O.
- **6.2 (ctx + rename)** blocks **6.4** and **6.5** because the inline `resourceAdder` / `presetWriter` interfaces reference the renamed types (`*AddRequest`, `*BatchAddResult`, `[]Orphan`) and the ctx-aware signatures (`AddResource(ctx, ...)` etc.).

Each task ends with `mise run check` passing.

## 6.0 Add `core.OperationError` foundation unit

`core.OperationError` does not exist in `internal/core/errors.go` today. `runAdd` Mode 2 (per-resource `name_conflict`) and the partial-success aggregate depend on it. This group introduces the type, wires `output.FormatError` to render it, and unit-tests the path. Parallel to slice-4 task group 4.0 (`core.NotFoundError`).

- [ ] 6.0.1 In `internal/core/errors.go`, add:
  - `type OperationError struct { Op, Resource string; Cause error }`
  - Constructor `func NewOperationError(op, resource string, cause error) *OperationError`
  - `func (e *OperationError) Error() string { return fmt.Sprintf("%s: %s", e.Op, e.Resource) }`
  - `func (e *OperationError) Unwrap() error { return e.Cause }`
- [ ] 6.0.2 In `internal/output/errors.go`, add a `case errors.As(err, &opErr)` branch in `FormatError`'s switch that calls a new `formatOperationError(io, *core.OperationError)` helper rendering `io.Styles.Error("Error: ") + "<op>: <resource>\n"` to **stderr** via the existing `writeErrOut(io, ...)` helper. Add `var opErr *core.OperationError` to the typed-error block at the top of `FormatError`.
- [ ] 6.0.3 Update `internal/AGENTS.md` line 31 ("typed domain errors" bullet) to add `OperationError` to the listed names.
- [ ] 6.0.4 In `internal/core/errors_test.go`, add a table-driven test `TestOperationError` covering: constructor stores `Op`/`Resource`/`Cause`; `Error()` returns `"<op>: <resource>"`; `Unwrap()` returns the wrapped cause; `errors.As(err, &target)` detects the type.
- [ ] 6.0.5 In `internal/output/output_test.go`, add a `TestFormatError_OperationError` case asserting dispatch via `errors.As` and that `io.ErrOut` (stderr) contains `"Error: register: skill/commit\n"` for `NewOperationError("register", "skill/commit", nil)`. Verify the cause-chain rendering by wrapping a sentinel error.
- [ ] 6.0.6 Run `mise run check`; confirm new unit tests pass and `cmdutil.ExitCodeFor` maps `*core.OperationError` to `ExitCodeError` (1) via the existing default-error case in `internal/cmdutil/exit.go:71`.

## 6.1 Add `core.CanInstallResource`

- [ ] 6.1.1 In `internal/core/rules.go`, add `CanInstallResource(ref string) error`:
  - Use `strings.Cut(ref, "/")` to parse `type/name`
  - If no `/` found or `type` is empty: return `*core.ValidationError{Message: "ref must be type/name"}`
  - If `name` is empty: return `*core.ValidationError{Message: "ref name must be non-empty"}`
  - If `type` is not in `{skill, agent, command, memory}`: return `*core.ValidationError{Message: "ref type must be one of skill, agent, command, memory"}`
  - Otherwise return nil
- [ ] 6.1.2 Add test cases in `internal/core/rules_test.go`:
  - Valid: `skill/commit`, `agent/reviewer`, `command/build`, `memory/project`
  - Invalid: `skills/commit` (wrong type), `skill/` (empty name), `skill` (no slash), `` (empty), `/commit` (empty type)
- [ ] 6.1.3 Confirm `internal/core/rules.go` still imports only stdlib (no `internal/library/`) — depguard enforces

## 6.2 Add `ctx` to library I/O functions and rename types

- [ ] 6.2.1 Update `internal/library/loader.go` (`LoadLibrary`): add `ctx context.Context` as the first parameter; check `ctx.Err()` after each I/O operation; return wrapped `ctx.Err()` on cancellation
- [ ] 6.2.2 Update `internal/library/adder.go` (`DiscoverOrphans` — at `adder.go:724`): add `ctx context.Context` as the first parameter; check `ctx.Err()` between files; return partial results + `ctx.Err()` on cancellation
- [ ] 6.2.3 Update `internal/library/adder.go` (`BatchAddResources` — at `adder.go:526`): add `ctx context.Context` as the first parameter; check `ctx.Err()` after each file in the loop; return partial results + `ctx.Err()` on cancellation; return type is `*library.BatchAddResult` (NOT a slice — keep the pointer per design Decision 6)
- [ ] 6.2.4 Update `internal/library/adder.go` (`AddResource` — at `adder.go:36`): add `ctx context.Context` as the first parameter; check `ctx.Err()` after each I/O operation; return wrapped `ctx.Err()` on cancellation
- [ ] 6.2.5 Rename `library.AddOptions` → `library.AddRequest` and `library.OrphanInfo` → `library.Orphan` in `internal/library/adder.go`; update all intra-package callers in `adder.go` and `adder_test.go`
- [ ] 6.2.6 Update all external callers of these functions. Confirmed call sites today:
  - `cmd/library_add.go` (calls `library.AddResource` via the package function) — updated in task 6.4.x
  - `internal/library/adder.go` (intra-package — `AddResource` and `BatchAddResources` call `LoadLibrary` at `adder.go:58, 641`; updated by 6.2.1-6.2.4)
  - `cmd/library_init.go`, `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_validate.go` (call `LoadLibrary` via the package function; migrate to slice-7)
  - Threading `context.Background()` through the legacy call sites in this slice (mechanical no-op; no behavior change for legacy commands). Each call site gets a one-line `ctx := context.Background()` immediately before the `library.LoadLibrary(...)` call.
- [ ] 6.2.7 Run `mise run check`; confirm zero issues from the API change and the renames
- [ ] 6.2.8 Add `library.ErrNameConflict` sentinel in `internal/library/adder.go`:
  - `var ErrNameConflict = errors.New("name conflict with existing resource")` (exported package-level sentinel; pattern matches slice-5's `*core.NotFoundError` introduction)
  - Rename `hasNameConflict` → `checkNameConflict` with signature `(lib *Library, orphan *Orphan) error` returning `ErrNameConflict` on collision, `nil` otherwise. The boolean return is replaced with a typed error so `runAdd` Mode 2 can wrap it as `Cause` in `*core.OperationError` (task 6.4.4).
  - Update the conflict-check call site (currently `adder.go:799`) to surface `ErrNameConflict` via the result type or as a typed return — choose one approach and document in the implementation commit.
  - Add `TestCheckNameConflict` in `internal/library/adder_test.go` covering: conflict returns `errors.Is(err, library.ErrNameConflict) == true`; no conflict returns `nil`; identical-type duplicate is unaffected (handled by a different code path).
  - Verify `errors.Is(err, library.ErrNameConflict)` works through `*core.OperationError`'s `Unwrap()` chain (covered by the `library-library-orphan-discovery/spec.md:41-47` scenario).
  - Re-run `mise run check` after this sub-task; no separate gate (rolled into 6.2.7's final verification).

## 6.3 Update `cmd/library.go` registration

- [ ] 6.3.1 In `cmd/library.go`, replace the two legacy registrations:
  - `cmd.AddCommand(NewLibraryAddCommand(bridge, &libraryPath))` → `cmd.AddCommand(NewCmdAdd(f, &libraryPath, nil))`
  - `cmd.AddCommand(NewLibraryCreateCommand(bridge, &libraryPath))` → `cmd.AddCommand(NewCmdCreatePreset(f, &libraryPath, nil))`
- [ ] 6.3.2 Run `mise run build`; confirm the build still passes. The legacy constructors become unused as their replacements land: `NewLibraryAddCommand` becomes unused after 6.4.x completes (deleted at the end of 6.4); `NewLibraryCreateCommand` becomes unused after 6.5.x completes (formally deleted in 6.5.7).

## 6.4 Migrate `cmd/library_add.go` (in place; flat layout)

- [ ] 6.4.1 In `cmd/library_add.go`, define `addOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `InputPaths []string`, `Name string`, `Description string`, `Type string`, `Platform string`, `Discover bool`, `Batch bool`, `Force bool`, `DryRun bool`, `Output string`
- [ ] 6.4.2 Declare the **`resourceAdder`** interface in `cmd/library_add.go` (NOT named `Library` — would shadow the `library.Library` struct returned by `f.Library()`). Methods:
  - `AddResource(ctx context.Context, req *library.AddRequest) error`
  - `DiscoverOrphans(ctx context.Context, opts library.DiscoverOptions) (*library.DiscoverResult, error)`
  - `BatchAddResources(ctx context.Context, opts library.BatchAddOptions) (*library.BatchAddResult, error)` (pointer return type per design Decision 6)
- [ ] 6.4.3 Implement `NewCmdAdd(f *cmdutil.Factory, libraryPath *string, runF func(*addOptions) error) *cobra.Command`:
  - Add flags: `--name`, `--description`, `--type`, `--platform`, `--discover`, `--batch`, `--force`, `--dry-run`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` (legacy `--json` is replaced)
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runAdd(opts)`
- [ ] 6.4.4 Implement `runAdd(opts *addOptions) error` (cross-reference: see `cmd/init.go:runInit` for the slice-5 partial-success pattern):
  - **Mode 1 (explicit files):** for each `InputPath`, call `lib.AddResource(opts.Ctx, ...)`; on error, collect into `[]core.InitializeError`; on success, add to successes list. If any failed, return `core.NewPartialSuccessError(...)`.
  - **Mode 2 (`--discover`):** call `lib.DiscoverOrphans(opts.Ctx, ...)`; for each orphan, validate ref via `core.CanInstallResource(orphan.Ref)`; if valid, add; if invalid, skip and collect error. On `name_conflict`, record `*core.OperationError{Op: "register", Resource: <ref>, Cause: <origErr>}` (task 6.0.1) and increment `Failed`. Return `*core.PartialSuccessError` on partial success.
  - **Mode 3 (`--discover --batch --force`):** same as Mode 2 but in a continuous loop; `--force` overrides existing resources; on cancellation, collect partial successes and return wrapped `ctx.Err()`.
- [ ] 6.4.5 Convert `cmd/library_add_test.go` to `iostreams.Test()` + `runF` injection; cover all three modes
- [ ] 6.4.6 Add explicit cancellation test: trigger batch mode with a slow operation, cancel context, verify partial results + non-nil error
- [ ] 6.4.7 Add golden-file tests in `cmd/library_add_test.go`:
  - **Success path**: pin the byte-identical-to-pre-change plain output (per design Decision 9)
  - **Partial-success path**: pin the per-resource status format for Mode 2 with mixed successes / `name_conflict` failures (per design Risks section, "Plain-output byte-identical guarantee" bullet)
- [ ] 6.4.8 Run `mise run check`; confirm all three modes work end-to-end

## 6.5 Migrate `cmd/library_create.go` (in place; group wrapper collapses to a leaf)

- [ ] 6.5.1 In `cmd/library_create.go`, define `createPresetOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Resources []string`, `Description string`, `Force bool`
- [ ] 6.5.2 Declare the **`presetWriter`** interface (NOT named `Library`). Methods:
  - `CreatePreset(ctx context.Context, req *library.CreatePresetRequest) error`
- [ ] 6.5.3 Implement `NewCmdCreatePreset(f *cmdutil.Factory, libraryPath *string, runF func(*createPresetOptions) error) *cobra.Command`:
  - Add flags: `--resources` (required), `--description`, `--force`
  - **No `--output` flag** (legacy didn't have `--json`)
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runCreatePreset(opts)`
- [ ] 6.5.4 Implement `runCreatePreset(opts *createPresetOptions) error`:
  - Validate `opts.Resources` is non-empty (return a usage error if empty — `*core.ValidationError` or wrap a Cobra-required-flag error)
  - For each ref in `opts.Resources`, validate via `core.CanInstallResource(ref)` (fast-fail)
  - Call `lib.CreatePreset(opts.Ctx, &library.CreatePresetRequest{Name: presetName, Resources: opts.Resources, Description: opts.Description, Force: opts.Force})`
- [ ] 6.5.5 Convert `cmd/library_create_test.go` to `iostreams.Test()` + `runF` injection; cover: success (single + multiple resources), description, `--force` overwrite, empty `Resources` returns usage error (exit 2), `CanInstallResource` rejects malformed ref (e.g., `skills/commit`).
- [ ] 6.5.6 Run `mise run check`; confirm `germinator library create preset <name> --resources skill/x,agent/y` works
- [ ] 6.5.7 Delete `cmd/library_create.go:NewLibraryCreateCommand` and the `library create` Cobra group wrapper; confirm `germinator library create preset --help` still resolves. Re-run `mise run build` to confirm no dead-code references.

## 6.6 Verification

- [ ] 6.6.1 Run `mise run lint` — if new `forbidigo` patterns appear (e.g., raw `fmt.Fprintf(os.Stdout, ...)` or `os.Exit(...)` in the new code), refresh `cmd/testdata/lint_baseline.txt` via `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` and commit the baseline alongside the source change
- [ ] 6.6.2 Run `mise run test` — confirm all unit tests pass (including cancellation tests and the new OperationError tests)
- [ ] 6.6.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 6.6.4 Run `mise run test:coverage` — confirm coverage for `cmd/library_add.go`, `cmd/library_create.go`, `internal/library/adder.go`, `internal/core/rules.go`, and `internal/core/errors.go` ≥ 70%
- [ ] 6.6.5 Run `mise run test:e2e` — confirm E2E tests for add and create pass (including all three add modes)
- [ ] 6.6.6 Smoke-test (expected success):
  - [ ] `germinator library add <file> --type skill --name test`
  - [ ] `germinator library add <file> --type skill --name test --dry-run`
  - [ ] `germinator library add --discover`
  - [ ] `germinator library add --discover --batch --force`
  - [ ] `germinator library add --discover --batch --force --output json`
  - [ ] `germinator library add --discover --output table`
  - [ ] `germinator library create preset test-preset --resources skill/x,agent/y`
  - [ ] `germinator library create preset test-preset --resources skill/x,agent/y --description "Test preset"`
- [ ] 6.6.7 Smoke-test (expected failure, validation/usage error → exit 1 or 2):
  - [ ] `germinator library create preset test-preset --resources ""` — expected exit 2 (Cobra required-flag violation → pflag typed error → `cmdutil.ExitCodeFor` returns 2)
  - [ ] `germinator library create preset test-preset --resources skills/commit` — expected exit 1 (invalid type caught by `core.CanInstallResource` → `*core.ValidationError` → `cmdutil.ExitCodeFor` returns 1 via the default-error case at `internal/cmdutil/exit.go:77`)
  - [ ] `germinator library add` (no inputs, no `--discover` — should report usage error) — expected exit 2 (Cobra args-validation error → `cmdutil.ExitCodeFor` returns 2 via the cobraUsagePrefixes branch)
- [ ] 6.6.8 Update `cmd/commands/AGENTS.md`:
  - [ ] Add "Library Add" section with the three modes (explicit files, `--discover`, `--discover --batch --force`), the corresponding flag tables, and per-mode output examples (plain / json / table)
  - [ ] Update "Library Create Preset" section to reflect the leaf command shape (`library create preset` is now a leaf, no group wrapper)
  - [ ] Add a "Stream discipline" note: primary data → `opts.IO.Out`; errors → `opts.IO.ErrOut` via `output.FormatError`; per-resource status → `opts.IO.ErrOut` so it doesn't pollute the JSON stream
- [ ] 6.6.9 Update `cmd/AGENTS.md`:
  - [ ] Add "Canonical example (slice 6)" section mirroring the slice-5 layout, with cross-references to `cmd/library_add.go:runAdd` (partial-success pattern, three modes, conflict handling, cancellation) and `cmd/library_create.go:runCreatePreset` (leaf command shape)
  - [ ] Note that the migrated files stay flat in `cmd/` (per Decision 7 in `design.md`)
  - [ ] Note that the inline interfaces are named after their behavior (`resourceAdder`, `presetWriter`), not after the `library.Library` struct they substitute for

## 7. Cleanup

- [ ] 7.1 Confirm `internal/service/` no longer contains `adder.go` or `creator.go` (currently absent; this is a regression check, not a deletion — files were never created)
- [ ] 7.2 Verify the delta specs at `openspec/changes/migrate-library-add-create/specs/` are valid:
  - [ ] `openspec validate migrate-library-add-create --type change --strict` reports `valid: true`
  - [ ] `openspec show migrate-library-add-create` renders the corrected proposal with the `library-library-json-output`, `library-library-orphan-discovery`, `library-library-resource-import`, `library-library-preset-creation`, and `errors-operation-error` capability names
  - [ ] `openspec show library-library-json-output --type spec` still resolves the main spec
  - [ ] `openspec show library-json-output --type spec` returns 404 (the bad name is fully removed)
  - [ ] `find openspec/changes/migrate-library-add-create -type d -name 'library-json-output'` returns no matches (the unprefixed folder name is gone)

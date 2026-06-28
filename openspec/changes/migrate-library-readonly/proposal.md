# Migrate library presets and library show

## Why

After changes 2 and 3 migrate the core domain commands and `library resources`, the next read-only library commands (`library presets` and `library show`) follow the same template. Migrating them completes the read-only library command set, ahead of the mutating library commands (changes 6, 7) and `init` (change-5).

Slice-1 listed `core.NotFoundError` in `internal/AGENTS.md` line 31 as a typed domain error, but no implementation exists in `internal/core/errors.go`. `library show`'s not-found behavior depends on it, so task group 4.0 introduces the type as a precondition in this slice (rather than a separate foundation slice). Without it, the delta spec's "Ref not found" scenario is unsatisfiable.

## What Changes

### Add foundation unit `core.NotFoundError`

- **ADD** `*core.NotFoundError{Entity, Key string}` to `internal/core/errors.go` with constructor `NewNotFoundError(entity, key string) *NotFoundError` and `func (e *NotFoundError) Error() string { return "not found: " + e.Key }`.
- **ADD** a dispatch branch in `output.FormatError` (`internal/output/errors.go`) that renders `Error: not found: <key>` to stderr via the existing `Styles.Error` channel.
- **UPDATE** `internal/AGENTS.md` line 31 ("typed domain errors" bullet) to confirm the type exists.
- **ADD** unit tests in `internal/core/errors_test.go` and `internal/output/output_test.go` covering constructor, `Error()` string, `errors.As` dispatch, and stderr rendering.

### Migrate `cmd/presets.go`

- **MIGRATE** `cmd/presets.go` (Go package semantics require the file to live in `cmd/`, not `cmd/library/`; see slice-3 archive task 2.3.1 + the "Implementation note" at line 155):
  - Declare `presetsOptions`: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Output string`
  - **Do not declare a `Library` interface.** `library.ListPresets` is a package function (`internal/library/lister.go:41`), not a method on `*library.Library`. Use the concrete `*library.Library` and call `library.ListPresets(lib)` directly (slice-3 precedent; archive task 2.3.3).
  - Constructor signature: `NewCmdPresets(f *cmdutil.Factory, libraryPath *string, runF func(*presetsOptions) error) *cobra.Command`. The `libraryPath *string` is the parent's `--library` shared pointer (matches `NewCmdResources` at `cmd/resources.go:47`).
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` for `--output json|table|plain`.
  - Implement `runPresets(opts)`: dispatch on `opts.Output` like `runResources` at `cmd/resources.go:97-124`. Move `formatPresetsList`, `outputPresetsJSON`, `PresetsJSONOutput`, and `PresetInfoJSON` from `cmd/library_formatters.go` into this file.

### Migrate `cmd/show.go`

- **MIGRATE** `cmd/show.go` (same Go package semantics as above):
  - Declare `showOptions`: `IO`, `Library`, `Ctx`, `Ref string`, `Output string`
  - **Do not declare a `Library` interface.** There is no `(*library.Library).Resolve` method. Dispatch on the ref format inside `cmd/show.go` using `library.ParseRef(opts.Ref)` for resource refs and direct `lib.Presets[name]` map lookup for preset refs.
  - Constructor signature: `NewCmdShow(f *cmdutil.Factory, libraryPath *string, runF func(*showOptions) error) *cobra.Command`.
  - On ref miss, return `core.NewNotFoundError("library ref", opts.Ref)`. `output.FormatError` renders this as `Error: not found: <ref>` to stderr; `cmdutil.ExitCodeFor` maps it to exit 1.
  - **Replace** the legacy brittle string check at `cmd/library.go:129` (`len(ref) > 7 && ref[:7] == "preset/"`) with `strings.HasPrefix(opts.Ref, "preset/")` + `strings.TrimPrefix`.
  - Move `formatResourceDetails`, `formatPresetDetails`, `outputShowResourceJSON`, `outputShowPresetJSON`, `ShowResourceJSONOutput`, and `ShowPresetJSONOutput` from `cmd/library_formatters.go` into this file.

### Update parent `cmd/library.go`

- **MODIFY** `cmd/library.go` lines 42-43: replace `NewLibraryPresetsCommand(bridge, &libraryPath)` and `NewLibraryShowCommand(bridge, &libraryPath)` with `NewCmdPresets(f, &libraryPath, runF)` and `NewCmdShow(f, &libraryPath, runF)` (matching the slice-3 `NewCmdResources` wiring at line 41).
- **DELETE** the legacy `NewLibraryPresetsCommand` and `NewLibraryShowCommand` constructors (no other caller remains after the slice-3 + slice-4 wiring).

### Partial move of `cmd/library_formatters.go`

- **MOVE** six helpers (and their associated types) from `cmd/library_formatters.go` into the per-command files (detailed above).
- **DO NOT DELETE** `cmd/library_formatters.go` in this slice. The remaining helpers (`formatResourcesList`, `formatPresetOutput`, `FormatBatchAddSummary`) are still consumed by `cmd/resources.go` (slice-3), `cmd/library_create.go` (slice-6), and `cmd/library_add.go` (slice-6). The file is deleted in slice-7 alongside `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go`. Add a TODO comment marking the slice-7 deletion.

## Capabilities

### Modified

- **`library-library-json-output`** (delta) — `--output` flag is now available on `library presets` and `library show`. Plain (default) output is byte-identical to the pre-change build for both commands. The legacy `--json` flag **registration was removed in slice-3** but dead `c.Flags().GetBool("json")` reads in `cmd/library.go:81,127` remain. Task 4.1.0 deletes these dead reads along with the legacy constructors. The `--json` short-form, where it was previously documented, has been non-functional since slice-3; consumers now use `--output json`.
- **`cli-error-formatting`** (delta) — `output.FormatError` SHALL dispatch on `*core.NotFoundError` (added in this slice via task group 4.0) and render `Error: not found: <key>` to stderr. `cmdutil.ExitCodeFor` SHALL map it to `ExitCodeError` (1) via the existing default-error case at `internal/cmdutil/exit.go:71`.

## Out of scope (deferred)

- Migrating `init` — change-5
- Migrating `library add`, `library create` — change-6
- Migrating remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) — change-7
- Migrating `config init`, `config validate` — change-8
- Migrating `completion`, `version` — change-9
- Deleting `cmd/library_formatters.go` — slice-7 (after `library_add`/`library_create` migrate)

## Impact

### Affected code

- **Added (new type):** `*core.NotFoundError` in `internal/core/errors.go`
- **Modified (2 files):** `internal/core/errors.go`, `internal/core/errors_test.go` (constructor + unit test for `NotFoundError`)
- **Modified (3 files):** `internal/output/errors.go`, `internal/output/output_test.go`, `internal/AGENTS.md` (dispatch branch + unit test + docs)
- **Rewritten (1 file):** `cmd/presets.go`
- **Rewritten (1 file):** `cmd/show.go`
- **Modified (1 file):** `cmd/library.go` (parent command wiring; lines 42-43)
- **Deleted (2 functions in `cmd/library.go`):** `NewLibraryPresetsCommand`, `NewLibraryShowCommand`
- **Modified (1 file):** `cmd/library_formatters.go` (six helpers removed; TODO comment added; file kept alive until slice-7)
- **Created (1 file):** `cmd/presets_test.go`
- **Created (1 file):** `cmd/show_test.go`
- **Modified (1 file):** `cmd/library_test.go` (`TestLibraryCommand_Presets` migrated to `runF` injection; `TestLibraryCommand_Show` migrated similarly; the no-op skeleton `TestLibraryCommand_InvalidRef` at lines 94-110 is replaced by a `runF`-injection test that asserts `errors.As(err, &core.NotFoundError{})` for `germinator library show invalidformat`)
- **Modified (1 file):** `cmd/AGENTS.md` (canonical-example entry for presets + show)

### Affected systems

- **CLI behavior:** `--output` flag added to `library presets` and `library show` (additive; default `plain` preserves current output).
- **CLI behavior:** `library show <missing-ref>` now returns `*core.NotFoundError` rendered as `Error: not found: <ref>` to stderr (semantic improvement over the legacy `fmt.Errorf("resource not found: ...")`).

### CHANGELOG entry

- **Additive:** `--output json|table|plain` flag is now available on `library presets` and `library show`. Default is `plain`; output is byte-identical to pre-change plain output.
- **Semantic:** `library show <missing-ref>` errors now report as `Error: not found: <ref>` (was `Error: resource not found: <ref>` for resource refs and `Error: preset not found: <name>` for preset refs; the two legacy strings are unified under the new `core.NotFoundError` type).

## Risks

- **`library show` ref resolution is complex** — the command accepts both `type/name` (resource) and `preset/name` formats; resolution logic may need refactoring. **Mitigation:** the resolution logic from `cmd/library.go` moves into `cmd/show.go`; existing tests cover both formats. The brittle `len(ref) > 7 && ref[:7] == "preset/"` check is replaced with `strings.HasPrefix` + `strings.TrimPrefix` per slice-3 modernization principles.
- **`library_formatters.go` partial move leaves residual helpers** — `formatResourcesList`, `formatPresetOutput`, `FormatBatchAddSummary` stay in the file because they are consumed by commands not migrated in this slice. **Mitigation:** a TODO comment marks the slice-7 deletion; tasks 4.4.2 + 4.4.8 verify the residual helpers still work via the unmigrated paths.
- **Introducing `core.NotFoundError` in this slice** couples library-show migration to a foundation change. **Mitigation:** task group 4.0 is self-contained, executes before task 4.2 (numeric ordering enforced), is fully unit-tested before downstream tasks, and is small enough (~30 LOC across two files + tests) to review in isolation. The alternative — a separate foundation slice — was considered and rejected as adding ceremony without proportional benefit.
- **Constructor signature change to `NewCmdXxx(f, libraryPath *string, runF)`** differs from slice-3's documented `NewCmdXxx(f, runF)` signature in the prose AGENTS.md guidance (slice-3's `NewCmdResources` happens to also take `libraryPath`, but other pilots like `NewCmdAdapt` do not). **Mitigation:** the per-library-command signature pattern is consistent across all library sub-commands (slice-3's resources + slice-4's presets + slice-4's show + future slice-7's other library commands); the `libraryPath *string` is a library-command-group concern, documented in the slice-3 archive task 2.3.1.
- **`legacyBridge` keeps `cmd/library_add.go` and `cmd/library_create.go` alive** via `FormatBatchAddSummary` and `formatPresetOutput`. **Mitigation:** task 4.4.8 smoke-tests the unmigrated commands end-to-end to confirm the bridge still works after this slice.

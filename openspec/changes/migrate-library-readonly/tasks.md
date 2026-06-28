# Tasks — Migrate library presets and library show

**Slice 4 of 9.** Migrates `cmd/presets.go` and `cmd/show.go` (Go package semantics require these to live in `cmd/`, not `cmd/library/`; see slice-3 archive task 2.3.1). Adds `--output` flag to both. Introduces the `core.NotFoundError` foundation unit (task group 4.0) that `library show` depends on. Partially moves `cmd/library_formatters.go` helpers to per-command files; the file itself is deleted in slice-7 (not this slice) because `cmd/resources.go` (slice-3), `cmd/library_add.go` (slice-6), and `cmd/library_create.go` (slice-6) still consume helpers from it.

Tasks execute in numeric order. Task group **4.0 (foundation) blocks 4.2 (show migration)** because `runShow` returns `*core.NotFoundError`. Each task ends with `mise run check` passing.

## 4.0 Add `core.NotFoundError` foundation unit

Slice-1 listed `NotFoundError` in `internal/AGENTS.md` line 31 but never implemented it. Task group 4.0 introduces the type, wires `output.FormatError` to render it, and unit-tests the path. `library show`'s not-found behavior (task 4.2.3) depends on this group.

- [ ] 4.0.1 In `internal/core/errors.go`, add `type NotFoundError struct { Entity, Key string }`, a constructor `NewNotFoundError(entity, key string) *NotFoundError`, and `func (e *NotFoundError) Error() string { return "not found: " + e.Key }`. Update `internal/AGENTS.md` line 31 ("typed domain errors" bullet) to confirm the type exists (remove the parenthetical "(renamed from `internal/domain/` in slice 1)" if it remains, and add `NotFoundError` to the listed names).
- [ ] 4.0.2 In `internal/output/errors.go`, add a `case errors.As(err, &notFound)` branch in `FormatError`'s switch that calls a new `formatNotFoundError(io, *core.NotFoundError)` helper rendering `io.Styles.Error("Error: ") + "not found: " + e.Key + "\n"` to **stderr** via `writeErrOut(io, ...)`. Add `var notFound *core.NotFoundError` to the typed-error block at the top of `FormatError`.
- [ ] 4.0.3 In `internal/core/errors_test.go`, add a table-driven test `TestNotFoundError` covering: constructor stores `Entity`/`Key`; `Error()` returns `"not found: <Key>"`; `errors.As(err, &target)` detects the type. In `internal/output/output_test.go`, add a `TestFormatError_NotFound` case asserting dispatch via `errors.As` and that `io.ErrOut` (stderr) contains `"Error: not found: nonexistent-ref\n"`.
- [ ] 4.0.4 Run `mise run check`; confirm new unit tests pass and `cmdutil.ExitCodeFor` maps `*core.NotFoundError` to `ExitCodeError` (1) via the existing default-error case in `internal/cmdutil/exit.go:71`.

## 4.1 Migrate `cmd/presets.go`

- [ ] 4.1.0 Update the parent `cmd/library.go`: replace lines 42-43 (`NewLibraryPresetsCommand(bridge, &libraryPath)` and `NewLibraryShowCommand(bridge, &libraryPath)`) with `NewCmdPresets(f, &libraryPath, runF)` and `NewCmdShow(f, &libraryPath, runF)` (slice-3 `NewCmdResources` signature pattern). Delete the legacy `NewLibraryPresetsCommand` and `NewLibraryShowCommand` constructors entirely (no other caller remains after the slice-3 + slice-4 wiring). Confirm `cmd/library.go:41` (`NewCmdResources(f, &libraryPath, runF)`) still compiles unchanged.
- [ ] 4.1.1 In `cmd/presets.go`, define `presetsOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Output string`
- [ ] 4.1.2 **Do not declare a `Library` interface** in this file. `library.ListPresets` is a **package function** (`internal/library/lister.go:41`) with signature `func ListPresets(lib *Library) []Preset` — it is not a method on `*library.Library`. Use the concrete `*library.Library` returned by `opts.Library()` and call `library.ListPresets(lib)` directly. (Matches slice-3 archive task 2.3.3 precedent.)
- [ ] 4.1.3 Implement `NewCmdPresets(f *cmdutil.Factory, libraryPath *string, runF func(*presetsOptions) error) *cobra.Command`:
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, and the parsed `--output` flag
  - Resolve the library path **per invocation** from `*libraryPath` and `os.Getenv("GERMINATOR_LIBRARY")` via `library.FindLibrary(...)`, then construct `opts.Library = func() (*library.Library, error) { return library.LoadLibrary(resolved) }` (matches `cmd/resources.go:74-81` pattern; the parent's `--library` flag must be honored)
  - Call `runF(opts)` if non-nil, else `runPresets(opts)`
- [ ] 4.1.4 Implement `runPresets(opts *presetsOptions) error`:
  - Resolve `lib, err := opts.Library()`, then call `library.ListPresets(lib)` (note: `ListPresets` does NOT take `ctx`; it's `func ListPresets(lib *Library) []Preset`)
  - Dispatch on `opts.Output` (plain / table / JSON) using `output.NewJSONExporter()` / `output.NewTableExporter()` for non-plain formats (matches slice-3 `runResources` pattern at `cmd/resources.go:97-124`)
  - Move `formatPresetsList`, `outputPresetsJSON`, `PresetsJSONOutput`, and `PresetInfoJSON` from `cmd/library_formatters.go` into this file as private functions/types (task 4.3.2a)
- [ ] 4.1.5 Create `cmd/presets_test.go` (not `cmd/library/presets_test.go`; Go package semantics) modeled on `cmd/resources_test.go`: helper `loadFixtureLibrary` + `newPresetsTestIO` + `newPresetsOpts(t, lib, output)` builders, then `TestRunPresets_Plain` (asserts byte-identical match against `formatPresetsList(lib)`), `TestRunPresets_PlainIsDefault` (asserts `""` and `"plain"` produce identical output), `TestRunPresets_JSON` (asserts 2-space indent JSON shape), `TestRunPresets_Table` (asserts TableExporter output), `TestRunPresets_StreamDiscipline` (asserts stderr is empty for non-verbose invocations). Also migrate `TestLibraryCommand_Presets` in `cmd/library_test.go:40` from the legacy `newTestConfig()` path to `runF` injection + `iostreams.Test()`.
- [ ] 4.1.6 Run `mise run check`; confirm `germinator library presets` produces byte-identical plain output, valid JSON (json), valid table (table), and that `cmd/resources_test.go` still passes (shared `formatResourcesList` helper still lives in `cmd/library_formatters.go` until slice-7).

## 4.2 Migrate `cmd/show.go`

- [ ] 4.2.1 In `cmd/show.go`, define `showOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Ref string`, `Output string`
- [ ] 4.2.2 **Do not declare a `Library` interface** in this file. There is no `(*library.Library).Resolve` method (verified: `internal/library/resolver.go` exposes package functions `ResolveResource`, `ResolvePreset`, plus `ParseRef`, `ValidateRef`). Use the concrete `*library.Library` returned by `opts.Library()` and dispatch on the ref format inside `cmd/show.go`.
- [ ] 4.2.3 Implement `NewCmdShow(f *cmdutil.Factory, libraryPath *string, runF func(*showOptions) error) *cobra.Command` and `runShow(opts *showOptions) error`:
  - In `RunE`: populate `opts` from `f.IOStreams`, `c.Context()`, `args[0]` (the ref), and the parsed `--output` flag. Resolve the library path per invocation via `library.FindLibrary(*libraryPath, os.Getenv("GERMINATOR_LIBRARY"))` (matches `cmd/resources.go:74-81` pattern).
  - In `runShow`: dispatch on the ref format. Resource refs (`type/name`) → call `library.ParseRef(opts.Ref)`, then look up `lib.Resources[typ][name]` directly (matches the existing `formatResourceDetails` pattern at `cmd/library_formatters.go:83-107`; see design Decision 3 for the rationale against introducing a `Resolve` method on `*library.Library` in this slice). Preset refs (`preset/name`) → strip the `preset/` prefix with `strings.TrimPrefix`, then look up `lib.Presets[name]`. **Replace the legacy brittle string check at `cmd/library.go:129` (`len(ref) > 7 && ref[:7] == "preset/"`) with `strings.HasPrefix(opts.Ref, "preset/")` + `strings.TrimPrefix`.**
  - On miss, return `core.NewNotFoundError("library ref", opts.Ref)`. `output.FormatError` (task 4.0.2) renders this as `Error: not found: <ref>` to stderr; `cmdutil.ExitCodeFor` maps it to exit 1.
  - Dispatch on `opts.Output` for the rendered output (plain / table / JSON).
  - Move `formatResourceDetails`, `formatPresetDetails`, `outputShowResourceJSON`, `outputShowPresetJSON`, `ShowResourceJSONOutput`, and `ShowPresetJSONOutput` from `cmd/library_formatters.go` into this file as private functions/types (task 4.3.2b).
- [ ] 4.2.4 Create `cmd/show_test.go` (not `cmd/library/show_test.go`; Go package semantics) modeled on `cmd/resources_test.go`: helper `loadFixtureLibrary` + `newShowTestIO` + `newShowOpts(t, lib, ref, output)` builders, then table-driven sub-tests under `TestRunShow` covering: resource ref plain/JSON/table, preset ref plain/JSON/table, not-found ref (asserts `errors.As(err, &core.NotFoundError{})` and `cmdutil.ExitCodeFor(err) == cmdutil.ExitCodeError`), invalid ref format (e.g., `""`, `"no-slash"`). Add `TestRunShow_StreamDiscipline` asserting stderr contains the not-found error and stdout is empty. Also migrate `TestLibraryCommand_Show` in `cmd/library_test.go` (presets-style block) from the legacy `newTestConfig()` path to `runF` injection + `iostreams.Test()`.
- [ ] 4.2.5 Run `mise run check`; confirm `germinator library show <ref>` works for both resource and preset refs in all three output formats, and `germinator library show nonexistent-ref` exits 1 with `Error: not found: nonexistent-ref` on stderr.

## 4.3 Partial move of `cmd/library_formatters.go`

- [ ] 4.3.1 Verify all moveable helpers in `cmd/library_formatters.go` have been moved to per-command files: `formatPresetsList`, `outputPresetsJSON`, `PresetsJSONOutput`, `PresetInfoJSON` → `cmd/presets.go` (from task 4.1.4); `formatResourceDetails`, `formatPresetDetails`, `outputShowResourceJSON`, `outputShowPresetJSON`, `ShowResourceJSONOutput`, `ShowPresetJSONOutput` → `cmd/show.go` (from task 4.2.3). Run `rg "formatPresetsList|formatResourceDetails|formatPresetDetails|outputPresetsJSON|outputShowResourceJSON|outputShowPresetJSON|PresetsJSONOutput|PresetInfoJSON|ShowResourceJSONOutput|ShowPresetJSONOutput" cmd/` and confirm zero matches remain outside the two new files.
- [ ] 4.3.2 `cmd/library_formatters.go` is **NOT deleted in this slice**. The remaining helpers (`formatResourcesList`, `formatPresetOutput`, `FormatBatchAddSummary`) are still consumed by `cmd/resources.go:119` (slice-3), `cmd/library_create.go:151` (slice-6), and `cmd/library_add.go:311,474` (slice-6). The file is deleted in slice-7 alongside `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go` (per slice-3 archive tasks 2.5.x). Add a one-line TODO comment at the top of `cmd/library_formatters.go`: `// TODO(slice-7): delete this file once slice-3/slice-6 consumers migrate off the remaining helpers.`

## 4.4 Verification

- [ ] 4.4.1 Run `mise run lint` — confirm no new violations. If new intentional violations were introduced (e.g., flag-string captures by `forbidigo` patterns in the migrated `cmd/presets.go` / `cmd/show.go` files), refresh the baseline via `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` and commit it alongside (per `cmd/AGENTS.md` "Lint Enforcement").
- [ ] 4.4.2 Run `mise run test` — confirm all unit tests pass (including the new `TestNotFoundError` and `TestFormatError_NotFound` from task 4.0.3, the migrated `cmd/presets_test.go` and `cmd/show_test.go`, and the unchanged `cmd/resources_test.go` whose `formatResourcesList` import remains valid)
- [ ] 4.4.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 4.4.4 Run `mise run test:coverage` — confirm coverage for `cmd/presets.go`, `cmd/show.go`, `internal/core/errors.go` (NotFoundError), and `internal/output/errors.go` (formatNotFoundError) ≥ 70%
- [ ] 4.4.5 Run `mise run test:e2e` — confirm E2E tests for presets and show pass
- [ ] 4.4.6 Smoke-test:
  - `germinator library presets`
  - `germinator library presets --output json`
  - `germinator library presets --output table`
  - `germinator library presets --output plain` (explicit default; byte-identical to no-flag invocation)
  - `germinator library presets --library /tmp/foo` (verify the parent's `--library` flag is honored via the `libraryPath *string` pattern)
  - `germinator library show <resource-ref>`
  - `germinator library show <preset-ref>`
  - `germinator library show nonexistent-ref` (assert exit 1 and stderr `Error: not found: nonexistent-ref`; stdout empty)
  - `germinator library show ""` (assert exit 1 and stderr `Error: not found: `; covers the invalid-empty edge case)
  - `germinator library show 'preset/'` (assert exit 1 and stderr `Error: not found: preset/`; covers the empty-name-after-prefix edge case)
  - `germinator library show skill/commit --output json | jq .` (verify the JSON is pipeable; no stderr leakage)
- [ ] 4.4.7 Update `cmd/AGENTS.md` (the file exists — `cmd/library/AGENTS.md` does not, Go package semantics) with the new pattern for presets and show. Add a "Canonical examples" entry mirroring the `cmd/resources.go` block at `cmd/AGENTS.md:100+`. Reference `cmd/presets.go` and `cmd/show.go` as the new templates.
- [ ] 4.4.8 Confirm `legacyBridge` still works for non-migrated commands (`germinator library add --help`, `germinator library create --help`, `germinator library init --help`, `germinator library refresh --help`, `germinator library remove --help`, `germinator library validate --help`). Smoke-test `germinator library add /nonexistent.md` end-to-end (the unmigrated path still works via `legacyBridge` + `cmd/library_formatters.go`'s `FormatBatchAddSummary`).

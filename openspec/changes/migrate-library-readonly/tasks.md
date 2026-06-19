# Tasks — Migrate library presets and library show

**Slice 4 of 9.** Migrates `cmd/library/presets.go` and `cmd/library/show.go` to the new pattern. Adds `--output` flag to both. Deletes `cmd/library_formatters.go`.

Each task ends with `mise run check` passing.

## 4.1 Migrate `cmd/library/presets.go`

- [ ] 4.1.1 In `cmd/library/presets.go`, define `presetsOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Output string`
- [ ] 4.1.2 Declare the `Library` interface in `cmd/library/presets.go` with the methods called: `ListPresets(ctx) ([]library.Preset, error)`
- [ ] 4.1.3 Implement `NewCmdPresets(f *cmdutil.Factory, runF func(*presetsOptions) error) *cobra.Command`:
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and the parsed `--output` flag
  - Call `runF(opts)` if non-nil, else `runPresets(opts)`
- [ ] 4.1.4 Implement `runPresets(opts *presetsOptions) error`:
  - Call `lib.ListPresets(opts.Ctx)`
  - Dispatch on `opts.Output` (plain / table / JSON) using `output.Exporter` for non-plain formats
  - Move any formatter helper from `cmd/library_formatters.go` into this file as a private function
- [ ] 4.1.5 Convert `cmd/library/presets_test.go` (if it exists; or add new tests) to `iostreams.Test()` + `runF` injection
- [ ] 4.1.6 Run `mise run check`; confirm `germinator library presets` produces byte-identical output (plain), valid JSON (json), valid table (table)

## 4.2 Migrate `cmd/library/show.go`

- [ ] 4.2.1 In `cmd/library/show.go`, define `showOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Ref string`, `Output string`
- [ ] 4.2.2 Declare the `Library` interface with the method called: `Resolve(ctx, ref string) (*library.Resource, error)` (covering both resource refs `type/name` and preset refs `preset/name`)
- [ ] 4.2.3 Implement `NewCmdShow(f, runF)` and `runShow(opts)`:
  - Parse `opts.Ref` to determine if it's a resource ref or a preset ref
  - Call the appropriate `Library.Resolve` method (or the unified `Resolve`)
  - Return `*core.NotFoundError` if the ref doesn't resolve
  - Dispatch on `opts.Output` for the rendered output
  - Move any formatter helper from `cmd/library_formatters.go` into this file as a private function
- [ ] 4.2.4 Convert `cmd/library/show_test.go` (or add new tests) to `iostreams.Test()` + `runF` injection; cover resource ref, preset ref, and not-found cases
- [ ] 4.2.5 Run `mise run check`; confirm `germinator library show <ref>` works for both resource and preset refs in all three output formats

## 4.3 Delete legacy formatters

- [ ] 4.3.1 Verify all helpers in `cmd/library_formatters.go` have been moved to per-command files (or to `internal/output/`); `rg "library_formatters" cmd/` should return no remaining imports
- [ ] 4.3.2 Delete `cmd/library_formatters.go`
- [ ] 4.3.3 Delete `internal/service/lister.go` if it exists (listing logic moved into command files)

## 4.4 Verification

- [ ] 4.4.1 Run `mise run lint` — confirm no new violations
- [ ] 4.4.2 Run `mise run test` — confirm all unit tests pass
- [ ] 4.4.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 4.4.4 Run `mise run test:coverage` — confirm coverage for `cmd/library/presets.go` and `cmd/library/show.go` ≥ 70%
- [ ] 4.4.5 Run `mise run test:e2e` — confirm E2E tests for presets and show pass
- [ ] 4.4.6 Smoke-test:
  - `germinator library presets`
  - `germinator library presets --output json`
  - `germinator library presets --output table`
  - `germinator library show <resource-ref>`
  - `germinator library show <preset-ref>`
  - `germinator library show nonexistent-ref` (exits 1 with formatted NotFoundError)
- [ ] 4.4.7 Update `cmd/library/AGENTS.md` (create it if it doesn't exist) with the new pattern for presets and show
- [ ] 4.4.8 Confirm `legacyBridge` still works for non-migrated commands

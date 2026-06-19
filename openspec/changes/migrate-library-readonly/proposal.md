# Migrate library presets and library show

## Why

After changes 2 and 3 migrate the core domain commands and `library resources`, the next read-only library commands (`library presets` and `library show`) follow the same template. Migrating them completes the read-only library command set, ahead of the mutating library commands (changes 6, 7) and `init` (change-5).

## What Changes

### Migrate `cmd/library/presets.go`

- **MIGRATE** `cmd/library/presets.go`:
  - Declare `presetsOptions`: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Output string`
  - Declare the `Library` interface (methods: `ListPresets(ctx) ([]library.Preset, error)`)
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` for `--output json|table|plain`
  - Implement `NewCmdPresets(f, runF)` and `runPresets(opts)`: dispatch on `opts.Output` like `library resources`

### Migrate `cmd/library/show.go`

- **MIGRATE** `cmd/library/show.go`:
  - Declare `showOptions`: `IO`, `Library`, `Ctx`, `Ref string`, `Output string`
  - Add `--output` flag via `cmdutil.AddOutputFlags`
  - Implement `NewCmdShow(f, runF)` and `runShow(opts)`:
    - Resolve `opts.Ref` (resource or preset) via the library
    - Return `core.NotFoundError` if not found
    - Dispatch on `opts.Output` (plain / table / JSON)

### Delete legacy files

- **DELETE** `cmd/library_formatters.go` (its helpers move to the per-command files or `internal/output/`)
- **DELETE** `internal/service/lister.go` if it exists (the listing logic moves to the command files)

## Capabilities

### Modified

- **`library/library-json-output`** (delta) — `--output` flag is now available on `library presets` and `library show` (the legacy `--json` flag, if any, is replaced).

## Out of scope (deferred)

- Migrating `init` — change-5
- Migrating `library add`, `library create` — change-6
- Migrating remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) — change-7
- Migrating `config init`, `config validate` — change-8
- Migrating `completion`, `version` — change-9

## Impact

### Affected code

- **Rewritten (1 file):** `cmd/library/presets.go`
- **Rewritten (1 file):** `cmd/library/show.go`
- **Deleted (1 file):** `cmd/library_formatters.go` (formatters move into per-command files)
- **Deleted (1 file, if present):** `internal/service/lister.go`
- **Modified (1 file):** `cmd/library/presets_test.go` (converted to `iostreams.Test()` + `runF` injection)
- **Modified (1 file):** `cmd/library/show_test.go` (converted similarly)

### Affected systems

- **CLI behavior:** `--output` flag added to `library presets` and `library show` (additive; default `plain` preserves current output)

## Risks

- **`library show` ref resolution is complex** — the command accepts both `type/name` (resource) and `preset/name` formats; resolution logic may need refactoring. **Mitigation:** the resolution logic from `cmd/library.go` moves into `cmd/library/show.go`; existing tests cover both formats.
- **`library_formatters.go` helpers are shared** — multiple read-only library commands use the same formatters; moving them per-command could duplicate code. **Mitigation:** the formatters are simple enough that duplication is acceptable; if complexity grows, they move to `internal/output/` in a future change.
- **`Library` interface methods differ per command** — `resources` needs `ListResources`, `presets` needs `ListPresets`, `show` needs `Resolve`. **Mitigation:** each command declares its own minimal interface (per the `application/command-options-pattern` capability); the production `Library` type implements all methods.

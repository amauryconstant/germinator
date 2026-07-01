# Migrate config init and config validate

## Why

The config commands are isolated (no library dependency) and small (2 commands). They follow the same template as the pilots and domain commands but have one nuance: the legacy `--output` flag (which wrote the config to a file path) must be renamed to `--output-path` to disambiguate from the new `--output` format flag.

## What Changes

### Split `cmd/config.go` into per-command files (flat)

- **KEEP** `cmd/config.go` as the parent command only (slimmed: remove the `init`/`validate` constructors and the `scaffoldedConfig` template); the parent has no flags itself
- Sub-commands become separate files in `cmd/` (`package cmd`), matching the flat layout used by every other migrated group (`library_*`, `init`, `adapt`, `validate`, …). No `cmd/config/` sub-directory is introduced (it would collide with the existing `internal/config` package)
- Update `cmd/config.go`'s `AddCommand` calls to `NewCmdConfigInit(f, nil)` / `NewCmdConfigValidate(f, nil)`, and stop discarding the Factory in `NewConfigCommand` (currently `NewConfigCommand(_ *cmdutil.Factory)`)

### Migrate `cmd/config_init.go`

- **MIGRATE** to `cmd/config_init.go` (new file, `package cmd`):
  - Declare `configInitOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`, `Force bool`
  - Implement `NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command`:
    - Add `--output-path` (string, the file path to write the config; **renamed from legacy `--output`**)
    - Add `--force` flag
    - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, and parsed flags
    - Call `runF(opts)` if non-nil, else `runConfigInit(opts)`
  - Implement `runConfigInit(opts *configInitOptions) error`:
    - Resolve `OutputPath` (default `$XDG_CONFIG_HOME/germinator/config.toml`)
    - Check if file exists; if so and `--force` not set, return `core.NewFileError(opts.OutputPath, "create", "config file already exists (use --force to overwrite)", nil)` (constructor — `FileError` fields are unexported)
    - Write the documented config template (the `scaffoldedConfig` constant moved out of `cmd/config.go`) to the file
    - Write a single success line to `opts.IO.Out`; nothing on `opts.IO.ErrOut`

### Migrate `cmd/config_validate.go`

- **MIGRATE** to `cmd/config_validate.go` (new file, `package cmd`):
  - Declare `configValidateOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`
  - Implement `NewCmdConfigValidate(f, runF)` and `runConfigValidate(opts)`:
    - Add `--output-path` (string, the file path to validate)
    - Read the file at `OutputPath`; on missing file return `core.NewFileError(opts.OutputPath, "read", "config file not found", statErr)` (not-found derived from the wrapped cause)
    - Delegate validation to `config.Validate()` (platform-only today); **return** the error — do not render inline (single-handling rule; `main.go` renders once via `output.FormatError`)
    - On success, write a single line to `opts.IO.Out`; nothing on `opts.IO.ErrOut`
- **NO `--output` format flag** (the command produces text output, not structured data)

## Capabilities

### Modified

- **`cli-config-commands`** (delta) — both `config init` and `config validate` follow the new command-options-pattern; the legacy `--output` flag is renamed to `--output-path` to disambiguate from the `cli-output-formats` capability's `--output` flag.

## Out of scope (deferred)

- Migrating `completion`, `version`, deleting `internal/models/`, finalizing `AGENTS.md` + CHANGELOG — change-9
- Swapping `internal/config`'s `internal/models.Platform*` constants → `core.ValidatePlatform` on the validate path — deferred to a dedicated cleanup change (keeps this change scoped to the cmd-layer migration)

## Migration sequence (slice → archive slug)

The slice numbers used in this proposal refer to the nine-step migration plan. Mapping to the archived change slugs for traceability:

| Slice | Archived slug | Subject |
|-------|---------------|---------|
| change-1 | `2026-06-24-scaffold-cli-foundation` | Initial CLI scaffolding |
| change-2 | `2026-06-26-wire-factory-and-pilots` | Factory + exit codes + deprecation canary |
| change-3 | (see `openspec/changes/archive/`) | Library init / show / list |
| change-4 | (see `openspec/changes/archive/`) | Library add / create preset |
| change-5 | (see `openspec/changes/archive/`) | Library refresh / validate / remove |
| change-6 | (see `openspec/changes/archive/`) | Adapt / canonicalize |
| change-7 | `2026-06-30-migrate-library-add-create` | Library add / create preset finalization (legacy shell deletion) |
| **change-8** | **(this change)** | **Config init / validate** |
| change-9 | (pending) | Completion / version / docs finalization |

## Impact

### Affected code

- **Modified (1 file):** `cmd/config.go` (slimmed to parent only; forwards `f` to the sub-command constructors)
- **Added (1 file):** `cmd/config_init.go` (`package cmd`)
- **Added (1 file):** `cmd/config_validate.go` (`package cmd`)
- **Modified (1 file):** `cmd/config_test.go` (split into `config_init_test.go` + `config_validate_test.go`)
- **Added (2 files):** `test/e2e/config_init_test.go`, `test/e2e/config_validate_test.go` — minimal Ginkgo E2E coverage following the `test/e2e/library_init_test.go` pattern

### Affected systems

- **CLI behavior:** the `--output` flag on `config init` and `config validate` is renamed to `--output-path` (BREAKING; mitigated by CHANGELOG entry)

## Risks

- **`--output` → `--output-path` rename is BREAKING** — any script using `germinator config init --output /path/to/config.toml` will break. **Mitigation:** CHANGELOG entry only. (The deprecation canary does NOT cover this — unknown flags yield a Cobra usage error, exit 2, and the canary fires only on exit 1.)
- **Default config path resolution** — `$XDG_CONFIG_HOME` may not be set in some environments. **Mitigation:** fall back to `~/.config/germinator/config.toml`; existing tests cover both paths.
- **Config file format** — the template must match what the existing implementation produces. **Mitigation:** task 8.2.5 byte-compares output against a pre-change fixture.

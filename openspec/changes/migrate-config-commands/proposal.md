# Migrate config init and config validate

## Why

The config commands are isolated (no library dependency) and small (2 commands). They follow the same template as the pilots and domain commands but have one nuance: the legacy `--output` flag (which wrote the config to a file path) must be renamed to `--output-path` to disambiguate from the new `--output` format flag.

## What Changes

### Split `cmd/config.go` into `cmd/config/` sub-directory

- **MOVE** `cmd/config.go` (parent command only) to `cmd/config/root.go`
- The parent command (`config`) has no flags itself; sub-commands carry their own flags

### Migrate `cmd/config/init.go`

- **MIGRATE** `cmd/config/init.go`:
  - Declare `configInitOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`, `Force bool`
  - Implement `NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command`:
    - Add `--output-path` (string, the file path to write the config; **renamed from legacy `--output`**)
    - Add `--force` flag
    - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, and parsed flags
    - Call `runF(opts)` if non-nil, else `runConfigInit(opts)`
  - Implement `runConfigInit(opts *configInitOptions) error`:
    - Resolve `OutputPath` (default `$XDG_CONFIG_HOME/germinator/config.toml`)
    - Check if file exists; if so and `--force` not set, return an error
    - Write the documented config template to the file

### Migrate `cmd/config/validate.go`

- **MIGRATE** `cmd/config/validate.go`:
  - Declare `configValidateOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`
  - Implement `NewCmdConfigValidate(f, runF)` and `runConfigValidate(opts)`:
    - Add `--output-path` (string, the file path to validate)
    - Read the file at `OutputPath`
    - Validate it (using koanf or similar)
    - Print any validation errors via `output.FormatError`
- **NO `--output` format flag** (the command produces text output, not structured data)

## Capabilities

### Modified

- **`config-commands`** (delta) — both `config init` and `config validate` follow the new command-options-pattern; the legacy `--output` flag is renamed to `--output-path` to disambiguate from the `output-formats` capability's `--output` flag.

## Out of scope (deferred)

- Migrating `completion`, `version`, deleting `internal/models/`, finalizing `AGENTS.md` + CHANGELOG — change-9

## Impact

### Affected code

- **Moved (1 file):** `cmd/config.go` → `cmd/config/root.go`
- **Rewritten (1 file):** `cmd/config/init.go` (new file in the sub-directory)
- **Rewritten (1 file):** `cmd/config/validate.go` (new file in the sub-directory)
- **Modified (1 file):** `cmd/config_test.go` (split into per-command test files)

### Affected systems

- **CLI behavior:** the `--output` flag on `config init` and `config validate` is renamed to `--output-path` (BREAKING; mitigated by CHANGELOG entry)

## Risks

- **`--output` → `--output-path` rename is BREAKING** — any script using `germinator config init --output /path/to/config.toml` will break. **Mitigation:** CHANGELOG entry documents the rename; the deprecation canary from change-2 (`EXIT_CODE_LEGACY`) emits a warning when unknown flags are used.
- **Default config path resolution** — `$XDG_CONFIG_HOME` may not be set in some environments. **Mitigation:** fall back to `~/.config/germinator/config.toml`; existing tests cover both paths.
- **Config file format** — the template must match what the existing implementation produces. **Mitigation:** task 8.2.3 byte-compares output against a pre-change fixture.

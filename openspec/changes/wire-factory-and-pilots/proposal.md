# Wire Factory and migrate pilot commands

## Why

After the foundation packages land (change-1), the next risk is the integration point: `main.go` must consume `iostreams/output/cmdutil`, and at least two commands must prove the new pattern works end-to-end before the bulk migration begins. This change lands the wiring + the two pilots (`adapt` and `library resources`), establishes the `legacyBridge` shim that lets non-migrated commands coexist, and deletes the legacy `cmd/container.go`, `cmd/command_config.go`, and `cmd/error_handler.go`.

## What Changes

### main.go rewiring

- **REWRITE** `main.go` body:
  - Construct root context via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` with `defer cancel()`
  - Construct `IOStreams` via `iostreams.System()`
  - Construct `Factory` with `IOStreams`, `AppVersion`, `Executable`, `RootContext`
  - Populate every lazy function field (`Config`, `Library`, `Transformer`, `Validator`, `Canonicalizer`, `Initializer`)
  - Set `rootCmd.SetContext(ctx)` so `c.Context()` returns the signal-aware context
- **REPLACE** post-`Execute` error handling with `output.FormatError(f.IOStreams, err)` followed by `os.Exit(int(cmdutil.ExitCodeFor(err)))`
- **ADD** `legacyBridge` shim: a temporary struct holding `Services *ServiceContainer`, `ErrorFormatter *ErrorFormatter`, `Verbosity Verbosity` — populated by calling Factory functions. Used by non-migrated commands until change-7 deletes it.

### Pilot: `cmd/adapt.go`

- **MIGRATE** `cmd/adapt.go`:
  - Declare `adaptOptions` struct: `IO *iostreams.IOStreams`, `Transformer func() (Transformer, error)`, `Ctx context.Context`, `InputPath`, `OutputPath`, `Platform string`
  - Declare the `Transformer` interface in the same file (one method: `Transform(ctx, *TransformRequest) (*core.TransformResult, error)`)
  - Implement `NewCmdAdapt(f *cmdutil.Factory, runF func(*adaptOptions) error) *cobra.Command`
  - Implement `runAdapt(opts *adaptOptions) error`: validate platform via `core.ValidatePlatform`, call `transformer.Transform`, write success to `opts.IO.Out`, verbose progress to `opts.IO.Verbosef`
- **CONVERT** `cmd/cmd_test.go` adapt test to use `iostreams.Test()` + `runF` injection + direct `runAdapt(opts)` invocation

### Pilot: `cmd/library/resources.go`

- **MIGRATE** `cmd/library/resources.go` (split from `cmd/library.go` if needed):
  - Declare `resourcesOptions`: `IO`, `Library func() (*library.Library, error)`, `Output string`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdResources(f, runF)` and `runResources(opts)`: dispatch on `opts.Output` (plain / table / JSON via `output.Exporter`)
- **CONVERT** resources test to `iostreams.Test()` + `runF` injection

### Deletions

- **DELETE** `cmd/container.go`
- **DELETE** `cmd/command_config.go`
- **DELETE** `cmd/error_handler.go`
- **DELETE** legacy body of `cmd/adapt.go` (replaced by migrated version)
- **DELETE** legacy body of `cmd/library/resources.go`

### Behavior changes

- **EXIT codes collapse from 0–6 to 0/1/2** in `main.go`'s error handling
- `--output` flag introduced on `library resources` (replaces any legacy `--json` flag, though `library resources` did not previously have one)

## Capabilities

### Modified (delta spec for library resources)

- **`library/library-json-output`** — The `--output json` flag is now available on `library resources` via `cmdutil.AddOutputFlags`. Plain (default) output is byte-identical to the legacy plain output.

## Out of scope (deferred)

- Migrating `validate`, `canonicalize` — **change-3**
- Migrating `library presets`, `library show` — **change-4**
- Migrating `init` — **change-5**
- Migrating `library add`, `library create` — **change-6**
- Migrating remaining library commands — **change-7**
- Migrating `config init`, `config validate` — **change-8**
- Migrating `completion`, `version` — **change-9**
- Deleting `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go` — **change-7** (after `legacyBridge` removed)

## Impact

### Affected code

- **Rewritten (1 file):** `main.go`
- **Rewritten (1 file):** `cmd/adapt.go`
- **Rewritten (1 file):** `cmd/library/resources.go` (may require splitting from `cmd/library.go`)
- **Deleted (3 files):** `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go`
- **Modified (1 file):** `cmd/library.go` (parent command only; sub-commands migrated)
- **Modified (1 file):** `cmd/cmd_test.go` (adapt test converted)

### Affected systems

- **CLI behavior:** exit codes 3–6 collapse to 1 (BREAKING; mitigated by CHANGELOG entry)
- **CLI behavior:** `--output` flag added to `library resources` (additive)
- **Scripts / CI:** any external consumer reading exit codes 3–6 must adapt

## Risks

- **`legacyBridge` is the only caller of legacy types** — sloppy migration could break non-migrated commands. **Mitigation:** each change's verification confirms `legacyBridge` still works (every command smoke-tested in tasks 1.7.x of this change, then again in changes 3-6).
- **Exit code canary** — emitting the deprecation warning on first `cmdutil.ExitCodeFor` call requires `Factory.IOStreams.Logger` to be wired in `main.go`. **Mitigation:** task 1.5.6 wires this explicitly; the warning is gated on `EXIT_CODE_LEGACY` env or TTY stderr to avoid noise in CI.
- **First time wiring the Factory** — `main.go` becomes the only composition root, which means every lazy function field must be populated correctly or commands fail at runtime. **Mitigation:** task 1.5.5 smoke-tests every command end-to-end; the pilot commands (adapt, library resources) provide positive verification that the wiring works.
- **Tests that depend on `cmd/container.go`** — the `internal/service/*_mock_test.go` mocks and `cmd/cmd_test.go` sections for non-pilot commands still call `NewServiceContainer()`. **Mitigation:** those tests are NOT touched in this change; only the adapt and resources sections are rewritten. The mocks are deleted in change-7.

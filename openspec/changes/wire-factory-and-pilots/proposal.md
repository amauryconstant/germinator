# Wire Factory and migrate pilot commands

## Why

After the foundation packages land (change-1), the next risk is the integration point: `main.go` must consume `iostreams/output/cmdutil`, and at least two commands must prove the new pattern works end-to-end before the bulk migration begins. This change lands the wiring + the two pilots (`adapt` and `library resources`), establishes the `LegacyBridge` shim that lets non-migrated commands coexist, and deletes the legacy `cmd/container.go`, `cmd/command_config.go`, and `cmd/error_handler.go`.

## What Changes

### main.go rewiring

- **REWRITE** `main.go` body:
  - Construct root context via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` with `defer cancel()`
  - Construct `IOStreams` via `iostreams.System()`
  - Construct `Factory` with `IOStreams`, `AppVersion`, `Executable`, `RootContext`
  - Populate every lazy function field (`Config`, `Library`, `Transformer`, `Validator`, `Canonicalizer`, `Initializer`)
  - Set `rootCmd.SetContext(ctx)` so `c.Context()` returns the signal-aware context
- **REPLACE** post-`Execute` error handling with `output.FormatError(f.IOStreams, err)` followed by `os.Exit(int(cmdutil.ExitCodeFor(err)))`
- **ADD** `cmd.LegacyBridge` shim: a temporary struct (declared in `cmd/legacy_bridge.go`, exported so it can cross the package boundary from `main.go` to `cmd/`). Holds `Services *LegacyServices`, `ErrorFormatter *ErrorFormatter`, `Verbosity Verbosity`. The struct is declared in task 2.1.3a (before `cmd/container.go` is deleted); `Services` is populated in task 2.1.3b (after task 2.5.1 deletes `cmd/container.go`) by calling `service.New*` constructors directly in `main.go`. Used by non-migrated commands until slice 7 deletes it.

### Pilot: `cmd/adapt.go`

- **MIGRATE** `cmd/adapt.go`:
  - Declare `adaptOptions` struct: `IO *iostreams.IOStreams`, `Transformer func() (Transformer, error)`, `Ctx context.Context`, `InputPath`, `OutputPath`, `Platform string`
  - Declare the `Transformer` interface in the same file (one method: `Transform(ctx, *TransformRequest) (*core.TransformResult, error)`)
  - Implement `NewCmdAdapt(f *cmdutil.Factory, runF func(*adaptOptions) error) *cobra.Command`
  - Implement `runAdapt(opts *adaptOptions) error`: validate platform via `core.ValidatePlatform`, call `transformer.Transform`, write success to `opts.IO.Out`, verbose progress to `opts.IO.Verbosef`
- **CONVERT** `cmd/cmd_test.go` adapt test to use `iostreams.Test()` + `runF` injection + direct `runAdapt(opts)` invocation

### Pilot: `library resources` (lives at `cmd/resources.go`)

- **MIGRATE** the `library resources` subcommand (originally proposed at `cmd/library/resources.go`, but Go package semantics require it to live in the same directory as the rest of the `cmd` package; the actual file is `cmd/resources.go`):
  - Declare `resourcesOptions`: `IO`, `Library func() (*library.Library, error)`, `Output string`
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Implement `NewCmdResources(f, runF)` and `runResources(opts)`: dispatch on `opts.Output` (plain / table / JSON via `output.Exporter`)
- **CONVERT** resources test to `iostreams.Test()` + `runF` injection

### Deletions

- **DELETE** `cmd/container.go`
- **DELETE** `cmd/command_config.go`
- **DELETE** `cmd/error_handler.go`
- **DELETE** legacy body of `cmd/adapt.go` (replaced by migrated version)
- **DELETE** legacy body of the `library resources` subcommand (now at `cmd/resources.go`)

### Behavior changes

- **EXIT codes collapse from 0–6 to 0/1/2** in `main.go`'s error handling
- `--output` flag introduced on `library resources` (replaces the parent-inherited `--json` flag from base spec `library-library-json-output`; consumers must switch from `--json` to `--output json`)

## Capabilities

### Modified (delta spec for library resources)

- **`library-library-json-output`** — The `--output json|table|plain` flag is now available on `library resources` via `cmdutil.AddOutputFlags`. Plain (default) output is byte-identical to the legacy plain output. The 7 obsolete `--json` parent-inherited requirements are explicitly REMOVED (1 parent + 6 sub-commands: resources, presets, remove, add, show, init).

### Modified (delta spec for exit code canary)

- **`cli-exit-codes`** — A new requirement "Exit code deprecation canary" is ADDED, defining the one-time stderr warning emitted from `main.go`'s post-`Execute` path when the exit code is `1` and either `EXIT_CODE_LEGACY` is set or stderr is a TTY. `cmdutil.ExitCodeFor` purity is preserved.

## Out of scope (deferred)

- Migrating `validate`, `canonicalize` — **change-3**
- Migrating `library presets`, `library show` — **change-4**
- Migrating `init` — **change-5**
- Migrating `library add`, `library create` — **change-6**
- Migrating remaining library commands — **change-7**
- Migrating `config init`, `config validate` — **change-8**
- Migrating `completion`, `version` — **change-9**
- Deleting `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go` — **change-7** (after `LegacyBridge` removed)

## Impact

### Affected code

- **Rewritten (1 file):** `main.go`
- **Rewritten (1 file):** `cmd/adapt.go`
- **Rewritten (1 file):** `cmd/resources.go` (originally proposed as `cmd/library/resources.go`; relocated to `cmd/resources.go` because Go requires same-directory files to share a `package` declaration). The parent `library` command stays in `cmd/library.go`.
- **Deleted (3 files):** `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go`
- **Modified (1 file):** `cmd/library.go` (parent command only — also remove the `--json` persistent flag per task 2.3.1; sub-commands migrated)
- **Modified (1 file):** `cmd/cmd_test.go` (adapt test converted)
- **Modified (1 file):** `cmd/root.go` (`NewRootCommand` signature changes to `NewRootCommand(f *cmdutil.Factory, bridge *LegacyBridge)`)
- **Modified (1 file):** `internal/iostreams/iostreams.go` (add `IsStderrTTY()` / `SetStderrTTY()` methods and `Warnf()` method per decisions 1 and 2)
- **Modified (1 file):** `internal/cmdutil/factory.go` (`NewFactory` signature changes to `NewFactory(ctx context.Context, io *iostreams.IOStreams, appVersion, executable string)` per decision 4)
- **New (1 file):** `cmd/legacy_bridge.go` (`LegacyBridge` + `LegacyServices` types per decision 3)
- **New (1 file):** `internal/warning/canary.go` (`MaybeWarnLegacyExitCode` + `ResetCanaryForTest` helpers)
- **New (1 file):** `internal/warning/canary_test.go` (six scenarios from task 2.4.4)
- **New (1 file):** `internal/warning/AGENTS.md` (document the package per `internal/AGENTS.md` pattern)
- **New (1 file):** `cmd/legacy_test_helpers_test.go` (legacy test-only adapter for non-pilot sections per task 2.4.2a; deleted in slice 7)

### CHANGELOG entry

- **BREAKING (exit codes):** the seven legacy exit codes collapse to three (`0, 1, 2`). The four removed codes — `ExitCodeConfig` (3), `ExitCodeGit` (4), `ExitCodeValidation` (5), `ExitCodeNotFound` (6) — all map to `1` (`ExitCodeError`). Scripts that dispatched on `3`, `4`, `5`, or `6` must adapt. Semantic meaning is preserved in the typed error (`output.FormatError` dispatches via `errors.As`); consumers should parse stderr rather than exit code for semantic dispatch. A one-time deprecation warning is emitted on stderr (gated on `EXIT_CODE_LEGACY` env var or stderr being a TTY) per design Decisions 6 and 8.

### Affected systems

- **CLI behavior:** exit codes 3–6 collapse to 1 (BREAKING; mitigated by CHANGELOG entry)
- **CLI behavior:** `--output` flag added to `library resources` (additive)
- **Scripts / CI:** any external consumer reading exit codes 3–6 must adapt

## Risks

- **`LegacyBridge` couples this change to `cmd/container.go`** — sloppy migration of any non-pilot command could break `LegacyBridge.Services`. **Mitigation:** (a) `LegacyBridge.Services` is constructed in `main.go` by calling each underlying service constructor directly per design Decision 7 (no indirection through the deleted `cmd/container.go`); (b) the smoke-test in task 2.7.6 exercises every command end-to-end through `LegacyBridge`; (c) changes 3-6 each re-verify `LegacyBridge` after their respective migrations.
- **Exit code canary emission** — the deprecation warning needs the full `IOStreams` (not just the logger) accessible from `main.go` so the helper can check `EXIT_CODE_LEGACY` and TTY state. **Mitigation:** task 2.1.4 calls `warning.MaybeWarnLegacyExitCode(f.IOStreams)` (the helper takes the full `IOStreams`, per the signature established in task 2.1.6); the canary is emitted from `main.go`, not `cmdutil.ExitCodeFor`, per design Decisions 6 and 8 — keeping `ExitCodeFor` a pure function with no logger parameter and no side effects.
- **First time wiring the Factory** — `main.go` becomes the only composition root, which means every lazy function field must be populated correctly or commands fail at runtime. **Mitigation:** task 2.1.2 uses `cmdutil.OnceValuesFunc[T]` wrappers (per `internal/cmdutil/AGENTS.md`) for every lazy field; task 2.7.6 smoke-tests every command end-to-end; the pilot commands (adapt, library resources) provide positive verification that the wiring works.
- **Tests that depend on `cmd/container.go`** — the `internal/service/*_mock_test.go` mocks and `cmd/cmd_test.go` sections for non-pilot commands still call `NewServiceContainer()`. **Mitigation:** those tests are NOT touched in this change; only the adapt and resources sections are rewritten. The mocks are deleted in change-7.

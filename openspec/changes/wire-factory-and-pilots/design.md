# Design — Wire Factory and migrate pilot commands

## Context

After change-1 (`scaffold-cli-foundation`) lands the new packages and lint rules, the codebase still uses the legacy `cmd/container.go` (eager `ServiceContainer`), `cmd/command_config.go` (mutable shared `CommandConfig`), and `cmd/error_handler.go` (seven exit codes + `CategorizeError`). All existing commands use the legacy `NewXCommand(cfg *CommandConfig)` constructor and mutate the shared config.

This change (change-2 of 9) wires the new architecture into `main.go` and migrates the two smallest commands (`adapt` + `library resources`) as proof that the pattern works. The `legacyBridge` shim keeps non-migrated commands working until they're each migrated in subsequent changes.

## Goals / Non-Goals

**Goals:**

- `main.go` constructs `IOStreams` + `Factory` and populates all lazy function fields.
- The post-`Execute` error path uses `output.FormatError` + `cmdutil.ExitCodeFor`.
- Exit codes collapse from 0–6 to 0/1/2.
- `cmd/adapt.go` and `cmd/library/resources.go` are migrated to `NewCmdXxx(f, runF)` + `runXxx(opts)` per the `application/command-options-pattern` capability.
- All non-migrated commands continue to work via the `legacyBridge` shim.
- `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go` are deleted.
- Every command's `--help` produces byte-identical output to the pre-change build.

**Non-Goals:**

- Migrating `validate`, `canonicalize` — change-3.
- Migrating any library command other than `resources` — changes 4, 6, 7.
- Migrating `init` (which has unique partial-success semantics) — change-5.
- Migrating `config init`, `config validate` — change-8.
- Migrating `completion`, `version` — change-9.
- Deleting `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go` — change-7 (after `legacyBridge` is removed by the last command migration).
- Adding `--output` flag to commands other than `library resources` (the `add`, `init`, `refresh`, `remove`, `validate` library commands get `--output` in their own changes).

## Decisions

### 1. `legacyBridge` is the only caller of legacy types during the migration window

**Choice**: A single `legacyBridge` struct is constructed in `main.go` and passed to all non-migrated commands. The struct holds:
- `Services *ServiceContainer` (from `cmd/container.go` — still alive)
- `ErrorFormatter *ErrorFormatter` (from `cmd/error_formatter.go` — still alive)
- `Verbosity Verbosity` (from `cmd/verbose.go` — still alive)

**Rationale**: without `legacyBridge`, every non-migrated command would need its own shim, multiplying the touch surface. With a single shim, the dependency is centralized and the deletion in change-7 is mechanical.

**Alternatives considered**:

- Migrate every command in one change → defeats the purpose of the split.
- Keep `ServiceContainer` alongside `Factory` → defeats the purpose of the foundation.
- Construct `legacyBridge` lazily on each non-migrated command's invocation → unnecessary; one construction in `main.go` is fine.

### 2. Pilots are `adapt` + `library resources`

**Choice**: The two pilots are `adapt` (no library dependency, exercises `Factory.Transformer`) and `library resources` (exercises `Factory.Library` + the new `--output` flag).

**Rationale**: together they exercise the three most important Factory lazy fields (`Transformer`, `Library`, and the output formatting), validate the `NewCmdXxx(f, runF)` pattern, and produce visible output that can be smoke-tested. Both are small (~50–100 LOC) so the migration is low-risk.

**Alternatives considered**:

- Just `adapt` → only exercises `Transformer`; doesn't validate `--output` or library wiring.
- Just `library resources` → only exercises library; doesn't validate the `Transformer` path.
- All three core domain commands (`adapt`, `validate`, `canonicalize`) → triples the size of this change; deferred to change-3.

### 3. `--output` is opt-in per command, not global

**Choice**: `cmdutil.AddOutputFlags` is called explicitly by `library resources` (and future commands in changes 4, 6, 7). Commands that don't produce structured output (`adapt`, `validate`, `canonicalize`, `init`, `version`, `completion`, `config init`, `library create`) do NOT call `AddOutputFlags`.

**Rationale**: matches the `output-formats` capability; avoids polluting `--help` for commands that don't benefit.

### 4. Exit code collapse is in `main.go`, not in `cmdutil`

**Choice**: `main.go` is the only place that calls `cmdutil.ExitCodeFor` and translates the result to an `os.Exit(int(...))` call. The `cmdutil` package exposes only `ExitCodeFor` (returns `ExitCode` enum).

**Rationale**: matches the foundation's design (Decision #4 in `scaffold-cli-foundation/design.md`); the `os.Exit` call is centralized; the `forbidigo` rule from change-1 ensures `os.Exit` only appears in `main.go`.

### 5. Signal-aware context via `signal.NotifyContext`

**Choice**: `main.go` calls `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` and stores the resulting context as `Factory.RootContext` AND passes it to `rootCmd.SetContext(ctx)`.

**Rationale**: `c.Context()` returns the context Cobra was invoked with, so `opts.Ctx = c.Context()` in each command's `RunE` automatically picks up the signal-aware context. Cancellation propagates through all Factory lazy functions and command logic.

### 6. Exit code deprecation canary

**Choice**: When `EXIT_CODE_LEGACY` env var is set OR stderr is a TTY, `cmdutil.ExitCodeFor` emits a single `Logger.Warn(...)` to `opts.IO.Logger` on its first call per process. The warning is suppressed in CI (where stderr is typically not a TTY and `EXIT_CODE_LEGACY` is unset).

**Rationale**: matches the foundation's `cli/exit-codes` ADDED requirement (canary mechanism); gives consumers a heads-up about the BREAKING change without spamming CI logs.

## Risks / Trade-offs

- **`legacyBridge` couples this change to `cmd/container.go`** — if `cmd/container.go` is deleted before all commands migrate, the bridge breaks. **Mitigation:** the file is deleted IN this change; `legacyBridge` populates `Services` by calling the remaining service constructors directly in `main.go` (not via `cmd/container.go`'s `NewServiceContainer`). The bridge lives as long as `cmd/verbose.go` and `cmd/error_formatter.go` live (until change-7).
- **First real test of the Factory pattern** — bugs in the pattern would surface here. **Mitigation:** the foundation already includes full Factory tests; the pilot commands are small and easy to verify; smoke tests in task 1.5.5 cover every command.
- **`exit code 5 → 1` mapping loses semantic info** — scripts that distinguished `ExitCodeValidation` (5) from `ExitCodeConfig` (3) can no longer do so. **Mitigation:** semantic info is preserved in the error message (formatted via `output.FormatError`); consumers should parse stderr, not exit codes, for semantic dispatch.
- **`legacyBridge` keeps `cmd/verbose.go` and `cmd/error_formatter.go` alive** — these files are not deleted in this change. **Mitigation:** they are deleted in change-7 along with `internal/service/` and `internal/application/`.

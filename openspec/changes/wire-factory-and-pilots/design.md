# Design — Wire Factory and migrate pilot commands

## Context

After change-1 (`scaffold-cli-foundation`) lands the new packages and lint rules, the codebase still uses the legacy `cmd/container.go` (eager `ServiceContainer`), `cmd/command_config.go` (mutable shared `CommandConfig`), and `cmd/error_handler.go` (seven exit codes + `CategorizeError`). All existing commands use the legacy `NewXCommand(cfg *CommandConfig)` constructor and mutate the shared config.

This change (change-2 of 9) wires the new architecture into `main.go` and migrates the two smallest commands (`adapt` + `library resources`) as proof that the pattern works. The `LegacyBridge` shim keeps non-migrated commands working until they're each migrated in subsequent changes.

## Goals / Non-Goals

**Goals:**

- `main.go` constructs `IOStreams` + `Factory` and populates all lazy function fields.
- The post-`Execute` error path uses `output.FormatError` + `cmdutil.ExitCodeFor`.
- Exit codes collapse from 0–6 to 0/1/2.
- `cmd/adapt.go` and `cmd/library/resources.go` are migrated to `NewCmdXxx(f, runF)` + `runXxx(opts)` per the `application/command-options-pattern` capability.
- All non-migrated commands continue to work via the `LegacyBridge` shim.
- `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go` are deleted.
- Every command's `--help` produces byte-identical output to the pre-change build.

**Non-Goals:**

- Migrating `validate`, `canonicalize` — change-3.
- Migrating any library command other than `resources` — changes 4, 6, 7.
- Migrating `init` (which has unique partial-success semantics) — change-5.
- Migrating `config init`, `config validate` — change-8.
- Migrating `completion`, `version` — change-9.
- Deleting `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go` — change-7 (after `LegacyBridge` is removed by the last command migration).
- Adding `--output` flag to commands other than `library resources` (the `add`, `init`, `refresh`, `remove`, `validate` library commands get `--output` in their own changes).

## Decisions

### 1. `LegacyBridge` is the only caller of legacy types during the migration window

**Choice**: A single `cmd.LegacyBridge` struct (exported, declared in `cmd/legacy_bridge.go`) is constructed in `main.go` and passed to all non-migrated commands via the `cmd.NewRootCommand(f, bridge)` signature. The struct holds:
- `Services *LegacyServices` (populated by calling `application.New*` constructors directly in `main.go` after task 2.5.1 deletes `cmd/container.go` — see Decision 7)
- `ErrorFormatter *ErrorFormatter` (from `cmd/error_formatter.go` — still alive, deleted in slice 7)
- `Verbosity Verbosity` (from `cmd/verbose.go` — still alive, deleted in slice 7)

**Rationale**: without `LegacyBridge`, every non-migrated command would need its own shim, multiplying the touch surface. With a single shim, the dependency is centralized and the deletion in change-7 is mechanical. The type lives in the `cmd` package (exported) so it can cross the package boundary from `main.go`; see Decision 7 for the declare/populate split.

**Alternatives considered**:

- Migrate every command in one change → defeats the purpose of the split.
- Keep `ServiceContainer` alongside `Factory` → defeats the purpose of the foundation.
- Construct `LegacyBridge` lazily on each non-migrated command's invocation → unnecessary; one construction in `main.go` is fine.
- Declare `LegacyBridge` in `package main` (exported) → breaks the "package main is the composition root, not a types package" idiom.

### 2. Pilots are `adapt` + `library resources`

**Choice**: The two pilots are `adapt` (no library dependency, exercises `Factory.Transformer`) and `library resources` (exercises `Factory.Library` + the new `--output` flag).

**Rationale**: together they exercise the `Transformer` and `Library` lazy fields and the new `output.Exporter` formatting path, validate the `NewCmdXxx(f, runF)` pattern, and produce visible output that can be smoke-tested. Both are small (~50–100 LOC) so the migration is low-risk.

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

**Choice**: When `EXIT_CODE_LEGACY` env var is set OR stderr is a TTY, the deprecation warning is emitted via `f.IOStreams.Warnf(...)` exactly once per process. The warning is written to `io.ErrOut` (the user-facing stderr channel) — NOT to `io.Logger`, which is a developer debug channel gated on `GERMINATOR_DEBUG` and would be a no-op in production. `Warnf` is a new method on `IOStreams` (mirroring `Verbosef`) that wraps the message in `Styles.Warning` and appends a trailing newline. The gate rule is:

```go
if os.Getenv("EXIT_CODE_LEGACY") != "" || io.IsStderrTTY() { ... }
```

The warning SHALL be emitted at most once per process via `sync.Once`; a `ResetCanaryForTest()` helper resets the once-state for test isolation.

**Rationale**: matches the foundation's `cli/exit-codes` ADDED requirement (canary mechanism); gives consumers a heads-up about the BREAKING change without spamming CI logs.

### 7. `LegacyBridge.Services` is constructed in `main.go` from underlying service constructors (split into two tasks)

**Choice**: `LegacyBridge` is declared in the `cmd` package as an exported type (in `cmd/legacy_bridge.go`) so it can cross the package boundary from `main.go` to non-migrated commands via the `cmd.NewRootCommand(f, bridge)` signature. The struct holds:

- `Services *legacyServices` (set in task 2.1.3b, AFTER task 2.5.1 deletes `cmd/container.go`)
- `ErrorFormatter *ErrorFormatter` (from `cmd/error_formatter.go`, deleted in slice 7)
- `Verbosity Verbosity` (from `cmd/verbose.go`, deleted in slice 7)

Task 2.1.3a (BEFORE task 2.5.1) declares the type and the `cmd.NewRootCommand(f, bridge)` signature; `bridge` may have a nil `Services` field for now. Task 2.1.3b (AFTER task 2.5.1) populates `bridge.Services` in `main.go` by calling each underlying service constructor directly (`application.NewTransformer`, `application.NewValidator`, `application.NewCanonicalizer`, `application.NewInitializer`). The `main.go` import of `internal/application` is temporary and removed in slice 7.

**Rationale**: Without this split, the temporal ordering is muddled — task 2.1.3 (early) cannot reference `application.NewTransformer` while `cmd/container.go` still exists. Splitting "declare" from "populate" makes the dependency on `cmd/container.go`'s deletion explicit.

**Alternatives considered**:

- Keep `cmd/container.go` alive as the `LegacyBridge.Services` factory → contradicts task 2.5.1's deletion; the file must go.
- Construct `LegacyBridge` lazily on each non-migrated command's invocation → duplicates service construction across commands; loses the `main.go` is the only composition root invariant.
- Move `LegacyBridge` into a separate package (e.g., `internal/legacybridge`) → adds a package boundary for code that exists only during the migration window; not worth the ceremony.
- Declare `LegacyBridge` in `package main` (exported) → breaks the "package main is the composition root, not a types package" idiom.

### 8. Exit-code canary emission lives in `main.go`, not in `cmdutil.ExitCodeFor`

**Choice**: `cmdutil.ExitCodeFor(err error) ExitCode` remains a pure function (no logger parameter, no side effects). The canary warning is emitted from `main.go` (the imperative shell) immediately before `os.Exit(int(...))`, guarded by a package-level `sync.Once` flag in a small new helper (e.g., `internal/warning/canary.go` exposing `MaybeWarnLegacyExitCode(logger *slog.Logger)`). Each function does one thing: `ExitCodeFor` maps errors to codes; the canary helper emits the deprecation warning.

**Rationale**: matches the foundation's `ExitCodeFor` signature (pure, single-responsibility); complies with the `golang-cli-architecture` skill's principles (4: pure exit-code mapping; 10: side effects are imperative); preserves testability (`exit_test.go` doesn't need a logger to assert the mapping). Allocating the side effect to `main.go` follows the Functional Core / Imperative Shell split. The emission itself uses `io.Warnf(...)` (the `IOStreams` method) so the message is styled and routed to `io.ErrOut` — the user-facing channel — rather than `io.Logger`, which is a developer debug channel gated on `GERMINATOR_DEBUG`.

**Alternatives considered**:

- Add a `*slog.Logger` parameter to `ExitCodeFor` → deviates from the pure-function pattern; would require updating foundation tests in `internal/cmdutil/exit_test.go`; bloats every call site.
- Emit the warning from `output.FormatError` → mixes exit-code concerns into error formatting; couples two unrelated packages.
- Emit the warning from a Cobra `PersistentPostRunE` hook → runs even on success paths, which doesn't match the "first call to `ExitCodeFor`" semantics.

## Risks / Trade-offs

- **`LegacyBridge` couples this change to `cmd/container.go`** — if `cmd/container.go` is deleted before all commands migrate, the bridge breaks. **Mitigation:** per Decision 7, the task is split into 2.1.3a (declare `cmd.LegacyBridge` type, nil `Services`) and 2.1.3b (populate `Services` from `application.New*` constructors, runs after task 2.5.1 deletes `cmd/container.go`). The bridge lives as long as `cmd/verbose.go` and `cmd/error_formatter.go` live (until change-7).
- **First real test of the Factory pattern** — bugs in the pattern would surface here. **Mitigation:** the foundation already includes full Factory tests; the pilot commands are small and easy to verify; smoke tests in task 2.7.6 cover every command.
- **Exit codes `3, 4, 5, 6 → 1` collapse loses semantic info** — scripts that distinguished `ExitCodeConfig` (3), `ExitCodeGit` (4), `ExitCodeValidation` (5), or `ExitCodeNotFound` (6) from `ExitCodeError` (1) can no longer do so via exit code alone. **Mitigation:** semantic info is preserved in the error message (formatted via `output.FormatError`); consumers should parse stderr, not exit codes, for semantic dispatch. The canary warning (Decision 8) gives consumers a one-time heads-up that the codes changed.
- **`LegacyBridge` keeps `cmd/verbose.go` and `cmd/error_formatter.go` alive** — these files are not deleted in this change. **Mitigation:** they are deleted in change-7 along with `internal/service/` and `internal/application/`.

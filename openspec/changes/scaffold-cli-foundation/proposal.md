# Scaffold golang-cli-architecture foundation

## Why

The full migration to `golang-cli-architecture` (Functional Core / Imperative Shell, lazy `Factory`, per-command `Options` + `runF`, `IOStreams`, three exit codes) is too dense for the `osx-workflow` autonomous orchestrator to run as a single change (213 tasks, 17 specs, ~3k lines of artifact content). PHASE1 would exhaust its 30-min subprocess timeout and 10-iteration budget before making meaningful progress.

This change is **slice 0 of 9**: it lands only the foundation — three new packages, the `domain → core` rename, the `infrastructure/` flattening, and the supporting lint rules — so subsequent changes (2–9) can build incrementally without each one re-doing the mechanical work. No commands are migrated in this change; `main.go` is not rewired; legacy files are not deleted.

The full migration is described in `openspec/changes/scaffold-cli-foundation/design.md` (Decisions section); this change implements only the items below.

## What Changes

### New packages (zero behavior change yet)

- **NEW** `internal/iostreams/` — `IOStreams` struct (`In`, `Out`, `ErrOut`, `Verbosef`, `Logger`, `Styles`, `IsStdoutTTY`, `IsInteractive`, `SetStdoutTTY`), `System()` + `Test()` constructors
- **NEW** `internal/iostreams/styles.go` — `Styles` struct (Error, Success, Warning, Dim, Bold) using `github.com/charmbracelet/lipgloss`; respects `NO_COLOR`
- **NEW** `internal/output/` — `FormatError`, `Exporter` interface + `JSONExporter` + `TableExporter`, `AddOutputFlags` helper
- **NEW** `internal/cmdutil/` — `Factory` struct (lazy `func() (T, error)` fields, `sync.OnceValues` caching), `ExitCode` type + `ExitCodeFor(err)` mapping, `AddOutputFlags` re-export

### Domain rename

- **RENAME** `internal/domain/` → `internal/core/` (12 files; pure layer preserved; depguard rule path updated to `internal/core/**`)
- **NEW** `internal/core/rules.go` — pure business rules: `ValidatePlatform`, `ResolveOutputPath`
- **DEFERRED** `CanInstallResource` — defined in change-6 alongside its first consumer (`library add`)

### Flatten

- **FLATTEN** `internal/infrastructure/parsing/` → `internal/parser/`
- **FLATTEN** `internal/infrastructure/serialization/` → `internal/renderer/`
- **FLATTEN** `internal/infrastructure/adapters/claude-code/` → `internal/claude-code/`
- **FLATTEN** `internal/infrastructure/adapters/opencode/` → `internal/opencode/`
- **FLATTEN** `internal/infrastructure/config/` → `internal/config/`
- **FLATTEN** `internal/infrastructure/library/` → `internal/library/`
- **REMOVE** empty `internal/infrastructure/` directory tree

### Error type

- **NEW** `core.PartialSuccessError` (in `internal/core/errors.go`) — sentinel error type for partial-success flows (used by `init` in change-5; defined here because `cmdutil.ExitCodeFor` and `output.FormatError` need to recognize it for testability)

### Lint enforcement

- **ENABLE** `forbidigo` in `.golangci.yml` with five patterns: `fmt.Fprintf(os.Stdout|Stderr)` in `cmd/*.go` (excluding tests), `os.Exit(` in `cmd/**` (excluding `main.go`), `var global(Factory|CommandConfig)`, `SetGlobal(Factory|CommandConfig)`, and `context.Background()` in `cmd/**/*.go` (except `main.go`) — see design.md Decision 5
- **ADD** `nolintlint` with `require-explanation: true, require-specific: true`
- **ADD** `cmd/lint_test.go` that runs `mise run lint` and asserts no forbidden patterns slipped through

### Dependencies

- **ADD** `github.com/charmbracelet/lipgloss` (for `iostreams.Styles`)
- **ADD** `golang.org/x/term` (for TTY detection in `iostreams.System()`)
- **ADD** direct import of `github.com/spf13/pflag` (for typed error detection in `cmdutil.ExitCodeFor`)

## Capabilities

### New

- **`cli/cli-factory`** — The `cmdutil.Factory` pattern with lazy `func() (T, error)` fields; the only DI mechanism in the new architecture. Replaces the eager `ServiceContainer` + mutable `CommandConfig`.
- **`cli/iostreams`** — Centralized terminal I/O abstraction (`In`, `Out`, `ErrOut`, TTY detection, color/styles, `Verbosef`, `Logger`, `SetStdoutTTY`, `Test()` constructor).
- **`cli/output-formats`** — The `--output json|table|plain` flag and `Exporter` interface shared by all read-only commands. Built-in `JSONExporter` and `TableExporter`.
- **`application/command-options-pattern`** — The `NewCmdXxx(f *Factory, runF func(*XxxOptions) error)` + `runXxx(opts)` template.

### Modified (delta specs)

- **`application/dependency-injection`** — `ServiceContainer` is **removed**; replaced conceptually by `cmdutil.Factory` (the Factory itself is introduced in this change; the removal of all callers happens in change-7). `NewServiceContainer()` is **removed** in change-7.
- **`cli/exit-codes`** — The seven exit codes (`0`–`6`) collapse to three (`0, 1, 2`). `ExitCodeFor(err)` lives in `cmdutil` and uses `errors.As` to map typed errors. The `Category*` enum and `CategorizeError` function are **removed** in change-7.
- **`cli/framework`** — `CommandConfig` struct is **removed** in change-7; commands take `*cmdutil.Factory`. `RunE` populates an `Options` struct then delegates to `runXxx`. `runF` parameter enables test injection.
- **`cli/verbose-output`** — The `Verbosity` type and `VerbosePrint` helpers are **removed** in change-7; replaced by `opts.IO.Verbosef(format, args...)` on `IOStreams`. `-v`/`-vv` flag semantics preserved.
- **`cli/error-formatting`** — `ErrorFormatter` struct is **removed** in change-7; replaced by `output.FormatError(io, err)`. Type-specific formatters become private functions in `output/`.

> **Note on namespacing:** Capability names in this proposal use a domain prefix (`cli/` or `application/`) for clarity, but the delta specs in this change are stored in flat folders under `specs/<name>/spec.md` (matching the convention used by archived changes such as `2026-02-28-cli-infrastructure` and `2026-03-02-di-foundation`).

## Out of scope (deferred to subsequent changes)

- Migrating any command to the new pattern — **change-2** (`adapt` + `library resources` pilots)
- Wiring `main.go` to use `Factory` directly (without `legacyBridge`) — **change-2**
- Deleting `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go` — **change-2**
- Migrating `validate`, `canonicalize` — **change-3**
- Migrating `library presets`, `library show` — **change-4**
- Migrating `init` command + wiring `PartialSuccessError` into `ExitCodeFor`/`FormatError` — **change-5**
- Migrating `library add`, `library create` — **change-6**
- Migrating remaining library commands + deleting `internal/service/`, `internal/application/`, `legacyBridge` — **change-7**
- Migrating `config init`, `config validate` — **change-8**
- Migrating `completion`, `version`, deleting `internal/models/`, finalizing `AGENTS.md` + CHANGELOG — **change-9**

## Impact

### Affected code (this change)

- **New (≈10 files):** `internal/iostreams/{iostreams,styles,iostreams_test,styles_test}.go`, `internal/output/{errors,exporter,output_flags,output_test}.go`, `internal/cmdutil/{factory,exit,output_flags,factory_test,exit_test}.go`, `internal/core/rules.go`
- **Renamed (12 files):** `internal/domain/*.go` → `internal/core/*.go`
- **Moved (≈15 files):** `internal/infrastructure/parsing/*.go` → `internal/parser/*.go`; `internal/infrastructure/serialization/*.go` → `internal/renderer/*.go`; `internal/infrastructure/adapters/{claude-code,opencode}/*.go` → `internal/claude-code/*.go`, `internal/opencode/*.go`; `internal/infrastructure/config/*.go` → `internal/config/*.go`; `internal/infrastructure/library/*.go` → `internal/library/*.go`
- **Modified (1 file):** `internal/core/errors.go` (add `PartialSuccessError`)
- **Modified (1 file):** `.golangci.yml` (depguard rule path + forbidigo + nolintlint)
- **Modified (1 file):** `go.mod` (add lipgloss, term; pflag becomes direct)
- **New (1 file):** `cmd/lint_test.go`

### Affected dependencies

- **ADD:** `github.com/charmbracelet/lipgloss`, `golang.org/x/term`, direct `github.com/spf13/pflag`
- **DEFERRED:** `github.com/charmbracelet/huh` — added in a future change when the first interactive prompt is introduced (no current consumer; per design Open Question #1 resolution)

### Affected systems

- **None externally observable in this change.** No CLI flag changes, no exit code changes, no behavior changes. All changes are internal package restructuring + new testable units.
- **`internal/core/` rename**: every importer of `gitlab.com/amoconst/germinator/internal/domain` is updated to `gitlab.com/amoconst/germinator/internal/core`. A `type Domain = core` alias is added in `internal/core/doc.go` for any external consumer; removed in change-9.

## Risks

- **Large mechanical rename** — 12 files moved + ~15 files flattened + every import updated. **Mitigation:** `mise run check` runs after each major group of moves (tasks 1.2.4, 1.2a.7) so the tree stays green between groups; `gofmt -r` patterns can fix import aliases en masse.
- **Depguard rule update** — the rename to `internal/core/**` requires updating both the depguard file path glob and `wrapcheck.ignorePackageSig`. **Mitigation:** updated in single task 1.2.3; CI catches regressions.
- **Empty `infrastructure/` directory** — after flattening, the parent must be removed cleanly. **Mitigation:** task 1.2a.8 explicitly removes the tree; checked with `find internal/infrastructure -type f` returning nothing.
- **`PartialSuccessError` defined before any consumer** — the type is added in change-1 but only used in change-5 (`init`). **Mitigation:** the type has tests in `core/errors_test.go` that confirm it implements `error` and that `errors.As` works; it is not referenced from any command code yet, so unused-import lints won't fire.
- **Forbidigo false positives** — patterns are scoped to `cmd/**` (excluding `main.go` and tests). **Mitigation:** the patterns are reviewed against existing `cmd/*.go` content; if any command currently uses `fmt.Fprintf(os.Stdout)`, the command must be migrated to `opts.IO.Out` as part of its slice in changes 2-9. Until then, forbidigo only fires on NEW code (it doesn't flag pre-existing code at lint time).

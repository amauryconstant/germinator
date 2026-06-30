# Internal Package Patterns

**Location**: `internal/`
**Parent**: See `/AGENTS.md` for project overview

---

> Describes the target architecture post-rewrite. Slice 1 (scaffold-cli-foundation) landed: `internal/domain/` → `internal/core/` rename, `internal/infrastructure/` flattened, and the three new shell packages (`iostreams/`, `output/`, `cmdutil/`) introduced as units. Remaining work: wire `main.go` to `Factory`, migrate the eight command groups, and delete the legacy `application/`/`service/`/legacyBridge files (slices 2–7).

## Structure

The internal package tree is organized as a Functional Core surrounded by an Imperative Shell:

```
internal/
├── core/                ← Functional Core: pure logic, zero I/O
├── iostreams/           ← IOStreams abstraction (stdin/stdout/stderr, TTY, verbose)
├── output/              ← Shared output (FormatError, Exporter+AddJSONFlags, TablePrinter, prompts)
├── cmdutil/             ← Factory (lazy DI), ExitCode mapping, shared cmd helpers
├── config/              ← Config loading (koanf, XDG paths)
├── library/             ← Library I/O (loader, resolver, saver, validator, refresher)
├── claude-code/         ← Claude Code platform adapter (parse + render)
└── opencode/            ← OpenCode platform adapter (parse + render)
```

### Functional Core (`internal/core/`)

Pure computation with no I/O. Types with behavior, validation, business rules, decision logic. Depends on nothing except stdlib and `samber/lo`. Tested with values in, values out — no mocks.

- `core/types.go` — domain types (Agent, Command, Skill, Memory, Platform)
- `core/errors.go` — typed domain errors (ParseError, ValidationError, TransformError, FileError, ConfigError, NotFoundError, OperationError, InitializeError, PartialSuccessError) carrying semantic meaning only — no exit codes
- `core/validation.go` — generic validators with `Validator[T]` and `Pipeline[T]`
- `core/rules.go` — business rule functions spanning types/config
- `core/result.go` — `Result[T]` for composable error handling

**Core dependency policy** (enforced via `depguard`):
- stdlib (excluding I/O packages: no `os`, `net`, `exec`)
- `github.com/samber/lo`

### Imperative Shell

Everything that does I/O lives here.

#### `internal/iostreams/`
- `IOStreams` struct: `In`, `Out`, `ErrOut`, `Verbose`, `Logger`, `Styles`, TTY detection
- `System()` constructor (real I/O), `Test()` constructor (buffer-backed for tests)
- `IsStdoutTTY()`, `IsInteractive()`, `Verbosef()` methods

#### `internal/output/`
- `FormatError(io, err)` — dispatches on error type via `errors.As`, formats to stderr
- `Exporter` interface + `AddJSONFlags(cmd, &opts.Exporter, fields)` — wires `--json` flag to commands
- `TablePrinter` — multi-format table rendering
- `promptConfirm(io, msg)` — huh-based interactive confirmation

#### `internal/cmdutil/`
- `Factory` struct: `IOStreams`, `AppVersion`, `Executable`, plus **lazy function fields** for dependencies (`Config func() (*config.Config, error)`, etc.) with `sync.Once` caching
- `ExitCodeFor(err)` — maps errors to 0/1/2

#### `internal/config/`
- `AppConfig` struct with `toml`/`koanf` tags
- `Load()`, `DefaultConfig()`, XDG path resolution
- Missing file falls back to defaults (not an error); validation uses `ValidateAll()` collect mode

#### `internal/library/`
- `Library` struct with `Resources`, `Presets`
- Operations: `Load`, `Resolve`, `List`, `Add`, `Create`, `Remove`, `Refresh`, `Validate`, `Save`
- Path resolution: `--library` flag > `GERMINATOR_LIBRARY` env > XDG default
- Returns core types; no business logic

#### `internal/{claude-code,opencode}/`
- One package per platform
- Each provides: `ParsePlatformDocument(path, docType) (*core.Document, error)` and `RenderDocument(doc, docType) (string, error)`
- Platform-specific validation (e.g. OpenCode mode/temperature rules)
- Returns core types; depends on `internal/core/`

### Removed Packages

The following legacy packages are deleted by the rewrite:

| Old | Replaced by |
|---|---|
| `internal/application/` | Interfaces defined where consumed (in cmd files or in `internal/core/contracts.go`) |
| `internal/service/` | Service implementations live in cmd files or as private helpers in platform adapter packages |
| `internal/infrastructure/parsing/` | Merged into `internal/{claude-code,opencode}/` (parse functions) |
| `internal/infrastructure/serialization/` | Merged into `internal/{claude-code,opencode}/` (render functions) |
| `internal/infrastructure/adapters/{claude-code,opencode}/` | Renamed to `internal/{claude-code,opencode}/` |
| `internal/infrastructure/config/` | Renamed to `internal/config/` |
| `internal/infrastructure/library/` | Renamed to `internal/library/` |
| `internal/domain/` | Renamed to `internal/core/` |

---

# Package Dependency Rules

```
cmd/ ───────────► internal/output/
  │                    │
  │                    ▼
  ├───────────► internal/iostreams/
  │
  ├───────────► internal/core/ ◄────── (no outbound deps)
  │                    ▲
  ├───────────► internal/claude-code/ ─────► internal/core/
  │                    ▲
  ├───────────► internal/opencode/   ─────► internal/core/
  │
  ├───────────► internal/config/    ─────► internal/core/  (optional)
  │
  └───────────► internal/library/   ─────► internal/core/
```

- `cmd/` imports everything (composition root)
- `internal/core/` imports nothing (stdlib + `lo` only)
- Adapter packages (`claude-code/`, `opencode/`) depend on `core/` and return core types
- `internal/output/` imports `core/` (formats core errors)
- `internal/library/`, `internal/config/` are independent (or import `core/` for shared types)

**Enforced by `depguard`** in `.golangci.yml`:
```yaml
linters-settings:
  depguard:
    rules:
      core-isolation:
        files:
          - "**/core/**"
        allow:
          - $gostd
          - github.com/samber/lo
        deny:
          - pkg: "github.com/*"
            desc: "core allows only stdlib and lo"
```

---

# Testing Patterns

Table-driven tests with descriptive names. End-to-end: `LoadDocument → Validate → RenderDocument → Verify output`. See `test/AGENTS.md` for golden file testing patterns and the CLI testing pyramid (Core → Command → Integration → E2E).

### When to Mock vs. Use Real Implementations

| Scenario | Strategy |
|---|---|
| `internal/core/` logic | Table-driven, no mocks — pure values in/out |
| `cmd/<file>` logic | `runF` injection + `iostreams.Test()` buffers |
| `internal/claude-code/` / `internal/opencode/` | Use real templates + `t.TempDir()` fixtures |
| E2E behavior | Full binary via `testscript` / Ginkgo |

The `test/mocks/` package is **deprecated**. New tests use `runF` injection with `iostreams.Test()`.

---

# File Organization

## Test Files

- `<package>_test.go` — unit tests
- `integration_test.go` — integration tests (build tag `//go:build integration`)
- `<package>_golden_test.go` — golden file tests (build tag `//go:build golden`)

## Source Files

- `<package>.go` — main implementation
- `doc.go` — package documentation

---

# Migration Status

| Old package | New package | Status |
|---|---|---|
| `internal/domain/` | `internal/core/` | **Done** (slice 1; `type Domain = core` alias retained) |
| `internal/infrastructure/parsing/` | `internal/parser/` | **Done** (slice 1; not yet merged into adapters — kept separate) |
| `internal/infrastructure/serialization/` | `internal/renderer/` | **Done** (slice 1; not yet merged into adapters — kept separate) |
| `internal/infrastructure/config/` | `internal/config/` | **Done** (slice 1) |
| `internal/infrastructure/library/` | `internal/library/` | **Done** (slice 1) |
| `internal/infrastructure/adapters/claude-code/` | `internal/claude-code/` | **Done** (slice 1) |
| `internal/infrastructure/adapters/opencode/` | `internal/opencode/` | **Done** (slice 1) |
| `internal/infrastructure/` umbrella | (removed) | **Done** (slice 1) |
| New: `internal/iostreams/`, `internal/output/`, `internal/cmdutil/` | (shell) | **Done as units** (slice 1); consumed in slice 2+ |
| `internal/application/` | (removed) | **Pending** (slice 1 → slice 7) |
| `internal/service/` | (removed) | **Pending** (slice 7) |
| `cmd/{container,command_config,error_handler}.go` + `legacyBridge` | (removed) | **Pending** (slice 7) |
| 7 → 3 exit codes | `cmdutil.ExitCodeFor` | **In progress** (slice 1 unit + tests; wiring in slice 2) |

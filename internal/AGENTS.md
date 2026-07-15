# Internal Package Patterns

**Location**: `internal/`
**Parent**: See `/AGENTS.md` for project overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md` (Functional Core / Imperative Shell, three concerns, Factory pattern)

---

## Structure

The internal package tree is organized as a Functional Core surrounded by an Imperative Shell:

```
internal/
├── core/                ← Functional Core: pure logic, zero I/O
├── iostreams/           ← IOStreams abstraction (stdin/stdout/stderr, TTY, verbose)
├── output/              ← Shared output (FormatError, Exporter, AddOutputFlags)
├── cmdutil/             ← Factory (lazy DI), CompletionCache, ExitCode mapping, cmd helpers
├── config/              ← Config loading (koanf, XDG paths)
├── library/             ← Library I/O (loader, resolver, saver, validator, refresher)
├── parser/              ← Platform-agnostic document parsing (frontmatter + body)
├── renderer/            ← Template-driven rendering to platform output
├── claude-code/         ← Claude Code platform adapter
├── opencode/            ← OpenCode platform adapter
├── permission/          ← Permission-rule mapping for platform output
└── version/             ← Build-time version metadata (ldflags injection point)
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
- `Exporter` interface + `JSONExporter` (2-space indent) + `TableExporter` (`tab:"HEADER"` struct tag)
- `AddOutputFlags(cmd, *string)` — wires `--output` (`json`/`table`/`plain`) with shell completion
- `FormatResourcesList(lib)` — stable plain rendering of `library resources`

#### `internal/cmdutil/`
- `Factory` struct: `IOStreams`, `AppVersion`, `Executable`, `RootContext`, `CompletionCache`, plus **lazy function fields** for dependencies (`Config func() (*config.Config, error)`, `Library func() (*library.Library, error)`) with `sync.Once` caching
- `ExitCodeFor(err)` — maps errors to 0/1/2 (no 3–6)
- `CompletionCache` — per-Factory TTL cache for shell-completion library snapshots; `Invalidate()` called by mutating library commands

#### `internal/config/`
- `Config` struct with `koanf` tags (`Library`, `PlatformDefault`, `Debug`, `Completion`)
- `Load()` top-level wrapper, `DefaultConfig()`, XDG path resolution via `adrg/xdg`
- Koanf env provider merges `GERMINATOR_*` env vars (defaults → file → env)
- `Library: ""` is the canonical "no config-file override" signal; falls through to `library.DefaultLibraryPath()`
- Missing file falls back to defaults (not an error); validation uses `errors.Join` collect mode

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

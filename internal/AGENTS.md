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
├── transform/           ← `Transformer` I/O adapter (parse → render → write)
├── validate/            ← `Validator` I/O adapter (parse → core + platform validators)
├── canonicalize/        ← `Canonicalizer` I/O adapter (parse-platform-doc → marshal → write)
├── install/             ← `Initializer` I/O adapter (per-ref parse → render → write loop)
├── claude-code/         ← Claude Code platform adapter
├── opencode/            ← OpenCode platform adapter
├── permission/          ← Permission-rule mapping for platform output
├── paths/               ← Shared filesystem path helpers (tilde expansion); leaf shell package, no internal deps
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

Everything that does I/O lives here. Each shell package owns one cohesive concern — see per-package AGENTS.md (`internal/<x>/AGENTS.md`) for the full surface area. High-level responsibilities:

| Package | Concern |
|---|---|
| `iostreams` | Terminal I/O boundary, TTY detection, `Verbosef`, `Styles` |
| `output` | `FormatError`, `Exporter`/`JSONExporter`/`TableExporter`, `AddOutputFlags`, `FormatResourcesList` |
| `cmdutil` | `Factory` (lazy fn fields), `ExitCodeFor`, `CompletionCache`, cmd helpers |
| `config` | `Config` struct + `koanf` loading, XDG resolution via `adrg/xdg`, `GERMINATOR_*` env merge |
| `library` | Library I/O (load/resolve/list/add/create/remove/refresh/validate/save); path priority: flag > env > XDG |
| `parser`, `renderer` | Frontmatter parsing + template rendering (consumed by transform/validate/canonicalize/install) |
| `transform`, `validate`, `canonicalize`, `install` | Service-style I/O adapters over parser + renderer + library |
| `claude-code`, `opencode` | One per platform; parse + render + platform-specific validation; return core types |
| `permission` | Permission-rule mapping for platform output |
| `paths` | `ExpandHome` tilde-expansion; leaf package (stdlib only) |
| `version` | Build-time version metadata (ldflags injection point) |

---

# Package Dependency Rules

```
cmd/ ───────────► internal/output/  ─────► internal/core/
  │                    ├─────────────────► internal/config/
  │                    └─────────────────► internal/library/
  │                    │
  │                    ▼
  ├───────────► internal/iostreams/
  │
  ├───────────► internal/cmdutil/  ─────► internal/config/
  │                    └────────────► internal/library/
  │
  ├───────────► internal/core/ ◄────── (no outbound deps)
  │                    ▲
  ├───────────► internal/claude-code/ ─────► internal/core/
  │                    ▲
  ├───────────► internal/opencode/   ─────► internal/core/
  │                    ▲
  ├───────────► internal/config/    ─────────► internal/paths/
  │                    └─────────────────► internal/core/  (optional)
  │
  ├───────────► internal/library/   ─────► internal/core/
  │
  ├───────────► internal/parser/   ─────► internal/claude-code/  ─────► internal/core/
  │                    ├────────────► internal/core/
  │                    └────────────► internal/opencode/    ─────► internal/core/
  │
  ├───────────► internal/renderer/ ─────► internal/claude-code/  ─────► internal/core/
  │                    ├────────────► internal/opencode/    ─────► internal/core/
  │                    └────────────► internal/permission/
  │
  ├───────────► internal/transform/     ─────► internal/core/
  │                  ├────────────► internal/parser/   ─────► internal/core/
  │                  └────────────► internal/renderer/ ─────► internal/core/
  ├───────────► internal/validate/      ─────► internal/core/
  │                  └────────────► internal/parser/   ─────► internal/core/
  ├───────────► internal/canonicalize/  ─────► internal/core/
  │                  ├────────────► internal/parser/   ─────► internal/core/
  │                  └────────────► internal/renderer/ ─────► internal/core/
  └───────────► internal/install/       ─────► internal/core/
                     ├────────────► internal/library/  ─────► internal/core/
                     ├────────────► internal/parser/   ─────► internal/core/
                     └────────────► internal/renderer/ ─────► internal/core/
```

- `cmd/` imports everything (composition root)
- `internal/core/` imports nothing (stdlib + `lo` only); the one allowed self-import exception is `internal/core/opencode` for its OpenCode-specific validators (documented in `internal/core/AGENTS.md`)
- Adapter packages (`claude-code/`, `opencode/`) depend on `core/` and return core types
- `internal/parser/` and `internal/renderer/` both depend on the platform adapters (`claude-code/`, `opencode/`) for the parse and render paths; each declares its own narrow consumer-side interface (`platformAdapter`, `templateAdapter`) per the "interfaces where consumed" rule
- Service-style I/O adapters (`transform/`, `validate/`, `canonicalize/`, `install/`) depend on `core/` plus `parser/`, `renderer/`, and (for `install/`) `library/`
- `internal/install/` depends on `internal/library/` for resource resolution; `library/` does **not** depend on `install/` (one-way dependency, enforced by `go build ./...`)
- `internal/output/` imports `core/` (formats core errors), `config/` (renders `*config.WriteError`), and `library/` (`FormatResourcesList`)
- `internal/cmdutil/` imports `config/` (exit-code mapping for `*config.WriteError`) and `library/` (CompletionCache type)
- `internal/library/`, `internal/config/` are independent (or import `core/` for shared types)

**Enforced by `depguard`** in `.golangci.yml` — core allows stdlib + `samber/lo` only; see `.golangci.yml` for the full rule.

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

The `test/mocks/` package is deprecated. New tests use `runF` injection with `iostreams.Test()` buffers (see `cmd/AGENTS.md` → Testing).

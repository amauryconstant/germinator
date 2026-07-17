**Location**: `cmd/`
**Parent**: See `/AGENTS.md` for project overview
**Skill references**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`, `@.opencode/skills/golang-cli-architecture/references/08-completion.md`, `@.opencode/skills/golang-cli-architecture/references/13-versioning.md`

---

# CLI Entry Points

Cobra-based CLI built on the `NewCmdXxx(f *Factory, runF func(*XxxOptions) error) *cobra.Command` pattern (Functional Core / Imperative Shell). **Architectural patterns live here**; per-command flag tables and behavior live in [`cmd/commands/AGENTS.md`](commands/AGENTS.md).

## Files

| File | Purpose |
|------|---------|
| `root.go` | Root command with subcommand registration |
| `adapt.go` | `adapt <input> <output> --platform ...` — canonical `NewCmdAdapt` example (see below) |
| `validate.go` | Validate document against platform rules |
| `canonicalize.go` | Convert platform document to canonical format |
| `version.go` | Display version, commit, build date (reads `internal/version`) |
| `library.go` | Library parent command + `library create` routing shim |
| `library_init.go` | Library init subcommand (scaffolding new libraries) |
| `library_add.go` | Library add subcommand (import resources, discover orphans) |
| `library_create.go` | `library create preset` subcommand |
| `library_refresh.go` | Library refresh subcommand (sync metadata from files) |
| `library_remove.go` | Library remove subcommand (resource / preset) |
| `library_validate.go` | Library validate subcommand (integrity check + `--fix`) |
| `resources.go` / `presets.go` / `show.go` | Read-only library subcommands |
| `init.go` | Install resources from library to project (top-level `init`) |
| `completion.go` | Shell completion command (carapace-based, multi-shell) |
| `completions.go` | Dynamic completion actions with `Factory.CompletionCache` |
| `config.go` | Config command group (`init`, `validate`) |
| `config_init.go` | `config init` subcommand (scaffold config file) |
| `config_validate.go` | `config validate` subcommand (validate config file) |
| `lint_test.go` | Lint baseline enforcement test (see [Lint Enforcement](#lint-enforcement)) |
| `testdata/lint_baseline.txt` | Captured `mise run lint` output; the baseline against which `lint_test.go` diffs |

> The composition root (`main.go`) lives at the project root. See [`/AGENTS.md`](../AGENTS.md) for the entry-point layout.

---

# Dependency Injection

Composition flows through `*cmdutil.Factory` (`internal/cmdutil/factory.go`). `main.go` constructs a single `*Factory` and passes it to every `NewCmdXxx` constructor. The Factory exposes:

- `IOStreams` (`*iostreams.IOStreams`) — terminal I/O, `Verbosef`, TTY detection
- `RootContext` (`context.Context`) — propagated to every `RunE`
- Lazy function fields (`Config`, `Library`, `Executable`, `CompletionCache`, ...) backed by `sync.OnceValues`

No `init()` functions or global command variables.

---

# Canonical Command Pattern: `adapt`

`cmd/adapt.go` is the canonical reference for the `NewCmdXxx(f, runF) + runXxx(opts)` pattern. Every command follows the same shape.

## 1. Options struct

Holds resolved runtime state — Factory-derived values, parsed flags, and positional args. Lazy per-call injection seams (here `Transformer`) live here so tests can substitute a fake.

```go
type adaptOptions struct {
    IO          *iostreams.IOStreams
    Transformer func() (Transformer, error) // nil → production constructor
    Ctx         context.Context
    InputPath   string
    OutputPath  string
    Platform    string
}
```

## 2. Local contract (interfaces where consumed)

Command-side interfaces live in `cmd/`, next to the consumer — not in `internal/core/`. This keeps `core/` pure and lets each command narrow the surface to what it actually needs.

```go
type Transformer interface {
    Transform(ctx context.Context, req *transform.Request) (*core.TransformResult, error)
}

// Request type lives in internal/transform/ (slice-8 stage 3).
// The cmd-side Transformer interface imports it via the shared import.
```

## 3. Constructor: `NewCmdAdapt(f, runF) *cobra.Command`

- `f *cmdutil.Factory` — composition root; supplies `IOStreams`, `RootContext`, completion cache.
- `runF func(*adaptOptions) error` — test-injection seam. Production passes `nil` (the constructor falls back to `runAdapt`); tests pass a stub.
- Builds the `*cobra.Command`, wires flags + carapace completion, and constructs `*adaptOptions` inside `RunE`.

```go
func NewCmdAdapt(f *cmdutil.Factory, runF func(*adaptOptions) error) *cobra.Command {
    var platform string
    cmd := &cobra.Command{
        Use:   "adapt <input> <output>",
        Args:  cobra.ExactArgs(2),
        RunE: func(c *cobra.Command, args []string) error {
            opts := &adaptOptions{
                IO:         f.IOStreams,
                Ctx:        c.Context(),
                InputPath:  args[0],
                OutputPath: args[1],
                Platform:   platform,
            }
            if runF != nil {
                return runF(opts)
            }
            return runAdapt(opts)
        },
    }
    cmd.Flags().StringVar(&platform, "platform", "", "Target platform (required: claude-code, opencode)")
    _ = cmd.MarkFlagRequired("platform")
    carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
        "platform": actionPlatforms(f),
    })
    return cmd
}
```

## 4. `runAdapt(opts)` — production body

Pure function over `*adaptOptions`. Validates input (`core.ValidatePlatform`), resolves the lazy dependency with a nil-safe fallback, calls the contract, writes success to `opts.IO.Out`, and wraps every failure path with `fmt.Errorf("...: %w", err)` so `output.FormatError` + `cmdutil.ExitCodeFor` can dispatch by type. Full body in `cmd/adapt.go`.

---

# Required Flags

Both `adapt` and `validate` require `--platform`:
```go
_ = cmd.MarkFlagRequired("platform")
```

Runtime validation uses `core.ValidatePlatform` (returns `*core.ValidationError` for unknown platforms):
```go
if err := core.ValidatePlatform(opts.Platform); err != nil {
    return fmt.Errorf("validating platform: %w", err)
}
```

Use `core.PlatformClaudeCode` / `core.PlatformOpenCode` constants (defined in `internal/core/rules.go` alongside `ValidatePlatform`).

---

# Exit Codes

Three-value scheme:

- `0` (Success) — command completed; also returned for `*core.PartialSuccessError` with `Succeeded > 0`
- `1` (Error) — general errors (transform, file, unexpected)
- `2` (Usage) — Cobra argument/validation errors; full type → code mapping in [`internal/cmdutil/AGENTS.md`](../internal/cmdutil/AGENTS.md)

Mapping: `cmdutil.ExitCodeFor(err)` inspects the error chain with `errors.As` and returns the code. `main.go` calls it from a deferred handler — never call `os.Exit` directly (enforced by `forbidigo`).

---

# Error Handling

## Flow

Errors return from `RunE` → bubble up to `main.go` → `cmdutil.ExitCodeFor(err)` resolves the exit code → a deferred handler calls `output.FormatError(io, err)` (writes the formatted message to `ErrOut` via `errors.As` dispatch on `*core.*Error`) → `os.Exit(code)`.

```go
// inside a runXxx body
if err := core.ValidatePlatform(opts.Platform); err != nil {
    return fmt.Errorf("validating platform: %w", err)
}
```

Every error path wraps with `%w` so the typed `*core.*Error` stays inspectable through the chain.

## Typed Errors

Import from `internal/core/`:
```go
import "gitlab.com/amoconst/germinator/internal/core"

// Constructors
core.NewParseError(path, message, cause)
core.NewValidationError(message, field, suggestions)
core.NewTransformError(operation, platform, message, cause)
core.NewFileError(path, operation, message, cause)
core.NewConfigError(field, value, message).WithSuggestions([]string{...})
core.NewNotFoundError(resource, identifier)
core.NewOperationError(op, resource, cause)
core.NewInitializeError(ref, inputPath, outputPath, cause)
core.NewPartialSuccessError(succeeded, failed, errors)
```

`output.FormatError` renders each type with a contextual prefix (`Error: not found: <ref>`, `Error: <op> failed: <msg>`, etc.).

---

# Version Output

`fmt.Fprintf(opts.IO.Out, "germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)` — example: `germinator v0.3.20 (abc123def) 2026-02-04`. `internal/version` is the source of truth (populated via `-ldflags` at build time).

---

# Testing

Table-driven tests with descriptive names. Each command has a dedicated `*_test.go`. Two test seams:

- `runF` injection — tests pass a stub `runF func(*XxxOptions) error` to skip the real body and assert the constructor wired flags / options as the test specifies.
- `iostreams.Test()` — buffer-backed `IOStreams` so tests assert on `Out` / `ErrOut` contents.

Test files of note:
- `adapt_test.go`, `validate_test.go`, `version_test.go` — per-command tests
- `library_*_test.go` — library subcommand tests
- `completions_test.go` — completion action unit tests (`Factory.CompletionCache`, timeout, actions)
- `lint_test.go` — lint baseline enforcement (see [Lint Enforcement](#lint-enforcement))

> The `test/mocks/` package is **deprecated**. New tests use `runF` injection with `iostreams.Test()` buffers.

**Debug logging**: `IOStreams.SetDebug(bool)` driven by `cfg.Debug` from `main.go` after `config.Load()`. `GERMINATOR_DEBUG` env → koanf env provider → `cfg.Debug`. The env is **not** read directly by `iostreams.System()`; `main.go`'s fail-fast `BuildFactory` is the single source of truth.

---

# Lint Enforcement

`forbidigo` patterns enforced in `.golangci.yml` for `cmd/**` (excluding `main.go` and tests):

| Pattern | Why |
|---------|-----|
| `fmt.Fprintf(os.Stdout\|Stderr)` | Use `opts.IO.Out` / `opts.IO.ErrOut` instead |
| `os.Exit(` | Use `cmdutil.ExitCodeFor(err)` in a deferred handler |
| `var global(Factory\|CommandConfig)` | Composition must flow through `*Factory` |
| `SetGlobal(Factory\|CommandConfig)` | Same |
| `context.Background()` | Use `opts.IO.RootContext` (or `Factory.RootContext`) |

`nolintlint` requires both a reason and a specific linter name on every `//nolint:` directive.

## Lint Baseline Test

`cmd/lint_test.go` runs `mise run lint` (via `exec.Command`) **8 times**, unions the stable violations, and diffs against `cmd/testdata/lint_baseline.txt`. The test fails on any violation that is not in the baseline.

**Adding a new intentional violation**:

1. Make the change.
2. Run `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` to refresh the baseline.
3. Commit the baseline file alongside the change.

> The test re-runs lint multiple times because `golangci-lint`'s parallel linters report in non-deterministic order; the union of stable violations is what the test gates on.

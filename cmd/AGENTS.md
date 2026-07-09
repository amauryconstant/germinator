**Location**: `cmd/`
**Parent**: See `/AGENTS.md` for project overview
**Skill references**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`, `@.opencode/skills/golang-cli-architecture/references/08-completion.md`, `@.opencode/skills/golang-cli-architecture/references/13-versioning.md`

---

# CLI Entry Points

Cobra-based CLI built on the `NewCmdXxx(f *Factory, runF func(*XxxOptions) error) *cobra.Command` pattern (Functional Core / Imperative Shell). **Architectural patterns live here**; per-command flag tables and behavior live in [`cmd/commands/AGENTS.md`](commands/AGENTS.md).

## Files

| File | Purpose |
|------|---------|
| `main.go` | Composition root — constructs `*cmdutil.Factory`, registers commands, deferred exit-code handler |
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
| `lint_test.go` | Lint baseline enforcement test (see [Lint Enforcement](#lint-enforcement)) |
| `testdata/lint_baseline.txt` | Captured `mise run lint` output; the baseline against which `lint_test.go` diffs |

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
    Transform(ctx context.Context, req *TransformRequest) (*core.TransformResult, error)
}

type TransformRequest struct {
    InputPath  string
    OutputPath string
    Platform   string
}
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

Pure function over `*adaptOptions`. Validates input (`core.ValidatePlatform`), resolves the lazy dependency with a nil-safe fallback, calls the contract, writes success to `opts.IO.Out`, and wraps every failure path with `fmt.Errorf("...: %w", err)` so `output.FormatError` + `cmdutil.ExitCodeFor` can dispatch by type.

```go
func runAdapt(opts *adaptOptions) error {
    if err := core.ValidatePlatform(opts.Platform); err != nil {
        return fmt.Errorf("validating platform: %w", err)
    }
    opts.IO.Verbosef("transforming %s → %s", opts.InputPath, opts.OutputPath)

    resolve := opts.Transformer
    if resolve == nil {
        resolve = func() (Transformer, error) { return NewTransformer(), nil }
    }
    t, err := resolve()
    if err != nil {
        return fmt.Errorf("resolving transformer: %w", err)
    }
    if _, err := t.Transform(opts.Ctx, &TransformRequest{
        InputPath:  opts.InputPath,
        OutputPath: opts.OutputPath,
        Platform:   opts.Platform,
    }); err != nil {
        return fmt.Errorf("transforming document: %w", err)
    }
    _, _ = fmt.Fprintf(opts.IO.Out, "wrote %s\n", opts.OutputPath)
    return nil
}
```

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

## Supported Platforms

- `claude-code` - Claude Code document format
- `opencode` - OpenCode document format

---

# Verbosity

`-v` / `-vv` on the root command toggles `IOStreams.Verbose`. Commands call `opts.IO.Verbosef(format, args...)` (writes to `ErrOut`, auto-trailing newline, no-op when `Verbose == false`). Stdout stays clean for piping.

---

# Exit Codes

Three-value scheme:

- `0` (Success) — command completed; also returned for `*core.PartialSuccessError` with `Succeeded > 0`
- `1` (Error) — general errors (transform, file, unexpected)
- `2` (Usage) — Cobra argument/validation errors (invalid flags, missing args) detected via `cmdutil.cobraUsagePrefixes`

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

# Argument Count

`adapt` and `validate` use `cobra.ExactArgs(2)` and `cobra.ExactArgs(1)` respectively.
`root` and `version` use default (no arguments).

---

# Version Output Format

```go
fmt.Fprintf(opts.IO.Out, "germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)
```

Example: `germinator v0.3.20 (abc123def) 2026-02-04`

`internal/version` is the source of truth (populated via `-ldflags` at build time). `Factory.AppVersion` is a separate short-form string used elsewhere and is **not** read by `runVersion` (see `cmd/version.go`).

---

# Testing

Table-driven tests with descriptive names. Each command has a dedicated `*_test.go`. Two test seams:

- `runF` injection — tests pass a stub `runF func(*XxxOptions) error` to skip the real body and assert the constructor wired flags / options correctly.
- `iostreams.Test()` — buffer-backed `IOStreams` so tests assert on `Out` / `ErrOut` contents.

Test files of note:
- `adapt_test.go`, `validate_test.go`, `version_test.go` — per-command tests
- `library_*_test.go` — library subcommand tests
- `completions_test.go` — completion action unit tests (`Factory.CompletionCache`, timeout, actions)
- `lint_test.go` — lint baseline enforcement (see [Lint Enforcement](#lint-enforcement))

> The `test/mocks/` package is **deprecated**. New tests use `runF` injection with `iostreams.Test()` buffers.

---

# Foundation Units

The shell units consumed by `cmd/`:

| Unit | Package | Purpose |
|------|---------|---------|
| `iostreams.IOStreams` | `internal/iostreams/` | Single terminal I/O boundary; `System()` (real) and `Test()` (buffer) constructors; `Verbosef`, TTY detection, `Styles` |
| `iostreams.Styles` | `internal/iostreams/styles.go` | `Error`/`Success`/`Warning`/`Dim`/`Bold` via `lipgloss`; respects `NO_COLOR` and TTY |
| `output.FormatError` | `internal/output/errors.go` | Dispatches on typed `*core.*Error` via `errors.As`; writes to `io.ErrOut` |
| `output.Exporter` + `JSONExporter` + `TableExporter` | `internal/output/exporter.go` | Format-flexible output (`tab:"HEADER"` struct tag) |
| `output.AddOutputFlags` | `internal/output/output_flags.go` | Wires `--output` (`json`/`table`/`plain`) with completion |
| `cmdutil.Factory` | `internal/cmdutil/factory.go` | Lazy `func() (T, error)` fields with `sync.OnceValues` caching; `NewFactory(ctx, io, appVersion, executable)` |
| `cmdutil.ExitCodeFor` | `internal/cmdutil/exit.go` | Maps `error` → 0/1/2 (0 if `*core.PartialSuccessError{Succeeded>0}`) |
| `cmdutil.AddOutputFlags` | `internal/cmdutil/output_flags.go` | Re-export of `output.AddOutputFlags` so cmd files import only `cmdutil` |
| `core.ValidatePlatform` | `internal/core/rules.go` | Returns `*core.ValidationError` for unknown platform strings |
| `core.ResolveOutputPath` | `internal/core/rules.go` | `(docType, name, platform) → path` (e.g., `agents/reviewer.claude-code.md`) |

**`GERMINATOR_DEBUG=1`** enables a debug-level `slog.Logger` on `IOStreams.Logger` (writes to `ErrOut`). Unset = no-op handler.

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

**Location**: `cmd/`
**Parent**: See `/AGENTS.md` for project overview

---

# CLI Entry Points

Cobra-based CLI with platform-specific validation, typed errors, and verbosity control. **Architectural patterns live here**; per-command flag tables and behavior live in [`cmd/commands/AGENTS.md`](commands/AGENTS.md).

## Files

| File | Purpose |
|------|---------|
| `main.go` | Composition root - wires ServiceContainer and executes CLI |
| `container.go` | ServiceContainer for dependency injection (legacy; slice 7) |
| `root.go` | Root command with subcommand registration |
| `adapt.go` | Transform document to target platform format |
| `validate.go` | Validate document against platform rules |
| `canonicalize.go` | Convert platform document to canonical format |
| `version.go` | Display version, commit, build date |
| `library.go` | Library management commands (init, resources, presets, show) |
| `library_init.go` | Library init subcommand (scaffolding new libraries) |
| `library_add.go` | Library add subcommand (import resources, discover orphans) |
| `library_create.go` | Library create subcommand (create presets) |
| `library_refresh.go` | Library refresh subcommand (sync metadata from files) |
| `init.go` | Install resources from library to project |
| `completion.go` | Shell completion command (carapace-based, multi-shell) |
| `completions.go` | Dynamic completion actions with caching |
| `formatters.go` | Init command output formatting (dry-run, success) |
| `library_formatters.go` | Library command output formatting |
| `error_handler.go` | Error categorization and exit code handling (legacy; slice 7) |
| `error_formatter.go` | Typed error formatting with contextual hints (legacy; slice 7) |
| `verbose.go` | Verbosity levels and output helpers (legacy; slice 7) |
| `config.go` | Config command group (init, validate) and CommandConfig struct |
| `lint_test.go` | Lint baseline enforcement test (see [Lint Enforcement](#lint-enforcement)) |
| `testdata/lint_baseline.txt` | Captured `mise run lint` output; the baseline against which `lint_test.go` diffs |

---

# Dependency Injection

## ServiceContainer (legacy; slice 7)

Services passed through command tree via `ServiceContainer`:
```go
type ServiceContainer struct {
    Transformer   application.Transformer
    Validator     application.Validator
    Canonicalizer application.Canonicalizer
    Initializer   application.Initializer
}

services := cmd.NewServiceContainer()
```

> **Slice 1 (done):** `cmdutil.Factory` introduced as a unit (see [Foundation Units](#foundation-units-slice-1)). `main.go` is not yet rewired; `ServiceContainer` is still the live DI mechanism. **Slice 2** rewires `main.go` to construct a `*cmdutil.Factory`; **slice 7** deletes `container.go` and `NewServiceContainer()`.

## Composition Root

`main.go` wires all dependencies:
```go
services := cmd.NewServiceContainer()
cfg := &cmd.CommandConfig{
    Services:       services,
    ErrorFormatter: cmd.NewErrorFormatter(),
    Verbosity:      0,
}
rootCmd := cmd.NewRootCommand(cfg)
```

## Calling Services

Commands access services through interfaces:
```go
result, err := cfg.Services.Transformer.Transform(ctx, &application.TransformRequest{
    InputPath:  args[0],
    OutputPath: args[1],
    Platform:   platform,
})
```

## Constructor Pattern

Commands use `NewXCommand(cfg *CommandConfig)` constructors with RunE pattern:
```go
func NewValidateCommand(cfg *CommandConfig) *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.RunE = func(c *cobra.Command, args []string) error {
        verbosity, _ := c.Flags().GetCount("verbose")
        cfg.Verbosity = Verbosity(verbosity)
        // Use cfg.Services, cfg.ErrorFormatter
        // Return errors (bubble up to main.go for centralized handling)
        return nil
    }
    return cmd
}
```

No `init()` functions or global command variables.

> **Slice 2 transition:** constructor signature changes to `NewCmdXxx(f *Factory, runF func(*XxxOptions) error)`. `runF` is the test-injection seam; production wires it to `runXxx`, tests substitute a stub.

## Canonical examples (slice 2)

The two pilot migrations in slice 2 (`adapt` and `library resources`)
are the canonical references for the new pattern. See:

- `cmd/adapt.go` — `NewCmdAdapt(f, runF)` + `runAdapt(opts)`; uses
  `core.ValidatePlatform`, `opts.IO.Out`, `opts.IO.Verbosef`. Defines
  the `Transformer` interface inline (interfaces where consumed).
- `cmd/resources.go` — `NewCmdResources(f, libraryPath, runF)` +
  `runResources(opts)`; dispatches on `opts.Output` to the JSON or
  table exporter, or plain output via the shared `formatResourcesList`
  helper. The `libraryPath *string` parameter is the parent's
  shared `--library` pointer so the parent's flag value is honored.
- `cmd/legacy_bridge.go` — `LegacyBridge` shim (transitional; slice 7
  deletes it). `legacyCfgFrom(bridge)` builds the per-command
  `CommandConfig` consumed by non-migrated commands during the
  migration window.

---

# CommandConfig (legacy; slice 7)

Holds configuration and services for command execution:
```go
type CommandConfig struct {
    Services       *ServiceContainer
    ErrorFormatter *ErrorFormatter
    Verbosity      Verbosity
}
```

> Replaced by `*cmdutil.Factory` in slice 2; `CommandConfig` struct is deleted in slice 7.

---

# Required Flags

Both `adapt` and `validate` require `--platform` flag:
```go
_ = cmd.MarkFlagRequired("platform")
```

Validation uses typed ConfigError:
```go
if platform == "" {
    HandleError(cfg, gerrors.NewConfigError("platform", "",
        []string{models.PlatformClaudeCode, models.PlatformOpenCode},
        "--platform flag is required"))
}
```

> Use `core.PlatformClaudeCode` / `core.PlatformOpenCode` constants (slice 1+) instead of `models.Platform*`. The `core.ValidatePlatform(s)` helper (slice 1) replaces the inline `[]string{...}` checks for new code.

## Supported Platforms

- `claude-code` - Claude Code document format
- `opencode` - OpenCode document format

---

# Verbosity Flag (legacy helpers; slice 7)

Persistent `-v`/`-vv` flag on root command for all subcommands:
```go
rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (use -v or -vv)")
```

Levels:
- Level 0 (default): No verbose output
- Level 1 (`-v`): Basic progress info
- Level 2 (`-vv`): Detailed operation info

Usage:
```go
VerbosePrint(cfg, "Processing file: %s", filePath)      // Level 1+
VeryVerbosePrint(cfg, "Parsing YAML structure...")      // Level 2+
```

Output goes to stderr (stdout stays clean for piping).

> **Slice 1 introduced `iostreams.IOStreams.Verbosef` (slice 1 unit).** Slice 2+ commands use `opts.IO.Verbosef(format, args...)` directly. `VerbosePrint`/`VeryVerbosePrint` and the `Verbosity` type are removed in slice 7.

---

# Exit Codes (transitioning to 0/1/2)

Current exit codes (still in use; `CategorizeError` is the live mapping):
- `0` (Success) - Command completed successfully
- `1` (Error) - General errors (transform, file, unexpected)
- `2` (Usage) - Cobra argument/validation errors (invalid flags, missing args)
- `3` (Config) - Configuration/parsing errors (malformed YAML, config errors)
- `4` (Git) - Git-related errors
- `5` (Validation) - Document validation errors
- `6` (NotFound) - File/resource not found errors

Error categorization via `CategorizeError()` using `errors.As` for type detection.

> **Slice 1 introduced `cmdutil.ExitCodeFor(err)` returning 0/1/2.** The seven-code scheme collapses to three (`ExitCodeSuccess=0`, `ExitCodeError=1`, `ExitCodeUsage=2`) when `main.go` is rewired in slice 2. `CategorizeError` and the `Category*` enum are deleted in slice 7.

---

# Error Handling

## Centralized Error Handling

Errors bubble up to main.go via RunE pattern:
```go
// main.go
cmd.SetGlobalCommandConfig(cfg)
if err := rootCmd.Execute(); err != nil {
    exitCode := cmd.HandleCLIError(rootCmd, err)
    os.Exit(int(exitCode))
}
```

```go
func HandleCLIError(c *cobra.Command, err error) ExitCode {
    // Formats and outputs error, returns exit code
    // Uses global CommandConfig set during command construction
}
```

> **Slice 1 introduced `output.FormatError(io, err)`** (slice 1 unit). Slice 2+ replaces `ErrorFormatter` with `output.FormatError` calls in command `runXxx` bodies. `HandleCLIError` is removed in slice 7.

## Error Formatter (legacy; slice 7)

Type-specific formatting with contextual hints:
- ParseError → "Parse error: <message> File: <path>"
- ValidationError → "Validation error: <message>" + "Hint:" lines
- TransformError → "Transform error (<operation> for <platform>): <message>"
- FileError → "File error (<operation>): <message> Path: <path>"
- ConfigError → "Config error: <message>" + "Available: <options>"

## Typed Errors

Import from `internal/core/` (renamed from `internal/domain/` in slice 1):
```go
import "gitlab.com/amoconst/germinator/internal/core"

// Constructors
core.NewParseError(path, message, cause)
core.NewValidationError(message, field, suggestions)
core.NewTransformError(operation, platform, message, cause)
core.NewFileError(path, operation, message, cause)
core.NewConfigError(field, value, available, message)
core.NewInitializeError(ref, inputPath, outputPath, cause)  // slice 1; consumer in slice 5
core.NewPartialSuccessError(succeeded, failed, errors)     // slice 1; consumer in slice 5
```

---

# Argument Count

`adapt` and `validate` use `cobra.ExactArgs(2)` and `cobra.ExactArgs(1)` respectively.
`root` and `version` use default (no arguments).

---

# Version Output Format

```go
fmt.Printf("germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)
```

Example: `germinator v0.3.20 (abc123def) 2026-02-04`

---

# Testing

`cmd_test.go` contains integration tests for CLI workflows.
Test both platforms when testing platform-specific commands.

Test files:
- `verbose_test.go` - Verbosity type and helper function tests
- `error_formatter_test.go` - Error formatting tests
- `library_test.go` - Library and init command tests
- `completions_test.go` - Completion action unit tests (cache, timeout, actions)
- `validate_test.go` - Validate command integration tests
- `library_init_test.go` - Library init tests
- `library_create_test.go` - Library create tests
- `lint_test.go` - Lint baseline enforcement (see [Lint Enforcement](#lint-enforcement))

> For new command tests, prefer `runF` injection with `iostreams.Test()` buffers over the deprecated `test/mocks/` package. Slice 2+ commands receive `runF` as a constructor parameter; tests substitute a stub.

---

# Foundation Units (slice 1)

The following units exist as of slice 1; **`main.go` is not yet wired to them** (that happens in slice 2). They are testable, fully covered, and safe to consume from new code.

| Unit | Package | Purpose |
|------|---------|---------|
| `iostreams.IOStreams` | `internal/iostreams/` | Single terminal I/O boundary; `System()` (real) and `Test()` (buffer) constructors; `Verbosef`, TTY detection, `Styles` |
| `iostreams.Styles` | `internal/iostreams/styles.go` | `Error`/`Success`/`Warning`/`Dim`/`Bold` via `lipgloss`; respects `NO_COLOR` and TTY |
| `output.FormatError` | `internal/output/errors.go` | Dispatches on typed `*core.*Error` via `errors.As`; writes to `io.ErrOut` |
| `output.Exporter` + `JSONExporter` + `TableExporter` | `internal/output/exporter.go` | Format-flexible output (`tab:"HEADER"` struct tag) |
| `output.AddOutputFlags` | `internal/output/output_flags.go` | Wires `--output` (`json`/`table`/`plain`) with completion |
| `cmdutil.Factory` | `internal/cmdutil/factory.go` | Lazy `func() (T, error)` fields with `sync.OnceValues` caching; `NewFactory(io, ver, exe)` |
| `cmdutil.ExitCodeFor` | `internal/cmdutil/exit.go` | Maps `error` → 0/1/2 (0 if `*core.PartialSuccessError{Succeeded>0}`) |
| `cmdutil.AddOutputFlags` | `internal/cmdutil/output_flags.go` | Re-export of `output.AddOutputFlags` so cmd files import only `cmdutil` |
| `core.ValidatePlatform` | `internal/core/rules.go` | Returns `*core.ValidationError` for unknown platform strings |
| `core.ResolveOutputPath` | `internal/core/rules.go` | `(docType, name, platform) → path` (e.g., `agents/reviewer.claude-code.md`); consumer in slice 5 |

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

**Adding a new intentional violation** (e.g., during a slice migration):

1. Make the change.
2. Run `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` to refresh the baseline.
3. Commit the baseline file alongside the change.

> The test re-runs lint multiple times because `golangci-lint`'s parallel linters report in non-deterministic order; the union of stable violations is what the test gates on.

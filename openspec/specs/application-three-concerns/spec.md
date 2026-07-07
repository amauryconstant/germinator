# application-three-concerns Specification

## Purpose

Define the three-concern separation for every CLI command — Parse (Cobra flags + args + config), Execute (functional core validation + transformation), Respond (formatting + exit codes) — and the dependency rule that enforces it.

## Requirements

### Requirement: Three concerns per command

Every CLI command SHALL be structured as three concerns:

1. **Parse** — collect input from flags, arguments, environment variables, config files
2. **Execute** — validate inputs and produce a result via the functional core
3. **Respond** — format the result for the terminal (stdout, stderr, exit code)

These concerns map to the per-command files in `cmd/`:

- **Parse** lives in the `NewCmdXxx` constructor (flag definitions, RunE wiring)
- **Execute** lives in the `runXxx` function body (validates inputs, calls core, accumulates errors)
- **Respond** lives in shared output helpers (`output.FormatError`, `opts.IO.Verbosef`, `cmdutil.ExitCodeFor`)

#### Scenario: Per-command file layout

- **WHEN** `cmd/<command>.go` is inspected
- **THEN** it SHALL contain:
  - `<command>Options` struct (parsed inputs + Factory-derived dependencies)
  - `NewCmd<Command>(f *cmdutil.Factory, runF func(*<command>Options) error) *cobra.Command` (constructor)
  - `run<Command>(opts *<command>Options) error` (production body)
- **AND** the constructor SHALL NOT perform business logic
- **AND** the run body SHALL NOT parse flags

### Requirement: Parse concern boundaries

The constructor (`NewCmdXxx`) is responsible only for Parse. It SHALL NOT call into the functional core, perform file I/O, or write to `Out`/`ErrOut` directly (other than the implicit `--help` text rendered by Cobra).

#### Scenario: Constructor is parse-only

- **WHEN** `NewCmdAdapt(f, nil)` is called
- **THEN** the returned `*cobra.Command` SHALL have its flags registered
- **AND** `RunE` SHALL populate `*adaptOptions` from `f.IOStreams`, `f.RootContext`, parsed flags, and args
- **AND** `RunE` SHALL NOT call any function in `internal/core/`, `internal/library/`, or other shell packages beyond lazy resolution

### Requirement: Execute concern boundaries

The `runXxx` function is responsible only for Execute. It SHALL accept a fully-populated `*XxxOptions` and SHALL NOT call Cobra APIs (`cmd.Flags().GetString(...)`, `cobra.Command`, etc.).

#### Scenario: runXxx takes only options

- **WHEN** `runAdapt(opts *adaptOptions) error` is called
- **THEN** `opts` SHALL already contain all parsed flags (`opts.InputPath`, `opts.OutputPath`, `opts.Platform`)
- **AND** `runAdapt` SHALL NOT access `cobra.Command` or `pflag.FlagSet` APIs
- **AND** `runAdapt` SHALL be testable in isolation by constructing a fake `*adaptOptions` (no Cobra needed)

### Requirement: Respond concern boundaries

The Respond layer formats the result. Error formatting SHALL be centralized in `output.FormatError(io, err)` — individual commands SHALL NOT call `output.FormatError` themselves (the single-handling rule; `main.go` renders errors exactly once).

#### Scenario: Command returns error, main.go formats

- **WHEN** a command's `RunE` returns a typed error
- **THEN** `main.go`'s deferred handler SHALL call `output.FormatError(io, err)` exactly once
- **AND** the command body SHALL NOT have called `output.FormatError` itself (no double-rendering)
- **AND** the command body SHALL NOT have written to `ErrOut` directly for error cases

### Requirement: Three concerns map to package boundaries

The three concerns SHALL map to the following package boundaries:

- **Parse** lives in `cmd/<command>.go` (Cobra wiring + per-command options)
- **Execute** lives in `cmd/<command>.go` (the `runXxx` body) AND `internal/core/` (the pure business logic)
- **Respond** lives in `internal/output/` (formatting) + `internal/cmdutil/exit.go` (exit codes) + `internal/iostreams/` (TTY + Styles)

#### Scenario: Execute calls into core

- **WHEN** `runXxx(opts)` needs to validate input or transform a document
- **THEN** it SHALL call a function in `internal/core/` (e.g., `core.ValidatePlatform`, `core.NewValidationError`)
- **AND** it SHALL NOT inline the business logic in the run body

### Requirement: Concern-boundary lints (forbidigo)

The `forbidigo` linter SHALL enforce the three-concern boundaries. Patterns (in `cmd/**` excluding `main.go` and `_test.go`):

| Pattern | Why forbidden |
|---|---|
| `fmt.Fprintf(os.Stdout\|Stderr, ...)` | Use `opts.IO.Out` / `opts.IO.ErrOut` (Respond concern) |
| `os.Exit(...)` | Use `cmdutil.ExitCodeFor(err)` (Respond concern) |
| `var global(Factory\|CommandConfig\|ServiceContainer)` | Composition must flow through `*Factory` (Parse concern) |
| `SetGlobal(Factory\|CommandConfig)` | Same |
| `context.Background()` | Use `opts.IO.RootContext` (Factory-owned, signal-aware) |
| `output.FormatError(` | Single-handling rule: only `main.go` formats errors (Respond concern) |

#### Scenario: Lint baseline catches forbidden patterns

- **WHEN** a contributor adds `os.Exit(1)` inside a `runXxx` body
- **THEN** `mise run lint` SHALL report the violation
- **AND** `cmd/lint_test.go` SHALL fail the diff against the baseline

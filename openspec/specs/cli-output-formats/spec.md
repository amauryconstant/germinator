# cli-output-formats Specification

## Purpose

Define the `--output json|table|plain` flag and the `Exporter` interface shared by all read-only commands. This capability establishes the foundation (`output.AddOutputFlags`, the `Exporter` interface, and `JSONExporter`/`TableExporter`).

> **Note:** The `--output` flag was introduced alongside the broader `--output json|table|plain` design. Prior to this, several commands exposed a one-off `--json` flag; that flag was removed in favor of the unified `--output` flag described here. Historical references to the `--json` flag are intentionally not preserved in this spec — see CHANGELOG.md for the removal history.

## Requirements

### Requirement: AddOutputFlags helper

The `output.AddOutputFlags` function SHALL add a `--output` string flag to a Cobra command.

**Change**: rehome the function from `internal/cmdutil` to `internal/output`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Flag registration

- **WHEN** `output.AddOutputFlags(cmd, &opts.Output)` is called
- **THEN** `cmd` SHALL have a `--output` string flag with default value `"plain"`
- **AND** valid values are `json`, `table`, `plain`
- **AND** shell completion is wired for those three values via `cobra.RegisterFlagCompletionFunc`

#### Scenario: Default value

- **WHEN** the `--output` flag is not provided on the command line
- **THEN** `opts.Output` SHALL be `"plain"` after flag parsing

#### Scenario: cmdutil does not re-export output.AddOutputFlags

- **GIVEN** a command author writes `cmdutil.AddOutputFlags(...)` in any cmd file
- **WHEN** `mise run build` runs
- **THEN** the build SHALL fail with `undefined: cmdutil.AddOutputFlags`

### Requirement: Exporter interface

The `output.Exporter` interface SHALL define a single `Write` method.

#### Scenario: Exporter interface

- **WHEN** a type implements `Exporter`
- **THEN** it SHALL expose `Write(io *iostreams.IOStreams, data any) error`

> **Divergence from the `golang-cli-architecture` skill:** The skill recommends a `Write(w io.Writer, data any) error` signature for portable, single-purpose formatters. Germinator intentionally uses `*iostreams.IOStreams` because every exporter needs access to TTY detection (adaptive output) and `Styles` (color). Tests inject buffers via `iostreams.Test()`. Switching to `io.Writer` would force every consumer to thread the IOStreams separately with no functional gain.

### Requirement: JSONExporter

The `output.JSONExporter` type SHALL serialize data to JSON.

#### Scenario: JSONExporter.Write

- **GIVEN** a `JSONExporter` and a buffer-backed `IOStreams`
- **WHEN** `Write(io, data)` is called with a Go struct/map
- **THEN** it SHALL encode `data` to `io.Out` using `json.MarshalIndent(v, "", "  ")` (2-space indent)
- **AND** it SHALL append a trailing newline `\n`
- **AND** it SHALL return any encoding error

### Requirement: TableExporter

The `output.TableExporter` type SHALL render tabular data.

#### Scenario: TableExporter.Write

- **GIVEN** a `TableExporter` and a buffer-backed `IOStreams`
- **WHEN** `Write(io, data)` is called with a slice of structs (or a `[]map[string]string`)
- **THEN** it SHALL render the data as an aligned table to `io.Out`
- **AND** it SHALL use `text/tabwriter` for column alignment

### Requirement: Format dispatch in commands

Read-only commands SHALL dispatch on `opts.Output` and call the appropriate `Exporter`.

#### Scenario: Plain output (default)

- **GIVEN** `opts.Output == "plain"`
- **WHEN** a read-only command runs
- **THEN** the command SHALL print data using its built-in plain-text formatting (no Exporter call)

#### Scenario: JSON output

- **GIVEN** `opts.Output == "json"`
- **WHEN** a read-only command runs
- **THEN** the command SHALL construct a `JSONExporter` and call `Write(opts.IO, data)`

#### Scenario: Table output

- **GIVEN** `opts.Output == "table"`
- **WHEN** a read-only command runs
- **THEN** the command SHALL construct a `TableExporter` and call `Write(opts.IO, data)`

### Requirement: AddOutputFlags is opt-in per command

Each command SHALL call `output.AddOutputFlags` if and only if it produces structured machine-readable output that benefits from a `--output` flag.

**Commands that SHALL call `output.AddOutputFlags`**: any read-only command whose primary result is a structured data structure (validation result, list of resources, list of presets, per-resource action result). Examples include commands that emit validation results, per-resource install/add/remove results, resource or preset lists, resource/preset detail views, and per-resource sync results.

**Commands that SHALL NOT call `output.AddOutputFlags`**: any command whose primary result is a file, script, or short one-line text written to stdout. Examples include commands that write platform files, write canonical files, write one-line version strings, write shell completion scripts, write config files, or write library scaffolding.

> **Out of scope:** NDJSON / JSON-Lines streaming. All current read-only commands produce JSON arrays (or single objects); no exporter implements NDJSON. Adding NDJSON would require defining the array-vs-streaming boundary per command; deferred until a streaming use case arises.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export was deleted in change `remove-cmdutil-output-reexport`; callers must import `internal/output` directly.

#### Scenario: adapt has no --output flag

- **WHEN** `germinator adapt --help` is invoked
- **THEN** the help output SHALL NOT include an `--output` flag

#### Scenario: library resources has --output flag

- **WHEN** `germinator library resources --help` is invoked
- **THEN** the help output SHALL include an `--output` flag with values `plain|table|json`

### Requirement: PersistentFlags wiring for parent commands

The `output.AddOutputFlags` function SHALL bind to `cmd.Flags()` (local-only). Parent commands requiring inherited `--output` flag visibility on subcommands SHALL wire the flag manually using `cmd.PersistentFlags()` and SHALL NOT call `output.AddOutputFlags` (which binds to local `cmd.Flags()`).

**Change**: NEW requirement documenting the limitation of `output.AddOutputFlags` (it binds to `cmd.Flags()`, which is local-only). Library subcommands that need inherited flags (e.g., `library remove`) MUST wire the `--output` flag manually via `cmd.PersistentFlags()`. This contract prevents future contributors from "fixing" the inline wiring by extracting a helper that abstracts over `Flags()` vs `PersistentFlags()` — the two flag-set bindings are intentionally distinct.

#### Scenario: PersistentFlags wiring for library remove

- **WHEN** a parent command like `library remove` needs the `--output` flag to be inherited by its subcommands (`resource`, `preset`)
- **THEN** the parent SHALL wire the flag manually using `cmd.PersistentFlags().StringVar(&outputFormat, "output", "plain", "Output format")` and `cmd.RegisterFlagCompletionFunc`
- **AND** it SHALL NOT call `output.AddOutputFlags(cmd, ...)` (which binds to local `cmd.Flags()`)

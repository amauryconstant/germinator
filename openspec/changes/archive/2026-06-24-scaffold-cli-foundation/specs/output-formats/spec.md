# output-formats Specification

## Purpose

Define the `--output json|table|plain` flag and the `Exporter` interface shared by all read-only commands. This change establishes the foundation (`cmdutil.AddOutputFlags`, the `Exporter` interface, and `JSONExporter`/`TableExporter`); per-command adoption and replacement of legacy `--json` flags is sequenced across changes 2-9.

## ADDED Requirements

### Requirement: AddOutputFlags helper

The `cmdutil.AddOutputFlags` function SHALL add a `--output` string flag to a Cobra command.

#### Scenario: Flag registration

- **WHEN** `cmdutil.AddOutputFlags(cmd, &opts.Output)` is called
- **THEN** `cmd` SHALL have a `--output` string flag with default value `"plain"`
- **AND** valid values are `json`, `table`, `plain`
- **AND** shell completion is wired for those three values via `cobra.RegisterFlagCompletionFunc`

#### Scenario: Default value

- **WHEN** the `--output` flag is not provided on the command line
- **THEN** `opts.Output` SHALL be `"plain"` after flag parsing

### Requirement: Exporter interface

The `output.Exporter` interface SHALL define a single `Write` method.

#### Scenario: Exporter interface

- **WHEN** a type implements `Exporter`
- **THEN** it SHALL expose `Write(io *iostreams.IOStreams, data any) error`

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

Each command SHALL call `cmdutil.AddOutputFlags` if and only if it produces structured machine-readable output that benefits from a `--output` flag.

**Commands that SHALL call `cmdutil.AddOutputFlags`**: any read-only command whose primary result is a structured data structure (validation result, list of resources, list of presets, per-resource action result). Examples include commands that emit validation results, per-resource install/add/remove results, resource or preset lists, resource/preset detail views, and per-resource sync results.

**Commands that SHALL NOT call `cmdutil.AddOutputFlags`**: any command whose primary result is a file, script, or short one-line text written to stdout. Examples include commands that write platform files, write canonical files, write one-line version strings, write shell completion scripts, write config files, or write library scaffolding.

#### Scenario: adapt has no --output flag

- **WHEN** `germinator adapt --help` is invoked
- **THEN** the help output SHALL NOT include an `--output` flag

#### Scenario: library resources has --output flag

- **WHEN** `germinator library resources --help` is invoked
- **THEN** the help output SHALL include an `--output` flag with values `plain|table|json`

> **Status (slice 1 / foundation):** this requirement describes the target state after all commands are migrated (changes 2-9). This change creates `cmdutil.AddOutputFlags` and the `Exporter` interface; adoption is per-command in subsequent changes.

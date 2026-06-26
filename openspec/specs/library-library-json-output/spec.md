# Capability: Library JSON Output

## Purpose

Define per-command structured output for `germinator library` subcommands via a string `--output` flag accepting `plain`, `json`, or `table`. Each library sub-command opts in independently via `cmdutil.AddOutputFlags`. The parent-inherited `--json` flag mechanism is removed; consumers should use the per-command flag (e.g., `germinator library resources --output json`).

> **Stream contract:** All `--output` formats (plain, json, table) write primary data to **stdout** (`opts.IO.Out`). Verbose progress writes to **stderr** (`opts.IO.ErrOut` via `opts.IO.Verbosef`). Errors write to **stderr** via `output.FormatError`. Never mix diagnostic output into stdout — this preserves `germinator library resources --output json | jq '.'`.

## Requirements

### Requirement: library resources supports --output flag

The `library resources` command SHALL expose a per-command string flag accepting the values `json`, `table`, or `plain` (default `plain`).

#### Scenario: Plain is the default

- **WHEN** `germinator library resources` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library resources` output)
- **AND** the output SHALL be written to **stdout** (`opts.IO.Out`)

#### Scenario: JSON output via --output json

- **WHEN** `germinator library resources --output json` is invoked
- **THEN** the output SHALL be JSON-formatted (2-space indent, trailing newline) via `output.NewJSONExporter().Write(opts.IO, data)`
- **AND** the JSON output SHALL be written to **stdout**
- **AND** the JSON structure SHALL match the previous `--json` flag output: `{"resources": [{"type": "...", "name": "...", "description": "...", "path": "..."}, ...]}`

#### Scenario: Table output via --output table

- **WHEN** `germinator library resources --output table` is invoked
- **THEN** the output SHALL be a tab-aligned table via `output.NewTableExporter().Write(opts.IO, data)`
- **AND** the table SHALL be written to **stdout**
- **AND** the column order SHALL be derived from `tab:"HEADER"` struct tags on the resource type, falling back to field names when no tag is present (per `internal/output/exporter.go`)

#### Scenario: Old --json flag is rejected

- **WHEN** `germinator library resources --json` is invoked
- **THEN** the command SHALL return a usage error
- **AND** the process SHALL exit with code 2 (`ExitCodeUsage`)

### Requirement: JSON encoding uses proper formatting

All JSON output SHALL use `output.NewJSONExporter().Write(opts.IO, data)` for formatted output with 2-space indentation and a trailing newline.

#### Scenario: JSON output is pretty-printed

- **GIVEN** a library with resources
- **WHEN** user runs `germinator library resources --output json`
- **THEN** the JSON output has 2-space indentation and is human-readable

> **Status:** Other library commands (`presets`, `show`, `add`, `create`, `init`, `refresh`, `remove`, `validate`) get `--output` in their respective migration changes.

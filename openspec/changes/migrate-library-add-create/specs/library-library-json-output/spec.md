# library-library-json-output Specification (delta)

> **Stream contract:** All `--output` formats (plain, json, table) write primary data to **stdout** (`opts.IO.Out`). Per-resource status and errors write to **stderr** (`opts.IO.ErrOut` via `output.FormatError` and `opts.IO.Verbosef`). Never mix diagnostic output into stdout — this preserves `germinator library add --discover --batch --force --output json | jq '.'`.

## MODIFIED Requirements

### Requirement: library add supports --output flag

The `library add` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`. The legacy `--json` flag registration (if any) is replaced by `--output`.

#### Scenario: Default plain output

- **WHEN** `germinator library add <file>` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library add` output)
- **AND** the output SHALL be written to **stdout** (`opts.IO.Out`)

#### Scenario: JSON output

- **WHEN** `germinator library add --discover --batch --force --output json` is invoked
- **THEN** the per-resource results SHALL be JSON-formatted using `output.NewJSONExporter().Write(opts.IO, data)`
- **AND** the JSON output SHALL be written to **stdout**
- **AND** `stderr` SHALL be empty (no diagnostic leakage)

#### Scenario: Table output

- **WHEN** `germinator library add --discover --output table` is invoked
- **THEN** the per-resource results SHALL be rendered as an aligned table using `output.NewTableExporter().Write(opts.IO, data)`
- **AND** the table SHALL be written to **stdout**
- **AND** the column order SHALL be derived from `tab:"HEADER"` struct tags, falling back to field names

#### Scenario: Legacy --json flag is rejected

- **WHEN** `germinator library add --json` is invoked
- **THEN** the command SHALL return a usage error
- **AND** the process SHALL exit with code 2 (`ExitCodeUsage`)

## ADDED Requirements

### Requirement: library create does NOT support --output flag

The `library create` command SHALL NOT expose a `--output` flag. The legacy implementation did not support `--json`; the absence of `--output` is consistent with the `output-formats` capability's "only commands that previously supported `--json` get `--output`" rule.

#### Scenario: No --output flag on library create preset

- **WHEN** `germinator library create preset --help` is invoked
- **THEN** the help output SHALL NOT include an `--output` flag

### Requirement: library create preset is a leaf under library

The `library create preset` command SHALL be registered directly under the `library` parent command (not under an intermediate `library create` Cobra group wrapper). The user-facing command path `germinator library create preset <name> --resources ...` is unchanged.

#### Scenario: library create preset help resolves to a single command

- **WHEN** `germinator library create preset --help` is invoked
- **THEN** the help output SHALL show the `library create preset` command's own flags and description (not a subcommand list)

#### Scenario: library create has no subcommand list

- **WHEN** `germinator library create --help` is invoked
- **THEN** the help output SHALL list `preset` as a child command (Cobra's default parent behaviour)
- **AND** the help output SHALL NOT show the `library create` group's own help screen (i.e., `germinator library create` without `--help` does not display a separate description)

> **Status:** the `--output` flag is added to `library add` in change `migrate-library-add-create` (slice 6 of 9). The `library create preset` leaf collapse is also a slice-6 change. The other mutating library commands (`init`, `refresh`, `remove`, `validate`) get `--output` in change-7.

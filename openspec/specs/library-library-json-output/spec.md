# Capability: Library JSON Output

## Purpose

Define per-command structured output for `germinator library` subcommands via a string `--output` flag accepting `plain`, `json`, or `table`. Each library sub-command opts in independently via `output.AddOutputFlags`. The parent-inherited `--json` flag mechanism is removed; consumers should use the per-command flag (e.g., `germinator library resources --output json`).

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

### Requirement: library presets supports --output flag

The `library presets` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **WHEN** `germinator library presets` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library presets` output)
- **AND** the output SHALL be written to **stdout** (`opts.IO.Out`)

#### Scenario: Explicit plain output matches default

- **WHEN** `germinator library presets --output plain` is invoked
- **THEN** the output SHALL be byte-identical to the default (no-flag) invocation

#### Scenario: JSON output

- **WHEN** `germinator library presets --output json` is invoked
- **THEN** the output SHALL be JSON-formatted (2-space indent) using `output.NewJSONExporter().Write(opts.IO, data)`
- **AND** the output SHALL be written to **stdout**
- **AND** `stderr` SHALL be empty (no diagnostic leakage)

#### Scenario: Table output

- **WHEN** `germinator library presets --output table` is invoked
- **THEN** the output SHALL be an aligned table using `output.NewTableExporter().Write(opts.IO, data)`
- **AND** the output SHALL be written to **stdout**
- **AND** the column order SHALL be derived from `tab:"HEADER"` struct tags, falling back to field names

#### Scenario: Parent --library flag is honored

- **WHEN** `germinator library presets --library /path/to/library` is invoked
- **THEN** the command SHALL load the library from `/path/to/library` (overriding `GERMINATOR_LIBRARY` env)
- **AND** the command SHALL output the contents of that library

### Requirement: library add supports --output flag

The `library add` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`. The legacy `--json` flag registration (if any) is replaced by `--output`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

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

### Requirement: library show supports --output flag

The `library show` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output for a resource ref

- **WHEN** `germinator library show skill/commit` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library show <resource>` output)
- **AND** the output SHALL be written to **stdout**

#### Scenario: Default plain output for a preset ref

- **WHEN** `germinator library show preset/git-workflow` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library show preset/<name>` output)
- **AND** the output SHALL be written to **stdout**

#### Scenario: JSON output

- **WHEN** `germinator library show <ref> --output json` is invoked
- **THEN** the output SHALL be JSON-formatted (2-space indent)
- **AND** the output SHALL be written to **stdout**
- **AND** `stderr` SHALL be empty

#### Scenario: Table output

- **WHEN** `germinator library show <ref> --output table` is invoked
- **THEN** the output SHALL be an aligned table
- **AND** the output SHALL be written to **stdout**

#### Scenario: Parent --library flag is honored

- **WHEN** `germinator library show <ref> --library /path/to/library` is invoked
- **THEN** the command SHALL resolve `<ref>` against the library at `/path/to/library`

### Requirement: NotFoundError on missing ref

When `library show <ref>` is invoked with a ref that doesn't resolve (neither as a resource nor as a preset), the command SHALL return `*core.NotFoundError`.

#### Scenario: Ref not found

- **WHEN** `germinator library show nonexistent-ref` is invoked
- **THEN** the command SHALL return `*core.NotFoundError` with `Key == "nonexistent-ref"`
- **AND** `output.FormatError` SHALL render it as `Error: not found: nonexistent-ref\n` to **stderr** (`opts.IO.ErrOut`)
- **AND** `stdout` SHALL be empty (no data leakage on error paths)
- **AND** `cmdutil.ExitCodeFor` SHALL map it to `ExitCodeUsage` (2) via the `errors.As(err, &notFound)` branch in `internal/cmdutil/exit.go:73-75`

#### Scenario: Empty ref is a not-found error

- **WHEN** `germinator library show ""` is invoked
- **THEN** the command SHALL return `*core.NotFoundError` with `Key == ""`
- **AND** `output.FormatError` SHALL render it as `Error: not found: \n` to stderr
- **AND** the process SHALL exit with code 2

### Requirement: Stream discipline for library presets and library show

Both commands SHALL write primary data to stdout and diagnostic information (errors, future verbose progress) to stderr.

#### Scenario: JSON output is pipeable

- **GIVEN** `germinator library presets --output json` or `germinator library show <ref> --output json` produces JSON on stdout
- **WHEN** the output is piped through `jq '.presets[0].name'` (presets) or `jq '.ref'` (show)
- **THEN** `jq` SHALL successfully parse the JSON
- **AND** `stderr` SHALL contain no error or progress messages for a successful invocation

#### Scenario: Error output stays on stderr

- **GIVEN** `germinator library show nonexistent-ref` returns `*core.NotFoundError`
- **WHEN** the output is piped through `jq '.'`
- **THEN** `jq` SHALL report a parse error (because stdout is empty) — but more importantly, the user's error message SHALL be on stderr and discoverable via `2>&1`

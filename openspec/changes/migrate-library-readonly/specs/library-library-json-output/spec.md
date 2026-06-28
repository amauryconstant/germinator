# library-json-output Specification (delta)

> **Stream contract:** All `--output` formats (plain, json, table) write primary data to **stdout** (`opts.IO.Out`). Verbose progress writes to **stderr** (`opts.IO.ErrOut` via `opts.IO.Verbosef`). Errors write to **stderr** via `output.FormatError`. Never mix diagnostic output into stdout — this preserves `germinator library presets --output json | jq '.'` and `germinator library show <ref> --output json | jq '.'`.

> **Precondition:** this delta depends on the `*core.NotFoundError` foundation unit introduced in task group 4.0 of the same change. The type and its `output.FormatError` dispatch branch land before task 4.2.3 (which returns the type).

## MODIFIED Requirements

### Requirement: library presets supports --output flag

The `library presets` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

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

### Requirement: library show supports --output flag

The `library show` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

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
- **AND** `cmdutil.ExitCodeFor` SHALL map it to `ExitCodeError` (1) via the default-error case in `internal/cmdutil/exit.go:71`

#### Scenario: Empty ref is a not-found error

- **WHEN** `germinator library show ""` is invoked
- **THEN** the command SHALL return `*core.NotFoundError` with `Key == ""`
- **AND** `output.FormatError` SHALL render it as `Error: not found: \n` to stderr
- **AND** the process SHALL exit with code 1

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

> **Status:** the `--output` flag is added to `library presets` and `library show` in change-4 (`migrate-library-readonly`). Other library commands (`add`, `init`, `refresh`, `remove`, `validate`) get `--output` in changes 6, 7. The `*core.NotFoundError` foundation unit is introduced in task group 4.0 of this same change and is consumed by `library show`'s not-found path.

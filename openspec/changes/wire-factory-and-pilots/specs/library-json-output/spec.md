# library-json-output Specification (delta)

## MODIFIED Requirements

### Requirement: library resources supports --output flag

The `library resources` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Plain is the default

- **WHEN** `germinator library resources` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library resources` output)

#### Scenario: JSON output via --output json

- **WHEN** `germinator library resources --output json` is invoked
- **THEN** the output SHALL be JSON-formatted (2-space indent, trailing newline) using `output.JSONExporter`
- **AND** the format SHALL match the previous `--json` flag output (the old `--json` flag is removed; use `--output json` instead)

#### Scenario: Table output via --output table

- **WHEN** `germinator library resources --output table` is invoked
- **THEN** the output SHALL be an aligned table using `output.TableExporter`

> **Status (this change):** the `--output` flag is added to `library resources` only. Other library commands (`presets`, `show`, `add`, `init`, `refresh`, `remove`, `validate`) get `--output` in changes 4, 6, 7 respectively.

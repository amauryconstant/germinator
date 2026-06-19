# library-json-output Specification (delta)

## MODIFIED Requirements

### Requirement: library add supports --output flag

The `library add` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **WHEN** `germinator library add <file>` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library add` output)

#### Scenario: JSON output

- **WHEN** `germinator library add --discover --batch --force --output json` is invoked
- **THEN** the per-resource results SHALL be JSON-formatted using `output.JSONExporter`

#### Scenario: Table output

- **WHEN** `germinator library add --discover --output table` is invoked
- **THEN** the per-resource results SHALL be rendered as an aligned table

### Requirement: library create does NOT support --output flag

The `library create` command SHALL NOT expose a `--output` flag. The legacy implementation did not support `--json`; the absence of `--output` is consistent with the `output-formats` capability's "only commands that previously supported `--json` get `--output`" rule.

#### Scenario: No --output flag

- **WHEN** `germinator library create preset --help` is invoked
- **THEN** the help output SHALL NOT include an `--output` flag

> **Status:** the `--output` flag is added to `library add` in change-6 (`migrate-library-add-create`). The other mutating library commands (`init`, `refresh`, `remove`, `validate`) get `--output` in change-7.

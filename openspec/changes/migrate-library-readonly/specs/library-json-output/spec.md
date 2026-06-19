# library-json-output Specification (delta)

## MODIFIED Requirements

### Requirement: library presets supports --output flag

The `library presets` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **WHEN** `germinator library presets` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library presets` output)

#### Scenario: JSON output

- **WHEN** `germinator library presets --output json` is invoked
- **THEN** the output SHALL be JSON-formatted (2-space indent) using `output.JSONExporter`

#### Scenario: Table output

- **WHEN** `germinator library presets --output table` is invoked
- **THEN** the output SHALL be an aligned table using `output.TableExporter`

### Requirement: library show supports --output flag

The `library show` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **WHEN** `germinator library show <ref>` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library show` output)

#### Scenario: JSON output

- **WHEN** `germinator library show <ref> --output json` is invoked
- **THEN** the output SHALL be JSON-formatted

#### Scenario: Table output

- **WHEN** `germinator library show <ref> --output table` is invoked
- **THEN** the output SHALL be an aligned table

### Requirement: NotFoundError on missing ref

When `library show <ref>` is invoked with a ref that doesn't resolve (neither as a resource nor as a preset), the command SHALL return `*core.NotFoundError`.

#### Scenario: Ref not found

- **WHEN** `germinator library show nonexistent-ref` is invoked
- **THEN** the command SHALL return `*core.NotFoundError`
- **AND** `output.FormatError` SHALL render it as `Error: not found: nonexistent-ref`
- **AND** `cmdutil.ExitCodeFor` SHALL map it to `ExitCodeError` (1)

> **Status:** the `--output` flag is added to `library presets` and `library show` in change-4 (`migrate-library-readonly`). Other library commands (`add`, `init`, `refresh`, `remove`, `validate`) get `--output` in changes 6, 7.

# library-library-json-output Specification (delta)

## MODIFIED Requirements

### Requirement: library presets supports --output flag

The `library presets` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **WHEN** `germinator library presets` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library presets` output)

#### Scenario: JSON output

- **WHEN** `germinator library presets --output json` is invoked
- **THEN** the output SHALL be JSON-formatted

#### Scenario: Table output

- **WHEN** `germinator library presets --output table` is invoked
- **THEN** the output SHALL be a table with columns derived from each preset's structured fields

### Requirement: library add supports --output flag

The `library add` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`. The legacy `--json` flag registration (if any) is replaced by `--output`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **WHEN** `germinator library add <file>` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library add` output)

#### Scenario: JSON output

- **WHEN** `germinator library add <file> --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

### Requirement: library show supports --output flag

The `library show` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output for a resource ref

- **WHEN** `germinator library show skill/commit` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library show <resource>` output)

#### Scenario: JSON output for a resource ref

- **WHEN** `germinator library show skill/commit --output json` is invoked
- **THEN** the resource detail SHALL be JSON-formatted

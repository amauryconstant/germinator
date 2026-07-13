# library-library-validation Specification (delta)

## MODIFIED Requirements

### Requirement: library validate supports --output flag

The `library validate` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **GIVEN** a library with validation issues
- **WHEN** `germinator library validate` is invoked without `--output`
- **THEN** the output SHALL be plain text with a list of validation issues

#### Scenario: JSON output

- **GIVEN** a library with validation issues
- **WHEN** `germinator library validate --output json` is invoked
- **THEN** the validation issues SHALL be JSON-formatted

#### Scenario: Table output

- **GIVEN** a library with 3 validation issues (mixed severities)
- **WHEN** `germinator library validate --output table` is invoked
- **THEN** the output SHALL be a table with columns: severity, type, ref, message
- **AND** each issue SHALL appear as a row

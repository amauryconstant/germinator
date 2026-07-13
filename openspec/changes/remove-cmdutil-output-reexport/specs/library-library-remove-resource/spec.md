# library-library-remove-resource Specification (delta)

## MODIFIED Requirements

### Requirement: library remove resource supports --output flag

The `library remove resource` sub-command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **GIVEN** a library with an existing resource `skill/commit`
- **WHEN** `germinator library remove resource skill/commit` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the removal

#### Scenario: JSON output

- **GIVEN** a library with an existing resource `skill/commit`
- **WHEN** `germinator library remove resource skill/commit --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

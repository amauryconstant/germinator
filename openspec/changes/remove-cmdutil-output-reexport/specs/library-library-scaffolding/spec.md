# library-library-scaffolding Specification (delta)

## MODIFIED Requirements

### Requirement: library init supports --output flag

The `library init` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **GIVEN** a path where no library exists
- **WHEN** `germinator library init --path /tmp/my-library` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the library path created (or the dry-run preview)
- **AND** the output SHALL be written to **stdout** (`opts.IO.Out`) per `internal/output/` stream discipline

#### Scenario: JSON output

- **GIVEN** a path where no library exists
- **WHEN** `germinator library init --path /tmp/my-library --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

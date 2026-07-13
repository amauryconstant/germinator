# library-library-remove-preset Specification (delta)

## MODIFIED Requirements

### Requirement: library remove preset supports --output flag

The `library remove preset` sub-command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags` (inherited from the parent `library remove` command).

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience. The parent `library remove` command continues to wire the `--output` flag via `cmd.PersistentFlags()` (inherited by `resource` and `preset` sub-commands); see the `PersistentFlags wiring for parent commands` requirement in `cli-output-formats`.

#### Scenario: Default plain output

- **GIVEN** a library with an existing preset `git-workflow`
- **WHEN** `germinator library remove preset git-workflow` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the removal

#### Scenario: JSON output

- **GIVEN** a library with an existing preset `git-workflow`
- **WHEN** `germinator library remove preset git-workflow --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

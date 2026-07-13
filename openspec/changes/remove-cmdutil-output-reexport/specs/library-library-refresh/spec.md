# library-library-refresh Specification (delta)

## MODIFIED Requirements

### Requirement: library refresh supports --output flag

The `library refresh` command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **GIVEN** a library with mixed refresh outcomes (some refreshed, some unchanged, some skipped)
- **WHEN** `germinator library refresh` is invoked without `--output`
- **THEN** the output SHALL be plain text with per-resource status sections: Refreshed, Unchanged, Skipped, Errors
- **AND** conflicts (name mismatch, malformed frontmatter) SHALL appear in the Errors section with the reason

#### Scenario: Unchanged resources reported

- **GIVEN** a library with 3 resources that match `library.yaml` exactly
- **WHEN** `germinator library refresh` is invoked
- **THEN** unchanged resources SHALL be listed under the "Unchanged" section

#### Scenario: JSON output

- **GIVEN** a library with mixed refresh outcomes
- **WHEN** `germinator library refresh --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

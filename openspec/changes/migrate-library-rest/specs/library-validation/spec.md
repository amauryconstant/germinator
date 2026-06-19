# library-validation Specification (delta)

## MODIFIED Requirements

### Requirement: library validate follows command-options-pattern

The `library validate` command SHALL adopt the `NewCmdLibraryValidate(f *cmdutil.Factory, runF func(*libraryValidateOptions) error) *cobra.Command` + `runLibraryValidate(opts *libraryValidateOptions) error` template.

#### Scenario: libraryValidateOptions struct

- **WHEN** `cmd/library/validate.go` is inspected
- **THEN** it SHALL declare `libraryValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Fix bool`, `Output string`

### Requirement: library validate supports --output flag

The `library validate` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **WHEN** `germinator library validate` is invoked without `--output`
- **THEN** the output SHALL be plain text with a list of validation issues

#### Scenario: JSON output

- **WHEN** `germinator library validate --output json` is invoked
- **THEN** the validation issues SHALL be JSON-formatted

### Requirement: --fix flag preserved

The `--fix` flag SHALL be preserved and trigger auto-cleanup of `library.yaml` (e.g. removing ghost preset refs, removing entries for missing resource files).

#### Scenario: --fix auto-cleanup

- **WHEN** `germinator library validate --fix` is invoked
- **THEN** the command SHALL auto-clean `library.yaml` (remove ghost preset refs, remove missing entries)
- **AND** the output SHALL indicate what was fixed

#### Scenario: validate without --fix is read-only

- **WHEN** `germinator library validate` is invoked (without `--fix`)
- **THEN** the command SHALL be read-only
- **AND** `library.yaml` SHALL NOT be modified

> **Status:** the `--output` flag is added to `library validate` in change-7 (`migrate-library-rest`). The `--fix` flag is preserved from the legacy implementation.

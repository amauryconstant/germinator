# library-library-validation Specification (delta)

## MODIFIED Requirements

### Requirement: library validate follows command-options-pattern

The `library validate` command SHALL adopt the `NewCmdLibraryValidate(f *cmdutil.Factory, runF func(*libraryValidateOptions) error) *cobra.Command` + `runLibraryValidate(opts *libraryValidateOptions) error` template.

#### Scenario: libraryValidateOptions struct

- **GIVEN** the `library validate` command has been migrated
- **WHEN** `cmd/library_validate.go` is inspected
- **THEN** it SHALL declare `libraryValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Fix bool`, `Output string`

#### Scenario: Library interface method

- **GIVEN** `cmd/library_validate.go` declares its `validatorLibrary` interface
- **WHEN** the interface is inspected
- **THEN** it SHALL declare a `Validate(ctx context.Context, req *ValidateRequest) (*ValidateResult, error)` method
- **AND** the interface SHALL be satisfied directly by `*library.Library` (the `Validate` method is added to `*Library` in change-7)
- **AND** `var _ validatorLibrary = (*library.Library)(nil)` SHALL be a compile-time check at the bottom of `cmd/library_validate.go`
- **AND** `ValidateRequest` SHALL be a type defined in `internal/library/requests.go` with a `Fix bool` field
- **AND** when `req.Fix` is true, the `(*Library).Validate` method SHALL internally call `(*Library).Fix(ctx, &FixRequest{})`; the fix report (`RemovedEntries`, `StrippedRefs`) SHALL be merged into the JSON payload when `--output json` is combined with `--fix`

### Requirement: library validate supports --output flag

The `library validate` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

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

### Requirement: --fix flag preserved

The `--fix` flag SHALL be preserved and trigger auto-cleanup of `library.yaml` (e.g. removing ghost preset refs, removing entries for missing resource files).

#### Scenario: --fix auto-cleanup

- **GIVEN** a library with ghost preset refs and missing entries
- **WHEN** `germinator library validate --fix` is invoked
- **THEN** the command SHALL auto-clean `library.yaml` (remove ghost preset refs, remove missing entries)
- **AND** the output SHALL indicate what was fixed

#### Scenario: --fix with --output json returns machine-readable fix report

- **GIVEN** a library.yaml with 2 missing entries and 1 ghost preset ref
- **WHEN** `germinator library validate --fix --output json` is invoked
- **THEN** the command SHALL auto-clean library.yaml
- **AND** the JSON output SHALL include a `fix` field
- **AND** the `fix` field SHALL enumerate `RemovedEntries` (the 2 missing entries) and `StrippedRefs` (the 1 ghost ref)
- **AND** the output SHALL be valid JSON

#### Scenario: --fix with --output table renders action/ref table

- **GIVEN** a library.yaml with missing entries
- **WHEN** `germinator library validate --fix --output table` is invoked
- **THEN** the output SHALL be a table with columns: action, ref
- **AND** each removed entry / stripped ref SHALL appear as a row

#### Scenario: --fix with no issues is a no-op

- **GIVEN** a clean library with no validation issues
- **WHEN** `germinator library validate --fix` is invoked
- **THEN** the command SHALL NOT modify `library.yaml`
- **AND** the output SHALL indicate "no fixes needed"

#### Scenario: validate without --fix is read-only

- **GIVEN** a library with validation issues
- **WHEN** `germinator library validate` is invoked (without `--fix`)
- **THEN** the command SHALL be read-only
- **AND** `library.yaml` SHALL NOT be modified

> **Status:** the `--output` flag is added to `library validate` in change-7 (`migrate-library-rest`). The `--fix` flag is preserved from the legacy implementation.

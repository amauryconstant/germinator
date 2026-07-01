# library-library-refresh Specification (delta)

## MODIFIED Requirements

### Requirement: library refresh follows command-options-pattern

The `library refresh` command SHALL adopt the `NewCmdRefresh(f *cmdutil.Factory, runF func(*refreshOptions) error) *cobra.Command` + `runRefresh(opts *refreshOptions) error` template.

#### Scenario: refreshOptions struct

- **GIVEN** the `library refresh` command has been migrated
- **WHEN** `cmd/library_refresh.go` is inspected
- **THEN** it SHALL declare `refreshOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `DryRun bool`, `Force bool`, `Output string`

#### Scenario: Library interface method

- **GIVEN** `cmd/library_refresh.go` declares its `refresherLibrary` interface
- **WHEN** the interface is inspected
- **THEN** it SHALL declare a `Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResult, error)` method
- **AND** the interface SHALL be satisfied directly by `*library.Library` (the `Refresh` method is added to `*Library` in change-7, mirroring the slice-6 `(*Library).CreatePreset` precedent)
- **AND** `var _ refresherLibrary = (*library.Library)(nil)` SHALL be a compile-time check at the bottom of `cmd/library_refresh.go`
- **AND** `RefreshRequest` SHALL be a type defined in `internal/library/requests.go` with `DryRun bool` and `Force bool` fields

### Requirement: library refresh supports --output flag

The `library refresh` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **GIVEN** a library with mixed refresh outcomes (some refreshed, some unchanged, some skipped)
- **WHEN** `germinator library refresh` is invoked without `--output`
- **THEN** the output SHALL be plain text with per-resource status sections: Refreshed, Unchanged, Skipped, Errors
- **AND** conflicts (name mismatch, malformed frontmatter) SHALL appear in the Errors section with the reason

#### Scenario: Unchanged resources reported

- **GIVEN** a library with 3 resources that match `library.yaml` exactly
- **WHEN** `germinator library refresh` is invoked
- **THEN** the output SHALL include an `Unchanged:` section listing the 3 resources

#### Scenario: JSON output

- **GIVEN** a library with refresh activity
- **WHEN** `germinator library refresh --output json` is invoked
- **THEN** the per-resource status SHALL be JSON-formatted
- **AND** the payload SHALL include `refreshed`, `unchanged`, `skipped`, and `errors` arrays

#### Scenario: Table output

- **GIVEN** a library with 3 refreshed resources
- **WHEN** `germinator library refresh --output table` is invoked
- **THEN** the output SHALL be a table with columns: ref, field, old, new
- **AND** each refreshed resource SHALL appear as a row

#### Scenario: Dry-run mode

- **GIVEN** a library with resources that would be refreshed
- **WHEN** `germinator library refresh --dry-run` is invoked
- **THEN** the command SHALL preview changes without modifying `library.yaml`

> **Status:** the `--output` flag is added to `library refresh` in change-7 (`migrate-library-rest`). The legacy `--json` flag (if any) is replaced.

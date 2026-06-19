# library-refresh Specification (delta)

## MODIFIED Requirements

### Requirement: library refresh follows command-options-pattern

The `library refresh` command SHALL adopt the `NewCmdRefresh(f *cmdutil.Factory, runF func(*refreshOptions) error) *cobra.Command` + `runRefresh(opts *refreshOptions) error` template.

#### Scenario: refreshOptions struct

- **WHEN** `cmd/library/refresh.go` is inspected
- **THEN** it SHALL declare `refreshOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `DryRun bool`, `Force bool`, `Output string`

### Requirement: library refresh supports --output flag

The `library refresh` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **WHEN** `germinator library refresh` is invoked without `--output`
- **THEN** the output SHALL be plain text with per-resource status (updated, unchanged, conflict)

#### Scenario: JSON output

- **WHEN** `germinator library refresh --output json` is invoked
- **THEN** the per-resource status SHALL be JSON-formatted

#### Scenario: Dry-run mode

- **WHEN** `germinator library refresh --dry-run` is invoked
- **THEN** the command SHALL preview changes without modifying `library.yaml`

> **Status:** the `--output` flag is added to `library refresh` in change-7 (`migrate-library-rest`). The legacy `--json` flag (if any) is replaced.

# Capability: Library Refresh

## Purpose

The Library Refresh capability synchronizes metadata from registered resource files into `library.yaml`. It handles description drift after manual edits and path updates after file renames, while maintaining safety through conflict detection.

## Requirements

### Requirement: Refresh library metadata

The system SHALL sync metadata from registered resource files into library.yaml.

#### Scenario: Refresh updates stale description
- **GIVEN** a library with resource `skill/commit` whose library.yaml description is "old description"
- **WHEN** Refresh is called and the file `skills/commit.md` has frontmatter `description: new description`
- **THEN** library.yaml entry for `skill/commit` is updated to `description: new description`

#### Scenario: Refresh skips unchanged description
- **GIVEN** a library with resource `skill/commit` whose library.yaml description matches frontmatter
- **WHEN** Refresh is called
- **THEN** no update is made and the resource is not reported as changed

#### Scenario: Refresh discovers renamed file by searching directory
- **GIVEN** a library with resource `skill/commit` at path `skills/commit.md`
- **AND** the file has been renamed to `skills/commit-msg.md` with frontmatter `name: commit`
- **WHEN** Refresh is called
- **THEN** Refresh searches the `skills/` directory for a file with frontmatter `name: commit`
- **AND** the library.yaml entry path is updated to `skills/commit-msg.md`

#### Scenario: Refresh updates both path and description
- **GIVEN** a library with resource `skill/commit` at path `skills/commit.md` with description "old description"
- **AND** the file has been renamed to `skills/commit-msg.md` with frontmatter `name: commit` and `description: new description`
- **WHEN** Refresh is called
- **THEN** library.yaml entry path is updated to `skills/commit-msg.md`
- **AND** library.yaml entry description is updated to `new description`

### Requirement: Handle missing files gracefully

The system SHALL skip resources whose files are missing during refresh.

#### Scenario: Refresh skips missing file
- **GIVEN** a library with resource `skill/commit` at path `skills/commit.md`
- **AND** the file does not exist
- **WHEN** Refresh is called
- **THEN** the resource is skipped (not updated, not errored)

### Requirement: Detect name mismatch conflicts

The system SHALL error when frontmatter name does not match entry key.

#### Scenario: Refresh errors on name mismatch
- **GIVEN** a library with resource `skill/commit` at path `skills/old.md`
- **AND** the file at `skills/old.md` has been moved to `skills/new.md` with frontmatter `name: new`
- **WHEN** Refresh is called
- **THEN** an error is recorded for `skill/commit` with reason `name_mismatch`
- **AND** the resource is skipped

### Requirement: Handle malformed frontmatter

The system SHALL skip resources with malformed frontmatter and record errors.

#### Scenario: Refresh errors on malformed frontmatter
- **GIVEN** a library with resource `skill/commit` at path `skills/commit.md`
- **AND** the file has invalid YAML frontmatter
- **WHEN** Refresh is called
- **THEN** an error is recorded for `skill/commit` with reason `malformed_frontmatter`
- **AND** the resource is skipped

### Requirement: Collect all errors and continue

The system SHALL process all resources and collect errors rather than failing on first error.

#### Scenario: Refresh continues after error
- **GIVEN** a library with multiple resources, some with errors
- **WHEN** Refresh is called
- **THEN** all resources are processed
- **AND** all errors are collected and reported at end
- **AND** exit code is 1 if any errors occurred

### Requirement: Support dry-run mode

The system SHALL preview changes without modifying library.yaml in dry-run mode.

#### Scenario: Dry-run shows what would change
- **GIVEN** a library with resources needing updates
- **WHEN** Refresh is called with `--dry-run`
- **THEN** no files are modified
- **AND** the expected changes are reported

### Requirement: Support force mode to skip conflicts

The system SHALL skip resources with conflicts when `--force` is specified.

#### Scenario: Force skips name mismatch
- **GIVEN** a library with resource `skill/commit` with name mismatch conflict
- **WHEN** Refresh is called with `--force`
- **THEN** the conflicting resource is skipped
- **AND** no error is recorded for that resource

### Requirement: Discover library path

The system SHALL discover library path via flag, environment, or default.

#### Scenario: Discover library from --library flag
- **GIVEN** `--library=/custom/path` flag
- **WHEN** Refresh is called
- **THEN** the library at `/custom/path` is used

#### Scenario: Discover library from environment
- **GIVEN** `GERMINATOR_LIBRARY=/env/path` env var and no flag
- **WHEN** Refresh is called
- **THEN** the library at `/env/path` is used

#### Scenario: Discover library from default
- **GIVEN** no `--library` flag and no `GERMINATOR_LIBRARY` env
- **WHEN** Refresh is called
- **THEN** `~/.config/germinator/library/` is used

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

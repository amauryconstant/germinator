# Capability: Library Refresh

## Purpose

The Library Refresh capability synchronizes metadata from registered resource files into `library.yaml`. It handles description drift after manual edits and path updates after file renames, while maintaining safety through conflict detection.

## ADDED Requirements

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

### Requirement: Output machine-readable JSON

The system SHALL output refresh results as JSON when `--json` flag is specified.

#### Scenario: JSON output format
- **GIVEN** a refresh operation with some updates and errors
- **WHEN** Refresh is called with `--json`
- **THEN** JSON output includes `refreshed`, `skipped`, and `errors` fields
- **AND** each refreshed item includes `ref`, `field`, `old`, `new`
- **AND** each skipped item includes `ref` and `reason`

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

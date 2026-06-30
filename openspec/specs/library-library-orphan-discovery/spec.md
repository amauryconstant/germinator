# Capability: Library Orphan Discovery

## Purpose

The Library Orphan Discovery capability finds resource files on disk that are not registered in `library.yaml`. It enables users to discover and optionally register orphaned files that were added via file manager, git mv, or other direct filesystem operations.

## Requirements

### Requirement: Discover orphaned resource files

The system SHALL scan library directories for files not registered in library.yaml.

#### Scenario: Discover orphans in skills directory
- **GIVEN** a library with no `skill/commit` registered
- **AND** a file `skills/commit.md` exists with valid frontmatter
- **WHEN** AddResource is called with `--discover`
- **THEN** the file is reported as an orphan with type `skill`, name `commit`

#### Scenario: Discover orphans in all resource directories
- **GIVEN** a library with orphan files in `agents/`, `commands/`, `memory/` directories
- **WHEN** AddResource is called with `--discover`
- **THEN** orphans from all directories are reported

#### Scenario: No orphans when all files registered
- **GIVEN** a library where all files in `skills/`, `agents/`, `commands/`, `memory/` are registered
- **WHEN** AddResource is called with `--discover`
- **THEN** no orphans are reported

#### Scenario: Discover orphans recursively in skills directory
- **GIVEN** a library with no `skill/nested/deep` registered
- **AND** a file `skills/nested/deep.md` exists with valid frontmatter
- **WHEN** AddResource is called with `--discover`
- **THEN** the file is reported as an orphan with type `skill`, name `nested/deep`

#### Scenario: Discover orphans recursively in all resource directories
- **GIVEN** a library with orphan files in nested directories: `agents/team/reviewer.md`, `commands/git/commit.md`, `memory/notes/todo.md`
- **WHEN** AddResource is called with `--discover`
- **THEN** all orphans from all directories (including nested) are reported

#### Scenario: Discover orphans in deeply nested structure
- **GIVEN** a library with orphan files at multiple nesting levels: `skills/sub1/skill1.md`, `skills/sub1/sub2/skill2.md`
- **WHEN** AddResource is called with `--discover`
- **THEN** both `skill1` and `skill2` are reported as orphans

### Requirement: Detect orphan type from directory

The system SHALL determine resource type from the directory containing the file.

#### Scenario: Orphan type from skills directory
- **GIVEN** a file `skills/new-skill.md` not in library.yaml
- **WHEN** AddResource is called with `--discover`
- **THEN** the orphan type is detected as `skill`

#### Scenario: Orphan type from agents directory
- **GIVEN** a file `agents/reviewer.md` not in library.yaml
- **WHEN** AddResource is called with `--discover`
- **THEN** the orphan type is detected as `agent`

### Requirement: Detect orphan name from frontmatter or filename

The system SHALL detect orphan name from frontmatter `name` field or derive from filename.

#### Scenario: Detect orphan name from frontmatter
- **GIVEN** a file `skills/skill.md` with frontmatter `name: my-skill`
- **WHEN** AddResource is called with `--discover`
- **THEN** the orphan name is `my-skill`

#### Scenario: Detect orphan name from filename when no frontmatter
- **GIVEN** a file `skills/commit.md` with no `name` in frontmatter
- **WHEN** AddResource is called with `--discover`
- **THEN** the orphan name is derived from filename as `commit`

### Requirement: Detect orphan description from frontmatter

The system SHALL detect orphan description from frontmatter `description` field.

#### Scenario: Detect orphan description from frontmatter
- **GIVEN** a file `skills/skill.md` with frontmatter `description: My skill description`
- **WHEN** AddResource is called with `--discover`
- **THEN** the orphan description is `My skill description`

#### Scenario: Orphan without description
- **GIVEN** a file `skills/skill.md` with no `description` in frontmatter
- **WHEN** AddResource is called with `--discover`
- **THEN** the orphan description is empty string

### Requirement: Report-only mode by default

The system SHALL report orphans without modifying library.yaml unless `--force` is specified.

#### Scenario: Report-only shows orphans
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called with `--discover` (without `--force`)
- **THEN** the orphans are reported
- **AND** library.yaml is not modified

### Requirement: Force mode registers orphans

The system SHALL register orphaned files in library.yaml when `--force` is specified.

#### Scenario: Force registers orphan
- **GIVEN** a library with orphan file `skills/new-skill.md`
- **WHEN** AddResource is called with `--discover --force`
- **THEN** the file is registered in library.yaml
- **AND** the orphan is reported as added

### Requirement: Support dry-run with discover

The system SHALL preview orphan registration without modifying library.yaml in dry-run mode.

#### Scenario: Discover dry-run shows what would happen
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called with `--discover --dry-run`
- **THEN** no files are modified
- **AND** the expected additions are described

### Requirement: Conflict detection produces typed OperationError

The `library add --discover` command SHALL detect when an orphan has the same name as an existing resource. The conflict SHALL be reported as a `*core.OperationError{Op: "register", Resource: <ref>, Cause: <origErr>}` per file, aggregated into a `*core.PartialSuccessError`, and counted toward `Failed`. The conflict SHALL be rendered to stderr via `output.FormatError` per file.

> Replaces the pre-change behavior where `ConflictInfo{Issue: "name_conflict"}` was carried as a string field on `DiscoverResult`.

#### Scenario: Detect name conflict produces OperationError

- **GIVEN** a library with existing resource `skill/commit`
- **AND** an orphan file `skills/commit.md` not in library.yaml
- **WHEN** `germinator library add --discover --force` is invoked
- **THEN** a `*core.OperationError{Op: "register", Resource: "skill/commit", Cause: <origErr>}` SHALL be produced for the file
- **AND** the OperationError SHALL be aggregated into the partial-success result with `Failed` incremented by 1
- **AND** `output.FormatError` SHALL render `Error: register: skill/commit\n` to **stderr** (`opts.IO.ErrOut`)
- **AND** the orphan SHALL NOT be registered (the file is left untouched)

#### Scenario: Name conflict counts as failure, not success

- **GIVEN** a library with 2 orphans: `skills/orphan1.md` (valid, no conflict) and `skills/orphan2.md` (conflicts with existing `skill/orphan2`)
- **WHEN** `germinator library add --discover --force --batch` is invoked
- **THEN** the partial-success aggregate SHALL have `Succeeded == 1` and `Failed == 1`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeSuccess` (0) because `Succeeded > 0`
- **AND** stdout SHALL contain the success listing for `skill/orphan1`
- **AND** stderr SHALL contain `Error: register: skill/orphan2` from the per-file FormatError render

#### Scenario: All conflicts returns exit 1

- **GIVEN** a library with 2 orphans, both conflicting with existing resources
- **WHEN** `germinator library add --discover --force --batch` is invoked
- **THEN** the partial-success aggregate SHALL have `Succeeded == 0` and `Failed == 2`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeError` (1)
- **AND** stdout SHALL be empty (no data leakage on error paths)
- **AND** stderr SHALL contain the two `Error: register: ...` lines

#### Scenario: OperationError preserves wrapped cause

- **GIVEN** a name conflict where the library package returns a typed `library.ErrNameConflict` as the underlying cause
- **WHEN** the conflict is reported as `*core.OperationError{Op: "register", Resource: <ref>, Cause: library.ErrNameConflict}`
- **THEN** `errors.Is(err, library.ErrNameConflict)` SHALL be `true`
- **AND** `errors.Unwrap(err)` SHALL return the cause
- **AND** `output.FormatError` SHALL render both the typed error message and the cause on separate lines

### Requirement: Discover library path

The system SHALL discover library path via flag, environment, or default.

#### Scenario: Discover library from --library flag
- **GIVEN** `--library=/custom/path` flag
- **WHEN** AddResource is called with `--discover`
- **THEN** the library at `/custom/path` is used

#### Scenario: Discover library from environment
- **GIVEN** `GERMINATOR_LIBRARY=/env/path` env var and no flag
- **WHEN** AddResource is called with `--discover`
- **THEN** the library at `/env/path` is used

#### Scenario: Discover library from default
- **GIVEN** no `--library` flag and no `GERMINATOR_LIBRARY` env
- **WHEN** AddResource is called with `--discover`
- **THEN** `~/.config/germinator/library/` is used

### Requirement: Enhanced discover result structure

The system SHALL return a DiscoverResult with comprehensive information for batch integration.

#### Scenario: Discover result contains orphans with path information
- **GIVEN** a library with orphan file `skills/commit.md`
- **WHEN** AddResource is called with `--discover`
- **THEN** the result includes orphan with Path, Type, Name, and optional Issue fields

#### Scenario: Discover result contains summary statistics
- **GIVEN** a library with 5 `.md` files scanned and 3 orphans found
- **WHEN** AddResource is called with `--discover`
- **THEN** the result Summary includes TotalScanned=5 and TotalOrphans=3
#### Scenario: Discover result tracks added resources

- **GIVEN** a library with orphan `skills/new-skill.md`
- **WHEN** AddResource is called with `--discover --force`
- **THEN** the result Added contains the successfully registered orphan

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

### Requirement: Conflict detection for duplicate names

The system SHALL detect when an orphan has the same name as an existing resource.

#### Scenario: Detect name conflict with existing resource
- **GIVEN** a library with existing resource `skill/commit`
- **AND** an orphan file `skills/commit.md` not in library.yaml
- **WHEN** AddResource is called with `--discover --force`
- **THEN** an error is reported for the conflict
- **AND** the orphan is not registered

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

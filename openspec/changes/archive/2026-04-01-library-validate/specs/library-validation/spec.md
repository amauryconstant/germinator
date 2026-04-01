# Capability: Library Validation

## Purpose

The Library Validation capability provides integrity checking for library metadata in `library.yaml` against the actual filesystem state. It detects inconsistencies and optionally auto-repairs them.

## ADDED Requirements

### Requirement: Detect missing resource files

The system SHALL detect when entries in `library.yaml` reference files that do not exist on disk.

#### Scenario: Detect missing file for skill resource
- **GIVEN** a library.yaml entry `skill/commit` with path `skills/commit.md`
- **AND** the file `skills/commit.md` does not exist
- **WHEN** library validation runs
- **THEN** a missing-file issue is reported with ref `skill/commit` and path `skills/commit.md`
- **AND** severity is `error`

#### Scenario: Detect missing file for agent resource
- **GIVEN** a library.yaml entry `agent/reviewer` with path `agents/reviewer.md`
- **AND** the file `agents/reviewer.md` does not exist
- **WHEN** library validation runs
- **THEN** a missing-file issue is reported with ref `agent/reviewer` and path `agents/reviewer.md`
- **AND** severity is `error`

### Requirement: Detect ghost preset resources

The system SHALL detect when presets reference resources that do not exist in the library.

#### Scenario: Detect ghost resource in preset
- **GIVEN** a preset `git-workflow` with resources `[skill/commit, skill/ghost]`
- **AND** resource `skill/ghost` does not exist in the library
- **WHEN** library validation runs
- **THEN** a ghost-resource issue is reported with ref `skill/ghost` and inPreset `git-workflow`
- **AND** severity is `error`

#### Scenario: Detect multiple ghost resources in same preset
- **GIVEN** a preset `dev-setup` with resources `[skill/build, agent/missing, command/test]`
- **AND** resources `agent/missing` and `command/test` do not exist
- **WHEN** library validation runs
- **THEN** two ghost-resource issues are reported
- **AND** each references the preset `dev-setup`

### Requirement: Detect orphaned files

The system SHALL detect when resource files exist on disk but are not registered in `library.yaml`.

#### Scenario: Detect orphaned file in skills directory
- **GIVEN** a file `skills/extra.md` exists on disk
- **AND** no entry in library.yaml references `skills/extra.md`
- **WHEN** library validation runs
- **THEN** an orphan issue is reported with path `skills/extra.md`
- **AND** severity is `warning`

#### Scenario: Detect orphaned file in agents directory
- **GIVEN** a file `agents/legacy.md` exists on disk
- **AND** no entry in library.yaml references `agents/legacy.md`
- **WHEN** library validation runs
- **THEN** an orphan issue is reported with path `agents/legacy.md`
- **AND** severity is `warning`

#### Scenario: Detect multiple orphaned files
- **GIVEN** files `commands/old.md` and `memory/notes.md` exist on disk
- **AND** neither is registered in library.yaml
- **WHEN** library validation runs
- **THEN** two orphan issues are reported

### Requirement: Detect malformed frontmatter

The system SHALL detect when resource files have invalid or unparseable YAML frontmatter.

#### Scenario: Detect malformed YAML in frontmatter
- **GIVEN** a resource file `skills/commit.md` exists
- **AND** the file has frontmatter that is not valid YAML
- **WHEN** library validation runs
- **THEN** a malformed-frontmatter issue is reported with path `skills/commit.md`
- **AND** severity is `error`

#### Scenario: Valid frontmatter passes validation
- **GIVEN** a resource file `skills/commit.md` exists
- **AND** the file has valid YAML frontmatter with required fields
- **WHEN** library validation runs
- **THEN** no malformed-frontmatter issue is reported for that file

### Requirement: Report validation summary

The system SHALL report a summary of validation results including issue counts by severity.

#### Scenario: Report clean library
- **GIVEN** a library with no issues
- **WHEN** library validation runs
- **THEN** the output indicates the library is valid
- **AND** summary shows `errors: 0, warnings: 0`

#### Scenario: Report issues summary
- **GIVEN** a library with 2 errors and 3 warnings
- **WHEN** library validation runs
- **THEN** the output indicates the library has issues
- **AND** summary shows `errors: 2, warnings: 3`

### Requirement: Fix library metadata automatically

The system SHALL provide a `--fix` flag that auto-cleans library.yaml metadata.

#### Scenario: Fix removes missing file entries
- **GIVEN** a library.yaml entry `skill/commit` with missing file `skills/commit.md`
- **WHEN** library validate --fix is run
- **THEN** the entry `skill/commit` is removed from library.yaml
- **AND** the file `skills/commit.md` is NOT deleted

#### Scenario: Fix strips ghost resources from presets
- **GIVEN** a preset `git-workflow` with resources `[skill/commit, skill/ghost]`
- **AND** resource `skill/ghost` does not exist
- **WHEN** library validate --fix is run
- **THEN** `skill/ghost` is removed from the preset's resources
- **AND** `skill/commit` remains

#### Scenario: Fix does not delete orphaned files
- **GIVEN** an orphaned file `agents/extra.md` exists
- **WHEN** library validate --fix is run
- **THEN** the file `agents/extra.md` is NOT deleted
- **AND** the orphan is NOT reported as fixed (conservative fix only)

#### Scenario: Fix skips malformed frontmatter
- **GIVEN** a resource file `skills/commit.md` with malformed frontmatter
- **WHEN** library validate --fix is run
- **THEN** no change is made to the file
- **AND** a malformed-frontmatter issue is still reported

### Requirement: Support JSON output

The system SHALL provide a `--json` flag for machine-readable output.

#### Scenario: JSON output format
- **GIVEN** a library with issues
- **WHEN** library validate --json is run
- **THEN** output is valid JSON
- **AND** contains `valid`, `summary`, and `issues` fields
- **AND** each issue has `type`, `severity`, and relevant fields (`ref`, `path`, `inPreset`)

#### Scenario: JSON output for clean library
- **GIVEN** a library with no issues
- **WHEN** library validate --json is run
- **THEN** output is valid JSON
- **AND** `valid` is true
- **AND** `issues` is an empty array

### Requirement: Exit codes reflect validation status

The system SHALL return exit codes that reflect validation results.

#### Scenario: Clean library returns exit 0
- **GIVEN** a library with no issues
- **WHEN** library validate is run
- **THEN** exit code is 0

#### Scenario: Errors found returns exit 5
- **GIVEN** a library with error-level issues (missing-file, ghost-resource, or malformed)
- **WHEN** library validate is run
- **THEN** exit code is 5

#### Scenario: Warnings only returns exit 0
- **GIVEN** a library with only warning-level issues (orphan)
- **WHEN** library validate is run
- **THEN** exit code is 0

#### Scenario: Errors after --fix returns exit 5
- **GIVEN** a library with error-level issues that cannot be auto-fixed
- **WHEN** library validate --fix is run
- **THEN** exit code is 5

### Requirement: Discover library path for validation

The system SHALL discover the library path using the same priority as other library commands.

#### Scenario: Discover library path from flag
- **GIVEN** flag `--library=/custom/path`
- **WHEN** library validate is run
- **THEN** the library at `/custom/path` is validated

#### Scenario: Discover library path from environment
- **GIVEN** no `--library` flag and env `GERMINATOR_LIBRARY=/env/path`
- **WHEN** library validate is run
- **THEN** the library at `/env/path` is validated

#### Scenario: Discover library path from default
- **GIVEN** no `--library` flag and no `GERMINATOR_LIBRARY` env
- **WHEN** library validate is run
- **THEN** the default library path is used

### Requirement: Human-readable output format

The system SHALL provide human-readable output by default.

#### Scenario: Human-readable shows issue list
- **GIVEN** a library with issues
- **WHEN** library validate is run (no --json flag)
- **THEN** output is human-readable text
- **AND** each issue is displayed with severity, ref/path, and description

#### Scenario: Human-readable shows fix hint
- **GIVEN** a library with issues
- **WHEN** library validate is run (no --json flag)
- **THEN** output includes hint about `--fix` and `--json` flags

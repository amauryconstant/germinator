# Capability: Library Scaffolding

## Purpose

The Library Scaffolding capability enables users to create new library directory structures with valid `library.yaml` manifests and empty resource directories via the `germinator library init` command.

## Requirements

### Requirement: Create library directory structure

The system SHALL create a new library directory at the specified path with a valid `library.yaml` and empty resource directories.

#### Scenario: Create library at specified path
- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library` is executed
- **THEN** directory `/tmp/my-library/` is created
- **AND** file `/tmp/my-library/library.yaml` is created with version "1" and empty resources/presets
- **AND** directory `/tmp/my-library/skills/` is created
- **AND** directory `/tmp/my-library/agents/` is created
- **AND** directory `/tmp/my-library/commands/` is created
- **AND** directory `/tmp/my-library/memory/` is created

#### Scenario: Create library at default path
- **GIVEN** no library exists at `~/.local/share/germinator/library/`
- **WHEN** `germinator library init` is executed
- **THEN** library is created at `~/.local/share/germinator/library/`
- **OR** if `XDG_DATA_HOME` is set, library is created at `$XDG_DATA_HOME/germinator/library/`

### Requirement: Validate created library

The system SHALL validate created library by loading it via `LoadLibrary` to ensure structural correctness.

#### Scenario: Validate successful creation
- **GIVEN** library creation at `/tmp/my-library` succeeds
- **WHEN** `LoadLibrary("/tmp/my-library")` is called
- **THEN** a valid Library struct is returned
- **AND** `Library.Resources` contains empty maps for skill, agent, command, and memory
- **AND** `Library.Presets` is an empty map
- **AND** `Library.Version` equals "1"

### Requirement: Handle existing library

The system SHALL return an error if a library already exists at the target path unless `--force` is specified.

#### Scenario: Error when library exists
- **GIVEN** a library already exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library` is executed
- **THEN** an error is returned indicating the library already exists

#### Scenario: Overwrite with force flag
- **GIVEN** a library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --force` is executed
- **THEN** the library is replaced with a new empty library

### Requirement: Support dry-run mode

The system SHALL preview changes without creating files when `--dry-run` is specified.

#### Scenario: Dry-run does not create files
- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --dry-run` is executed
- **THEN** no files or directories are created
- **AND** a message is printed indicating what would be created

### Requirement: Create valid library.yaml content

The system SHALL create a `library.yaml` with version "1" and empty resource/preset sections.

#### Scenario: Library.yaml has correct structure
- **GIVEN** library is created at `/tmp/my-library`
- **WHEN** the `library.yaml` file is read
- **THEN** it parses as valid YAML
- **AND** it contains `version: "1"`
- **AND** it contains `resources:` with skill, agent, command, and memory types all empty maps
- **AND** it contains `presets:` as an empty map

#### Scenario: Permissions denied when creating directories
- **GIVEN** user lacks permission to create directories at `/protected/`
- **WHEN** `germinator library init --path /protected/library` is executed
- **THEN** an error is returned indicating permission was denied

#### Scenario: Disk full when writing files
- **GIVEN** disk space is exhausted at target path
- **WHEN** `germinator library init --path /full-disk/library` is executed
- **THEN** an error is returned indicating write failure

#### Scenario: Invalid path characters
- **GIVEN** path contains invalid characters for the filesystem
- **WHEN** `germinator library init --path "/tmp/invalid\x00path"` is executed
- **THEN** an error is returned indicating invalid path

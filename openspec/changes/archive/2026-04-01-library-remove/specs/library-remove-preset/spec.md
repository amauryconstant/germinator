# library-remove-preset

## ADDED Requirements (11)

### Requirement: Remove preset command exists

The system SHALL provide a `library remove preset <name>` command that removes a preset from the library.

#### Scenario: Remove preset subcommand is registered
- **WHEN** the user runs `germinator library remove --help`
- **THEN** the help output includes `preset` as a subcommand

### Requirement: Remove preset requires name

The system SHALL require a preset name as a non-empty string.

#### Scenario: Remove with valid name
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** the system attempts to remove the preset

#### Scenario: Remove with empty name
- **WHEN** user runs `germinator library remove preset ""`
- **THEN** the system returns an error about missing preset name

### Requirement: Remove preset validates library exists

The system SHALL error if the library does not exist at the specified path.

#### Scenario: Remove from non-existent library
- **WHEN** user runs `germinator library remove preset git-workflow --library /nonexistent`
- **THEN** the system returns an error that the library does not exist

### Requirement: Remove preset validates preset exists

The system SHALL error if the specified preset does not exist in the library.

#### Scenario: Remove non-existent preset
- **WHEN** user runs `germinator library remove preset nonexistent`
- **AND** the library does not contain a preset named `nonexistent`
- **THEN** the system returns an error: `preset not found`

### Requirement: Remove preset updates library.yaml

The system SHALL remove the preset entry from `library.yaml`.

#### Scenario: Update library.yaml on remove
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** the `presets.git-workflow` entry is removed from `library.yaml`

### Requirement: Remove preset has no physical file

The system SHALL NOT attempt to delete any physical files when removing a preset (presets exist only in library.yaml).

#### Scenario: No physical file deletion
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** no file outside of `library.yaml` is modified or deleted

### Requirement: Remove preset validates after update

The system SHALL validate the updated library.yaml after removal.

#### Scenario: Validation of updated library
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** the system loads the updated library to validate it is well-formed

### Requirement: Remove preset outputs success message

The system SHALL print a human-readable success message when removal succeeds.

#### Scenario: Success output
- **WHEN** user runs `germinator library remove preset git-workflow`
- **AND** removal succeeds
- **THEN** the system prints: `Removed preset: git-workflow`

### Requirement: Remove preset supports --json flag

The system SHALL support a `--json` flag that outputs structured JSON instead of human-readable text.

#### Scenario: JSON output format
- **WHEN** user runs `germinator library remove preset git-workflow --json`
- **THEN** the output is JSON with fields: `type`, `name`, `resourcesRemoved`

#### Scenario: JSON output example
- **WHEN** user runs `germinator library remove preset git-workflow --json`
- **AND** the preset contained resources `["skill/commit", "skill/pr"]`
- **THEN** the output is:
```json
{
  "type": "preset",
  "name": "git-workflow",
  "resourcesRemoved": ["skill/commit", "skill/pr"]
}
```

### Requirement: Remove preset --json error format

When `--json` is specified and an error occurs, the system SHALL return JSON with an `error` field.

#### Scenario: JSON error format
- **WHEN** user runs `germinator library remove preset nonexistent --json`
- **AND** the preset does not exist
- **THEN** the output is:
```json
{
  "error": "preset not found",
  "type": "preset",
  "name": "nonexistent"
}
```

### Requirement: Remove preset uses library path resolution

The system SHALL resolve the library path using the same priority as other library commands: explicit `--library` flag > `GERMINATOR_LIBRARY` env > default path.

#### Scenario: Library path from default
- **WHEN** user runs `germinator library remove preset git-workflow`
- **AND** no explicit library path is provided
- **THEN** the system uses the default library path

#### Scenario: Library path from environment
- **WHEN** `GERMINATOR_LIBRARY` is set to `/custom/path`
- **AND** user runs `germinator library remove preset git-workflow`
- **THEN** the system uses `/custom/path`

#### Scenario: Library path from flag
- **WHEN** user runs `germinator library remove preset git-workflow --library /custom/path`
- **THEN** the system uses `/custom/path`

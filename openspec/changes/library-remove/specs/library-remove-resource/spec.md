# library-remove-resource

## ADDED Requirements (12)

### Requirement: Remove resource command exists

The system SHALL provide a `library remove resource <ref>` command that removes a resource from the library.

#### Scenario: Remove resource subcommand is registered
- **WHEN** the user runs `germinator library remove --help`
- **THEN** the help output includes `resource` as a subcommand

### Requirement: Remove resource requires valid reference

The system SHALL require a resource reference in `type/name` format (e.g., `skill/commit`).

#### Scenario: Remove with valid reference
- **WHEN** user runs `germinator library remove resource skill/commit`
- **THEN** the system attempts to remove the resource

#### Scenario: Remove with invalid reference format
- **WHEN** user runs `germinator library remove resource invalid`
- **THEN** the system returns an error: `invalid resource reference format (expected type/name)`

#### Scenario: Remove with empty reference
- **WHEN** user runs `germinator library remove resource ""`
- **THEN** the system returns an error about missing resource reference

### Requirement: Remove resource validates library exists

The system SHALL error if the library does not exist at the specified path.

#### Scenario: Remove from non-existent library
- **WHEN** user runs `germinator library remove resource skill/commit --library /nonexistent`
- **THEN** the system returns an error that the library does not exist

### Requirement: Remove resource validates resource exists

The system SHALL error if the specified resource does not exist in the library.

#### Scenario: Remove non-existent resource
- **WHEN** user runs `germinator library remove resource skill/nonexistent`
- **AND** the library does not contain `skill/nonexistent`
- **THEN** the system returns an error: `resource skill/nonexistent not found`

### Requirement: Remove resource checks preset references

The system SHALL error if any preset in the library references the resource being removed.

#### Scenario: Remove resource referenced by preset
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** a preset `git-workflow` contains `skill/commit` in its resources
- **THEN** the system returns an error: `resource skill/commit is referenced by preset git-workflow`

#### Scenario: Remove resource not referenced by any preset
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** no presets reference `skill/commit`
- **THEN** the system proceeds with removal

### Requirement: Remove resource deletes physical file

The system SHALL delete the physical resource file from the library directory.

#### Scenario: Delete physical file on remove
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** the resource exists at `library/skills/commit.md`
- **THEN** the file `library/skills/commit.md` is deleted

### Requirement: Remove resource updates library.yaml

The system SHALL remove the resource entry from `library.yaml`.

#### Scenario: Update library.yaml on remove
- **WHEN** user runs `germinator library remove resource skill/commit`
- **THEN** the `resources.skill.commit` entry is removed from `library.yaml`

### Requirement: Remove resource validates after update

The system SHALL validate the updated library.yaml after removal.

#### Scenario: Validation of updated library
- **WHEN** user runs `germinator library remove resource skill/commit`
- **THEN** the system loads the updated library to validate it is well-formed

### Requirement: Remove resource outputs success message

The system SHALL print a human-readable success message when removal succeeds.

#### Scenario: Success output
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** removal succeeds
- **THEN** the system prints: `Removed resource: skill/commit`

### Requirement: Remove resource supports --json flag

The system SHALL support a `--json` flag that outputs structured JSON instead of human-readable text.

#### Scenario: JSON output format
- **WHEN** user runs `germinator library remove resource skill/commit --json`
- **THEN** the output is JSON with fields: `type`, `resourceType`, `name`, `fileDeleted`, `libraryPath`

#### Scenario: JSON output example
- **WHEN** user runs `germinator library remove resource skill/commit --json`
- **THEN** the output is:
```json
{
  "type": "resource",
  "resourceType": "skill",
  "name": "commit",
  "fileDeleted": "skills/commit.md",
  "libraryPath": "/path/to/library"
}
```

### Requirement: Remove resource --json error format

When `--json` is specified and an error occurs, the system SHALL return JSON with an `error` field.

#### Scenario: JSON error format
- **WHEN** user runs `germinator library remove resource skill/nonexistent --json`
- **AND** the resource does not exist
- **THEN** the output is:
```json
{
  "error": "resource not found",
  "type": "resource",
  "resourceType": "skill",
  "name": "nonexistent"
}
```

### Requirement: Remove resource uses library path resolution

The system SHALL resolve the library path using the same priority as other library commands: explicit `--library` flag > `GERMINATOR_LIBRARY` env > default path.

#### Scenario: Library path from default
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** no explicit library path is provided
- **THEN** the system uses the default library path

#### Scenario: Library path from environment
- **WHEN** `GERMINATOR_LIBRARY` is set to `/custom/path`
- **AND** user runs `germinator library remove resource skill/commit`
- **THEN** the system uses `/custom/path`

#### Scenario: Library path from flag
- **WHEN** user runs `germinator library remove resource skill/commit --library /custom/path`
- **THEN** the system uses `/custom/path`

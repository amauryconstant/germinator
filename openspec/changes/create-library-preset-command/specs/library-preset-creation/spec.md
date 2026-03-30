# Capability: Library Preset Creation

## Purpose

The Library Preset Creation capability enables creating new presets in the library through the CLI. It validates resource references, persists changes to library.yaml, and supports overwrite with --force flag.

## Requirements

### Requirement: Create preset with valid resources

The system SHALL create a preset when all referenced resources exist in the library.

#### Scenario: Create preset with single resource
- **GIVEN** a library with resource `skill/commit`
- **WHEN** CreatePreset is called with name `commit-tools` and resources `["skill/commit"]`
- **THEN** the preset is added to library.yaml with name `commit-tools` and resources `["skill/commit"]`

#### Scenario: Create preset with multiple resources
- **GIVEN** a library with resources `skill/commit` and `skill/merge-request`
- **WHEN** CreatePreset is called with name `git-workflow` and resources `["skill/commit", "skill/merge-request"]`
- **THEN** the preset is added with both resources

#### Scenario: Create preset with description
- **GIVEN** a library with resource `skill/commit`
- **WHEN** CreatePreset is called with name `commits` and description `Git commit best practices`
- **THEN** the preset is added with description `Git commit best practices`

### Requirement: Validate referenced resources exist

The system SHALL return an error if any referenced resource does not exist in the library.

#### Scenario: Create preset with nonexistent resource
- **GIVEN** a library without resource `skill/nonexistent`
- **WHEN** CreatePreset is called with resources `["skill/nonexistent"]`
- **THEN** an error "resource not found: skill/nonexistent" is returned

#### Scenario: Create preset with mixed valid and invalid resources
- **GIVEN** a library with resource `skill/commit` but not `skill/nonexistent`
- **WHEN** CreatePreset is called with resources `["skill/commit", "skill/nonexistent"]`
- **THEN** an error "resource not found: skill/nonexistent" is returned

### Requirement: Prevent duplicate preset names

The system SHALL return an error if a preset with the given name already exists unless --force is used.

#### Scenario: Create preset with duplicate name
- **GIVEN** a library with existing preset `git-workflow`
- **WHEN** CreatePreset is called with name `git-workflow`
- **THEN** an error "preset 'git-workflow' already exists (use --force to overwrite)" is returned

#### Scenario: Create preset with force flag
- **GIVEN** a library with existing preset `git-workflow`
- **WHEN** CreatePreset is called with name `git-workflow` and force=true
- **THEN** the existing preset is replaced with the new definition

### Requirement: Validate preset name is not empty

The system SHALL return an error if the preset name is empty or whitespace-only.

#### Scenario: Create preset with empty name
- **GIVEN** an empty preset name
- **WHEN** CreatePreset is called with name `""`
- **THEN** an error "preset name cannot be empty" is returned

#### Scenario: Create preset with whitespace name
- **GIVEN** a whitespace-only preset name
- **WHEN** CreatePreset is called with name `"   "`
- **THEN** an error "preset name cannot be whitespace only" is returned

### Requirement: Require at least one resource

The system SHALL return an error if no resources are specified.

#### Scenario: Create preset with no resources
- **GIVEN** no resources specified
- **WHEN** CreatePreset is called with empty resources list
- **THEN** an error "preset must have at least one resource" is returned

### Requirement: Persist library changes

The system SHALL persist the new preset to library.yaml.

#### Scenario: Persist preset to library.yaml
- **GIVEN** a valid preset creation request
- **WHEN** CreatePreset succeeds
- **THEN** library.yaml is updated with the new preset

### Requirement: CLI command interface

The system SHALL provide a CLI command `germinator library create preset <name>`.

#### Scenario: CLI help output
- **GIVEN** running `germinator library create preset --help`
- **THEN** help text shows usage, examples, and available flags

#### Scenario: CLI with required flags
- **GIVEN** running `germinator library create preset git-workflow --resources skill/commit,skill/pr`
- **THEN** the preset is created and success message is displayed

#### Scenario: CLI with missing --resources
- **GIVEN** running `germinator library create preset git-workflow`
- **THEN** an error "--resources is required" is displayed

#### Scenario: CLI with --description flag
- **GIVEN** running `germinator library create preset git-workflow --resources skill/commit --description "Git workflow tools"`
- **THEN** the preset is created with description "Git workflow tools"

#### Scenario: CLI with --force flag
- **GIVEN** running `germinator library create preset git-workflow --resources skill/commit --force`
- **THEN** any existing preset is overwritten

#### Scenario: CLI with --library flag
- **GIVEN** a library at `/custom/library`
- **WHEN** running `germinator library create preset git-workflow --resources skill/commit --library /custom/library`
- **THEN** the preset is created in the specified library

### Requirement: Display created preset details

The system SHALL display the created preset details on success.

#### Scenario: Display preset details on success
- **GIVEN** a successful preset creation
- **WHEN** the command completes
- **THEN** output shows preset name, description (or empty), and resources list

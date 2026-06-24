# Capability: Library System

## Purpose

The Library System capability provides resource management for the Germinator CLI. It handles library loading, resource resolution, preset expansion, and path discovery from filesystem-based library directories.

## Requirements

### Requirement: Load library from filesystem

The system SHALL load a library from a filesystem path with valid library.yaml.

#### Scenario: Load library from filesystem
- **GIVEN** a library directory with a valid `library.yaml`
- **WHEN** LoadLibrary is called with the directory path
- **THEN** the library is parsed and returned with all resources and presets indexed

### Requirement: Handle missing library.yaml

The system SHALL return a clear error when library.yaml is missing.

#### Scenario: Load library with missing library.yaml
- **GIVEN** a directory without a `library.yaml` file
- **WHEN** LoadLibrary is called
- **THEN** an error is returned indicating the library.yaml was not found

### Requirement: Resolve resource by reference

The system SHALL resolve resource references to file paths.

#### Scenario: Resolve resource by reference
- **GIVEN** a loaded library with a resource `skill/commit`
- **WHEN** ResolveResource is called with ref `skill/commit`
- **THEN** the absolute path to the resource file is returned

#### Scenario: Resolve nonexistent resource
- **GIVEN** a loaded library
- **WHEN** ResolveResource is called with ref `skill/nonexistent`
- **THEN** an error "resource not found: skill/nonexistent" is returned

#### Scenario: Resolve resource with invalid reference format
- **GIVEN** a loaded library
- **WHEN** ResolveResource is called with ref `invalidformat`
- **THEN** an error indicating invalid reference format is returned

### Requirement: Resolve presets to resource lists

The system SHALL expand preset names to lists of resource references.

#### Scenario: Resolve preset to resource list
- **GIVEN** a loaded library with preset `git-workflow` containing resources
- **WHEN** ResolvePreset is called with name `git-workflow`
- **THEN** a list of resource references (e.g., `["skill/commit", "skill/merge-request"]`) is returned

#### Scenario: Resolve nonexistent preset
- **GIVEN** a loaded library
- **WHEN** ResolvePreset is called with name `nonexistent`
- **THEN** an error "preset not found: nonexistent" is returned

### Requirement: Discover library path

The system SHALL discover the library path via flag, environment, or default.

#### Scenario: Discover library path from flag
- **GIVEN** flag `--library=/custom/path`
- **WHEN** FindLibrary is called
- **THEN** `/custom/path` is returned

#### Scenario: Discover library path from environment
- **GIVEN** no `--library` flag and env `GERMINATOR_LIBRARY=/env/path`
- **WHEN** FindLibrary is called
- **THEN** `/env/path` is returned

#### Scenario: Discover library path from default
- **GIVEN** no `--library` flag and no `GERMINATOR_LIBRARY` env
- **WHEN** FindLibrary is called
- **THEN** `~/.config/germinator/library/` is returned

### Requirement: List library contents

The system SHALL list resources and presets from a loaded library.

#### Scenario: List resources grouped by type
- **GIVEN** a loaded library with skills, agents, and commands
- **WHEN** ListResources is called
- **THEN** resources are returned grouped by type (skill, agent, command, memory)

#### Scenario: List presets
- **GIVEN** a loaded library with presets
- **WHEN** ListPresets is called
- **THEN** all presets with their descriptions are returned

### Requirement: Parse resource references

The system SHALL parse resource references in `type/name` format.

#### Scenario: Parse resource reference
- **GIVEN** a reference string `skill/commit`
- **WHEN** ParseRef is called
- **THEN** type `skill` and name `commit` are returned

### Requirement: Support all resource types

The system SHALL support skill, agent, command, and memory resource types.

#### Scenario: Support all resource types
- **GIVEN** a library with resources of types skill, agent, command, and memory
- **WHEN** resources are indexed and resolved
- **THEN** all four types are handled correctly

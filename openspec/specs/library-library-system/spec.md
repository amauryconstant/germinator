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
- **THEN** `~/.local/share/germinator/library/` is returned
- **AND** when `XDG_DATA_HOME` is set the default SHALL be `$XDG_DATA_HOME/germinator/library/`

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

### Requirement: Atomic library writes

The system SHALL write library metadata (`library.yaml`) atomically via the write-temp-then-rename pattern: write to `library.yaml.tmp`, fsync, then `os.Rename` to the final path. A crash mid-write leaves the previous `library.yaml` intact.

#### Scenario: Successful atomic write

- **GIVEN** a library loaded into memory with mutations
- **WHEN** `Save()` is called
- **THEN** the new content SHALL be written to `library.yaml.tmp` first
- **AND** on successful close, the temp file SHALL be renamed to `library.yaml`
- **AND** the final `library.yaml` SHALL contain the new content (observable via subsequent `LoadLibrary`)

#### Scenario: Crash mid-write preserves prior library

- **GIVEN** the temp-file write is interrupted (process killed before rename)
- **WHEN** `LoadLibrary` is called on the library directory
- **THEN** the prior `library.yaml` SHALL be returned intact
- **AND** no half-written data SHALL be visible

### Requirement: Library file permissions

When the system creates a new library directory or `library.yaml`, the permissions SHALL be:

- Library directory: `0750` (rwxr-x--- — owner full, group read+execute, others none)
- `library.yaml`: `0640` (rw-r----- — owner read+write, group read, others none)
- Resource subdirectories (skills/, agents/, commands/, memory/): `0750`

#### Scenario: Library directory permissions

- **WHEN** `Library.Init` creates the library root directory
- **THEN** the directory permissions SHALL be `0750`

#### Scenario: library.yaml permissions

- **WHEN** `Library.Init` writes the initial `library.yaml`
- **THEN** the file permissions SHALL be `0640`

### Requirement: Single-writer contract

The library SHALL be designed for single-writer usage (one user, one process at a time on a given library directory). Concurrent writes from multiple processes SHALL NOT be supported without external coordination (e.g., `flock`).

#### Scenario: Concurrent library modifications are last-writer-wins

- **GIVEN** two `germinator` processes modify the same library directory simultaneously
- **WHEN** both processes call `Save()` in overlapping windows
- **THEN** the final `library.yaml` SHALL reflect whichever `Save()` renamed last (no merging, no conflict detection)
- **AND** no corruption SHALL occur thanks to the atomic rename (each save is a complete snapshot)

> **Note:** For multi-writer scenarios (parallel CI, shared library on a network filesystem), use `github.com/gofrs/flock` to wrap library mutations in an advisory lock. The single-writer contract is the default; locking is opt-in for advanced use cases.

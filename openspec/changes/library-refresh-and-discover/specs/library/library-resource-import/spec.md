# Capability: Library Resource Import

## Purpose

The Library Resource Import capability enables importing existing canonical or platform documents into the library. It handles type detection, canonicalization, validation, file copying, and library.yaml updates. This spec extends the capability with orphan discovery mode.

## MODIFIED Requirements

### Requirement: Support orphan discovery mode

The system SHALL support discovering orphaned resource files via --discover flag.

#### Scenario: Discover flag shows orphans without modifying
- **GIVEN** a library with orphan files not in library.yaml
- **WHEN** AddResource is called with `--discover`
- **THEN** orphaned files are reported
- **AND** library.yaml is not modified

#### Scenario: Discover with force registers orphans
- **GIVEN** a library with orphan file `skills/new-skill.md`
- **WHEN** AddResource is called with `--discover --force`
- **THEN** the orphan is registered in library.yaml

#### Scenario: Discover requires explicit flag
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called without `--discover`
- **THEN** no discovery occurs
- **AND** orphans are not reported

#### Scenario: Discover dry-run previews registration
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called with `--discover --dry-run`
- **THEN** no files are modified
- **AND** expected additions are described

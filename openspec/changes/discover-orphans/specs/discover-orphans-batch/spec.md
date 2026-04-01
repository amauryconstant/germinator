# Capability: Discover Orphans Batch

## Purpose

The Discover Orphans Batch capability enables processing multiple orphans continuously with `--batch` mode, allowing partial success when some orphans fail validation or conflict.

## Requirements

### Requirement: Batch discover mode

The system SHALL support `--batch` flag for continuous orphan processing.

#### Scenario: Batch discover without force shows orphans only
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called with `--discover --batch`
- **THEN** orphans are reported
- **AND** no files are modified
- **AND** Summary shows TotalAdded=0

#### Scenario: Batch discover with force adds all orphans
- **GIVEN** a library with orphan files `skills/a.md`, `skills/b.md`
- **WHEN** AddResource is called with `--discover --batch --force`
- **THEN** all orphans are registered in library.yaml
- **AND** Summary shows TotalAdded=2

#### Scenario: Batch discover skips conflicts on force
- **GIVEN** a library with orphan `skills/existing.md` (name conflicts with existing)
- **AND** orphan `skills/new.md`
- **WHEN** AddResource is called with `--discover --batch --force`
- **THEN** `skills/new.md` is added
- **AND** `skills/existing.md` is skipped with conflict
- **AND** Summary shows TotalAdded=1, TotalSkipped=1

### Requirement: Batch discover with dry-run

The system SHALL preview batch operations without modifying files when `--dry-run` is specified.

#### Scenario: Batch discover dry-run shows all orphans
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called with `--discover --batch --dry-run`
- **THEN** no files are modified
- **AND** all orphans are listed with full details

### Requirement: Batch discover result summary

The system SHALL provide summary with counts for integration and reporting.

#### Scenario: Batch discover summary includes all counts
- **GIVEN** a library that scans 10 files finding 5 orphans
- **AND** 3 added successfully, 1 skipped (conflict), 1 failed (error)
- **WHEN** AddResource is called with `--discover --batch --force`
- **THEN** Summary contains TotalScanned=10, TotalOrphans=5, TotalAdded=3, TotalSkipped=1, TotalFailed=1

#### Scenario: Batch discover empty library
- **GIVEN** a library with no orphan files
- **WHEN** AddResource is called with `--discover --batch`
- **THEN** Summary shows TotalScanned=0, TotalOrphans=0, TotalAdded=0

### Requirement: Batch discover continues on individual failures

The system SHALL continue processing remaining orphans when one fails.

#### Scenario: Batch discover continues after registration error
- **GIVEN** a library with orphan `skills/valid.md` and orphan `skills/invalid.md` (causes error)
- **WHEN** AddResource is called with `--discover --batch --force`
- **THEN** `skills/valid.md` is added
- **AND** processing continues to attempt `skills/invalid.md`
- **AND** Summary shows both Added and Failed counts

### Requirement: Batch discover JSON output

The system SHALL support JSON output for batch operations.

#### Scenario: Batch discover with JSON flag
- **GIVEN** a library with orphan `skills/test.md`
- **WHEN** AddResource is called with `--discover --batch --force --json`
- **THEN** output is valid JSON with Orphans, Added, Conflicts, and Summary fields

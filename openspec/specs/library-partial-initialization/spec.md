# Capability: Partial Initialization

## Purpose

The Partial Initialization capability extends resource installation to continue processing all resources even when individual failures occur. It provides comprehensive result reporting so users can see all successes and failures in a single execution.

## Requirements

### Requirement: Continue on individual resource failure

The system SHALL process all resources in a request regardless of individual failures.

#### Scenario: Continue processing after resource resolution failure
- **GIVEN** resources `["skill/commit", "skill/nonexistent", "skill/merge-request"]`
- **WHEN** Initialize is called
- **THEN** `skill/commit` is processed successfully
- **AND** `skill/nonexistent` fails with resolution error
- **AND** `skill/merge-request` is processed successfully
- **AND** all three results are returned

#### Scenario: Continue processing after file write failure
- **GIVEN** resources where one has invalid content that fails rendering
- **WHEN** Initialize is called
- **THEN** resources before the failure are written successfully
- **AND** the failing resource has an error in its result
- **AND** resources after the failure are still attempted

#### Scenario: Continue processing after file exists error
- **GIVEN** resources `["skill/commit", "skill/existing", "skill/merge-request"]`
- **AND** `skill/existing` maps to a file that already exists without --force
- **WHEN** Initialize is called
- **THEN** `skill/commit` is processed successfully
- **AND** `skill/existing` fails with file exists error
- **AND** `skill/merge-request` is still processed

### Requirement: Return partial results

The system SHALL return results for all resources, including failures.

#### Scenario: Return success result for successful resource
- **GIVEN** a valid resource `skill/commit`
- **WHEN** Initialize is called
- **THEN** the result for `skill/commit` has Ref, InputPath, OutputPath set
- **AND** Error is nil

#### Scenario: Return error result for failed resource
- **GIVEN** an invalid resource `skill/nonexistent`
- **WHEN** Initialize is called
- **THEN** the result for `skill/nonexistent` has Ref set
- **AND** Error contains the resolution error
- **AND** InputPath and OutputPath may be empty

#### Scenario: Mixed results returned
- **GIVEN** three resources where the second fails
- **WHEN** Initialize is called
- **THEN** the returned slice contains three results
- **AND** first result has no error
- **AND** second result has an error
- **AND** third result has no error

### Requirement: Error return only when all resources fail

The system SHALL return nil error if at least one resource succeeded.

#### Scenario: Return nil error on partial success
- **GIVEN** resources where some succeed and some fail
- **WHEN** Initialize is called
- **THEN** the returned error is nil
- **AND** individual results contain the failures

#### Scenario: Return error when all resources fail
- **GIVEN** resources where all fail
- **WHEN** Initialize is called
- **THEN** a non-nil error is returned
- **AND** individual results contain the failures

### Requirement: Support dry-run with partial processing

The system SHALL preview all resources in dry-run mode regardless of errors.

#### Scenario: Dry-run continues through resolution errors
- **GIVEN** resources with one invalid ref
- **WHEN** Initialize is called with `--dry-run`
- **THEN** all resources are evaluated
- **AND** valid resources show their would-be output paths
- **AND** invalid resource shows its error

### Requirement: Support force flag with partial processing

The system SHALL apply force flag to all resources during partial processing.

#### Scenario: Force applies to all resources
- **GIVEN** resources where some would overwrite existing files
- **WHEN** Initialize is called with `--force`
- **THEN** all resources are processed
- **AND** existing files are overwritten
- **AND** no "file exists" errors are returned

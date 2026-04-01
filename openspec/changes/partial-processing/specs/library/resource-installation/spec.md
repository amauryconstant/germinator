# Capability: Resource Installation

## Purpose

The Resource Installation capability handles the installation of library resources to a target project. It orchestrates resource loading, transformation, and file writing with support for dry-run mode and force overwrite. This spec modifies the error handling behavior from fail-fast to partial processing.

## MODIFIED Requirements

### Requirement: Process all resources regardless of errors

**This replaces the previous "Fail-fast on errors" requirement.**

The system SHALL process all resources in the request, continuing on individual errors.

#### Scenario: Process all resources with mixed results
- **GIVEN** resources `["skill/commit", "skill/invalid", "skill/merge-request"]`
- **WHEN** InitializeResources is called and `skill/invalid` fails
- **THEN** `skill/commit` is processed successfully
- **AND** `skill/invalid` has an error in its result
- **AND** `skill/merge-request` is processed

#### Scenario: Return all results even on errors
- **GIVEN** a batch of resources with some failures
- **WHEN** InitializeResources is called
- **THEN** a result is returned for every resource
- **AND** successful results have no error
- **AND** failed results have the error set

#### Scenario: Continue through file write errors
- **GIVEN** resources where one fails to write due to permissions
- **WHEN** InitializeResources is called
- **THEN** the failing resource has an error in its result
- **AND** other resources are still processed

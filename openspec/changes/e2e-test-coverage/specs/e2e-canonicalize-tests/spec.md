## ADDED Requirements

### Requirement: Canonicalize Command E2E Tests

The canonicalize command SHALL be tested for all expected behaviors.

#### Scenario: Canonicalize valid document succeeds
- **WHEN** `germinator canonicalize <valid-platform-doc> <output> --platform opencode --type agent` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "Successfully canonicalized"
- **AND** output file SHALL be created

#### Scenario: Canonicalize with missing platform flag fails
- **WHEN** `germinator canonicalize <doc> <output> --type agent` is executed without `--platform`
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain "required" or "platform"

#### Scenario: Canonicalize with missing type flag fails
- **WHEN** `germinator canonicalize <doc> <output> --platform opencode` is executed without `--type`
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain "required" or "type"

#### Scenario: Canonicalize with invalid platform fails
- **WHEN** `germinator canonicalize <doc> <output> --platform invalid --type agent` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL indicate the platform is invalid

#### Scenario: Canonicalize with invalid type fails
- **WHEN** `germinator canonicalize <doc> <output> --platform opencode --type invalid` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL indicate the type is invalid

#### Scenario: Canonicalize nonexistent file fails
- **WHEN** `germinator canonicalize nonexistent.yaml <output> --platform opencode --type agent` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

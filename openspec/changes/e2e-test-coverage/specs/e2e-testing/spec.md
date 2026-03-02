## ADDED Requirements

### Requirement: Validate Command Platform Parity

The validate command SHALL be tested for both supported platforms.

#### Scenario: Validate valid document succeeds with claude-code platform
- **WHEN** `germinator validate <valid-doc> --platform claude-code` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "Document is valid"

#### Scenario: Validate nonexistent file fails with claude-code platform
- **WHEN** `germinator validate nonexistent.yaml --platform claude-code` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

#### Scenario: Validate invalid document fails with claude-code platform
- **WHEN** `germinator validate <invalid-doc> --platform claude-code` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain validation errors

---

### Requirement: Adapt Command Platform Parity

The adapt command SHALL be tested for both supported platforms.

#### Scenario: Adapt document succeeds with claude-code platform
- **WHEN** `germinator adapt <valid-doc> <output> --platform claude-code` is executed
- **THEN** exit code SHALL be 0
- **AND** stdout SHALL contain "transformed successfully"
- **AND** output file SHALL be created

#### Scenario: Adapt nonexistent file fails with claude-code platform
- **WHEN** `germinator adapt nonexistent.yaml <output> --platform claude-code` is executed
- **THEN** exit code SHALL be 1
- **AND** stderr SHALL contain an error message

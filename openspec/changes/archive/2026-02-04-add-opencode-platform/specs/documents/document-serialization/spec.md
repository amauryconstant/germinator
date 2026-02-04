## MODIFIED Requirements

### Requirement: Template Function Registration
The system SHALL register custom template functions for platform-specific transformations.

#### Scenario: Register template functions
- **GIVEN** RenderDocument function is called
- **WHEN** a template is parsed
- **THEN** it SHALL create a FuncMap with registered functions
- **AND** transformPermissionMode function SHALL be available

### Requirement: Permission Transformation Function
The system SHALL provide transformPermissionMode template function to convert Claude Code permissionMode enum to OpenCode permission objects.

#### Scenario: Transform permissionMode to OpenCode format
- **GIVEN** an Agent with permissionMode "acceptEdits"
- **WHEN** {{transformPermissionMode .PermissionMode}} is used in template
- **THEN** it SHALL output permission object with edit: {"*": "allow"} and bash: {"*": "ask"}

#### Scenario: Transform permissionMode "default"
- **WHEN** {{transformPermissionMode "default"}} is called
- **THEN** it SHALL output permission: {"edit": {"*": "ask"}, "bash": {"*": "ask"}}

### Requirement: Template Function Documentation
Template functions SHALL have Go documentation describing their purpose and usage.

#### Scenario: Function documentation exists
- **GIVEN** template functions are implemented in internal/core/template_funcs.go
- **WHEN** code is inspected
- **THEN** transformPermissionMode function SHALL have Go doc comment
- **AND** each doc SHALL describe purpose, parameters, and return value

### Requirement: Template Function Testing
Template functions SHALL have unit tests covering all scenarios.

#### Scenario: Unit tests pass for transformPermissionMode
- **GIVEN** transformPermissionMode function is tested
- **WHEN** unit tests are run
- **THEN** all 5 Claude Code modes SHALL be tested
- **AND** unknown mode SHALL be tested
- **AND** all tests SHALL pass

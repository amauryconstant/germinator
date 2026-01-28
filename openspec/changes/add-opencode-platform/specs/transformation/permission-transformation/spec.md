## ADDED Requirements

### Requirement: Claude Code to OpenCode Permission Transformation
The system SHALL provide transformPermissionMode function to convert Claude Code permissionMode enum to OpenCode permission object format.

#### Scenario: Transform "default" mode
- **GIVEN** permissionMode is "default"
- **WHEN** transformPermissionMode("default") is called
- **THEN** it SHALL return {"edit": {"*": "ask"}, "bash": {"*": "ask"}}

#### Scenario: Transform "acceptEdits" mode
- **GIVEN** permissionMode is "acceptEdits"
- **WHEN** transformPermissionMode("acceptEdits") is called
- **THEN** it SHALL return {"edit": {"*": "allow"}, "bash": {"*": "ask"}}

#### Scenario: Transform "dontAsk" mode
- **GIVEN** permissionMode is "dontAsk"
- **WHEN** transformPermissionMode("dontAsk") is called
- **THEN** it SHALL return {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
- **AND** This preserves semantic meaning: "don't ask user" means allow without prompting

#### Scenario: Transform "bypassPermissions" mode
- **GIVEN** permissionMode is "bypassPermissions"
- **WHEN** transformPermissionMode("bypassPermissions") is called
- **THEN** it SHALL return {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
- **AND** This preserves semantic meaning: "bypass permissions" means allow without restrictions}

#### Scenario: Transform "plan" mode
- **GIVEN** permissionMode is "plan"
- **WHEN** transformPermissionMode("plan") is called
- **THEN** it SHALL return {"edit": {"*": "deny"}, "bash": {"*": "deny"}}

#### Scenario: Handle unknown permissionMode
- **GIVEN** permissionMode is "unknown"
- **WHEN** transformPermissionMode("unknown") is called
- **THEN** it SHALL return nil

### Requirement: Permission Transformation Limitations
The permission transformation SHALL document its limitations due to fundamental differences between platforms.

#### Scenario: Basic approximation only
- **GIVEN** permission systems are fundamentally different (enum vs. nested objects)
- **WHEN** transforming permissions
- **THEN** it SHALL provide basic approximation for top-level edit and bash permissions only
- **AND** it SHALL NOT represent command-level granularity (e.g., "git push": "deny")
- **AND** dontAsk and bypassPermissions map to allow/allow (preserves semantic intent)

#### Scenario: No global wildcard support
- **GIVEN** OpenCode doesn't support global wildcards
- **WHEN** transforming to OpenCode format
- **THEN** it SHALL use tool-specific wildcards (edit: {"*": "ask"})

### Requirement: Permission Transformation Testing
Permission transformation functions SHALL have comprehensive test coverage.

#### Scenario: Unit tests for all permission modes
- **GIVEN** transformPermissionMode is tested
- **WHEN** tests are run
- **THEN** all 5 Claude Code modes SHALL be tested
- **AND** unknown mode SHALL be tested

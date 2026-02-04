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

#### Scenario: Eight tools are mapped for common permissions
- **GIVEN** OpenCode supports 15+ permissionable tools (read, write, edit, grep, glob, list, bash, task, skill, lsp, todoread, todowrite, webfetch, websearch, codesearch, external_directory, doom_loop)
- **WHEN** transformPermissionMode is called
- **THEN** Eight tools SHALL be set in permission object: edit, bash, read, grep, glob, list, webfetch, websearch
- **AND** Seven tools SHALL remain at undefined state: task, skill, lsp, todoread, todowrite, codesearch, external_directory, doom_loop
- **AND** Limitation SHALL be documented prominently in field mapping tables

#### Scenario: Command-level granularity not supported
- **GIVEN** Claude Code permissionMode enum cannot represent command-specific rules
- **WHEN** Transforming to OpenCode format
- **THEN** Transformation SHALL use tool-specific wildcards only (e.g., `{"edit": {"*": "allow"}}`)
- **AND** Command-level rules (e.g., `{"bash": {"git push": "deny"}}`) SHALL NOT be supported
- **AND** Limitation SHALL be documented

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

### Requirement: Expanded Permission Coverage
The system SHALL provide permission mappings for 8 common tools used in code development workflows.

#### Scenario: Core code analysis tools mapped
- **GIVEN** Claude Code permissionMode enum
- **WHEN** transformPermissionMode is called
- **THEN** read, grep, glob, list tools SHALL be mapped to appropriate permissions (ask/allow/deny) based on mode
- **AND** These tools are commonly used in read-only analysis workflows

#### Scenario: Web tools mapped
- **GIVEN** Claude Code permissionMode enum
- **WHEN** transformPermissionMode is called
- **THEN** webfetch, websearch tools SHALL be mapped to appropriate permissions (ask/allow/deny) based on mode
- **AND** These tools enable web-based research capabilities

### Requirement: Permission Transformation Testing
Permission transformation functions SHALL have comprehensive test coverage.

#### Scenario: Unit tests for all permission modes
- **GIVEN** transformPermissionMode is tested
- **WHEN** tests are run
- **THEN** all 5 Claude Code modes SHALL be tested
- **AND** unknown mode SHALL be tested

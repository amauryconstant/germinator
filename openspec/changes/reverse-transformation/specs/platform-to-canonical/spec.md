# platform-to-canonical Specification

## Purpose

Transform Claude Code or OpenCode documents into canonical YAML format via CLI command and reverse transformation pipeline, supporting all document types with user-specified platform and type parameters.

## Requirements

### Requirement: CLI Command Interface

The system SHALL provide a `canonicalize` CLI command that accepts input file path, output file path, platform flag, and document type flag.

#### Scenario: Invoke canonicalize command with valid arguments

- **GIVEN** a valid platform YAML file exists at input path
- **WHEN** user runs `germinator canonicalize <input> <output> --platform claude-code --type agent`
- **THEN** command SHALL parse platform file
- **AND** command SHALL validate canonical model
- **AND** command SHALL write canonical YAML to output file
- **AND** command SHALL return success message with output file path

#### Scenario: Invoke canonicalize command with missing platform flag

- **GIVEN** a valid platform YAML file exists at input path
- **WHEN** user runs `germinator canonicalize <input> <output> --type agent` (no --platform)
- **THEN** command SHALL display error message to stderr: "Error: --platform flag is required (valid: claude-code, opencode)"
- **AND** command SHALL exit with status code 1
- **AND** command SHALL NOT write any output file

#### Scenario: Invoke canonicalize command with missing type flag

- **GIVEN** a valid platform YAML file exists at input path
- **WHEN** user runs `germinator canonicalize <input> <output> --platform claude-code` (no --type)
- **THEN** command SHALL display error message to stderr: "Error: --type flag is required (valid: agent, command, skill, memory)"
- **AND** command SHALL exit with status code 1
- **AND** command SHALL NOT write any output file

#### Scenario: Invoke canonicalize command with invalid platform value

- **GIVEN** a valid platform YAML file exists at input path
- **WHEN** user runs `germinator canonicalize <input> <output> --platform invalid --type agent`
- **THEN** command SHALL display error message to stderr: "Error: invalid platform 'invalid' (valid: claude-code, opencode)"
- **AND** command SHALL exit with status code 1
- **AND** command SHALL NOT write any output file

#### Scenario: Invoke canonicalize command with invalid type value

- **GIVEN** a valid platform YAML file exists at input path
- **WHEN** user runs `germinator canonicalize <input> <output> --platform claude-code --type invalid`
- **THEN** command SHALL display error message to stderr: "Error: invalid document type 'invalid' (valid: agent, command, skill, memory)"
- **AND** command SHALL exit with status code 1
- **AND** command SHALL NOT write any output file

#### Scenario: Invoke canonicalize command with non-existent input file

- **GIVEN** no file exists at input path
- **WHEN** user runs `germinator canonicalize <input> <output> --platform claude-code --type agent`
- **THEN** command SHALL display error message to stderr: "Error: input file not found: <input>"
- **AND** command SHALL exit with status code 1
- **AND** command SHALL NOT write any output file

---

### Requirement: Platform File Parsing

The system SHALL parse platform YAML files into canonical models using adapter.ToCanonical() methods.

#### Scenario: Parse Claude Code agent file to canonical model

- **GIVEN** a Claude Code agent YAML with name, description, tools array, permissionMode enum, and skills array
- **WHEN** ParsePlatformDocument(path, "claude-code", "agent") is called
- **THEN** function SHALL read and parse YAML file
- **AND** function SHALL return CanonicalAgent struct with populated fields
- **AND** agent.Name SHALL match parsed name
- **AND** agent.Description SHALL match parsed description
- **AND** agent.Tools SHALL be lowercase array of tool names
- **AND** agent.PermissionPolicy SHALL be converted from permissionMode enum (default→restrictive, acceptEdits→balanced, etc.)
- **AND** agent.Targets["claude-code"]["skills"] SHALL contain parsed skills array
- **AND** function SHALL return nil error on success

#### Scenario: Parse OpenCode agent file to canonical model

- **GIVEN** an OpenCode agent YAML with description, tools map (boolean values), permission object (tool→action map), mode, temperature, maxSteps
- **WHEN** ParsePlatformDocument(path, "opencode", "agent") is called
- **THEN** function SHALL read and parse YAML file
- **AND** function SHALL return CanonicalAgent struct with populated fields
- **AND** agent.Description SHALL match parsed description
- **AND** agent.Tools SHALL contain tools with true value
- **AND** agent.DisallowedTools SHALL contain tools with false value
- **AND** agent.PermissionPolicy SHALL be inferred from permission object (balanced if mixed allow/ask, permissive if all allow, analysis if edit/bash denied)
- **AND** agent.Behavior.Mode SHALL match parsed mode
- **AND** agent.Behavior.Temperature SHALL match parsed temperature
- **AND** agent.Behavior.Steps SHALL match parsed maxSteps
- **AND** function SHALL return nil error on success

#### Scenario: Parse Claude Code command file to canonical model

- **GIVEN** a Claude Code command YAML with name, description, tools array, execution.context, execution.subtask, execution.agent, arguments.hint, model
- **WHEN** ParsePlatformDocument(path, "claude-code", "command") is called
- **THEN** function SHALL return CanonicalCommand struct with populated fields
- **AND** cmd.Name SHALL match parsed name
- **AND** cmd.Description SHALL match parsed description
- **AND** cmd.Tools SHALL be lowercase array of tool names
- **AND** cmd.Execution.Context SHALL match parsed execution.context
- **AND** cmd.Execution.Subtask SHALL match parsed execution.subtask
- **AND** cmd.Execution.Agent SHALL match parsed execution.agent
- **AND** cmd.Arguments.Hint SHALL match parsed arguments.hint
- **AND** cmd.Model SHALL match parsed model

#### Scenario: Parse OpenCode command file to canonical model

- **GIVEN** an OpenCode command YAML with name, description, allowed-tools array, context, subtask, agent, argument-hint, model
- **WHEN** ParsePlatformDocument(path, "opencode", "command") is called
- **THEN** function SHALL return CanonicalCommand struct with populated fields
- **AND** cmd.Name SHALL match parsed name
- **AND** cmd.Description SHALL match parsed description
- **AND** cmd.Tools SHALL match parsed allowed-tools array
- **AND** cmd.Execution.Context SHALL match parsed context
- **AND** cmd.Execution.Subtask SHALL match parsed subtask
- **AND** cmd.Execution.Agent SHALL match parsed agent
- **AND** cmd.Arguments.Hint SHALL match parsed argument-hint
- **AND** cmd.Model SHALL match parsed model

#### Scenario: Parse skill file to canonical model (both platforms)

- **GIVEN** a platform skill YAML (Claude Code or OpenCode) with name, description, tools, extensions (license, compatibility, metadata, hooks), execution (context, agent, userInvocable), model
- **WHEN** ParsePlatformDocument(path, platform, "skill") is called
- **THEN** function SHALL return CanonicalSkill struct with populated fields
- **AND** skill.Name SHALL match parsed name
- **AND** skill.Description SHALL match parsed description
- **AND** skill.Tools SHALL be lowercase array of tool names
- **AND** skill.Extensions.License SHALL match parsed license
- **AND** skill.Extensions.Compatibility SHALL match parsed compatibility array
- **AND** skill.Extensions.Metadata SHALL match parsed metadata map
- **AND** skill.Extensions.Hooks SHALL match parsed hooks map
- **AND** skill.Execution.Context SHALL match parsed execution context
- **AND** skill.Execution.Agent SHALL match parsed execution agent
- **AND** skill.Execution.UserInvocable SHALL match parsed userInvocable or user-invocable field
- **AND** skill.Model SHALL match parsed model

#### Scenario: Parse memory file to canonical model (both platforms)

- **GIVEN** a platform memory YAML with paths array and/or content string (in frontmatter or after --- delimiter)
- **WHEN** ParsePlatformDocument(path, platform, "memory") is called
- **THEN** function SHALL return CanonicalMemory struct with populated fields
- **AND** memory.Paths SHALL match parsed paths array
- **AND** memory.Content SHALL match parsed content (from frontmatter field or markdown body)
- **AND** function SHALL return nil error on success

#### Scenario: Parse platform file with invalid YAML syntax

- **GIVEN** a file with malformed YAML (e.g., unclosed brackets, invalid indentation)
- **WHEN** ParsePlatformDocument(path, platform, docType) is called
- **THEN** function SHALL return parsing error
- **AND** function SHALL NOT return any canonical model
- **AND** error message SHALL indicate YAML parsing failure

#### Scenario: Parse platform file with content after YAML frontmatter

- **GIVEN** a file with YAML frontmatter between --- delimiters and markdown content after second delimiter
- **WHEN** ParsePlatformDocument(path, platform, docType) is called
- **THEN** function SHALL parse YAML frontmatter to canonical model
- **AND** function SHALL extract markdown content (everything after second ---)
- **AND** canonical model's Content field SHALL contain extracted markdown
- **AND** function SHALL return nil error on success

---

### Requirement: Canonical Model Validation

The system SHALL validate canonical models after platform parsing using existing Validate() methods.

#### Scenario: Validate canonical agent with required fields

- **GIVEN** a CanonicalAgent struct with name, description, and valid permissionPolicy
- **WHEN** agent.Validate() is called
- **THEN** function SHALL return empty error array
- **AND** agent SHALL be considered valid

#### Scenario: Validate canonical agent with missing name

- **GIVEN** a CanonicalAgent struct with empty name field
- **WHEN** agent.Validate() is called
- **THEN** function SHALL return error array with "name is required" error
- **AND** agent SHALL be considered invalid

#### Scenario: Validate canonical agent with invalid name format

- **GIVEN** a CanonicalAgent struct with name containing uppercase letters or spaces (e.g., "My Agent")
- **WHEN** agent.Validate() is called
- **THEN** function SHALL return error array with name regex error
- **AND** error message SHALL indicate pattern ^[a-z0-9]+(-[a-z0-9]+)\*$ is required
- **AND** agent SHALL be considered invalid

#### Scenario: Validate canonical agent with invalid permission policy

- **GIVEN** a CanonicalAgent struct with permissionPolicy set to "invalid"
- **WHEN** agent.Validate() is called
- **THEN** function SHALL return error array with permission policy error
- **AND** error message SHALL list valid policies (restrictive, balanced, permissive, analysis, unrestricted)
- **AND** agent SHALL be considered invalid

#### Scenario: Validate canonical behavior with invalid temperature

- **GIVEN** a CanonicalAgent struct with behavior.temperature set to 1.5
- **WHEN** agent.Validate() is called
- **THEN** function SHALL return error array with temperature range error
- **AND** error message SHALL indicate temperature must be between 0.0 and 1.0
- **AND** agent SHALL be considered invalid

#### Scenario: Validate canonical command with required fields

- **GIVEN** a CanonicalCommand struct with name and description
- **WHEN** command.Validate() is called
- **THEN** function SHALL return empty error array
- **AND** command SHALL be considered valid

#### Scenario: Validate canonical memory with no paths or content

- **GIVEN** a CanonicalMemory struct with empty paths array and empty content string
- **WHEN** memory.Validate() is called
- **THEN** function SHALL return error array with "paths or content is required" error
- **AND** memory SHALL be considered invalid

#### Scenario: Validate canonical skill with invalid name length

- **GIVEN** a CanonicalSkill struct with name containing more than 64 characters
- **WHEN** skill.Validate() is called
- **THEN** function SHALL return error array with name length error
- **AND** error message SHALL indicate name must be 1-64 characters
- **AND** skill SHALL be considered invalid

---

### Requirement: Canonical YAML Serialization

The system SHALL serialize canonical models to YAML strings using canonical templates in config/templates/canonical/.

#### Scenario: Serialize canonical agent to YAML

- **GIVEN** a CanonicalAgent struct with name, description, tools, permissionPolicy, behavior, model
- **WHEN** MarshalCanonical(agent) is called
- **THEN** function SHALL return YAML string with proper structure
- **AND** YAML SHALL start with ---
- **AND** YAML SHALL contain name field with agent's name
- **AND** YAML SHALL contain description field with agent's description
- **AND** YAML SHALL contain tools array if agent has tools (each tool on separate line with dash)
- **AND** YAML SHALL contain permissionPolicy field if set
- **AND** YAML SHALL contain behavior object if any behavior fields are set
- **AND** YAML SHALL omit empty fields (no behavior object if all behavior fields are empty)
- **AND** YAML SHALL contain model field if set
- **AND** YAML SHALL end with ---
- **AND** YAML SHALL contain agent's Content field after second ---
- **AND** function SHALL return nil error on success

#### Scenario: Serialize canonical command to YAML

- **GIVEN** a CanonicalCommand struct with name, description, tools, execution, model
- **WHEN** MarshalCanonical(command) is called
- **THEN** function SHALL return YAML string with proper structure
- **AND** YAML SHALL contain name field with command's name
- **AND** YAML SHALL contain description field with command's description
- **AND** YAML SHALL contain tools array if command has tools
- **AND** YAML SHALL contain execution object if any execution fields are set
- **AND** YAML SHALL omit empty execution object
- **AND** YAML SHALL contain model field if set
- **AND** YAML SHALL contain command's Content field after second ---

#### Scenario: Serialize canonical skill to YAML

- **GIVEN** a CanonicalSkill struct with name, description, tools, extensions, execution, model
- **WHEN** MarshalCanonical(skill) is called
- **THEN** function SHALL return YAML string with proper structure
- **AND** YAML SHALL contain name field with skill's name
- **AND** YAML SHALL contain description field with skill's description
- **AND** YAML SHALL contain tools array if skill has tools
- **AND** YAML SHALL contain extensions object if any extension fields are set
- **AND** YAML SHALL omit empty extensions object
- **AND** YAML SHALL contain execution object if any execution fields are set
- **AND** YAML SHALL omit empty execution object
- **AND** YAML SHALL contain model field if set
- **AND** YAML SHALL contain skill's Content field after second ---

#### Scenario: Serialize canonical memory to YAML

- **GIVEN** a CanonicalMemory struct with paths and content
- **WHEN** MarshalCanonical(memory) is called
- **THEN** function SHALL return YAML string with proper structure
- **AND** YAML SHALL contain paths array if memory has paths
- **AND** YAML SHALL omit paths array if empty
- **AND** YAML SHALL contain content field if memory.Content is non-empty
- **AND** YAML SHALL use pipe (|) syntax for content field to preserve multiline formatting
- **AND** YAML SHALL omit content field if empty and paths exist

#### Scenario: Serialize canonical model targets section

- **GIVEN** a CanonicalAgent struct with targets["claude-code"]["skills"] set to ["skill1", "skill2"]
- **WHEN** MarshalCanonical(agent) is called
- **THEN** YAML SHALL contain targets section
- **AND** YAML SHALL contain claude-code subsection under targets
- **AND** YAML SHALL contain skills array under targets.claude-code
- **AND** YAML SHALL preserve array structure ["skill1", "skill2"]

#### Scenario: Serialize canonical model with all empty optional fields

- **GIVEN** a CanonicalAgent struct with only name and description set (tools empty, permissionPolicy empty, behavior all zeros, model empty)
- **WHEN** MarshalCanonical(agent) is called
- **THEN** YAML SHALL contain only name and description fields
- **AND** YAML SHALL NOT contain tools array
- **AND** YAML SHALL NOT contain permissionPolicy field
- **AND** YAML SHALL NOT contain behavior object
- **AND** YAML SHALL NOT contain model field
- **AND** YAML SHALL NOT contain targets section

---

### Requirement: File Writing

The system SHALL write canonical YAML output to specified file path with appropriate permissions.

#### Scenario: Write canonical output to new file

- **GIVEN** canonical YAML string is ready and output path does not exist
- **WHEN** system writes output to file
- **THEN** system SHALL create new file at output path
- **AND** file SHALL contain complete canonical YAML string
- **AND** file permissions SHALL be 0644 (rw-r--r--)
- **AND** write operation SHALL succeed without error

#### Scenario: Write canonical output to existing file

- **GIVEN** canonical YAML string is ready and output path already exists
- **WHEN** system writes output to file
- **THEN** system SHALL overwrite existing file
- **AND** file SHALL contain complete canonical YAML string (replacing previous content)
- **AND** file permissions SHALL be 0644 (rw-r--r--)
- **AND** write operation SHALL succeed without error

#### Scenario: Write canonical output to read-only directory

- **GIVEN** canonical YAML string is ready and output path is in read-only directory
- **WHEN** system attempts to write to file
- **THEN** system SHALL receive file write error
- **AND** system SHALL NOT create output file
- **AND** system SHALL return error to caller
- **AND** error message SHALL indicate permission denied

#### Scenario: Write canonical output with YAML extension

- **GIVEN** output path ends with .yaml extension
- **WHEN** system writes output to file
- **THEN** file SHALL be created with .yaml extension
- **AND** file content SHALL be valid YAML
- **AND** YAML tools/editors shall recognize file format

---

### Requirement: Reverse Transformation Pipeline Orchestration

The system SHALL provide CanonicalizeDocument() function that orchestrates Parse → Validate → Serialize → Write pipeline in order.

#### Scenario: Execute successful reverse transformation

- **GIVEN** a valid platform YAML file and writable output path
- **WHEN** CanonicalizeDocument(input, output, "claude-code", "agent") is called
- **THEN** system SHALL parse platform file to canonical model
- **AND** system SHALL validate canonical model (no errors)
- **AND** system SHALL serialize canonical model to YAML string
- **AND** system SHALL write YAML string to output file
- **AND** system SHALL return nil error on success

#### Scenario: Fail fast on parsing error

- **GIVEN** a platform YAML file with invalid YAML syntax
- **WHEN** CanonicalizeDocument(input, output, platform, docType) is called
- **THEN** system SHALL attempt to parse platform file
- **AND** parsing SHALL fail with error
- **AND** system SHALL NOT attempt validation
- **AND** system SHALL NOT attempt serialization
- **AND** system SHALL NOT write output file
- **AND** system SHALL return parsing error

#### Scenario: Fail fast on validation error

- **GIVEN** a platform YAML file that parses to invalid canonical model (e.g., missing name)
- **WHEN** CanonicalizeDocument(input, output, platform, docType) is called
- **THEN** system SHALL parse platform file successfully
- **AND** validation SHALL fail with errors
- **AND** system SHALL NOT attempt serialization
- **AND** system SHALL NOT write output file
- **AND** system SHALL return validation errors

#### Scenario: Fail on serialization error

- **GIVEN** a canonical model that is valid but cannot be serialized (e.g., template parse error)
- **WHEN** CanonicalizeDocument(input, output, platform, docType) is called
- **THEN** system SHALL parse and validate successfully
- **AND** serialization SHALL fail with error
- **AND** system SHALL NOT write output file
- **AND** system SHALL return serialization error

#### Scenario: Fail on file write error

- **GIVEN** a canonical model that parses, validates, and serializes successfully
- **WHEN** CanonicalizeDocument(input, output, platform, docType) is called with unwritable output path
- **THEN** system SHALL parse, validate, and serialize successfully
- **AND** file write SHALL fail with error
- **AND** system SHALL return file write error

#### Scenario: Execute reverse transformation for all document types

- **GIVEN** valid platform files for agent, command, skill, and memory
- **WHEN** CanonicalizeDocument() is called for each document type with appropriate --type flag
- **THEN** all four transformations SHALL succeed
- **AND** all four output files SHALL be valid canonical YAML
- **AND** all four output files SHALL pass model validation

#### Scenario: Execute reverse transformation for both platforms

- **GIVEN** valid Claude Code and OpenCode agent files
- **WHEN** CanonicalizeDocument() is called with --platform claude-code for first file and --platform opencode for second file
- **THEN** both transformations SHALL succeed
- **AND** both output files SHALL be valid canonical YAML
- **AND** both canonical agents SHALL have equivalent semantic meaning (name, description, tools, permissionPolicy, behavior)

---

### Requirement: Error Messages and User Guidance

The system SHALL provide clear, actionable error messages for all failure scenarios.

#### Scenario: Display error message for missing required flag

- **GIVEN** user invokes canonicalize command without --platform or --type flag
- **WHEN** command validates flags
- **THEN** system SHALL display error message to stderr: "Error: --platform flag is required (valid: claude-code, opencode)" or "Error: --type flag is required (valid: agent, command, skill, memory)"
- **AND** command SHALL exit with status code 1

#### Scenario: Display error message for invalid flag value

- **GIVEN** user provides invalid platform value (e.g., "github") or invalid type value (e.g., "service")
- **WHEN** command validates flags
- **THEN** system SHALL display error message to stderr: "Error: invalid platform 'github' (valid: claude-code, opencode)" or "Error: invalid document type 'service' (valid: agent, command, skill, memory)"
- **AND** command SHALL exit with status code 1

#### Scenario: Display error message for parsing failure

- **GIVEN** platform YAML file has invalid syntax or structure
- **WHEN** CanonicalizeDocument() attempts to parse file
- **THEN** system SHALL return error to CLI
- **AND** error message SHALL indicate parsing failed
- **AND** error message MAY include YAML parsing details (line number, specific error)
- **AND** command SHALL display error message to stderr
- **AND** command SHALL exit with status code 1

#### Scenario: Display error message for validation failure

- **GIVEN** canonical model fails validation (e.g., missing name, invalid permission policy)
- **WHEN** CanonicalizeDocument() validates canonical model
- **THEN** system SHALL return validation errors to CLI
- **AND** error message SHALL display all validation errors (not just first)
- **AND** each validation error SHALL be on separate line
- **AND** errors SHALL clearly indicate which field is invalid and why
- **AND** command SHALL display errors to stderr
- **AND** command SHALL exit with status code 1

#### Scenario: Display success message on successful transformation

- **GIVEN** canonicalize command completes successfully
- **WHEN** transformation finishes
- **THEN** command SHALL display success message to stdout: "Successfully canonicalized document to: <output>"
- **AND** command SHALL exit with status code 0

#### Scenario: Display error message for file not found

- **GIVEN** input path does not exist
- **WHEN** command attempts to read file
- **THEN** system SHALL display error message to stderr
- **AND** error message SHALL indicate file not found
- **AND** error message SHALL include the input path that was not found
- **AND** command SHALL exit with status code 1

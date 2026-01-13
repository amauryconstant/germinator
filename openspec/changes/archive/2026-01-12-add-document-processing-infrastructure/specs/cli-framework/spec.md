## ADDED Requirements

### Requirement: Validate Command

The CLI SHALL provide a validate command that validates document files.

#### Scenario: Validate single document
**Given** a document file path is provided
**When** `germinator validate <file>` is run
**Then** the command SHALL parse the document
**And** it SHALL validate the document
**And** it SHALL display validation errors if any exist
**And** it SHALL exit with code 0 if valid, non-zero if invalid

#### Scenario: Validate displays clear error messages
**Given** a document with validation errors
**When** `germinator validate <file>` is run
**Then** each error SHALL be displayed on a separate line
**And** errors SHALL be clearly formatted for human reading
**And** the output SHALL indicate the total number of errors

#### Scenario: Validate command help
**Given** the validate command exists
**When** `germinator validate --help` is run
**Then** it SHALL display usage information
**And** it SHALL describe the command's purpose

#### Scenario: Validate handles missing file
**Given** a file that doesn't exist
**When** `germinator validate <file>` is run
**Then** it SHALL display an error message indicating the file doesn't exist
**And** it SHALL exit with a non-zero code

---

### Requirement: Adapt Command

The CLI SHALL provide an adapt command that transforms documents to target platforms.

#### Scenario: Adapt document to platform
**Given** an input document file and output file path
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL parse the input document
**And** it SHALL validate the document
**And** it SHALL serialize the document
**And** it SHALL write to the output file
**And** it SHALL display success message

#### Scenario: Adapt fails on validation
**Given** an invalid input document
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL parse the document
**And** it SHALL detect validation errors
**And** it SHALL NOT create the output file
**And** it SHALL display validation errors
**And** it SHALL exit with a non-zero code

#### Scenario: Adapt with Claude Code platform
**Given** a valid input document and output file
**When** `germinator adapt <input> <output> --platform claude-code` is run
**Then** it SHALL transform the document (pass-through validation and serialization)
**And** the output file SHALL contain the validated document

#### Scenario: Adapt command help
**Given** the adapt command exists
**When** `germinator adapt --help` is run
**Then** it SHALL display usage information
**And** it SHALL describe the command's purpose
**And** it SHALL list required arguments (input, output)
**And** it SHALL list required flags (--platform)

#### Scenario: Adapt handles read errors
**Given** an input file that cannot be read (permission denied)
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL display an error message indicating the read failure
**And** it SHALL NOT create the output file
**And** it SHALL exit with a non-zero code

#### Scenario: Adapt handles write errors
**Given** a valid input document but output directory is read-only
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL display an error message indicating the write failure
**And** it SHALL exit with a non-zero code

---

### Requirement: Command Registration

The CLI SHALL register all new subcommands with the root command.

#### Scenario: Register validate command
**Given** the CLI is initialized
**When** the root command is inspected
**Then** it SHALL have a "validate" subcommand
**And** the subcommand SHALL be accessible via `germinator validate`

#### Scenario: Register adapt command
**Given** the CLI is initialized
**When** the root command is inspected
**Then** it SHALL have an "adapt" subcommand
**And** the subcommand SHALL be accessible via `germinator adapt`

#### Scenario: Commands appear in help
**Given** the CLI is initialized
**When** `germinator --help` is run
**Then** the help output SHALL list all available commands
**And** it SHALL include "validate" in the commands list
**And** it SHALL include "adapt" in the commands list

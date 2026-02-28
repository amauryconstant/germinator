# cli-framework Specification (Delta)

## Purpose

Define Cobra CLI framework with validate, adapt, and canonicalize commands for document transformation.

## MODIFIED Requirements

### Requirement: Root Command Structure

The project SHALL have a root command named "germinator" with basic functionality and persistent flags.

#### Scenario: Root command exists

- **GIVEN** the project has been initialized
- **WHEN** a developer builds and runs the binary
- **THEN** a root command named "germinator" SHALL be available
- **AND** it SHALL display help when run with --help flag

#### Scenario: Root command has description

- **GIVEN** the root command exists
- **WHEN** a developer runs `./germinator --help`
- **THEN** the output SHALL include a description explaining the tool's purpose

#### Scenario: Root command has persistent verbose flag

- **GIVEN** the root command exists
- **WHEN** a developer runs `./germinator --help`
- **THEN** the output SHALL include a `-v` flag description
- **AND** the flag SHALL be marked as persistent

---

### Requirement: CommandConfig Pattern

The CLI SHALL use a CommandConfig struct for dependency injection.

#### Scenario: CommandConfig initialization

- **GIVEN** the root command is executed
- **WHEN** PreRun or Run is called
- **THEN** a CommandConfig SHALL be created with ErrorFormatter and Verbosity
- **AND** CommandConfig SHALL be passed to command functions

#### Scenario: CommandConfig contains ErrorFormatter

- **GIVEN** a CommandConfig instance
- **WHEN** ErrorFormatter field is accessed
- **THEN** it SHALL be a valid ErrorFormatter instance
- **AND** it SHALL be used for all error formatting in the command

#### Scenario: CommandConfig contains Verbosity

- **GIVEN** a CommandConfig instance
- **WHEN** Verbosity field is accessed
- **THEN** it SHALL reflect the current verbosity level from flags
- **AND** it SHALL be used for verbose output decisions

---

### Requirement: Validate Command

The CLI SHALL provide a validate command that validates document files.

#### Scenario: Validate single document

- **GIVEN** a document file path is provided
- **WHEN** `germinator validate <file>` is run
- **THEN** the command SHALL parse the document
- **AND** it SHALL validate the document
- **AND** it SHALL display validation errors if any exist
- **AND** it SHALL exit with code 0 if valid
- **AND** it SHALL exit with code 2 for validation errors
- **AND** it SHALL exit with code 3 for parse errors

#### Scenario: Validate displays clear error messages

- **GIVEN** a document with validation errors
- **WHEN** `germinator validate <file>` is run
- **THEN** each error SHALL be displayed on a separate line
- **AND** errors SHALL be clearly formatted for human reading
- **AND** contextual hints SHALL be provided when available

#### Scenario: Validate command help

- **GIVEN** the validate command exists
- **WHEN** `germinator validate --help` is run
- **THEN** it SHALL display usage information
- **AND** it SHALL describe the command's purpose

#### Scenario: Validate handles missing file

- **GIVEN** a file that doesn't exist
- **WHEN** `germinator validate <file>` is run
- **THEN** it SHALL display a file error message
- **AND** it SHALL exit with code 1

#### Scenario: Validate handles invalid platform

- **GIVEN** an invalid platform flag
- **WHEN** `germinator validate <file> --platform invalid` is run
- **THEN** it SHALL display a config error with valid platforms
- **AND** it SHALL exit with code 2

#### Scenario: Validate uses CommandConfig

- **GIVEN** the validate command is run
- **WHEN** execution begins
- **THEN** it SHALL receive a CommandConfig with ErrorFormatter and Verbosity
- **AND** it SHALL use CommandConfig.ErrorFormatter for error output
- **AND** it SHALL use CommandConfig.Verbosity for verbose output

---

### Requirement: Adapt Command

The CLI SHALL provide an adapt command that transforms documents to target platforms.

#### Scenario: Adapt document to platform

- **GIVEN** an input document file and output file path
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** it SHALL parse the input document
- **AND** it SHALL validate the document
- **AND** it SHALL serialize the document
- **AND** it SHALL write to the output file
- **AND** it SHALL display success message
- **AND** it SHALL exit with code 0

#### Scenario: Adapt fails on validation

- **GIVEN** an invalid input document
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** it SHALL parse the document
- **AND** it SHALL detect validation errors
- **AND** it SHALL NOT create the output file
- **AND** it SHALL display validation errors with hints
- **AND** it SHALL exit with code 2

#### Scenario: Adapt fails on parse error

- **GIVEN** a document with invalid YAML
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** it SHALL NOT create the output file
- **AND** it SHALL display parse error with file path
- **AND** it SHALL exit with code 3

#### Scenario: Adapt with Claude Code platform

- **GIVEN** a valid input document and output file
- **WHEN** `germinator adapt <input> <output> --platform claude-code` is run
- **THEN** it SHALL transform the document (pass-through validation and serialization)
- **AND** the output file SHALL contain the validated document

#### Scenario: Adapt command help

- **GIVEN** the adapt command exists
- **WHEN** `germinator adapt --help` is run
- **THEN** it SHALL display usage information
- **AND** it SHALL describe the command's purpose
- **AND** it SHALL list required arguments (input, output)
- **AND** it SHALL list required flags (--platform)

#### Scenario: Adapt handles read errors

- **GIVEN** an input file that cannot be read (permission denied)
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** it SHALL display a file error with path and operation
- **AND** it SHALL NOT create the output file
- **AND** it SHALL exit with code 1

#### Scenario: Adapt handles write errors

- **GIVEN** a valid input document but output directory is read-only
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** it SHALL display a file error with path and operation
- **AND** it SHALL exit with code 1

#### Scenario: Adapt handles invalid platform

- **GIVEN** an invalid platform flag
- **WHEN** `germinator adapt <input> <output> --platform invalid` is run
- **THEN** it SHALL display a config error with valid platforms
- **AND** it SHALL exit with code 2

#### Scenario: Adapt uses CommandConfig

- **GIVEN** the adapt command is run
- **WHEN** execution begins
- **THEN** it SHALL receive a CommandConfig with ErrorFormatter and Verbosity
- **AND** it SHALL use CommandConfig.ErrorFormatter for error output
- **AND** it SHALL use CommandConfig.Verbosity for verbose output

---

### Requirement: Canonicalize Command

The CLI SHALL provide a canonicalize command that converts platform documents to canonical format.

#### Scenario: Canonicalize uses CommandConfig

- **GIVEN** the canonicalize command is run
- **WHEN** execution begins
- **THEN** it SHALL receive a CommandConfig with ErrorFormatter and Verbosity
- **AND** it SHALL use CommandConfig.ErrorFormatter for error output
- **AND** it SHALL use CommandConfig.Verbosity for verbose output

#### Scenario: Canonicalize handles errors with typed errors

- **GIVEN** a platform document with errors
- **WHEN** `germinator canonicalize <input> <output>` is run
- **THEN** it SHALL display typed error messages
- **AND** it SHALL exit with appropriate exit code

---

### Requirement: Central Error Handler

The CLI SHALL provide a central error handler for consistent error processing.

#### Scenario: HandleError processes all errors

- **GIVEN** an error occurs in any command
- **WHEN** HandleError is called with CommandConfig and the error
- **THEN** it SHALL format the error using CommandConfig.ErrorFormatter
- **AND** it SHALL write formatted output to stderr
- **AND** it SHALL exit with the appropriate exit code

#### Scenario: HandleError with nil error

- **GIVEN** HandleError is called with nil
- **WHEN** the function executes
- **THEN** it SHALL exit with code 0
- **AND** it SHALL NOT write to stderr

# cli-framework Specification

## Purpose

Define Cobra CLI framework with validate, adapt, canonicalize, init, library, config, and version commands for document transformation. All commands obtain dependencies through `*cmdutil.Factory` (see `cli-factory` and `command-options-pattern`).

## Requirements

### Requirement: Cobra CLI Framework

The project SHALL use the Cobra framework for CLI command structure.

#### Scenario: Cobra dependency is installed

**Given** the Go module is initialized
**When** a developer runs `go get github.com/spf13/cobra@latest`
**Then** Cobra SHALL be added to go.mod
**And** go.sum SHALL be updated
**And** `go build ./...` SHALL succeed

#### Scenario: Cobra is imported in cmd package

**Given** the cmd/root.go file exists
**When** the file is inspected
**Then** it SHALL import github.com/spf13/cobra
**And** it SHALL use cobra.Command for command definition

---

### Requirement: Root Command Structure

The project SHALL have a root command named "germinator" with basic functionality and persistent flags.

#### Scenario: Root command exists

**Given** the project has been initialized
**When** a developer builds and runs the binary
**Then** a root command named "germinator" SHALL be available
**And** it SHALL display help when run with --help flag

#### Scenario: Root command has description

**Given** the root command exists
**When** a developer runs `./germinator --help`
**Then** the output SHALL include a description explaining the tool's purpose

#### Scenario: Root command has persistent verbose flag

**Given** the root command exists
**When** a developer runs `./germinator --help`
**Then** the output SHALL include a `-v` flag description
**And** the flag SHALL be marked as persistent

---

### Requirement: Commands take Factory, not CommandConfig

Each command's constructor SHALL take `*cmdutil.Factory` as its first parameter (after the optional `runF` for test injection). No command SHALL take `*CommandConfig`. The `CommandConfig` type and `cmd/command_config.go` SHALL be removed.

#### Scenario: NewCmdXxx signature

- **WHEN** a command's constructor signature is inspected
- **THEN** it SHALL match `NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command`
- **AND** it SHALL NOT have any parameter of type `*CommandConfig`

### Requirement: No global CommandConfig

The `cmd.SetGlobalCommandConfig(*CommandConfig)` function and any package-level `CommandConfig` variable SHALL be **removed**. All command state SHALL flow through `opts`.

#### Scenario: No global command config

- **WHEN** the codebase is inspected
- **THEN** there SHALL be no `var globalConfig *CommandConfig` or similar
- **AND** there SHALL be no `SetGlobalCommandConfig` function
- **AND** no command SHALL call `cmd.GetCommandConfig()` or similar getter

---

### Requirement: RunE Command Pattern

All CLI commands that can fail SHALL use the `RunE` pattern instead of `Run`.

#### Scenario: Commands use RunE instead of Run

- **GIVEN** a command that can return errors (validate, adapt, canonicalize, init, library)
- **WHEN** the command definition is inspected
- **THEN** it SHALL use `RunE` field instead of `Run`
- **AND** the function signature SHALL be `func(cmd *cobra.Command, args []string) error`

#### Scenario: Commands return errors instead of calling os.Exit

- **GIVEN** a command encounters an error
- **WHEN** the error occurs in the RunE function
- **THEN** the error SHALL be returned
- **AND** os.Exit SHALL NOT be called within the command

#### Scenario: Version command can use Run

- **GIVEN** the version command cannot fail
- **WHEN** the command definition is inspected
- **THEN** it MAY use `Run` instead of `RunE`

---

### Requirement: Centralized Error Handling

Error handling SHALL be centralized in `main.go`. Commands return typed errors; `main.go` formats them and maps them to exit codes.

#### Scenario: main.go handles all errors

- **GIVEN** main.go executes the root command
- **WHEN** rootCmd.Execute() returns an error
- **THEN** `output.FormatError(io, err)` SHALL be called to render the error
- **AND** `cmdutil.ExitCodeFor(err)` SHALL be called to determine the exit code
- **AND** the process SHALL exit with that exit code

#### Scenario: Commands delegate error formatting

- **GIVEN** a command's `RunE` returns a typed error (e.g. `*core.ValidationError`)
- **WHEN** main.go processes the error
- **THEN** the typed-error dispatch in `output.FormatError` SHALL choose the correct rendering
- **AND** the command SHALL NOT call `output.FormatError` itself

#### Scenario: Cobra argument errors map to ExitCodeUsage

- **GIVEN** rootCmd.Execute() returns a Cobra usage error (unknown flag, missing arg, etc.)
- **WHEN** main.go calls `cmdutil.ExitCodeFor(err)`
- **THEN** the result SHALL be `ExitCodeUsage` (2)

---

### Requirement: CLI Entry Point

The project SHALL have a main entry point in cmd/root.go that initializes the CLI.

#### Scenario: Main function exists

**Given** the project has been initialized
**When** the cmd/root.go file is inspected
**Then** a main() function SHALL exist
**And** it SHALL execute the root command

#### Scenario: Root command is executable

**Given** the main function exists
**When** a developer runs `go build ./cmd`
**Then** a binary SHALL be produced
**And** the binary SHALL be executable
**And** running the binary SHALL execute the root command

---

### Requirement: Validate Command

The CLI SHALL provide a validate command that validates document files.

#### Scenario: Validate single document

**Given** a document file path is provided
**When** `germinator validate <file>` is run
**Then** the command SHALL parse the document
**And** it SHALL validate the document
**And** it SHALL display validation errors if any exist
**And** it SHALL return nil if valid
**And** it SHALL return `*core.ValidationError` for validation errors
**And** it SHALL return `*core.ConfigError` for parse errors

#### Scenario: Validate displays clear error messages

**Given** a document with validation errors
**When** `germinator validate <file>` is run
**Then** each error SHALL be displayed on a separate line
**And** errors SHALL be clearly formatted for human reading
**And** contextual hints SHALL be provided when available

#### Scenario: Validate command help

**Given** the validate command exists
**When** `germinator validate --help` is run
**Then** it SHALL display usage information
**And** it SHALL describe the command's purpose

#### Scenario: Validate handles missing file

**Given** a file that doesn't exist
**When** `germinator validate <file>` is run
**Then** the command SHALL display a file error message
**And** it SHALL return `*core.FileError`

#### Scenario: Validate handles invalid platform

**Given** an invalid platform flag
**When** `germinator validate <file> --platform invalid` is run
**Then** the command SHALL display a config error with valid platforms
**And** it SHALL return `*core.ConfigError`

#### Scenario: Validate takes a Factory

- **GIVEN** the validate command is run
- **WHEN** execution begins
- **THEN** its `NewCmdValidate(f *cmdutil.Factory, runF)` SHALL be invoked
- **AND** the constructed `validateOptions` SHALL hold `IO *iostreams.IOStreams`, lazy `Validator func() (Validator, error)`, and parsed flags
- **AND** the command SHALL NOT reference a `*CommandConfig`

---

### Requirement: Adapt Command

The CLI SHALL provide an adapt command that transforms documents to target platforms.

#### Scenario: Adapt document to platform

**Given** an input document file and output file path
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** the command SHALL parse the input document
**And** it SHALL validate the document
**And** it SHALL serialize the document
**And** it SHALL write to the output file
**And** it SHALL display a success message
**And** it SHALL return nil on success

#### Scenario: Adapt fails on validation

**Given** an invalid input document
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** the command SHALL parse the document
**And** it SHALL detect validation errors
**And** it SHALL NOT create the output file
**And** it SHALL display validation errors with hints
**And** it SHALL return `*core.ValidationError`

#### Scenario: Adapt fails on parse error

**Given** a document with invalid YAML
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL NOT create the output file
**And** it SHALL display a parse error with file path
**And** it SHALL return `*core.ConfigError`

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
**Then** it SHALL display a file error with path and operation
**And** it SHALL NOT create the output file
**And** it SHALL return `*core.FileError`

#### Scenario: Adapt handles write errors

**Given** a valid input document but output directory is read-only
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL display a file error with path and operation
**And** it SHALL return `*core.FileError`

#### Scenario: Adapt handles invalid platform

**Given** an invalid platform flag
**When** `germinator adapt <input> <output> --platform invalid` is run
**Then** it SHALL display a config error with valid platforms
**And** it SHALL return `*core.ConfigError`

#### Scenario: Adapt takes a Factory

- **GIVEN** the adapt command is run
- **WHEN** execution begins
- **THEN** its `NewCmdAdapt(f *cmdutil.Factory, runF)` SHALL be invoked
- **AND** the constructed `adaptOptions` SHALL hold `IO`, lazy `Transformer` / `Validator`, and parsed flags
- **AND** the command SHALL NOT reference a `*CommandConfig`

---

### Requirement: Canonicalize Command

The CLI SHALL provide a canonicalize command that converts platform documents to canonical format.

#### Scenario: Canonicalize handles errors with typed errors

- **GIVEN** a platform document with errors
- **WHEN** `germinator canonicalize <input> <output>` is run
- **THEN** it SHALL display typed error messages
- **AND** it SHALL return the appropriate typed error

#### Scenario: Canonicalize takes a Factory

- **GIVEN** the canonicalize command is run
- **WHEN** execution begins
- **THEN** its `NewCmdCanonicalize(f *cmdutil.Factory, runF)` SHALL be invoked
- **AND** the constructed `canonicalizeOptions` SHALL hold `IO`, lazy `Canonicalizer`, and parsed flags
- **AND** the command SHALL NOT reference a `*CommandConfig`

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

#### Scenario: Register canonicalize command

**Given** the CLI is initialized
**When** the root command is inspected
**Then** it SHALL have a "canonicalize" subcommand
**And** the subcommand SHALL be accessible via `germinator canonicalize`

#### Scenario: Commands appear in help

**Given** the CLI is initialized
**When** `germinator --help` is run
**Then** the help output SHALL list all available commands
**And** it SHALL include "validate" in the commands list
**And** it SHALL include "adapt" in the commands list
**And** it SHALL include "canonicalize" in the commands list

---

### Requirement: Enhanced Version Display

The version command SHALL display version, commit SHA, and build date for better debugging.

#### Scenario: Version command shows full info

**Given** germinator is built with version information
**When** a user runs `germinator version`
**Then** it SHALL display format: `germinator {version} ({commit}) {date}`
**And** version SHALL be the semantic version (e.g., v0.3.0)
**And** commit SHALL be 7-character commit SHA (e.g., abc1234)
**And** date SHALL be YYYY-MM-DD format (e.g., 2025-01-13)

#### Scenario: Version with tag

**Given** germinator is built from a Git tag (e.g., v0.3.0)
**When** version command runs
**Then** it SHALL display: `germinator v0.3.0 (abc1234) 2025-01-13`
**And** version SHALL match Git tag
**And** commit SHALL be tag's commit SHA
**And** date SHALL be commit date

#### Scenario: Version without tag

**Given** germinator is built from non-tagged commit
**When** version command runs
**Then** it SHALL display: `germinator v0.3.0-1-gabc1234 (abc1234) 2025-01-13`
**And** version SHALL include git describe output
**And** commit SHALL be current HEAD SHA
**And** date SHALL be current date

---

### Requirement: Version Package Variables

The version package SHALL use variables instead of constants for build-time injection.

#### Scenario: Version is variable

**Given** `internal/version/version.go` is inspected
**When** version variable is declared
**Then** it SHALL use `var` instead of `const`
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "dev"

#### Scenario: Commit is variable

**Given** `internal/version/version.go` is inspected
**When** commit variable is declared
**Then** it SHALL use `var` for commit SHA
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "" (empty string)

#### Scenario: Date is variable

**Given** `internal/version/version.go` is inspected
**When** date variable is declared
**Then** it SHALL use `var` for build date
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "" (empty string)

#### Scenario: Variables exported

**Given** version package is inspected
**When** exports are checked
**Then** `Version` variable SHALL be exported
**And** `Commit` variable SHALL be exported
**And** `Date` variable SHALL be exported

---

### Requirement: Version Command

The version command SHALL display version information for debugging and release tracking.

#### Scenario: Version command works

**Given** germinator is installed
**When** a user runs `germinator version`
**Then** it SHALL execute successfully
**And** it SHALL display version in format: `germinator {version} ({commit}) {date}`
**And** it SHALL exit with code 0

#### Scenario: Version help is available

**Given** a user runs `germinator version --help`
**Then** it SHALL display command help
**And** it SHALL show description: "Show version of germinator"

# cli-error-handling Specification (Delta)

## Purpose

Migrate CLI error handling from per-command `HandleError()` calls to the `RunE` pattern with centralized error handling in `main.go`, and expand exit codes to match the unified standard.

## MODIFIED Requirements

### Requirement: Exit Code Constants

The system SHALL define semantic exit code constants.

#### Scenario: Exit code values

- **WHEN** exit codes are defined
- **THEN** ExitCodeSuccess SHALL be 0
- **AND** ExitCodeError SHALL be 1
- **AND** ExitCodeUsage SHALL be 2
- **AND** ExitCodeConfig SHALL be 3
- **AND** ExitCodeGit SHALL be 4
- **AND** ExitCodeValidation SHALL be 5
- **AND** ExitCodeNotFound SHALL be 6

---

### Requirement: Error Categories

The system SHALL define error categories for exit code mapping.

#### Scenario: Category enumeration

- **WHEN** error categories are defined
- **THEN** CategoryCobra SHALL exist for Cobra framework errors
- **AND** CategoryConfig SHALL exist for configuration errors (renamed from CategoryParse)
- **AND** CategoryValidation SHALL exist for validation errors
- **AND** CategoryNotFound SHALL exist for not-found errors (NEW)
- **AND** CategoryGit SHALL exist for git-related errors (NEW)
- **AND** CategoryGeneric SHALL exist for unclassified errors

#### Scenario: CategoryParse renamed to CategoryConfig

- **GIVEN** current code has `CategoryParse` for parse errors
- **WHEN** the migration is complete
- **THEN** `CategoryParse` SHALL be renamed to `CategoryConfig`
- **AND** parse errors SHALL map to the Config exit code (3)

---

### Requirement: Exit Code Mapping

The system SHALL provide a function to map errors to exit codes.

#### Scenario: Config error exit code

- **WHEN** GetExitCodeForError receives CategoryConfig
- **THEN** it SHALL return ExitCodeConfig (3)

#### Scenario: Validation error exit code

- **WHEN** GetExitCodeForError receives CategoryValidation
- **THEN** it SHALL return ExitCodeValidation (5)

#### Scenario: NotFound error exit code

- **WHEN** GetExitCodeForError receives CategoryNotFound
- **THEN** it SHALL return ExitCodeNotFound (6)

#### Scenario: Git error exit code

- **WHEN** GetExitCodeForError receives CategoryGit
- **THEN** it SHALL return ExitCodeGit (4)

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

Error handling SHALL be centralized in `main.go`.

#### Scenario: main.go handles all errors

- **GIVEN** main.go executes the root command
- **WHEN** rootCmd.Execute() returns an error
- **THEN** HandleCLIError SHALL be called with the command and error
- **AND** the process SHALL exit with the appropriate exit code

#### Scenario: HandleCLIError signature

- **WHEN** HandleCLIError is defined
- **THEN** it SHALL accept `*cobra.Command` and `error` parameters
- **AND** it SHALL return `ExitCode`

#### Scenario: HandleCLIError formats and outputs errors

- **GIVEN** HandleCLIError receives an error
- **WHEN** it executes
- **THEN** it SHALL format the error using ErrorFormatter
- **AND** it SHALL write to stderr
- **AND** it SHALL return the appropriate ExitCode

#### Scenario: Cobra argument errors handled specially

- **GIVEN** HandleCLIError receives a Cobra argument error
- **WHEN** IsCobraArgumentError returns true
- **THEN** the error message SHALL be output without full formatting
- **AND** exit code SHALL be ExitCodeUsage (2)

---

### Requirement: CommandConfig Error Formatter Access

CommandConfig SHALL provide access to ErrorFormatter for command usage.

#### Scenario: Commands access ErrorFormatter via cfg

- **GIVEN** a command receives CommandConfig
- **WHEN** the command needs to format an error
- **THEN** it MAY use cfg.ErrorFormatter
- **AND** the formatted output MAY be returned as part of the error

---

## REMOVED Requirements

### Requirement: Central Error Handler

**Reason:** Replaced by RunE pattern with centralized handling in main.go. The HandleError function that called os.Exit is removed; commands now return errors.

**Migration:**
- Commands using `HandleError(cfg, err)` → return `err`
- main.go handles errors via `HandleCLIError(rootCmd, err)`

### Requirement: Exit Code Parse Constant

**Reason:** Renamed to ExitCodeConfig for clarity. Parse errors now map to Config exit code.

**Migration:**
- `ExitCodeParse` constant → renamed to `ExitCodeConfig` (same value 3)
- `CategoryParse` constant → renamed to `CategoryConfig`
- All references updated to use new names

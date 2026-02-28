# verbose-output Specification

## Purpose

Provide multi-level verbosity control for CLI commands with structured output formatting.

## ADDED Requirements

### Requirement: Verbosity Type

The CLI SHALL provide a Verbosity type for type-safe verbosity level handling.

#### Scenario: Verbosity type exists

- **GIVEN** the cmd/verbose.go file is inspected
- **WHEN** the Verbosity type is defined
- **THEN** it SHALL be based on int type
- **AND** it SHALL have methods IsVerbose() and IsVeryVerbose()

#### Scenario: Verbosity level detection

- **GIVEN** a Verbosity value
- **WHEN** IsVerbose() is called with level >= 1
- **THEN** it SHALL return true
- **WHEN** IsVeryVerbose() is called with level >= 2
- **THEN** it SHALL return true

---

### Requirement: Persistent Verbose Flag

The root command SHALL have a persistent verbose flag available to all subcommands.

#### Scenario: Verbose flag exists on root

- **GIVEN** the CLI is initialized
- **WHEN** `germinator --help` is run
- **THEN** the help output SHALL include `-v` flag
- **AND** the flag description SHALL mention multiple uses for increased verbosity

#### Scenario: Verbose flag increments count

- **GIVEN** the verbose flag is defined with CountP
- **WHEN** user passes `-v`
- **THEN** verbosity level SHALL be 1
- **WHEN** user passes `-vv`
- **THEN** verbosity level SHALL be 2

#### Scenario: Verbose flag is persistent

- **GIVEN** the verbose flag is on root command
- **WHEN** any subcommand is run
- **THEN** the verbose flag SHALL be available
- **AND** `germinator validate file.yaml -v` SHALL work
- **AND** `germinator adapt in.yaml out.md -vv` SHALL work

---

### Requirement: Verbose Output Destination

Verbose output SHALL go to stderr, keeping stdout clean.

#### Scenario: Verbose output to stderr

- **GIVEN** a command is run with `-v` or `-vv`
- **WHEN** verbose messages are printed
- **THEN** they SHALL be written to stderr
- **AND** stdout SHALL contain only normal command output

---

### Requirement: Validate Command Verbose Output

The validate command SHALL support verbose output at two levels.

#### Scenario: Validate with level 1 (-v)

- **GIVEN** a valid document file
- **WHEN** `germinator validate <file> --platform <platform> -v` is run
- **THEN** stderr SHALL include the file path being validated
- **AND** stderr SHALL include the platform name
- **AND** stdout SHALL include "Document is valid"

#### Scenario: Validate with level 2 (-vv)

- **GIVEN** a valid document file
- **WHEN** `germinator validate <file> --platform <platform> -vv` is run
- **THEN** stderr SHALL include loading details with indented format
- **AND** stderr SHALL include parsing details
- **AND** stderr SHALL include validation steps

#### Scenario: Validate with no verbose flag

- **GIVEN** a valid document file
- **WHEN** `germinator validate <file> --platform <platform>` is run
- **THEN** stdout SHALL include "Document is valid"
- **AND** stderr SHALL NOT include verbose messages

---

### Requirement: Adapt Command Verbose Output

The adapt command SHALL support verbose output at two levels.

#### Scenario: Adapt with level 1 (-v)

- **GIVEN** valid input and output paths
- **WHEN** `germinator adapt <input> <output> --platform <platform> -v` is run
- **THEN** stderr SHALL include transformation description
- **AND** stderr SHALL include output path
- **AND** stdout SHALL include success message

#### Scenario: Adapt with level 2 (-vv)

- **GIVEN** valid input and output paths
- **WHEN** `germinator adapt <input> <output> --platform <platform> -vv` is run
- **THEN** stderr SHALL include loading details with indented format
- **AND** stderr SHALL include rendering details
- **AND** stderr SHALL include template path

#### Scenario: Adapt with no verbose flag

- **GIVEN** valid input and output paths
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** stdout SHALL include success message
- **AND** stderr SHALL NOT include verbose messages

---

### Requirement: Canonicalize Command Verbose Output

The canonicalize command SHALL support verbose output at two levels.

#### Scenario: Canonicalize with level 1 (-v)

- **GIVEN** valid input and output paths
- **WHEN** `germinator canonicalize <input> <output> -v` is run
- **THEN** stderr SHALL include canonicalization description
- **AND** stderr SHALL include output path
- **AND** stdout SHALL include success message

#### Scenario: Canonicalize with level 2 (-vv)

- **GIVEN** valid input and output paths
- **WHEN** `germinator canonicalize <input> <output> -vv` is run
- **THEN** stderr SHALL include parsing details with indented format
- **AND** stderr SHALL include validation details

---

### Requirement: Verbose Output Format

Verbose output SHALL follow consistent formatting patterns.

#### Scenario: Level 1 format

- **GIVEN** verbosity level 1
- **WHEN** verbose output is printed
- **THEN** messages SHALL be single-line descriptions
- **AND** messages SHALL NOT be indented

#### Scenario: Level 2 format

- **GIVEN** verbosity level 2
- **WHEN** detailed output is printed
- **THEN** detail lines SHALL be indented with 2 spaces
- **AND** detail lines SHALL describe specific operations

---

### Requirement: Verbose Output Helper Functions

The system SHALL provide helper functions for verbose output.

#### Scenario: VerbosePrint function

- **WHEN** VerbosePrint(cfg, format, args) is called with verbosity >= 1
- **THEN** it SHALL write formatted output to stderr
- **WHEN** verbosity is 0
- **THEN** it SHALL NOT write any output

#### Scenario: VeryVerbosePrint function

- **WHEN** VeryVerbosePrint(cfg, format, args) is called with verbosity >= 2
- **THEN** it SHALL write formatted output to stderr with 2-space indentation
- **WHEN** verbosity is < 2
- **THEN** it SHALL NOT write any output

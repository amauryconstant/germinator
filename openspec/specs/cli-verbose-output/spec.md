# verbose-output Specification

## Purpose

Provide multi-level verbosity control for CLI commands. Verbose output is emitted through `opts.IO.Verbosef` on the `iostreams.IOStreams` struct (see `cli-iostreams`), with the `-v` / `-vv` flag semantics preserved at the Cobra flag layer.

## Requirements

### Requirement: Verbose output via IOStreams.Verbosef

Verbose output SHALL be emitted via `opts.IO.Verbosef(format, args...)` on `iostreams.IOStreams` (introduced in `cli-iostreams`). The legacy `Verbosity` type and `cmd.VerbosePrint` / `cmd.VeryVerbosePrint` helpers SHALL be removed (see `cli-iostreams` for the new mechanism).

#### Scenario: Verbosef writes to ErrOut when verbose

- **WHEN** `opts.IO.Verbose == true`
- **AND** a command calls `opts.IO.Verbosef("loading %d files", 5)`
- **THEN** the formatted string SHALL be written to `opts.IO.ErrOut`
- **AND** a trailing newline SHALL be appended

### Requirement: Verbosity flag semantics preserved

The `-v` (verbose level 1) and `-vv` (verbose level 2) flag semantics SHALL be preserved at the Cobra flag level. The new mechanism (`opts.IO.Verbosef`) SHALL only fire when `opts.IO.Verbose == true`.

#### Scenario: -v sets Verbose=true

- **GIVEN** a command that defines a `-v` flag
- **WHEN** the user invokes the command with `-v`
- **THEN** `opts.IO.Verbose` SHALL be `true`
- **AND** calls to `opts.IO.Verbosef(...)` SHALL write to `opts.IO.ErrOut`

#### Scenario: -vv sets Verbose=true (level 2)

- **GIVEN** a command that defines a `-v` count flag
- **WHEN** the user invokes the command with `-vv`
- **THEN** `opts.IO.Verbose` SHALL be `true`
- **AND** a future `Verbosef2` mechanism MAY distinguish level 1 from level 2 (NOT required by this capability)

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
- **THEN** the flag count SHALL be 1
- **WHEN** user passes `-vv`
- **THEN** the flag count SHALL be 2

#### Scenario: Verbose flag is persistent

- **GIVEN** the verbose flag is on root command
- **WHEN** any subcommand is run
- **THEN** the verbose flag SHALL be available
- **AND** `germinator validate file.yaml -v` SHALL work
- **AND** `germinator adapt in.yaml out.md -vv` SHALL work

#### Scenario: Flag value propagates to opts.IO.Verbose

- **GIVEN** the root command's `-v` count flag was parsed
- **WHEN** a subcommand's `RunE` populates `opts`
- **THEN** `opts.IO.Verbose` SHALL be `true` if the count >= 1, `false` otherwise

### Requirement: Verbose Output Destination

Verbose output SHALL go to stderr, keeping stdout clean.

#### Scenario: Verbose output to stderr

- **GIVEN** a command is run with `-v` or `-vv`
- **WHEN** `opts.IO.Verbosef(...)` is called
- **THEN** the formatted message SHALL be written to `opts.IO.ErrOut`
- **AND** `opts.IO.Out` SHALL contain only normal command output

### Requirement: Validate Command Verbose Output

The validate command SHALL support verbose output via `opts.IO.Verbosef`.

#### Scenario: Validate with -v

- **GIVEN** a valid document file
- **WHEN** `germinator validate <file> --platform <platform> -v` is run
- **THEN** `opts.IO.ErrOut` SHALL include the file path being validated
- **AND** `opts.IO.ErrOut` SHALL include the platform name
- **AND** `opts.IO.Out` SHALL include "Document is valid"

#### Scenario: Validate with -vv

- **GIVEN** a valid document file
- **WHEN** `germinator validate <file> --platform <platform> -vv` is run
- **THEN** `opts.IO.ErrOut` SHALL include loading details with indented format
- **AND** `opts.IO.ErrOut` SHALL include parsing details
- **AND** `opts.IO.ErrOut` SHALL include validation steps

#### Scenario: Validate with no verbose flag

- **GIVEN** a valid document file
- **WHEN** `germinator validate <file> --platform <platform>` is run
- **THEN** `opts.IO.Out` SHALL include "Document is valid"
- **AND** `opts.IO.Verbosef` calls SHALL be no-ops
- **AND** `opts.IO.ErrOut` SHALL NOT include verbose messages

### Requirement: Adapt Command Verbose Output

The adapt command SHALL support verbose output via `opts.IO.Verbosef`.

#### Scenario: Adapt with -v

- **GIVEN** valid input and output paths
- **WHEN** `germinator adapt <input> <output> --platform <platform> -v` is run
- **THEN** `opts.IO.ErrOut` SHALL include transformation description
- **AND** `opts.IO.ErrOut` SHALL include output path
- **AND** `opts.IO.Out` SHALL include success message

#### Scenario: Adapt with -vv

- **GIVEN** valid input and output paths
- **WHEN** `germinator adapt <input> <output> --platform <platform> -vv` is run
- **THEN** `opts.IO.ErrOut` SHALL include loading details with indented format
- **AND** `opts.IO.ErrOut` SHALL include rendering details
- **AND** `opts.IO.ErrOut` SHALL include template path

#### Scenario: Adapt with no verbose flag

- **GIVEN** valid input and output paths
- **WHEN** `germinator adapt <input> <output> --platform <platform>` is run
- **THEN** `opts.IO.Out` SHALL include success message
- **AND** `opts.IO.Verbosef` calls SHALL be no-ops
- **AND** `opts.IO.ErrOut` SHALL NOT include verbose messages

### Requirement: Canonicalize Command Verbose Output

The canonicalize command SHALL support verbose output via `opts.IO.Verbosef`.

#### Scenario: Canonicalize with -v

- **GIVEN** valid input and output paths
- **WHEN** `germinator canonicalize <input> <output> -v` is run
- **THEN** `opts.IO.ErrOut` SHALL include canonicalization description
- **AND** `opts.IO.ErrOut` SHALL include output path
- **AND** `opts.IO.Out` SHALL include success message

#### Scenario: Canonicalize with -vv

- **GIVEN** valid input and output paths
- **WHEN** `germinator canonicalize <input> <output> -vv` is run
- **THEN** `opts.IO.ErrOut` SHALL include parsing details with indented format
- **AND** `opts.IO.ErrOut` SHALL include validation details

### Requirement: Verbose Output Format

Verbose output SHALL follow consistent formatting patterns.

#### Scenario: Single-line format

- **GIVEN** `opts.IO.Verbose == true`
- **WHEN** `opts.IO.Verbosef(...)` is called
- **THEN** the formatted string SHALL be a single line
- **AND** a trailing newline SHALL be appended

#### Scenario: Level 2 details are indented

- **GIVEN** `opts.IO.Verbose == true`
- **WHEN** a command emits a level-2 detail line
- **THEN** the line SHALL be prefixed with 2 spaces
- **AND** the line SHALL describe a specific operation

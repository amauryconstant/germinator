# verbose-output Specification (delta)

## MODIFIED Requirements

### Requirement: Verbosity type replaced

The `cmd.Verbosity` type and the `cmd.VerbosePrint(cfg *CommandConfig, ...)` and `cmd.VeryVerbosePrint(cfg *CommandConfig, ...)` helpers SHALL be **removed**. Verbose output SHALL be emitted via `opts.IO.Verbosef(format, args...)` on `iostreams.IOStreams` (introduced in `cli/iostreams`).

#### Scenario: Verbosity type removed

- **WHEN** the codebase is inspected
- **THEN** the `Verbosity` type SHALL NOT be defined
- **AND** `VerbosePrint` and `VeryVerbosePrint` SHALL NOT be defined
- **AND** `cmd/verbose.go` SHALL be deleted (deletion happens in change-7 after all consumers migrated)

> **Status (slice 1 / foundation):** `iostreams.Verbosef` exists on `IOStreams`. `Verbosity` and `VerbosePrint` still exist in `cmd/verbose.go`. Adoption in commands happens in changes 2-9; file deletion in change-7.

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
- **AND** a future `Verbosef2` mechanism MAY distinguish level 1 from level 2 (NOT required by this change)

> **Status (slice 1 / foundation):** the flag parsing logic still uses the legacy `Verbosity` type. Migration of each command's flag wiring happens in changes 2-9.

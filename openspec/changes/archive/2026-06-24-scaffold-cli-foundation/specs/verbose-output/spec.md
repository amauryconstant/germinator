# verbose-output Specification (delta)

## MODIFIED Requirements

### Requirement: Verbose output via IOStreams.Verbosef

Verbose output SHALL be emitted via `opts.IO.Verbosef(format, args...)` on `iostreams.IOStreams` (introduced in `iostreams`). The legacy `Verbosity` type and `cmd.VerbosePrint` / `cmd.VeryVerbosePrint` helpers SHALL be deleted in change-7 (see `## REMOVED Requirements` below).

#### Scenario: Verbosef writes to ErrOut when verbose

- **WHEN** `opts.IO.Verbose == true`
- **AND** a command calls `opts.IO.Verbosef("loading %d files", 5)`
- **THEN** the formatted string SHALL be written to `opts.IO.ErrOut`
- **AND** a trailing newline SHALL be appended

> **Status (slice 1 / foundation):** `iostreams.Verbosef` exists on `IOStreams` with table-driven tests. The legacy `Verbosity` and helpers still exist in `cmd/verbose.go`. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

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

> **Status (slice 1 / foundation):** the flag parsing logic still uses the legacy `Verbosity` type. Migration of each command's flag wiring happens in changes 2-9. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

## REMOVED Requirements

### Requirement: Verbosity type and VerbosePrint helpers

**Reason**: `Verbosity` and `VerbosePrint`/`VeryVerbosePrint` are tightly coupled to the deleted `CommandConfig` (they take `cfg *CommandConfig` as first argument) and emit via `fmt.Fprintf(os.Stderr, ...)` rather than through an injected writer.

**Migration**: Replace `cfg.Verbosity.VerbosePrint(...)` with `opts.IO.Verbosef(...)`. The `Verbose bool` field on `IOStreams` is set from `-v`/`-vv` flag parsing in each migrated command.

#### Scenario: Legacy verbose helpers removed

- **WHEN** the codebase is inspected after change-7
- **THEN** the `Verbosity` type SHALL NOT be defined
- **AND** `VerbosePrint` and `VeryVerbosePrint` SHALL NOT be defined
- **AND** `cmd/verbose.go` SHALL be deleted

> **Status (slice 1 / foundation):** legacy helpers still exist. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

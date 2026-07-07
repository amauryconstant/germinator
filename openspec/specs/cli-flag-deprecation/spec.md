# cli-flag-deprecation Specification

## Purpose

Define the deprecation policy for CLI flags and commands. The policy ensures that breaking changes are telegraphed ahead of time, are easy to find in `--help` output, and follow a predictable removal cadence.

## Requirements

### Requirement: MarkDeprecated for flags

The CLI SHALL use Cobra's `cmd.Flags().MarkDeprecated(name, message)` to mark a flag as deprecated. Deprecated flags continue to work but print a deprecation warning to `ErrOut` on each invocation.

#### Scenario: Deprecated flag prints warning

- **GIVEN** flag `--old-name` is marked deprecated with message `"use --new-name instead"`
- **WHEN** `germinator <cmd> --old-name foo` is invoked
- **THEN** `ErrOut` SHALL contain: `Flag --old-name has been deprecated, use --new-name instead`
- **AND** the command SHALL proceed using the deprecated flag's value
- **AND** exit code SHALL be 0 on success

#### Scenario: Deprecated flag still appears in --help

- **WHEN** `germinator <cmd> --help` is invoked
- **THEN** the help output SHALL list the deprecated flag with `[deprecated]` annotation

### Requirement: Removal cadence

Deprecated flags SHALL be removed after 2–3 minor versions of deprecation, with the deprecation announcement in CHANGELOG.md.

#### Scenario: Three-minor-version deprecation window

- **GIVEN** flag `--old-name` is deprecated in version `v1.5.0`
- **WHEN** version `v1.8.0` is released
- **THEN** the deprecated flag MAY be removed (3 minor versions elapsed)
- **AND** CHANGELOG.md SHALL record the deprecation date and projected removal version

### Requirement: Hard renames are not deprecation

Flags that have been renamed without a deprecation window (the rename is the breaking change) SHALL NOT use `MarkDeprecated`. They SHALL be documented in CHANGELOG.md and in the spec's "Note on flag naming" footnote (see `cli-config-commands/spec.md` for the `--output` → `--output-path` precedent).

#### Scenario: --output → --output-path hard rename

- **GIVEN** `germinator config init --output-path /tmp/config.toml` is the new canonical form
- **WHEN** `germinator config init --output /tmp/config.toml` is invoked (legacy flag)
- **THEN** Cobra SHALL emit an "unknown flag" usage error
- **AND** exit code SHALL be 2 (`ExitCodeUsage`)
- **AND** no deprecation warning SHALL be emitted (the flag was removed, not deprecated)

### Requirement: Command deprecation

Cobra commands (not just flags) MAY be marked deprecated via the `Deprecated` field. Deprecated commands SHALL print a warning on each invocation.

#### Scenario: Deprecated command

- **GIVEN** command `oldcmd` has `Deprecated: "use 'newcmd' instead"`
- **WHEN** `germinator oldcmd` is invoked
- **THEN** `ErrOut` SHALL contain: `Command "oldcmd" is deprecated, use 'newcmd' instead`
- **AND** the command SHALL still execute its body

### Requirement: Deprecation must be visible

A deprecation SHALL NOT be silent. The user SHALL encounter the warning via:

- `--help` output (annotation)
- Invocation of the deprecated flag/command (warning to `ErrOut`)
- CHANGELOG.md (visible on the project's release notes)

#### Scenario: --help annotation

- **GIVEN** flag `--foo` is deprecated
- **WHEN** `germinator <cmd> --help` is invoked
- **THEN** the deprecated flag SHALL appear with `[deprecated]` or similar suffix in the flag list

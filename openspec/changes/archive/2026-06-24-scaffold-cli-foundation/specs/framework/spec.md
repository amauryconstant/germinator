# framework Specification (delta)

## MODIFIED Requirements

### Requirement: Commands take Factory, not CommandConfig

Each command's constructor SHALL take `*cmdutil.Factory` as its first parameter (after the optional `runF` for test injection). No command SHALL take `*CommandConfig`. The `CommandConfig` type and `cmd/command_config.go` SHALL be deleted in change-7 (see `## REMOVED Requirements` below).

#### Scenario: NewCmdXxx signature

- **WHEN** a command's constructor signature is inspected
- **THEN** it SHALL match `NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command`
- **AND** it SHALL NOT have any parameter of type `*CommandConfig`

> **Status (slice 1 / foundation):** the new signature is established; no command uses it yet. Adoption happens in changes 2-9. The legacy `CommandConfig` and `cmd/command_config.go` are removed in changes 2 and 7 as noted in the REMOVED section.

### Requirement: No global CommandConfig

The `cmd.SetGlobalCommandConfig(*CommandConfig)` function and any package-level `CommandConfig` variable SHALL be **removed**. All command state SHALL flow through `opts`.

#### Scenario: No global command config

- **WHEN** the codebase is inspected
- **THEN** there SHALL be no `var globalConfig *CommandConfig` or similar
- **AND** there SHALL be no `SetGlobalCommandConfig` function
- **AND** no command SHALL call `cmd.GetCommandConfig()` or similar getter

> **Status (slice 1 / foundation):** `SetGlobalCommandConfig` and the global variable still exist. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

## REMOVED Requirements

### Requirement: CommandConfig struct and global setter

**Reason**: `CommandConfig` is a mutable, package-level singleton shared by reference; commands read and write its fields directly, which makes concurrent command execution and unit testing brittle.

**Migration**: Use `*cmdutil.Factory` (passed to each `NewCmdXxx`) instead. Read eagerly-stamped fields like `AppVersion` directly from the Factory; lazy fields (`Config`, `Library`, etc.) via their `func() (T, error)` accessor. See `specs/cli-factory/spec.md`.

#### Scenario: CommandConfig removed

- **WHEN** the codebase is inspected after change-7
- **THEN** the `CommandConfig` type SHALL NOT be defined
- **AND** `NewCommandConfig` SHALL NOT be defined
- **AND** `cmd/command_config.go` SHALL be deleted

> **Status (slice 1 / foundation):** `CommandConfig` still exists in `cmd/command_config.go`. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

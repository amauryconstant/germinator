# framework Specification (delta)

## MODIFIED Requirements

### Requirement: CommandConfig removed

The `cmd.CommandConfig` struct and the `cmd.NewCommandConfig(...)` constructor SHALL be **removed**. Commands SHALL take `*cmdutil.Factory` (introduced in `cli/cli-factory`) and populate a per-command `XxxOptions` struct.

#### Scenario: CommandConfig type removed

- **WHEN** the codebase is inspected
- **THEN** the `CommandConfig` type SHALL NOT be defined
- **AND** `NewCommandConfig` SHALL NOT be defined
- **AND** `cmd/command_config.go` SHALL be deleted (deletion happens in change-2)

> **Status (slice 1 / foundation):** `CommandConfig` still exists in `cmd/command_config.go`. Deletion of the file happens in change-2.

### Requirement: Commands take Factory, not CommandConfig

Each command's constructor SHALL take `*cmdutil.Factory` as its first parameter (after the optional `runF` for test injection). No command SHALL take `*CommandConfig`.

#### Scenario: NewCmdXxx signature

- **WHEN** a command's constructor signature is inspected
- **THEN** it SHALL match `NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command`
- **AND** it SHALL NOT have any parameter of type `*CommandConfig`

> **Status (slice 1 / foundation):** no command uses the new signature yet. Adoption happens in changes 2-9.

### Requirement: No global CommandConfig

The `cmd.SetGlobalCommandConfig(*CommandConfig)` function and any package-level `CommandConfig` variable SHALL be **removed**. All command state SHALL flow through `opts`.

#### Scenario: No global command config

- **WHEN** the codebase is inspected
- **THEN** there SHALL be no `var globalConfig *CommandConfig` or similar
- **AND** there SHALL be no `SetGlobalCommandConfig` function
- **AND** no command SHALL call `cmd.GetCommandConfig()` or similar getter

> **Status (slice 1 / foundation):** `SetGlobalCommandConfig` and the global variable still exist. Removal happens in change-2 (delete `cmd/command_config.go`).

# Capability: Library Remove Preset

## Purpose

The Library Remove Preset capability enables removing existing presets from the library through the CLI. It validates preset existence, updates library.yaml, and supports both human-readable and JSON output formats.

## Requirements

### Requirement: Remove preset command exists

The system SHALL provide a `library remove preset <name>` command that removes a preset from the library.

#### Scenario: Remove preset subcommand is registered
- **WHEN** the user runs `germinator library remove --help`
- **THEN** the help output includes `preset` as a subcommand

### Requirement: Remove preset requires name

The system SHALL require a preset name as a non-empty string.

#### Scenario: Remove with valid name
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** the system attempts to remove the preset

#### Scenario: Remove with empty name
- **WHEN** user runs `germinator library remove preset ""`
- **THEN** the system returns an error about missing preset name

### Requirement: Remove preset validates library exists

The system SHALL error if the library does not exist at the specified path.

#### Scenario: Remove from non-existent library
- **WHEN** user runs `germinator library remove preset git-workflow --library /nonexistent`
- **THEN** the system returns an error that the library does not exist

### Requirement: Remove preset validates preset exists

The system SHALL error if the specified preset does not exist in the library.

#### Scenario: Remove non-existent preset
- **WHEN** user runs `germinator library remove preset nonexistent`
- **AND** the library does not contain a preset named `nonexistent`
- **THEN** the system returns an error: `preset not found`

### Requirement: Remove preset updates library.yaml

The system SHALL remove the preset entry from `library.yaml`.

#### Scenario: Update library.yaml on remove
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** the `presets.git-workflow` entry is removed from `library.yaml`

### Requirement: Remove preset has no physical file

The system SHALL NOT attempt to delete any physical files when removing a preset (presets exist only in library.yaml).

#### Scenario: No physical file deletion
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** no file outside of `library.yaml` is modified or deleted

### Requirement: Remove preset validates after update

The system SHALL validate the updated library.yaml after removal.

#### Scenario: Validation of updated library
- **WHEN** user runs `germinator library remove preset git-workflow`
- **THEN** the system loads the updated library to validate it is well-formed

### Requirement: Remove preset outputs success message

The system SHALL print a human-readable success message when removal succeeds.

#### Scenario: Success output
- **WHEN** user runs `germinator library remove preset git-workflow`
- **AND** removal succeeds
- **THEN** the system prints: `Removed preset: git-workflow`

### Requirement: Remove preset uses library path resolution

The system SHALL resolve the library path using the same priority as other library commands: explicit `--library` flag > `GERMINATOR_LIBRARY` env > default path.

#### Scenario: Library path from default
- **WHEN** user runs `germinator library remove preset git-workflow`
- **AND** no explicit library path is provided
- **THEN** the system uses the default library path

#### Scenario: Library path from environment
- **WHEN** `GERMINATOR_LIBRARY` is set to `/custom/path`
- **AND** user runs `germinator library remove preset git-workflow`
- **THEN** the system uses `/custom/path`

#### Scenario: Library path from flag
- **WHEN** user runs `germinator library remove preset git-workflow --library /custom/path`
- **THEN** the system uses `/custom/path`

### Requirement: library remove preset follows command-options-pattern

The `library remove preset` sub-command SHALL adopt the `NewCmdRemove(f, runF)` + `runRemove(opts)` template. When `opts.PresetName` is non-empty, the function calls `Library.RemovePreset(ctx, req)` instead of `Library.RemoveResource(ctx, req)`.

#### Scenario: Preset removal dispatch

- **GIVEN** a library with an existing preset `git-workflow`
- **WHEN** `germinator library remove preset git-workflow` is invoked
- **THEN** `runRemove(opts)` SHALL call `lib.RemovePreset(opts.Ctx, &RemovePresetRequest{Name: opts.PresetName, Force: opts.Force})`
- **AND** `opts.PresetName` SHALL be populated from the positional `args[0]` of `cobra.ExactArgs(1)` — the legacy CLI surface is preserved (no flag substitution)

#### Scenario: Library interface method

- **GIVEN** `cmd/library_remove.go` declares its `removerLibrary` interface (shared with the resource removal scenario above)
- **WHEN** the interface is inspected
- **THEN** it SHALL declare a `RemovePreset(ctx context.Context, req *RemovePresetRequest) error` method
- **AND** the interface SHALL be satisfied directly by `*library.Library` (the `RemovePreset` method is added to `*Library` in change-7, parallel to `RemoveResource`)
- **AND** `RemovePresetRequest` SHALL be a type defined in `internal/library/requests.go` with `Name string` and `Force bool` fields

### Requirement: library remove preset supports --output flag

The `library remove preset` sub-command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags` (inherited from the parent `library remove` command).

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience. The parent `library remove` command continues to wire the `--output` flag via `cmd.PersistentFlags()` (inherited by `resource` and `preset` sub-commands); see the `PersistentFlags wiring for parent commands` requirement in `cli-output-formats`.

#### Scenario: Default plain output

- **GIVEN** a library with an existing preset `git-workflow`
- **WHEN** `germinator library remove preset git-workflow` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the removal

#### Scenario: JSON output

- **GIVEN** a library with an existing preset `git-workflow`
- **WHEN** `germinator library remove preset git-workflow --output json` is invoked
- **THEN** the result SHALL be JSON-formatted
- **AND** the payload SHALL include `type: "preset"`, `name`, and `resourcesRemoved` fields

#### Scenario: Table output

- **GIVEN** a library with an existing preset `git-workflow`
- **WHEN** `germinator library remove preset git-workflow --output table` is invoked
- **THEN** the output SHALL be a table with columns: name, action
- **AND** the removed preset SHALL appear as a single row

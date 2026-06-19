# library-remove-preset Specification (delta)

## MODIFIED Requirements

### Requirement: library remove preset follows command-options-pattern

The `library remove preset` sub-command SHALL adopt the `NewCmdRemove(f, runF)` + `runRemove(opts)` template. When `opts.PresetName` is non-empty, the function calls `Library.RemovePreset(ctx, req)` instead of `Library.RemoveResource(ctx, req)`.

#### Scenario: Preset removal dispatch

- **WHEN** `germinator library remove preset <name>` is invoked
- **THEN** `runRemove(opts)` SHALL call `lib.RemovePreset(opts.Ctx, &RemovePresetRequest{Name: opts.PresetName, Force: opts.Force})`

### Requirement: library remove preset supports --output flag

The `library remove preset` sub-command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags` (inherited from the parent `library remove` command).

#### Scenario: Default plain output

- **WHEN** `germinator library remove preset <name>` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the removal

#### Scenario: JSON output

- **WHEN** `germinator library remove preset <name> --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

> **Status:** the `--output` flag is added to `library remove preset` in change-7 (`migrate-library-rest`).

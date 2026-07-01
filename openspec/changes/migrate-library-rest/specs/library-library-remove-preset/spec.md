# library-library-remove-preset Specification (delta)

## MODIFIED Requirements

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

The `library remove preset` sub-command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags` (inherited from the parent `library remove` command).

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

> **Status:** the `--output` flag is added to `library remove preset` in change-7 (`migrate-library-rest`).

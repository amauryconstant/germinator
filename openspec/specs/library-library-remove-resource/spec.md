# Capability: Library Remove Resource

## Purpose

The Library Remove Resource capability enables removing existing resources from the library through the CLI. It validates resource existence, checks preset references, deletes physical files, updates library.yaml, and supports both human-readable and JSON output formats.

## Requirements

### Requirement: Remove resource command exists

The system SHALL provide a `library remove resource <ref>` command that removes a resource from the library.

#### Scenario: Remove resource subcommand is registered
- **WHEN** the user runs `germinator library remove --help`
- **THEN** the help output includes `resource` as a subcommand

### Requirement: Remove resource requires valid reference

The system SHALL require a resource reference in `type/name` format (e.g., `skill/commit`).

#### Scenario: Remove with valid reference
- **WHEN** user runs `germinator library remove resource skill/commit`
- **THEN** the system attempts to remove the resource

#### Scenario: Remove with invalid reference format
- **WHEN** user runs `germinator library remove resource invalid`
- **THEN** the system returns an error: `invalid resource reference format (expected type/name)`

#### Scenario: Remove with empty reference
- **WHEN** user runs `germinator library remove resource ""`
- **THEN** the system returns an error about missing resource reference

### Requirement: Remove resource validates library exists

The system SHALL error if the library does not exist at the specified path.

#### Scenario: Remove from non-existent library
- **WHEN** user runs `germinator library remove resource skill/commit --library /nonexistent`
- **THEN** the system returns an error that the library does not exist

### Requirement: Remove resource validates resource exists

The system SHALL error if the specified resource does not exist in the library.

#### Scenario: Remove non-existent resource
- **WHEN** user runs `germinator library remove resource skill/nonexistent`
- **AND** the library does not contain `skill/nonexistent`
- **THEN** the system returns an error: `resource skill/nonexistent not found`

### Requirement: Remove resource checks preset references

The system SHALL error if any preset in the library references the resource being removed.

#### Scenario: Remove resource referenced by preset
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** a preset `git-workflow` contains `skill/commit` in its resources
- **THEN** the system returns an error: `resource skill/commit is referenced by preset git-workflow`

#### Scenario: Remove resource not referenced by any preset
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** no presets reference `skill/commit`
- **THEN** the system proceeds with removal

### Requirement: Remove resource deletes physical file

The system SHALL delete the physical resource file from the library directory.

#### Scenario: Delete physical file on remove
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** the resource exists at `library/skills/commit.md`
- **THEN** the file `library/skills/commit.md` is deleted

### Requirement: Remove resource updates library.yaml

The system SHALL remove the resource entry from `library.yaml`.

#### Scenario: Update library.yaml on remove
- **WHEN** user runs `germinator library remove resource skill/commit`
- **THEN** the `resources.skill.commit` entry is removed from `library.yaml`

### Requirement: Remove resource validates after update

The system SHALL validate the updated library.yaml after removal.

#### Scenario: Validation of updated library
- **WHEN** user runs `germinator library remove resource skill/commit`
- **THEN** the system loads the updated library to validate it is well-formed

### Requirement: Remove resource outputs success message

The system SHALL print a human-readable success message when removal succeeds.

#### Scenario: Success output
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** removal succeeds
- **THEN** the system prints: `Removed resource: skill/commit`

### Requirement: Remove resource uses library path resolution

The system SHALL resolve the library path using the same priority as other library commands: explicit `--library` flag > `GERMINATOR_LIBRARY` env > default path.

#### Scenario: Library path from default
- **WHEN** user runs `germinator library remove resource skill/commit`
- **AND** no explicit library path is provided
- **THEN** the system uses the default library path

#### Scenario: Library path from environment
- **WHEN** `GERMINATOR_LIBRARY` is set to `/custom/path`
- **AND** user runs `germinator library remove resource skill/commit`
- **THEN** the system uses `/custom/path`

#### Scenario: Library path from flag
- **WHEN** user runs `germinator library remove resource skill/commit --library /custom/path`
- **THEN** the system uses `/custom/path`

### Requirement: library remove resource follows command-options-pattern

The `library remove resource` sub-command SHALL adopt the `NewCmdRemove(f, runF)` + `runRemove(opts)` template.

#### Scenario: removeOptions struct

- **GIVEN** the `library remove` command has been migrated
- **WHEN** `cmd/library_remove.go` is inspected
- **THEN** it SHALL declare `removeOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Ref string`, `PresetName string`, `Force bool`, `Output string`
- **AND** the struct SHALL NOT carry `ResourceType` or `ResourceName` fields — the legacy positional `<ref>` argument is preserved (no `--type` / `--name` flag substitution)

#### Scenario: Library interface method

- **GIVEN** `cmd/library_remove.go` declares its `removerLibrary` interface
- **WHEN** the interface is inspected
- **THEN** it SHALL declare a `RemoveResource(ctx context.Context, req *RemoveResourceRequest) error` method
- **AND** the interface SHALL be satisfied directly by `*library.Library` (the `RemoveResource` method is added to `*Library` in change-7)
- **AND** `var _ removerLibrary = (*library.Library)(nil)` SHALL be a compile-time check at the bottom of `cmd/library_remove.go`
- **AND** `RemoveResourceRequest` SHALL be a type defined in `internal/library/requests.go` with `Ref string` (a `"type/name"` string like `"skill/commit"`) and `Force bool` fields

#### Scenario: Positional ref argument preserved

- **GIVEN** the legacy CLI surface `germinator library remove resource <ref>`
- **WHEN** the migrated command is invoked as `germinator library remove resource skill/commit`
- **THEN** the command SHALL parse `args[0]` as `opts.Ref` via the `Args: cobra.ExactArgs(1)` validator
- **AND** the command SHALL pass `opts.Ref` unchanged into `RemoveResourceRequest.Ref`
- **AND** no `--type` or `--name` flag SHALL be required or accepted

### Requirement: library remove resource supports --output flag

The `library remove resource` sub-command SHALL expose a `--output json|table|plain` flag via `output.AddOutputFlags`.

**Change**: rehome the function reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of 7 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Default plain output

- **GIVEN** a library with an existing resource `skill/commit`
- **WHEN** `germinator library remove resource skill/commit` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the removal

#### Scenario: JSON output

- **GIVEN** a library with an existing resource `skill/commit`
- **WHEN** `germinator library remove resource skill/commit --output json` is invoked
- **THEN** the result SHALL be JSON-formatted
- **AND** the payload SHALL include `type`, `name`, and `fileDeleted` fields

#### Scenario: Table output

- **GIVEN** a library with an existing resource `skill/commit`
- **WHEN** `germinator library remove resource skill/commit --output table` is invoked
- **THEN** the output SHALL be a table with columns: ref, action
- **AND** the removed resource SHALL appear as a single row

#### Scenario: --force flag

- **GIVEN** a library with an existing resource `skill/commit`
- **WHEN** `germinator library remove resource skill/commit --force` is invoked
- **THEN** the command SHALL skip confirmation prompts and remove the resource unconditionally

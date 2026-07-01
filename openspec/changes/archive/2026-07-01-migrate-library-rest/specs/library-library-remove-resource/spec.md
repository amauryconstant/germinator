# library-library-remove-resource Specification (delta)

## MODIFIED Requirements

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

The `library remove resource` sub-command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

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

> **Status:** the `--output` flag is added to `library remove` in change-7 (`migrate-library-rest`).

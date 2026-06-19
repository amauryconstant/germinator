# library-remove-resource Specification (delta)

## MODIFIED Requirements

### Requirement: library remove resource follows command-options-pattern

The `library remove resource` sub-command SHALL adopt the `NewCmdRemove(f, runF)` + `runRemove(opts)` template.

#### Scenario: removeOptions struct

- **WHEN** `cmd/library/remove.go` is inspected
- **THEN** it SHALL declare `removeOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `ResourceType string`, `ResourceName string`, `PresetName string`, `Force bool`, `Output string`

### Requirement: library remove resource supports --output flag

The `library remove resource` sub-command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **WHEN** `germinator library remove resource <type>/<name>` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the removal

#### Scenario: JSON output

- **WHEN** `germinator library remove resource <type>/<name> --output json` is invoked
- **THEN** the result SHALL be JSON-formatted

#### Scenario: --force flag

- **WHEN** `germinator library remove resource <type>/<name> --force` is invoked
- **THEN** the command SHALL skip confirmation prompts and remove the resource unconditionally

> **Status:** the `--output` flag is added to `library remove` in change-7 (`migrate-library-rest`).

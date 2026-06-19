# framework Specification (delta)

## MODIFIED Requirements

### Requirement: validate and canonicalize take Factory

The `validate` and `canonicalize` commands SHALL take `*cmdutil.Factory` (per the `cli-factory` capability) instead of `*CommandConfig`. They SHALL adopt the `command-options-pattern` shape (`NewCmdValidate(f, runF)` + `validateOptions` + `runValidate`; `NewCmdCanonicalize(f, runF)` + `canonicalizeOptions` + `runCanonicalize`).

#### Scenario: validate command signature

- **WHEN** `cmd/validate.go` is inspected
- **THEN** the constructor SHALL be `NewCmdValidate(f *cmdutil.Factory, runF func(*validateOptions) error) *cobra.Command`
- **AND** it SHALL NOT have any parameter of type `*CommandConfig`
- **AND** `validateOptions` SHALL declare `IO *iostreams.IOStreams`, `Validator func() (Validator, error)`, `Ctx context.Context`, `InputPath string`, `Platform string`

#### Scenario: canonicalize command signature

- **WHEN** `cmd/canonicalize.go` is inspected
- **THEN** the constructor SHALL be `NewCmdCanonicalize(f *cmdutil.Factory, runF func(*canonicalizeOptions) error) *cobra.Command`
- **AND** `canonicalizeOptions` SHALL declare `IO *iostreams.IOStreams`, `Canonicalizer func() (Canonicalizer, error)`, `Ctx context.Context`, `InputPath string`, `OutputPath string`, `Platform string`, `DocType string`

> **Status:** the migration is implemented in change-3 (`migrate-domain-commands`). The internal `internal/service/validator.go` and `internal/service/canonicalizer.go` are deleted; their logic moves into the per-command files as private helpers.

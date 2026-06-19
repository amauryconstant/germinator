# config-commands Specification (delta)

## MODIFIED Requirements

### Requirement: config init follows command-options-pattern

The `config init` command SHALL adopt the `NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command` + `runConfigInit(opts *configInitOptions) error` template.

#### Scenario: configInitOptions struct

- **WHEN** `cmd/config/init.go` is inspected
- **THEN** it SHALL declare `configInitOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`, `Force bool`

### Requirement: --output renamed to --output-path

The `config init` and `config validate` commands SHALL accept `--output-path <file>` instead of the legacy `--output <file>`. The rename disambiguates from the `output-formats` capability's `--output` format flag.

#### Scenario: --output-path flag

- **WHEN** `germinator config init --help` is invoked
- **THEN** the help output SHALL include `--output-path <file>` (not `--output`)

#### Scenario: Legacy --output returns usage error

- **WHEN** `germinator config init --output /tmp/config.toml` is invoked
- **THEN** the command SHALL return a usage error (exit 2) because `--output` is no longer a recognized flag

#### Scenario: Default output path

- **WHEN** `germinator config init` is invoked without `--output-path`
- **THEN** the default path SHALL be `$XDG_CONFIG_HOME/germinator/config.toml` if `XDG_CONFIG_HOME` is set
- **AND** `~/.config/germinator/config.toml` otherwise

### Requirement: --force flag preserved

The `config init --force` flag SHALL be preserved; it allows overwriting an existing config file.

#### Scenario: --force overwrites existing file

- **WHEN** `germinator config init --force` is invoked AND the default config file already exists
- **THEN** the file SHALL be overwritten with the config template

#### Scenario: Without --force, existing file fails

- **WHEN** `germinator config init` is invoked AND the default config file already exists
- **THEN** the command SHALL return `*core.ConfigError{Message: "config file already exists; use --force to overwrite"}`
- **AND** `cmdutil.ExitCodeFor` SHALL return `ExitCodeError` (1)

### Requirement: config validate follows command-options-pattern

The `config validate` command SHALL adopt the `NewCmdConfigValidate(f *cmdutil.Factory, runF func(*configValidateOptions) error) *cobra.Command` + `runConfigValidate(opts *configValidateOptions) error` template.

#### Scenario: configValidateOptions struct

- **WHEN** `cmd/config/validate.go` is inspected
- **THEN** it SHALL declare `configValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`

### Requirement: config validate renders errors via FormatError

When `config validate` encounters validation errors, each error SHALL be rendered via `output.FormatError(opts.IO, err)`.

#### Scenario: Invalid field type

- **WHEN** `germinator config validate --output-path /tmp/bad-config.toml` is invoked AND the file has an invalid field type
- **THEN** the command SHALL return `*core.ConfigError` describing the invalid field
- **AND** `output.FormatError` SHALL render it to `opts.IO.ErrOut`

> **Status:** the `--output` → `--output-path` rename is implemented in change-8 (`migrate-config-commands`). This is a BREAKING change; CHANGELOG entry documents it.

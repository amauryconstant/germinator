# Capability: Config Commands

## Purpose

CLI commands for scaffolding and validating Germinator configuration files.

## Requirements

### Requirement: Config init scaffolds a new config file

The `germinator config init` command SHALL create a new config file with explanatory comments for each field. The legacy `--output` flag is renamed to `--output-path` to disambiguate from the `cli-output-formats` capability's `--output` format flag.

#### Scenario: Init creates config at default location

- **WHEN** user runs `germinator config init` with no flags
- **AND** no config file exists at the default location
- **THEN** system creates the config file at `~/.config/germinator/config.toml`
- **AND** file contains comments explaining each field

#### Scenario: Init creates config at custom location

- **WHEN** user runs `germinator config init --output-path /custom/path/config.toml`
- **AND** no file exists at `/custom/path/config.toml`
- **THEN** system creates the config file at `/custom/path/config.toml`
- **AND** parent directories are created with permissions 0750

#### Scenario: Legacy --output returns usage error

- **WHEN** `germinator config init --output /tmp/config.toml` is invoked
- **THEN** the command SHALL return a usage error (exit 2) because `--output` is no longer a recognized flag

#### Scenario: Default output path

- **WHEN** `germinator config init` is invoked without `--output-path`
- **THEN** the default path SHALL be `$XDG_CONFIG_HOME/germinator/config.toml` if `XDG_CONFIG_HOME` is set
- **AND** `~/.config/germinator/config.toml` otherwise

#### Scenario: Init refuses to overwrite without force

- **WHEN** `germinator config init` is invoked with no `--force` flag
- **AND** a config file already exists at the target location
- **THEN** the command SHALL return `core.NewFileError(OutputPath, "create", "config file already exists (use --force to overwrite)", nil)` (constructed via the constructor — `FileError` fields are unexported)
- **AND** `main.go` SHALL render it via `output.FormatError` to `opts.IO.ErrOut`
- **AND** no file SHALL be written

#### Scenario: Init overwrites with force flag

- **WHEN** `germinator config init --force` is invoked
- **AND** a config file already exists at the target location
- **THEN** the file SHALL be overwritten with the scaffolded content

#### Scenario: Init produces byte-identical output

- **WHEN** `germinator config init --output-path /tmp/golden-config.toml` is invoked on an empty path
- **THEN** the generated file content SHALL be byte-identical to the pre-change build's output (golden file comparison)

### Requirement: Config init output contains all fields

The scaffolded config file SHALL contain entries for all configurable fields with explanatory comments. All field values SHALL be commented out by default, requiring users to explicitly uncomment settings they wish to override.

#### Scenario: Config contains library field

- **WHEN** config file is scaffolded
- **THEN** file contains `# library` field with comment explaining it accepts a path
- **AND** default value `~/.local/share/germinator/library` is set (uses `XDG_DATA_HOME` if set)

#### Scenario: Config contains platform field

- **WHEN** config file is scaffolded
- **THEN** file contains `# platform` field with comment explaining valid values
- **AND** default value is empty string (requiring --platform flag)

#### Scenario: Config contains completion settings

- **WHEN** config file is scaffolded
- **THEN** file contains `[completion]` table header (active, not commented)
- **AND** file contains `# timeout` field with comment explaining its purpose and default value
- **AND** file contains `# cache_ttl` field with comment explaining its purpose and default value

### Requirement: Config validate checks existing config

The `germinator config validate` command SHALL validate an existing config file. The command SHALL return any error (rendered once at the boundary by `main.go`); on success it SHALL write a single success line to `opts.IO.Out` and nothing to `opts.IO.ErrOut`.

#### Scenario: Validate succeeds for valid config

- **WHEN** `germinator config validate --output-path /tmp/valid-config.toml` is invoked AND the file contains a valid TOML config
- **THEN** the command SHALL return nil
- **AND** a single success line SHALL be written to `opts.IO.Out`
- **AND** nothing SHALL be written to `opts.IO.ErrOut`

#### Scenario: Validate fails when file not found

- **WHEN** `germinator config validate --output-path /nonexistent.toml` is invoked AND the path does not exist
- **THEN** the command SHALL return `core.NewFileError(opts.OutputPath, "read", "config file not found", <os.Stat error>)`
- **AND** `FileError.IsNotFound()` SHALL return true (derived from the wrapped cause)
- **AND** `main.go` SHALL render it once via `output.FormatError` to `opts.IO.ErrOut`

#### Scenario: Validate fails on malformed TOML

- **WHEN** the config file exists but contains invalid TOML syntax
- **THEN** the command SHALL return a parse error carrying the file path
- **AND** `main.go` SHALL render it once via `output.FormatError`

#### Scenario: Validate fails on invalid platform value

- **WHEN** the config file contains `platform = "unknown"`
- **THEN** the command SHALL return the `*core.ConfigError` produced by `config.Validate()` (platform-only scope today)
- **AND** `main.go` SHALL render it once via `output.FormatError` to `opts.IO.ErrOut`

### Requirement: Config validate uses specified output path

The `--output-path` flag SHALL specify which config file to validate. The legacy `--output` flag SHALL return a usage error.

#### Scenario: Validate uses default path

- **WHEN** user runs `germinator config validate` with no `--output-path` flag
- **THEN** system validates config at `~/.config/germinator/config.toml`

#### Scenario: Validate uses custom path

- **WHEN** user runs `germinator config validate --output-path /custom/path/config.toml`
- **THEN** system validates config at `/custom/path/config.toml`

#### Scenario: Legacy --output returns usage error

- **WHEN** `germinator config validate --output /tmp/config.toml` is invoked
- **THEN** the command SHALL return a usage error (exit 2) because `--output` is no longer a recognized flag

### Requirement: Config command registration

The `germinator config` parent command SHALL be registered in the root command.

#### Scenario: Config subcommands are accessible

- **WHEN** user runs `germinator config --help`
- **THEN** help output shows `init` and `validate` subcommands

### Requirement: config init follows command-options-pattern

The `config init` command SHALL adopt the `NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command` + `runConfigInit(opts *configInitOptions) error` template, consistent with the other migrated commands.

#### Scenario: config init supports runF injection

- **WHEN** `NewCmdConfigInit` is constructed with a non-nil `runF`
- **THEN** invoking the command SHALL call `runF(opts)` instead of the production `runConfigInit`
- **AND** the production path SHALL be exercisable end-to-end when `runF` is nil

### Requirement: config validate follows command-options-pattern

The `config validate` command SHALL adopt the `NewCmdConfigValidate(f *cmdutil.Factory, runF func(*configValidateOptions) error) *cobra.Command` + `runConfigValidate(opts *configValidateOptions) error` template.

#### Scenario: config validate supports runF injection

- **WHEN** `NewCmdConfigValidate` is constructed with a non-nil `runF`
- **THEN** invoking the command SHALL call `runF(opts)` instead of the production `runConfigValidate`

### Requirement: config validate renders errors at the boundary

When `config validate` encounters an error, the command SHALL return it; `main.go` SHALL render it exactly once via `output.FormatError`. The command SHALL NOT call `output.FormatError` itself (single-handling rule: errors are either rendered OR returned, never both).

#### Scenario: No double-rendering of validation errors

- **WHEN** `config validate` returns an error
- **THEN** `main.go` SHALL render it once via `output.FormatError` to `opts.IO.ErrOut`
- **AND** the command body SHALL NOT have already rendered the same error

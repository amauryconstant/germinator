# Capability: Config Commands

## Purpose

CLI commands for scaffolding and validating Germinator configuration files.

## Requirements

### Requirement: Config init scaffolds a new config file

The `germinator config init` command SHALL create a new config file with explanatory comments for each field.

#### Scenario: Init creates config at default location

- **WHEN** user runs `germinator config init` with no flags
- **AND** no config file exists at the default location
- **THEN** system creates the config file at `~/.config/germinator/config.toml`
- **AND** file contains comments explaining each field

#### Scenario: Init creates config at custom location

- **WHEN** user runs `germinator config init --output /custom/path/config.toml`
- **AND** no file exists at `/custom/path/config.toml`
- **THEN** system creates the config file at `/custom/path/config.toml`
- **AND** parent directories are created with permissions 0755

#### Scenario: Init refuses to overwrite without force

- **WHEN** user runs `germinator config init` with no `--force` flag
- **AND** a config file already exists at the target location
- **THEN** system returns an error indicating the file exists
- **AND** no file is written

#### Scenario: Init overwrites with force flag

- **WHEN** user runs `germinator config init --force`
- **AND** a config file already exists at the target location
- **THEN** system overwrites the existing file with the scaffolded content

### Requirement: Config init output contains all fields

The scaffolded config file SHALL contain entries for all configurable fields with explanatory comments.

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
- **THEN** file contains `[completion]` table with `timeout` and `cache_ttl` fields
- **AND** each field has a comment explaining its purpose and default value

### Requirement: Config validate checks existing config

The `germinator config validate` command SHALL validate an existing config file.

#### Scenario: Validate succeeds for valid config

- **WHEN** user runs `germinator config validate`
- **AND** a valid config file exists at the target location
- **THEN** system returns success message

#### Scenario: Validate fails when file not found

- **WHEN** user runs `germinator config validate`
- **AND** no config file exists at the target location
- **THEN** system returns an error indicating file not found

#### Scenario: Validate fails on malformed TOML

- **WHEN** user runs `germinator config validate`
- **AND** config file exists but contains invalid TOML syntax
- **THEN** system returns a parse error with file path

#### Scenario: Validate fails on invalid platform value

- **WHEN** user runs `germinator config validate`
- **AND** config file contains `platform = "unknown"`
- **THEN** system returns a validation error listing valid platforms

### Requirement: Config validate uses specified output path

The `--output` flag SHALL specify which config file to validate.

#### Scenario: Validate uses default path

- **WHEN** user runs `germinator config validate` with no `--output` flag
- **THEN** system validates config at `~/.config/germinator/config.toml`

#### Scenario: Validate uses custom path

- **WHEN** user runs `germinator config validate --output /custom/path/config.toml`
- **THEN** system validates config at `/custom/path/config.toml`

### Requirement: Config command registration

The `germinator config` parent command SHALL be registered in the root command.

#### Scenario: Config subcommands are accessible

- **WHEN** user runs `germinator config --help`
- **THEN** help output shows `init` and `validate` subcommands

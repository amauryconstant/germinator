## ADDED Requirements

### Requirement: Config file location discovery

The system SHALL discover the config file location using XDG Base Directory specification with fallbacks.

#### Scenario: XDG_CONFIG_HOME is set
- **WHEN** environment variable XDG_CONFIG_HOME is set to `/custom/config`
- **THEN** system looks for config at `/custom/config/germinator/config.toml`

#### Scenario: XDG_CONFIG_HOME not set
- **WHEN** environment variable XDG_CONFIG_HOME is not set
- **AND** HOME is `/home/user`
- **THEN** system looks for config at `/home/user/.config/germinator/config.toml`

#### Scenario: Config file in current directory
- **WHEN** no config file exists in XDG locations
- **AND** file `./config.toml` exists in current working directory
- **THEN** system loads config from current directory

#### Scenario: No config file exists
- **WHEN** no config file exists in any location
- **THEN** system uses default values silently (no error)

### Requirement: Config file parsing

The system SHALL parse TOML config files into a structured Config object.

#### Scenario: Parse valid config file
- **WHEN** config file contains valid TOML:
  ```toml
  library = "/path/to/library"
  platform = "opencode"
  ```
- **THEN** system returns Config with Library="/path/to/library" and Platform="opencode"

#### Scenario: Parse partial config file
- **WHEN** config file contains only some fields:
  ```toml
  library = "/path/to/library"
  ```
- **THEN** system returns Config with Library="/path/to/library" and Platform uses default (empty)

#### Scenario: Parse invalid TOML syntax
- **WHEN** config file contains invalid TOML syntax
- **THEN** system returns parse error with file path and line number

### Requirement: Default values

The system SHALL provide default values for all config fields.

#### Scenario: Library default
- **WHEN** library field is not specified in config
- **THEN** system uses default `~/.config/germinator/library`

#### Scenario: Platform default
- **WHEN** platform field is not specified in config
- **THEN** system uses empty string (requires explicit specification)

### Requirement: Config validation

The system SHALL validate config values on load and return errors for invalid values.

#### Scenario: Invalid platform value
- **WHEN** config file contains `platform = "unknown"`
- **THEN** system returns validation error listing valid platforms (opencode, claude-code)

#### Scenario: Valid platform values
- **WHEN** config file contains `platform = "opencode"` or `platform = "claude-code"`
- **THEN** system accepts the value

#### Scenario: Empty platform is valid
- **WHEN** platform is empty or not specified
- **THEN** system accepts the value (no validation error)

### Requirement: Manager interface

The system SHALL provide a ConfigManager interface for loading and accessing config.

#### Scenario: Load config
- **WHEN** Load() is called
- **THEN** system discovers config file, parses it, validates it, and returns Config object

#### Scenario: Get config after load
- **WHEN** GetConfig() is called after successful Load()
- **THEN** system returns the loaded Config object

#### Scenario: Get config before load
- **WHEN** GetConfig() is called before Load()
- **THEN** system returns nil or panics (undefined behavior)

### Requirement: Path expansion

The system SHALL expand tilde (~) in library path to user's home directory.

#### Scenario: Tilde expansion in library path
- **WHEN** config file contains `library = "~/custom/library"`
- **THEN** system expands to `/home/user/custom/library` (or appropriate home directory)

#### Scenario: Absolute path unchanged
- **WHEN** config file contains `library = "/absolute/path/library"`
- **THEN** system uses path as-is

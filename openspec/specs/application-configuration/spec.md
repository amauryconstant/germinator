# Capability: Configuration Management

## Purpose

The Configuration Management capability provides centralized configuration loading and management for the Germinator CLI. It handles config file discovery, parsing, validation, and access through a manager interface.

## Requirements

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
- **THEN** system uses default `~/.local/share/germinator/library/`
- **AND** when `XDG_DATA_HOME` is set the default SHALL be `$XDG_DATA_HOME/germinator/library/`

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

- **WHEN** `GetConfig()` is called before any explicit `Load()`
- **THEN** the manager SHALL return the default config seeded by `NewConfigManager()` (every field takes its documented default value)
- **AND** no panic or nil dereference SHALL occur (default-seeded access is the canonical "fresh state")
### Requirement: Path expansion

The system SHALL expand tilde (~) in library path to user's home directory.

#### Scenario: Tilde expansion in library path

- **WHEN** config file contains `library = "~/custom/library"`
- **THEN** system expands to `/home/user/custom/library` (or appropriate home directory)

#### Scenario: Absolute path unchanged

- **WHEN** config file contains `library = "/absolute/path/library"`
- **THEN** system uses path as-is

### Requirement: Configuration precedence

The system SHALL apply configuration sources in the following order, with later sources overriding earlier ones (last write wins):

1. **Defaults** — hardcoded in `DefaultConfig()`
2. **Config file** — `$XDG_CONFIG_HOME/germinator/config.toml` or `~/.config/germinator/config.toml`
3. **Environment variables** — `GERMINATOR_*` prefix (e.g., `GERMINATOR_LIBRARY`, `GERMINATOR_PLATFORM`, `GERMINATOR_DEBUG`)
4. **Flags** — explicit user intent for this invocation (Cobra flags override everything)

Library path resolution follows a parallel three-tier chain within the config layer: `--library` flag > `GERMINATOR_LIBRARY` env > config file > XDG default.

#### Scenario: Flag overrides env var

- **GIVEN** the environment variable `GERMINATOR_LIBRARY=/env/lib`
- **AND** the config file sets `library = "/file/lib"`
- **WHEN** the user runs `germinator --library /flag/lib library resources`
- **THEN** the library path SHALL be `/flag/lib`

#### Scenario: Env var overrides config file

- **GIVEN** the environment variable `GERMINATOR_PLATFORM=opencode`
- **AND** the config file sets `platform = "claude-code"`
- **WHEN** `germinator --help` is run (any command without `--platform` flag)
- **THEN** the resolved platform SHALL be `opencode` (env wins)

#### Scenario: Config file overrides defaults

- **GIVEN** no environment variables are set
- **AND** the config file sets `library = "/custom/lib"`
- **WHEN** `germinator library resources` is run
- **THEN** the library path SHALL be `/custom/lib`

### Requirement: XDG resolution via adrg/xdg

XDG path resolution SHALL use `github.com/adrg/xdg` (cross-platform: handles `XDG_CONFIG_HOME`/`XDG_DATA_HOME`/`XDG_CACHE_HOME` on Unix, and the Windows equivalents — `%AppData%`/`%LocalAppData%` — transparently). Resolution rules:

- Config: `$XDG_CONFIG_HOME/germinator/config.toml` → fallback `~/.config/germinator/config.toml`
- Library (data): `$XDG_DATA_HOME/germinator/library/` → fallback `~/.local/share/germinator/library/`

#### Scenario: adrg/xdg handles missing XDG_DATA_HOME

- **GIVEN** neither `XDG_DATA_HOME` nor `HOME` is set to a writable location on a Unix system
- **WHEN** `library.DefaultLibraryPath()` is called
- **THEN** the function SHALL return the platform-appropriate fallback (e.g., `~/.local/share/germinator/library/` on Unix, `%LocalAppData%\germinator\library\` on Windows) via `adrg/xdg.DataFile("germinator/library")` or equivalent

### Requirement: Environment variable naming

All environment variables SHALL use the `GERMINATOR_` prefix, underscore word separation, and uppercase letters (e.g., `GERMINATOR_LIBRARY`, `GERMINATOR_PLATFORM`, `GERMINATOR_DEBUG`).

#### Scenario: GERMINATOR_DEBUG enables debug logging

- **GIVEN** the environment variable `GERMINATOR_DEBUG=1` is set
- **WHEN** `germinator library resources` is run
- **THEN** `IOStreams.Logger` SHALL be a debug-level structured handler writing to `ErrOut`
- **AND** debug log lines SHALL appear on `ErrOut` interleaved with normal verbose output

#### Scenario: GERMINATOR_LIBRARY overrides config

- **GIVEN** the config file sets `library = "/file/lib"`
- **AND** the environment variable `GERMINATOR_LIBRARY=/env/lib` is set
- **WHEN** `germinator library resources` is run
- **THEN** the library SHALL be loaded from `/env/lib`

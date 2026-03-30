## Why

Germinator's configuration system uses sensible defaults but provides no guidance on what settings are available or how to customize them. Users who want to configure germinator must either read source code or manually create a config file without understanding the available options. There is no command to scaffold a well-documented config file.

## What Changes

- New `germinator config` parent command with subcommands:
  - `germinator config init` - Scaffold a config file with documented fields explaining each setting
  - `germinator config validate` - Validate an existing config file
- `config init` flags:
  - `--output <path>` - Output file path (default: XDG config path `~/.config/germinator/config.toml`)
  - `--force` - Overwrite existing file (default: error if file exists)
- `config validate` flags:
  - `--output <path>` - Config file to validate (default: XDG config path)
- `config validate` behavior:
  - Returns error if file does not exist
  - Returns error if TOML parsing fails
  - Returns error if validation fails (e.g., unknown platform value)
  - Returns success if config is valid

## Capabilities

### New Capabilities

- `config-commands`: CLI commands for config file scaffolding and validation
  - `germinator config init` scaffolds a new config file with explanatory comments
  - `germinator config validate` checks if a config file is parseable and valid

### Modified Capabilities

(None)

## Impact

- New file: `cmd/config.go` - Config command group with init and validate subcommands
- No changes to existing commands, services, or domain models
- Reuses existing `internal/infrastructure/config` package (Manager, GetConfigPath, etc.)
- No breaking changes

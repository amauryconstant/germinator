## Why

Germinator lacks a global configuration system. All settings must be passed via flags on every invocation. As features like the library system are added, users need a way to persist preferences (library path, default platform) rather than specifying them repeatedly.

## What Changes

- Add `internal/config/` package with Koanf-based configuration loading
- TOML config file at XDG-compliant location (`$XDG_CONFIG_HOME/germinator/config.toml`)
- Manager pattern for config loading and access
- Validation on load with clear error messages
- Config file is optional - missing file uses defaults silently

## Capabilities

### New Capabilities

- `configuration`: Load, validate, and access user configuration from TOML file with XDG-compliant location discovery

### Modified Capabilities

(none)

## Impact

**New code:**
- `internal/config/config.go` - Config struct, defaults, validation
- `internal/config/manager.go` - ConfigManager interface + Koanf implementation

**Consumed by:**
- `internal/library/` (library-system change) - for library path discovery
- Future features needing user preferences

**No changes to:**
- Existing transformation pipeline
- Existing CLI commands (yet)

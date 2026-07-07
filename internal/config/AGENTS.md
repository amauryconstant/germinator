**Location**: `internal/config/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview

---

# Config Package

Loads Germinator's TOML configuration from disk using koanf, with XDG path resolution and sensible defaults. Holds three concerns: the `Config` value type, the `Manager` interface + koanf implementation, and standalone path helpers.

## Files

| File | Purpose |
|------|---------|
| `config.go` | `Config`, `CompletionConfig`, `DefaultConfig`, `Validate`, `ExpandPaths`, `expandTilde` |
| `manager.go` | `Manager` interface, `koanfConfigManager`, `NewConfigManager`, `Load`, `resolveConfigPath`, `GetConfigPath` |
| `config_test.go` | `Config` / `Validate` / `ExpandPaths` tests |
| `manager_test.go` | `Load` + path resolution tests |

## Public Surface

### `Config` value type (`config.go`)

- `Config` — top-level config value. Fields: `Library string` (path to library dir), `Platform string` (default platform; empty = require flag), `Completion CompletionConfig`.
- `CompletionConfig` — `Timeout string` (default `"500ms"`), `CacheTTL string` (default `"5s"`). Both are duration strings parsed elsewhere.
- `DefaultConfig() *Config` — returns the baseline config used when no file is found.
- `(*Config).Validate() error` — no-op when `Platform == ""`; otherwise checks against `core.PlatformClaudeCode` / `core.PlatformOpenCode` and returns `*core.ConfigError` (with suggestions) on mismatch.
- `(*Config).ExpandPaths() error` — expands `~/` in `Library` via `os.UserHomeDir`.

### `Manager` interface (`manager.go`)

- `Manager` — `Load() error` + `GetConfig() *Config`.
- `NewConfigManager() Manager` — returns the koanf-backed implementation seeded with `DefaultConfig()`.
- `(*koanfConfigManager).Load()` — resolves the config path (XDG → `~/.config` → `./config.toml`), loads via `koanf` + `toml/v2` parser, unmarshals into `*Config`, runs `Validate`, then `ExpandPaths`. **A missing config file is not an error** — defaults are kept.
- `GetConfigPath() (string, error)` — returns the XDG-preferred config path even if the file does not exist (used by `config init` / messages).

### Path resolution order (`resolveConfigPath`)

1. `$XDG_CONFIG_HOME/germinator/config.toml`
2. `$HOME/.config/germinator/config.toml`
3. `./config.toml` (current working directory)

First existing file wins; none found → returns `""` and the caller keeps defaults.

## Design Notes

**Why a separate `Manager` interface.** Decouples loading mechanics (koanf, file I/O) from the `Config` value type so tests can substitute a fake manager. `cmdutil.Factory` exposes config lazily via a `Config func() (*config.Config, error)` field rather than embedding a `*Manager`.

**Missing file ≠ error.** This is deliberate: a fresh install with no config file should silently use defaults. Parse failures and validation failures still return typed errors (`*core.FileError`, `*core.ParseError`, `*core.ConfigError`).

**Dependencies.** `internal/core` (for `PlatformClaudeCode`/`PlatformOpenCode` constants and typed errors — moved here from `internal/models` in slice 9.3), `knadh/koanf/v2` + `koanf/parsers/toml/v2` + `koanf/providers/file`. The package does not depend on `internal/iostreams` or `internal/cmdutil` — it is a leaf shell package.

**Location**: `internal/config/`
**Parent**: See `internal/AGENTS.md` for core patterns

---

# Configuration Package

Koanf-based configuration loading with XDG-compliant file discovery.

## Usage

```go
cm := config.NewConfigManager()
if err := cm.Load(); err != nil {
    // handle error (missing file is not an error)
}
cfg := cm.GetConfig()
```

## Config File

| Field | Default | Description |
|-------|---------|-------------|
| `library` | `~/.config/germinator/library` | Path to library directory |
| `platform` | `""` (empty) | Default platform (empty = must specify via flag) |
| `completion.timeout` | `500ms` | Max time for library loading during shell completion |
| `completion.cache_ttl` | `5s` | Cache duration for library data during completion |

**Valid platforms**: `opencode`, `claude-code`

**File location** (checked in order):
1. `$XDG_CONFIG_HOME/germinator/config.toml`
2. `$HOME/.config/germinator/config.toml`
3. `./config.toml` (cwd)

## Patterns

| Pattern | Detail |
| ------- | ------ |
| Manager interface | `ConfigManager` with `Load() error` and `GetConfig() *Config` |
| Optional config | Missing file uses defaults silently |
| Path expansion | `~` expanded to home directory after load |
| Validation | Platform validated against known values at load time |

## Example Config

```toml
# ~/.config/germinator/config.toml
library = "/path/to/library"
platform = "opencode"

[completion]
timeout = "500ms"    # Max time for library loading
cache_ttl = "5s"     # Cache duration for completion data
```

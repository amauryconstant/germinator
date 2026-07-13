**Location**: `internal/config/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/03-input-config.md`

---

# Config Package

Loads Germinator's TOML configuration from disk using koanf, with XDG path resolution and sensible defaults. Holds three concerns: the `Config` value type, the `Manager` interface + koanf implementation, and standalone path helpers.

## Files

| File | Purpose |
|------|---------|
| `config.go` | `Config`, `CompletionConfig`, `DefaultConfig`, `Validate`, `ExpandPaths` (delegates tilde expansion to `internal/paths`) |
| `manager.go` | `Manager` interface, `koanfConfigManager`, `NewConfigManager`, `Load`, `resolveConfigPath`, `GetConfigPath` |
| `load.go` | Top-level `Load()` wrapper with `loadFn` test-injection seam (mutex-protected) |
| `config_test.go` | `Config` / `Validate` / `ExpandPaths` tests |
| `manager_test.go` | `Load` + env-provider + path resolution tests |
| `manager_xdg_test.go` | `//go:build !windows` — XDG resolution tests |

## Public Surface

### `Config` value type (`config.go`)

| Field | Type | Default | Source |
|-------|------|---------|--------|
| `Library` | `string` | `""` (XDG falls through) | koanf `library`, env `GERMINATOR_LIBRARY` |
| `PlatformDefault` | `string` | `""` | koanf `platform`, env `GERMINATOR_PLATFORM` (lowercased) |
| `Debug` | `bool` | `false` | koanf `debug`, env `GERMINATOR_DEBUG` (strconv.ParseBool) |
| `Completion` | `CompletionConfig` | `Timeout:"500ms"`, `CacheTTL:"5s"` | koanf `completion.*` |

- `Library: ""` is the canonical "no config-file override" signal — `library.DefaultLibraryPath()` resolves the XDG path when `Library` is empty at resolution time.
- `DefaultConfig() *Config` — returns the baseline config used when no file is found. All fields are seeded; `Library` is intentionally empty.
- `(*Config).Validate() error` — returns `nil` for valid configs; otherwise returns `*core.ConfigError` (with suggestions) joined via `errors.Join` (collect-all semantics). Checks `PlatformDefault` against `core.PlatformClaudeCode` / `core.PlatformOpenCode` (empty allowed), and validates that `Completion.Timeout` / `Completion.CacheTTL` parse via `time.ParseDuration` (empty allowed).
- `(*Config).ExpandPaths() error` — expands `~/` in `Library` via `internal/paths.ExpandHome`.

### `Manager` interface (`manager.go`)

- `Manager` — `Load() error` + `GetConfig() *Config`.
- `NewConfigManager() Manager` — returns the koanf-backed implementation seeded with `DefaultConfig()`.
- `(*koanfConfigManager).Load()` — resolves the config path via `adrg/xdg`, loads via `koanf` + `toml/v2` parser, then **merges `GERMINATOR_*` env vars** via `koanf/providers/env` with `.` delimiter and lowercased keys. Merge order: defaults → file → env. Unmarshals into `*Config`, runs `Validate`, then `ExpandPaths`. **A missing config file is not an error** — defaults + env are kept.
- `GetConfigPath() (string, error)` — returns the XDG-preferred config path even if the file does not exist (used by `config init` / messages).

### Top-level `Load()` wrapper (`load.go`)

- `Load() (*Config, error)` — convenience entry point that callers (notably `cmdutil.BuildFactory`) use without instantiating a Manager. Contract: `*Config` is always non-nil (returns the `DefaultConfig()`-seeded struct on error too); the error chain (`*core.FileError` / `*core.ParseError` / `*core.ConfigError`, dispatched via `errors.As` by `output.FormatError`) is the authoritative signal.
- `swapLoadFn` (test-only, mutex-protected) — replaces the underlying `loadFn` for stub injection. Use with `t.Cleanup(swapLoadFn(...))`.

### Path resolution order (`resolveConfigPath`)

1. `xdg.ConfigFile("germinator/config.toml")` (resolved by `adrg/xdg` per `$XDG_CONFIG_HOME`)
2. `./config.toml` (current working directory, for projects that ship their own config)

Mutex-protected via `xdgReload()` so parallel tests that mutate XDG env vars via `t.Setenv` see updated base directories. The function is called via `xdg.ConfigFile(...)` which may create the directory on disk; production callers handle missing files gracefully.

### Env-var mapping (via koanf env provider)

| Koanf key | Env var |
|-----------|---------|
| `library` | `GERMINATOR_LIBRARY` |
| `platform` (→ `PlatformDefault`) | `GERMINATOR_PLATFORM` (NOT `_DEFAULT` — prefix stripped, remaining key lowercased) |
| `debug` | `GERMINATOR_DEBUG` |
| `completion.timeout` | `GERMINATOR_COMPLETION.TIMEOUT` |
| `completion.cache_ttl` | `GERMINATOR_COMPLETION.CACHE_TTL` |

Bool truthiness: koanf parses via `strconv.ParseBool` semantics — `1` / `t` / `T` / `true` / `TRUE` / `True` → `true`; all other non-empty strings → `false`; unset → struct default. See `TestConfig_EnvVarBoolTruthinessRule`.

## Design Notes

**Why a separate `Manager` interface.** Decouples loading mechanics (koanf, file I/O) from the `Config` value type so tests can substitute a fake manager. `cmdutil.Factory` exposes config lazily via a `Config func() (*config.Config, error)` field rather than embedding a `*Manager`.

**Missing file ≠ error.** This is deliberate: a fresh install with no config file should silently use defaults. Parse failures and validation failures still return typed errors (`*core.FileError`, `*core.ParseError`, `*core.ConfigError`).

**Dependencies.** `internal/core` (for `PlatformClaudeCode`/`PlatformOpenCode` constants and typed errors), `knadh/koanf/v2` + `koanf/parsers/toml/v2` + `koanf/providers/file`. The package does not depend on `internal/iostreams` or `internal/cmdutil` — it is a leaf shell package.

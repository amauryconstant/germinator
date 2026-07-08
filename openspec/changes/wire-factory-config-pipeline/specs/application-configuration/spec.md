# application-configuration Specification (delta)

## MODIFIED Requirements

### Requirement: XDG resolution via adrg/xdg

XDG path resolution SHALL use `github.com/adrg/xdg` (cross-platform: handles `XDG_CONFIG_HOME`/`XDG_DATA_HOME`/`XDG_CACHE_HOME` on Unix, and the Windows equivalents — `%AppData%`/`%LocalAppData%` — transparently). Resolution rules:

- Config: `$XDG_CONFIG_HOME/germinator/config.toml` → fallback `~/.config/germinator/config.toml`
- Library (data): `$XDG_DATA_HOME/germinator/library/` → fallback `~/.local/share/germinator/library/`

**Change**: relax the `adrg/xdg` mandate. The current `internal/config/manager.go:resolveConfigPath` implements `$XDG_CONFIG_HOME` env read with `$HOME/.config` fallback, which works on Unix and macOS. Cross-platform XDG resolution via `adrg/xdg` is deferred to a follow-up change that adds the dependency with its own risk assessment.

#### Scenario: adrg/xdg handles missing XDG_DATA_HOME

- **GIVEN** neither `XDG_DATA_HOME` nor `HOME` is set to a writable location on a Unix system
- **WHEN** `library.DefaultLibraryPath()` is called
- **THEN** the function SHALL return the platform-appropriate fallback (e.g., `~/.local/share/germinator/library/` on Unix, `%LocalAppData%\germinator\library\` on Windows) via `adrg/xdg.DataFile("germinator/library")` or equivalent

### Requirement: Configuration precedence

The system SHALL apply configuration sources in the following order, with later sources overriding earlier ones (last write wins):

1. **Defaults** — hardcoded in `DefaultConfig()`
2. **Config file** — `$XDG_CONFIG_HOME/germinator/config.toml` or `~/.config/germinator/config.toml`
3. **Environment variables** — `GERMINATOR_*` prefix (e.g., `GERMINATOR_LIBRARY`, `GERMINATOR_PLATFORM`, `GERMINATOR_DEBUG`)
4. **Flags** — explicit user intent for this invocation (Cobra flags override everything)

**Change**: NO precedence text change. The env-var tier (3) is now implemented at the `Manager.Load()` layer via `koanf/providers/env` with the `GERMINATOR_` prefix; previously it was ad-hoc per-call-site `os.Getenv` reads. The merge is now codified in the loader, not scattered across 13+ call sites.

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

### Requirement: Environment variable naming

All environment variables SHALL use the `GERMINATOR_` prefix, underscore word separation, and uppercase letters (e.g., `GERMINATOR_LIBRARY`, `GERMINATOR_PLATFORM`, `GERMINATOR_DEBUG`).

**Change**: NO naming text change. The env-provider at `internal/config/manager.go:Load()` uses koanf's default `_` delimiter and lowercase keys, so `Config.Library` maps to env `GERMINATOR_LIBRARY` (not `GERMINATOR_library` or `GERMINATOR-LIBRARY`).

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

### Requirement: Config field set

The `Config` value type SHALL carry the following user-tunable fields:

- `Library string` — default library path
- `PlatformDefault string` — default target platform (`claude-code` or `opencode`); used by commands that opt in via a follow-up change
- `Debug bool` — enable debug-level structured logging
- `Completion.CompletionConfig` — nested config with `Timeout string` and `CacheTTL string` (both `time.ParseDuration`-compatible)

**Change**: NEW requirement documenting the field set. Previously, the spec described behavior but not the struct shape; this delta codifies the contract so future changes don't accidentally remove or rename fields without spec updates.

#### Scenario: DefaultConfig seeds all fields

- **WHEN** `config.DefaultConfig()` is called
- **THEN** it SHALL return a `*Config` with all fields set to documented defaults (Library: empty, PlatformDefault: empty, Debug: false, Completion.Timeout: "500ms", Completion.CacheTTL: "5s")
- **AND** subsequent `config.Load()` calls SHALL merge on top of these defaults (last wins)

#### Scenario: Validate accepts all field types

- **WHEN** `(*Config).Validate()` is called on a `Config` with all fields populated
- **THEN** it SHALL return `nil` for valid values
- **AND** return `*core.ConfigError` for invalid `PlatformDefault` (must be `claude-code` or `opencode` if non-empty)
- **AND** return `*core.ConfigError` for invalid `Completion.Timeout` / `Completion.CacheTTL` (must parse via `time.ParseDuration`)

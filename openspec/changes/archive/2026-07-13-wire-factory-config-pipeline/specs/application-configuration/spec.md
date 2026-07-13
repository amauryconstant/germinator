# application-configuration Specification (delta)

## ADDED Requirements

### Requirement: Config field set

The `Config` value type SHALL carry the following user-tunable fields:

- `Library string` — config-file override for the library path. When empty (the `DefaultConfig()` seed), library-path resolution falls through to `DefaultLibraryPath()` (XDG-resolved via `adrg/xdg.DataFile("germinator/library")`, see requirement "XDG resolution via adrg/xdg" in the source-of-truth spec).
- `PlatformDefault string` — default target platform (`claude-code` or `opencode`); used by commands that opt in via a follow-up change
- `Debug bool` — enable debug-level structured logging
- `Completion CompletionConfig` — nested struct with `Timeout string` and `CacheTTL string` fields (both `time.ParseDuration`-compatible)

**Change**: NEW requirement documenting the field set. Previously, the spec described behavior but not the struct shape; this delta codifies the contract so future changes don't accidentally remove or rename fields without spec updates. The `Library` field documents its empty default explicitly so the `cfg.Library == ""` "no override" sentinel is part of the public contract.

#### Scenario: DefaultConfig seeds all fields

- **WHEN** `config.DefaultConfig()` is called
- **THEN** it SHALL return a `*Config` with all fields set to documented defaults (Library: `""` (empty string — XDG falls through at resolution time), PlatformDefault: `""`, Debug: `false`, Completion.Timeout: `"500ms"`, Completion.CacheTTL: `"5s"`)
- **AND** subsequent `config.Load()` calls SHALL merge on top of these defaults (last wins)

#### Scenario: Validate accepts all field types

- **WHEN** `(*Config).Validate()` is called on a `Config` with all fields populated
- **THEN** it SHALL return `nil` for valid values
- **AND** return `*core.ConfigError` for invalid `PlatformDefault` (must be `claude-code` or `opencode` if non-empty)
- **AND** return `*core.ConfigError` for invalid `Completion.Timeout` / `Completion.CacheTTL` (must parse via `time.ParseDuration`); empty values are valid and the completion helpers fall back to their defaults for nil cfg or empty strings
- **AND** when multiple fields are invalid, return all errors via `errors.Join` so users see every problem at once (collect-all semantics)
- **AND** `Library` SHALL be valid by `(*Config).Validate()` (always nil) for the empty string and for paths that do not start with `~/`; a non-empty `Library` starting with `~/` MAY surface a `*core.ConfigError` from the post-Validate `ExpandPaths()` step if `os.UserHomeDir()` cannot determine the user's home directory (this error path uses the `internal/paths.ExpandHome` canonical helper shared with `cmd/completions.go`)

#### Scenario: Tilde-prefixed Library surfaces ConfigError when HOME is unset

- **GIVEN** the config file sets `library = "~/custom/library"`
- **AND** `os.UserHomeDir()` cannot determine the user's home directory (e.g., `HOME` unset on a Unix-like system)
- **WHEN** `config.Load()` runs (via `Manager.Load()` → `(*Config).Validate()` → `(*Config).ExpandPaths()` → `paths.ExpandHome`)
- **THEN** it SHALL return `(*Config, *core.ConfigError)` with the message wrapping the underlying `os.UserHomeDir` failure

## MODIFIED Requirements

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
- **AND** `runResources` SHALL emit a debug log line (e.g., `"listing library resources"`) to `ErrOut`
- **AND** debug log lines SHALL appear on `ErrOut` interleaved with normal verbose output

#### Scenario: GERMINATOR_LIBRARY overrides config

- **GIVEN** the config file sets `library = "/file/lib"`
- **AND** the environment variable `GERMINATOR_LIBRARY=/env/lib` is set
- **WHEN** `germinator library resources` is run
- **THEN** the library SHALL be loaded from `/env/lib`

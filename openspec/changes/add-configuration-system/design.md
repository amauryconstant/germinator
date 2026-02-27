## Context

Germinator currently has no configuration system. All settings are passed via CLI flags on every invocation. As the library system and other features are added, users need a way to persist preferences rather than repeating flags.

**Constraints:**
- Minimal external dependencies (Koanf is acceptable)
- No environment variable overrides for config values (XDG env only affects file location)
- Config file is optional - missing file is not an error
- Follow existing Go patterns in the codebase

**Assumptions:**
- Single config file per user (no profiles)
- Config is loaded once at startup, not reloaded during runtime
- Users manage config file manually (no CLI commands to modify it)

## Goals / Non-Goals

**Goals:**
- Load config from XDG-compliant location
- Parse TOML config file
- Return structured Config object with defaults applied
- Validate config on load, fail with clear errors if invalid
- Provide Manager pattern for loading and access

**Non-Goals:**
- Environment variable overrides for config values
- Config modification commands (`config set`, `config get`)
- Multiple config profiles
- Hot-reloading of config changes
- Integration with Cobra flag binding

## Decisions

### D1: Config File Format

**Choice:** TOML

**Rationale:** Simpler than YAML for flat config structures. No significant whitespace issues. Matches twiggit pattern for consistency across projects.

**Alternatives:**
- YAML - rejected (whitespace sensitivity, over-engineering for simple config)
- JSON - rejected (no comments, less human-friendly)

### D2: Config Library

**Choice:** Koanf (`github.com/knadh/koanf/v2`)

**Rationale:** Lighter than Viper. Focused on config loading without Viper's complexity. Used successfully in twiggit. Clean separation of providers and parsers.

**Alternatives:**
- Viper - rejected (heavier, more features than needed)
- Hand-rolled - rejected (unnecessary complexity for TOML parsing)

### D3: File Location Strategy

**Choice:** XDG Base Directory specification with fallbacks

```
1. $XDG_CONFIG_HOME/germinator/config.toml
2. $HOME/.config/germinator/config.toml
3. ./config.toml (current working directory)
```

**Rationale:** Standard Linux convention. Predictable location. No custom env var for file location.

**Alternatives:**
- Custom env var (`GERMINATOR_CONFIG`) - rejected (unnecessary complexity)
- Single hardcoded path - rejected (inflexible)

### D4: Architecture Pattern

**Choice:** Manager pattern with interface

```go
type ConfigManager interface {
    Load() (*Config, error)
    GetConfig() *Config
}

type Config struct {
    Library  string
    Platform string
}
```

**Rationale:** Matches twiggit pattern. Allows mocking in tests. Clean separation of loading logic from config data.

**Alternatives:**
- Simple function `LoadConfig()` - rejected (less testable, no caching)
- Global singleton - rejected (harder to test, implicit state)

### D5: Package Location

**Choice:** `internal/config/`

```
internal/config/
├── config.go    # Config struct, defaults, validation
├── manager.go   # ConfigManager interface + Koanf implementation
└── *_test.go    # tests
```

**Rationale:** Follows existing `internal/` patterns. Config is internal implementation detail.

**Alternatives:**
- `pkg/config/` - rejected (not meant for external consumption)
- `internal/core/config.go` - rejected (separate domain, deserves own package)

### D6: Error Handling

**Choice:** Terse errors with context

**Examples:**
- `config file not found: /path/to/config.toml` (info only, not fatal)
- `invalid config: library path is required`
- `invalid config: unknown platform "foo" (valid: opencode, claude-code)`

**Rationale:** v1 simplicity. Clear enough for users to understand and fix.

**Alternatives:**
- Verbose errors with suggestions - rejected (over-engineering)

### D7: Default Values

**Choice:**
```go
Library:  "~/.config/germinator/library"
Platform: ""  // empty = must specify via flag
```

**Rationale:** Library has sensible default. Platform forces explicit choice (no hidden default behavior).

**Alternatives:**
- Default platform to "opencode" - rejected (implicit behavior, user should be explicit)

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Config file missing | Not an error - use defaults silently |
| Config file unreadable | Return error with path and system error |
| Invalid TOML syntax | Return parse error with line number |
| Invalid config values | Return validation error with field name |
| Library path doesn't exist | Not validated at config load time (deferred to library loading) |
| Platform value invalid | Validate against known platforms at load time |

## Open Questions

(none - design is complete for v1 scope)

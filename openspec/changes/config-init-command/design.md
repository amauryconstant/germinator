## Context

Germinator's configuration system loads settings from TOML files at XDG-compliant paths but provides no CLI to scaffold or validate these files. The existing `internal/infrastructure/config` package provides `Manager`, `Load()`, `GetConfigPath()`, and `DefaultConfig()`, but there is no user-facing command to create an initial config file or verify an existing one.

The config file has 4 fields: `library`, `platform`, `completion.timeout`, `completion.cache_ttl`. All fields have sensible defaults.

## Goals / Non-Goals

**Goals:**
- Provide `germinator config init` to scaffold a config file with explanatory comments
- Provide `germinator config validate` to verify a config file is parseable and valid
- Reuse existing `internal/infrastructure/config` infrastructure

**Non-Goals:**
- Creating or modifying library directory structure
- Changing the config file schema or adding new fields
- Interactive prompts (flags-only interface)

## Decisions

### 1: Command Structure

**Choice:** `germinator config` as parent command with `init` and `validate` subcommands.

**Rationale:** Groups related functionality under a common parent. Matches CLI patterns in the codebase. `config init` scaffolds a file; `config validate` checks it.

**Alternatives Considered:**
- Separate top-level commands (`germinator init-config`, `germinator validate-config`): Less coherent grouping
- `germinator setup`: Too generic, doesn't communicate "config"

### 2: Default Output Path

**Choice:** Use `config.GetConfigPath()` which returns `$XDG_CONFIG_HOME/germinator/config.toml` if set, otherwise `$HOME/.config/germinator/config.toml`.

**Rationale:** Matches the existing config resolution logic. Users running `config init` without flags get the same location the application will look for.

**Alternatives Considered:**
- CWD `config.toml`: Less useful for user-level config that persists across projects
- Always ask interactively: Over-engineered for a flag-based CLI

### 3: `--output` Flag Semantics

**Choice:** `--output` accepts an exact file path (not directory).

**Rationale:** Matches `init` command's `--output` which is also an exact output path. Explicit and predictable.

**Alternatives Considered:**
- `--output` as directory: Requires constructing filename, less flexible
- `--global`/`--local` flags: Adds complexity for marginal gain

### 4: TOML Comment Style

**Choice:** Comments **above** each field explaining the setting.

**Rationale:** More readable than trailing comments. Comments above is standard TOML documentation practice.

### 5: Error on Existing File Without `--force`

**Choice:** `config init` errors if the output file already exists, unless `--force` is set.

**Rationale:** Prevents accidental overwrites. `init` command uses same pattern with `force` flag.

## Scaffolded Config File Content

```toml
# Germinator configuration
# https://github.com/anomalyco/germinator
#
# This file configures germinator's global behavior.
# All settings are optional - defaults are used if omitted.

# Path to your library directory containing skills, agents, commands, and presets.
# The library must contain a library.yaml index file.
# Supports ~ expansion for home directory.
# Default: ~/.config/germinator/library
library = "~/.config/germinator/library"

# Default platform when --platform is not specified.
# Options: "opencode" (default), "claude-code"
# Leave empty to require --platform on every command.
# Default: "" (none)
platform = ""

# Shell completion configuration
[completion]

# Maximum time to wait for library loading during completion suggestions.
# Lower values = faster but may timeout on large libraries.
# Default: "500ms"
timeout = "500ms"

# How long to cache library data for completion performance.
# Higher values = faster completions but may show stale results.
# Default: "5s"
cache_ttl = "5s"
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| TOML comment syntax is valid | Comments use `#` which is valid TOML for top-level keys and table keys |
| Users confuse `germinator init` (resource install) with `germinator config init` | Different parent commands (`init` vs `config init`); help text distinguishes them |
| Config file location confusion | Success message prints the exact path created |

## Open Questions

(None)

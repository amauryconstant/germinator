**Location**: `cmd/commands/`
**Parent**: See `/cmd/AGENTS.md` for CLI architecture (DI, errors, exit codes, verbosity, lint enforcement)

---

# Per-Command Reference

User-facing flag tables, output formats, and behavior for each command group. This is reference material; the architectural patterns (DI, error handling, exit codes) live in `cmd/AGENTS.md`.

---

# Library Command

Manage the canonical resource library containing skills, agents, commands, and memory.

## Subcommands

| Command | Description |
|---------|-------------|
| `library init` | Scaffold a new library directory structure |
| `library add` | Import a resource to the library |
| `library refresh` | Sync metadata from resource files into library.yaml |
| `library remove` | Remove a resource or preset (sub-commands: `resource <ref>`, `preset <name>`) |
| `library validate` | Check library integrity (with optional `--fix`) |
| `library resources` | List all resources in library (grouped by type) |
| `library presets` | List all presets in library |
| `library create preset` | Create a new preset |
| `library show <ref>` | Display resource or preset details |

### Library Init

Scaffolds a new library directory with `library.yaml` and empty resource directories.

| Flag | Default | Description |
|------|---------|-------------|
| `--path` | `$XDG_DATA_HOME/germinator/library/` or `~/.local/share/germinator/library/` | Path to create library |
| `--dry-run` | false | Preview changes without creating files |
| `--force` | false | Overwrite existing library |

```bash
# Create at default path
germinator library init

# Custom location
germinator library init --path /tmp/my-library

# Preview
germinator library init --dry-run

# Overwrite existing
germinator library init --force
```

## Library Refresh

Syncs metadata from registered resource files into `library.yaml`. Updates description from frontmatter when stale, discovers renamed files by searching directories.

| Flag | Default | Description |
|------|---------|-------------|
| `--library` | XDG default | Path to library directory |
| `--dry-run` | false | Preview changes without modifying |
| `--force` | false | Skip resources with conflicts |
| `--output` | `plain` | `plain` (default, per-section `Refreshed/Unchanged/Skipped/Errors` report), `json`, or `table` |

```bash
# Sync metadata from files
germinator library refresh

# Preview what would change
germinator library refresh --dry-run

# Skip conflicts
germinator library refresh --force

# JSON output for scripting
germinator library refresh --output json

# Per-change table
germinator library refresh --output table
```

**What it does:**
- Updates `description` from frontmatter when stale
- Updates `path` when file renamed (only if frontmatter name matches entry key)
- Skips missing files silently
- Reports per-section results: `Refreshed`, `Unchanged`, `Skipped`, `Errors`
- Exit code 1 if any errors occurred (mapped by `cmdutil.ExitCodeFor`)

## Library Remove

Removes a resource (deletes file + YAML entry) or a preset (YAML entry only). One Cobra parent command with two sub-commands dispatched on positional args.

| Sub-command | Positional arg | Action |
|-------------|----------------|--------|
| `library remove resource <ref>` | `ref` (e.g., `skill/commit`) | Delete file + drop YAML entry; refuses if any preset references the resource |
| `library remove preset <name>` | `name` (e.g., `git-workflow`) | Drop YAML entry only; resources stay registered |

> The legacy positional `<ref>` argument is preserved (no `--type` / `--name` flag substitution).

**Flags** (inherited from parent): `--library` (XDG default), `--force` (no-op for preset), `--output plain|json|table`. See `cmd/library_remove.go` for full type/error mapping.

```bash
# Remove a skill (and its file)
germinator library remove resource skill/commit

# JSON for scripts
germinator library remove resource agent/reviewer --output json

# Remove a preset
germinator library remove preset git-workflow
```

## Library Validate

Checks library integrity against four issue types: `missing-file` / `ghost-resource` / `orphan` / `malformed-frontmatter`. Use `--fix` to auto-clean `library.yaml` (removes missing entries, strips ghost preset refs).

**Flags**: `--library` (XDG default), `--fix` (mutating; opt-in), `--output plain|json|table`. Without `--fix`, validate is read-only — `library.yaml` is never modified. Exit codes via `cmdutil.ExitCodeFor`: `0` clean/warnings-only, `1` load failure, `2` Cobra usage error. See `cmd/library_validate.go` for full output-shape details.

```bash
# Check integrity
germinator library validate

# JSON for CI scripts
germinator library validate --output json

# Auto-fix
germinator library validate --fix

# Machine-readable fix report
germinator library validate --fix --output json
```

## Library Path Discovery

Priority: `--library` flag > `GERMINATOR_LIBRARY` env > `$XDG_DATA_HOME/germinator/library/` or `~/.local/share/germinator/library/`

```bash
# Use default path
germinator library resources

# Use custom path via flag
germinator library resources --library /path/to/library

# Use custom path via environment
GERMINATOR_LIBRARY=/path/to/library germinator library resources
```

## Resource References

Format: `type/name` (e.g., `skill/commit`, `agent/reviewer`)

Valid types: `skill`, `agent`, `command`, `memory`

## Output Format

`library resources` outputs grouped sections:
```
Skills:
  skill/commit - Git commit best practices
  skill/merge-request - Generate merge request descriptions

Agents:
  agent/reviewer - Code review agent
```

`library presets` outputs preset details with resources:
```
git-workflow - Git workflow tools
  - skill/commit
  - skill/merge-request
```

## Library Create Preset

Create a new preset that references existing resources in the library.
The `library create` Cobra group wrapper was collapsed to a leaf in
slice 6 — `library create preset` is now a single command (no group
indirection), registered directly under `library` in `cmd/library.go`.

| Flag | Default | Description |
|------|---------|-------------|
| `--resources` | (required) | Comma-separated resource references (e.g., `skill/commit,agent/reviewer`) |
| `--description` | empty | Preset description |
| `--force` | false | Overwrite existing preset |
| `--library` | XDG default | Path to library directory |

**Output flag:** `library create preset` does NOT expose `--output`
(the legacy implementation did not have `--json`). This matches the
`output-formats` capability's "only commands that previously
supported `--json` get `--output`" rule.

Validation: Fails if referenced resources don't exist; fails if preset exists without `--force`. Empty `--resources` returns a usage error (exit 2).

```bash
# Create preset with single resource
germinator library create preset commit-tools --resources skill/commit

# Create preset with multiple resources
germinator library create preset git-workflow --resources skill/commit,skill/pr

# Create with description
germinator library create preset dev-setup --resources skill/build,agent/reviewer --description "Development setup"

# Overwrite existing preset
germinator library create preset git-workflow --resources skill/commit --force
```

**Output:** Displays preset name, description, and resources on success.

## Library Add

Import resources to the library. Supports three modes dispatched on
the `--discover` and `--batch` flags.

### Mode 1: explicit files

```bash
germinator library add <file> --type skill --name test
germinator library add <file> --type skill --name test --dry-run
germinator library add <file> --force
```

### Mode 2: `--discover` (report-only)

Scans `skills/`, `agents/`, `commands/`, `memory/` directories for
orphan files (not registered in `library.yaml`) and reports them.
Without `--batch --force`, this is report-only.

```bash
germinator library add --discover
germinator library add --discover --dry-run
```

`name_conflict` (orphan name already registered under a different
type) is recorded as a `*core.OperationError` per file and counts
toward `PartialSuccessError.Failed`. The exit code is 0 if any
succeeded and 1 otherwise.

### Mode 3: `--discover --batch --force` (continuous)

Continuously processes all orphans, registering them in
`library.yaml`. On cancellation (Ctrl-C), partial successes are
reported and the function returns wrapped `ctx.Err()`.

```bash
germinator library add --discover --batch --force
germinator library add --discover --batch --force --dry-run
```

### Output formats

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `plain` | `plain` (byte-identical to legacy format), `json` (via `output.NewJSONExporter`), or `table` (via `output.NewTableExporter`) |

```bash
# Plain (default)
germinator library add --discover --batch --force
# -> Added resource: skill/foo
# -> Orphaned resources:
# ->   skill/foo (/path/to/skills/foo.md)
# -> Summary: scanned=N, orphans=N, added=N, skipped=N, failed=N
# -> Added N, skipped N, failed N

# JSON
germinator library add --discover --batch --force --output json
# -> { "added": [...], "summary": { "totalScanned": N, ... } }

# Table
germinator library add --discover --output table
# -> TYPE   NAME   PATH
# -> skill  foo    /path/to/skills/foo.md
```

### Stream discipline

- **Stdout (`opts.IO.Out`):** primary data — added resources, summary lines, table rows, JSON output.
- **Stderr (`opts.IO.ErrOut`):** per-resource errors via `output.FormatError`; verbose progress via `opts.IO.Verbosef`; partial-success aggregates.

Never mix diagnostic output into stdout — this preserves
`germinator library add --discover --batch --force --output json | jq '.'`.

---

# Init Command

Install resources from the library to a target project directory.

## Required Flags

- `--platform` (required): Target platform (`opencode` or `claude-code`)
- `--resources` OR `--preset` (one required): Resources to install

## Optional Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--library` | XDG default | Path to library directory |
| `--output` | `.` | Output directory |
| `--dry-run` | false | Preview changes without writing |
| `--force` | false | Overwrite existing files |

## Mutually Exclusive

`--resources` and `--preset` are mutually exclusive.

## Output Path Derivation

| Type | OpenCode | Claude Code |
|------|----------|-------------|
| skill | `.opencode/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` |
| agent | `.opencode/agents/<name>.md` | `.claude/agents/<name>.md` |
| command | `.opencode/commands/<name>.md` | `.claude/commands/<name>.md` |
| memory | `.opencode/memory/<name>.md` | `.claude/memory/<name>.md` |

## Examples

```bash
# Install specific resources
germinator init --platform opencode --resources skill/commit,skill/merge-request

# Install from preset
germinator init --platform opencode --preset git-workflow

# Preview changes
germinator init --platform opencode --preset git-workflow --dry-run

# Overwrite existing files
germinator init --platform opencode --resources skill/commit --force

# Install to custom directory
germinator init --platform opencode --preset git-workflow --output /project
```

## Error Handling

**Partial Processing**: The init command processes all resources regardless of individual failures, collecting per-resource results.

| Scenario | Behavior |
|----------|----------|
| At least one resource succeeds | Returns success (exit 0), displays per-resource status and summary |
| All resources fail | Returns error (exit 1), individual errors visible in results |
| File exists without `--force` | Returns error for that resource, continues with others |
| Resource not found | Returns error for that resource, continues with others |

**Result reporting**: Each resource gets its own `InitializeResult` with individual error status. Use `--json` for structured output.

**Examples**:
```bash
# See all successes and failures in one run
germinator init --platform opencode --resources skill/commit,skill/invalid,skill/pr

# JSON output for scripting
germinator init --platform opencode --resources skill/commit,skill/invalid --json
```

---

# Config Command

Scaffold and validate germinator configuration files.

## Subcommands

| Command | Description |
|---------|-------------|
| `config init` | Scaffold a config file with documented fields |
| `config validate` | Validate an existing config file |

## Config Init

Scaffolds `~/.config/germinator/config.toml` (or custom path) with **all settings commented out** by default.

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | XDG config path | Output file path |
| `--force` | false | Overwrite existing file |

Settings are commented to allow selective override; users uncomment only what they need.

```bash
# Scaffold default config
germinator config init

# Custom output path
germinator config init --output /path/to/config.toml

# Overwrite existing
germinator config init --force
```

## Config Validate

Validates a config file is parseable and conformant.

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | XDG config path | Config file to validate |

```bash
# Validate default config
germinator config validate

# Validate custom path
germinator config validate --output /path/to/config.toml
```

**Returns:** Success (0), NotFound (6), Config error (3), Parse error (1)

---

# Completion Command

Generate shell completion scripts for 10+ shells via carapace.

## Supported Shells

| Shell | Subcommand |
|-------|------------|
| Bash | `completion bash` |
| Zsh | `completion zsh` |
| Fish | `completion fish` |
| PowerShell | `completion powershell` |
| Elvish | `completion elvish` |
| Nushell | `completion nushell` |
| Oil | `completion oil` |
| Tcsh | `completion tcsh` |
| Xonsh | `completion xonsh` |
| Clink (Windows) | `completion cmd-clink` |

## Installation Examples

```bash
# Bash
echo 'source <(germinator completion bash)' >> ~/.bashrc

# Zsh
germinator completion zsh > ~/.zfunc/_germinator

# Fish
germinator completion fish > ~/.config/fish/completions/germinator.fish
```

## Dynamic Completions

| Flag/Argument | Source |
|---------------|--------|
| `--resources` | Library resources (e.g., `skill/commit`) |
| `--preset` | Library presets |
| `library show <ref>` | Resources + presets |
| `--platform` | Static: `claude-code`, `opencode` |

## Completion Actions (completions.go)

| Function | Purpose |
|----------|---------|
| `actionPlatforms()` | Static platform values |
| `actionResources(cmd)` | Dynamic from library with caching |
| `actionPresets(cmd)` | Dynamic from library with caching |
| `actionLibraryRefs(cmd)` | Combined resources + presets |

## Caching

- Package-level cache with mutex for thread safety
- 5-second TTL (configurable via `completion.cache_ttl`)
- 500ms timeout for library loading (configurable via `completion.timeout`)
- Silent fallback to empty completions on error/timeout

## Library Path Resolution

Priority: `--library` flag > `GERMINATOR_LIBRARY` env > config > default

**Location**: `cmd/`
**Parent**: See `/AGENTS.md` for project overview

---

# CLI Entry Points

Cobra-based CLI with platform-specific validation, typed errors, and verbosity control.

## Files

| File | Purpose |
|------|---------|
| `main.go` | Composition root - wires ServiceContainer and executes CLI |
| `container.go` | ServiceContainer for dependency injection |
| `root.go` | Root command with subcommand registration |
| `adapt.go` | Transform document to target platform format |
| `validate.go` | Validate document against platform rules |
| `canonicalize.go` | Convert platform document to canonical format |
| `version.go` | Display version, commit, build date |
| `library.go` | Library management commands (init, resources, presets, show) |
| `library_init.go` | Library init subcommand (scaffolding new libraries) |
| `init.go` | Install resources from library to project |
| `completion.go` | Shell completion command (carapace-based, multi-shell) |
| `completions.go` | Dynamic completion actions with caching |
| `formatters.go` | Init command output formatting (dry-run, success) |
| `library_formatters.go` | Library command output formatting |
| `error_handler.go` | Error categorization and exit code handling |
| `error_formatter.go` | Typed error formatting with contextual hints |
| `verbose.go` | Verbosity levels and output helpers |
| `config.go` | Config command group (init, validate) and CommandConfig struct |

---

# Dependency Injection

## ServiceContainer

Services passed through command tree via `ServiceContainer`:
```go
type ServiceContainer struct {
    Transformer   application.Transformer
    Validator     application.Validator
    Canonicalizer application.Canonicalizer
    Initializer   application.Initializer
}

services := cmd.NewServiceContainer()
```

## Composition Root

`main.go` wires all dependencies:
```go
services := cmd.NewServiceContainer()
cfg := &cmd.CommandConfig{
    Services:       services,
    ErrorFormatter: cmd.NewErrorFormatter(),
    Verbosity:      0,
}
rootCmd := cmd.NewRootCommand(cfg)
```

## Calling Services

Commands access services through interfaces:
```go
result, err := cfg.Services.Transformer.Transform(ctx, &application.TransformRequest{
    InputPath:  args[0],
    OutputPath: args[1],
    Platform:   platform,
})
```

## Constructor Pattern

Commands use `NewXCommand(cfg *CommandConfig)` constructors with RunE pattern:
```go
func NewValidateCommand(cfg *CommandConfig) *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.RunE = func(c *cobra.Command, args []string) error {
        verbosity, _ := c.Flags().GetCount("verbose")
        cfg.Verbosity = Verbosity(verbosity)
        // Use cfg.Services, cfg.ErrorFormatter
        // Return errors (bubble up to main.go for centralized handling)
        return nil
    }
    return cmd
}
```

No `init()` functions or global command variables.

---

# CommandConfig

Holds configuration and services for command execution:
```go
type CommandConfig struct {
    Services       *ServiceContainer
    ErrorFormatter *ErrorFormatter
    Verbosity      Verbosity
}
```

---

# Required Flags

Both `adapt` and `validate` require `--platform` flag:
```go
_ = cmd.MarkFlagRequired("platform")
```

Validation uses typed ConfigError:
```go
if platform == "" {
    HandleError(cfg, gerrors.NewConfigError("platform", "",
        []string{models.PlatformClaudeCode, models.PlatformOpenCode},
        "--platform flag is required"))
}
```

## Supported Platforms

- `claude-code` - Claude Code document format
- `opencode` - OpenCode document format

Platform strings from `models.PlatformClaudeCode`, `models.PlatformOpenCode`.

---

# Verbosity Flag

Persistent `-v`/`-vv` flag on root command for all subcommands:
```go
rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (use -v or -vv)")
```

Levels:
- Level 0 (default): No verbose output
- Level 1 (`-v`): Basic progress info
- Level 2 (`-vv`): Detailed operation info

Usage:
```go
VerbosePrint(cfg, "Processing file: %s", filePath)      // Level 1+
VeryVerbosePrint(cfg, "Parsing YAML structure...")      // Level 2+
```

Output goes to stderr (stdout stays clean for piping).

---

# Exit Codes

Semantic exit codes for programmatic handling:
- `0` (Success) - Command completed successfully
- `1` (Error) - General errors (transform, file, unexpected)
- `2` (Usage) - Cobra argument/validation errors (invalid flags, missing args)
- `3` (Config) - Configuration/parsing errors (malformed YAML, config errors)
- `4` (Git) - Git-related errors
- `5` (Validation) - Document validation errors
- `6` (NotFound) - File/resource not found errors

Error categorization via `CategorizeError()` using `errors.As` for type detection.

---

# Error Handling

## Centralized Error Handling

Errors bubble up to main.go via RunE pattern:
```go
// main.go
cmd.SetGlobalCommandConfig(cfg)
if err := rootCmd.Execute(); err != nil {
    exitCode := cmd.HandleCLIError(rootCmd, err)
    os.Exit(int(exitCode))
}
```

```go
func HandleCLIError(c *cobra.Command, err error) ExitCode {
    // Formats and outputs error, returns exit code
    // Uses global CommandConfig set during command construction
}
```

## Error Formatter

Type-specific formatting with contextual hints:
- ParseError → "Parse error: <message> File: <path>"
- ValidationError → "Validation error: <message>" + "Hint:" lines
- TransformError → "Transform error (<operation> for <platform>): <message>"
- FileError → "File error (<operation>): <message> Path: <path>"
- ConfigError → "Config error: <message>" + "Available: <options>"

## Typed Errors

Import from `internal/errors`:
```go
import gerrors "gitlab.com/amoconst/germinator/internal/errors"

// Constructors
gerrors.NewParseError(path, message, cause)
gerrors.NewValidationError(message, field, suggestions)
gerrors.NewTransformError(operation, platform, message, cause)
gerrors.NewFileError(path, operation, message, cause)
gerrors.NewConfigError(field, value, available, message)
```

---

# Argument Count

`adapt` and `validate` use `cobra.ExactArgs(2)` and `cobra.ExactArgs(1)` respectively.
`root` and `version` use default (no arguments).

---

# Version Output Format

```go
fmt.Printf("germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)
```

Example: `germinator v0.3.20 (abc123def) 2026-02-04`

---

# Testing

`cmd_test.go` contains integration tests for CLI workflows.
Test both platforms when testing platform-specific commands.

New test files:
- `verbose_test.go` - Verbosity type and helper function tests
- `error_formatter_test.go` - Error formatting tests
- `library_test.go` - Library and init command tests
- `completions_test.go` - Completion action unit tests (cache, timeout, actions)

---

# Library Command

Manage the canonical resource library containing skills, agents, commands, and memory.

## Subcommands

| Command | Description |
|---------|-------------|
| `library init` | Scaffold a new library directory structure |
| `library resources` | List all resources in library (grouped by type) |
| `library presets` | List all presets in library |
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

- Fail-fast: Stops on first error
- File exists error without `--force`
- Resource not found error for missing resources
- Preset not found error for missing presets

---

# Config Command

Scaffold and validate germinator configuration files.

## Subcommands

| Command | Description |
|---------|-------------|
| `config init` | Scaffold a config file with documented fields |
| `config validate` | Validate an existing config file |

## Config Init

Scaffolds `~/.config/germinator/config.toml` (or custom path) with explanatory comments.

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | XDG config path | Output file path |
| `--force` | false | Overwrite existing file |

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

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
| `library.go` | Library management commands (resources, presets, show) |
| `init.go` | Install resources from library to project |

---

# Dependency Injection

## ServiceContainer

Services passed through command tree via `ServiceContainer`:
```go
type ServiceContainer struct {
    // Services added as application grows
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

## Constructor Pattern

Commands use `NewXCommand(cfg *CommandConfig)` constructors:
```go
func NewValidateCommand(cfg *CommandConfig) *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.Run = func(c *cobra.Command, args []string) {
        verbosity, _ := c.Flags().GetCount("verbose")
        cfg.Verbosity = Verbosity(verbosity)
        // Use cfg.Services, cfg.ErrorFormatter
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
- `2` (Usage) - Config/validation errors (invalid flags, missing args)
- `3` (Parse) - Parse errors (malformed YAML, unrecognized document type)

Error categorization via `CategorizeError()` using `errors.As` for type detection.

---

# Error Handling

## Central Error Handler

```go
func HandleError(cfg *CommandConfig, err error) {
    fmt.Fprintln(os.Stderr, cfg.ErrorFormatter.Format(err))
    os.Exit(int(GetExitCodeForError(err)))
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

---

# Library Command

Manage the canonical resource library containing skills, agents, commands, and memory.

## Subcommands

| Command | Description |
|---------|-------------|
| `library resources` | List all resources in library (grouped by type) |
| `library presets` | List all presets in library |
| `library show <ref>` | Display resource or preset details |

## Library Path Discovery

Priority: `--library` flag > `GERMINATOR_LIBRARY` env > `~/.config/germinator/library/`

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

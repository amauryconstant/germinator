**Location**: `cmd/`
**Parent**: See `/AGENTS.md` for project overview

---

# CLI Entry Points

Cobra-based CLI with platform-specific validation, typed errors, and verbosity control.

## Commands

- `root.go` - Entry point, runs help when no subcommand provided
- `adapt` - Transform document to target platform format
- `validate` - Validate document against platform rules
- `canonicalize` - Convert platform document to canonical format
- `version` - Display version, commit, build date

---

# Command Pattern

## CommandConfig Pattern

Commands use `CommandConfig` for dependency injection:
```go
func(cmd *cobra.Command, args []string) {
    cfg := NewCommandConfig(cmd)
    // Use cfg.ErrorFormatter, cfg.Verbosity
}
```

## Required Flags

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

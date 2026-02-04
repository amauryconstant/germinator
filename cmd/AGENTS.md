**Location**: `cmd/`
**Parent**: See `/AGENTS.md` for project overview

---

# CLI Entry Points

Cobra-based CLI with platform-specific validation.

## Commands

- `root.go` - Entry point, runs help when no subcommand provided
- `adapt` - Transform document to target platform format
- `validate` - Validate document against platform rules
- `version` - Display version, commit, build date

---

# Command Pattern

## Required Flags

Both `adapt` and `validate` require `--platform` flag:
```go
_ = cmd.MarkFlagRequired("platform")
```

Validation in Run function (not Cobra validation):
```go
if platform == "" {
    fmt.Fprintf(os.Stderr, "Error: --platform flag is required (available: %s, %s)\n", ...)
    os.Exit(1)
}
```

## Supported Platforms

- `claude-code` - Claude Code document format
- `opencode` - OpenCode document format

Platform strings from `models.PlatformClaudeCode`, `models.PlatformOpenCode`.

---

# Exit Codes

All commands use `os.Exit(1)` for errors:
- Missing required flag → exit 1
- Validation errors → exit 1
- Transformation errors → exit 1

Success: implicit exit 0 (no `os.Exit` call)

---

# Error Output

Use `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` for error messages.
Validation errors printed one per line to stderr.

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

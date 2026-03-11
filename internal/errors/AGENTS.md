**Location**: `internal/errors/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Errors Package

Typed domain errors with immutable builders for fluent construction.

---

# Error Types

| Type | Constructor | Use Case |
|------|-------------|----------|
| `ParseError` | `NewParseError(path, message, cause)` | Malformed YAML, unrecognized document type |
| `ValidationError` | `NewValidationError(request, field, value, message)` | Invalid field values |
| `TransformError` | `NewTransformError(operation, platform, message, cause)` | Template/render failures |
| `FileError` | `NewFileError(path, operation, message, cause)` | File read/write errors |
| `ConfigError` | `NewConfigError(field, value, message)` | Invalid configuration |

---

# Immutable Builder Pattern

All error types use immutable builders that return a **copy**:

```go
err := errors.NewParseError("agent.yaml", "invalid YAML", cause).
    WithSuggestions([]string{"Check indentation", "Validate syntax"}).
    WithContext("agent definition file")
```

| Method | Returns | Purpose |
|--------|---------|---------|
| `WithSuggestions([]string)` | New copy | Add actionable suggestions |
| `WithContext(string)` | New copy | Add context information |

**Important**: Builders return new instances; original error is unchanged.

---

# Getters

All types provide getters for programmatic access:

| Method | Returns |
|--------|---------|
| `Message()` | Error message string |
| `Cause()` | Underlying error (for chaining) |
| `Suggestions()` | Copy of suggestions slice |
| `Context()` | Additional context |

### Type-Specific Getters

| Type | Getters |
|------|---------|
| `ParseError` | `Path()` |
| `ValidationError` | `Field()`, `Value()`, `Request()` |
| `TransformError` | `Operation()`, `Platform()` |
| `FileError` | `Path()`, `Operation()`, `IsNotFound()` |
| `ConfigError` | `Field()`, `Value()` |

---

# Error Formatting

All types implement `fmt.Stringer` with consistent formatting:

```
<type> [<context>]: <message>: <cause>
💡 <suggestion 1>
💡 <suggestion 2>
```

### ParseError

```
parse error in agent.yaml: invalid YAML: yaml: line 5: could not find expected ':'
💡 Check indentation
💡 Validate syntax
```

### ValidationError

```
validation failed for Agent.temperature: must be between 0.0 and 1.0 (value: 2.5)
💡 Use a value in range [0.0, 1.0]
```

### TransformError

```
transform error (render for opencode): failed to execute template: template: agent:12: unexpected EOF
```

### FileError

```
file error (read /path/to/file.yaml): failed to read file: open /path/to/file.yaml: no such file or directory
```

### ConfigError

```
config error: invalid platform 'unknown': must be claude-code or opencode
💡 Use 'claude-code' for Claude Code
💡 Use 'opencode' for OpenCode
```

---

# Error Chaining

All types implement `Unwrap()` for `errors.Is` / `errors.As`:

```go
var parseErr *errors.ParseError
if errors.As(err, &parseErr) {
    fmt.Println(parseErr.Path())
}
```

---

# FileError.IsNotFound()

Helper method for file-not-found detection:

```go
if fileErr.IsNotFound() {
    // Handle missing file
}
```

Checks both message and cause for: "not found", "does not exist", "no such file".

---

# Usage Patterns

## Creating Errors

```go
// Basic
err := errors.NewParseError(path, "invalid frontmatter", nil)

// With cause
err := errors.NewParseError(path, "invalid YAML", yamlErr)

// Full builder chain
err := errors.NewValidationError("Agent", "mode", "invalid", "must be primary, subagent, or all").
    WithSuggestions([]string{"Use 'primary' for main agents", "Use 'subagent' for helpers"}).
    WithContext("agent behavior configuration")
```

## Handling Errors

```go
var parseErr *errors.ParseError
if errors.As(err, &parseErr) {
    log.Printf("Parse error in %s: %s", parseErr.Path(), parseErr.Message())
    for _, s := range parseErr.Suggestions() {
        log.Printf("  Suggestion: %s", s)
    }
}
```

---

# File Structure

| File | Content |
|------|---------|
| `types.go` | All error type definitions (557 lines) |
| `types_test.go` | Unit tests for all types |

---

# See Also

- `internal/validation/AGENTS.md` - How validation errors integrate with `Result[T]`
- `internal/core/AGENTS.md` - How core layer uses these errors

**Location**: `internal/paths/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview

---

# Paths Package

Shared filesystem path helpers used by both `internal/config` and `cmd/`. Centralizes tilde expansion so the two call sites (which previously had independent implementations) stay in sync.

## Files

| File | Purpose |
|------|---------|
| `expand.go` | `ExpandHome(path string) (string, error)` — canonical tilde expansion |
| `expand_test.go` | Table-driven coverage of `ExpandHome` |

## Public Surface

### `ExpandHome(path string) (string, error)`

Expands a leading `~` or `~/` in `path` to the current user's home directory.

| Input | Output | Error? |
|-------|--------|--------|
| `""` (empty) | `""` | no |
| `"~"` | home dir | no |
| `"~/foo"` | `<home>/foo` | no |
| `"/abs/path"` | `"/abs/path"` (unchanged) | no |
| `"rel/path"` | `"rel/path"` (unchanged) | no |
| `"/path/with~tilde"` | unchanged (only LEADING `~` is expanded) | no |
| `"~/foo"` with `HOME` unset | `""` | yes (wrapped via `errors.Join`) |

**Error contract:** An error is returned only when the input starts with `~` AND `os.UserHomeDir()` fails. The error chain is built with `errors.Join` so callers can inspect the underlying failure via `errors.Is`.

**Silent fallback:** Callers wanting the legacy silent-fallback behavior (used by `cmd/completions.go`) should catch the error and return the original path unchanged:

```go
expanded, err := paths.ExpandHome(path)
if err != nil {
    return path
}
return expanded
```

## Dependencies

- `errors`, `fmt`, `os`, `path/filepath` from stdlib only.
- No third-party deps. No `internal/core` or shell-package deps (this is a leaf shell package).

**Location**: `internal/iostreams/`
**Parent**: See `/internal/AGENTS.md` for package overview

---

# IOStreams Package

Centralized terminal I/O boundary used by all commands. Constructed by `cmdutil.Factory` in slice-2+; not yet consumed by `cmd/` in slice 1.

## Files

| File | Purpose |
|------|---------|
| `iostreams.go` | `IOStreams` struct, `System()` / `Test()` constructors, TTY detection, `Verbosef`, debug logger |
| `styles.go` | `Styles` (Error/Success/Warning/Dim/Bold) backed by `lipgloss`; respects `NO_COLOR` |
| `iostreams_test.go` | TTY detection, `Test()` buffers, `Verbosef`, `NO_COLOR` cases |
| `styles_test.go` | TTY vs non-TTY output for each `Styles` method |

## Key Surface

- `System()` — real I/O, TTY detected from fd 0/1, debug logger gated on `GERMINATOR_DEBUG`
- `Test()` — buffer-backed (`*bytes.Buffer`); `Out`/`ErrOut` assertable for test inspection
- `IsStdoutTTY()` / `IsInteractive()` — TTY predicates
- `Verbosef(format, args...)` — writes to `ErrOut` when `Verbose == true`; adds trailing newline
- `Styles.Error/Success/Warning/Dim/Bold(s)` — no-op when `!isTTY` or `NO_COLOR` is set

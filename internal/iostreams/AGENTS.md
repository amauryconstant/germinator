**Location**: `internal/iostreams/`
**Parent**: See `/internal/AGENTS.md` for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/09-logging.md`

---

# IOStreams Package

Centralized terminal I/O boundary used by all commands. Constructed by `cmdutil.Factory` (`System()` in production via `main.go`, `Test()` in tests) and threaded through every command's options struct as `IO *iostreams.IOStreams`.

## Files

| File | Purpose |
|------|---------|
| `iostreams.go` | `IOStreams` struct, `System()` / `Test()` constructors, TTY detection, `Verbosef`, debug logger |
| `styles.go` | `Styles` (Error/Success/Warning/Dim/Bold) backed by `lipgloss`; respects `NO_COLOR` |
| `iostreams_test.go` | TTY detection, `Test()` buffers, `Verbosef`, `NO_COLOR` cases |
| `styles_test.go` | TTY vs non-TTY output for each `Styles` method |

## Key Surface

- `System()` — real I/O, TTY detected from fd 0/1/2, debug logger gated on `GERMINATOR_DEBUG`
- `Test()` — buffer-backed (`*bytes.Buffer`); `Out`/`ErrOut` assertable for test inspection; TTY overrides preset to `false`
- `IsStdoutTTY()` / `IsStdinTTY()` / `IsStderrTTY()` — TTY predicates (stdout/stderr overridable for tests)
- `IsInteractive()` — true only when both stdin AND stdout are TTYs
- `SetStdoutTTY(v)` / `SetStderrTTY(v)` — override TTY detection (tests, exit-code canary)
- `Verbosef(format, args...)` — writes to `ErrOut` when `Verbose == true`; adds trailing newline
- `Warnf(format, args...)` — yellow-styled `Warning:` prefix to `ErrOut`, independent of `Verbose`
- `Styles.Error/Success/Warning/Dim/Bold(s)` — no-op when `!isTTY` or `NO_COLOR` is set

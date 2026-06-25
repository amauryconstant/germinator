**Location**: `internal/warning/`
**Parent**: See `/internal/AGENTS.md` for package overview

---

# Warning Package

User-facing warning emission that is independent of verbosity and the debug Logger. Currently hosts the exit-code deprecation canary (slice 2).

## Files

| File | Purpose |
|------|---------|
| `canary.go` | `MaybeWarnLegacyExitCode(io)` — one-time stderr warning gated on `EXIT_CODE_LEGACY` env var OR `IsStderrTTY()`; `ResetCanaryForTest()` for unit-test isolation |

## Key Surface

- `MaybeWarnLegacyExitCode(io *iostreams.IOStreams)` — emits once per process via `sync.Once`. Nil-receiver safe. Suppressed in non-TTY, non-env-var environments (typical CI).
- `ResetCanaryForTest()` — resets the once-state. Pair with `t.Cleanup(warning.ResetCanaryForTest)` in sub-tests.

## Why a dedicated package

Per the `golang-cli-architecture` skill: **pure logic stays in core/cmdutil, side effects live in imperative shell packages**. `cmdutil.ExitCodeFor` remains a pure function with no logger or io parameter; this package owns the canary emission and can be invoked from `main.go` only.

## Caller contract

Call sites MUST invoke `MaybeWarnLegacyExitCode` only when the resolved exit code is `1` (the "general error" code, mapped from typed errors). Exit code `2` (Cobra/pflag usage errors) MUST NOT trigger the canary.

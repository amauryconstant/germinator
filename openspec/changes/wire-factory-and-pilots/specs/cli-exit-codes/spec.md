# cli-exit-codes Specification (delta)

## ADDED Requirements

### Requirement: Exit code deprecation canary

The `germinator` post-`Execute` error path SHALL emit a one-time deprecation warning to stderr when the resolved exit code is `1` AND at least one of the following conditions holds:

- The `EXIT_CODE_LEGACY` environment variable is set to a non-empty value, OR
- stderr is a TTY (per `IOStreams.IsStderrTTY()`).

The warning SHALL be emitted at most once per process via `sync.Once` and SHALL be suppressed in non-TTY, non-`EXIT_CODE_LEGACY` environments (typical CI). The helper is exposed as `internal/warning.MaybeWarnLegacyExitCode(io *iostreams.IOStreams)` and is invoked from `main.go` immediately before `os.Exit(int(cmdutil.ExitCodeFor(err)))`. The warning is written to `io.ErrOut` via `io.Warnf(...)` (a method on `IOStreams` that wraps `Styles.Warning`); it does NOT depend on `IOStreams.Logger` (which is gated on `GERMINATOR_DEBUG`). `cmdutil.ExitCodeFor` remains a pure function with no side effects and no logger parameter.

#### Scenario: Canary fires in interactive session

- **WHEN** the process exits with code `1` AND (stderr is a TTY OR `EXIT_CODE_LEGACY` is set to a non-empty value)
- **THEN** the deprecation warning SHALL be written to stderr exactly once

#### Scenario: Canary suppressed in non-interactive, non-env-var invocation

- **WHEN** the process exits with code `1` AND stderr is not a TTY AND `EXIT_CODE_LEGACY` is unset
- **THEN** no deprecation warning SHALL be emitted

#### Scenario: Canary fires when explicitly requested

- **WHEN** `EXIT_CODE_LEGACY=1` is set AND the process exits with code `1`
- **THEN** the deprecation warning SHALL be written to stderr exactly once

#### Scenario: Single emission per process

- **GIVEN** `MaybeWarnLegacyExitCode` has already been called once during the current process
- **WHEN** any subsequent exit code is `1`
- **THEN** the deprecation warning SHALL NOT be emitted again

#### Scenario: ResetCanaryForTest resets once-state

- **GIVEN** `MaybeWarnLegacyExitCode` has been called once during the current process
- **WHEN** `ResetCanaryForTest()` is invoked
- **THEN** the next call to `MaybeWarnLegacyExitCode` SHALL be permitted to emit the warning (subject to gate conditions)

#### Scenario: Warning emission is independent of Logger

- **WHEN** `MaybeWarnLegacyExitCode` is called with an `IOStreams` whose `Logger` field is nil
- **THEN** the function SHALL still write the warning to `io.ErrOut` (the canary does not depend on the Logger)

#### Scenario: Exit code 2 does not trigger the canary

- **WHEN** the process exits with code `2` (`ExitCodeUsage`) under any TTY or env-var conditions
- **THEN** no deprecation warning SHALL be emitted (the canary is gated on exit code `1` only)

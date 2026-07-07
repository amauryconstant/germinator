# cli-self-update Specification

## Purpose

Define opt-in update notification: on first run, ask the user whether to enable update checks. If enabled, check once per day (TTL-cached), display a non-blocking notice on stderr, and never auto-update without explicit consent.

> **Library:** `github.com/creativeprojects/go-selfupdate` for GitHub-based self-update detection.

## Requirements

### Requirement: First-run consent prompt

The CLI SHALL prompt the user on first run to opt into update checking. The first-run state is detected by the absence of a `last_check` timestamp in the local state file.

#### Scenario: First-run prompt on a TTY

- **GIVEN** the user has never run `germinator` before
- **AND** stdin AND stdout are TTYs
- **WHEN** any `germinator` command is invoked
- **THEN** the CLI SHALL print a single consent prompt to `ErrOut`: `Enable update checking? [y/N]`
- **AND** on `y`, persist a `consent = true` flag in the local state file
- **AND** on `N` or empty, persist a `consent = false` flag (no future prompts)

#### Scenario: First-run prompt suppressed in non-interactive contexts

- **GIVEN** the user has never run `germinator` before
- **AND** stdin is not a TTY (CI, pipe, redirect)
- **WHEN** any `germinator` command is invoked
- **THEN** the CLI SHALL default `consent = false` (no prompt, no update checks)
- **AND** no error SHALL be returned for missing consent

### Requirement: Daily check TTL

The CLI SHALL perform an update check at most once per 24-hour window. The last-check timestamp is persisted in the local state file at `$XDG_DATA_HOME/germinator/state/last_check`.

#### Scenario: Check within TTL skips the network call

- **GIVEN** `consent = true` was set 12 hours ago
- **WHEN** any `germinator` command is invoked
- **THEN** no update check SHALL be performed (within 24-hour TTL)
- **AND** the command SHALL proceed normally

#### Scenario: Check past TTL fetches the latest release

- **GIVEN** `consent = true` was set 25 hours ago
- **WHEN** any `germinator` command is invoked
- **THEN** a non-blocking goroutine SHALL fetch the latest release from the configured GitHub repo
- **AND** `last_check` SHALL be updated to the current time on success

### Requirement: Non-blocking update notice

When an update is available, the CLI SHALL print a notice to `ErrOut` (not `Out`), without blocking the command's primary output.

#### Scenario: New version available

- **GIVEN** `consent = true`
- **AND** the latest GitHub release version is greater than the current binary's version
- **WHEN** any `germinator` command is invoked
- **THEN** the CLI SHALL print to `ErrOut`: `A new version of germinator is available: <latest> (current: <current>)`
- **AND** the command SHALL complete its primary work without delay
- **AND** the notice SHALL NOT appear on `Out` (stdout stays clean for piping)

#### Scenario: Up-to-date

- **GIVEN** `consent = true`
- **AND** the latest GitHub release version equals the current binary's version
- **WHEN** any `germinator` command is invoked
- **THEN** no update notice SHALL be printed
- **AND** `last_check` SHALL be updated regardless

### Requirement: Update fetch timeout

The background goroutine SHALL bound the update fetch with a short timeout (default 5 seconds) and silently swallow any error. Update checking MUST NEVER fail the primary command.

#### Scenario: Network timeout

- **GIVEN** `consent = true` and the TTL has expired
- **WHEN** the background goroutine fetches the latest release
- **AND** the fetch exceeds the 5-second timeout
- **THEN** the goroutine SHALL exit silently (no error surfaced to the user)
- **AND** the primary command SHALL exit 0 regardless

### Requirement: Update check is disable-able via flag or env

The user SHALL be able to disable update checking for a single invocation via `--no-update-check` or the `GERMINATOR_NO_UPDATE_CHECK=1` environment variable.

#### Scenario: --no-update-check bypasses check

- **GIVEN** `consent = true` and the TTL has expired
- **WHEN** `germinator --no-update-check library resources` is invoked
- **THEN** no update check SHALL be performed
- **AND** `last_check` SHALL NOT be updated

### Requirement: No automatic update

The CLI SHALL NEVER auto-update the binary. The notice SHALL suggest the user run an explicit command (`germinator upgrade` — to be implemented in a future change) or visit the releases page.

#### Scenario: Notice suggests manual action

- **GIVEN** a new version is detected
- **WHEN** the notice is printed
- **THEN** it SHALL include either `Run 'germinator upgrade'` or the releases page URL
- **AND** it SHALL NOT trigger any download or replacement of the current binary

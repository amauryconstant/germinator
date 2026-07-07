# cli-color-policy Specification

## Purpose

Define when the CLI emits ANSI color codes. Color is enabled by default on a TTY, disabled when stdout is not a TTY, controllable via `--color=always|never|auto`, and respects the `NO_COLOR` environment variable (see [no-color.org](https://no-color.org/)).

## Requirements

### Requirement: Default color behavior

The CLI SHALL enable color output when both (a) stdout is a TTY and (b) the `NO_COLOR` environment variable is unset or empty. Otherwise color SHALL be disabled.

#### Scenario: Color on TTY without NO_COLOR

- **GIVEN** `NO_COLOR` is unset
- **AND** `IOStreams.IsStdoutTTY() == true`
- **WHEN** any styled output is rendered
- **THEN** the output SHALL contain ANSI escape codes (color, bold)

#### Scenario: Color disabled when stdout is a pipe

- **GIVEN** `IOStreams.IsStdoutTTY() == false`
- **WHEN** any styled output is rendered
- **THEN** the output SHALL NOT contain ANSI escape codes

#### Scenario: NO_COLOR disables color even on TTY

- **GIVEN** `NO_COLOR=1` is set
- **AND** `IOStreams.IsStdoutTTY() == true`
- **WHEN** any styled output is rendered
- **THEN** the output SHALL NOT contain ANSI escape codes (per no-color.org)

### Requirement: --color flag

The CLI SHALL expose a `--color=always|never|auto` flag on the root command. `auto` is the default and follows the rules above; `always` forces color regardless of TTY detection; `never` disables color regardless of detection or `NO_COLOR`.

#### Scenario: --color=always forces color on a pipe

- **GIVEN** `IOStreams.IsStdoutTTY() == false`
- **WHEN** `germinator --color=always library resources` is invoked
- **THEN** the output SHALL contain ANSI escape codes
- **AND** the user takes responsibility for the pipe-broken output

#### Scenario: --color=never disables color on TTY

- **GIVEN** `IOStreams.IsStdoutTTY() == true`
- **AND** `NO_COLOR` is unset
- **WHEN** `germinator --color=never library resources` is invoked
- **THEN** the output SHALL NOT contain ANSI escape codes

#### Scenario: --color=auto follows defaults

- **GIVEN** `IOStreams.IsStdoutTTY() == true`
- **AND** `NO_COLOR` is unset
- **WHEN** `germinator --color=auto library resources` is invoked
- **THEN** the output SHALL contain ANSI escape codes (TTY + no NO_COLOR → auto enables)

### Requirement: Color alone never conveys meaning

The CLI SHALL NOT use color as the sole conveyor of meaning. Every styled element SHALL also include a text label or prefix that conveys the meaning without color.

#### Scenario: Error includes both prefix and color

- **WHEN** `output.FormatError(io, err)` writes an error to `ErrOut`
- **THEN** the rendered text SHALL begin with `Error: ` (plain text)
- **AND** the `Error:` prefix SHALL be red when color is enabled
- **AND** the error message SHALL be fully readable when color is disabled (the prefix `Error:` still signals the severity)

### Requirement: Severity palette

The CLI SHALL use a consistent severity palette via `iostreams.Styles`:

| Severity | Color (when enabled) | Method on `Styles` |
|---|---|---|
| Error | red, bold | `Styles.Error(s)` |
| Success | green | `Styles.Success(s)` |
| Warning | yellow | `Styles.Warning(s)` |
| Dim / hint | gray | `Styles.Dim(s)` |
| Bold / emphasis | bold | `Styles.Bold(s)` |

#### Scenario: Styles methods exist

- **WHEN** `iostreams.NewStyles(true)` is called
- **THEN** the returned `Styles` SHALL expose `Error`, `Success`, `Warning`, `Dim`, `Bold` methods
- **AND** each method SHALL return a styled string when called
- **AND** each method SHALL return the input string unchanged when color is disabled

### Requirement: Color is decided once

The color decision SHALL be made at `IOStreams` construction time (in `iostreams.System()`), not per-call. The `Styles` struct SHALL be a snapshot of the decision; later changes to TTY state or `NO_COLOR` SHALL NOT affect already-constructed `Styles`.

#### Scenario: SetStdoutTTY does not recolor already-rendered output

- **WHEN** `IOStreams.SetStdoutTTY(true)` is called after some output has already been rendered
- **THEN** previously-rendered output SHALL NOT change retroactively
- **AND** only subsequent output SHALL be affected by the new TTY state

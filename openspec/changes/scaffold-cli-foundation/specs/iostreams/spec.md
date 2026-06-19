# iostreams Specification

## Purpose

Centralize terminal I/O behind an `IOStreams` struct that all commands use for stdin, stdout, stderr, verbose output, TTY detection, color rendering, and structured logging.

## ADDED Requirements

### Requirement: IOStreams struct

The `iostreams.IOStreams` struct SHALL provide the only terminal I/O boundary for commands.

#### Scenario: IOStreams fields

- **WHEN** an `IOStreams` instance is inspected
- **THEN** it SHALL have `In io.Reader`, `Out io.Writer`, `ErrOut io.Writer` fields
- **AND** a `Verbose bool` field
- **AND** a `Logger *slog.Logger` field (gated on `GERMINATOR_DEBUG`)
- **AND** a `Styles iostreams.Styles` field

### Requirement: System and Test constructors

The `iostreams` package SHALL provide two constructors.

#### Scenario: System() for production

- **WHEN** `iostreams.System()` is called from `main.go`
- **THEN** it SHALL return an `IOStreams` with `In = os.Stdin`, `Out = os.Stdout`, `ErrOut = os.Stderr`
- **AND** it SHALL use `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))` for TTY detection
- **AND** it SHALL set `Styles` to a TTY-aware or non-TTY-aware `Styles` based on TTY detection and `NO_COLOR`

#### Scenario: Test() for unit tests

- **WHEN** `iostreams.Test()` is called from a test
- **THEN** it SHALL return an `IOStreams` with `In`, `Out`, `ErrOut` backed by `bytes.Buffer` instances
- **AND** it SHALL set `IsStdoutTTY() = false`
- **AND** it SHALL expose the buffers so tests can assert on the captured output

### Requirement: TTY detection

The `IOStreams` struct SHALL expose TTY state for `stdout` and `stdin`.

#### Scenario: IsStdoutTTY

- **WHEN** `IOStreams.IsStdoutTTY()` is called
- **THEN** it SHALL return `true` if stdout is connected to a terminal
- **AND** `false` otherwise (file, pipe, buffer)

#### Scenario: IsInteractive

- **WHEN** `IOStreams.IsInteractive()` is called
- **THEN** it SHALL return `true` if both stdin AND stdout are TTYs
- **AND** `false` otherwise

### Requirement: Verbosef method

The `IOStreams.Verbosef` method SHALL print to `ErrOut` when `Verbose` is true.

#### Scenario: Verbosef with Verbose=true

- **GIVEN** an `IOStreams` instance with `Verbose = true` and a buffer-backed `ErrOut`
- **WHEN** `io.Verbosef("loading %d files", 5)` is called
- **THEN** the formatted string `loading 5 files` SHALL be written to `ErrOut`
- **AND** a trailing newline SHALL be appended

#### Scenario: Verbosef with Verbose=false

- **GIVEN** an `IOStreams` instance with `Verbose = false`
- **WHEN** `io.Verbosef("loading %d files", 5)` is called
- **THEN** nothing SHALL be written to `ErrOut`

### Requirement: Structured Logger

The `IOStreams.Logger` SHALL be a `*slog.Logger` gated on the `GERMINATOR_DEBUG` environment variable.

#### Scenario: Logger disabled by default

- **GIVEN** the `GERMINATOR_DEBUG` env var is unset
- **WHEN** `iostreams.System()` constructs the IOStreams
- **THEN** `Logger` SHALL be `slog.New(slog.NewTextHandler(io.Discard, nil))` (no-op)

#### Scenario: Logger enabled in debug mode

- **GIVEN** the `GERMINATOR_DEBUG` env var is set to any non-empty value
- **WHEN** `iostreams.System()` constructs the IOStreams
- **THEN** `Logger` SHALL be a debug-level JSON handler writing to `ErrOut`

### Requirement: Styles struct

The `iostreams.Styles` struct SHALL provide color-rendering helpers via `github.com/charmbracelet/lipgloss`.

#### Scenario: Styles methods

- **WHEN** an `IOStreams` instance is inspected
- **THEN** its `Styles` field SHALL expose `Error()`, `Success()`, `Warning()`, `Dim()`, `Bold()` methods
- **AND** each method SHALL take a `string` and return the styled (or unstyled) `string`

#### Scenario: NO_COLOR respected

- **GIVEN** the `NO_COLOR` env var is set to any non-empty value
- **WHEN** any `Styles` method is called
- **THEN** it SHALL return the input string unchanged (no ANSI codes)

### Requirement: SetStdoutTTY override

`IOStreams.SetStdoutTTY(bool)` SHALL allow tests to override TTY detection.

#### Scenario: Override in tests

- **GIVEN** an `IOStreams` instance from `iostreams.Test()` (buffer-backed, `IsStdoutTTY() = false`)
- **WHEN** `io.SetStdoutTTY(true)` is called
- **THEN** `IsStdoutTTY()` SHALL return `true` until `SetStdoutTTY(false)` is called

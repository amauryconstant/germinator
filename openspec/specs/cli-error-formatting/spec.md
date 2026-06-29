# error-formatting Specification

## Purpose

Provide typed-error formatting with a single centralized `output.FormatError` entry point. Errors are dispatched to private per-type helpers via `errors.As`, so the rendering rules live next to the error type itself and the command layer never branches on error type.

## Requirements

### Requirement: Errors formatted via output.FormatError

Error formatting SHALL be centralized in `output.FormatError(io *iostreams.IOStreams, err error)` (introduced in `output-formats`). The legacy `cmd.ErrorFormatter` struct and `cmd.NewErrorFormatter(...)` constructor SHALL be removed (see `output-formats` for the rendering contract).

#### Scenario: FormatError writes to ErrOut

- **WHEN** a command's `RunE` returns a typed error
- **AND** `output.FormatError(opts.IO, err)` is called
- **THEN** the formatted error SHALL be written to `opts.IO.ErrOut`
- **AND** `opts.IO.Styles.Error()` SHALL style the prefix

### Requirement: Typed-error dispatch

`output.FormatError(io, err)` SHALL dispatch on typed errors via `errors.As`:

- `*core.ParseError` → render: `Error: parse failed at <path>: <message>`
- `*core.ValidationError` → render: `Error: validation failed: <message>` followed by per-error list
- `*core.TransformError` → render: `Error: transform failed: <message>`
- `*core.FileError` → render: `Error: <op> <path>: <message>`
- `*core.ConfigError` → render: `Error: config: <message>`
- `*core.PartialSuccessError` → render: `partial success: N succeeded, M failed` followed by per-error lines
- `*core.NotFoundError` → render: `Error: not found: <key>`
- generic error → render: `Error: <err.Error()>`

#### Scenario: FormatError table-driven

- **WHEN** `FormatError` is called with each error type from the list above
- **THEN** it SHALL write the corresponding formatted string to `io.ErrOut`
- **AND** `io.Styles.Error()` SHALL style the prefix

### Requirement: Type-specific formatters are private

Each typed-error formatter in `output/errors.go` SHALL be a private function (lowercase). Only `FormatError` is exported.

#### Scenario: Private formatters

- **WHEN** the `internal/output/errors.go` file is inspected
- **THEN** only `FormatError` SHALL be exported (start with uppercase)
- **AND** per-type helpers (e.g. `formatParseError`, `formatValidationError`) SHALL be package-private

### Requirement: Multiple validation errors

The system SHALL format multiple validation errors clearly.

#### Scenario: Format validation error list

- **WHEN** a `*core.ValidationError` carries multiple `Errors`
- **THEN** `FormatError` SHALL write each error on a separate line
- **AND** each error SHALL be numbered or bulleted

### Requirement: Error cause chain

The system MAY include the wrapped cause in the output for debugging.

#### Scenario: Include cause for debugging

- **WHEN** a typed error wraps an underlying error via `fmt.Errorf("...: %w", inner)`
- **THEN** `FormatError` MAY append a clearly separated cause line
- **AND** the cause SHALL be indented to distinguish it from the primary message

### Requirement: FormatError dispatches on core.NotFoundError

`output.FormatError` SHALL dispatch on `*core.NotFoundError` (introduced in this slice's task group 4.0) and render a styled message to stderr.

#### Scenario: NotFoundError renders to stderr

- **WHEN** `output.FormatError(io, err)` is called and `err` is (or wraps) a `*core.NotFoundError`
- **THEN** the message SHALL be written to `io.ErrOut`
- **AND** the message SHALL be `io.Styles.Error("Error: ") + "not found: " + e.Key + "\n"`
- **AND** `stdout` SHALL NOT receive the error message

#### Scenario: NotFoundError detection via errors.As

- **WHEN** `errors.As(err, &target)` is called with `var target *core.NotFoundError`
- **THEN** the call SHALL return `true` for any error (or wrapped error) of type `*core.NotFoundError`

#### Scenario: NotFoundError maps to ExitCodeError

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `cmdutil.ExitCodeError` (1) via the default-error case at `internal/cmdutil/exit.go:71`

### Requirement: core.NotFoundError type and constructor

`internal/core/errors.go` SHALL define a `NotFoundError` struct, a constructor, and an `Error()` method that produces the canonical not-found message.

#### Scenario: NotFoundError.Error format

- **WHEN** `core.NewNotFoundError("library ref", "nonexistent-ref")` is called
- **THEN** the returned error's `Error()` method SHALL return `"not found: nonexistent-ref"`

#### Scenario: NotFoundError fields

- **WHEN** a `*core.NotFoundError` is constructed with `NewNotFoundError(entity, key)`
- **THEN** the struct SHALL expose the `Entity` and `Key` fields for programmatic inspection (via accessor methods or exported fields)

> **Status:** `core.NotFoundError` and the `FormatError` dispatch branch are introduced in task group 4.0 of `migrate-library-readonly`. No other commands in the slice consume the type directly, but downstream slices (5: `init`, 6: `library add`/`library create`, 7: remaining library commands) may use it for additional not-found scenarios.

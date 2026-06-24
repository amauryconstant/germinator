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

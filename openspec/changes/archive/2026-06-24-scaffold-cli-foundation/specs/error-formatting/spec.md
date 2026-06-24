# error-formatting Specification (delta)

## MODIFIED Requirements

### Requirement: Errors formatted via output.FormatError

Error formatting SHALL be centralized in `output.FormatError(io *iostreams.IOStreams, err error)` (introduced in `output-formats`). The legacy `cmd.ErrorFormatter` struct and `cmd.NewErrorFormatter(...)` constructor SHALL be deleted in change-7 (see `## REMOVED Requirements` below).

#### Scenario: FormatError writes to ErrOut

- **WHEN** a command's `RunE` returns a typed error
- **AND** `output.FormatError(opts.IO, err)` is called
- **THEN** the formatted error SHALL be written to `opts.IO.ErrOut`
- **AND** `opts.IO.Styles.Error()` SHALL style the prefix

> **Status (slice 1 / foundation):** `output.FormatError` exists with full table-driven tests covering each typed error. `ErrorFormatter` still exists in `cmd/error_formatter.go`. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

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

> **Status (slice 1 / foundation):** the function exists with full table-driven tests in `internal/output/output_test.go`. The legacy `ErrorFormatter` is removed in changes 2 and 7 as noted in the REMOVED section.

### Requirement: Type-specific formatters are private

Each typed-error formatter in `output/errors.go` SHALL be a private function (lowercase). Only `FormatError` is exported.

#### Scenario: Private formatters

- **WHEN** the `internal/output/errors.go` file is inspected
- **THEN** only `FormatError` SHALL be exported (start with uppercase)
- **AND** per-type helpers (e.g. `formatParseError`, `formatValidationError`) SHALL be package-private

> **Status (slice 1 / foundation):** the structure is established. Per-type formatters are private in `internal/output/errors.go`. The legacy `ErrorFormatter` is removed in changes 2 and 7 as noted in the REMOVED section.

## REMOVED Requirements

### Requirement: ErrorFormatter struct

**Reason**: `ErrorFormatter` is a per-command formatter that depends on `CommandConfig` and is constructed in `RunE`. It duplicates logic that belongs to typed-error dispatch via `errors.As`.

**Migration**: Use `output.FormatError(opts.IO, err)` from the command's `RunE` closure (or defer to `cmdutil.ExitCodeFor` for the exit code mapping).

#### Scenario: ErrorFormatter removed

- **WHEN** the codebase is inspected after change-7
- **THEN** the `ErrorFormatter` type SHALL NOT be defined
- **AND** `NewErrorFormatter` SHALL NOT be defined
- **AND** `cmd/error_formatter.go` SHALL be deleted

> **Status (slice 1 / foundation):** `ErrorFormatter` still exists. The legacy surface is removed in changes 2 and 7 as noted in the REMOVED section.

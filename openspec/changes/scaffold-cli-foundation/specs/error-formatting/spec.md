# error-formatting Specification (delta)

## MODIFIED Requirements

### Requirement: ErrorFormatter replaced by FormatError

The `cmd.ErrorFormatter` struct and `cmd.NewErrorFormatter(...)` constructor SHALL be **removed**. Error formatting SHALL be centralized in `output.FormatError(io *iostreams.IOStreams, err error)` (introduced in `cli/output-formats`).

#### Scenario: ErrorFormatter type removed

- **WHEN** the codebase is inspected
- **THEN** the `ErrorFormatter` type SHALL NOT be defined
- **AND** `NewErrorFormatter` SHALL NOT be defined
- **AND** `cmd/error_formatter.go` SHALL be deleted (deletion happens in change-7)

> **Status (slice 1 / foundation):** `output.FormatError` exists with table-driven tests covering each typed error. `ErrorFormatter` still exists in `cmd/error_formatter.go`. Removal happens in change-7 (after all commands using it are migrated in changes 2-6).

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

> **Status (slice 1 / foundation):** the function exists with table-driven tests in `internal/output/output_test.go`.

### Requirement: Type-specific formatters are private

Each typed-error formatter in `output/errors.go` SHALL be a private function (lowercase). Only `FormatError` is exported.

#### Scenario: Private formatters

- **WHEN** the `internal/output/errors.go` file is inspected
- **THEN** only `FormatError` SHALL be exported (start with uppercase)
- **AND** per-type helpers (e.g. `formatParseError`, `formatValidationError`) SHALL be package-private

> **Status (slice 1 / foundation):** the structure is established. Per-type formatters are private in `internal/output/errors.go`.

# error-formatting Specification (delta)

## MODIFIED Requirements

### Requirement: Typed-error dispatch

`output.FormatError(io, err)` SHALL dispatch on typed errors via `errors.As`:

- `*core.ParseError` → render: `Error: parse failed at <path>: <message>`
- `*core.ValidationError` → render: `Error: validation failed: <message>` followed by per-error list
- `*core.TransformError` → render: `Error: transform failed: <message>`
- `*core.FileError` → render: `Error: <op> <path>: <message>`
- `*core.ConfigError` → render: `Error: config: <message>`
- `*core.PartialSuccessError` → render: `partial success: N succeeded, M failed` followed by per-error lines
- `*core.NotFoundError` → render: `Error: not found: <key>`
- `*core.OperationError` → render: `Error: <op> failed: <resource>: <message>`
- `*core.InitializeError` → render: `Error: initialize failed: <ref>` (NEW in change `enforce-error-discipline`)
- `*core.UsageError` → render: `Error: <flag>: <reason>` (NEW in change `enforce-error-discipline`)
- generic error → render: `Error: <err.Error()>`

**Change**: ADDED `*core.InitializeError` and `*core.UsageError` cases. The pre-change switch handled only 8 of the 9 existing typed errors (missing `InitializeError`); `UsageError` is a new type introduced in this change. The dispatch set is now complete for all 10 typed errors.

#### Scenario: FormatError table-driven

- **WHEN** `FormatError` is called with each error type from the list above
- **THEN** it SHALL write the corresponding formatted string to `io.ErrOut`
- **AND** `io.Styles.Error()` SHALL style the prefix

#### Scenario: InitializeError renders to stderr

- **WHEN** `output.FormatError(io, err)` is called and `err` is (or wraps) a `*core.InitializeError`
- **THEN** the message SHALL be written to `io.ErrOut`
- **AND** the message SHALL be `io.Styles.Error("Error: ") + "initialize failed: " + e.Ref + "\n"`

#### Scenario: UsageError renders to stderr

- **WHEN** `output.FormatError(io, err)` is called and `err` is (or wraps) a `*core.UsageError`
- **THEN** the message SHALL be written to `io.ErrOut`
- **AND** the message SHALL be `io.Styles.Error("Error: ") + e.Flag + ": " + e.Reason + "\n"`

#### Scenario: NotFoundError maps to ExitCodeError

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `cmdutil.ExitCodeError` (1) — see the cross-reference to `cli-exit-codes/spec.md` for the corrected mapping (was 2; corrected in change `enforce-error-discipline`).

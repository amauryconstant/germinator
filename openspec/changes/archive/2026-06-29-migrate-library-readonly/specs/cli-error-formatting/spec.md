# cli-error-formatting Specification (delta)

## MODIFIED Requirements

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

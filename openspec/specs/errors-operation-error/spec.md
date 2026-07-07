# Capability: Operation Error

## Purpose

The Operation Error capability defines `core.OperationError`, a typed error for per-operation failures that flow through `*core.PartialSuccessError` aggregations and through `output.FormatError`. It is the foundation unit used by per-resource library operations (e.g., the batch-discovery path of `library add --discover --batch`, where each orphan discovery failure is wrapped in an `OperationError`) to surface a uniform `<op>: <resource>` message plus the wrapped underlying cause.

## Requirements

### Requirement: OperationError type definition

The `core` package SHALL provide `*core.OperationError{Op, Resource string; Cause error}` for typed operation-level errors. The constructor SHALL be `core.NewOperationError(op, resource string, cause error) *OperationError`.

#### Scenario: Constructor stores fields

- **WHEN** `core.NewOperationError("register", "skill/commit", errors.New("name taken"))` is called
- **THEN** the returned `*OperationError` SHALL have `Op == "register"`
- **AND** `Resource == "skill/commit"`
- **AND** `Cause` SHALL be the wrapped error

#### Scenario: Error() formats as "<op>: <resource>"

- **WHEN** `e := core.NewOperationError("register", "skill/commit", nil); e.Error()` is called
- **THEN** the result SHALL be the string `"register: skill/commit"` (no trailing newline; the formatter adds it)

#### Scenario: Unwrap returns the cause

- **WHEN** `cause := errors.New("original"); e := core.NewOperationError("register", "skill/commit", cause)` is created
- **THEN** `errors.Unwrap(e)` SHALL return `cause`
- **AND** `errors.Is(e, cause)` SHALL be `true`

#### Scenario: errors.As detects the type

- **WHEN** `var target *core.OperationError; errors.As(err, &target)` is called with any `*core.OperationError` value (including wrapped ones via `fmt.Errorf("...%w", e)`)
- **THEN** `target` SHALL be the unwrapped `*core.OperationError`
- **AND** `target != nil`

### Requirement: OperationError rendering via FormatError

`output.FormatError` SHALL dispatch on `*core.OperationError` via `errors.As` and render `Error: <op>: <resource>\n` to **stderr** (`opts.IO.ErrOut`) via the existing `Styles.Error` channel. The wrapped `Cause` is rendered on a separate indented line if non-nil.

#### Scenario: FormatError renders OperationError to stderr

- **GIVEN** `err := core.NewOperationError("register", "skill/commit", nil)`
- **WHEN** `output.FormatError(io, err)` is called
- **THEN** `io.ErrOut` SHALL contain `"Error: register: skill/commit\n"`
- **AND** `io.Out` SHALL be empty (no data leakage on error paths)

#### Scenario: FormatError renders wrapped cause

- **GIVEN** `err := core.NewOperationError("register", "skill/commit", errors.New("name taken by skill/x"))`
- **WHEN** `output.FormatError(io, err)` is called
- **THEN** `io.ErrOut` SHALL contain `"Error: register: skill/commit"` on the first line
- **AND** `io.ErrOut` SHALL contain the wrapped cause `"name taken by skill/x"` on a subsequent indented line
- **AND** `io.Out` SHALL be empty

#### Scenario: FormatError dispatch precedence

- **GIVEN** an error chain `fmt.Errorf("registering: %w", core.NewOperationError("register", "skill/commit", nil))`
- **WHEN** `output.FormatError(io, err)` is called
- **THEN** the `OperationError` branch SHALL be selected via `errors.As`
- **AND** the outer wrapping string SHALL NOT appear in the rendered output (the typed error owns the user-facing message)

### Requirement: cmdutil.ExitCodeFor maps OperationError

`cmdutil.ExitCodeFor(err)` SHALL map `*core.OperationError` to `ExitCodeError` (1) via the default-error case in `internal/cmdutil/exit.go:82` (the final `return ExitCodeError` after all `errors.As` branches miss).

#### Scenario: ExitCodeFor returns 1 for OperationError

- **WHEN** `cmdutil.ExitCodeFor(core.NewOperationError("register", "skill/commit", nil))` is called
- **THEN** the result SHALL be `cmdutil.ExitCodeError` (1)

#### Scenario: ExitCodeFor returns 0 for OperationError wrapped in PartialSuccessError

- **GIVEN** `err := core.NewPartialSuccessError(succeeded=1, failed=1, errs: []core.InitializeError{{Ref: "skill/commit", Cause: core.NewOperationError("register", "skill/commit", nil)}})`
- **WHEN** `cmdutil.ExitCodeFor(err)` is called
- **THEN** the result SHALL be `cmdutil.ExitCodeSuccess` (0) because `Succeeded > 0` (the slice-5 partial-success rule takes precedence)

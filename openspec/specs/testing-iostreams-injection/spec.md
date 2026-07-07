# testing-iostreams-injection Specification

## Purpose

Define the `iostreams.Test()` pattern used across all command tests: how to inject buffer-backed `IOStreams`, how to assert on captured `Out` and `ErrOut`, and how to test TTY vs non-TTY behavior paths.

## Requirements

### Requirement: iostreams.Test() constructor

The `iostreams` package SHALL provide a `Test() *IOStreams` constructor that returns an `IOStreams` instance backed by `*bytes.Buffer` instances for `In`, `Out`, and `ErrOut`. The TTY predicates SHALL default to `false`.

#### Scenario: Test() returns buffer-backed IOStreams

- **WHEN** `io := iostreams.Test()` is called
- **THEN** `io.Out`, `io.ErrOut`, and `io.In` SHALL each be `*bytes.Buffer` (assertable in tests via type assertion)
- **AND** `io.IsStdoutTTY()` SHALL return `false`
- **AND** `io.IsStdinTTY()` SHALL return `false`
- **AND** `io.IsInteractive()` SHALL return `false`

#### Scenario: Test() captures Verbosef output

- **GIVEN** an `IOStreams` from `iostreams.Test()`
- **WHEN** `io.Verbose = true; io.Verbosef("loading %d files", 5)` is called
- **THEN** `io.ErrOut.(*bytes.Buffer).String()` SHALL contain `loading 5 files`

### Requirement: Test() disables color

`iostreams.Test()` SHALL construct a `Styles` value with color disabled (no ANSI codes).

#### Scenario: Test styles are unstyled

- **GIVEN** an `IOStreams` from `iostreams.Test()`
- **WHEN** `io.Styles.Error("oops")` is called
- **THEN** the returned string SHALL be `oops` (no ANSI codes)
- **AND** tests can assert against the bare string without escape-code stripping

### Requirement: Test() gates Logger on no-op handler

`iostreams.Test()` SHALL construct a `Logger` that writes to `io.Discard` by default (not `ErrOut`). Tests that want to assert on log output MAY swap `io.Logger` for a custom handler.

#### Scenario: Test Logger does not pollute ErrOut

- **GIVEN** an `IOStreams` from `iostreams.Test()`
- **WHEN** `io.Logger.Debug("internal state", "key", "value")` is called
- **THEN** `io.ErrOut.(*bytes.Buffer)` SHALL be empty (Logger writes to `io.Discard`)
- **AND** `io.Out.(*bytes.Buffer)` SHALL be empty

### Requirement: SetStdoutTTY override for TTY tests

Tests SHALL be able to override TTY state via `io.SetStdoutTTY(true)` to exercise the colored/styled code path without a real terminal.

#### Scenario: SetStdoutTTY enables styled output

- **GIVEN** an `IOStreams` from `iostreams.Test()`
- **WHEN** `io.SetStdoutTTY(true)` is called
- **THEN** `io.IsStdoutTTY()` SHALL return `true`
- **AND** `io.Styles.Error("oops")` SHALL return a string with ANSI codes (e.g., `\x1b[31;1moops\x1b[0m`)

### Requirement: Table-driven IOStreams tests

The recommended pattern for command tests SHALL be a table-driven test that iterates over scenarios, each capturing `Out` and `ErrOut` from a fresh `iostreams.Test()` buffer pair.

#### Scenario: Captured-buffer assertion

```go
ios, _, out, errOut := iostreams.Test()
// ... invoke the command ...
assert.Contains(t, out.String(), "Installed: skill/commit -> ...")
assert.Contains(t, errOut.String(), "Error: not found")
```

- **WHEN** a command test asserts on `out.String()` and `errOut.String()`
- **THEN** the assertions SHALL reflect only the command's `Out` and `ErrOut` writes (no stderr leak from Logger)

### Requirement: RunF injection for testability

Commands constructed via `NewCmdXxx(f, runF)` SHALL accept a `runF func(*XxxOptions) error` parameter for test injection. When `runF` is non-nil, the constructor's `RunE` SHALL call `runF(opts)` instead of `runXxx(opts)`. Tests pass a closure that captures `opts` for assertion.

#### Scenario: runF captures options

- **GIVEN** a command `NewCmdFoo(f, func(opts *FooOptions) error { captured = opts; return nil })`
- **WHEN** `cmd.SetArgs([]string{"arg1", "--flag", "value"})` is set and `cmd.Execute()` is called
- **THEN** `captured.Name` SHALL be `"arg1"`
- **AND** `captured.Flag` SHALL be `"value"`
- **AND** `runFoo` SHALL NOT have been invoked (production body skipped)

#### Scenario: runF can inject errors

- **GIVEN** a command constructed with `runF` returning `errors.New("boom")`
- **WHEN** the command is invoked
- **THEN** the command's `RunE` SHALL return the injected error
- **AND** `cmdutil.ExitCodeFor(err)` SHALL map the error (test asserts exit code behavior)

### Requirement: No real OS streams in unit tests

Unit tests SHALL NOT reference `os.Stdout`, `os.Stderr`, `os.Stdin`, or `iostreams.System()`. All tests SHALL use `iostreams.Test()`. The `forbidigo` linter MAY enforce this rule (see `cmd/lint_helpers.go` for the patterns).

#### Scenario: Linter flags direct os.Stdout/Stderr usage

- **WHEN** a test file in `cmd/` contains `fmt.Fprintf(os.Stdout, ...)` or `os.Exit(...)`
- **THEN** `forbidigo` SHALL report a violation
- **AND** the test SHALL be rewritten to use `iostreams.Test()` and `cmdutil.ExitCodeFor`

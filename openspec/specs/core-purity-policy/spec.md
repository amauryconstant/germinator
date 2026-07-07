# core-purity-policy Specification

## Purpose

Define and enforce the purity rule for `internal/core/`: the core may not import any I/O packages (`os`, `net`, `exec`), may not perform filesystem access, network calls, or subprocess execution, and depends only on the Go standard library and `github.com/samber/lo`.

## Requirements

### Requirement: Core may not import I/O packages

The `internal/core/` package SHALL NOT import any of the following:

- `os` (and subpackages like `os/exec`, `os/signal`)
- `net` (and subpackages like `net/http`, `net/url`)
- `io/fs` (filesystem abstraction)
- `path/filepath` (filesystem path manipulation)
- `io` (limited; only for plain byte streams when a generic helper is unavoidable — but never for actual file I/O)
- `github.com/spf13/cobra`, `github.com/spf13/pflag`, `github.com/spf13/viper` (CLI concerns)
- `os/exec` (subprocess execution)
- `log/slog` (logging is a shell concern)

#### Scenario: depguard denies I/O imports

- **WHEN** a file under `internal/core/` imports `os` or `net`
- **THEN** `mise run lint` SHALL fail the depguard rule `core-isolation`
- **AND** the violation SHALL be reported with the line number and offending import

### Requirement: Core depends only on stdlib + samber/lo

The `internal/core/` package SHALL import only:

- Go standard library (excluding I/O packages per the rule above)
- `github.com/samber/lo` (collection utilities — zero dependencies)

#### Scenario: depguard allows stdlib + lo

- **WHEN** `internal/core/errors.go` imports `fmt`, `strings`, and `github.com/samber/lo`
- **THEN** the depguard rule `core-isolation` SHALL pass
- **AND** no violations SHALL be reported

### Requirement: Pure functions

All exported functions and methods in `internal/core/` SHALL be pure: same inputs → same outputs, no observable side effects, no global state mutation, no time-dependent behavior except via injected `time.Time` or `time.Clock` values.

#### Scenario: Pure validation

- **WHEN** `core.NewValidationError("Agent", "name", "invalid", "name is required")` is called twice with identical arguments
- **THEN** both returned errors SHALL have identical field values
- **AND** no global state SHALL be modified
- **AND** no I/O SHALL be performed

#### Scenario: Side effects are forbidden

- **WHEN** a function in `internal/core/` is reviewed
- **THEN** it SHALL NOT call `os.Open`, `os.WriteFile`, `os.Stat`, `net.Dial`, `exec.Command`, or any I/O primitive
- **AND** it SHALL NOT mutate any global variable
- **AND** it SHALL NOT call `time.Now()` directly (time must be injectable)

### Requirement: Errors are typed and exit-code-free

Errors defined in `internal/core/` SHALL carry semantic meaning (field, value, message) but SHALL NOT carry exit codes or formatting logic. Exit-code mapping happens in `internal/cmdutil/exit.go` (`ExitCodeFor`); formatting happens in `internal/output/errors.go` (`FormatError`).

#### Scenario: Core errors have no ExitCode method

- **WHEN** `core.NotFoundError` is inspected
- **THEN** it SHALL NOT expose an `ExitCode()` method
- **AND** it SHALL NOT contain a `code int` field
- **AND** `cmdutil.ExitCodeFor(notFoundErr)` SHALL return the correct code (2) via `errors.As`

### Requirement: Core has zero non-stdlib, non-lo dependencies

The `internal/core/` package's `go.mod` requirements graph (excluding transitive dependencies of `lo` itself, which are zero) SHALL contain only:

- `github.com/samber/lo` (direct)

No other direct or indirect dependencies SHALL appear.

#### Scenario: depguard denies non-stdlib, non-lo imports

- **GIVEN** `internal/core/agent.go` accidentally imports `github.com/charmbracelet/lipgloss`
- **WHEN** `mise run lint` runs the depguard rules
- **THEN** the depguard rule SHALL deny the import with the message: `core allows only stdlib and lo`

### Requirement: The depguard rule is documented and stable

The depguard rule SHALL live in `.golangci.yml` (or the equivalent linter configuration file). The rule SHALL be:

```yaml
linters-settings:
  depguard:
    rules:
      core-isolation:
        files:
          - "**/core/**"
        allow:
          - $gostd
          - github.com/samber/lo
        deny:
          - pkg: "github.com/*"
            desc: "core allows only stdlib and lo"
```

#### Scenario: depguard config exists

- **WHEN** `.golangci.yml` is inspected
- **THEN** it SHALL contain a `depguard.rules.core-isolation` entry
- **AND** the entry SHALL match the documented shape above

### Requirement: Purity rule is reflected in tests

`internal/core/` tests SHALL NOT use mocks, stubs, or test doubles. All tests SHALL pass plain Go values to core functions and assert on the returned values. The functional core is tested with values in, values out.

#### Scenario: Core test uses no mocks

- **WHEN** `internal/core/errors_test.go` is inspected
- **THEN** it SHALL NOT contain any mock framework import (`github.com/stretchr/testify/mock`, `github.com/stretchr/testify/assert`, etc., except for `assert` for assertions)
- **AND** every test SHALL be table-driven with literal input/output values

> **Note:** `testify/assert` and `testify/require` are allowed in core tests because they do not perform I/O. They are assertion libraries, not mocking frameworks.

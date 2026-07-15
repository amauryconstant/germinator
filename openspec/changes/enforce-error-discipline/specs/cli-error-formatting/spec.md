# error-formatting Specification (delta)

> **Cross-references:** this change also modifies `errors-typed-errors`, `cli-exit-codes`, `errors-enhanced-errors`. See those delta specs.

## REMOVED Requirements

### Scenario: NotFoundError maps to ExitCodeUsage

The prior scenario *"NotFoundError maps to ExitCodeUsage"* (from `openspec/specs/cli-error-formatting/spec.md:85-88`) is removed. The exit-code mapping for `*core.NotFoundError` lives in `cli-exit-codes/spec.md` after this change lands; `cli-error-formatting` is scoped to the `output.FormatError` rendering contract only.

> **Why:** the exit-code contract is owned by `cli-exit-codes`; keeping the two specs aligned avoids duplicate assertions that can drift.

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
- `*core.InitializeError` → render: `Error: <e.Error()>` (NEW in change `enforce-error-discipline`; delegates to `InitializeError.Error()` which renders `initialize failed: <ref>: output: <outputPath>: <cause.Error()>` as a single colon-joined line per `internal/core/errors.go:670-687` — parts joined by `: `; `output` segment optional; cause optional; optional `(context)` suffix; optional `\n💡 <suggestions>` block)
- `*core.UsageError` → render: `Error: <flag>: <reason>` (NEW in change `enforce-error-discipline`)
- `*config.WriteError` → render: `Error: <op> <path>: <message>` (NEW in change `enforce-error-discipline`; `*config.WriteError` carries `op`, `path`, `cause` per `internal/config/errors.go`)
- generic error → render: `Error: <err.Error()>`

**Change**: ADDED `*core.InitializeError` and `*core.UsageError` cases. The pre-change switch handled only 8 of the 9 existing typed errors (missing `InitializeError`); `UsageError` is a new type introduced in this change. The dispatch set is now complete for all 11 core typed errors (9 existing + 2 new) plus 1 Imperative Shell typed error (`*config.WriteError`) — 12 arms + default fallback.

The dispatch ordering matches the order above (full canonical order: Parse, Validation, Transform, File, Config, PartialSuccess, NotFound, Operation, Initialize, Usage, WriteError, default); the new cases sit after `OperationError` and before the generic-error fallback.

#### Scenario: FormatError table-driven

- **WHEN** `FormatError` is called with each error type from the list above
- **THEN** it SHALL write the corresponding formatted string to `io.ErrOut`
- **AND** `io.Styles.Error()` SHALL style the prefix

#### Scenario: InitializeError renders to stderr

- **WHEN** `output.FormatError(io, err)` is called and `err` is (or wraps) a `*core.InitializeError`
- **THEN** the message SHALL be written to `io.ErrOut`
- **AND** the message SHALL be `io.Styles.Error("Error: ") + e.Error() + "\n"` (delegates to `InitializeError.Error()`; the dispatch case MUST NOT manually concatenate the segments — `e.Error()` is the single source of truth per `internal/core/errors.go`)

#### Scenario: InitializeError with cause and outputPath renders the chain

- **WHEN** `*core.InitializeError` carries a non-nil `Cause` and a non-empty `OutputPath()`
- **THEN** the rendered message SHALL be `io.Styles.Error("Error: ") + e.Error() + "\n"` — delegating to `InitializeError.Error()` (which renders `initialize failed: <ref>: output: <outputPath>: <cause.Error()>` as a single colon-joined line per `InitializeError.Error()` declaration in `internal/core/errors.go`). The dispatch case MUST NOT manually concatenate the segments; `e.Error()` is the single source of truth.

#### Scenario: UsageError renders to stderr

- **WHEN** `output.FormatError(io, err)` is called and `err` is (or wraps) a `*core.UsageError`
- **THEN** the message SHALL be written to `io.ErrOut`
- **AND** the message SHALL be `io.Styles.Error("Error: ") + e.Flag() + ": " + e.Reason() + "\n"`

#### Scenario: UsageError renders via the new clean-break wording

- **WHEN** `output.FormatError(io, err)` is called with `*core.UsageError{Flag: "--resources", Reason: "must be non-empty list of refs"}`
- **THEN** the message SHALL be `io.Styles.Error("Error: ") + "--resources: must be non-empty list of refs\n"`
- **AND** the message SHALL NOT contain the string `"flag needs an argument"` (the prior Cobra-encoded phrasing is dropped)

#### Scenario: Dispatch ordering — wrapped OperationError carries an InitializeError

- **WHEN** `output.FormatError(io, err)` is called with `*core.OperationError{Op: "init", Resource: "skill/commit", Cause: innerErr}` where `innerErr` is or wraps `*core.InitializeError`
- **THEN** the message SHALL be rendered via the `OperationError` dispatch case (first matching case in the switch order at line 87 wins; OperationError is case 8, InitializeError is case 9)
- **AND** the primary body SHALL be `<Op>: <Resource>` (per `formatOperationError` in `internal/output/errors.go:120-132`)
- **AND** when `Cause` is non-nil, the InitializeError's `Error()` SHALL appear as an indented sub-line (per the existing `Cause` rendering block in `formatOperationError`)
- **AND** the message SHALL NOT use InitializeError's colon-joined single-line body as the primary message
- **NOTE**: the rule "first matching case in switch order wins" is the documented dispatch contract; the scenario below ("wrapped InitializeError inside fmt.Errorf %w") demonstrates the rule for the non-OperationError case where InitializeError wins because no other typed case appears before it in the switch.

#### Scenario: Dispatch ordering — wrapped InitializeError inside fmt.Errorf %w

- **WHEN** `output.FormatError(io, err)` is called with `fmt.Errorf("loading library: %w", core.NewInitializeError("skill/commit", "/in", "/out", underlyingErr))`
- **THEN** the message SHALL be rendered via the `InitializeError` dispatch case (errors.As traverses the wrap chain)
- **AND** the message SHALL be `io.Styles.Error("Error: ") + e.Error() + "\n"` (delegate to `InitializeError.Error()`; the outer `loading library:` prefix is not part of the typed error and is NOT rendered because the `InitializeError` case wins over the generic fallback that would otherwise include the outer text via `err.Error()`).

## ADDED Requirements

### Requirement: InitializeError dispatch table member

`output.FormatError` SHALL dispatch on `*core.InitializeError` and render the user-facing message to stderr. The pre-change switch missed this case; the case is added in `internal/output/errors.go` per change `enforce-error-discipline` (the switch spans lines 31-50; the new `case *core.InitializeError` arm sits after `OperationError` and before the generic-error fallback).

#### Scenario: InitializeError renders to stderr via dispatch

- **WHEN** `output.FormatError(io, err)` is called with `*core.InitializeError`
- **THEN** the dispatch SHALL match the `case *core.InitializeError` arm
- **AND** no other case shall match first (the case order in the switch is: Parse, Validation, Transform, File, Config, PartialSuccess, NotFound, Operation, **Initialize**, Usage, WriteError, default)

### Requirement: UsageError dispatch table member

`output.FormatError` SHALL dispatch on `*core.UsageError` and render the user-facing message to stderr. The case is added per change `enforce-error-discipline`.

#### Scenario: UsageError renders to stderr via dispatch

- **WHEN** `output.FormatError(io, err)` is called with `*core.UsageError`
- **THEN** the dispatch SHALL match the `case *core.UsageError` arm
- **AND** the rendered message SHALL be `"Error: <flag>: <reason>\n"` (where `<flag>` and `<reason>` come from the typed-error's accessors)

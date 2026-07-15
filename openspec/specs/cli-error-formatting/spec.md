# error-formatting Specification

> **Cross-references:** this change also modifies `errors-typed-errors`, `cli-exit-codes`, `errors-enhanced-errors`. See those delta specs.

## Purpose

Provide typed-error formatting with a single centralized `output.FormatError` entry point. Errors are dispatched to private per-type helpers via `errors.As`, so the rendering rules live next to the error type itself and the command layer never branches on error type.

## Requirements

### Requirement: Errors formatted via output.FormatError

Error formatting SHALL be centralized in `output.FormatError(io *iostreams.IOStreams, err error)` (see `cli-output-formats` for the rendering contract).

#### Scenario: FormatError writes to ErrOut

- **WHEN** a command's `RunE` returns a typed error
- **AND** `output.FormatError(opts.IO, err)` is called
- **THEN** the formatted error SHALL be written to `opts.IO.ErrOut`
- **AND** `opts.IO.Styles.Error()` SHALL style the prefix

### Requirement: Typed-error dispatch

`output.FormatError(io, err)` SHALL dispatch on typed errors via `errors.As`, in the canonical order documented below (first matching case wins):

- `*core.ParseError` → render: `Error: parse failed at <path>: <message>`
- `*core.ValidationError` → render: `Error: validation failed: <message>` followed by per-error list
- `*core.TransformError` → render: `Error: transform failed: <message>`
- `*core.FileError` → render: `Error: <op> <path>: <message>`
- `*core.ConfigError` → render: `Error: config: <message>`
- `*core.PartialSuccessError` → render: `partial success: N succeeded, M failed` followed by per-error lines (**partial-success supersedes not-found for wrapped chains**; placed before NotFound in the canonical order)
- `*core.NotFoundError` → render: `Error: not found: <key>`
- `*core.OperationError` → render: `Error: <op> failed: <resource>: <message>`
- `*core.InitializeError` → render: `Error: <e.Error()>` (delegates to `InitializeError.Error()` which renders `initialize failed: <ref>: output: <outputPath>: <cause.Error()>` as a single colon-joined line per `internal/core/errors.go:670-687` — parts joined by `: `; `output` segment optional; cause optional; optional `(context)` suffix; optional `\n💡 <suggestions>` block)
- `*core.UsageError` → render: `Error: <flag>: <reason>`
- `*config.WriteError` → render: `Error: <op> <path>: <message>`
- `*core.CobraUsageError` → render: `Error: <e.Error()>` (delegates to the wrapped cause's `Error()`, e.g., `"Error: requires at least 1 arg(s), only received 0"`)
- generic error → render: `Error: <err.Error()>`

The dispatch ordering matches the order above (full canonical order: Parse, Validation, Transform, File, Config, PartialSuccess, NotFound, Operation, Initialize, Usage, WriteError, CobraUsage, default). `PartialSuccess` intentionally precedes `NotFound` so that a `*core.PartialSuccessError` aggregated from per-resource `NotFoundError` failures dispatches to the partial-success renderer (which lists every per-resource line) rather than to the terse `not found:` renderer. The new `UsageError`, `WriteError`, and `CobraUsageError` cases sit after `OperationError` and before the generic-error fallback.

#### Scenario: FormatError table-driven

- **WHEN** `FormatError` is called with each error type from the list above
- **THEN** it SHALL write the corresponding formatted string to `io.ErrOut`
- **AND** `io.Styles.Error()` SHALL style the prefix

#### Scenario: PartialSuccessError precedes NotFoundError in dispatch order

- **WHEN** `output.FormatError(io, err)` is called with `errors.Join(PartialSuccessError, NotFoundError)` (or any other combination where both `*core.PartialSuccessError` and `*core.NotFoundError` are reachable via `errors.As`)
- **THEN** the rendered output SHALL be the PartialSuccessError shape (`partial success: <N> succeeded, <M> failed\n…`), not the `Error: not found: <key>` shape
- **AND** the canonical switch order at `internal/output/errors.go` SHALL keep `case *core.PartialSuccessError` before `case *core.NotFoundError`
- **NOTE**: this scenario codifies the "PartialSuccess supersedes NotFound" dispatch order documented in the Requirement above; tests that consume `errors.Join` of both types assert the partial-success renderer wins.

#### Scenario: InitializeError renders to stderr

- **WHEN** `output.FormatError(io, err)` is called and `err` is (or wraps) a `*core.InitializeError`
- **THEN** the message SHALL be written to `io.ErrOut`
- **AND** the message SHALL be `io.Styles.Error("Error: ") + e.Error() + "\n"` (delegates to `InitializeError.Error()`; the dispatch case MUST NOT manually concatenate the segments — `e.Error()` is the single source of truth per `internal/core/errors.go`)

#### Scenario: InitializeError with cause and outputPath renders the chain

- **WHEN** `*core.InitializeError` carries a non-nil `Cause` and a non-empty `OutputPath()`
- **THEN** the rendered message SHALL be `io.Styles.Error("Error: ") + e.Error() + "\n"` — delegating to `InitializeError.Error()` (which renders `initialize failed: <ref>: output: <outputPath>: <cause.Error()` as a single colon-joined line per `InitializeError.Error()` declaration in `internal/core/errors.go`). The dispatch case MUST NOT manually concatenate the segments; `e.Error()` is the single source of truth.

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

The system SHALL include the wrapped cause in the output for debugging, indented on a separate line.

#### Scenario: Include cause for debugging

- **WHEN** a typed error wraps an underlying error via `fmt.Errorf("...: %w", inner)`
- **THEN** `FormatError` SHALL append a clearly separated cause line
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

### Requirement: InitializeError dispatch table member

`output.FormatError` SHALL dispatch on `*core.InitializeError` and render the user-facing message to stderr. The pre-change switch missed this case; the case is added in `internal/output/errors.go` (the switch spans lines 31-50; the new `case *core.InitializeError` arm sits after `OperationError` and before the generic-error fallback).

#### Scenario: InitializeError renders to stderr via dispatch

- **WHEN** `output.FormatError(io, err)` is called with `*core.InitializeError`
- **THEN** the dispatch SHALL match the `case *core.InitializeError` arm
- **AND** no other case shall match first (the case order in the switch is: Parse, Validation, Transform, File, Config, PartialSuccess, NotFound, Operation, **Initialize**, Usage, WriteError, default)

### Requirement: UsageError dispatch table member

`output.FormatError` SHALL dispatch on `*core.UsageError` and render the user-facing message to stderr.

#### Scenario: UsageError renders to stderr via dispatch

- **WHEN** `output.FormatError(io, err)` is called with `*core.UsageError`
- **THEN** the dispatch SHALL match the `case *core.UsageError` arm
- **AND** the rendered message SHALL be `"Error: <flag>: <reason>\n"` (where `<flag>` and `<reason>` come from the typed-error's accessors)

### Requirement: CobraUsageError dispatch table member

`output.FormatError` SHALL dispatch on `*core.CobraUsageError` and render the wrapped cause's user-facing message to stderr. The case is added so the dispatch set covers every `core.*Error` type as the implementation comment claims (`internal/output/errors.go:19-22`).

#### Scenario: CobraUsageError renders to stderr via dispatch

- **WHEN** `output.FormatError(io, err)` is called with `*core.CobraUsageError`
- **THEN** the dispatch SHALL match the `case *core.CobraUsageError` arm
- **AND** the rendered message SHALL be `"Error: <wrapped.Error()>\n"` — the wrapped cause's `Error()` is the body; the dispatch delegates to the cause rather than prefixing or wrapping further.

### Requirement: core.NotFoundError type and constructor

`internal/core/errors.go` SHALL define a `NotFoundError` struct, a constructor, and an `Error()` method that produces the canonical not-found message.

#### Scenario: NotFoundError.Error format

- **WHEN** `core.NewNotFoundError("library ref", "nonexistent-ref")` is called
- **THEN** the returned error's `Error()` method SHALL return `"not found: nonexistent-ref"`

#### Scenario: NotFoundError fields

- **WHEN** a `*core.NotFoundError` is constructed with `NewNotFoundError(entity, key)`
- **THEN** the struct SHALL expose the `Entity` and `Key` fields for programmatic inspection (via accessor methods or exported fields)

## Fulfilled

**Change:** `migrate-library-rest` (slice 7 of 9)
**Date:** 2026-07-01

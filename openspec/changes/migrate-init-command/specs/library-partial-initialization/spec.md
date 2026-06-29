# library-partial-initialization Specification (delta)

## REMOVED Requirements

### Requirement: Error return only when all resources fail (removed)

The previous contract required `Initialize` to return `nil` error on partial success and a non-nil error on full failure. The new contract reverses this: `Initialize` always returns the full result slice; per-resource errors live in `core.InitializeResult.Error`; the caller (`runInit`) synthesizes `*core.PartialSuccessError` based on the success count.

#### Scenario: Removed — partial success returns nil error (no longer applies)

The old scenario "Return nil error on partial success" is superseded by the new contract below.

#### Scenario: Removed — non-nil error on all-failure (no longer applies)

The old scenario "Return error when all resources fail" is superseded by the new contract below.

## ADDED Requirements

### Requirement: Initializer.Initialize contract

The `Initializer.Initialize(ctx, req) ([]core.InitializeResult, error)` method SHALL always return the full list of per-resource results, even on partial success or full failure. The `error` return is reserved for transport-level failures (e.g. library not found); per-resource failures are encoded in `core.InitializeResult.Error`.

#### Scenario: All success

- **WHEN** `Initialize` processes N refs and all succeed
- **THEN** it SHALL return `([]result{N items, all Succeeded: true}, nil)`

#### Scenario: Partial success

- **WHEN** `Initialize` processes N refs and M fail
- **THEN** it SHALL return `([]result{N items, M with Succeeded: false and Error: ...}, nil)` — the error return is `nil`; per-resource failures are in the result slice

#### Scenario: Transport failure

- **WHEN** the library cannot be loaded
- **THEN** it SHALL return `(nil, err)` — the result slice is nil and the error is non-nil

### Requirement: core.InitializeResult

Each `core.InitializeResult` SHALL carry: `Ref string`, `InputPath string`, `OutputPath string`, `Error error`. Success is implied by `Error == nil`; there is no separate `Succeeded` field.

#### Scenario: InitializeResult fields

- **WHEN** an `InitializeResult` is inspected
- **THEN** it SHALL have the four fields above
- **AND** `Error == nil` SHALL indicate a successful initialization
- **AND** `Error != nil` SHALL indicate a failed initialization

### Requirement: core.InitializeError wraps the cause

`core.InitializeError` SHALL carry a `Cause error` field and SHALL implement `Unwrap() error` returning the cause so `errors.As(err, &typedErr)` reaches the underlying typed error.

#### Scenario: Unwrap chain reachable

- **WHEN** `core.InitializeError` wraps a typed error
- **THEN** `errors.As` SHALL reach the wrapped cause
- **AND** `core.PartialSuccessError.Errors()` SHALL yield `core.InitializeError` values consumable by `output.FormatError`

### Requirement: Caller distinguishes partial vs full failure

The caller (`runInit`) SHALL distinguish partial success from full failure by inspecting the count of `Succeeded == true` results and synthesizing the appropriate `*core.PartialSuccessError`.

#### Scenario: Partial → exit 0

- **WHEN** the result slice has at least one `Succeeded: true` entry
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: <count>, Failed: <count>}`
- **AND** `cmdutil.ExitCodeFor` SHALL return 0

#### Scenario: Full failure → exit 1

- **WHEN** the result slice has zero `Succeeded: true` entries
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: 0, Failed: <count>}`
- **AND** `cmdutil.ExitCodeFor` SHALL return 1

### Requirement: Preset-not-found reported as usage error

When `--preset <name>` references a non-existent preset, `runInit` SHALL return `*core.NotFoundError{Entity: "preset", Key: <name>}`. The `cmdutil.ExitCodeFor` mapping (added as preliminary task 5.0.1 in this change) returns `ExitCodeUsage` (2) for `*core.NotFoundError`.

#### Scenario: Preset not found → exit 2

- **GIVEN** no preset named `ghost` in the library
- **WHEN** `germinator init --platform opencode --preset ghost` is run
- **THEN** `runInit` SHALL return `*core.NotFoundError{Entity: "preset", Key: "ghost"}`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeUsage` (2)

> **Status:** the new `Initialize` contract is implemented in this change (`migrate-init-command`). `core.InitializeResult`, `core.InitializeError`, and `core.PartialSuccessError` are defined in the `scaffold-cli-foundation` change. The `(*library.Library).ResolvePreset` method, the `cmdutil.Factory.Initializer` field, and the `cmdutil.ExitCodeFor` mapping for `*core.NotFoundError` → exit 2 are added as preliminary code-change tasks in `tasks.md` §5.0.

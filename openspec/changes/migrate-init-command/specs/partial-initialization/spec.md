# partial-initialization Specification (delta)

## MODIFIED Requirements

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

Each `core.InitializeResult` SHALL carry: `Ref string`, `InputPath string`, `OutputPath string`, `Succeeded bool`, `Error error`.

#### Scenario: InitializeResult fields

- **WHEN** an `InitializeResult` is inspected
- **THEN** it SHALL have the five fields above
- **AND** `Succeeded == true` ⇒ `Error == nil`
- **AND** `Succeeded == false` ⇒ `Error != nil`

### Requirement: Caller distinguishes partial vs full failure

The caller (`runInit`) SHALL distinguish partial success from full failure by inspecting the count of `Succeeded == true` results.

#### Scenario: Partial → exit 0

- **WHEN** the result slice has at least one `Succeeded: true` entry
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: <count>, Failed: <count>}`
- **AND** `cmdutil.ExitCodeFor` SHALL return 0

#### Scenario: Full failure → exit 1

- **WHEN** the result slice has zero `Succeeded: true` entries
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: 0, Failed: <count>}`
- **AND** `cmdutil.ExitCodeFor` SHALL return 1

> **Status:** the new `Initialize` contract is implemented in change-5 (`migrate-init-command`). The `core.InitializeResult` and `core.InitializeError` types are defined in change-1 (`scaffold-cli-foundation`).

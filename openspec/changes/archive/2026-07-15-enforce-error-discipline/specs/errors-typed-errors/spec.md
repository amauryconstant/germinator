# typed-errors Specification (delta)

> **Cross-references:** this change also modifies `cli-error-formatting`, `cli-exit-codes`, `errors-enhanced-errors`. See those delta specs.

## REMOVED Requirements

### Scenario: NotFoundError maps to ExitCodeUsage

The prior scenario *"NotFoundError maps to ExitCodeUsage"* (from `openspec/specs/errors-typed-errors/spec.md:137-140`) is removed. `*core.NotFoundError` no longer maps to exit code 2; the corrected mapping lives in `cli-exit-codes/spec.md`.

> **Why:** the prior mapping was semantically wrong per the 2026-07-08 review and the `golang-cli-architecture/references/05-errors.md` rule that "ValidationError is a runtime failure → exit 1, NOT a CLI usage error." A lookup miss is a runtime state, not a user-input validation error. The prior parenthetical rationale ("the lookup miss is treated as user input that did not resolve, not an operational error") is now considered incorrect: a not-found error is not user input, it is a system state the user is querying.

## ADDED Requirements

### Requirement: UsageError type

The system SHALL provide a `UsageError` type for CLI flag validation errors that are not caught by Cobra's `MarkFlagRequired`, `Args` validators, or pflag's typed errors. `UsageError` carries the flag name, the reason, and optional suggestions as private fields exposed via accessor methods (`Flag()`, `Reason()`, `Suggestions()`) following the project's builder pattern (matching `ParseError`, `ValidationError`, `TransformError`, `FileError`, `ConfigError`, `InitializeError`). The immutable `WithSuggestions([]string) *UsageError` builder returns a new instance with the same `flag` and `reason` and the new suggestions; suggestions are reserved for future flag-validation errors that may want to suggest valid alternatives (e.g., mis-spelled flag names).

`UsageError` maps to `ExitCodeUsage` (2) via `cmdutil.ExitCodeFor`. `Unwrap()` returns `nil` — `UsageError` is a leaf error (it does not wrap an underlying cause); the godoc explicitly notes this so future maintainers do not add a `cause` field and break the contract.

**Extension rationale**: the canonical `UsageError` shape in `golang-cli-architecture/references/05-errors.md` line 122-128 uses a single `Message string` field. This change extends it to carry `flag`, `reason`, and `suggestions` (with private fields + getters + immutable builder) because pflag-style error text does not preserve which flag failed, and the suggestion list is reserved for future flag-validation errors that want to suggest valid alternatives (e.g., mis-spelled flag names). The `{flag, reason, suggestions}` shape is the canonical extension for CLI-layer usage errors that need to preserve positional context.

#### Scenario: UsageError has private Flag, Reason, and Suggestions fields

- **WHEN** `NewUsageError(flag, reason string)` is called
- **THEN** it SHALL return a `*UsageError` with private `flag`, `reason`, and (initially nil) `suggestions` fields populated
- **AND** the fields SHALL NOT be directly accessible from outside `internal/core`
- **AND** `e.Flag()` SHALL return the flag
- **AND** `e.Reason()` SHALL return the reason
- **AND** `e.Suggestions()` SHALL return a defensive copy of the suggestions slice (or nil if no suggestions are set)

#### Scenario: UsageError Error format

- **WHEN** `err.Error()` is called on a `*UsageError{flag: "--resources", reason: "must be non-empty list of refs"}`
- **THEN** it SHALL return the string `"--resources: must be non-empty list of refs"`

#### Scenario: UsageError follows Go error-string convention

- **WHEN** a `*UsageError` is constructed via `NewUsageError(flag, reason)` for any input
- **THEN** the rendered `Error()` SHALL be a single line in the form `<flag>: <reason>`, where `<flag>` starts with `--` and the rest of the flag segment is lowercase kebab-case (e.g. `--resources`), and `<reason>` starts with a lowercase letter, contains no trailing `.`, `!`, or `?`
- **AND** the godoc for `*UsageError` SHALL explicitly state the convention: "the `flag` parameter MUST match `^--[a-z][a-z0-9-]*$`; the `reason` parameter MUST start with a lowercase letter and have no trailing punctuation (Go error-string convention per `golang-error-handling` rule 3 — `references/error-creation.md:32`)". The constructor does NOT validate the convention at runtime; callers passing upper-case or trailing punctuation violate the contract and are a programmer error.
- **AND** the convention is enforced via godoc only; no test asserts upper-case or trailing-punctuation rejection (such an assertion would require either rejecting at the constructor or normalizing the input, neither of which is desired — the typed-error surface must round-trip the supplied bytes verbatim).

#### Scenario: UsageError WithSuggestions builder returns a new instance

- **WHEN** `e.WithSuggestions([]string{"hint1", "hint2"})` is called on a `*UsageError`
- **THEN** the returned `*UsageError` SHALL have the same `flag` and `reason` as `e`
- **AND** the returned `*UsageError`'s `Suggestions()` SHALL return `[]string{"hint1", "hint2"}`
- **AND** the original `e` SHALL NOT be modified (immutable builder)

#### Scenario: UsageError is in FormatError dispatch set

- **WHEN** `output.FormatError(io, err)` is called with `*core.UsageError`
- **THEN** the rendered message SHALL be `"Error: --resources: must be non-empty list of refs"` written to `io.ErrOut`

#### Scenario: UsageError implements json.Marshaler

- **WHEN** `json.Marshal(*core.UsageError{...})` is called
- **THEN** it SHALL return the JSON bytes `{"error": "--resources: must be non-empty list of refs"}`

### Requirement: CobraUsageError sentinel

The system SHALL provide a `CobraUsageError` sentinel that wraps the underlying Cobra arg-validation error. Commands wrap the error returned by `cobra.ExactArgs`/`MinimumNArgs`/`MaximumNArgs`/`RangeArgs` failures (currently emitted as `fmt.Errorf` strings via `cobra/args.go`) with `MustNewCobraUsageError(err)` so `cmdutil.ExitCodeFor` can match the typed error and return `ExitCodeUsage` (2). The `Must*` prefix telegraphs the panic on nil cause (a violated invariant per `golang-design-patterns`); mirroring `regexp.MustCompile` and `template.Must`.

`CobraUsageError` is the project-owned contract that replaces the brittle substring-prefix dispatch; pflag's four typed errors (`*pflag.NotExistError`, `*pflag.ValueRequiredError`, `*pflag.InvalidValueError`, `*pflag.InvalidSyntaxError`) already cover pflag-emitted types directly.

#### Scenario: CobraUsageError wraps an existing error

- **WHEN** `MustNewCobraUsageError(err)` is called with a non-nil `err`
- **THEN** the returned `*CobraUsageError` SHALL expose the underlying error via `Unwrap()`
- **AND** the constructor SHALL NOT panic (the cause is non-nil)

#### Scenario: CobraUsageError Error format

- **WHEN** `err.Error()` is called on a `*CobraUsageError` wrapping `errors.New("requires at least 1 arg(s), only received 0")`
- **THEN** it SHALL return the wrapped error's `Error()` string verbatim

#### Scenario: MustNewCobraUsageError panics on nil cause

- **WHEN** `MustNewCobraUsageError(nil)` is called
- **THEN** it SHALL panic with a message indicating the cause is required — a nil cause is a programmer error, not a recoverable state. The `Must*` constructor prefix telegraphs the panic to callers (matching `regexp.MustCompile` and `template.Must`). No nil-guard fallback is provided; if a caller requires nil-safety, they must use a try/recv pattern or a separate `New*` constructor (not provided in this change).

### Requirement: ValidateDocumentType helper

The system SHALL provide a `ValidateDocumentType(docType string) error` helper in `internal/core/rules.go` that validates a bare document type against the canonical resource-type set `{skill, agent, command, memory}` (the same set `validResourceTypes` constrains `CanInstallResource` to). This helper is the canonical guardrail for command-line `--type` validation (e.g., `germinator canonicalize --type <docType>`).

`CanInstallResource(ref)` validates the `"type/name"` ref shape and is the WRONG guardrail for bare document types; `ValidateDocumentType` is the new sibling helper for the bare-type case.

#### Scenario: ValidateDocumentType accepts canonical types

- **WHEN** `core.ValidateDocumentType("agent")` is called (and similarly for `"command"`, `"skill"`, `"memory"`)
- **THEN** it SHALL return nil

#### Scenario: ValidateDocumentType rejects the plural form

- **WHEN** `core.ValidateDocumentType("skills")` is called (the plural form)
- **THEN** it SHALL return a `*core.ValidationError`
- **AND** the error SHALL include a suggestion listing the canonical types

#### Scenario: ValidateDocumentType rejects unknown / empty input

- **WHEN** `core.ValidateDocumentType("bot")` or `core.ValidateDocumentType("")` is called
- **THEN** it SHALL return a `*core.ValidationError`

### Requirement: MarshalJSON on all core typed errors

All typed errors defined in `internal/core/errors.go` SHALL implement `MarshalJSON() ([]byte, error)`. The complete set is:

- **Existing (9)**: `ParseError`, `ValidationError`, `TransformError`, `FileError`, `ConfigError`, `NotFoundError`, `OperationError`, `InitializeError`, `PartialSuccessError`.
- **New (2)**: `UsageError`, `CobraUsageError`.

Each `MarshalJSON()` SHALL return the JSON bytes `{"error": "<Error()>"}`.

#### Scenario: MarshalJSON returns structured JSON

- **WHEN** `json.Marshal(*core.NotFoundError{Key: "ghost"})` is called
- **THEN** it SHALL return the JSON bytes `{"error": "not found: ghost"}`
- **AND** the underlying error's `Error()` string SHALL be the value of the `error` field

#### Scenario: MarshalJSON for wrapped typed errors

- **WHEN** `json.Marshal(&core.OperationError{Op: "add", Resource: "skill/commit", Cause: ...})` is called (the pointer-literal form is required because `MarshalJSON` is defined on the pointer receiver per task 1.7 — `encoding/json` does not auto-take-address on non-addressable value literals)
- **THEN** the rendered JSON SHALL be `{"error": "add: skill/commit"}` (the single-key shape, delegating to `e.Error()`)
- **AND** cause chains SHALL NOT be re-encoded recursively (the standard `Cause` chain is exposed via `Error()`, not as a nested object). `BatchFailureInfo.Cause` (see `errors-enhanced-errors/spec.md`) wraps a typed error whose own `MarshalJSON` returns the single-key shape; that single-key shape is what `BatchFailureInfo` JSON serialization renders, not a recursive walk of the cause chain.

#### Scenario: MarshalJSON contract for all 11 typed errors

- **WHEN** `json.Marshal(e)` is called for each of the 11 typed errors (ParseError, ValidationError, TransformError, FileError, ConfigError, NotFoundError, OperationError, InitializeError, PartialSuccessError, UsageError, CobraUsageError)
- **THEN** the rendered JSON SHALL be `{"error": "<e.Error()>"}`
- **AND** the underlying error's `Error()` string SHALL be the value of the `error` field

## MODIFIED Requirements

### Requirement: NotFoundError type

The system SHALL provide a `NotFoundError` type for missing-entity lookups (library refs, presets, library.yaml, source files, library files). It carries `Entity` and `Key` as exported fields and maps to exit code 1 (operational error) via `cmdutil.ExitCodeFor`.

**Change**: clarify that `NotFoundError` maps to `ExitCodeError` (1), not `ExitCodeUsage` (2). The prior mapping (2) was semantically wrong per the 2026-07-08 review: "not found" is a runtime state, not a user-input validation error. The change `enforce-error-discipline` updates `internal/cmdutil/exit.go:73-74` and `internal/cmdutil/exit_test.go:58` accordingly.

The migration also widens the type swap from `cmd/library_add.go` lookup branches to all 9 production sites: `internal/library/resolver.go:21, 26, 62`, `internal/library/loader.go:36, 53`, `internal/library/adder.go:157`, `internal/library/remover.go:82, 87, 140`.

#### Scenario: NewNotFoundError constructor

- **WHEN** `NewNotFoundError(entity, key string)` is called
- **THEN** it SHALL return a `*NotFoundError{Entity: entity, Key: key}`

#### Scenario: NotFoundError Error format

- **WHEN** `err.Error()` is called on a `*NotFoundError{Key: "ghost"}`
- **THEN** it SHALL return the string `"not found: ghost"`

#### Scenario: NotFoundError maps to ExitCodeError

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeError` (1) — `*core.NotFoundError` represents a runtime lookup miss, not a user-input validation error; the prior mapping (2) was semantically wrong and is corrected in change `enforce-error-discipline`.

#### Scenario: NotFoundError implements json.Marshaler

- **WHEN** `json.Marshal(*core.NotFoundError{Key: "ghost"})` is called
- **THEN** it SHALL return the JSON bytes `{"error": "not found: ghost"}`

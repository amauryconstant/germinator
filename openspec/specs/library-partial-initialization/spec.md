# Capability: Partial Initialization

## Purpose

The Partial Initialization capability extends resource installation to continue processing all resources even when individual failures occur. It provides comprehensive result reporting so users can see all successes and failures in a single execution.

## Requirements

### Requirement: Continue on individual resource failure

The system SHALL process all resources in a request regardless of individual failures.

#### Scenario: Continue processing after resource resolution failure
- **GIVEN** resources `["skill/commit", "skill/nonexistent", "skill/merge-request"]`
- **WHEN** Initialize is called
- **THEN** `skill/commit` is processed successfully
- **AND** `skill/nonexistent` fails with resolution error
- **AND** `skill/merge-request` is processed successfully
- **AND** all three results are returned

#### Scenario: Continue processing after file write failure
- **GIVEN** resources where one has invalid content that fails rendering
- **WHEN** Initialize is called
- **THEN** resources before the failure are written successfully
- **AND** the failing resource has an error in its result
- **AND** resources after the failure are still attempted

#### Scenario: Continue processing after file exists error
- **GIVEN** resources `["skill/commit", "skill/existing", "skill/merge-request"]`
- **AND** `skill/existing` maps to a file that already exists without --force
- **WHEN** Initialize is called
- **THEN** `skill/commit` is processed successfully
- **AND** `skill/existing` fails with file exists error
- **AND** `skill/merge-request` is still processed

### Requirement: Return partial results

The system SHALL return results for all resources, including failures.

#### Scenario: Return success result for successful resource
- **GIVEN** a valid resource `skill/commit`
- **WHEN** Initialize is called
- **THEN** the result for `skill/commit` has Ref, InputPath, OutputPath set
- **AND** Error is nil

#### Scenario: Return error result for failed resource
- **GIVEN** an invalid resource `skill/nonexistent`
- **WHEN** Initialize is called
- **THEN** the result for `skill/nonexistent` has Ref set
- **AND** Error contains the resolution error
- **AND** InputPath and OutputPath may be empty

#### Scenario: Mixed results returned
- **GIVEN** three resources where the second fails
- **WHEN** Initialize is called
- **THEN** the returned slice contains three results
- **AND** first result has no error
- **AND** second result has an error
- **AND** third result has no error

### Requirement: Initializer.Initialize contract

The `Initializer.Initialize(ctx, req) ([]core.InitializeResult, error)` method SHALL always return the full list of per-resource results, even on partial success or full failure. The `error` return is reserved for transport-level failures (e.g. library not found); per-resource failures are encoded in `core.InitializeResult.Error`.

#### Scenario: All success
- **WHEN** `Initialize` processes N refs and all succeed
- **THEN** it SHALL return `([]result{N items, all with Error: nil}, nil)`

#### Scenario: Partial success
- **WHEN** `Initialize` processes N refs and M fail
- **THEN** it SHALL return `([]result{N items, M with non-nil Error}, nil)` — the error return is `nil`; per-resource failures are encoded in the `Error` field of the result slice

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

The caller (`runInit`) SHALL distinguish partial success from full failure by inspecting the count of results with `Error == nil` and synthesizing the appropriate `*core.PartialSuccessError`.

#### Scenario: Partial → exit 0
- **WHEN** the result slice has at least one entry with `Error == nil`
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: <count>, Failed: <count>}`
- **AND** `cmdutil.ExitCodeFor` SHALL return 0

#### Scenario: Full failure → exit 1
- **WHEN** the result slice has zero entries with `Error == nil`
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: 0, Failed: <count>}`
- **AND** `cmdutil.ExitCodeFor` SHALL return 1

### Requirement: Preset-not-found reported as usage error

When `--preset <name>` references a non-existent preset, `runInit` SHALL return `*core.NotFoundError{Entity: "preset", Key: <name>}`. The `cmdutil.ExitCodeFor` mapping returns `ExitCodeUsage` (2) for `*core.NotFoundError`.

#### Scenario: Preset not found → exit 2
- **GIVEN** no preset named `ghost` in the library
- **WHEN** `germinator init --platform opencode --preset ghost` is run
- **THEN** `runInit` SHALL return `*core.NotFoundError{Entity: "preset", Key: "ghost"}`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeUsage` (2)

### Requirement: Support dry-run with partial processing

The system SHALL preview all resources in dry-run mode regardless of errors.

#### Scenario: Dry-run continues through resolution errors
- **GIVEN** resources with one invalid ref
- **WHEN** Initialize is called with `--dry-run`
- **THEN** all resources are evaluated
- **AND** valid resources show their would-be output paths
- **AND** invalid resource shows its error

### Requirement: Support force flag with partial processing

The system SHALL apply force flag to all resources during partial processing.

#### Scenario: Force applies to all resources
- **GIVEN** resources where some would overwrite existing files
- **WHEN** Initialize is called with `--force`
- **THEN** all resources are processed
- **AND** existing files are overwritten
- **AND** no "file exists" errors are returned

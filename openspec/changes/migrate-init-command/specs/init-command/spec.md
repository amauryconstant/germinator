# init-command Specification (delta)

## MODIFIED Requirements

### Requirement: init command follows command-options-pattern

The `init` command SHALL adopt the `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command` + `runInit(opts *initOptions) error` template.

#### Scenario: init command signature

- **WHEN** `germinator init --help` is invoked
- **THEN** the help output SHALL list the flags: `--platform`, `--output-dir`, `--resources`, `--preset`, `--dry-run`, `--force`
- **AND** the constructor signature SHALL be `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`

#### Scenario: initOptions struct

- **WHEN** `cmd/init.go` is inspected
- **THEN** it SHALL declare `initOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Initializer func() (Initializer, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`

### Requirement: Partial-success exit code semantics

The `init` command SHALL return exit code 0 if at least one resource was processed successfully, and exit code 1 only if all resources failed.

#### Scenario: All succeeded

- **WHEN** `runInit` processes N refs and all succeed
- **THEN** `runInit` SHALL return `nil`
- **AND** `cmdutil.ExitCodeFor(nil)` SHALL return `ExitCodeSuccess` (0)

#### Scenario: Partial success

- **WHEN** `runInit` processes N refs and M (1 ≤ M < N) fail
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: N - M, Failed: M}`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeSuccess` (0)
- **AND** `output.FormatError(io, err)` SHALL write `partial success: N-M succeeded, M failed` followed by per-resource error lines

#### Scenario: All failed

- **WHEN** `runInit` processes N refs and all fail
- **THEN** `runInit` SHALL return `*core.PartialSuccessError{Succeeded: 0, Failed: N}`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeError` (1)

### Requirement: Per-resource errors rendered via FormatError

When `runInit` encounters per-resource errors, each error SHALL be rendered via `output.FormatError(io, err)` to `io.ErrOut`.

#### Scenario: Per-resource error rendering

- **WHEN** a single resource fails during init
- **THEN** the per-resource error SHALL be formatted via `output.FormatError`
- **AND** the error SHALL appear in `opts.IO.ErrOut`
- **AND** the overall command exit code SHALL reflect partial-success semantics

### Requirement: Preset expansion

When `init --preset <name>` is invoked, the preset SHALL be expanded to a list of refs before processing.

#### Scenario: Preset expansion

- **WHEN** `germinator init --preset git-workflow` is invoked
- **THEN** `runInit` SHALL call `f.Library().ResolvePreset(ctx, "git-workflow")` to get the list of refs
- **AND** each ref SHALL be processed in turn
- **AND** partial-success semantics SHALL apply to the full list of expanded refs

> **Status:** the `init` command is migrated in change-5 (`migrate-init-command`). The `core.PartialSuccessError` type was defined in change-1 (`scaffold-cli-foundation`); `cmdutil.ExitCodeFor` and `output.FormatError` recognize it from day one.

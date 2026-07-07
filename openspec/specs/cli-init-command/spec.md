# Capability: Init Command

## Purpose

The Init Command provides the CLI entry point for initializing AI coding assistant resources in a target project. It handles flag parsing, validation, and orchestrates the resource installation process.

## Requirements

### Requirement: Install with explicit resources

The CLI SHALL install explicitly specified resources.

#### Scenario: Install with explicit resources
- **GIVEN** a library with resources `skill/commit` and `skill/merge-request`
- **WHEN** `germinator init --platform opencode --resources skill/commit,skill/merge-request` is run
- **THEN** both resources are installed to the current directory

### Requirement: Install with preset

The CLI SHALL install all resources from a preset.

#### Scenario: Install with preset
- **GIVEN** a library with preset `git-workflow`
- **WHEN** `germinator init --platform opencode --preset git-workflow` is run
- **THEN** all resources in the preset are installed

### Requirement: init command follows command-options-pattern

The `init` command SHALL adopt the `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command` + `runInit(opts *initOptions) error` template.

#### Scenario: init command signature
- **WHEN** `germinator init --help` is invoked
- **THEN** the help output SHALL list the flags: `--platform`, `--output-dir`, `--library`, `--resources`, `--preset`, `--dry-run`, `--force`
- **AND** the constructor signature SHALL be `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`
#### Scenario: initOptions struct

- **WHEN** `cmd/init.go` is inspected
- **THEN** it SHALL declare `initOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`
- **AND** `runInit` SHALL obtain the initializer via the direct constructor `NewInitializer()` from `cmd/initializer.go` (no per-command lazy field; the initializer is an internal adapter that needs no test seam because it has no injectable dependency)

### Requirement: Validate required flags

The `init` command SHALL require `--platform` and exactly one of `--resources` or `--preset`.

#### Scenario: Require --platform
- **GIVEN** no `--platform` flag
- **WHEN** `germinator init --resources skill/commit` is run
- **THEN** an error indicates `--platform` is required

#### Scenario: Require --resources or --preset
- **GIVEN** no `--resources` or `--preset` flag
- **WHEN** `germinator init --platform opencode` is run
- **THEN** an error indicates either `--resources` or `--preset` is required

#### Scenario: Reject both --resources and --preset
- **GIVEN** both `--resources` and `--preset` flags
- **WHEN** `germinator init --platform opencode --resources skill/commit --preset git-workflow` is run
- **THEN** an error indicates flags are mutually exclusive

### Requirement: Support --library flag

The `init` command SHALL accept a custom library path via `--library`, populated into `opts.LibraryPath`.

#### Scenario: Custom library path
- **GIVEN** a library at `/custom/library`
- **WHEN** `germinator init --platform opencode --resources skill/commit --library /custom/library` is run
- **THEN** resources are loaded from the custom library
### Requirement: --output-dir flag

The flag for the output directory SHALL be `--output-dir`. There is no short form. The default value is `"."` (the current working directory); when omitted the CLI writes files relative to the invocation directory.

#### Scenario: Custom output directory

- **GIVEN** output directory `/target/project`
- **WHEN** `germinator init --platform opencode --resources skill/commit --output-dir /target/project` is run
- **THEN** resources are installed under `/target/project/...`

### Requirement: Validate platform value

The `init` command SHALL validate the platform flag value via `core.ValidatePlatform`.

#### Scenario: Validate platform value
- **GIVEN** an invalid platform `invalid-platform`
- **WHEN** `germinator init --platform invalid-platform --resources skill/commit` is run
- **THEN** an error indicates valid platforms are `opencode` and `claude-code`

### Requirement: Support dry-run preview

The `init` command SHALL support `--dry-run` to preview changes without writing files.

#### Scenario: Dry-run preview
- **GIVEN** dry-run mode
- **WHEN** `germinator init --platform opencode --resources skill/commit --dry-run` is run
- **THEN** output shows what would be written without creating files
### Requirement: Support --force overwrite

The `init` command SHALL support `--force` to overwrite existing files. The destructive-operation resolution order SHALL be:

1. `--force` flag set → overwrite without prompting, exit 0 on success
2. stdin is a TTY (`IOStreams.IsInteractive()` returns true) → print a confirmation prompt; abort if the user declines
3. stdin is not a TTY (pipe, CI, redirect) AND `--force` not set → return `*core.NewFileError(...)` indicating the file exists and `--force` is required (non-interactive, no prompt)

#### Scenario: Force overwrite

- **GIVEN** existing output files
- **WHEN** `germinator init --platform opencode --resources skill/commit --force` is run
- **THEN** existing files are overwritten without prompting

#### Scenario: Non-interactive overwrite without --force fails

- **GIVEN** existing output files
- **AND** stdin is not a TTY (e.g., piped from another command)
- **WHEN** `germinator init --platform opencode --resources skill/commit` is run without `--force`
- **THEN** the command SHALL return a `*core.FileError` explaining the file exists and `--force` is required (no prompt in non-interactive context)

#### Scenario: Interactive overwrite without --force prompts

- **GIVEN** existing output files
- **AND** stdin AND stdout are both TTYs
- **WHEN** `germinator init --platform opencode --resources skill/commit` is run without `--force`
- **THEN** the command SHALL print a confirmation prompt before overwriting
- **AND** on user confirmation, the existing files SHALL be overwritten

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

When `runInit` encounters per-resource errors, each error SHALL be rendered via `output.FormatError(io, err)` to `opts.IO.ErrOut`.

#### Scenario: Per-resource error rendering
- **WHEN** a single resource fails during init
- **THEN** the per-resource error SHALL be formatted via `output.FormatError`
- **AND** the error SHALL appear in `opts.IO.ErrOut`
- **AND** the overall command exit code SHALL reflect partial-success semantics

### Requirement: Preset expansion via Library method

When `--preset <name>` is invoked, the preset SHALL be expanded via `(*library.Library).ResolvePreset(ctx, name)`.

#### Scenario: Preset expansion
- **WHEN** `germinator init --platform opencode --preset git-workflow` is invoked
- **THEN** `runInit` SHALL call `f.Library().ResolvePreset(opts.Ctx, "git-workflow")` to get the list of refs
- **AND** each ref SHALL be processed in turn
- **AND** partial-success semantics SHALL apply to the full list of expanded refs

### Requirement: Preset-not-found is a usage error

When `--preset <name>` references a preset that does not exist, the command SHALL return `*core.NotFoundError` and exit 2.

#### Scenario: Preset not found
- **GIVEN** no preset named `ghost` in the library
- **WHEN** `germinator init --platform opencode --preset ghost` is run
- **THEN** `runInit` SHALL return `*core.NotFoundError{Entity: "preset", Key: "ghost"}`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeUsage` (2)
### Requirement: Initializer resolved via direct constructor

The `Initializer` adapter SHALL be obtained in `runInit` via the direct constructor `NewInitializer()` from `cmd/initializer.go`. The adapter has no injectable dependencies (its only inputs are passed via `*InitializeRequest`), so it does not need a Factory lazy field or a per-command `initOptions` field.

#### Scenario: NewInitializer has no test seam

- **WHEN** `cmd/init.go` is inspected
- **THEN** `initOptions` SHALL NOT contain an `Initializer` field
- **AND** `runInit` SHALL call `NewInitializer().Initialize(ctx, req)` directly on the request constructed from validated flags

### Requirement: Format success output

The CLI SHALL display a summary of installed resources on success.

#### Scenario: Success output
- **GIVEN** successful installation of resources
- **WHEN** init completes
- **THEN** a summary lists installed resources and their paths

### Requirement: Format error output

The CLI SHALL display errors with resource reference and file path.

#### Scenario: Error output with file paths
- **GIVEN** a resource that fails to load
- **WHEN** init encounters an error
- **THEN** the error message includes the resource reference and file path

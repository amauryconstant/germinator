# cli-init-command Specification (delta)

## MODIFIED Requirements

### Requirement: init command follows command-options-pattern

The `init` command SHALL adopt the `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command` + `runInit(opts *initOptions) error` template.

#### Scenario: init command signature

- **WHEN** `germinator init --help` is invoked
- **THEN** the help output SHALL list the flags: `--platform`, `--output-dir`, `--library`, `--resources`, `--preset`, `--dry-run`, `--force`
- **AND** the constructor signature SHALL be `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`

#### Scenario: initOptions struct

- **WHEN** `cmd/init.go` is inspected
- **THEN** it SHALL declare `initOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Initializer func() (application.Initializer, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`

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

### Requirement: --output-dir flag (breaking rename)

The flag for the output directory SHALL be `--output-dir` (replacing the legacy `--output`/`-o` short form). This is a breaking change relative to the legacy `cmd/init.go`.

#### Scenario: Custom output directory

- **GIVEN** output directory `/target/project`
- **WHEN** `germinator init --platform opencode --resources skill/commit --output-dir /target/project` is run
- **THEN** resources are installed to `/target/project/.opencode/skills/commit/SKILL.md`

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

The `init` command SHALL support `--force` to overwrite existing files.

#### Scenario: Force overwrite

- **GIVEN** existing output files
- **WHEN** `germinator init --platform opencode --resources skill/commit --force` is run
- **THEN** existing files are overwritten

## ADDED Requirements

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
- **THEN** `runInit` SHALL return `*core.NotFoundError{Entity: "preset", Name: "ghost"}`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeUsage` (2)

### Requirement: Initializer wired through Factory

The `cmdutil.Factory` SHALL expose a lazy `Initializer func() (application.Initializer, error)` field so `runInit` can resolve it without import cycles.

#### Scenario: Factory exposes Initializer

- **WHEN** `cmd/init.go` is inspected
- **THEN** `initOptions.Initializer` SHALL be populated from `f.Initializer` in the RunE hook
- **AND** the field SHALL be lazily evaluated (called only when `runInit` invokes it)

> **Status:** the `init` command is migrated in this change (`migrate-init-command`). Foundation types (`core.PartialSuccessError`, `core.InitializeResult`, `core.InitializeError`) were defined in the `scaffold-cli-foundation` change. The `cmdutil.ExitCodeFor` mapping for `*core.NotFoundError` → exit 2, the `(*library.Library).ResolvePreset` method, and the `Factory.Initializer` field are added as preliminary code-change tasks in `tasks.md` §5.0.

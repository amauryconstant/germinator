# library-library-scaffolding Specification (delta)

## MODIFIED Requirements

### Requirement: library init follows command-options-pattern

The `library init` command SHALL adopt the `NewCmdLibraryInit(f *cmdutil.Factory, runF func(*libraryInitOptions) error) *cobra.Command` + `runLibraryInit(opts *libraryInitOptions) error` template.

#### Scenario: libraryInitOptions struct

- **GIVEN** the `library init` command has been migrated
- **WHEN** `cmd/library_init.go` is inspected
- **THEN** it SHALL declare `libraryInitOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `Path string`, `Force bool`, `DryRun bool`, `Output string`
- **AND** the struct SHALL NOT carry a `Library` lazy loader field — `runLibraryInit` calls `library.CreateLibrary` (a package-level function) directly because `init` creates a fresh library and there is no pre-existing `*library.Library` to receive a method call

#### Scenario: Library interface not declared

- **GIVEN** `cmd/library_init.go` is the only library command where the I/O is creation (not mutation of an existing library)
- **WHEN** `cmd/library_init.go` is inspected
- **THEN** it SHALL NOT declare a `Library` interface in the cmd-side style of slice-6/7 migrations
- **AND** it SHALL directly invoke `library.CreateLibrary(library.CreateOptions{...})` (or the thin wrapper `library.Init(ctx, *InitRequest) error` introduced in this change) inside `runLibraryInit`

### Requirement: library init supports --output flag

The `library init` command SHALL expose a `--output json|table|plain` flag via `cmdutil.AddOutputFlags`.

#### Scenario: Default plain output

- **GIVEN** a path where no library exists
- **WHEN** `germinator library init --path /tmp/my-library` is invoked without `--output`
- **THEN** the output SHALL be plain text confirming the library path created (or the dry-run preview)
- **AND** the output SHALL be written to **stdout** (`opts.IO.Out`) per `internal/output/` stream discipline

#### Scenario: JSON output

- **GIVEN** a path where no library exists
- **WHEN** `germinator library init --path /tmp/my-library --output json` is invoked
- **THEN** the result SHALL be JSON-formatted (`{"path": "...", "dryRun": false, "created": true}`)
- **AND** the JSON SHALL be 2-space indented with a trailing newline (matching `library-library-json-output` stream discipline)
- **AND** the JSON SHALL be written to **stdout**

#### Scenario: Table output

- **GIVEN** a path where no library exists
- **WHEN** `germinator library init --path /tmp/my-library --output table` is invoked
- **THEN** the output SHALL be a table with columns: path, dryRun, created
- **AND** the library path SHALL appear as a single row

#### Scenario: --path flag preserved

- **GIVEN** a user-supplied target path
- **WHEN** `germinator library init --path /tmp/my-library` is invoked
- **THEN** the library SHALL be created at `/tmp/my-library`
- **AND** the existing `--force` and `--dry-run` flags SHALL be preserved (byte-compatible with the legacy CLI)

#### Scenario: Legacy --json flag is rejected

- **WHEN** `germinator library init --json` is invoked
- **THEN** the command SHALL return a usage error
- **AND** the process SHALL exit with code 2 (`ExitCodeUsage`)

### Requirement: --force flag is preserved

The `--force` flag SHALL remain opt-in. When set, the command SHALL overwrite an existing library at the target path.

#### Scenario: --force overwrites existing library

- **GIVEN** a library already exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --force` is invoked
- **THEN** the existing library SHALL be overwritten (existing `library.yaml` and resource files may be modified or removed per the legacy behavior)
- **AND** no confirmation prompt SHALL appear when `--force` is set

#### Scenario: init without --force refuses to overwrite

- **GIVEN** a library already exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library` is invoked (without `--force`)
- **THEN** the command SHALL refuse to overwrite the existing library
- **AND** a typed error SHALL be returned (exit code 1 via `cmdutil.ExitCodeFor`)

### Requirement: --dry-run flag is preserved

The `--dry-run` flag SHALL preview the operation without modifying the filesystem.

#### Scenario: --dry-run previews without changes

- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --dry-run` is invoked
- **THEN** the command SHALL output a preview (plain or JSON per `--output`)
- **AND** the filesystem SHALL NOT be modified (no `library.yaml`, no resource directories)

### Requirement: InitRequest request type

The `library.InitRequest` type SHALL be defined in `internal/library/requests.go` for the thin `library.Init(ctx, *InitRequest) error` wrapper added by this change.

#### Scenario: InitRequest struct

- **GIVEN** the request/result convention established in slice-6 (`library.CreatePresetRequest`)
- **WHEN** `internal/library/requests.go` is inspected
- **THEN** it SHALL declare `InitRequest` with fields `Path string`, `Force bool`, `DryRun bool`
- **AND** the package-level `library.Init(ctx, *InitRequest) error` function SHALL map request fields to `library.CreateOptions{Path, Force, DryRun}` and delegate to `library.CreateLibrary`

> **Status:** the `--output` flag is added to `library init` in change-7 (`migrate-library-rest`). The legacy `--json` flag is replaced by `--output json`. The `--path`, `--force`, and `--dry-run` flags are preserved (no breaking CLI change).

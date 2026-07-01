# Capability: Library Scaffolding

## Purpose

The Library Scaffolding capability enables users to create new library directory structures with valid `library.yaml` manifests and empty resource directories via the `germinator library init` command.

## Requirements

### Requirement: Create library directory structure

The system SHALL create a new library directory at the specified path with a valid `library.yaml` and empty resource directories.

#### Scenario: Create library at specified path
- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library` is executed
- **THEN** directory `/tmp/my-library/` is created
- **AND** file `/tmp/my-library/library.yaml` is created with version "1" and empty resources/presets
- **AND** directory `/tmp/my-library/skills/` is created
- **AND** directory `/tmp/my-library/agents/` is created
- **AND** directory `/tmp/my-library/commands/` is created
- **AND** directory `/tmp/my-library/memory/` is created

#### Scenario: Create library at default path
- **GIVEN** no library exists at `~/.local/share/germinator/library/`
- **WHEN** `germinator library init` is executed
- **THEN** library is created at `~/.local/share/germinator/library/`
- **OR** if `XDG_DATA_HOME` is set, library is created at `$XDG_DATA_HOME/germinator/library/`

### Requirement: Validate created library

The system SHALL validate created library by loading it via `LoadLibrary` to ensure structural correctness.

#### Scenario: Validate successful creation
- **GIVEN** library creation at `/tmp/my-library` succeeds
- **WHEN** `LoadLibrary("/tmp/my-library")` is called
- **THEN** a valid Library struct is returned
- **AND** `Library.Resources` contains empty maps for skill, agent, command, and memory
- **AND** `Library.Presets` is an empty map
- **AND** `Library.Version` equals "1"

### Requirement: Handle existing library

The system SHALL return an error if a library already exists at the target path unless `--force` is specified.

#### Scenario: Error when library exists
- **GIVEN** a library already exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library` is executed
- **THEN** an error is returned indicating the library already exists

#### Scenario: Overwrite with force flag
- **GIVEN** a library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --force` is executed
- **THEN** the library is replaced with a new empty library

### Requirement: Support dry-run mode

The system SHALL preview changes without creating files when `--dry-run` is specified.

#### Scenario: Dry-run does not create files
- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --dry-run` is executed
- **THEN** no files or directories are created
- **AND** a message is printed indicating what would be created

### Requirement: Create valid library.yaml content

The system SHALL create a `library.yaml` with version "1" and empty resource/preset sections.

#### Scenario: Library.yaml has correct structure
- **GIVEN** library is created at `/tmp/my-library`
- **WHEN** the `library.yaml` file is read
- **THEN** it parses as valid YAML
- **AND** it contains `version: "1"`
- **AND** it contains `resources:` with skill, agent, command, and memory types all empty maps
- **AND** it contains `presets:` as an empty map

#### Scenario: Permissions denied when creating directories
- **GIVEN** user lacks permission to create directories at `/protected/`
- **WHEN** `germinator library init --path /protected/library` is executed
- **THEN** an error is returned indicating permission was denied

#### Scenario: Disk full when writing files
- **GIVEN** disk space is exhausted at target path
- **WHEN** `germinator library init --path /full-disk/library` is executed
- **THEN** an error is returned indicating write failure

#### Scenario: Invalid path characters
- **GIVEN** path contains invalid characters for the filesystem
- **WHEN** `germinator library init --path "/tmp/invalid\x00path"` is executed
- **THEN** an error is returned indicating invalid path

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

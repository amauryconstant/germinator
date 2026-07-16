# cli-framework Specification

## Purpose

Define Cobra CLI framework with validate, adapt, canonicalize, init, library, config, completion, and version commands for document transformation. All commands obtain dependencies through `*cmdutil.Factory` (see `cli-cli-factory` and `cli-command-options-pattern`).

## Requirements

### Requirement: Cobra CLI Framework

The project SHALL use the Cobra framework for CLI command structure.

#### Scenario: Cobra dependency is installed

**Given** the Go module is initialized
**When** a developer runs `go get github.com/spf13/cobra@latest`
**Then** Cobra SHALL be added to go.mod
**And** go.sum SHALL be updated
**And** `go build ./...` SHALL succeed

#### Scenario: Cobra is imported in cmd package

**Given** the cmd/root.go file exists
**When** the file is inspected
**Then** it SHALL import github.com/spf13/cobra
**And** it SHALL use cobra.Command for command definition

---

### Requirement: Root Command Structure

The project SHALL have a root command named "germinator" with basic functionality and persistent flags.

#### Scenario: Root command exists

**Given** the project has been initialized
**When** a developer builds and runs the binary
**Then** a root command named "germinator" SHALL be available
**And** it SHALL display help when run with --help flag

#### Scenario: Root command has description

**Given** the root command exists
**When** a developer runs `./germinator --help`
**Then** the output SHALL include a description explaining the tool's purpose

#### Scenario: Root command has persistent verbose flag

**Given** the root command exists
**When** a developer runs `./germinator --help`
**Then** the output SHALL include a `-v` flag description
**And** the flag SHALL be marked as persistent

---

### Requirement: Commands take Factory

Each command's constructor SHALL take `*cmdutil.Factory` as its first parameter (after the optional `runF` for test injection). No command SHALL take `*CommandConfig`. State flows through per-command `*XxxOptions` structs.

#### Scenario: NewCmdXxx signature

- **WHEN** a command's constructor signature is inspected
- **THEN** it SHALL match `NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command`
- **AND** it SHALL NOT have any parameter of type `*CommandConfig`

### Requirement: No global command state

The `cmd` package SHALL NOT maintain any package-level mutable state for command configuration. All command state SHALL flow through per-command `*XxxOptions` structs populated in each command's `RunE` hook from `*cmdutil.Factory`.

#### Scenario: No global command config

- **WHEN** the `cmd` package is inspected
- **THEN** there SHALL be no package-level variable of type `*CommandConfig` or similar mutable state
- **AND** there SHALL be no `SetGlobalCommandConfig` function or equivalent
- **AND** no command SHALL call `cmd.GetCommandConfig()` or any global getter

---

### Requirement: RunE Command Pattern

All CLI commands that can fail SHALL use the `RunE` pattern instead of `Run`.

#### Scenario: Commands use RunE instead of Run

- **GIVEN** a command that can return errors (validate, adapt, canonicalize, init, library)
- **WHEN** the command definition is inspected
- **THEN** it SHALL use `RunE` field instead of `Run`
- **AND** the function signature SHALL be `func(cmd *cobra.Command, args []string) error`

#### Scenario: Commands return errors instead of calling os.Exit

- **GIVEN** a command encounters an error
- **WHEN** the error occurs in the RunE function
- **THEN** the error SHALL be returned
- **AND** os.Exit SHALL NOT be called within the command

#### Scenario: Version command can use Run

- **GIVEN** the version command cannot fail
- **WHEN** the command definition is inspected
- **THEN** it MAY use `Run` instead of `RunE`

---

### Requirement: Centralized Error Handling

Error handling SHALL be centralized in `main.go`. Commands return typed errors; `main.go` formats them and maps them to exit codes.

#### Scenario: main.go handles all errors

- **GIVEN** main.go executes the root command
- **WHEN** rootCmd.Execute() returns an error
- **THEN** `output.FormatError(io, err)` SHALL be called to render the error
- **AND** `cmdutil.ExitCodeFor(err)` SHALL be called to determine the exit code
- **AND** the process SHALL exit with that exit code

#### Scenario: Commands delegate error formatting

- **GIVEN** a command's `RunE` returns a typed error (e.g. `*core.ValidationError`)
- **WHEN** main.go processes the error
- **THEN** the typed-error dispatch in `output.FormatError` SHALL choose the correct rendering
- **AND** the command SHALL NOT call `output.FormatError` itself

#### Scenario: Cobra argument errors map to ExitCodeUsage

- **GIVEN** rootCmd.Execute() returns a Cobra usage error (unknown flag, missing arg, etc.)
- **WHEN** main.go calls `cmdutil.ExitCodeFor(err)`
- **THEN** the result SHALL be `ExitCodeUsage` (2)

---

### Requirement: CLI Entry Point

The project SHALL have a main entry point in cmd/root.go that initializes the CLI.

#### Scenario: Main function exists

**Given** the project has been initialized
**When** the cmd/root.go file is inspected
**Then** a main() function SHALL exist
**And** it SHALL execute the root command

#### Scenario: Root command is executable

**Given** the main function exists
**When** a developer runs `go build ./cmd`
**Then** a binary SHALL be produced
**And** the binary SHALL be executable
**And** running the binary SHALL execute the root command

---

### Requirement: Validate Command

The CLI SHALL provide a validate command that validates document files.

#### Scenario: Validate single document

**Given** a document file path is provided
**When** `germinator validate <file>` is run
**Then** the command SHALL parse the document
**And** it SHALL validate the document
**And** it SHALL display validation errors if any exist
**And** it SHALL return nil if valid
**And** it SHALL return `*core.ValidationError` for validation errors
**And** it SHALL return `*core.ConfigError` for parse errors

#### Scenario: Validate displays clear error messages

**Given** a document with validation errors
**When** `germinator validate <file>` is run
**Then** each error SHALL be displayed on a separate line
**And** errors SHALL be clearly formatted for human reading
**And** contextual hints SHALL be provided when available

#### Scenario: Validate command help

**Given** the validate command exists
**When** `germinator validate --help` is run
**Then** it SHALL display usage information
**And** it SHALL describe the command's purpose

#### Scenario: Validate handles missing file

**Given** a file that doesn't exist
**When** `germinator validate <file>` is run
**Then** the command SHALL display a file error message
**And** it SHALL return `*core.FileError`

#### Scenario: Validate handles invalid platform

**Given** an invalid platform flag
**When** `germinator validate <file> --platform invalid` is run
**Then** the command SHALL display a config error with valid platforms
**And** it SHALL return `*core.ConfigError`

#### Scenario: Validate takes a Factory

- **GIVEN** the validate command is run
- **WHEN** execution begins
- **THEN** its `NewCmdValidate(f *cmdutil.Factory, runF)` SHALL be invoked
- **AND** the constructed `validateOptions` SHALL hold `IO *iostreams.IOStreams`, lazy `Validator func() (Validator, error)`, and parsed flags
- **AND** the command SHALL NOT reference a `*CommandConfig`

---

### Requirement: Adapt Command

The CLI SHALL provide an adapt command that transforms documents to target platforms.

#### Scenario: Adapt document to platform

**Given** an input document file and output file path
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** the command SHALL parse the input document
**And** it SHALL validate the document
**And** it SHALL serialize the document
**And** it SHALL write to the output file
**And** it SHALL display a success message
**And** it SHALL return nil on success

#### Scenario: Adapt fails on validation

**Given** an invalid input document
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** the command SHALL parse the document
**And** it SHALL detect validation errors
**And** it SHALL NOT create the output file
**And** it SHALL display validation errors with hints
**And** it SHALL return `*core.ValidationError`

#### Scenario: Adapt fails on parse error

**Given** a document with invalid YAML
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL NOT create the output file
**And** it SHALL display a parse error with file path
**And** it SHALL return `*core.ConfigError`

#### Scenario: Adapt with Claude Code platform

**Given** a valid input document and output file
**When** `germinator adapt <input> <output> --platform claude-code` is run
**Then** it SHALL transform the document (pass-through validation and serialization)
**And** the output file SHALL contain the validated document

#### Scenario: Adapt command help

**Given** the adapt command exists
**When** `germinator adapt --help` is run
**Then** it SHALL display usage information
**And** it SHALL describe the command's purpose
**And** it SHALL list required arguments (input, output)
**And** it SHALL list required flags (--platform)

#### Scenario: Adapt handles read errors

**Given** an input file that cannot be read (permission denied)
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL display a file error with path and operation
**And** it SHALL NOT create the output file
**And** it SHALL return `*core.FileError`

#### Scenario: Adapt handles write errors

**Given** a valid input document but output directory is read-only
**When** `germinator adapt <input> <output> --platform <platform>` is run
**Then** it SHALL display a file error with path and operation
**And** it SHALL return `*core.FileError`

#### Scenario: Adapt handles invalid platform

**Given** an invalid platform flag
**When** `germinator adapt <input> <output> --platform invalid` is run
**Then** it SHALL display a config error with valid platforms
**And** it SHALL return `*core.ConfigError`

#### Scenario: Adapt takes a Factory

- **GIVEN** the adapt command is run
- **WHEN** execution begins
- **THEN** its `NewCmdAdapt(f *cmdutil.Factory, runF)` SHALL be invoked
- **AND** the constructed `adaptOptions` SHALL hold `IO`, lazy `Transformer` / `Validator`, and parsed flags
- **AND** the command SHALL NOT reference a `*CommandConfig`

---

### Requirement: Canonicalize Command

The CLI SHALL provide a canonicalize command that converts platform documents to canonical format.

#### Scenario: Canonicalize handles errors with typed errors

- **GIVEN** a platform document with errors
- **WHEN** `germinator canonicalize <input> <output>` is run
- **THEN** it SHALL display typed error messages
- **AND** it SHALL return the appropriate typed error

#### Scenario: Canonicalize takes a Factory

- **GIVEN** the canonicalize command is run
- **WHEN** execution begins
- **THEN** its `NewCmdCanonicalize(f *cmdutil.Factory, runF)` SHALL be invoked
- **AND** the constructed `canonicalizeOptions` SHALL hold `IO`, lazy `Canonicalizer`, and parsed flags
- **AND** the command SHALL NOT reference a `*CommandConfig`

---

### Requirement: validate and canonicalize command signatures

The `validate` and `canonicalize` commands SHALL take `*cmdutil.Factory` (per the `cli-cli-factory` capability) and follow the `cli-command-options-pattern` shape: `NewCmdValidate(f, runF)` + `validateOptions` + `runValidate`; `NewCmdCanonicalize(f, runF)` + `canonicalizeOptions` + `runCanonicalize`.

> **Note:** this requirement defines the command **signature shape**; behavioral requirements remain in the "Validate Command" / "Canonicalize Command" requirements of this spec.

#### Scenario: validate command signature

- **WHEN** `cmd/validate.go` is inspected
- **THEN** the constructor SHALL be `NewCmdValidate(f *cmdutil.Factory, runF func(*validateOptions) error) *cobra.Command`
- **AND** it SHALL NOT have any parameter of type `*CommandConfig`
- **AND** `validateOptions` SHALL declare `IO *iostreams.IOStreams`, `Validator func() (Validator, error)`, `Ctx context.Context`, `InputPath string`, `Platform string`

#### Scenario: canonicalize command signature

- **WHEN** `cmd/canonicalize.go` is inspected
- **THEN** the constructor SHALL be `NewCmdCanonicalize(f *cmdutil.Factory, runF func(*canonicalizeOptions) error) *cobra.Command`
- **AND** `canonicalizeOptions` SHALL declare `IO *iostreams.IOStreams`, `Canonicalizer func() (Canonicalizer, error)`, `Ctx context.Context`, `InputPath string`, `OutputPath string`, `Platform string`, `DocType string`

> **Note:** The validate/canonicalize behavior is implemented entirely within the per-command files (`cmd/validate.go`, `cmd/canonicalize.go`) as private helpers. There is no separate `internal/service/validator.go` or `internal/service/canonicalizer.go`.

---

### Requirement: Command Registration

The CLI SHALL register all new subcommands with the root command.

#### Scenario: Register validate command

**Given** the CLI is initialized
**When** the root command is inspected
**Then** it SHALL have a "validate" subcommand
**And** the subcommand SHALL be accessible via `germinator validate`

#### Scenario: Register adapt command

**Given** the CLI is initialized
**When** the root command is inspected
**Then** it SHALL have an "adapt" subcommand
**And** the subcommand SHALL be accessible via `germinator adapt`

#### Scenario: Register canonicalize command

**Given** the CLI is initialized
**When** the root command is inspected
**Then** it SHALL have a "canonicalize" subcommand
**And** the subcommand SHALL be accessible via `germinator canonicalize`

#### Scenario: Commands appear in help

**Given** the CLI is initialized
**When** `germinator --help` is run
**Then** the help output SHALL list all available commands
**And** it SHALL include "validate" in the commands list
**And** it SHALL include "adapt" in the commands list
**And** it SHALL include "canonicalize" in the commands list

---

### Requirement: Enhanced Version Display

The version command SHALL display version, commit SHA, and build date for better debugging.

#### Scenario: Version command shows full info

**Given** germinator is built with version information
**When** a user runs `germinator version`
**Then** it SHALL display format: `germinator {version} ({commit}) {date}`
**And** version SHALL be the semantic version (e.g., v0.3.0)
**And** commit SHALL be 7-character commit SHA (e.g., abc1234)
**And** date SHALL be YYYY-MM-DD format (e.g., 2025-01-13)

#### Scenario: Version with tag

**Given** germinator is built from a Git tag (e.g., v0.3.0)
**When** version command runs
**Then** it SHALL display: `germinator v0.3.0 (abc1234) 2025-01-13`
**And** version SHALL match Git tag
**And** commit SHALL be tag's commit SHA
**And** date SHALL be commit date

#### Scenario: Version without tag

**Given** germinator is built from non-tagged commit
**When** version command runs
**Then** it SHALL display: `germinator v0.3.0-1-gabc1234 (abc1234) 2025-01-13`
**And** version SHALL include git describe output
**And** commit SHALL be current HEAD SHA
**And** date SHALL be current date

---

### Requirement: Version Package Variables

The version package SHALL use variables instead of constants for build-time injection.

#### Scenario: Version is variable

**Given** `internal/version/version.go` is inspected
**When** version variable is declared
**Then** it SHALL use `var` instead of `const`
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "dev"

#### Scenario: Commit is variable

**Given** `internal/version/version.go` is inspected
**When** commit variable is declared
**Then** it SHALL use `var` for commit SHA
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "" (empty string)

#### Scenario: Date is variable

**Given** `internal/version/version.go` is inspected
**When** date variable is declared
**Then** it SHALL use `var` for build date
**And** it SHALL allow ldflags injection
**And** it SHALL have default value "" (empty string)

#### Scenario: Variables exported

**Given** version package is inspected
**When** exports are checked
**Then** `Version` variable SHALL be exported
**And** `Commit` variable SHALL be exported
**And** `Date` variable SHALL be exported

---

### Requirement: Version Command

The version command SHALL display version information for debugging and release tracking.

#### Scenario: Version command works

**Given** germinator is installed
**When** a user runs `germinator version`
**Then** it SHALL execute successfully
**And** it SHALL display version in format: `germinator {version} ({commit}) {date}`
**And** it SHALL exit with code 0

#### Scenario: Version help is available

**Given** a user runs `germinator version --help`
**Then** it SHALL display command help
**And** it SHALL show description: "Show version of germinator"

#### Scenario: Version with --output json

- **GIVEN** germinator is built with version metadata
- **WHEN** a user runs `germinator version --output json`
- **THEN** the command SHALL emit a single JSON object on stdout with keys `version`, `commit`, `date`, and `go`
- **AND** `stdout` SHALL contain exactly one JSON object (no trailing newline required)
- **AND** `stderr` SHALL be empty
- **AND** the process SHALL exit with code 0

### Requirement: Version emits runtime.GoVersion

The version command SHALL include the Go runtime version (via `runtime.Version()`) in its output. This field aids bug reports by recording the exact Go toolchain that produced the binary.

#### Scenario: Version output includes Go runtime version

- **GIVEN** germinator is built with any Go toolchain
- **WHEN** a user runs `germinator version` (any output format)
- **THEN** the rendered output SHALL contain the Go runtime version string (e.g., `go1.25.5`)
- **AND** for `--output json`, the `go` key SHALL be populated from `runtime.Version()`

### Requirement: I/O adapter placement

Service-style I/O adapters (Transformer, Validator, Canonicalizer, Initializer, and per-resource adders) MUST live in dedicated `internal/<x>/` shell packages, not in `cmd/`. The Functional Core / Imperative Shell pattern requires that any code performing I/O (filesystem reads/writes, external tool calls, network requests) live at the package boundary (`internal/<shell>/`), not in the action layer (`cmd/`).

**Change**: NEW requirement codifying the post-extraction state. The slice-3 design rationale (`openspec/changes/archive/2026-06-26-migrate-domain-commands/design.md:38-48`) was a one-adapter argument for keeping validator logic in `cmd/`; after `extract-io-adapters` all 5 adapters live in `internal/<x>/`.

#### Scenario: Adapters live in internal/<x>/, not cmd/

- **WHEN** the codebase is searched for `transformerAdapter`, `validatorAdapter`, `canonicalizerAdapter`, `initializerAdapter`, `libraryAdapter`
- **THEN** zero matches SHALL appear in `cmd/`
- **AND** matches SHALL appear in `internal/transform/`, `internal/validate/`, `internal/canonicalize/`, `internal/install/`, and `internal/library/` (as methods on `*Library`) respectively

#### Scenario: Cmd-side interfaces remain in cmd/

- **WHEN** the codebase is searched for the `Transformer`, `Validator`, `Canonicalizer`, `Initializer`, `resourceAdder` interfaces
- **THEN** each interface SHALL be declared in its consumer's `cmd/<command>.go` file (per the skill's "interfaces where consumed" principle)
- **AND** the interfaces SHALL NOT be re-exported or duplicated in `internal/<x>/` (the cmd-side contract is the canonical declaration)

#### Scenario: NewService constructors live in shell packages

- **WHEN** the codebase is searched for `NewService` or `NewXxx` constructors of the extracted adapters
- **THEN** each constructor SHALL live in its shell package (`internal/transform/transform.go`, `internal/validate/validate.go`, `internal/canonicalize/canonicalize.go`, `internal/install/install.go`)
- **AND** each constructor SHALL return the cmd-side interface type (the package implements the interface, the consumer declares it)
- **AND** `cmd/` SHALL import the shell package only to call the constructor; the cmd file does NOT re-implement the adapter

#### Scenario: Package name `install` (not `init`) avoids Go identifier collision

- **WHEN** a shell package for the init/initialize logic is created
- **THEN** the package SHALL be named `install`, not `init`
- **AND** the `Initializer` interface (the cmd-side contract) remains in `cmd/init.go`; only the implementation moves to `internal/install/`
- **Rationale**: `init` is a reserved Go package name; using it causes issues with Go tooling (`go test ./init/...` triggers linter warnings). `install` is semantically equivalent and avoids the collision (per design Decision 1).

#### Scenario: Library package methods replace the libraryAdapter

- **WHEN** the codebase is searched for `libraryAdapter`
- **THEN** zero matches SHALL appear (the type was deleted)
- **AND** `cmd/library_add.go` SHALL declare `var _ adderLibrary = (*library.Library)(nil)` as the compile-time check
- **AND** `*library.Library` SHALL expose `Add`, `BatchAddResources`, `DiscoverOrphans` methods (per slice-7 decision 6, completed by this change)

#### Scenario: New shell packages follow the internal/library convention

- **WHEN** any of the 4 new shell packages (`internal/validate/`, `internal/canonicalize/`, `internal/transform/`, `internal/install/`) is inspected
- **THEN** it SHALL have a `Service` interface, `Request`/`Result` types, and a `NewService` constructor
- **AND** it SHALL return `core.*` types (not package-local types)
- **AND** it SHALL take `ctx context.Context` as the first parameter of each public method
- **AND** it SHALL have an `AGENTS.md` following the `internal/library/AGENTS.md` template (Files table + Key Surface + skill reference)

### Requirement: I/O adapter ctx propagation

Service-style I/O adapters (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`, and the per-resource adders) — implemented in `internal/{transform,validate,canonicalize,install}/` shell packages and as `*Library` methods on `internal/library/library.go` — SHALL accept `ctx context.Context` as the first parameter of every public method. The `ctx` SHALL be forwarded to all downstream calls (`parser.LoadDocument`, `renderer.RenderDocument`, `LoadLibrary`, `SaveLibrary`, `*Library.Refresh`, `*Library.RemoveResource`, `*Library.ResolvePreset`, etc.).

The adapter SHALL NOT discard `ctx` (e.g., via `_ context.Context`). If a method does not need cancellation, the `ctx` SHALL still be accepted (and may be ignored), so the call site is uniform across the codebase. The canonical example of the accept-and-may-ignore pattern is `(*Library).ResolvePreset` at `library-library-resolution/spec.md` — a pure in-memory map lookup that accepts `ctx` for spec symmetry with no I/O to forward to today.

**Change**: NEW requirement. The pre-change adapters in `cmd/{initializer,transformer,canonicalize,validate}.go` accepted `ctx` as a method parameter but discarded it via the `_` underscore binding. The `extract-io-adapters` change relocates the adapters to `internal/<x>/` shell packages; this change threads `ctx` through, fulfilling the spec promise at `openspec/changes/extract-io-adapters/specs/cli-framework/spec.md:42`. The `internal/library/resolver.go:67` `ResolvePreset` underscore binding is also replaced with `ctx context.Context` for spec symmetry.

#### Scenario: Service method accepts ctx as first parameter

- **WHEN** an adapter's public method is called
- **THEN** the method signature SHALL have `ctx context.Context` as the first parameter
- **AND** the method SHALL forward `ctx` to all downstream `parser.*` / `renderer.*` / `library.*` calls
- **AND** any `ctx.Err()` check site SHALL wrap the sentinel with `%w` (e.g., `fmt.Errorf("...: %w", ctx.Err())`) so callers can `errors.Is(err, context.Canceled)`
- **AND** the method SHALL NOT use `context.Background()` or `context.TODO()` in place of the caller's `ctx` — synthesizing any context inside a request path violates the `golang-context` best practice of "never create a new context in the middle of a request path"

#### Scenario: Cancellation propagates through the adapter

- **GIVEN** a cmd-side `ctx` that is cancelled mid-call
- **WHEN** the adapter method is called with that `ctx`
- **THEN** the call SHALL return within bounded time (verified by `goleak` and `-race` tests)
- **AND** the returned error SHALL be `context.Canceled` or `context.DeadlineExceeded` (or wrap one of them via `%w`)

#### Scenario: Adapter signature inspection

- **WHEN** `mise run lint` runs `staticcheck` (enabled via `.golangci.yml:14`) against `internal/{transform,validate,canonicalize,install}/` (once `extract-io-adapters` lands; otherwise the legacy `cmd/{transformer,validate,canonicalize,initializer}.go` paths) and the `*Library` methods in `internal/library/library.go` and `internal/library/resolver.go`
- **THEN** zero `lostcancel` (go vet) violations SHALL appear
- **AND** zero `_ context\.Context` underscore-binding matches SHALL appear in the production-code scope (verified by `rg --type=go -g '!*_test.go' "_ context\.Context" cmd/ internal/`)
- **AND** zero `context\.TODO\(\)` matches SHALL appear in the production-code scope (verified by `rg --type=go -g '!*_test.go' "context\.TODO\(\)" cmd/ internal/`) — the `internal/library/resolver.go:54-56` package-level `ResolvePreset` shim was deleted by `propagate-context-through-shell` task 3.4b; this rg ensures no replacement shim is introduced
- **AND** every public method on the adapters SHALL have `ctx context.Context` as a named, non-underscore parameter

> **Note:** The pre-change adapters in `cmd/{initializer,transformer,canonicalize,validate}.go` accepted `ctx` but discarded it via the `_` underscore binding. After `extract-io-adapters` relocates the adapters to `internal/<x>/` shell packages, this scenario verifies the future-proofed state. The scenario applies regardless of whether the adapter is currently in `cmd/` (legacy) or `internal/<x>/` (post-extraction). Test fakes in `cmd/*_test.go` are exempt from the underscore-binding check (the rg glob excludes `*_test.go`); test fakes renamed to `ctx context.Context` (zero behavior change) are encouraged but not required.

## Coverage Note

The scenarios above codify **structural invariants** (placement, naming, package conventions) enforced primarily by compile-time checks (`var _ X = (*Y)(nil)`), `go build ./...`, and `rg` patterns from `tasks.md` Task 4.1.

**Behavioral coverage** for the same architectural surface lives in:

- `internal/{validate,canonicalize,transform,install}/*_test.go` — direct exercise of each shell-package `Service` interface
- `internal/library/methods_test.go` — table-driven coverage of the new `(*Library).Add` / `(*Library).BatchAddResources` / `(*Library).DiscoverOrphans` methods
- `cmd/library_add_test.go` — runtime mirror of the `var _ adderLibrary = (*library.Library)(nil)` contract assertion
- `cmd/{adapt,validate,canonicalize,init}_test.go` and `cmd/cmd_test.go` — `runF` injection and direct `xxx.NewService()` calls exercise the cmd→shell wiring

The `mise run test:coverage` threshold (70% per `openspec/config.yaml`) gates the new packages.

## Fulfilled

**Change:** `migrate-library-rest` (slice 7 of 9)
**Date:** 2026-07-01

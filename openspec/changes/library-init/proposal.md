## Why

Users need to create new library structures to organize their skills, agents, commands, and memory resources. Currently, the Germinator CLI can load existing libraries and install resources from them, but provides no way to scaffold a new library directory. Users must manually create the directory structure and `library.yaml` file, which is error-prone and not discoverable through the CLI.

## What Changes

- New `germinator library init` command that scaffolds a new library directory structure
- Creates `library.yaml` with empty resources and presets sections
- Creates empty resource type directories: `skills/`, `agents/`, `commands/`, `memory/`
- Validates created library by loading it to ensure structural correctness
- Supports `--path` flag to specify library location (defaults to `~/.config/germinator/library/`)
- Supports `--dry-run` flag to preview what would be created
- Supports `--force` flag to overwrite existing library
- Returns error if library already exists at target path without `--force`

## Capabilities

### New Capabilities

- `library-scaffolding`: Creates and validates new library directory structures with `library.yaml` and empty resource directories. Includes post-creation validation via `LoadLibrary` to ensure the created library is well-formed and loadable.

### Modified Capabilities

- None. This capability does not change existing library-system or resource-installation requirements.

## Impact

### New Files

- `internal/infrastructure/library/creator.go` - Library creation logic with `CreateLibrary` function and `CreateOptions` struct
- `cmd/library_init.go` - CLI command implementation for `germinator library init`
- `cmd/library_init_test.go` - Unit tests for library creation
- `test/e2e/library_init_test.go` - E2E tests for CLI integration
- `test/fixtures/library-init/` - Test fixtures for E2E tests

### Modified Files

- `cmd/library.go` - Add `NewLibraryInitCommand` to the library command's subcommands

### No Impact

- No changes to existing library loading or resource installation behavior
- No changes to domain models or service interfaces
- No changes to the `application.Initializer` interface (separate concern from library creation)

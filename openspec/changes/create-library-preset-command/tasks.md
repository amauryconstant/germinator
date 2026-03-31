## 1. Infrastructure: Library Persistence

- [x] 1.1 Create `internal/infrastructure/library/saver.go` with `SaveLibrary()` function
- [x] 1.2 Implement `AddPreset()` function to add preset to library in-memory
- [x] 1.3 Add `PresetExists()` function to check if preset name exists
- [x] 1.4 Implement YAML marshaling and file writing in `SaveLibrary()`
- [x] 1.5 Create `saver_test.go` with unit tests

## 2. CLI Command: library create

- [x] 2.1 Create `cmd/library_create.go`
- [x] 2.2 Implement `NewLibraryCreateCommand()` returning `create` subcommand group
- [x] 2.3 Implement `NewCreatePresetCommand()` with full flag handling
- [x] 2.4 Add `--resources` flag (required, comma-separated strings)
- [x] 2.5 Add `--description` flag (optional string)
- [x] 2.6 Add `--force` flag (optional boolean)
- [x] 2.7 Add `--library` flag (optional string, overrides default discovery)

## 3. CLI Integration

- [x] 3.1 Register `NewLibraryCreateCommand` in `NewLibraryCommand()` (cmd/library.go)
- [x] 3.2 Wire up error handling with typed errors (FileError, ConfigError)
- [x] 3.3 Add completions for `--resources` flag in completions.go (use existing `actionResources()`)

## 4. Output Formatting

- [x] 4.1 Create `formatPresetOutput()` function in `cmd/library_formatters.go`
- [x] 4.2 Display preset name, description, and resources list on success
- [x] 4.3 Match existing output formatting style (verbose vs quiet modes)

## 5. Error Handling

- [x] 5.1 Validate preset name is not empty/whitespace (return clear error)
- [x] 5.2 Validate resources list is not empty (return clear error)
- [x] 5.3 Check preset doesn't exist without --force (return ConfigError with suggestion)
- [x] 5.4 Validate all referenced resources exist in library before saving
- [x] 5.5 Handle file write errors with FileError

## 6. Testing

- [x] 6.1 Add unit tests for `SaveLibrary()` in `saver_test.go`
- [x] 6.2 Add unit tests for `AddPreset()` edge cases
- [x] 6.3 Add unit tests for preset validation
- [x] 6.4 Add CLI integration test in `cmd/library_create_test.go`
- [x] 6.5 Add E2E test in `test/e2e/` (if applicable)

## 7. Validation

- [x] 7.1 Run `mise run lint` and fix any issues
- [x] 7.2 Run `mise run format` to ensure code formatting
- [x] 7.3 Run `mise run test` to ensure all tests pass
- [x] 7.4 Run `mise run check` for full validation before commit

(End of file - total 52 lines)

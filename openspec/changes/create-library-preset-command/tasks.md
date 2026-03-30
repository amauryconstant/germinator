## 1. Infrastructure: Library Persistence

- [ ] 1.1 Create `internal/infrastructure/library/saver.go` with `SaveLibrary()` function
- [ ] 1.2 Implement `AddPreset()` function to add preset to library in-memory
- [ ] 1.3 Add `PresetExists()` function to check if preset name exists
- [ ] 1.4 Implement YAML marshaling and file writing in `SaveLibrary()`
- [ ] 1.5 Create `saver_test.go` with unit tests

## 2. CLI Command: library create

- [ ] 2.1 Create `cmd/library_create.go`
- [ ] 2.2 Implement `NewLibraryCreateCommand()` returning `create` subcommand group
- [ ] 2.3 Implement `NewCreatePresetCommand()` with full flag handling
- [ ] 2.4 Add `--resources` flag (required, comma-separated strings)
- [ ] 2.5 Add `--description` flag (optional string)
- [ ] 2.6 Add `--force` flag (optional boolean)
- [ ] 2.7 Add `--library` flag (optional string, overrides default discovery)

## 3. CLI Integration

- [ ] 3.1 Register `NewLibraryCreateCommand` in `NewLibraryCommand()` (cmd/library.go)
- [ ] 3.2 Wire up error handling with typed errors (FileError, ConfigError)
- [ ] 3.3 Add completions for `--resources` flag in completions.go (use existing `actionResources()`)

## 4. Output Formatting

- [ ] 4.1 Create `formatPresetOutput()` function in `cmd/library_formatters.go`
- [ ] 4.2 Display preset name, description, and resources list on success
- [ ] 4.3 Match existing output formatting style (verbose vs quiet modes)

## 5. Error Handling

- [ ] 5.1 Validate preset name is not empty/whitespace (return clear error)
- [ ] 5.2 Validate resources list is not empty (return clear error)
- [ ] 5.3 Check preset doesn't exist without --force (return ConfigError with suggestion)
- [ ] 5.4 Validate all referenced resources exist in library before saving
- [ ] 5.5 Handle file write errors with FileError

## 6. Testing

- [ ] 6.1 Add unit tests for `SaveLibrary()` in `saver_test.go`
- [ ] 6.2 Add unit tests for `AddPreset()` edge cases
- [ ] 6.3 Add unit tests for preset validation
- [ ] 6.4 Add CLI integration test in `cmd/library_create_test.go`
- [ ] 6.5 Add E2E test in `test/e2e/` (if applicable)

## 7. Validation

- [ ] 7.1 Run `mise run lint` and fix any issues
- [ ] 7.2 Run `mise run format` to ensure code formatting
- [ ] 7.3 Run `mise run test` to ensure all tests pass
- [ ] 7.4 Run `mise run check` for full validation before commit

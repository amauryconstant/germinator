## 1. Infrastructure Layer

- [ ] 1.1 Create `internal/infrastructure/library/remover.go`
- [ ] 1.2 Implement `RemoveResource()` function with ref parsing, library loading, existence check, preset reference check, file deletion, YAML update, and validation
- [ ] 1.3 Implement `RemovePreset()` function with name validation, library loading, existence check, YAML update, and validation
- [ ] 1.4 Add JSON output types for `RemoveResource` and `RemovePreset`

## 2. Command Layer

- [ ] 2.1 Create `cmd/library_remove.go`
- [ ] 2.2 Implement `NewLibraryRemoveCommand()` with `resource` and `preset` subcommands
- [ ] 2.3 Register `remove` command in `cmd/library.go`
- [ ] 2.4 Add `--json` flag to both subcommands
- [ ] 2.5 Add `--library` flag to both subcommands for path override (implements library path resolution from specs)

## 3. Testing

- [ ] 3.1 Create `internal/infrastructure/library/remover_test.go` with table-driven tests
- [ ] 3.2 Test `RemoveResource()` - success case, resource not found, preset reference conflict, invalid ref format
- [ ] 3.3 Test `RemovePreset()` - success case, preset not found, invalid name
- [ ] 3.4 Create E2E tests for `library remove resource` command
- [ ] 3.5 Create E2E tests for `library remove preset` command

## 4. Validation

- [ ] 4.1 Run `mise run check` to verify linting, formatting, and tests pass
- [ ] 4.2 Verify all existing library tests still pass

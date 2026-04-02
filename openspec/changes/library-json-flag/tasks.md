## 1. Add --json Flag to Parent Library Command

- [ ] 1.1 Add persistent `--json` flag to `NewLibraryCommand` in `cmd/library.go`
- [ ] 1.2 Remove local `--json` flag from `NewLibraryRefreshCommand` (it will inherit from parent)
- [ ] 1.3 Remove local `--json` flags from `NewLibraryRemoveResourceCommand` and `NewLibraryRemovePresetCommand` (they will inherit from parent)
- [ ] 1.4 Remove local `--json` flag from `NewLibraryValidateCommand` (it will inherit from parent)

## 2. Add JSON Output to Library Resources Command

- [ ] 2.1 Create `outputResourcesJSON` function in `cmd/library_formatters.go`
- [ ] 2.2 Modify `NewLibraryResourcesCommand` to check `--json` flag and call `outputResourcesJSON`
- [ ] 2.3 Test `germinator library resources --json` outputs correct JSON structure

## 3. Add JSON Output to Library Presets Command

- [ ] 3.1 Create `outputPresetsJSON` function in `cmd/library_formatters.go`
- [ ] 3.2 Modify `NewLibraryPresetsCommand` to check `--json` flag and call `outputPresetsJSON`
- [ ] 3.3 Test `germinator library presets --json` outputs correct JSON structure

## 4. Add JSON Output to Library Add Command

- [ ] 4.1 Modify `NewLibraryAddCommand` to check `--json` flag (inherited from parent)
- [ ] 4.2 Create JSON output for add success in `cmd/library_add.go`
- [ ] 4.3 Create JSON output for add with failures in `cmd/library_add.go`
- [ ] 4.4 Test `germinator library add <files> --json` outputs correct JSON structure

## 5. Add JSON Output to Library Show Command

- [ ] 5.1 Create `outputShowResourceJSON` function in `cmd/library_formatters.go`
- [ ] 5.2 Create `outputShowPresetJSON` function in `cmd/library_formatters.go`
- [ ] 5.3 Modify `NewLibraryShowCommand` to check `--json` flag and call appropriate JSON output function
- [ ] 5.4 Test `germinator library show skill/commit --json` outputs correct JSON structure
- [ ] 5.5 Test `germinator library show preset/git-workflow --json` outputs correct JSON structure

## 6. Add JSON Output to Library Init Command

- [ ] 6.1 Modify `NewLibraryInitCommand` to accept and check `--json` flag (inherited from parent)
- [ ] 6.2 Create JSON output for init success/failure in `cmd/library_init.go`
- [ ] 6.3 Test `germinator library init --path /tmp/test-lib --json` outputs correct JSON structure

## 7. Ensure Backward Compatibility

- [ ] 7.1 Verify `germinator library refresh --json` still works (existing implementation)
- [ ] 7.2 Verify `germinator library remove resource --json` still works (existing implementation)
- [ ] 7.3 Verify `germinator library remove preset --json` still works (existing implementation)
- [ ] 7.4 Verify `germinator library validate --json` still works (existing implementation)

## 8. Add Tests

- [ ] 8.1 Add unit tests for `outputResourcesJSON` function
- [ ] 8.2 Add unit tests for `outputPresetsJSON` function
- [ ] 8.3 Add unit tests for JSON output in show command
- [ ] 8.4 Add unit tests for JSON output in add command
- [ ] 8.5 Add E2E test for `library resources --json`
- [ ] 8.6 Add E2E test for `library presets --json`
- [ ] 8.7 Add E2E test for `library show --json`
- [ ] 8.8 Add E2E test for `library add --json`
- [ ] 8.9 Add E2E test for `library init --json`

## 9. Final Verification

- [ ] 9.1 Run `mise run check` and ensure all checks pass
- [ ] 9.2 Run `mise run test:full` and ensure all tests pass

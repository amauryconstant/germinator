## 1. Add --json Flag to Parent Library Command

- [ ] 1.1 Add persistent `--json` flag to `NewLibraryCommand` in `cmd/library.go`
- [ ] 1.2 Remove local `--json` flag from `NewLibraryRefreshCommand` (it will inherit from parent)
- [ ] 1.3 Remove local `--json` flag from `NewLibraryRemoveCommand` (it will inherit from parent)
- [ ] 1.4 Remove local `--json` flag from `NewLibraryValidateCommand` (it will inherit from parent)

## 2. Add JSON Output to Library Resources Command

- [ ] 2.1 Create `outputResourcesJSON` function in `cmd/library_formatters.go`
- [ ] 2.2 Modify `NewLibraryResourcesCommand` to check `--json` flag and call `outputResourcesJSON`
- [ ] 2.3 Test `germinator library resources --json` outputs correct JSON structure

## 3. Add JSON Output to Library Presets Command

- [ ] 3.1 Create `outputPresetsJSON` function in `cmd/library_formatters.go`
- [ ] 3.2 Modify `NewLibraryPresetsCommand` to check `--json` flag and call `outputPresetsJSON`
- [ ] 3.3 Test `germinator library presets --json` outputs correct JSON structure

## 4. Add JSON Output to Library Show Command

- [ ] 4.1 Create `outputShowResourceJSON` function in `cmd/library_formatters.go`
- [ ] 4.2 Create `outputShowPresetJSON` function in `cmd/library_formatters.go`
- [ ] 4.3 Modify `NewLibraryShowCommand` to check `--json` flag and call appropriate JSON output function
- [ ] 4.4 Test `germinator library show skill/commit --json` outputs correct JSON structure
- [ ] 4.5 Test `germinator library show preset/git-workflow --json` outputs correct JSON structure

## 5. Add JSON Output to Library Init Command

- [ ] 5.1 Modify `NewLibraryInitCommand` to accept and check `--json` flag (inherited from parent)
- [ ] 5.2 Create JSON output for init success/failure in `cmd/library_init.go`
- [ ] 5.3 Test `germinator library init --path /tmp/test-lib --json` outputs correct JSON structure

## 6. Ensure Backward Compatibility

- [ ] 6.1 Verify `germinator library refresh --json` still works (existing implementation)
- [ ] 6.2 Verify `germinator library remove --json` still works (existing implementation)
- [ ] 6.3 Verify `germinator library validate --json` still works (existing implementation)

## 7. Add Tests

- [ ] 7.1 Add unit tests for `outputResourcesJSON` function
- [ ] 7.2 Add unit tests for `outputPresetsJSON` function
- [ ] 7.3 Add unit tests for JSON output in show command
- [ ] 7.4 Add E2E test for `library resources --json`
- [ ] 7.5 Add E2E test for `library presets --json`
- [ ] 7.6 Add E2E test for `library show --json`
- [ ] 7.7 Add E2E test for `library init --json`

## 8. Final Verification

- [ ] 8.1 Run `mise run check` and ensure all checks pass
- [ ] 8.2 Run `mise run test:full` and ensure all tests pass

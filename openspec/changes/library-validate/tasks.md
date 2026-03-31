## 1. Infrastructure

- [ ] 1.1 Create `internal/infrastructure/library/validator.go` with Issue and IssueType types
- [ ] 1.2 Implement `ValidateLibrary()` that runs all four checks
- [ ] 1.3 Implement `CheckMissingFiles()` - verifies entries in library.yaml have corresponding files
- [ ] 1.4 Implement `CheckOrphanedFiles()` - finds files on disk not in library.yaml
- [ ] 1.5 Implement `CheckGhostResources()` - verifies preset refs exist in library
- [ ] 1.6 Implement `CheckMalformedFrontmatter()` - parses frontmatter from each resource file
- [ ] 1.7 Implement `FixLibrary()` - removes missing entries and ghost preset refs from library.yaml
- [ ] 1.8 Create `validator_test.go` with table-driven tests for each check

## 2. Command

- [ ] 2.1 Create `cmd/library_validate.go` with NewLibraryValidateCommand
- [ ] 2.2 Add `--library` flag (persistent, shared with other library commands)
- [ ] 2.3 Add `--fix` flag for auto-cleanup
- [ ] 2.4 Add `--json` flag for machine-readable output
- [ ] 2.5 Implement human-readable output format with severity indicators
- [ ] 2.6 Implement JSON output format
- [ ] 2.7 Wire up exit codes (0 clean, 5 errors, 1 unexpected)
- [ ] 2.8 Register command in library command group
- [ ] 2.9 Create `library_validate_test.go` with command integration tests

## 3. Verification

- [ ] 3.1 Run `mise run check` to verify linting and formatting
- [ ] 3.2 Run `mise run test` to verify unit tests pass
- [ ] 3.3 Run `mise run test:e2e` if applicable

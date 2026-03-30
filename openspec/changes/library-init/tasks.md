## 1. Library Infrastructure

- [ ] 1.1 Create `internal/infrastructure/library/creator.go` with `CreateLibrary` function
- [ ] 1.2 Add `CreateOptions` struct with `Path`, `DryRun`, `Force` fields
- [ ] 1.3 Implement `CreateLibrary` to create directory structure and `library.yaml`
- [ ] 1.4 Add `defaultLibraryYAML()` function returning valid YAML content
- [ ] 1.5 Add error handling for existing library without `--force`
- [ ] 1.6 Validate created library via `LoadLibrary` after creation; on failure, return error but leave partial structure for debugging

## 2. CLI Command

- [ ] 2.1 Create `cmd/library_init.go` with `NewLibraryInitCommand` function
- [ ] 2.2 Add `--path` flag (defaults to `DefaultLibraryPath()`)
- [ ] 2.3 Add `--dry-run` flag for preview mode
- [ ] 2.4 Add `--force` flag for overwriting
- [ ] 2.5 Wire command into `cmd/library.go` as subcommand

## 3. Unit Tests

- [ ] 3.1 Create `cmd/library_init_test.go` with test cases
- [ ] 3.2 Test library creation at custom path
- [ ] 3.3 Test error when library exists (without force)
- [ ] 3.4 Test force overwrite existing library
- [ ] 3.5 Test dry-run mode
- [ ] 3.6 Test created library is valid and loadable

## 4. E2E Tests

- [ ] 4.1 Create `test/e2e/library_init_test.go` with Ginkgo v2 tests
- [ ] 4.2 Create `test/fixtures/library-init/` test fixture directory
- [ ] 4.3 Test `library init --path <tmp>` creates valid structure
- [ ] 4.4 Test `library init --dry-run` shows preview without creating
- [ ] 4.5 Test `library init --force` overwrites existing
- [ ] 4.6 Test running `germinator library init --path <tmp>` with invalid path returns appropriate error
- [ ] 4.7 Test running `germinator library init` with permissions denied returns appropriate error

## 5. Validation

- [ ] 5.1 Run `mise run lint` and fix any issues
- [ ] 5.2 Run `mise run format` to ensure code is formatted
- [ ] 5.3 Run `mise run test` and ensure all tests pass
- [ ] 5.4 Run `mise run test:e2e` and ensure E2E tests pass

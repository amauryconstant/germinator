## 1. Library Infrastructure

- [x] 1.1 Create `internal/infrastructure/library/creator.go` with `CreateLibrary` function
- [x] 1.2 Add `CreateOptions` struct with `Path`, `DryRun`, `Force` fields
- [x] 1.3 Implement `CreateLibrary` to create directory structure and `library.yaml`
- [x] 1.4 Add `defaultLibraryYAML()` function returning valid YAML content
- [x] 1.5 Add error handling for existing library without `--force`
- [x] 1.6 Validate created library via `LoadLibrary` after creation; on failure, return error but leave partial structure for debugging

## 2. CLI Command

- [x] 2.1 Create `cmd/library_init.go` with `NewLibraryInitCommand` function
- [x] 2.2 Add `--path` flag (defaults to `DefaultLibraryPath()`)
- [x] 2.3 Add `--dry-run` flag for preview mode
- [x] 2.4 Add `--force` flag for overwriting
- [x] 2.5 Wire command into `cmd/library.go` as subcommand

## 3. Unit Tests

- [x] 3.1 Create `cmd/library_init_test.go` with test cases
- [x] 3.2 Test library creation at custom path
- [x] 3.3 Test error when library exists (without force)
- [x] 3.4 Test force overwrite existing library
- [x] 3.5 Test dry-run mode
- [x] 3.6 Test created library is valid and loadable

## 4. E2E Tests

- [x] 4.1 Create `test/e2e/library_init_test.go` with Ginkgo v2 tests
- [x] 4.2 Create `test/fixtures/library-init/` test fixture directory
- [x] 4.3 Test `library init --path <tmp>` creates valid structure
- [x] 4.4 Test `library init --dry-run` shows preview without creating
- [x] 4.5 Test `library init --force` overwrites existing
- [x] 4.6 Test running `germinator library init --path <tmp>` with invalid path returns appropriate error
- [x] 4.7 Test running `germinator library init` with permissions denied returns appropriate error

## 5. Validation

- [x] 5.1 Run `mise run lint` and fix any issues
- [x] 5.2 Run `mise run format` to ensure code is formatted
- [x] 5.3 Run `mise run test` and ensure all tests pass
- [x] 5.4 Run `mise run test:e2e` and ensure E2E tests pass

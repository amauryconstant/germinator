## 1. Library Refresh Infrastructure

- [x] 1.1 Create `internal/infrastructure/library/refresher.go` with `RefreshOptions`, `RefreshResult`, `RefreshError` types
- [x] 1.2 Implement `RefreshLibrary(opts RefreshOptions) (*RefreshResult, error)` function
- [x] 1.3 Reuse `extractFrontmatterField` from `adder.go` for frontmatter extraction
- [x] 1.4 Add helper to detect frontmatter name for conflict checking
- [x] 1.5 Implement description sync logic (compare frontmatter desc with library.yaml desc)
- [x] 1.6 Implement path update detection (file found at different path, name matches key)
- [x] 1.7 Implement name mismatch conflict detection
- [x] 1.8 Implement malformed frontmatter error handling
- [x] 1.9 Implement error collection (all errors collected, not fail-fast)
- [x] 1.10 Add `refresher_test.go` with table-driven tests for all scenarios

## 2. Library Refresh CLI Command

- [x] 2.1 Create `cmd/library_refresh.go` with `NewLibraryRefreshCommand`
- [x] 2.2 Add flags: `--dry-run`, `--force`, `--json`
- [x] 2.3 Implement `runLibraryRefresh` with library path discovery
- [x] 2.4 Wire to `library.RefreshLibrary`
- [x] 2.5 Add refresh output formatting in `cmd/library_formatters.go`
- [x] 2.6 Register `refresh` subcommand in `cmd/library.go`

## 3. Library Add Discover Mode

- [x] 3.1 Add `--discover` flag to `cmd/library_add.go`
- [x] 3.2 Add `DiscoverOptions` type and `DiscoverOrphans` function in `adder.go`
- [x] 3.3 Implement orphan scanning (scan skills/, agents/, commands/, memory/ directories)
- [x] 3.4 Implement orphan detection (type from directory, name from frontmatter/filename, description from frontmatter)
- [x] 3.5 Implement conflict detection (orphan name matches existing resource)
- [x] 3.6 Implement orphan registration with `--force`
- [x] 3.7 Add discover output formatting in `cmd/library_formatters.go`
- [x] 3.8 Update help text and examples for --discover flag

## 4. Testing

- [x] 4.1 Add unit tests for `RefreshLibrary` in `refresher_test.go`
- [x] 4.2 Add unit tests for orphan discovery in `adder_test.go` (TestDiscoverOrphans exists in refresher_test.go)
- [x] 4.3 Create `test/e2e/library_refresh_test.go` with E2E scenarios
- [x] 4.4 Create `test/e2e/library_discover_test.go` with E2E scenarios
- [x] 4.5 Add fixtures for refresh and discover test scenarios (existing fixtures library used)

## 5. Integration

- [x] 5.1 Run `mise run check` to verify lint, format, tests pass
- [x] 5.2 Verify all new code follows conventions (no comments, table-driven tests)

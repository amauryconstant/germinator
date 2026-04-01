## 1. Infrastructure Layer - BatchAddResult Types

- [ ] 1.1 Add BatchAddResult struct and nested types (AddSuccess, SkipInfo, FailureInfo, BatchSummary) to internal/infrastructure/library/adder.go
- [ ] 1.2 Add BatchAddOptions struct with Sources, LibraryPath, DryRun, Force, Name, Description, Type, Platform fields
- [ ] 1.3 Implement BatchAddResources function signature that returns (*BatchAddResult, error)

## 2. Infrastructure Layer - BatchAddResources Implementation

- [ ] 2.1 Implement directory scanning: walk directory tree to collect all .md files
- [ ] 2.2 Implement batch processing loop: iterate sources, detect type/name/platform per file
- [ ] 2.3 Implement skip detection: check if resource already exists (ErrResourceExists case)
- [ ] 2.4 Implement conflict detection: orphan name matches existing resource
- [ ] 2.5 Implement failure handling: collect errors without stopping
- [ ] 2.6 Implement summary calculation: count added/skipped/failed

## 3. Command Layer - CLI Changes

- [ ] 3.1 Add `--batch` flag to NewLibraryAddCommand in cmd/library_add.go
- [ ] 3.2 Modify Args validation: batch mode accepts 0+ args, non-batch requires 1 arg
- [ ] 3.3 Create new runBatchAdd function to handle batch mode
- [ ] 3.4 Modify runLibraryAdd to call batch or single-file path based on flag
- [ ] 3.5 Ensure batch mode always returns nil error (exit code 0)

## 4. Command Layer - Output Formatting

- [ ] 4.1 Add FormatBatchAddSummary function in cmd/library_formatters.go
- [ ] 4.2 Implement human-readable summary output: "Added N, skipped M, failed K"
- [ ] 4.3 Check json flag on parent command and output JSON result when set

## 5. Discover Integration

- [ ] 5.1 Modify runLibraryDiscover to support batch mode
- [ ] 5.2 When --discover --batch is set, process all orphans through BatchAddResources
- [ ] 5.3 Handle --discover --batch --force to register all orphans

## 6. Testing

- [ ] 6.1 Add unit tests for BatchAddResources in internal/infrastructure/library/adder_test.go
- [ ] 6.2 Add test cases for directory scanning (recursive .md discovery)
- [ ] 6.3 Add test cases for skip detection (already_exists, conflict)
- [ ] 6.4 Add test cases for failure collection
- [ ] 6.5 Add E2E tests for batch flag in test/e2e/library_test.go
- [ ] 6.6 Add E2E tests for --discover --batch combination

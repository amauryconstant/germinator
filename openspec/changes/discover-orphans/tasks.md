## 1. Update Discover Result Types

- [x] 1.1 Add `DiscoverSummary` struct with TotalScanned, TotalOrphans, TotalAdded, TotalSkipped, TotalFailed fields
- [x] 1.2 Rename `DiscoverOrphan` to `OrphanInfo` with Path, Type, Name, Issue fields
- [x] 1.3 Rename `DiscoverConflict` to `ConflictInfo` with Orphan and Issue fields
- [x] 1.4 Add `AddSuccess` struct with Type, Name, Path fields
- [x] 1.5 Update `DiscoverResult` to include Summary field

## 2. Implement Recursive Directory Scanning

- [ ] 2.1 Modify `scanDirectory` to use `filepath.WalkDir` for recursive traversal
- [ ] 2.2 Update file matching to include `.md` extension check
- [ ] 2.3 Track TotalScanned count during directory walk
- [ ] 2.4 Handle nested paths correctly for orphan name detection

## 3. Implement Batch Mode Logic

- [ ] 3.1 Add `Batch` field to `DiscoverOptions` struct
- [ ] 3.2 Modify DiscoverOrphans to process all orphans continuously
- [ ] 3.3 Track TotalAdded, TotalSkipped, TotalFailed in Summary
- [ ] 3.4 Continue processing on individual registration errors

## 4. Update CLI Command

- [ ] 4.1 Add `--batch` flag to `library add` command
- [ ] 4.2 Update output formatting for batch mode
- [ ] 4.3 Add human-readable summary output
- [ ] 4.4 Support `--json` flag for discover results

## 5. Add Unit Tests

- [x] 5.1 Add tests for recursive scanning with nested directories
- [x] 5.2 Add tests for enhanced DiscoverResult structure
- [x] 5.3 Add tests for batch mode with conflicts and errors
- [x] 5.4 Run `mise run test` to verify all tests pass

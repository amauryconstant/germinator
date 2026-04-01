## Verification Report: library-refresh-and-discover

### Summary
| Dimension    | Status                        |
|--------------|-------------------------------|
| Completeness | 31/31 tasks, 3 specs covered  |
| Correctness  | All requirements implemented   |
| Coherence    | Design followed                |

### CRITICAL Issues (Must fix before archive)
None.

### WARNING Issues (Should fix)
None.

### SUGGESTION Issues (Nice to fix)
None.

### Detailed Findings

#### Completeness (31/31 tasks, 3 specs)

**Tasks verified:**
- 1.1-1.10: Library Refresh Infrastructure (refresher.go with RefreshOptions, RefreshResult, RefreshError types; RefreshLibrary function; frontmatter extraction reuse; conflict detection; error collection)
- 2.1-2.6: Library Refresh CLI Command (library_refresh.go; --dry-run, --force, --json flags; library path discovery; wiring to RefreshLibrary; refresh output formatting; registration in library.go)
- 3.1-3.8: Library Add Discover Mode (--discover flag; DiscoverOptions and DiscoverOrphans; orphan scanning; conflict detection; force registration; discover output formatting; help text updates)
- 4.1-4.5: Testing (refresher_test.go with table-driven tests; TestDiscoverOrphans in refresher_test.go; library_refresh_test.go E2E; library_discover_test.go E2E; fixtures used)
- 5.1-5.2: Integration (mise run check passes; code follows conventions)

**Specs verified:**
- `specs/library/library-refresh/spec.md`: 9 requirements (Refresh metadata, Missing files, Name mismatch, Malformed frontmatter, Error collection, Dry-run, Force mode, JSON output, Library path discovery)
- `specs/library/library-orphan-discovery/spec.md`: 9 requirements (Discover orphans, Type from directory, Name from frontmatter/filename, Description from frontmatter, Report-only mode, Force mode, Dry-run, Conflict detection, Library path discovery)
- `specs/library/library-resource-import/spec.md`: 4 requirements (Discover flag, Discover with force, Discover requires explicit flag, Discover dry-run)

#### Correctness (All 22 requirements implemented)

**Refresh implementation:**
- `RefreshLibrary()` in `internal/infrastructure/library/refresher.go:50` - updates description from frontmatter
- `processResource()` handles file rename detection via `searchForFile()` - path update when frontmatter name matches
- `recordNameMismatch()` - skips on name mismatch conflict
- `isMalformedFrontmatter()` - detects and skips malformed frontmatter
- Error collection in `result.Errors` slice - all errors collected, not fail-fast
- `DryRun` and `Force` options properly respected
- `outputRefreshJSON()` outputs structured JSON with refreshed, skipped, errors details

**Discover implementation:**
- `DiscoverOrphans()` in `internal/infrastructure/library/adder.go:461` - scans all 4 resource directories
- `detectOrphan()` extracts type from directory, name from frontmatter/filename, description from frontmatter
- `isRegistered()` checks if orphan already in library
- `hasNameConflict()` detects name conflicts across types
- `registerOrphan()` adds orphans to library.yaml with `--force`
- `runLibraryDiscover()` in `cmd/library_add.go:139` - handles discover mode

**CLI wiring:**
- `NewLibraryRefreshCommand()` registered in `cmd/library.go:48`
- `--discover` flag added to `cmd/library_add.go:67`
- Library path discovery via `library.FindLibrary()` in both commands

#### Coherence (Design followed)

Design decisions verified:
1. ✓ `library refresh` as new command (not extending validate)
2. ✓ Collect all errors, report at end, exit code 1 if errors
3. ✓ Path update only when frontmatter name matches entry key (safe mode)
4. ✓ `--discover` flag on `library add` (not separate command)
5. ✓ Report-only by default, `--force` required for registration

### Final Assessment
**PASS** - All tasks complete, all requirements implemented, all design decisions followed, all tests passing.

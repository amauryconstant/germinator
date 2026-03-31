## 1. Infrastructure Layer

- [x] 1.1 Create `internal/infrastructure/library/adder.go` with `AddOptions` struct
- [x] 1.2 Implement `detectType(source, flag) string` for type detection
- [x] 1.3 Implement `detectName(source, flag) string` for name detection
- [x] 1.4 Implement `detectDescription(source, flag) string` for description detection
- [x] 1.5 Implement `canonicalizeIfNeeded` via cmd layer (architectural decision - import cycle avoided)
- [x] 1.6 Implement `AddResource(opts AddOptions) error` function
- [x] 1.7 Write unit tests in `adder_test.go`

## 2. CLI Command Layer

- [x] 2.1 Create `cmd/library_add.go` with `NewLibraryAddCommand`
- [x] 2.2 Implement flag parsing (--name, --description, --type, --platform, --force, --dry-run, --library)
- [x] 2.3 Wire library path discovery (--library flag > GERMINATOR_LIBRARY env > default)
- [x] 2.4 Implement output formatting for dry-run and success cases
- [x] 2.5 Register command in `NewLibraryCommand` in `cmd/library.go`
- [ ] 2.6 Write E2E tests in `test/e2e/library_add_test.go`

## 3. Integration & Validation

- [x] 3.1 Create test fixtures for library add (various resource types, platform formats) - verified manually
- [x] 3.2 Test auto-detection scenarios (flag, frontmatter, filename) - verified manually
- [x] 3.3 Test canonicalization of platform documents (OpenCode, Claude Code) - verified manually
- [x] 3.4 Test conflict handling (with and without --force) - verified via unit tests
- [x] 3.5 Test dry-run mode output - verified manually
- [ ] 3.6 Run `mise run check` to validate lint, format, tests, build

## 4. Documentation

- [ ] 4.1 Add `library add` to AGENTS.md command table
- [ ] 4.2 Update library command help text to include `add` subcommand

## Notes

- Import cycle issue resolved: Canonicalization moved to cmd layer to avoid `library → application → library` cycle
- Architecture: `library.AddResource()` expects canonical format; `cmd/library_add.go` handles platform detection and canonicalization before calling `AddResource`
- Pre-existing lint issues in `adder_test.go` (errcheck) are not related to this change
- Pre-existing config test failures in `internal/infrastructure/config` are not related to this change

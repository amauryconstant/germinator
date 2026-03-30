## 1. Infrastructure Layer

- [ ] 1.1 Create `internal/infrastructure/library/adder.go` with `AddOptions` struct
- [ ] 1.2 Implement `detectType(source, flag) string` for type detection
- [ ] 1.3 Implement `detectName(source, flag) string` for name detection
- [ ] 1.4 Implement `detectDescription(source, flag) string` for description detection
- [ ] 1.5 Implement `canonicalizeIfNeeded(source, platform, docType) (canonicalPath, error)`
- [ ] 1.6 Implement `AddResource(opts AddOptions) error` function
- [ ] 1.7 Write unit tests in `adder_test.go`

## 2. CLI Command Layer

- [ ] 2.1 Create `cmd/library_add.go` with `NewLibraryAddCommand`
- [ ] 2.2 Implement flag parsing (--name, --description, --type, --platform, --force, --dry-run, --library)
- [ ] 2.3 Wire library path discovery (--library flag > GERMINATOR_LIBRARY env > default)
- [ ] 2.4 Implement output formatting for dry-run and success cases
- [ ] 2.5 Register command in `NewLibraryCommand` in `cmd/library.go`
- [ ] 2.6 Write E2E tests in `test/e2e/library_add_test.go`

## 3. Integration & Validation

- [ ] 3.1 Create test fixtures for library add (various resource types, platform formats)
- [ ] 3.2 Test auto-detection scenarios (flag, frontmatter, filename)
- [ ] 3.3 Test canonicalization of platform documents (OpenCode, Claude Code)
- [ ] 3.4 Test conflict handling (with and without --force)
- [ ] 3.5 Test dry-run mode output
- [ ] 3.6 Run `mise run check` to validate lint, format, tests, build

## 4. Documentation

- [ ] 4.1 Add `library add` to AGENTS.md command table
- [ ] 4.2 Update library command help text to include `add` subcommand

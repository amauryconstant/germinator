## 1. Create Package Structure

- [x] 1.1 Create `internal/domain/` directory

## 2. Move Domain Types

- [x] 2.1 Split `internal/models/canonical/models.go` into type-specific files in `internal/domain/`:
  - [x] 2.1a Move Agent types to `internal/domain/agent.go`
  - [x] 2.1b Move Command types to `internal/domain/command.go`
  - [x] 2.1c Move Skill types to `internal/domain/skill.go`
  - [x] 2.1d Move Memory types to `internal/domain/memory.go`
  - [x] 2.1e Move Platform/PermissionPolicy types to `internal/domain/platform.go`
- [x] 2.2 Move `internal/errors/` to `internal/domain/errors.go`
- [x] 2.3 Move `internal/validation/validators.go` to `internal/domain/validation.go`
- [x] 2.4 Move `internal/validation/result.go` to `internal/domain/result.go`
- [x] 2.5 Move `internal/validation/opencode/` to `internal/domain/opencode/` (preserve subdirectory)
- [x] 2.6 Move `internal/application/results.go` to `internal/domain/results.go`
- [x] 2.7 Verify `internal/application/requests.go` remains in place (InitializeRequest has library dependency)
- [x] 2.8 Create `internal/domain/doc.go` with package documentation

## 3. Update Import Paths

- [x] 3.1 Update all imports from `internal/models/canonical` to `internal/domain`
- [x] 3.2 Update all imports from `internal/errors` to `internal/domain`
- [x] 3.3 Update all imports from `internal/validation` to `internal/domain` and `internal/domain/opencode`
- [x] 3.4 Update all imports from `internal/application` results to `internal/domain`
- [x] 3.5 Verify compilation with `go build ./...`

## 4. Move Tests

- [x] 4.1 Move `internal/models/canonical/*_test.go` to `internal/domain/`
- [x] 4.2 Move `internal/errors/*_test.go` to `internal/domain/`
- [x] 4.3 Move `internal/validation/*_test.go` to `internal/domain/` (except opencode/ tests)
- [x] 4.4 Move `internal/validation/opencode/*_test.go` to `internal/domain/opencode/`
- [x] 4.5 Update all test file import paths
- [x] 4.6 Verify tests pass with `go test ./...`

## 5. Add Domain Purity Enforcement

- [x] 5.1 Add `depguard` to `linters.enable` list in `.golangci.yml`
- [x] 5.2 Add depguard rule for domain purity (allow only stdlib and internal/domain)
- [x] 5.3 Run `golangci-lint run` to verify no domain purity violations

## 6. Cleanup Old Directories

- [x] 6.1 Remove empty `internal/errors/` directory
- [x] 6.2 Remove empty `internal/validation/` directory
- [x] 6.3 Remove empty `internal/models/canonical/` directory
- [x] 6.4 Remove empty `internal/models/` directory (if empty after canonical removal)

## 7. Update Documentation

- [ ] 7.1 Create `internal/domain/AGENTS.md` with domain layer documentation
- [x] 7.2 Update `internal/application/AGENTS.md` (remove requests/results references)
- [x] 7.3 Update root `AGENTS.md` architecture diagram

## 8. Final Verification

- [x] 8.1 Run `go build ./...` to verify compilation
- [x] 8.2 Run `go test ./...` to verify all tests pass
- [x] 8.3 Run `golangci-lint run` to verify no linting errors
- [x] 8.4 Run `mise run check` to verify full validation passes

## 1. Create Package Structure

- [ ] 1.1 Create `internal/domain/` directory

## 2. Move Domain Types

- [ ] 2.1 Split `internal/models/canonical/models.go` into type-specific files in `internal/domain/`:
  - [ ] 2.1a Move Agent types to `internal/domain/agent.go`
  - [ ] 2.1b Move Command types to `internal/domain/command.go`
  - [ ] 2.1c Move Skill types to `internal/domain/skill.go`
  - [ ] 2.1d Move Memory types to `internal/domain/memory.go`
  - [ ] 2.1e Move Platform/PermissionPolicy types to `internal/domain/platform.go`
- [ ] 2.2 Move `internal/errors/` to `internal/domain/errors.go`
- [ ] 2.3 Move `internal/validation/validators.go` to `internal/domain/validation.go`
- [ ] 2.4 Move `internal/validation/result.go` to `internal/domain/result.go`
- [ ] 2.5 Move `internal/validation/opencode/` to `internal/domain/opencode/` (preserve subdirectory)
- [ ] 2.6 Move `internal/application/results.go` to `internal/domain/results.go`
- [ ] 2.7 Verify `internal/application/requests.go` remains in place (InitializeRequest has library dependency)
- [ ] 2.8 Create `internal/domain/doc.go` with package documentation

## 3. Update Import Paths

- [ ] 3.1 Update all imports from `internal/models/canonical` to `internal/domain`
- [ ] 3.2 Update all imports from `internal/errors` to `internal/domain`
- [ ] 3.3 Update all imports from `internal/validation` to `internal/domain` and `internal/domain/opencode`
- [ ] 3.4 Update all imports from `internal/application` results to `internal/domain`
- [ ] 3.5 Verify compilation with `go build ./...`

## 4. Move Tests

- [ ] 4.1 Move `internal/models/canonical/*_test.go` to `internal/domain/`
- [ ] 4.2 Move `internal/errors/*_test.go` to `internal/domain/`
- [ ] 4.3 Move `internal/validation/*_test.go` to `internal/domain/` (except opencode/ tests)
- [ ] 4.4 Move `internal/validation/opencode/*_test.go` to `internal/domain/opencode/`
- [ ] 4.5 Update all test file import paths
- [ ] 4.6 Verify tests pass with `go test ./...`

## 5. Add Domain Purity Enforcement

- [ ] 5.1 Add `depguard` to `linters.enable` list in `.golangci.yml`
- [ ] 5.2 Add depguard rule for domain purity (allow only stdlib and internal/domain)
- [ ] 5.3 Run `golangci-lint run` to verify no domain purity violations

## 6. Cleanup Old Directories

- [ ] 6.1 Remove empty `internal/errors/` directory
- [ ] 6.2 Remove empty `internal/validation/` directory
- [ ] 6.3 Remove empty `internal/models/canonical/` directory
- [ ] 6.4 Remove empty `internal/models/` directory (if empty after canonical removal)

## 7. Update Documentation

- [ ] 7.1 Create `internal/domain/AGENTS.md` with domain layer documentation
- [ ] 7.2 Update `internal/application/AGENTS.md` (remove requests/results references)
- [ ] 7.3 Update root `AGENTS.md` architecture diagram

## 8. Final Verification

- [ ] 8.1 Run `go build ./...` to verify compilation
- [ ] 8.2 Run `go test ./...` to verify all tests pass
- [ ] 8.3 Run `golangci-lint run` to verify no linting errors
- [ ] 8.4 Run `mise run check` to verify full validation passes

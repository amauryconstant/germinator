## 1. Linter Configuration - Update golangci.yml

- [ ] 1.1 Add essential linters: `staticcheck`, `typecheck`
- [ ] 1.2 Add code quality linters: `gocyclo`, `gocognit`, `funlen` with thresholds
- [ ] 1.3 Add error handling linters: `errorlint`, `wrapcheck`, `errname`
- [ ] 1.4 Add performance linters: `prealloc`, `perfsprint`
- [ ] 1.5 Add security linter: `gosec` with exclusions
- [ ] 1.6 Add test linters: `testifylint`, `tparallel`, `thelper`
- [ ] 1.7 Add best practices linters: `nakedret`, `unconvert`, `unparam`, `wastedassign`
- [ ] 1.8 Verify depguard domain purity rule is properly configured for `internal/domain/`
- [ ] 1.9 Configure test file exclusions for complexity linters (`funlen`, `gocyclo`, `gocognit`)
- [ ] 1.10 Configure wrapcheck to ignore `internal/domain` package

## 2. Linter Configuration - Fix Errors

- [ ] 2.1 Run `golangci-lint run` and categorize all errors
- [ ] 2.2 Fix all errcheck errors
- [ ] 2.3 Fix all staticcheck errors
- [ ] 2.4 Fix all typecheck errors
- [ ] 2.5 Fix all gocyclo/gocognit errors (refactor complex functions)
- [ ] 2.6 Fix all funlen errors (split long functions)
- [ ] 2.7 Fix all errorlint errors
- [ ] 2.8 Fix all wrapcheck errors
- [ ] 2.9 Fix all gosec errors or add justified exclusions
- [ ] 2.10 Fix all testifylint errors
- [ ] 2.11 Fix all other linter errors
- [ ] 2.12 Verify `golangci-lint run` passes cleanly

## 3. Documentation Update

- [ ] 3.1 Document linter categories and thresholds in `.golangci.yml` comments
- [ ] 3.2 Document GoSec and wrapcheck exclusion rationale in `.golangci.yml` comments

## 4. Final Verification

- [ ] 4.1 Run `go build ./...` to verify compilation
- [ ] 4.2 Run `go test ./...` to verify all tests pass
- [ ] 4.3 Run `mise run test:e2e` to verify E2E tests pass
- [ ] 4.4 Run `golangci-lint run` to verify no linting errors
- [ ] 4.5 Run `mise run check` to verify full validation passes

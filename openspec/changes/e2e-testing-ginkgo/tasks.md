## 1. Dependencies and Setup

- [ ] 1.1 Add Ginkgo v2, Gomega, and gexec dependencies to go.mod
- [ ] 1.2 Create test/e2e/ directory structure

## 2. Test Suite Infrastructure

- [ ] 2.1 Create test/e2e/e2e_suite_test.go with BeforeSuite/AfterSuite
- [ ] 2.2 Implement binary build in BeforeSuite (build to bin/germinator-e2e)
- [ ] 2.3 Implement artifact cleanup in AfterSuite using gexec.CleanupBuildArtifacts()

## 3. CLI Helper

- [ ] 3.1 Create test/e2e/helpers/cli_helper.go with GerminatorCLI struct
- [ ] 3.2 Implement NewGerminatorCLI() constructor with binary path resolution
- [ ] 3.3 Implement Run(args ...string) method returning gexec.Session
- [ ] 3.4 Implement ShouldSucceed(session) assertion helper
- [ ] 3.5 Implement ShouldFailWithExit(session, code) assertion helper
- [ ] 3.6 Implement ShouldOutput(session, expected) assertion helper
- [ ] 3.7 Implement ShouldErrorOutput(session, expected) assertion helper

## 4. Test Fixtures

- [ ] 4.1 Create test/e2e/fixtures/document_fixtures.go with fixture management
- [ ] 4.2 Add valid document fixture for testing success cases
- [ ] 4.3 Add invalid document fixture for testing validation errors
- [ ] 4.4 Add nonexistent file path helper for file-not-found tests

## 5. Validate Command Tests

- [ ] 5.1 Create test/e2e/validate_test.go
- [ ] 5.2 Test: validate valid document succeeds (exit 0, "Document is valid")
- [ ] 5.3 Test: validate without --platform flag fails (exit 1)
- [ ] 5.4 Test: validate nonexistent file fails (exit 1)
- [ ] 5.5 Test: validate with invalid platform fails (exit 1)
- [ ] 5.6 Test: validate invalid document fails with validation errors (exit 1)

## 6. Adapt Command Tests

- [ ] 6.1 Create test/e2e/adapt_test.go
- [ ] 6.2 Test: adapt valid document succeeds (exit 0, output file created)
- [ ] 6.3 Test: adapt without --platform flag fails (exit 1)
- [ ] 6.4 Test: adapt nonexistent file fails (exit 1)

## 7. Version Command Tests

- [ ] 7.1 Create test/e2e/version_test.go
- [ ] 7.2 Test: version displays version info matching pattern

## 8. Root Command Tests

- [ ] 8.1 Create test/e2e/root_test.go
- [ ] 8.2 Test: root command without args shows help (exit 0)
- [ ] 8.3 Test: --help flag shows help (exit 0)

## 9. Mise Tasks

- [ ] 9.1 Add test:e2e task to .mise/config.toml
- [ ] 9.2 Add test:full task to .mise/config.toml (unit tests + E2E)

## 10. Verification

- [ ] 10.1 Run mise run test:e2e and verify all tests pass
- [ ] 10.2 Run mise run test:full and verify all tests pass
- [ ] 10.3 Run mise run check and verify no issues

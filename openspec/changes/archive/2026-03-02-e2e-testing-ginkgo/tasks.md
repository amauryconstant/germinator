## 1. Dependencies and Setup

- [x] 1.1 Add Ginkgo v2, Gomega, and gexec dependencies to go.mod
- [x] 1.2 Create test/e2e/ directory structure

## 2. Test Suite Infrastructure

- [x] 2.1 Create test/e2e/e2e_suite_test.go with BeforeSuite/AfterSuite
- [x] 2.2 Implement binary build in BeforeSuite (build to bin/germinator-e2e)
- [x] 2.3 Implement artifact cleanup in AfterSuite using gexec.CleanupBuildArtifacts()

## 3. CLI Helper

- [x] 3.1 Create test/e2e/helpers/cli_helper.go with GerminatorCLI struct
- [x] 3.2 Implement NewGerminatorCLI() constructor with binary path resolution
- [x] 3.3 Implement Run(args ...string) method returning gexec.Session
- [x] 3.4 Implement ShouldSucceed(session) assertion helper
- [x] 3.5 Implement ShouldFailWithExit(session, code) assertion helper
- [x] 3.6 Implement ShouldOutput(session, expected) assertion helper
- [x] 3.7 Implement ShouldErrorOutput(session, expected) assertion helper

## 4. Test Fixtures

- [x] 4.1 Create test/e2e/fixtures/document_fixtures.go with fixture management
- [x] 4.2 Add valid document fixture for testing success cases
- [x] 4.3 Add invalid document fixture for testing validation errors
- [x] 4.4 Add nonexistent file path helper for file-not-found tests

## 5. Validate Command Tests

- [x] 5.1 Create test/e2e/validate_test.go
- [x] 5.2 Test: validate valid document succeeds (exit 0, "Document is valid")
- [x] 5.3 Test: validate without --platform flag fails (exit 1)
- [x] 5.4 Test: validate nonexistent file fails (exit > 0)
- [x] 5.5 Test: validate with invalid platform fails (exit > 0)
- [x] 5.6 Test: validate invalid document fails with validation errors (exit > 0)

## 6. Adapt Command Tests

- [x] 6.1 Create test/e2e/adapt_test.go
- [x] 6.2 Test: adapt valid document succeeds (exit 0, output file created)
- [x] 6.3 Test: adapt without --platform flag fails (exit 1)
- [x] 6.4 Test: adapt nonexistent file fails (exit > 0)

## 7. Version Command Tests

- [x] 7.1 Create test/e2e/version_test.go
- [x] 7.2 Test: version displays version info

## 8. Root Command Tests

- [x] 8.1 Create test/e2e/root_test.go
- [x] 8.2 Test: root command without args shows help (exit 0)
- [x] 8.3 Test: --help flag shows help (exit 0)

## 9. Mise Tasks

- [x] 9.1 Add test:e2e task to .mise/config.toml
- [x] 9.2 Add test:full task to .mise/config.toml (unit tests + E2E)

## 10. Verification

- [x] 10.1 Run mise run test:e2e and verify all tests pass
- [x] 10.2 Run mise run test:full and verify all tests pass
- [x] 10.3 Run mise run check and verify no issues

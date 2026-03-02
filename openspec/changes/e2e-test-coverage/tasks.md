## 1. Canonicalize Command E2E Tests

- [x] 1.1 Create test/e2e/canonicalize_test.go with suite setup
- [x] 1.2 Add success scenario: canonicalize valid document with both platforms
- [x] 1.3 Add error scenario: canonicalize without --platform flag
- [x] 1.4 Add error scenario: canonicalize without --type flag
- [x] 1.5 Add error scenario: canonicalize with invalid platform
- [x] 1.6 Add error scenario: canonicalize with invalid type
- [x] 1.7 Add error scenario: canonicalize nonexistent file

## 2. Platform Parity for Validate Command

- [x] 2.1 Add claude-code success scenario to validate_test.go
- [x] 2.2 Add claude-code nonexistent file scenario to validate_test.go
- [x] 2.3 Add claude-code invalid document scenario to validate_test.go

## 3. Platform Parity for Adapt Command

- [x] 3.1 Add claude-code success scenario to adapt_test.go
- [x] 3.2 Add claude-code nonexistent file scenario to adapt_test.go

## 4. Verification

- [x] 4.1 Run mise run test:e2e to verify all tests pass
- [x] 4.2 Run mise run check to verify lint and format

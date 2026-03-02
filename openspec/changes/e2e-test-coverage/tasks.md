## 1. Canonicalize Command E2E Tests

- [ ] 1.1 Create test/e2e/canonicalize_test.go with suite setup
- [ ] 1.2 Add success scenario: canonicalize valid document with both platforms
- [ ] 1.3 Add error scenario: canonicalize without --platform flag
- [ ] 1.4 Add error scenario: canonicalize without --type flag
- [ ] 1.5 Add error scenario: canonicalize with invalid platform
- [ ] 1.6 Add error scenario: canonicalize with invalid type
- [ ] 1.7 Add error scenario: canonicalize nonexistent file

## 2. Platform Parity for Validate Command

- [ ] 2.1 Add claude-code success scenario to validate_test.go
- [ ] 2.2 Add claude-code nonexistent file scenario to validate_test.go
- [ ] 2.3 Add claude-code invalid document scenario to validate_test.go

## 3. Platform Parity for Adapt Command

- [ ] 3.1 Add claude-code success scenario to adapt_test.go
- [ ] 3.2 Add claude-code nonexistent file scenario to adapt_test.go

## 4. Verification

- [ ] 4.1 Run mise run test:e2e to verify all tests pass
- [ ] 4.2 Run mise run check to verify lint and format

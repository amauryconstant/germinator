
## 2026-03-04 - PHASE2 Verification

- [ ] **[cosmetic]** Rename misleading test "agent pipeline stops on first error"
  - Location: internal/validation/validators_test.go:401-413
  - Impact: Low
  - Notes: Test name suggests early exit verification but just checks failure case

- [ ] **[cosmetic]** Rename misleading test in OpenCode validators
  - Location: internal/validation/opencode/validators_test.go:209-225
  - Impact: Low
  - Notes: Same issue - test named "stops on first error" but doesn't verify early exit

- [ ] **[docs]** Update validation-pipeline spec to match implementation
  - Location: openspec/changes/add-validation-pipeline/specs/validation-pipeline/spec.md
  - Impact: Low
  - Notes: Spec says "early exit" but implementation "collects all errors" which is better UX

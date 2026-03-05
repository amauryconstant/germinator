
## 2026-03-05 - PHASE2 Verification

- [ ] **[cosmetic]** Consider adding Context to Error() method output
  - Location: internal/errors/types.go
  - Impact: Low
  - Notes: While CLI output via error_formatter includes context, the Error() method does not. This could cause confusion when errors are logged or printed directly without going through the formatter.

- [ ] **[docs]** Update spec to reflect actual Error() implementation
  - Location: openspec/changes/enhance-all-errors/specs/enhanced-errors/spec.md:277-291
  - Impact: Low
  - Notes: Spec says Error() should include context and use "Hint:" format, but implementation uses 💡 emoji and doesn't include context. CLI output is correct via error_formatter.

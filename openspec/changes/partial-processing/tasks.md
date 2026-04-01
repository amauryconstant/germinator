## 1. Modify Initializer Service (Fail-Fast → Continue on Error)

- [ ] 1.1 Remove early return at line 42 (resource resolution error) - change to append result and continue
- [ ] 1.2 Remove early return at line 51 (parse ref error) - change to append result and continue
- [ ] 1.3 Remove early return at line 58 (get output path error) - change to append result and continue
- [ ] 1.4 Remove early return at line 67 (file exists error) - change to append result and continue
- [ ] 1.5 Remove early return at line 82 (load document error) - change to append result and continue
- [ ] 1.6 Remove early return at line 90 (render document error) - change to append result and continue
- [ ] 1.7 Remove early return at line 98 (mkdir error) - change to append result and continue
- [ ] 1.8 Remove early return at line 105 (write file error) - change to append result and continue
- [ ] 1.9 Add logic to return error only when ALL resources fail

## 2. Update Unit Tests

- [ ] 2.1 Add test for partial success scenario (one resource fails, others succeed)
- [ ] 2.2 Add test for all resources fail scenario (error returned)
- [ ] 2.3 Add test verifying all results are returned regardless of errors
- [ ] 2.4 Update existing fail-fast test to reflect new behavior

## 3. Run Verification

- [ ] 3.1 Run `mise run check` to ensure all validation passes
- [ ] 3.2 Run `mise run test` to verify all unit tests pass

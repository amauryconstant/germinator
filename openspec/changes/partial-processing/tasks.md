## 1. Modify Initializer Service (Fail-Fast → Continue on Error)

- [x] 1.1 Remove early return on resource resolution error - change to append result and continue
- [x] 1.2 Remove early return on parse ref error - change to append result and continue
- [x] 1.3 Remove early return on get output path error - change to append result and continue
- [x] 1.4 Remove early return on file exists error - change to append result and continue
- [x] 1.5 Remove early return on load document error - change to append result and continue
- [x] 1.6 Remove early return on render document error - change to append result and continue
- [x] 1.7 Remove early return on mkdir error - change to append result and continue
- [x] 1.8 Remove early return on write file error - change to append result and continue
- [x] 1.9 Add logic to return error only when ALL resources fail (nil error if at least one succeeds)

## 2. Update Unit Tests

- [x] 2.1 Add test for partial success scenario (one resource fails, others succeed)
- [x] 2.2 Add test for all resources fail scenario (error returned)
- [x] 2.3 Add test verifying all results are returned regardless of errors
- [x] 2.4 Update existing fail-fast test to reflect new behavior

## 3. Update CLI Output (Per-Resource Status and Summary)

- [x] 3.1 Update CLI to display per-resource success/failure status
- [x] 3.2 Add summary line showing "Initialized N resources, M failed"

## 4. Run Verification

- [x] 4.1 Run `mise run check` to ensure all validation passes
- [x] 4.2 Run `mise run test` to verify all unit tests pass

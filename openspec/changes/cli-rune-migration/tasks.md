## 1. CLI Pattern Migration - Error Types

- [ ] 1.1 Rename `ExitCodeParse` to `ExitCodeConfig` in `cmd/error_handler.go`
- [ ] 1.2 Add new exit codes: ExitCodeGit (4), ExitCodeNotFound (6)
- [ ] 1.3 Rename `CategoryParse` to `CategoryConfig`, add `CategoryNotFound`, `CategoryGit`
- [ ] 1.4 Update `CategorizeError()` to handle new error types
- [ ] 1.5 Update `GetExitCodeForError()` with new mappings (Config→3, Git→4, Validation→5, NotFound→6)
- [ ] 1.6 Rename `HandleError()` to `HandleCLIError()` with updated signature accepting `*cobra.Command`
- [ ] 1.7 Remove `HandleValidationErrors()` - validation errors go through `HandleCLIError()`
- [ ] 1.8 Add `IsCobraArgumentError()` helper function for Cobra argument error detection

## 2. CLI Pattern Migration - Commands to RunE

- [ ] 2.1 Update `main.go` to use centralized error handling pattern
- [ ] 2.2 Convert `cmd/validate.go` from `Run` to `RunE`
- [ ] 2.3 Convert `cmd/adapt.go` from `Run` to `RunE`
- [ ] 2.4 Convert `cmd/canonicalize.go` from `Run` to `RunE`
- [ ] 2.5 Convert `cmd/init.go` from `Run` to `RunE`
- [ ] 2.6 Convert `cmd/library.go` and subcommands from `Run` to `RunE`
- [ ] 2.7 Remove `HandleError()` calls from all commands (return errors instead)
- [ ] 2.8 Verify E2E tests pass with correct exit codes
  - [ ] 2.8.1 Verify config errors exit with code 3
  - [ ] 2.8.2 Verify validation errors exit with code 5
  - [ ] 2.8.3 Verify usage errors exit with code 2
  - [ ] 2.8.4 Verify git errors exit with code 4 (when git operations added)
  - [ ] 2.8.5 Verify not-found errors exit with code 6 (when not-found errors occur)

## 3. Final Verification

- [ ] 3.1 Run `go build ./...` to verify compilation
- [ ] 3.2 Run `go test ./...` to verify all tests pass
- [ ] 3.3 Run `mise run test:e2e` to verify E2E tests pass with correct exit codes

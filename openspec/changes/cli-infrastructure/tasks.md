## 1. Error Types Package

- [x] 1.1 Create `internal/errors/types.go` with ParseError struct (Path, Message, Cause fields, Error() and Unwrap() methods)
- [x] 1.2 Add ValidationError struct with Message, Field, and Suggestions fields, plus Suggestions() method
- [x] 1.3 Add TransformError struct with Operation, Platform, Message, and Cause fields
- [x] 1.4 Add FileError struct with Path, Operation, Message, and Cause fields, plus IsNotFound() helper
- [x] 1.5 Add ConfigError struct with Field, Value, Available, and Message fields
- [x] 1.6 Create constructor functions: NewParseError, NewValidationError, NewTransformError, NewFileError, NewConfigError

## 2. Error Types Tests

- [x] 2.1 Create `internal/errors/types_test.go` with table-driven tests for each error type
- [x] 2.2 Test Error() string formatting for each type
- [x] 2.3 Test Unwrap() method returns correct cause
- [x] 2.4 Test errors.As detection for each type
- [x] 2.5 Test FileError.IsNotFound() helper method
- [x] 2.6 Test ValidationError.Suggestions() method

## 3. Exit Codes and Categorization

- [x] 3.1 Create `cmd/error_handler.go` with ExitCode type and constants (Success=0, Error=1, Usage=2, Parse=3)
- [x] 3.2 Define ErrorCategory enum (Cobra, Config, Parse, Validation, Transform, File, Generic)
- [x] 3.3 Implement CategorizeError(err error) ErrorCategory function using errors.As for type detection
- [x] 3.4 Implement GetExitCodeForError(err error) ExitCode mapping function
- [x] 3.5 Implement HandleError(cfg *CommandConfig, err error) function that formats, outputs to stderr, and exits

## 4. Error Formatter

- [x] 4.1 Create `cmd/error_formatter.go` with ErrorFormatter struct and formatters map
- [x] 4.2 Implement NewErrorFormatter() with default formatters for all error types
- [x] 4.3 Implement Format(err error) string method with type lookup using errors.As
- [x] 4.4 Add formatParseError with "Parse error:" prefix and file path
- [x] 4.5 Add formatValidationError with "Validation error:" prefix and Hint: suggestions
- [x] 4.6 Add formatTransformError with "Transform error:" prefix and operation/platform
- [x] 4.7 Add formatFileError with "File error:" prefix, operation, and path
- [x] 4.8 Add formatConfigError with "Config error:" prefix and available options

## 5. Error Formatter Tests

- [x] 5.1 Create `cmd/error_formatter_test.go` with table-driven tests
- [x] 5.2 Test formatting for each error type
- [x] 5.3 Test formatting with wrapped errors (errors.As detection)
- [x] 5.4 Test default formatting for unknown error types
- [x] 5.5 Test ValidationError with and without suggestions

## 6. Verbosity Infrastructure

- [x] 6.1 Create `cmd/verbose.go` with Verbosity type and helper methods (IsVerbose(), IsVeryVerbose())
- [x] 6.2 Add VerbosePrint(cfg *CommandConfig, format string, args ...any) function
- [x] 6.3 Add VeryVerbosePrint(cfg *CommandConfig, format string, args ...any) function with 2-space indentation

## 7. CommandConfig Pattern

- [x] 7.1 Create `cmd/config.go` with CommandConfig struct (ErrorFormatter *ErrorFormatter, Verbosity Verbosity)
- [x] 7.2 Add NewCommandConfig(cmd *cobra.Command) function to build config from command flags
- [x] 7.3 Update `cmd/root.go` to add persistent verbose flag using CountP("verbose", "v", ...)
- [x] 7.4 Update `cmd/root.go` to initialize CommandConfig in PersistentPreRun

## 8. Core Package Integration

- [x] 8.1 Update `internal/core/loader.go` LoadDocument to return typed ParseError for unrecognized filenames
- [x] 8.2 Update LoadDocument to return typed ParseError for YAML parsing failures
- [x] 8.3 Update LoadDocument to return typed FileError for file read failures
- [x] 8.4 Update validatePlatform helper to return typed ConfigError

## 9. Services Package Integration

- [x] 9.1 Update `internal/services/transformer.go` TransformDocument to return typed errors (ParseError, TransformError, FileError)
- [x] 9.2 Update ValidateDocument to return typed ConfigError and ParseError
- [x] 9.3 Update `internal/services/canonicalizer.go` CanonicalizeDocument to return typed errors

## 10. Validate Command Integration

- [x] 10.1 Update `cmd/validate.go` to receive and use CommandConfig
- [x] 10.2 Update validate command to use HandleError instead of fmt.Fprintf + os.Exit(1)
- [x] 10.3 Add level 1 verbose output for validate (file path, platform)
- [x] 10.4 Add level 2 verbose output for validate (loading, parsing, validation details)
- [x] 10.5 Ensure validate command exits with correct semantic codes

## 11. Adapt Command Integration

- [x] 11.1 Update `cmd/adapt.go` to receive and use CommandConfig
- [x] 11.2 Update adapt command to use HandleError instead of fmt.Fprintf + os.Exit(1)
- [x] 11.3 Add level 1 verbose output for adapt (transformation description, output path)
- [x] 11.4 Add level 2 verbose output for adapt (loading, rendering, template details)
- [x] 11.5 Ensure adapt command exits with correct semantic codes

## 12. Canonicalize Command Integration

- [x] 12.1 Update `cmd/canonicalize.go` to receive and use CommandConfig
- [x] 12.2 Update canonicalize command to use HandleError instead of fmt.Fprintf + os.Exit(1)
- [x] 12.3 Add level 1 verbose output for canonicalize (canonicalization description, output path)
- [x] 12.4 Add level 2 verbose output for canonicalize (parsing, validation details)
- [x] 12.5 Ensure canonicalize command exits with correct semantic codes

## 13. Testing

- [x] 13.1 Add unit tests for Verbosity type methods in `cmd/verbose_test.go`
- [x] 13.2 Add E2E tests for validate command with `-v` flag
- [x] 13.3 Add E2E tests for validate command with `-vv` flag
- [x] 13.4 Add E2E tests for adapt command with `-v` flag
- [x] 13.5 Add E2E tests for adapt command with `-vv` flag
- [x] 13.6 Add E2E tests for default behavior (no verbose output)
- [x] 13.7 Add integration tests for exit codes in `cmd_test.go`
- [x] 13.8 Test validate command exits with code 2 for validation errors
- [x] 13.9 Test validate command exits with code 3 for parse errors
- [x] 13.10 Test validate command exits with code 1 for file errors
- [x] 13.11 Test adapt command exit codes for each error category

## 14. Documentation

- [ ] 14.1 Update `cmd/AGENTS.md` with verbose flag pattern documentation
- [ ] 14.2 Update `cmd/AGENTS.md` with CommandConfig pattern documentation
- [ ] 14.3 Update `cmd/AGENTS.md` with error handling pattern documentation

## 15. Final Verification

- [x] 15.1 Run `mise run check` (lint, format, test, build)
- [x] 15.2 Verify backward compatibility - existing error messages remain helpful
- [x] 15.3 Manual testing of error output formatting
- [x] 15.4 Manual testing of verbose output formatting

package cmd

import (
	"errors"
	"fmt"
	"os"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

// ExitCode represents the process exit code.
type ExitCode int

// Exit codes for different error categories.
const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeError   ExitCode = 1
	ExitCodeUsage   ExitCode = 2
	ExitCodeParse   ExitCode = 3
)

// ErrorCategory represents the category of an error.
type ErrorCategory int

// Error categories for classification.
const (
	CategoryCobra ErrorCategory = iota
	CategoryConfig
	CategoryParse
	CategoryValidation
	CategoryTransform
	CategoryFile
	CategoryGeneric
)

// CategorizeError determines the error category based on error type.
func CategorizeError(err error) ErrorCategory {
	var parseErr *gerrors.ParseError
	var validationErr *gerrors.ValidationError
	var transformErr *gerrors.TransformError
	var fileErr *gerrors.FileError
	var configErr *gerrors.ConfigError

	if errors.As(err, &parseErr) {
		return CategoryParse
	}
	if errors.As(err, &validationErr) {
		return CategoryValidation
	}
	if errors.As(err, &transformErr) {
		return CategoryTransform
	}
	if errors.As(err, &fileErr) {
		return CategoryFile
	}
	if errors.As(err, &configErr) {
		return CategoryConfig
	}

	return CategoryGeneric
}

// GetExitCodeForError returns the appropriate exit code for an error.
func GetExitCodeForError(err error) ExitCode {
	category := CategorizeError(err)

	switch category {
	case CategoryParse:
		return ExitCodeParse
	case CategoryConfig, CategoryValidation, CategoryCobra:
		return ExitCodeUsage
	case CategoryTransform, CategoryFile, CategoryGeneric:
		return ExitCodeError
	default:
		return ExitCodeError
	}
}

// HandleError formats and outputs the error, then exits with the appropriate code.
func HandleError(cfg *CommandConfig, err error) {
	fmt.Fprintln(os.Stderr, cfg.ErrorFormatter.Format(err))
	os.Exit(int(GetExitCodeForError(err)))
}

// ValidationResultError wraps multiple validation errors for unified handling.
type ValidationResultError struct {
	Errors []error
}

func (e *ValidationResultError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return e.Errors[0].Error()
}

// HandleValidationErrors formats and outputs multiple validation errors, then exits.
func HandleValidationErrors(cfg *CommandConfig, errs []error) {
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, cfg.ErrorFormatter.Format(e))
	}
	os.Exit(int(ExitCodeUsage))
}

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	gerrors "gitlab.com/amoconst/germinator/internal/domain"
)

// globalCommandConfig holds the CommandConfig for error handling in main.go.
// This is set during root command construction and used by HandleCLIError.
var globalCommandConfig *CommandConfig

// SetGlobalCommandConfig stores the CommandConfig for use in error handling.
// This is called during root command construction.
func SetGlobalCommandConfig(cfg *CommandConfig) {
	globalCommandConfig = cfg
}

// ExitCode represents the process exit code.
type ExitCode int

// Exit codes for different error categories.
const (
	ExitCodeSuccess    ExitCode = 0
	ExitCodeError      ExitCode = 1
	ExitCodeUsage      ExitCode = 2
	ExitCodeConfig     ExitCode = 3
	ExitCodeGit        ExitCode = 4
	ExitCodeValidation ExitCode = 5
	ExitCodeNotFound   ExitCode = 6
)

// ErrorCategory represents the category of an error.
type ErrorCategory int

// Error categories for classification.
const (
	CategoryCobra ErrorCategory = iota
	CategoryConfig
	CategoryValidation
	CategoryTransform
	CategoryFile
	CategoryGit
	CategoryNotFound
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
		return CategoryConfig
	}
	if errors.As(err, &validationErr) {
		return CategoryValidation
	}
	if errors.As(err, &transformErr) {
		return CategoryTransform
	}
	if errors.As(err, &fileErr) {
		if fileErr.IsNotFound() {
			return CategoryNotFound
		}
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
	case CategoryConfig:
		return ExitCodeConfig
	case CategoryValidation:
		return ExitCodeValidation
	case CategoryGit:
		return ExitCodeGit
	case CategoryNotFound:
		return ExitCodeNotFound
	case CategoryCobra:
		return ExitCodeUsage
	case CategoryTransform, CategoryFile, CategoryGeneric:
		return ExitCodeError
	default:
		return ExitCodeError
	}
}

// HandleCLIError formats and outputs the error, then returns the appropriate exit code.
// The caller (main.go) should use this code with os.Exit().
func HandleCLIError(cmd *cobra.Command, err error) ExitCode {
	// Check for Cobra argument errors to provide better UX
	if IsCobraArgumentError(err) {
		// Cobra will have already printed the error
		return ExitCodeUsage
	}

	// Handle ValidationResultError specially to print all errors
	if validationErr, ok := err.(*ValidationResultError); ok {
		if globalCommandConfig != nil {
			for _, e := range validationErr.Errors {
				fmt.Fprintln(os.Stderr, globalCommandConfig.ErrorFormatter.Format(e))
			}
		} else {
			// Fallback to basic formatting
			for _, e := range validationErr.Errors {
				fmt.Fprintln(os.Stderr, e.Error())
			}
		}
		return ExitCodeValidation
	}

	// Format and print the error using the global config
	if globalCommandConfig != nil {
		fmt.Fprintln(os.Stderr, globalCommandConfig.ErrorFormatter.Format(err))
	} else {
		// Fallback to basic formatting if config is not available
		fmt.Fprintln(os.Stderr, err.Error())
	}

	// Return the appropriate exit code
	return GetExitCodeForError(err)
}

// IsCobraArgumentError detects if an error is a Cobra argument validation error.
// These errors have already been printed by Cobra and should just return ExitCodeUsage.
func IsCobraArgumentError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "accepts") ||
		strings.Contains(errStr, "requires") ||
		strings.Contains(errStr, "at least") ||
		strings.Contains(errStr, "at most") ||
		strings.Contains(errStr, "unknown flag") ||
		strings.Contains(errStr, "invalid argument")
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

// Unwrap returns the first error for compatibility with error chain support.
func (e *ValidationResultError) Unwrap() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}

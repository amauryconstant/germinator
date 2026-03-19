package cmd

import (
	"errors"
	"fmt"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/domain"
)

type formatterFunc func(error) string

// ErrorFormatter formats errors based on their type.
type ErrorFormatter struct {
	formatters []struct {
		match     func(error) bool
		formatter formatterFunc
	}
}

// NewErrorFormatter creates a new ErrorFormatter with default formatters registered.
func NewErrorFormatter() *ErrorFormatter {
	f := &ErrorFormatter{}
	f.registerDefaultFormatters()
	return f
}

// Format returns a formatted string for the given error.
func (f *ErrorFormatter) Format(err error) string {
	for _, entry := range f.formatters {
		if entry.match(err) {
			return entry.formatter(err)
		}
	}
	return f.defaultFormat(err)
}

func (f *ErrorFormatter) defaultFormat(err error) string {
	return fmt.Sprintf("Error: %s\n", err.Error())
}

func (f *ErrorFormatter) registerDefaultFormatters() {
	f.formatters = []struct {
		match     func(error) bool
		formatter formatterFunc
	}{
		{match: isParseError, formatter: formatParseError},
		{match: isValidationError, formatter: formatValidationError},
		{match: isTransformError, formatter: formatTransformError},
		{match: isFileError, formatter: formatFileError},
		{match: isConfigError, formatter: formatConfigError},
	}
}

func isParseError(err error) bool {
	var e *gerrors.ParseError
	return errors.As(err, &e)
}

func isValidationError(err error) bool {
	var e *gerrors.ValidationError
	return errors.As(err, &e)
}

func isTransformError(err error) bool {
	var e *gerrors.TransformError
	return errors.As(err, &e)
}

func isFileError(err error) bool {
	var e *gerrors.FileError
	return errors.As(err, &e)
}

func isConfigError(err error) bool {
	var e *gerrors.ConfigError
	return errors.As(err, &e)
}

func formatParseError(err error) string {
	var parseErr *gerrors.ParseError
	if !errors.As(err, &parseErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Parse error: %s\n", parseErr.Message())
	fmt.Fprintf(&sb, "  File: %s\n", parseErr.Path())
	if parseErr.Cause() != nil {
		fmt.Fprintf(&sb, "  Cause: %s\n", parseErr.Cause().Error())
	}
	for _, suggestion := range parseErr.Suggestions() {
		fmt.Fprintf(&sb, "  Hint: %s\n", suggestion)
	}
	if parseErr.Context() != "" {
		fmt.Fprintf(&sb, "  Context: %s\n", parseErr.Context())
	}
	return sb.String()
}

func formatValidationError(err error) string {
	var validationErr *gerrors.ValidationError
	if !errors.As(err, &validationErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	if validationErr.Field() != "" {
		fmt.Fprintf(&sb, "Validation error: %s (field: %s)\n", validationErr.Message(), validationErr.Field())
	} else {
		fmt.Fprintf(&sb, "Validation error: %s\n", validationErr.Message())
	}

	for _, suggestion := range validationErr.Suggestions() {
		fmt.Fprintf(&sb, "  Hint: %s\n", suggestion)
	}
	return sb.String()
}

func formatTransformError(err error) string {
	var transformErr *gerrors.TransformError
	if !errors.As(err, &transformErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	if transformErr.Platform() != "" {
		fmt.Fprintf(&sb, "Transform error (%s for %s): %s\n", transformErr.Operation(), transformErr.Platform(), transformErr.Message())
	} else {
		fmt.Fprintf(&sb, "Transform error (%s): %s\n", transformErr.Operation(), transformErr.Message())
	}
	if transformErr.Cause() != nil {
		fmt.Fprintf(&sb, "  Cause: %s\n", transformErr.Cause().Error())
	}
	for _, suggestion := range transformErr.Suggestions() {
		fmt.Fprintf(&sb, "  Hint: %s\n", suggestion)
	}
	if transformErr.Context() != "" {
		fmt.Fprintf(&sb, "  Context: %s\n", transformErr.Context())
	}
	return sb.String()
}

func formatFileError(err error) string {
	var fileErr *gerrors.FileError
	if !errors.As(err, &fileErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "File error (%s): %s\n", fileErr.Operation(), fileErr.Message())
	fmt.Fprintf(&sb, "  Path: %s\n", fileErr.Path())
	if fileErr.Cause() != nil {
		fmt.Fprintf(&sb, "  Cause: %s\n", fileErr.Cause().Error())
	}
	for _, suggestion := range fileErr.Suggestions() {
		fmt.Fprintf(&sb, "  Hint: %s\n", suggestion)
	}
	if fileErr.Context() != "" {
		fmt.Fprintf(&sb, "  Context: %s\n", fileErr.Context())
	}
	return sb.String()
}

func formatConfigError(err error) string {
	var configErr *gerrors.ConfigError
	if !errors.As(err, &configErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Config error: %s\n", configErr.Message())
	if len(configErr.Suggestions()) > 0 {
		fmt.Fprintf(&sb, "  Hint: %s\n", strings.Join(configErr.Suggestions(), ", "))
	}
	if configErr.Field() != "" && configErr.Value() != "" {
		fmt.Fprintf(&sb, "  Field: %s, Value: %s\n", configErr.Field(), configErr.Value())
	}
	if configErr.Context() != "" {
		fmt.Fprintf(&sb, "  Context: %s\n", configErr.Context())
	}
	return sb.String()
}

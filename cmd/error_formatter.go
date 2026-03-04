package cmd

import (
	"errors"
	"fmt"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
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
	sb.WriteString(fmt.Sprintf("Parse error: %s\n", parseErr.Message))
	sb.WriteString(fmt.Sprintf("  File: %s\n", parseErr.Path))
	if parseErr.Cause != nil {
		sb.WriteString(fmt.Sprintf("  Cause: %s\n", parseErr.Cause.Error()))
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
		sb.WriteString(fmt.Sprintf("Validation error: %s (field: %s)\n", validationErr.Message(), validationErr.Field()))
	} else {
		sb.WriteString(fmt.Sprintf("Validation error: %s\n", validationErr.Message()))
	}

	for _, suggestion := range validationErr.Suggestions() {
		sb.WriteString(fmt.Sprintf("  Hint: %s\n", suggestion))
	}
	return sb.String()
}

func formatTransformError(err error) string {
	var transformErr *gerrors.TransformError
	if !errors.As(err, &transformErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	if transformErr.Platform != "" {
		sb.WriteString(fmt.Sprintf("Transform error (%s for %s): %s\n", transformErr.Operation, transformErr.Platform, transformErr.Message))
	} else {
		sb.WriteString(fmt.Sprintf("Transform error (%s): %s\n", transformErr.Operation, transformErr.Message))
	}
	if transformErr.Cause != nil {
		sb.WriteString(fmt.Sprintf("  Cause: %s\n", transformErr.Cause.Error()))
	}
	return sb.String()
}

func formatFileError(err error) string {
	var fileErr *gerrors.FileError
	if !errors.As(err, &fileErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File error (%s): %s\n", fileErr.Operation, fileErr.Message))
	sb.WriteString(fmt.Sprintf("  Path: %s\n", fileErr.Path))
	if fileErr.Cause != nil {
		sb.WriteString(fmt.Sprintf("  Cause: %s\n", fileErr.Cause.Error()))
	}
	return sb.String()
}

func formatConfigError(err error) string {
	var configErr *gerrors.ConfigError
	if !errors.As(err, &configErr) {
		return fmt.Sprintf("Error: %s\n", err.Error())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Config error: %s\n", configErr.Message))
	if len(configErr.Available) > 0 {
		sb.WriteString(fmt.Sprintf("  Available: %s\n", strings.Join(configErr.Available, ", ")))
	}
	if configErr.Field != "" && configErr.Value != "" {
		sb.WriteString(fmt.Sprintf("  Field: %s, Value: %s\n", configErr.Field, configErr.Value))
	}
	return sb.String()
}

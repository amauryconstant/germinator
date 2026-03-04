package errors

import (
	"fmt"
	"strings"
)

type ParseError struct {
	Path    string
	Message string
	Cause   error
}

func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("parse error in %s: %s: %v", e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("parse error in %s: %s", e.Path, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// ValidationError represents a validation failure with immutable builders for fluent construction.
type ValidationError struct {
	request     string
	field       string
	value       string
	message     string
	suggestions []string
	context     string
}

// NewValidationError creates a new ValidationError with the given parameters.
func NewValidationError(request, field, value, message string) *ValidationError {
	return &ValidationError{
		request:     request,
		field:       field,
		value:       value,
		message:     message,
		suggestions: nil,
		context:     "",
	}
}

// WithSuggestions returns a new ValidationError with the given suggestions (immutable builder).
func (e *ValidationError) WithSuggestions(suggestions []string) *ValidationError {
	return &ValidationError{
		request:     e.request,
		field:       e.field,
		value:       e.value,
		message:     e.message,
		suggestions: suggestions,
		context:     e.context,
	}
}

// WithContext returns a new ValidationError with the given context (immutable builder).
func (e *ValidationError) WithContext(context string) *ValidationError {
	return &ValidationError{
		request:     e.request,
		field:       e.field,
		value:       e.value,
		message:     e.message,
		suggestions: e.suggestions,
		context:     context,
	}
}

// Field returns the field name that failed validation.
func (e *ValidationError) Field() string {
	return e.field
}

// Value returns the invalid value that failed validation.
func (e *ValidationError) Value() string {
	return e.value
}

// Message returns the validation error message.
func (e *ValidationError) Message() string {
	return e.message
}

// Request returns the request type context.
func (e *ValidationError) Request() string {
	return e.request
}

// Suggestions returns a copy of the suggestions slice.
func (e *ValidationError) Suggestions() []string {
	if e.suggestions == nil {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns additional context information.
func (e *ValidationError) Context() string {
	return e.context
}

// Error formats the validation error as a string.
func (e *ValidationError) Error() string {
	var parts []string

	if e.request != "" && e.field != "" {
		parts = append(parts, fmt.Sprintf("validation failed for %s.%s", e.request, e.field))
	} else if e.field != "" {
		parts = append(parts, fmt.Sprintf("validation failed for field '%s'", e.field))
	} else {
		parts = append(parts, "validation failed")
	}

	if e.message != "" {
		parts = append(parts, e.message)
	}

	result := strings.Join(parts, ": ")

	if e.value != "" {
		result += fmt.Sprintf(" (value: %s)", e.value)
	}

	if len(e.suggestions) > 0 {
		for _, suggestion := range e.suggestions {
			result += fmt.Sprintf("\n💡 %s", suggestion)
		}
	}

	return result
}

type TransformError struct {
	Operation string
	Platform  string
	Message   string
	Cause     error
}

func (e *TransformError) Error() string {
	if e.Platform != "" {
		if e.Cause != nil {
			return fmt.Sprintf("transform error (%s for %s): %s: %v", e.Operation, e.Platform, e.Message, e.Cause)
		}
		return fmt.Sprintf("transform error (%s for %s): %s", e.Operation, e.Platform, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("transform error (%s): %s: %v", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("transform error (%s): %s", e.Operation, e.Message)
}

func (e *TransformError) Unwrap() error {
	return e.Cause
}

type FileError struct {
	Path      string
	Operation string
	Message   string
	Cause     error
}

func (e *FileError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("file error (%s %s): %s: %v", e.Operation, e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("file error (%s %s): %s", e.Operation, e.Path, e.Message)
}

func (e *FileError) Unwrap() error {
	return e.Cause
}

func (e *FileError) IsNotFound() bool {
	msg := strings.ToLower(e.Message)
	if strings.Contains(msg, "not found") || strings.Contains(msg, "does not exist") || strings.Contains(msg, "no such file") {
		return true
	}
	if e.Cause != nil {
		causeMsg := strings.ToLower(e.Cause.Error())
		return strings.Contains(causeMsg, "not found") || strings.Contains(causeMsg, "does not exist") || strings.Contains(causeMsg, "no such file")
	}
	return false
}

type ConfigError struct {
	Field     string
	Value     string
	Available []string
	Message   string
}

func (e *ConfigError) Error() string {
	if len(e.Available) > 0 {
		return fmt.Sprintf("config error: %s (available: %s)", e.Message, strings.Join(e.Available, ", "))
	}
	if e.Field != "" && e.Value != "" {
		return fmt.Sprintf("config error: invalid %s '%s': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func NewParseError(path, message string, cause error) *ParseError {
	return &ParseError{
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

func NewTransformError(operation, platform, message string, cause error) *TransformError {
	return &TransformError{
		Operation: operation,
		Platform:  platform,
		Message:   message,
		Cause:     cause,
	}
}

func NewFileError(path, operation, message string, cause error) *FileError {
	return &FileError{
		Path:      path,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

func NewConfigError(field, value string, available []string, message string) *ConfigError {
	return &ConfigError{
		Field:     field,
		Value:     value,
		Available: available,
		Message:   message,
	}
}

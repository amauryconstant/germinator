package errors

import (
	"fmt"
	"strings"
)

// ParseError represents a parsing failure with immutable builders for fluent construction.
type ParseError struct {
	path        string
	message     string
	cause       error
	suggestions []string
	context     string
}

// NewParseError creates a new ParseError with the given parameters.
func NewParseError(path, message string, cause error) *ParseError {
	return &ParseError{
		path:        path,
		message:     message,
		cause:       cause,
		suggestions: nil,
		context:     "",
	}
}

// WithSuggestions returns a new ParseError with the given suggestions (immutable builder).
func (e *ParseError) WithSuggestions(suggestions []string) *ParseError {
	return &ParseError{
		path:        e.path,
		message:     e.message,
		cause:       e.cause,
		suggestions: suggestions,
		context:     e.context,
	}
}

// WithContext returns a new ParseError with the given context (immutable builder).
func (e *ParseError) WithContext(context string) *ParseError {
	return &ParseError{
		path:        e.path,
		message:     e.message,
		cause:       e.cause,
		suggestions: e.suggestions,
		context:     context,
	}
}

// Path returns the file path where the parse error occurred.
func (e *ParseError) Path() string {
	return e.path
}

// Message returns the parse error message.
func (e *ParseError) Message() string {
	return e.message
}

// Cause returns the underlying error that caused the parse failure.
func (e *ParseError) Cause() error {
	return e.cause
}

// Suggestions returns a copy of the suggestions slice.
func (e *ParseError) Suggestions() []string {
	if e.suggestions == nil {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns additional context information.
func (e *ParseError) Context() string {
	return e.context
}

// Error formats the parse error as a string.
func (e *ParseError) Error() string {
	var parts []string

	if e.path != "" {
		parts = append(parts, fmt.Sprintf("parse error in %s", e.path))
	} else {
		parts = append(parts, "parse error")
	}

	if e.message != "" {
		parts = append(parts, e.message)
	}

	result := strings.Join(parts, ": ")

	if e.cause != nil {
		result += fmt.Sprintf(": %v", e.cause)
	}

	if len(e.suggestions) > 0 {
		for _, suggestion := range e.suggestions {
			result += fmt.Sprintf("\n💡 %s", suggestion)
		}
	}

	return result
}

// Unwrap returns the underlying cause for error chain support.
func (e *ParseError) Unwrap() error {
	return e.cause
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

// Unwrap returns nil (validation errors don't wrap other errors).
// Provided for API consistency with other error types.
func (e *ValidationError) Unwrap() error {
	return nil
}

type TransformError struct {
	operation   string
	platform    string
	message     string
	cause       error
	suggestions []string
	context     string
}

// NewTransformError creates a new TransformError with the given parameters.
func NewTransformError(operation, platform, message string, cause error) *TransformError {
	return &TransformError{
		operation:   operation,
		platform:    platform,
		message:     message,
		cause:       cause,
		suggestions: nil,
		context:     "",
	}
}

// WithSuggestions returns a new TransformError with the given suggestions (immutable builder).
func (e *TransformError) WithSuggestions(suggestions []string) *TransformError {
	return &TransformError{
		operation:   e.operation,
		platform:    e.platform,
		message:     e.message,
		cause:       e.cause,
		suggestions: suggestions,
		context:     e.context,
	}
}

// WithContext returns a new TransformError with the given context (immutable builder).
func (e *TransformError) WithContext(context string) *TransformError {
	return &TransformError{
		operation:   e.operation,
		platform:    e.platform,
		message:     e.message,
		cause:       e.cause,
		suggestions: e.suggestions,
		context:     context,
	}
}

// Operation returns the operation that failed.
func (e *TransformError) Operation() string {
	return e.operation
}

// Platform returns the target platform.
func (e *TransformError) Platform() string {
	return e.platform
}

// Message returns the transform error message.
func (e *TransformError) Message() string {
	return e.message
}

// Cause returns the underlying error that caused the transform failure.
func (e *TransformError) Cause() error {
	return e.cause
}

// Suggestions returns a copy of the suggestions slice.
func (e *TransformError) Suggestions() []string {
	if e.suggestions == nil {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns additional context information.
func (e *TransformError) Context() string {
	return e.context
}

// Error formats the transform error as a string.
func (e *TransformError) Error() string {
	var parts []string

	if e.platform != "" {
		parts = append(parts, fmt.Sprintf("transform error (%s for %s)", e.operation, e.platform))
	} else {
		parts = append(parts, fmt.Sprintf("transform error (%s)", e.operation))
	}

	if e.message != "" {
		parts = append(parts, e.message)
	}

	result := strings.Join(parts, ": ")

	if e.cause != nil {
		result += fmt.Sprintf(": %v", e.cause)
	}

	if len(e.suggestions) > 0 {
		for _, suggestion := range e.suggestions {
			result += fmt.Sprintf("\n💡 %s", suggestion)
		}
	}

	return result
}

// Unwrap returns the underlying cause for error chain support.
func (e *TransformError) Unwrap() error {
	return e.cause
}

type FileError struct {
	path        string
	operation   string
	message     string
	cause       error
	suggestions []string
	context     string
}

// NewFileError creates a new FileError with the given parameters.
func NewFileError(path, operation, message string, cause error) *FileError {
	return &FileError{
		path:        path,
		operation:   operation,
		message:     message,
		cause:       cause,
		suggestions: nil,
		context:     "",
	}
}

// WithSuggestions returns a new FileError with the given suggestions (immutable builder).
func (e *FileError) WithSuggestions(suggestions []string) *FileError {
	return &FileError{
		path:        e.path,
		operation:   e.operation,
		message:     e.message,
		cause:       e.cause,
		suggestions: suggestions,
		context:     e.context,
	}
}

// WithContext returns a new FileError with the given context (immutable builder).
func (e *FileError) WithContext(context string) *FileError {
	return &FileError{
		path:        e.path,
		operation:   e.operation,
		message:     e.message,
		cause:       e.cause,
		suggestions: e.suggestions,
		context:     context,
	}
}

// Path returns the file path where the error occurred.
func (e *FileError) Path() string {
	return e.path
}

// Operation returns the operation that failed (read, write, etc.).
func (e *FileError) Operation() string {
	return e.operation
}

// Message returns the file error message.
func (e *FileError) Message() string {
	return e.message
}

// Cause returns the underlying error that caused the file operation failure.
func (e *FileError) Cause() error {
	return e.cause
}

// Suggestions returns a copy of the suggestions slice.
func (e *FileError) Suggestions() []string {
	if e.suggestions == nil {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns additional context information.
func (e *FileError) Context() string {
	return e.context
}

// Error formats the file error as a string.
func (e *FileError) Error() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("file error (%s %s)", e.operation, e.path))

	if e.message != "" {
		parts = append(parts, e.message)
	}

	result := strings.Join(parts, ": ")

	if e.cause != nil {
		result += fmt.Sprintf(": %v", e.cause)
	}

	if len(e.suggestions) > 0 {
		for _, suggestion := range e.suggestions {
			result += fmt.Sprintf("\n💡 %s", suggestion)
		}
	}

	return result
}

// Unwrap returns the underlying cause for error chain support.
func (e *FileError) Unwrap() error {
	return e.cause
}

// IsNotFound returns true if the error indicates the file was not found.
func (e *FileError) IsNotFound() bool {
	msg := strings.ToLower(e.message)
	if strings.Contains(msg, "not found") || strings.Contains(msg, "does not exist") || strings.Contains(msg, "no such file") {
		return true
	}
	if e.cause != nil {
		causeMsg := strings.ToLower(e.cause.Error())
		return strings.Contains(causeMsg, "not found") || strings.Contains(causeMsg, "does not exist") || strings.Contains(causeMsg, "no such file")
	}
	return false
}

type ConfigError struct {
	field       string
	value       string
	message     string
	suggestions []string
	context     string
}

// NewConfigError creates a new ConfigError with the given parameters.
// Note: The constructor signature has changed - 'available' parameter removed.
// Use WithSuggestions() builder to add available options.
func NewConfigError(field, value, message string) *ConfigError {
	return &ConfigError{
		field:       field,
		value:       value,
		message:     message,
		suggestions: nil,
		context:     "",
	}
}

// WithSuggestions returns a new ConfigError with the given suggestions (immutable builder).
func (e *ConfigError) WithSuggestions(suggestions []string) *ConfigError {
	return &ConfigError{
		field:       e.field,
		value:       e.value,
		message:     e.message,
		suggestions: suggestions,
		context:     e.context,
	}
}

// WithContext returns a new ConfigError with the given context (immutable builder).
func (e *ConfigError) WithContext(context string) *ConfigError {
	return &ConfigError{
		field:       e.field,
		value:       e.value,
		message:     e.message,
		suggestions: e.suggestions,
		context:     context,
	}
}

// Field returns the configuration field that caused the error.
func (e *ConfigError) Field() string {
	return e.field
}

// Value returns the invalid value that caused the error.
func (e *ConfigError) Value() string {
	return e.value
}

// Message returns the config error message.
func (e *ConfigError) Message() string {
	return e.message
}

// Suggestions returns a copy of the suggestions slice.
func (e *ConfigError) Suggestions() []string {
	if e.suggestions == nil {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns additional context information.
func (e *ConfigError) Context() string {
	return e.context
}

// Error formats the config error as a string.
func (e *ConfigError) Error() string {
	var parts []string

	if e.field != "" && e.value != "" {
		parts = append(parts, fmt.Sprintf("config error: invalid %s '%s'", e.field, e.value))
	} else {
		parts = append(parts, "config error")
	}

	if e.message != "" {
		parts = append(parts, e.message)
	}

	result := strings.Join(parts, ": ")

	if len(e.suggestions) > 0 {
		result += fmt.Sprintf("\n💡 %s", strings.Join(e.suggestions, "\n💡 "))
	}

	return result
}

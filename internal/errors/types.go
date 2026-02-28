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

type ValidationError struct {
	Message        string
	Field          string
	suggestionList []string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s (field: %s)", e.Message, e.Field)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Suggestions() []string {
	return e.suggestionList
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

func NewValidationError(message, field string, suggestions []string) *ValidationError {
	return &ValidationError{
		Message:        message,
		Field:          field,
		suggestionList: suggestions,
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

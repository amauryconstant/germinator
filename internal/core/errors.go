package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// NotFoundError represents a missing-entity lookup (e.g., library ref,
// preset name). It is the typed error consumers use to detect
// "not found" conditions via errors.As.
type NotFoundError struct {
	Entity string
	Key    string
}

// NewNotFoundError creates a NotFoundError for the named entity and key.
func NewNotFoundError(entity, key string) *NotFoundError {
	return &NotFoundError{Entity: entity, Key: key}
}

// Error formats the not-found error as a canonical "not found: <key>".
func (e *NotFoundError) Error() string {
	return "not found: " + e.Key
}

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
		parts = append(parts, "parse error in "+e.path)
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
		var b strings.Builder
		for _, suggestion := range e.suggestions {
			b.WriteString("\n💡 ")
			b.WriteString(suggestion)
		}
		result += b.String()
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
		var b strings.Builder
		for _, suggestion := range e.suggestions {
			b.WriteString("\n💡 ")
			b.WriteString(suggestion)
		}
		result += b.String()
	}

	return result
}

// Unwrap returns nil (validation errors don't wrap other errors).
// Provided for API consistency with other error types.
func (e *ValidationError) Unwrap() error {
	return nil
}

// TransformError represents a transformation failure with immutable builders for fluent construction.
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
		var b strings.Builder
		for _, suggestion := range e.suggestions {
			b.WriteString("\n💡 ")
			b.WriteString(suggestion)
		}
		result += b.String()
	}

	return result
}

// Unwrap returns the underlying cause for error chain support.
func (e *TransformError) Unwrap() error {
	return e.cause
}

// FileError represents a file operation failure with immutable builders for fluent construction.
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
		var b strings.Builder
		for _, suggestion := range e.suggestions {
			b.WriteString("\n💡 ")
			b.WriteString(suggestion)
		}
		result += b.String()
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

// ConfigError represents a configuration failure with immutable builders for fluent construction.
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
		result += "\n💡 " + strings.Join(e.suggestions, "\n💡 ")
	}

	return result
}

// InitializeError represents a single resource installation failure.
type InitializeError struct {
	ref         string
	inputPath   string
	outputPath  string
	cause       error
	suggestions []string
	context     string
}

// NewInitializeError creates a new InitializeError.
func NewInitializeError(ref, inputPath, outputPath string, cause error) *InitializeError {
	return &InitializeError{
		ref:         ref,
		inputPath:   inputPath,
		outputPath:  outputPath,
		cause:       cause,
		suggestions: nil,
		context:     "",
	}
}

// WithSuggestions returns a new InitializeError with the given suggestions.
func (e *InitializeError) WithSuggestions(suggestions []string) *InitializeError {
	return &InitializeError{
		ref:         e.ref,
		inputPath:   e.inputPath,
		outputPath:  e.outputPath,
		cause:       e.cause,
		suggestions: suggestions,
		context:     e.context,
	}
}

// WithContext returns a new InitializeError with the given context.
func (e *InitializeError) WithContext(context string) *InitializeError {
	return &InitializeError{
		ref:         e.ref,
		inputPath:   e.inputPath,
		outputPath:  e.outputPath,
		cause:       e.cause,
		suggestions: e.suggestions,
		context:     context,
	}
}

// Ref returns the resource reference.
func (e *InitializeError) Ref() string { return e.ref }

// InputPath returns the input path of the failed resource.
func (e *InitializeError) InputPath() string { return e.inputPath }

// OutputPath returns the output path of the failed resource.
func (e *InitializeError) OutputPath() string { return e.outputPath }

// Cause returns the underlying cause.
func (e *InitializeError) Cause() error { return e.cause }

// Suggestions returns a copy of the suggestions slice.
func (e *InitializeError) Suggestions() []string {
	if e.suggestions == nil {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns the additional context.
func (e *InitializeError) Context() string { return e.context }

// Error formats the initialize error.
func (e *InitializeError) Error() string {
	var parts []string
	parts = append(parts, "initialize failed: "+e.ref)
	if e.outputPath != "" {
		parts = append(parts, "output: "+e.outputPath)
	}
	if e.cause != nil {
		parts = append(parts, e.cause.Error())
	}
	result := strings.Join(parts, ": ")
	if e.context != "" {
		result += " (" + e.context + ")"
	}
	if len(e.suggestions) > 0 {
		result += "\n💡 " + strings.Join(e.suggestions, "\n💡 ")
	}
	return result
}

// Unwrap returns the underlying cause.
func (e *InitializeError) Unwrap() error { return e.cause }

// PartialSuccessError represents an aggregation of installation results
// in which some resources succeeded and others failed.
type PartialSuccessError struct {
	succeeded int
	failed    int
	errors    []InitializeError
}

// NewPartialSuccessError creates a new PartialSuccessError.
func NewPartialSuccessError(succeeded, failed int, errs []InitializeError) *PartialSuccessError {
	return &PartialSuccessError{
		succeeded: succeeded,
		failed:    failed,
		errors:    errs,
	}
}

// Succeeded returns the number of resources that succeeded.
func (e *PartialSuccessError) Succeeded() int { return e.succeeded }

// Failed returns the number of resources that failed.
func (e *PartialSuccessError) Failed() int { return e.failed }

// Errors returns a copy of the per-resource failure slice.
func (e *PartialSuccessError) Errors() []InitializeError {
	if e.errors == nil {
		return nil
	}
	result := make([]InitializeError, len(e.errors))
	copy(result, e.errors)
	return result
}

// Error formats the partial-success error.
func (e *PartialSuccessError) Error() string {
	return fmt.Sprintf("partial success: %d succeeded, %d failed", e.succeeded, e.failed)
}

// Unwrap returns nil (PartialSuccessError aggregates other errors but
// does not chain a single cause).
func (e *PartialSuccessError) Unwrap() error { return nil }

// OperationError represents a per-operation failure with an optional
// wrapped cause. It is the typed error used for per-file library
// operations (orphan discovery name_conflict aggregation, etc.) so
// callers can errors.As detect it and output.FormatError can render it
// through the typed-error dispatcher chain.
type OperationError struct {
	Op       string
	Resource string
	Cause    error
}

// NewOperationError creates an OperationError for the given operation,
// resource, and wrapped cause. Cause may be nil.
func NewOperationError(op, resource string, cause error) *OperationError {
	return &OperationError{Op: op, Resource: resource, Cause: cause}
}

// Error renders the operation error as "<op>: <resource>".
// Error strings are lowercase with no trailing punctuation.
func (e *OperationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.Resource)
}

// Unwrap returns the wrapped cause so errors.Is and errors.As traverse
// the chain.
func (e *OperationError) Unwrap() error { return e.Cause }

// UsageError represents a CLI flag validation error that is not caught
// by Cobra's MarkFlagRequired or Args validators, nor by pflag's typed
// errors. It carries the flag name, the reason, and optional suggestions
// following the project's typed-error builder pattern (matching
// ParseError, ValidationError, TransformError, FileError, ConfigError,
// InitializeError). It maps to ExitCodeUsage (2) via cmdutil.ExitCodeFor.
//
// The flag and reason parameters MUST be lowercase with no trailing
// punctuation (Go error-string convention per golang-error-handling
// rule 3 — references/error-creation.md:32). Callers passing upper-case
// flag names violate the convention and are a programmer error.
type UsageError struct {
	flag        string
	reason      string
	suggestions []string
}

// NewUsageError creates a new UsageError with the given flag name and
// reason. Suggestions are initially nil; use WithSuggestions to add
// remediation hints.
func NewUsageError(flag, reason string) *UsageError {
	return &UsageError{flag: flag, reason: reason, suggestions: nil}
}

// Flag returns the flag name that triggered the usage error.
func (e *UsageError) Flag() string { return e.flag }

// Reason returns the human-readable reason for the usage error.
func (e *UsageError) Reason() string { return e.reason }

// Suggestions returns a defensive copy of the suggestions slice via
// slices.Clone (Go 1.21+) so callers cannot mutate the receiver's
// internal slice. Returns nil when no suggestions are set.
func (e *UsageError) Suggestions() []string { return slices.Clone(e.suggestions) }

// WithSuggestions returns a NEW *UsageError with the same flag and
// reason and a freshly-allocated suggestions slice (immutable builder).
// The input slice is defensive-copied via slices.Clone to prevent
// caller mutation.
func (e *UsageError) WithSuggestions(suggestions []string) *UsageError {
	return &UsageError{
		flag:        e.flag,
		reason:      e.reason,
		suggestions: slices.Clone(suggestions),
	}
}

// Error formats the usage error as "<flag>: <reason>".
func (e *UsageError) Error() string { return e.flag + ": " + e.reason }

// Unwrap returns nil (UsageError is a leaf error; it does not wrap an
// underlying cause). The godoc on the type explicitly notes this so
// future maintainers do not add a cause field and break the contract.
func (e *UsageError) Unwrap() error { return nil }

// CobraUsageError is a sentinel that wraps a Cobra arg-validation
// error (emitted by cobra.ExactArgs, MinimumNArgs, MaximumNArgs,
// RangeArgs, and MarkFlagRequired-derived "required flag(s) ..." strings)
// so cmdutil.ExitCodeFor can match the typed error and return
// ExitCodeUsage (2). The Must* prefix telegraphs the panic on nil
// cause (a violated invariant), mirroring regexp.MustCompile and
// template.Must.
type CobraUsageError struct {
	err error
}

// MustNewCobraUsageError creates a new CobraUsageError wrapping err.
// It panics if err is nil — a nil cause is a programmer error, not a
// recoverable state. The Must* prefix telegraphs the panic to callers,
// matching regexp.MustCompile and template.Must. No nil-guard fallback
// is provided; callers requiring nil-safety must use a try/recv pattern
// or a separate New* constructor (not provided in this change).
func MustNewCobraUsageError(err error) *CobraUsageError {
	if err == nil {
		panic("MustNewCobraUsageError: cause is required (programmer error)")
	}
	return &CobraUsageError{err: err}
}

// Error returns the wrapped error's Error() string verbatim.
func (e *CobraUsageError) Error() string { return e.err.Error() }

// Unwrap returns the wrapped cause so errors.Is and errors.As traverse
// the chain.
func (e *CobraUsageError) Unwrap() error { return e.err }

// MarshalJSON implementations.
//
// Each MarshalJSON returns the JSON bytes {"error": "<Error()>"}. The
// shape is the only sensible JSON projection of a Go error interface
// (stdlib's default would marshal as {}). Any future exported struct
// fields on these typed errors must be exposed via MarshalJSON to
// appear in JSON output — json.Marshaler precedence in stdlib means
// MarshalJSON wins over struct-field marshaling.

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *ParseError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal ParseError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *ValidationError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal ValidationError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *TransformError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal TransformError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *FileError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal FileError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *ConfigError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal ConfigError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *NotFoundError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal NotFoundError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *OperationError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal OperationError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *InitializeError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal InitializeError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *PartialSuccessError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal PartialSuccessError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *UsageError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal UsageError: %w", err)
	}
	return b, nil
}

// MarshalJSON renders the typed error as {"error": "<Error()>"}.
// See the package-level MarshalJSON block for the rationale.
func (e *CobraUsageError) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: e.Error()})
	if err != nil {
		return nil, fmt.Errorf("marshal CobraUsageError: %w", err)
	}
	return b, nil
}

// Compile-time interface checks. Catches method-signature typos at
// build time (per golang-design-patterns rule 19).

var (
	_ error = (*ParseError)(nil)
	_ error = (*ValidationError)(nil)
	_ error = (*TransformError)(nil)
	_ error = (*FileError)(nil)
	_ error = (*ConfigError)(nil)
	_ error = (*NotFoundError)(nil)
	_ error = (*OperationError)(nil)
	_ error = (*InitializeError)(nil)
	_ error = (*PartialSuccessError)(nil)
	_ error = (*UsageError)(nil)
	_ error = (*CobraUsageError)(nil)
)

var (
	_ json.Marshaler = (*ParseError)(nil)
	_ json.Marshaler = (*ValidationError)(nil)
	_ json.Marshaler = (*TransformError)(nil)
	_ json.Marshaler = (*FileError)(nil)
	_ json.Marshaler = (*ConfigError)(nil)
	_ json.Marshaler = (*NotFoundError)(nil)
	_ json.Marshaler = (*OperationError)(nil)
	_ json.Marshaler = (*InitializeError)(nil)
	_ json.Marshaler = (*PartialSuccessError)(nil)
	_ json.Marshaler = (*UsageError)(nil)
	_ json.Marshaler = (*CobraUsageError)(nil)
)

package config

import (
	"fmt"
)

// WriteError represents a configuration file I/O failure. It is
// emitted by helpers that write or scaffold config files (e.g.,
// WriteDefault) when an underlying *os.PathError occurs. The type
// follows the project's typed-error pattern: private fields,
// exported accessors, and Unwrap() so errors.Is / errors.As traverse
// the chain.
//
// Per the cli-exit-codes spec, *WriteError maps to ExitCodeError (1) —
// an I/O failure during config scaffolding is an operational error,
// not a user-input validation error.
//
// The optional `message` field carries user-facing context (e.g.,
// "config file already exists (use --force to overwrite)") that is
// rendered between the op/path segment and the cause. Construct via
// NewWriteErrorWithMessage when context text is needed; NewWriteError
// is sufficient when the cause alone is informative.
type WriteError struct {
	op      string
	path    string
	message string
	cause   error
}

// NewWriteError creates a WriteError for the given operation, path,
// and underlying cause. Cause may be nil (rare; the helper layer
// surfaces the underlying *os.PathError when available).
func NewWriteError(op, path string, cause error) *WriteError {
	return &WriteError{op: op, path: path, cause: cause}
}

// NewWriteErrorWithMessage creates a WriteError with a user-facing
// message rendered between the op/path segment and the cause. Use
// this when the failure is not adequately described by the cause
// alone (e.g., a precondition violation like "file already exists"
// where the cause is nil). The message follows Go error-string
// conventions (lowercase, no trailing punctuation) per
// references/error-creation.md in the golang-error-handling skill.
func NewWriteErrorWithMessage(op, path, message string, cause error) *WriteError {
	return &WriteError{op: op, path: path, message: message, cause: cause}
}

// Op returns the I/O operation that failed (e.g., "mkdir", "write",
// "stat").
func (e *WriteError) Op() string { return e.op }

// Path returns the file path involved in the failed operation.
func (e *WriteError) Path() string { return e.path }

// Message returns the user-facing context attached to the failure,
// or the empty string when no message was set.
func (e *WriteError) Message() string { return e.message }

// Cause returns the underlying error that caused the write failure,
// or nil when none was attached.
func (e *WriteError) Cause() error { return e.cause }

// Error formats the write error as "<op> <path>" plus an optional
// "<message>" segment and an optional "<cause>" segment, each
// colon-joined when present. Examples:
//
//	NewWriteError("write", "/etc/cfg", nil) -> "write /etc/cfg"
//	NewWriteError("write", "/etc/cfg", err) -> "write /etc/cfg: <err>"
//	NewWriteErrorWithMessage("create", "/etc/cfg", "file already exists", nil)
//	  -> "create /etc/cfg: file already exists"
//	NewWriteErrorWithMessage("write", "/etc/cfg", "write failed", err)
//	  -> "write /etc/cfg: write failed: <err>"
func (e *WriteError) Error() string {
	body := fmt.Sprintf("%s %s", e.op, e.path)
	if e.message != "" {
		body += ": " + e.message
	}
	if e.cause != nil {
		body += ": " + e.cause.Error()
	}
	return body
}

// Unwrap returns the underlying cause so errors.Is and errors.As
// traverse the chain (e.g., errors.Is(err, os.ErrNotExist)).
func (e *WriteError) Unwrap() error { return e.cause }

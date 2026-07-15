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
type WriteError struct {
	op    string
	path  string
	cause error
}

// NewWriteError creates a WriteError for the given operation, path,
// and underlying cause. Cause may be nil (rare; the helper layer
// surfaces the underlying *os.PathError when available).
func NewWriteError(op, path string, cause error) *WriteError {
	return &WriteError{op: op, path: path, cause: cause}
}

// Op returns the I/O operation that failed (e.g., "mkdir", "write",
// "stat").
func (e *WriteError) Op() string { return e.op }

// Path returns the file path involved in the failed operation.
func (e *WriteError) Path() string { return e.path }

// Cause returns the underlying error that caused the write failure,
// or nil when none was attached.
func (e *WriteError) Cause() error { return e.cause }

// Error formats the write error as "<op> <path>: <message>".
// When cause is non-nil, the cause's message is appended after a
// colon. When cause is nil, only the "<op> <path>" segment is
// returned.
func (e *WriteError) Error() string {
	body := fmt.Sprintf("%s %s", e.op, e.path)
	if e.cause != nil {
		body += ": " + e.cause.Error()
	}
	return body
}

// Unwrap returns the underlying cause so errors.Is and errors.As
// traverse the chain (e.g., errors.Is(err, os.ErrNotExist)).
func (e *WriteError) Unwrap() error { return e.cause }

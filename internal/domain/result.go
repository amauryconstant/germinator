// Package validation provides a functional validation pipeline with Result[T] type
// for composable, early-exit validation with clean error handling.
package domain

// Result[T] represents either a success value or an error, enabling functional
// error handling without exceptions or multiple return values.
type Result[T any] struct {
	Value T
	Error error
}

// NewResult creates a successful Result containing the given value.
func NewResult[T any](value T) Result[T] {
	return Result[T]{
		Value: value,
		Error: nil,
	}
}

// NewErrorResult creates an error Result with the zero value of T and the given error.
func NewErrorResult[T any](err error) Result[T] {
	var zero T
	return Result[T]{
		Value: zero,
		Error: err,
	}
}

// IsSuccess returns true if the Result contains a success value (Error is nil).
func (r Result[T]) IsSuccess() bool {
	return r.Error == nil
}

// IsError returns true if the Result contains an error (Error is not nil).
func (r Result[T]) IsError() bool {
	return r.Error != nil
}

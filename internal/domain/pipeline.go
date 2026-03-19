package domain

import (
	"errors"
)

// ValidationFunc validates input of type T and returns a Result[bool].
type ValidationFunc[T any] func(T) Result[bool]

// ValidationPipeline chains multiple ValidationFunc functions and collects all errors.
type ValidationPipeline[T any] struct {
	validations []ValidationFunc[T]
}

// NewValidationPipeline creates a new ValidationPipeline with the given validation functions.
// Validators are executed in the order provided.
func NewValidationPipeline[T any](validations ...ValidationFunc[T]) *ValidationPipeline[T] {
	return &ValidationPipeline[T]{
		validations: validations,
	}
}

// Validate runs all validators in order, collecting all errors.
// Returns NewResult(true) if all validators pass, or an error Result with a combined error if any fail.
func (p *ValidationPipeline[T]) Validate(input T) Result[bool] {
	var allErrors []error

	for _, validation := range p.validations {
		result := validation(input)
		if result.IsError() {
			allErrors = append(allErrors, result.Error)
		}
	}

	if len(allErrors) > 0 {
		// Combine all errors into a single error
		combinedErr := errors.Join(allErrors...)
		return NewErrorResult[bool](combinedErr)
	}

	return NewResult(true)
}

package validation

// ValidationFunc[T] is a function that validates input of type T and returns a Result[bool].
type ValidationFunc[T any] func(T) Result[bool]

// ValidationPipeline[T] chains multiple ValidationFunc[T] functions with early exit on first error.
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

// Validate runs all validators in order, exiting early on the first error.
// Returns NewResult(true) if all validators pass, or the first error Result encountered.
func (p *ValidationPipeline[T]) Validate(input T) Result[bool] {
	for _, validation := range p.validations {
		result := validation(input)
		if result.IsError() {
			return result
		}
	}
	return NewResult(true)
}

## Purpose

Provide a composable ValidationPipeline[T] that chains multiple ValidationFunc[T] functions with early exit on first error.

## ADDED Requirements

### Requirement: ValidationFunc type definition

The system SHALL provide a `ValidationFunc[T any] func(T) Result[bool]` type in `internal/validation/pipeline.go`.

#### Scenario: ValidationFunc accepts input and returns Result

- **WHEN** a ValidationFunc[T] is called with input of type T
- **THEN** it SHALL return `Result[bool]`

---

### Requirement: ValidationPipeline type definition

The system SHALL provide a `ValidationPipeline[T any]` struct in `internal/validation/pipeline.go`.

#### Scenario: ValidationPipeline stores validation functions

- **WHEN** a ValidationPipeline[T] is created
- **THEN** it SHALL store a slice of `ValidationFunc[T]`

---

### Requirement: NewValidationPipeline constructor

The system SHALL provide a `NewValidationPipeline[T any](validations ...ValidationFunc[T]) *ValidationPipeline[T]` function.

#### Scenario: NewValidationPipeline with no validators

- **WHEN** `NewValidationPipeline[string]()` is called with no arguments
- **THEN** it SHALL return a ValidationPipeline with empty validations slice

#### Scenario: NewValidationPipeline with multiple validators

- **WHEN** `NewValidationPipeline(validator1, validator2, validator3)` is called
- **THEN** it SHALL return a ValidationPipeline containing all three validators in order

---

### Requirement: ValidationPipeline.Validate method

The system SHALL provide a `Validate(input T) Result[bool]` method on ValidationPipeline[T].

#### Scenario: Validate runs all validators on valid input

- **WHEN** `pipeline.Validate(input)` is called with input that passes all validators
- **THEN** it SHALL return `NewResult(true)`
- **AND** all validators SHALL be called in order

#### Scenario: Validate exits early on first error

- **WHEN** `pipeline.Validate(input)` is called with input that fails the second validator
- **THEN** it SHALL return the error Result from the second validator
- **AND** the third validator SHALL NOT be called

#### Scenario: Validate with empty pipeline

- **WHEN** `pipeline.Validate(input)` is called on an empty pipeline
- **THEN** it SHALL return `NewResult(true)`

#### Scenario: Validate stops on first failure

- **WHEN** a validator returns an error Result
- **THEN** no subsequent validators SHALL be called
- **AND** the error Result SHALL be returned immediately

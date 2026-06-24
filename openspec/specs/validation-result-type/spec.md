## Purpose

Provide a generic Result[T] type for functional error handling, enabling functions to return either a success value or an error without using exceptions or multiple return values.

## Requirements

### Requirement: Result[T] type definition

The system SHALL provide a generic `Result[T any]` type in `internal/validation/result.go`.

#### Scenario: Result struct has Value and Error fields

- **WHEN** a Result[T] is examined
- **THEN** it SHALL have a `Value T` field
- **AND** it SHALL have an `Error error` field

---

### Requirement: NewResult constructor

The system SHALL provide a `NewResult[T any](value T) Result[T]` function.

#### Scenario: NewResult creates success result

- **WHEN** `NewResult(42)` is called
- **THEN** the returned Result SHALL have `Value` equal to `42`
- **AND** the returned Result SHALL have `Error` equal to `nil`

#### Scenario: NewResult works with any type

- **WHEN** `NewResult(true)` is called
- **THEN** the returned Result SHALL have `Value` equal to `true`
- **AND** the returned Result SHALL have `Error` equal to `nil`

---

### Requirement: NewErrorResult constructor

The system SHALL provide a `NewErrorResult[T any](err error) Result[T]` function.

#### Scenario: NewErrorResult creates error result

- **WHEN** `NewErrorResult[int](someError)` is called
- **THEN** the returned Result SHALL have `Value` equal to zero value of T
- **AND** the returned Result SHALL have `Error` equal to `someError`

#### Scenario: NewErrorResult zero value for bool

- **WHEN** `NewErrorResult[bool](someError)` is called
- **THEN** the returned Result SHALL have `Value` equal to `false` (zero value)

---

### Requirement: IsSuccess method

The system SHALL provide an `IsSuccess() bool` method on Result[T].

#### Scenario: IsSuccess returns true when Error is nil

- **WHEN** `result.IsSuccess()` is called on a Result with `Error == nil`
- **THEN** it SHALL return `true`

#### Scenario: IsSuccess returns false when Error is not nil

- **WHEN** `result.IsSuccess()` is called on a Result with `Error != nil`
- **THEN** it SHALL return `false`

---

### Requirement: IsError method

The system SHALL provide an `IsError() bool` method on Result[T].

#### Scenario: IsError returns true when Error is not nil

- **WHEN** `result.IsError()` is called on a Result with `Error != nil`
- **THEN** it SHALL return `true`

#### Scenario: IsError returns false when Error is nil

- **WHEN** `result.IsError()` is called on a Result with `Error == nil`
- **THEN** it SHALL return `false`

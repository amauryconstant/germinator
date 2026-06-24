# mock-infrastructure Specification

## Purpose

Provide testify/mock implementations for all application service interfaces to enable isolated unit testing without real implementations.

## Requirements

### Requirement: Mock Package Structure

The test infrastructure SHALL provide a `test/mocks/` directory containing mock implementations.

#### Scenario: Mocks directory exists

- **WHEN** the project structure is inspected
- **THEN** a `test/mocks/` directory SHALL exist
- **AND** it SHALL contain mock files for each application interface

#### Scenario: Mock files follow naming convention

- **WHEN** mock files are created
- **THEN** each file SHALL be named `<interface>_mock.go`
- **AND** the mock struct SHALL be named `Mock<Interface>`

---

### Requirement: MockTransformer

The test infrastructure SHALL provide a mock implementation of the Transformer interface.

#### Scenario: MockTransformer implements Transformer interface

- **WHEN** `MockTransformer` is instantiated
- **THEN** it SHALL implement `application.Transformer`
- **AND** it SHALL embed `testify/mock.Mock`

#### Scenario: MockTransformer Transform method can be configured

- **GIVEN** a `MockTransformer` instance
- **WHEN** `On("Transform", ctx, request).Return(result, error)` is called
- **THEN** subsequent calls to `Transform(ctx, request)` SHALL return the configured result

#### Scenario: MockTransformer Transform method records calls

- **GIVEN** a `MockTransformer` that has been called
- **WHEN** `AssertCalled(t, "Transform", ctx, request)` is invoked
- **THEN** the assertion SHALL pass if Transform was called with matching arguments

---

### Requirement: MockValidator

The test infrastructure SHALL provide a mock implementation of the Validator interface.

#### Scenario: MockValidator implements Validator interface

- **WHEN** `MockValidator` is instantiated
- **THEN** it SHALL implement `application.Validator`
- **AND** it SHALL embed `testify/mock.Mock`

#### Scenario: MockValidator Validate method can be configured

- **GIVEN** a `MockValidator` instance
- **WHEN** `On("Validate", ctx, request).Return(result, error)` is called
- **THEN** subsequent calls to `Validate(ctx, request)` SHALL return the configured result

#### Scenario: MockValidator returns validation result with errors

- **GIVEN** a `MockValidator` configured with `Return(&ValidateResult{Errors: []error{err}}, nil)`
- **WHEN** `Validate` is called
- **THEN** the result SHALL contain the configured errors
- **AND** `result.Valid()` SHALL return false

---

### Requirement: MockCanonicalizer

The test infrastructure SHALL provide a mock implementation of the Canonicalizer interface.

#### Scenario: MockCanonicalizer implements Canonicalizer interface

- **WHEN** `MockCanonicalizer` is instantiated
- **THEN** it SHALL implement `application.Canonicalizer`
- **AND** it SHALL embed `testify/mock.Mock`

#### Scenario: MockCanonicalizer Canonicalize method can be configured

- **GIVEN** a `MockCanonicalizer` instance
- **WHEN** `On("Canonicalize", ctx, request).Return(result, error)` is called
- **THEN** subsequent calls to `Canonicalize(ctx, request)` SHALL return the configured result

---

### Requirement: MockInitializer

The test infrastructure SHALL provide a mock implementation of the Initializer interface.

#### Scenario: MockInitializer implements Initializer interface

- **WHEN** `MockInitializer` is instantiated
- **THEN** it SHALL implement `application.Initializer`
- **AND** it SHALL embed `testify/mock.Mock`

#### Scenario: MockInitializer Initialize method can be configured

- **GIVEN** a `MockInitializer` instance
- **WHEN** `On("Initialize", ctx, request).Return(results, error)` is called
- **THEN** subsequent calls to `Initialize(ctx, request)` SHALL return the configured results

#### Scenario: MockInitializer returns multiple results

- **GIVEN** a `MockInitializer` configured with multiple `InitializeResult` items
- **WHEN** `Initialize` is called
- **THEN** the result slice SHALL contain all configured items

---

### Requirement: Mock Usage in Unit Tests

Unit tests SHALL be able to use mocks for interface isolation.

#### Scenario: Command test uses mock validator

- **GIVEN** a unit test for the validate command
- **WHEN** the test creates a `MockValidator` and configures its behavior
- **THEN** the command SHALL use the mock instead of real implementation
- **AND** the test SHALL verify expected mock calls

#### Scenario: Mock assertions verify call count

- **GIVEN** a mock that was called during test execution
- **WHEN** `AssertNumberOfCalls(t, "Method", n)` is invoked
- **THEN** the assertion SHALL pass if the method was called exactly n times

#### Scenario: Mock assertions verify call arguments

- **GIVEN** a mock that was called with specific arguments
- **WHEN** `AssertCalled(t, "Method", arg1, arg2)` is invoked
- **THEN** the assertion SHALL pass if arguments match

---

### Requirement: Mocks Coexist with Real Implementations

The mock infrastructure SHALL coexist with existing golden file and integration tests.

#### Scenario: Unit tests use mocks, integration tests use real implementations

- **GIVEN** both mock and real implementations exist
- **WHEN** unit tests are run with `go test -short`
- **THEN** mocks MAY be used for isolation
- **WHEN** integration tests are run
- **THEN** real implementations SHALL be used

#### Scenario: Golden file tests unchanged

- **GIVEN** existing golden file tests in `internal/services/*_golden_test.go`
- **WHEN** mock infrastructure is added
- **THEN** golden file tests SHALL continue to use real implementations
- **AND** golden file tests SHALL continue to pass unchanged

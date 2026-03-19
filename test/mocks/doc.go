// Package mocks provides testify/mock implementations of application service interfaces.
//
// This package enables isolated unit testing by allowing tests to mock service implementations
// without relying on real implementations or external dependencies.
//
// Mock Usage Pattern:
//
// 1. Create a mock instance:
//
//	mockValidator := new(mocks.MockValidator)
//
// 2. Set up expected method calls:
//
//	mockValidator.On("Validate", ctx, mock.AnythingOfType("*application.ValidateRequest")).
//	    Return(&domain.ValidateResult{}, nil)
//
// 3. Call the method in your test code
//
// 4. Assert that the method was called:
//
//	mockValidator.AssertCalled(t, "Validate", ctx, mock.AnythingOfType("*application.ValidateRequest"))
//	mockValidator.AssertNumberOfCalls(t, "Validate", 1)
//
// 5. Assert expectations after the test:
//
//	mockValidator.AssertExpectations(t)
//
// Available Mocks:
//
//   - MockTransformer - implements application.Transformer
//   - MockValidator - implements application.Validator
//   - MockCanonicalizer - implements application.Canonicalizer
//   - MockInitializer - implements application.Initializer
//
// See test/AGENTS.md for comprehensive mock usage patterns and examples.
package mocks

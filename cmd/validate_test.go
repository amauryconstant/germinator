package cmd_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/test/mocks"
)

// TestMockValidatorUsage demonstrates how to use MockValidator for isolated unit testing.
//
// This example shows the complete mock lifecycle:
// 1. Create a mock instance
// 2. Set up expected method calls with On()
// 3. Call the method under test
// 4. Verify behavior with AssertCalled() and AssertExpectations()
func TestMockValidatorUsage(t *testing.T) {
	// Setup: Create a mock validator
	mockValidator := new(mocks.MockValidator)
	ctx := context.Background()

	// Test case: Successful validation with no errors
	t.Run("successful validation", func(t *testing.T) {
		// Arrange: Set up expected method call
		expectedReq := &application.ValidateRequest{
			InputPath: "/path/to/document.md",
			Platform:  "opencode",
		}
		expectedResult := &domain.ValidateResult{
			Errors: []error{},
		}

		mockValidator.On("Validate", ctx, expectedReq).
			Return(expectedResult, nil)

		// Act: Call the method being tested
		result, err := mockValidator.Validate(ctx, expectedReq)

		// Assert: Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid(), "expected valid result")
		assert.Empty(t, result.Errors, "expected no validation errors")

		// Verify: Ensure the method was called as expected
		mockValidator.AssertCalled(t, "Validate", ctx, expectedReq)
		mockValidator.AssertNumberOfCalls(t, "Validate", 1)
		mockValidator.AssertExpectations(t)
	})

	// Test case: Validation with errors
	t.Run("validation with errors", func(t *testing.T) {
		// Arrange: Set up expected method call with validation errors
		mockValidator = new(mocks.MockValidator) // Reset mock for this test case
		req := &application.ValidateRequest{
			InputPath: "/path/to/invalid.md",
			Platform:  "claude-code",
		}
		expectedErrors := []error{
			errors.New("missing required field: name"),
			errors.New("invalid permission mode"),
		}
		expectedResult := &domain.ValidateResult{
			Errors: expectedErrors,
		}

		mockValidator.On("Validate", ctx, req).
			Return(expectedResult, nil)

		// Act: Call the method being tested
		result, err := mockValidator.Validate(ctx, req)

		// Assert: Verify results
		assert.NoError(t, err, "method call should succeed even with validation errors")
		assert.NotNil(t, result)
		assert.False(t, result.Valid(), "expected invalid result")
		assert.Len(t, result.Errors, 2, "expected 2 validation errors")
		assert.Contains(t, result.Errors[0].Error(), "missing required field")
		assert.Contains(t, result.Errors[1].Error(), "invalid permission mode")

		// Verify: Ensure the method was called as expected
		mockValidator.AssertCalled(t, "Validate", ctx, req)
		mockValidator.AssertExpectations(t)
	})

	// Test case: Fatal error during validation
	t.Run("fatal error during validation", func(t *testing.T) {
		// Arrange: Set up expected method call with fatal error
		mockValidator = new(mocks.MockValidator) // Reset mock for this test case
		req := &application.ValidateRequest{
			InputPath: "/path/to/missing.md",
			Platform:  "opencode",
		}
		fatalErr := errors.New("file not found")

		mockValidator.On("Validate", ctx, req).
			Return(nil, fatalErr)

		// Act: Call the method being tested
		result, err := mockValidator.Validate(ctx, req)

		// Assert: Verify results
		assert.Error(t, err, "expected fatal error")
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "file not found")

		// Verify: Ensure the method was called as expected
		mockValidator.AssertCalled(t, "Validate", ctx, req)
		mockValidator.AssertExpectations(t)
	})

	// Test case: Using mock.Anything for flexible matching
	t.Run("mock with argument matching", func(t *testing.T) {
		// Arrange: Set up expected call with flexible argument matching
		mockValidator = new(mocks.MockValidator) // Reset mock for this test case

		mockValidator.On("Validate", ctx, mock.AnythingOfType("*application.ValidateRequest")).
			Return(&domain.ValidateResult{Errors: []error{}}, nil)

		// Act: Call with any request
		result, err := mockValidator.Validate(ctx, &application.ValidateRequest{
			InputPath: "/any/path.md",
			Platform:  "opencode",
		})

		// Assert: Verify results
		assert.NoError(t, err)
		assert.True(t, result.Valid())

		// Verify: Ensure the method was called
		mockValidator.AssertCalled(t, "Validate", ctx, mock.AnythingOfType("*application.ValidateRequest"))
		mockValidator.AssertExpectations(t)
	})
}

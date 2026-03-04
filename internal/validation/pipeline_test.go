package validation

import (
	"errors"
	"testing"
)

func TestValidationFunc(t *testing.T) {
	t.Run("accepts input and returns Result", func(t *testing.T) {
		validator := func(s string) Result[bool] {
			if s == "" {
				return NewErrorResult[bool](errors.New("empty string"))
			}
			return NewResult(true)
		}

		result := validator("test")
		if !result.IsSuccess() {
			t.Error("expected validator to return success for non-empty string")
		}

		result = validator("")
		if !result.IsError() {
			t.Error("expected validator to return error for empty string")
		}
	})
}

func TestNewValidationPipeline(t *testing.T) {
	t.Run("creates pipeline with no validators", func(t *testing.T) {
		pipeline := NewValidationPipeline[string]()
		if len(pipeline.validations) != 0 {
			t.Errorf("expected empty validations slice, got %d validators", len(pipeline.validations))
		}
	})

	t.Run("creates pipeline with multiple validators", func(t *testing.T) {
		v1 := func(s string) Result[bool] { return NewResult(true) }
		v2 := func(s string) Result[bool] { return NewResult(true) }
		v3 := func(s string) Result[bool] { return NewResult(true) }

		pipeline := NewValidationPipeline(v1, v2, v3)
		if len(pipeline.validations) != 3 {
			t.Errorf("expected 3 validators, got %d", len(pipeline.validations))
		}
	})
}

func TestValidationPipeline_Validate(t *testing.T) {
	t.Run("runs all validators on valid input", func(t *testing.T) {
		callCount := 0
		v1 := func(s string) Result[bool] {
			callCount++
			return NewResult(true)
		}
		v2 := func(s string) Result[bool] {
			callCount++
			return NewResult(true)
		}
		v3 := func(s string) Result[bool] {
			callCount++
			return NewResult(true)
		}

		pipeline := NewValidationPipeline(v1, v2, v3)
		result := pipeline.Validate("test")

		if !result.IsSuccess() {
			t.Error("expected pipeline to return success for valid input")
		}
		if callCount != 3 {
			t.Errorf("expected all 3 validators to be called, got %d calls", callCount)
		}
	})

	t.Run("exits early on first error", func(t *testing.T) {
		callCount := 0
		v1 := func(s string) Result[bool] {
			callCount++
			return NewResult(true)
		}
		v2 := func(s string) Result[bool] {
			callCount++
			return NewErrorResult[bool](errors.New("validation failed"))
		}
		v3 := func(s string) Result[bool] {
			callCount++
			return NewResult(true)
		}

		pipeline := NewValidationPipeline(v1, v2, v3)
		result := pipeline.Validate("test")

		if !result.IsError() {
			t.Error("expected pipeline to return error when validator fails")
		}
		if callCount != 2 {
			t.Errorf("expected only 2 validators to be called (early exit), got %d calls", callCount)
		}
	})

	t.Run("returns success for empty pipeline", func(t *testing.T) {
		pipeline := NewValidationPipeline[string]()
		result := pipeline.Validate("test")

		if !result.IsSuccess() {
			t.Error("expected empty pipeline to return success")
		}
		if result.Value != true {
			t.Error("expected success result to have Value = true")
		}
	})

	t.Run("stops on first failure and returns error", func(t *testing.T) {
		testErr := errors.New("validation failed")
		v1 := func(s string) Result[bool] {
			return NewErrorResult[bool](testErr)
		}
		v2 := func(s string) Result[bool] {
			t.Error("second validator should not be called after first failure")
			return NewResult(true)
		}

		pipeline := NewValidationPipeline(v1, v2)
		result := pipeline.Validate("test")

		if !result.IsError() {
			t.Error("expected error result")
		}
		if result.Error != testErr {
			t.Errorf("expected error %v, got %v", testErr, result.Error)
		}
	})
}

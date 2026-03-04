package validation

import (
	"errors"
	"testing"
)

func TestNewResult(t *testing.T) {
	t.Run("creates success result with int value", func(t *testing.T) {
		result := NewResult(42)
		if result.Value != 42 {
			t.Errorf("expected Value to be 42, got %v", result.Value)
		}
		if result.Error != nil {
			t.Errorf("expected Error to be nil, got %v", result.Error)
		}
	})

	t.Run("creates success result with bool value", func(t *testing.T) {
		result := NewResult(true)
		if result.Value != true {
			t.Errorf("expected Value to be true, got %v", result.Value)
		}
		if result.Error != nil {
			t.Errorf("expected Error to be nil, got %v", result.Error)
		}
	})

	t.Run("creates success result with string value", func(t *testing.T) {
		result := NewResult("test")
		if result.Value != "test" {
			t.Errorf("expected Value to be 'test', got %v", result.Value)
		}
		if result.Error != nil {
			t.Errorf("expected Error to be nil, got %v", result.Error)
		}
	})
}

func TestNewErrorResult(t *testing.T) {
	testErr := errors.New("test error")

	t.Run("creates error result with int zero value", func(t *testing.T) {
		result := NewErrorResult[int](testErr)
		if result.Value != 0 {
			t.Errorf("expected Value to be 0 (zero value), got %v", result.Value)
		}
		if result.Error != testErr {
			t.Errorf("expected Error to be testErr, got %v", result.Error)
		}
	})

	t.Run("creates error result with bool zero value", func(t *testing.T) {
		result := NewErrorResult[bool](testErr)
		if result.Value != false {
			t.Errorf("expected Value to be false (zero value), got %v", result.Value)
		}
		if result.Error != testErr {
			t.Errorf("expected Error to be testErr, got %v", result.Error)
		}
	})

	t.Run("creates error result with string zero value", func(t *testing.T) {
		result := NewErrorResult[string](testErr)
		if result.Value != "" {
			t.Errorf("expected Value to be empty string (zero value), got %v", result.Value)
		}
		if result.Error != testErr {
			t.Errorf("expected Error to be testErr, got %v", result.Error)
		}
	})
}

func TestIsSuccess(t *testing.T) {
	t.Run("returns true when Error is nil", func(t *testing.T) {
		result := NewResult(42)
		if !result.IsSuccess() {
			t.Error("expected IsSuccess() to return true for success result")
		}
	})

	t.Run("returns false when Error is not nil", func(t *testing.T) {
		result := NewErrorResult[int](errors.New("test error"))
		if result.IsSuccess() {
			t.Error("expected IsSuccess() to return false for error result")
		}
	})
}

func TestIsError(t *testing.T) {
	t.Run("returns true when Error is not nil", func(t *testing.T) {
		result := NewErrorResult[int](errors.New("test error"))
		if !result.IsError() {
			t.Error("expected IsError() to return true for error result")
		}
	})

	t.Run("returns false when Error is nil", func(t *testing.T) {
		result := NewResult(42)
		if result.IsError() {
			t.Error("expected IsError() to return false for success result")
		}
	})
}

package core

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeResult_Shape(t *testing.T) {
	t.Run("has exactly four expected fields", func(t *testing.T) {
		rt := reflect.TypeOf(InitializeResult{})
		if rt.NumField() != 4 {
			t.Fatalf("InitializeResult field count drift: got %d, want 4 (Ref, InputPath, OutputPath, Error)",
				rt.NumField())
		}
		expected := []string{"Ref", "InputPath", "OutputPath", "Error"}
		for i, name := range expected {
			got := rt.Field(i).Name
			if got != name {
				t.Errorf("InitializeResult field %d: got %q, want %q", i, got, name)
			}
		}
	})

	t.Run("nil Error indicates success", func(t *testing.T) {
		r := InitializeResult{
			Ref:        "skill/commit",
			InputPath:  "/lib/skills/commit.md",
			OutputPath: ".opencode/skills/commit/SKILL.md",
		}
		if r.Error != nil {
			t.Errorf("expected nil Error to indicate success, got %v", r.Error)
		}
	})

	t.Run("non-nil Error indicates failure", func(t *testing.T) {
		r := InitializeResult{
			Ref:        "skill/commit",
			InputPath:  "/lib/skills/commit.md",
			OutputPath: ".opencode/skills/commit/SKILL.md",
			Error:      errors.New("boom"),
		}
		if r.Error == nil {
			t.Error("expected non-nil Error to indicate failure, got nil")
		}
	})
}

// TestValidateResult_Valid covers (*ValidateResult).Valid() at
// core/results.go:17. The zero value (no errors set) must report
// Valid; a result with one or more errors must report !Valid.
func TestValidateResult_Valid(t *testing.T) {
	t.Run("zero-value is valid", func(t *testing.T) {
		r := &ValidateResult{}
		assert.True(t, r.Valid(),
			"zero-value ValidateResult must be Valid (no errors)")
	})

	t.Run("explicit empty slice is valid", func(t *testing.T) {
		r := &ValidateResult{Errors: []error{}}
		assert.True(t, r.Valid(),
			"ValidateResult with empty Errors slice must be Valid")
	})

	t.Run("nil Errors slice is valid", func(t *testing.T) {
		r := &ValidateResult{Errors: nil}
		assert.True(t, r.Valid(),
			"ValidateResult with nil Errors must be Valid")
	})

	t.Run("single error is invalid", func(t *testing.T) {
		r := &ValidateResult{Errors: []error{errors.New("boom")}}
		assert.False(t, r.Valid(),
			"ValidateResult with one error must NOT be Valid")
	})

	t.Run("multiple errors are invalid", func(t *testing.T) {
		r := &ValidateResult{
			Errors: []error{errors.New("a"), errors.New("b"), errors.New("c")},
		}
		assert.False(t, r.Valid(),
			"ValidateResult with multiple errors must NOT be Valid")
	})
}

// TestUnwrapChains exercises the Unwrap() method on every typed error
// in the core package that implements it. Two flavors are tested:
//
//   - Chain errors (Unwrap returns non-nil cause): ParseError,
//     TransformError, FileError, InitializeError, OperationError,
//     CobraUsageError. For these, errors.Is on the chain must reach
//     the wrapped sentinel; errors.As on the chain must reach the
//     typed error.
//   - Leaf errors (Unwrap returns nil): ValidationError, UsageError,
//     PartialSuccessError. For these, Unwrap must return nil so the
//     chain terminates at the typed error.
//
// The proposal listed `ValidationError` in the chain group; the actual
// implementation has ValidationError.Unwrap returning nil (it's a leaf),
// so this test treats ValidationError as a leaf per the implementation.
func TestUnwrapChains(t *testing.T) {
	sentinel := errors.New("root cause")

	t.Run("ParseError unwraps to cause", func(t *testing.T) {
		err := NewParseError("path.md", "boom", sentinel)
		require.NotNil(t, err.Unwrap())
		assert.True(t, errors.Is(err, sentinel),
			"errors.Is must reach sentinel through ParseError chain")
	})

	t.Run("TransformError unwraps to cause", func(t *testing.T) {
		err := NewTransformError("render", "opencode", "boom", sentinel)
		require.NotNil(t, err.Unwrap())
		assert.True(t, errors.Is(err, sentinel),
			"errors.Is must reach sentinel through TransformError chain")
	})

	t.Run("FileError unwraps to cause", func(t *testing.T) {
		err := NewFileError("/path", "read", "boom", sentinel)
		require.NotNil(t, err.Unwrap())
		assert.True(t, errors.Is(err, sentinel),
			"errors.Is must reach sentinel through FileError chain")
	})

	t.Run("InitializeError unwraps to cause", func(t *testing.T) {
		err := NewInitializeError("skill/commit", "/in.md", "/out.md", sentinel)
		require.NotNil(t, err.Unwrap())
		assert.True(t, errors.Is(err, sentinel),
			"errors.Is must reach sentinel through InitializeError chain")
	})

	t.Run("OperationError unwraps to cause", func(t *testing.T) {
		err := NewOperationError("add", "skill/commit", sentinel)
		require.NotNil(t, err.Unwrap())
		assert.True(t, errors.Is(err, sentinel),
			"errors.Is must reach sentinel through OperationError chain")
	})

	t.Run("CobraUsageError unwraps to cause", func(t *testing.T) {
		err := MustNewCobraUsageError(sentinel)
		require.NotNil(t, err.Unwrap())
		assert.True(t, errors.Is(err, sentinel),
			"errors.Is must reach sentinel through CobraUsageError chain")
	})

	t.Run("errors.As walks the chain via Unwrap", func(t *testing.T) {
		inner := NewParseError("inner.md", "inner", nil)
		outer := fmt.Errorf("wrapped: %w", inner)
		var target *ParseError
		require.True(t, errors.As(outer, &target),
			"errors.As must traverse the chain to the typed *ParseError")
		assert.Equal(t, "inner.md", target.Path())
	})

	t.Run("ValidationError Unwrap returns nil (leaf)", func(t *testing.T) {
		err := NewValidationError("req", "field", "value", "msg")
		assert.Nil(t, err.Unwrap(),
			"ValidationError must be a leaf; Unwrap returns nil")
	})

	t.Run("UsageError Unwrap returns nil (leaf)", func(t *testing.T) {
		err := NewUsageError("--flag", "reason")
		assert.Nil(t, err.Unwrap(),
			"UsageError must be a leaf; Unwrap returns nil")
	})

	t.Run("PartialSuccessError Unwrap returns nil (leaf)", func(t *testing.T) {
		err := NewPartialSuccessError(1, 2, []InitializeError{
			*NewInitializeError("skill/x", "/x", "/y", errors.New("fail")),
		})
		assert.Nil(t, err.Unwrap(),
			"PartialSuccessError aggregates errors but is a leaf; Unwrap returns nil")
	})
}

package core

import (
	"errors"
	"reflect"
	"testing"
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

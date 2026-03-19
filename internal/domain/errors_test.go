package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseError(t *testing.T) {
	tests := []struct {
		name       string
		err        *ParseError
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:       "with cause",
			err:        NewParseError("test.yaml", "invalid YAML", fmt.Errorf("yaml: line 5")),
			wantMsg:    "parse error in test.yaml: invalid YAML: yaml: line 5",
			wantUnwrap: fmt.Errorf("yaml: line 5"),
		},
		{
			name:       "without cause",
			err:        NewParseError("agent.md", "unrecognized document type", nil),
			wantMsg:    "parse error in agent.md: unrecognized document type",
			wantUnwrap: nil,
		},
		{
			name:       "with suggestions",
			err:        NewParseError("test.yaml", "invalid YAML", nil).WithSuggestions([]string{"Check indentation", "Verify quotes"}),
			wantMsg:    "parse error in test.yaml: invalid YAML\n💡 Check indentation\n💡 Verify quotes",
			wantUnwrap: nil,
		},
		{
			name:       "with context",
			err:        NewParseError("test.yaml", "invalid YAML", nil).WithContext("while parsing agent definition"),
			wantMsg:    "parse error in test.yaml: invalid YAML",
			wantUnwrap: nil,
		},
		{
			name:       "with suggestions and cause",
			err:        NewParseError("test.yaml", "invalid YAML", fmt.Errorf("yaml: line 5")).WithSuggestions([]string{"Check indentation"}),
			wantMsg:    "parse error in test.yaml: invalid YAML: yaml: line 5\n💡 Check indentation",
			wantUnwrap: fmt.Errorf("yaml: line 5"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ParseError.Error() = %q, want %q", got, tt.wantMsg)
			}
			if tt.wantUnwrap != nil {
				if got := tt.err.Unwrap(); got == nil || got.Error() != tt.wantUnwrap.Error() {
					t.Errorf("ParseError.Unwrap() = %v, want %v", got, tt.wantUnwrap)
				}
			} else {
				if got := tt.err.Unwrap(); got != nil {
					t.Errorf("ParseError.Unwrap() = %v, want nil", got)
				}
			}
		})
	}
}

func TestParseErrorGetters(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewParseError("test.yaml", "invalid format", cause).
		WithSuggestions([]string{"Try this", "Or that"}).
		WithContext("additional context")

	t.Run("Path", func(t *testing.T) {
		if got := err.Path(); got != "test.yaml" {
			t.Errorf("Path() = %q, want %q", got, "test.yaml")
		}
	})

	t.Run("Message", func(t *testing.T) {
		if got := err.Message(); got != "invalid format" {
			t.Errorf("Message() = %q, want %q", got, "invalid format")
		}
	})

	t.Run("Cause", func(t *testing.T) {
		if got := err.Cause(); got != cause {
			t.Errorf("Cause() = %v, want %v", got, cause)
		}
	})

	t.Run("Suggestions", func(t *testing.T) {
		got := err.Suggestions()
		if len(got) != 2 || got[0] != "Try this" || got[1] != "Or that" {
			t.Errorf("Suggestions() = %v, want [Try this Or that]", got)
		}
		// Verify it returns a copy
		got[0] = "modified"
		if err.Suggestions()[0] == "modified" {
			t.Error("Suggestions() should return a copy")
		}
	})

	t.Run("Context", func(t *testing.T) {
		if got := err.Context(); got != "additional context" {
			t.Errorf("Context() = %q, want %q", got, "additional context")
		}
	})

	t.Run("empty suggestions and context", func(t *testing.T) {
		err := NewParseError("test.yaml", "error", nil)
		if got := err.Suggestions(); got != nil {
			t.Errorf("Suggestions() = %v, want nil", got)
		}
		if got := err.Context(); got != "" {
			t.Errorf("Context() = %q, want empty", got)
		}
	})
}

func TestParseErrorImmutableBuilders(t *testing.T) {
	t.Run("WithSuggestions returns new instance", func(t *testing.T) {
		err1 := NewParseError("test.yaml", "error", nil)
		err2 := err1.WithSuggestions([]string{"try this"})

		if err1 == err2 {
			t.Error("WithSuggestions should return a new instance")
		}

		if len(err1.Suggestions()) != 0 {
			t.Error("original error should not have suggestions")
		}

		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "try this" {
			t.Error("new error should have suggestions")
		}
	})

	t.Run("WithContext returns new instance", func(t *testing.T) {
		err1 := NewParseError("test.yaml", "error", nil)
		err2 := err1.WithContext("context info")

		if err1 == err2 {
			t.Error("WithContext should return a new instance")
		}

		if err1.Context() != "" {
			t.Error("original error should not have context")
		}

		if err2.Context() != "context info" {
			t.Error("new error should have context")
		}
	})

	t.Run("builders preserve existing fields", func(t *testing.T) {
		cause := fmt.Errorf("cause")
		err1 := NewParseError("test.yaml", "error", cause).WithSuggestions([]string{"hint1"})
		err2 := err1.WithContext("context info")

		if err2.Path() != "test.yaml" {
			t.Error("Path should be preserved")
		}
		if err2.Message() != "error" {
			t.Error("Message should be preserved")
		}
		if err2.Cause() != cause {
			t.Error("Cause should be preserved")
		}
		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "hint1" {
			t.Error("Suggestions should be preserved")
		}
	})

	t.Run("chained builders", func(t *testing.T) {
		err := NewParseError("test.yaml", "error", nil).
			WithSuggestions([]string{"hint1"}).
			WithContext("context")

		if len(err.Suggestions()) != 1 {
			t.Error("Suggestions should be set")
		}
		if err.Context() != "context" {
			t.Error("Context should be set")
		}
	})
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name            string
		err             *ValidationError
		wantMsg         string
		wantSuggestions []string
		wantField       string
		wantValue       string
		wantMessage     string
		wantRequest     string
	}{
		{
			name:            "with all fields and suggestions",
			err:             NewValidationError("Agent", "permission", "invalid", "invalid permission value").WithSuggestions([]string{"read", "write"}),
			wantMsg:         "validation failed for Agent.permission: invalid permission value (value: invalid)\n💡 read\n💡 write",
			wantSuggestions: []string{"read", "write"},
			wantField:       "permission",
			wantValue:       "invalid",
			wantMessage:     "invalid permission value",
			wantRequest:     "Agent",
		},
		{
			name:            "without field",
			err:             NewValidationError("Agent", "", "", "missing required field"),
			wantMsg:         "validation failed: missing required field",
			wantSuggestions: nil,
			wantField:       "",
			wantValue:       "",
			wantMessage:     "missing required field",
			wantRequest:     "Agent",
		},
		{
			name:            "with field no suggestions",
			err:             NewValidationError("Command", "model", "invalid-model", "invalid format"),
			wantMsg:         "validation failed for Command.model: invalid format (value: invalid-model)",
			wantSuggestions: nil,
			wantField:       "model",
			wantValue:       "invalid-model",
			wantMessage:     "invalid format",
			wantRequest:     "Command",
		},
		{
			name:            "with context",
			err:             NewValidationError("Skill", "name", "", "name is required").WithContext("additional context"),
			wantMsg:         "validation failed for Skill.name: name is required",
			wantSuggestions: nil,
			wantField:       "name",
			wantValue:       "",
			wantMessage:     "name is required",
			wantRequest:     "Skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.wantMsg)
			}
			got := tt.err.Suggestions()
			if len(got) != len(tt.wantSuggestions) {
				t.Errorf("ValidationError.Suggestions() = %v, want %v", got, tt.wantSuggestions)
			}
			if tt.err.Field() != tt.wantField {
				t.Errorf("ValidationError.Field() = %q, want %q", tt.err.Field(), tt.wantField)
			}
			if tt.err.Value() != tt.wantValue {
				t.Errorf("ValidationError.Value() = %q, want %q", tt.err.Value(), tt.wantValue)
			}
			if tt.err.Message() != tt.wantMessage {
				t.Errorf("ValidationError.Message() = %q, want %q", tt.err.Message(), tt.wantMessage)
			}
			if tt.err.Request() != tt.wantRequest {
				t.Errorf("ValidationError.Request() = %q, want %q", tt.err.Request(), tt.wantRequest)
			}
		})
	}
}

func TestValidationErrorImmutableBuilders(t *testing.T) {
	t.Run("WithSuggestions returns new instance", func(t *testing.T) {
		err1 := NewValidationError("Agent", "name", "", "name is required")
		err2 := err1.WithSuggestions([]string{"try this"})

		if err1 == err2 {
			t.Error("WithSuggestions should return a new instance")
		}

		if len(err1.Suggestions()) != 0 {
			t.Error("original error should not have suggestions")
		}

		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "try this" {
			t.Error("new error should have suggestions")
		}
	})

	t.Run("WithContext returns new instance", func(t *testing.T) {
		err1 := NewValidationError("Agent", "name", "", "name is required")
		err2 := err1.WithContext("additional info")

		if err1 == err2 {
			t.Error("WithContext should return a new instance")
		}

		if err1.Context() != "" {
			t.Error("original error should not have context")
		}

		if err2.Context() != "additional info" {
			t.Error("new error should have context")
		}
	})

	t.Run("Suggestions returns copy", func(t *testing.T) {
		err := NewValidationError("Agent", "name", "", "name is required").WithSuggestions([]string{"suggestion1", "suggestion2"})
		suggestions1 := err.Suggestions()
		suggestions2 := err.Suggestions()

		if &suggestions1[0] == &suggestions2[0] {
			t.Error("Suggestions should return a copy, not the original slice")
		}
	})
}

func TestTransformError(t *testing.T) {
	tests := []struct {
		name       string
		err        *TransformError
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:       "with platform and cause",
			err:        NewTransformError("render", "opencode", "template failed", fmt.Errorf("missing field")),
			wantMsg:    "transform error (render for opencode): template failed: missing field",
			wantUnwrap: fmt.Errorf("missing field"),
		},
		{
			name:       "with platform without cause",
			err:        NewTransformError("convert", "claude-code", "unsupported type", nil),
			wantMsg:    "transform error (convert for claude-code): unsupported type",
			wantUnwrap: nil,
		},
		{
			name:       "without platform with cause",
			err:        NewTransformError("process", "", "internal error", fmt.Errorf("oops")),
			wantMsg:    "transform error (process): internal error: oops",
			wantUnwrap: fmt.Errorf("oops"),
		},
		{
			name:       "without platform or cause",
			err:        NewTransformError("validate", "", "invalid state", nil),
			wantMsg:    "transform error (validate): invalid state",
			wantUnwrap: nil,
		},
		{
			name:       "with suggestions",
			err:        NewTransformError("render", "opencode", "template failed", nil).WithSuggestions([]string{"Check template path"}),
			wantMsg:    "transform error (render for opencode): template failed\n💡 Check template path",
			wantUnwrap: nil,
		},
		{
			name:       "with suggestions and cause",
			err:        NewTransformError("render", "opencode", "template failed", fmt.Errorf("missing field")).WithSuggestions([]string{"Check syntax"}),
			wantMsg:    "transform error (render for opencode): template failed: missing field\n💡 Check syntax",
			wantUnwrap: fmt.Errorf("missing field"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("TransformError.Error() = %q, want %q", got, tt.wantMsg)
			}
			if tt.wantUnwrap != nil {
				if got := tt.err.Unwrap(); got == nil || got.Error() != tt.wantUnwrap.Error() {
					t.Errorf("TransformError.Unwrap() = %v, want %v", got, tt.wantUnwrap)
				}
			} else {
				if got := tt.err.Unwrap(); got != nil {
					t.Errorf("TransformError.Unwrap() = %v, want nil", got)
				}
			}
		})
	}
}

func TestTransformErrorGetters(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewTransformError("render", "opencode", "template failed", cause).
		WithSuggestions([]string{"Try this", "Or that"}).
		WithContext("additional context")

	t.Run("Operation", func(t *testing.T) {
		if got := err.Operation(); got != "render" {
			t.Errorf("Operation() = %q, want %q", got, "render")
		}
	})

	t.Run("Platform", func(t *testing.T) {
		if got := err.Platform(); got != "opencode" {
			t.Errorf("Platform() = %q, want %q", got, "opencode")
		}
	})

	t.Run("Message", func(t *testing.T) {
		if got := err.Message(); got != "template failed" {
			t.Errorf("Message() = %q, want %q", got, "template failed")
		}
	})

	t.Run("Cause", func(t *testing.T) {
		if got := err.Cause(); got != cause {
			t.Errorf("Cause() = %v, want %v", got, cause)
		}
	})

	t.Run("Suggestions", func(t *testing.T) {
		got := err.Suggestions()
		if len(got) != 2 || got[0] != "Try this" || got[1] != "Or that" {
			t.Errorf("Suggestions() = %v, want [Try this Or that]", got)
		}
		// Verify it returns a copy
		got[0] = "modified"
		if err.Suggestions()[0] == "modified" {
			t.Error("Suggestions() should return a copy")
		}
	})

	t.Run("Context", func(t *testing.T) {
		if got := err.Context(); got != "additional context" {
			t.Errorf("Context() = %q, want %q", got, "additional context")
		}
	})

	t.Run("empty suggestions and context", func(t *testing.T) {
		err := NewTransformError("render", "opencode", "error", nil)
		if got := err.Suggestions(); got != nil {
			t.Errorf("Suggestions() = %v, want nil", got)
		}
		if got := err.Context(); got != "" {
			t.Errorf("Context() = %q, want empty", got)
		}
	})
}

func TestTransformErrorImmutableBuilders(t *testing.T) {
	t.Run("WithSuggestions returns new instance", func(t *testing.T) {
		err1 := NewTransformError("render", "opencode", "error", nil)
		err2 := err1.WithSuggestions([]string{"try this"})

		if err1 == err2 {
			t.Error("WithSuggestions should return a new instance")
		}

		if len(err1.Suggestions()) != 0 {
			t.Error("original error should not have suggestions")
		}

		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "try this" {
			t.Error("new error should have suggestions")
		}
	})

	t.Run("WithContext returns new instance", func(t *testing.T) {
		err1 := NewTransformError("render", "opencode", "error", nil)
		err2 := err1.WithContext("context info")

		if err1 == err2 {
			t.Error("WithContext should return a new instance")
		}

		if err1.Context() != "" {
			t.Error("original error should not have context")
		}

		if err2.Context() != "context info" {
			t.Error("new error should have context")
		}
	})

	t.Run("builders preserve existing fields", func(t *testing.T) {
		cause := fmt.Errorf("cause")
		err1 := NewTransformError("render", "opencode", "error", cause).WithSuggestions([]string{"hint1"})
		err2 := err1.WithContext("context info")

		if err2.Operation() != "render" {
			t.Error("Operation should be preserved")
		}
		if err2.Platform() != "opencode" {
			t.Error("Platform should be preserved")
		}
		if err2.Message() != "error" {
			t.Error("Message should be preserved")
		}
		if err2.Cause() != cause {
			t.Error("Cause should be preserved")
		}
		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "hint1" {
			t.Error("Suggestions should be preserved")
		}
	})

	t.Run("chained builders", func(t *testing.T) {
		err := NewTransformError("render", "opencode", "error", nil).
			WithSuggestions([]string{"hint1"}).
			WithContext("context")

		if len(err.Suggestions()) != 1 {
			t.Error("Suggestions should be set")
		}
		if err.Context() != "context" {
			t.Error("Context should be set")
		}
	})
}

func TestFileError(t *testing.T) {
	tests := []struct {
		name         string
		err          *FileError
		wantMsg      string
		wantUnwrap   error
		wantNotFound bool
	}{
		{
			name:         "read error with cause",
			err:          NewFileError("test.yaml", "read", "failed to read file", fmt.Errorf("permission denied")),
			wantMsg:      "file error (read test.yaml): failed to read file: permission denied",
			wantUnwrap:   fmt.Errorf("permission denied"),
			wantNotFound: false,
		},
		{
			name:         "file not found",
			err:          NewFileError("missing.yaml", "read", "file not found", nil),
			wantMsg:      "file error (read missing.yaml): file not found",
			wantUnwrap:   nil,
			wantNotFound: true,
		},
		{
			name:         "does not exist variant",
			err:          NewFileError("gone.md", "read", "file does not exist", nil),
			wantNotFound: true,
		},
		{
			name:         "no such file variant",
			err:          NewFileError("vanished.yaml", "read", "no such file or directory", nil),
			wantNotFound: true,
		},
		{
			name:         "write error",
			err:          NewFileError("output.md", "write", "disk full", nil),
			wantMsg:      "file error (write output.md): disk full",
			wantNotFound: false,
		},
		{
			name:         "with suggestions",
			err:          NewFileError("test.yaml", "read", "failed", nil).WithSuggestions([]string{"Check file permissions"}),
			wantMsg:      "file error (read test.yaml): failed\n💡 Check file permissions",
			wantNotFound: false,
		},
		{
			name:         "with suggestions and cause",
			err:          NewFileError("test.yaml", "read", "failed", fmt.Errorf("permission denied")).WithSuggestions([]string{"Check permissions"}),
			wantMsg:      "file error (read test.yaml): failed: permission denied\n💡 Check permissions",
			wantUnwrap:   fmt.Errorf("permission denied"),
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMsg != "" {
				if got := tt.err.Error(); got != tt.wantMsg {
					t.Errorf("FileError.Error() = %q, want %q", got, tt.wantMsg)
				}
			}
			if got := tt.err.IsNotFound(); got != tt.wantNotFound {
				t.Errorf("FileError.IsNotFound() = %v, want %v", got, tt.wantNotFound)
			}
			if tt.wantUnwrap != nil {
				if got := tt.err.Unwrap(); got == nil || got.Error() != tt.wantUnwrap.Error() {
					t.Errorf("FileError.Unwrap() = %v, want %v", got, tt.wantUnwrap)
				}
			} else {
				if got := tt.err.Unwrap(); got != nil {
					t.Errorf("FileError.Unwrap() = %v, want nil", got)
				}
			}
		})
	}
}

func TestFileErrorGetters(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewFileError("test.yaml", "read", "failed to read", cause).
		WithSuggestions([]string{"Try this", "Or that"}).
		WithContext("additional context")

	t.Run("Path", func(t *testing.T) {
		if got := err.Path(); got != "test.yaml" {
			t.Errorf("Path() = %q, want %q", got, "test.yaml")
		}
	})

	t.Run("Operation", func(t *testing.T) {
		if got := err.Operation(); got != "read" {
			t.Errorf("Operation() = %q, want %q", got, "read")
		}
	})

	t.Run("Message", func(t *testing.T) {
		if got := err.Message(); got != "failed to read" {
			t.Errorf("Message() = %q, want %q", got, "failed to read")
		}
	})

	t.Run("Cause", func(t *testing.T) {
		if got := err.Cause(); got != cause {
			t.Errorf("Cause() = %v, want %v", got, cause)
		}
	})

	t.Run("Suggestions", func(t *testing.T) {
		got := err.Suggestions()
		if len(got) != 2 || got[0] != "Try this" || got[1] != "Or that" {
			t.Errorf("Suggestions() = %v, want [Try this Or that]", got)
		}
		// Verify it returns a copy
		got[0] = "modified"
		if err.Suggestions()[0] == "modified" {
			t.Error("Suggestions() should return a copy")
		}
	})

	t.Run("Context", func(t *testing.T) {
		if got := err.Context(); got != "additional context" {
			t.Errorf("Context() = %q, want %q", got, "additional context")
		}
	})

	t.Run("empty suggestions and context", func(t *testing.T) {
		err := NewFileError("test.yaml", "read", "error", nil)
		if got := err.Suggestions(); got != nil {
			t.Errorf("Suggestions() = %v, want nil", got)
		}
		if got := err.Context(); got != "" {
			t.Errorf("Context() = %q, want empty", got)
		}
	})
}

func TestFileErrorImmutableBuilders(t *testing.T) {
	t.Run("WithSuggestions returns new instance", func(t *testing.T) {
		err1 := NewFileError("test.yaml", "read", "error", nil)
		err2 := err1.WithSuggestions([]string{"try this"})

		if err1 == err2 {
			t.Error("WithSuggestions should return a new instance")
		}

		if len(err1.Suggestions()) != 0 {
			t.Error("original error should not have suggestions")
		}

		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "try this" {
			t.Error("new error should have suggestions")
		}
	})

	t.Run("WithContext returns new instance", func(t *testing.T) {
		err1 := NewFileError("test.yaml", "read", "error", nil)
		err2 := err1.WithContext("context info")

		if err1 == err2 {
			t.Error("WithContext should return a new instance")
		}

		if err1.Context() != "" {
			t.Error("original error should not have context")
		}

		if err2.Context() != "context info" {
			t.Error("new error should have context")
		}
	})

	t.Run("builders preserve existing fields", func(t *testing.T) {
		cause := fmt.Errorf("cause")
		err1 := NewFileError("test.yaml", "read", "error", cause).WithSuggestions([]string{"hint1"})
		err2 := err1.WithContext("context info")

		if err2.Path() != "test.yaml" {
			t.Error("Path should be preserved")
		}
		if err2.Operation() != "read" {
			t.Error("Operation should be preserved")
		}
		if err2.Message() != "error" {
			t.Error("Message should be preserved")
		}
		if err2.Cause() != cause {
			t.Error("Cause should be preserved")
		}
		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "hint1" {
			t.Error("Suggestions should be preserved")
		}
	})

	t.Run("chained builders", func(t *testing.T) {
		err := NewFileError("test.yaml", "read", "error", nil).
			WithSuggestions([]string{"hint1"}).
			WithContext("context")

		if len(err.Suggestions()) != 1 {
			t.Error("Suggestions should be set")
		}
		if err.Context() != "context" {
			t.Error("Context should be set")
		}
	})
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ConfigError
		wantMsg string
	}{
		{
			name:    "with suggestions",
			err:     NewConfigError("platform", "invalid", "unknown platform").WithSuggestions([]string{"claude-code", "opencode"}),
			wantMsg: "config error: invalid platform 'invalid': unknown platform\n💡 claude-code\n💡 opencode",
		},
		{
			name:    "with field and value",
			err:     NewConfigError("type", "invalid", "must be one of the supported types"),
			wantMsg: "config error: invalid type 'invalid': must be one of the supported types",
		},
		{
			name:    "message only",
			err:     NewConfigError("", "", "missing required configuration"),
			wantMsg: "config error: missing required configuration",
		},
		{
			name:    "with context",
			err:     NewConfigError("platform", "", "platform is required").WithContext("use --platform flag"),
			wantMsg: "config error: platform is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ConfigError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestConfigErrorGetters(t *testing.T) {
	err := NewConfigError("platform", "invalid", "unknown platform").
		WithSuggestions([]string{"claude-code", "opencode"}).
		WithContext("use --platform flag")

	t.Run("Field", func(t *testing.T) {
		if got := err.Field(); got != "platform" {
			t.Errorf("Field() = %q, want %q", got, "platform")
		}
	})

	t.Run("Value", func(t *testing.T) {
		if got := err.Value(); got != "invalid" {
			t.Errorf("Value() = %q, want %q", got, "invalid")
		}
	})

	t.Run("Message", func(t *testing.T) {
		if got := err.Message(); got != "unknown platform" {
			t.Errorf("Message() = %q, want %q", got, "unknown platform")
		}
	})

	t.Run("Suggestions", func(t *testing.T) {
		got := err.Suggestions()
		if len(got) != 2 || got[0] != "claude-code" || got[1] != "opencode" {
			t.Errorf("Suggestions() = %v, want [claude-code opencode]", got)
		}
		// Verify it returns a copy
		got[0] = "modified"
		if err.Suggestions()[0] == "modified" {
			t.Error("Suggestions() should return a copy")
		}
	})

	t.Run("Context", func(t *testing.T) {
		if got := err.Context(); got != "use --platform flag" {
			t.Errorf("Context() = %q, want %q", got, "use --platform flag")
		}
	})

	t.Run("empty suggestions and context", func(t *testing.T) {
		err := NewConfigError("field", "value", "error")
		if got := err.Suggestions(); got != nil {
			t.Errorf("Suggestions() = %v, want nil", got)
		}
		if got := err.Context(); got != "" {
			t.Errorf("Context() = %q, want empty", got)
		}
	})
}

func TestConfigErrorImmutableBuilders(t *testing.T) {
	t.Run("WithSuggestions returns new instance", func(t *testing.T) {
		err1 := NewConfigError("platform", "invalid", "unknown")
		err2 := err1.WithSuggestions([]string{"claude-code"})

		if err1 == err2 {
			t.Error("WithSuggestions should return a new instance")
		}

		if len(err1.Suggestions()) != 0 {
			t.Error("original error should not have suggestions")
		}

		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "claude-code" {
			t.Error("new error should have suggestions")
		}
	})

	t.Run("WithContext returns new instance", func(t *testing.T) {
		err1 := NewConfigError("platform", "invalid", "unknown")
		err2 := err1.WithContext("context info")

		if err1 == err2 {
			t.Error("WithContext should return a new instance")
		}

		if err1.Context() != "" {
			t.Error("original error should not have context")
		}

		if err2.Context() != "context info" {
			t.Error("new error should have context")
		}
	})

	t.Run("builders preserve existing fields", func(t *testing.T) {
		err1 := NewConfigError("platform", "invalid", "unknown").WithSuggestions([]string{"claude-code"})
		err2 := err1.WithContext("context info")

		if err2.Field() != "platform" {
			t.Error("Field should be preserved")
		}
		if err2.Value() != "invalid" {
			t.Error("Value should be preserved")
		}
		if err2.Message() != "unknown" {
			t.Error("Message should be preserved")
		}
		if len(err2.Suggestions()) != 1 || err2.Suggestions()[0] != "claude-code" {
			t.Error("Suggestions should be preserved")
		}
	})

	t.Run("chained builders", func(t *testing.T) {
		err := NewConfigError("platform", "invalid", "unknown").
			WithSuggestions([]string{"claude-code"}).
			WithContext("context")

		if len(err.Suggestions()) != 1 {
			t.Error("Suggestions should be set")
		}
		if err.Context() != "context" {
			t.Error("Context should be set")
		}
	})
}

func TestErrorsAsDetection(t *testing.T) {
	t.Run("ParseError", func(t *testing.T) {
		err := NewParseError("test.yaml", "failed", nil)
		var target *ParseError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect ParseError")
		}
		if target.Path() != "test.yaml" {
			t.Errorf("target.Path() = %q, want %q", target.Path(), "test.yaml")
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := NewValidationError("Agent", "field", "", "invalid")
		var target *ValidationError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect ValidationError")
		}
		if target.Message() != "invalid" {
			t.Errorf("target.Message() = %q, want %q", target.Message(), "invalid")
		}
	})

	t.Run("TransformError", func(t *testing.T) {
		err := NewTransformError("op", "platform", "failed", nil)
		var target *TransformError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect TransformError")
		}
		if target.Operation() != "op" {
			t.Errorf("target.Operation() = %q, want %q", target.Operation(), "op")
		}
	})

	t.Run("FileError", func(t *testing.T) {
		err := NewFileError("path", "read", "failed", nil)
		var target *FileError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect FileError")
		}
		if target.Path() != "path" {
			t.Errorf("target.Path() = %q, want %q", target.Path(), "path")
		}
	})

	t.Run("ConfigError", func(t *testing.T) {
		err := NewConfigError("field", "value", "invalid")
		var target *ConfigError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect ConfigError")
		}
		if target.Field() != "field" {
			t.Errorf("target.Field() = %q, want %q", target.Field(), "field")
		}
	})
}

func TestWrappedErrors(t *testing.T) {
	inner := fmt.Errorf("inner error")
	wrapped := NewParseError("file.yaml", "outer", inner)

	var target *ParseError
	if !errors.As(wrapped, &target) {
		t.Error("errors.As failed on wrapped error")
	}

	if !errors.Is(wrapped, wrapped) {
		t.Error("errors.Is failed on self-comparison")
	}
}

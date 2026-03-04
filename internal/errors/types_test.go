package errors

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

func TestConfigError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ConfigError
		wantMsg string
	}{
		{
			name:    "with available options",
			err:     NewConfigError("platform", "invalid", []string{"claude-code", "opencode"}, "unknown platform"),
			wantMsg: "config error: unknown platform (available: claude-code, opencode)",
		},
		{
			name:    "with field and value",
			err:     NewConfigError("type", "invalid", nil, "must be one of the supported types"),
			wantMsg: "config error: invalid type 'invalid': must be one of the supported types",
		},
		{
			name:    "message only",
			err:     NewConfigError("", "", nil, "missing required configuration"),
			wantMsg: "config error: missing required configuration",
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

func TestErrorsAsDetection(t *testing.T) {
	t.Run("ParseError", func(t *testing.T) {
		err := NewParseError("test.yaml", "failed", nil)
		var target *ParseError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect ParseError")
		}
		if target.Path != "test.yaml" {
			t.Errorf("target.Path = %q, want %q", target.Path, "test.yaml")
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
		if target.Operation != "op" {
			t.Errorf("target.Operation = %q, want %q", target.Operation, "op")
		}
	})

	t.Run("FileError", func(t *testing.T) {
		err := NewFileError("path", "read", "failed", nil)
		var target *FileError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect FileError")
		}
		if target.Path != "path" {
			t.Errorf("target.Path = %q, want %q", target.Path, "path")
		}
	})

	t.Run("ConfigError", func(t *testing.T) {
		err := NewConfigError("field", "value", nil, "invalid")
		var target *ConfigError
		if !errors.As(err, &target) {
			t.Error("errors.As failed to detect ConfigError")
		}
		if target.Field != "field" {
			t.Errorf("target.Field = %q, want %q", target.Field, "field")
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

package cmd

import (
	"fmt"
	"strings"
	"testing"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

func TestErrorFormatter_Format(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "ParseError",
			err:      gerrors.NewParseError("test.yaml", "invalid YAML", fmt.Errorf("line 5")),
			contains: []string{"Parse error:", "test.yaml", "invalid YAML"},
		},
		{
			name:     "ValidationError with suggestions",
			err:      gerrors.NewValidationError("", "permission", "", "invalid value").WithSuggestions([]string{"try read", "try write"}),
			contains: []string{"Validation error:", "invalid value", "field: permission", "Hint: try read", "Hint: try write"},
		},
		{
			name:     "ValidationError without suggestions",
			err:      gerrors.NewValidationError("", "model", "", "missing field"),
			contains: []string{"Validation error:", "missing field", "field: model"},
		},
		{
			name:     "TransformError with platform",
			err:      gerrors.NewTransformError("render", "opencode", "template failed", nil),
			contains: []string{"Transform error", "render", "opencode", "template failed"},
		},
		{
			name:     "FileError",
			err:      gerrors.NewFileError("test.yaml", "read", "file not found", nil),
			contains: []string{"File error", "read", "test.yaml"},
		},
		{
			name:     "ConfigError with available",
			err:      gerrors.NewConfigError("platform", "invalid", "unknown platform").WithSuggestions([]string{"claude-code", "opencode"}),
			contains: []string{"Config error", "unknown platform", "Hint:", "claude-code, opencode"},
		},
		{
			name:     "Generic error",
			err:      fmt.Errorf("something went wrong"),
			contains: []string{"Error:", "something went wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.Format(tt.err)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Format() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestErrorFormatter_WrappedErrors(t *testing.T) {
	formatter := NewErrorFormatter()

	inner := fmt.Errorf("inner error")
	wrapped := gerrors.NewParseError("file.yaml", "outer", inner)

	got := formatter.Format(wrapped)

	if !strings.Contains(got, "Parse error:") {
		t.Error("should format as ParseError even when wrapped")
	}
	if !strings.Contains(got, "inner error") {
		t.Error("should include cause in output")
	}
}

func TestErrorFormatter_DefaultFormat(t *testing.T) {
	formatter := NewErrorFormatter()

	customErr := fmt.Errorf("custom error")

	got := formatter.Format(customErr)

	if !strings.Contains(got, "Error:") {
		t.Error("default format should include 'Error:' prefix")
	}
}

func TestNewErrorFormatterReadyToUse(t *testing.T) {
	formatter := NewErrorFormatter()

	if formatter == nil {
		t.Fatal("NewErrorFormatter() returned nil")
	}

	tests := []struct {
		name  string
		err   error
		check string
	}{
		{
			name:  "ParseError formatted",
			err:   gerrors.NewParseError("test.yaml", "failed", nil),
			check: "Parse error:",
		},
		{
			name:  "ValidationError formatted",
			err:   gerrors.NewValidationError("", "field", "", "invalid"),
			check: "Validation error:",
		},
		{
			name:  "TransformError formatted",
			err:   gerrors.NewTransformError("op", "platform", "failed", nil),
			check: "Transform error",
		},
		{
			name:  "FileError formatted",
			err:   gerrors.NewFileError("path", "read", "failed", nil),
			check: "File error",
		},
		{
			name:  "ConfigError formatted",
			err:   gerrors.NewConfigError("field", "value", "invalid"),
			check: "Config error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.Format(tt.err)
			if !strings.Contains(got, tt.check) {
				t.Errorf("NewErrorFormatter() should format %s, got: %q", tt.name, got)
			}
		})
	}
}

func TestFormatMultipleErrors(t *testing.T) {
	formatter := NewErrorFormatter()

	err1 := gerrors.NewValidationError("", "name", "", "missing name")
	err2 := gerrors.NewValidationError("", "temperature", "", "invalid temperature").WithSuggestions([]string{"0.0 - 1.0"})

	var formatted strings.Builder
	for i, err := range []error{err1, err2} {
		fmt.Fprintf(&formatted, "%d. %s", i+1, formatter.Format(err))
	}

	result := formatted.String()

	if !strings.Contains(result, "missing name") {
		t.Error("should contain first error message")
	}
	if !strings.Contains(result, "invalid temperature") {
		t.Error("should contain second error message")
	}
	if !strings.Contains(result, "Hint: 0.0 - 1.0") {
		t.Error("should contain suggestion for second error")
	}
}

func TestFormatValidationErrorWithMultipleSuggestions(t *testing.T) {
	formatter := NewErrorFormatter()

	err := gerrors.NewValidationError(
		"",
		"mode",
		"",
		"invalid mode",
	).WithSuggestions([]string{"primary", "subagent", "all"})

	got := formatter.Format(err)

	if !strings.Contains(got, "Validation error:") {
		t.Error("should contain Validation error prefix")
	}
	if !strings.Contains(got, "Hint: primary") {
		t.Error("should contain first suggestion")
	}
	if !strings.Contains(got, "Hint: subagent") {
		t.Error("should contain second suggestion")
	}
	if !strings.Contains(got, "Hint: all") {
		t.Error("should contain third suggestion")
	}
}

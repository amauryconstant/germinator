package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVerbosityMethods(t *testing.T) {
	tests := []struct {
		name            string
		verbosity       Verbosity
		wantVerbose     bool
		wantVeryVerbose bool
	}{
		{
			name:            "level 0",
			verbosity:       0,
			wantVerbose:     false,
			wantVeryVerbose: false,
		},
		{
			name:            "level 1",
			verbosity:       1,
			wantVerbose:     true,
			wantVeryVerbose: false,
		},
		{
			name:            "level 2",
			verbosity:       2,
			wantVerbose:     true,
			wantVeryVerbose: true,
		},
		{
			name:            "level 3",
			verbosity:       3,
			wantVerbose:     true,
			wantVeryVerbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.verbosity.IsVerbose(); got != tt.wantVerbose {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.wantVerbose)
			}
			if got := tt.verbosity.IsVeryVerbose(); got != tt.wantVeryVerbose {
				t.Errorf("IsVeryVerbose() = %v, want %v", got, tt.wantVeryVerbose)
			}
		})
	}
}

func TestVerbosePrint(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &CommandConfig{
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      1,
	}

	VerbosePrint(cfg, "test message: %s", "value")

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test message: value") {
		t.Errorf("VerbosePrint() output = %q, should contain %q", output, "test message: value")
	}
}

func TestVerbosePrintLevelZero(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &CommandConfig{
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      0,
	}

	VerbosePrint(cfg, "test message")

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if output != "" {
		t.Errorf("VerbosePrint() at level 0 should produce no output, got: %q", output)
	}
}

func TestVeryVerbosePrint(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &CommandConfig{
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      2,
	}

	VeryVerbosePrint(cfg, "detail: %s", "info")

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "  detail: info") {
		t.Errorf("VeryVerbosePrint() output = %q, should contain %q", output, "  detail: info")
	}
}

func TestVeryVerbosePrintLevelOne(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &CommandConfig{
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      1,
	}

	VeryVerbosePrint(cfg, "detail")

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if output != "" {
		t.Errorf("VeryVerbosePrint() at level 1 should produce no output, got: %q", output)
	}
}

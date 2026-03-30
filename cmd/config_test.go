package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigInitCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		outputPath  string
		force       bool
		setup       func(t *testing.T, path string)
		expectError bool
		errorMsg    string
		checkFile   bool
	}{
		{
			name:        "creates config at custom path",
			outputPath:  filepath.Join(tmpDir, "config.toml"),
			force:       false,
			setup:       nil,
			expectError: false,
			checkFile:   true,
		},
		{
			name:       "refuses to overwrite without force",
			outputPath: filepath.Join(tmpDir, "config2.toml"),
			force:      false,
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}
			},
			expectError: true,
			errorMsg:    "already exists",
			checkFile:   false,
		},
		{
			name:       "overwrites with force flag",
			outputPath: filepath.Join(tmpDir, "config3.toml"),
			force:      true,
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}
			},
			expectError: false,
			checkFile:   true,
		},
		{
			name:        "creates parent directories",
			outputPath:  filepath.Join(tmpDir, "subdir", "config.toml"),
			force:       false,
			setup:       nil,
			expectError: false,
			checkFile:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t, tt.outputPath)
			}

			cfg := &CommandConfig{
				Services:       NewServiceContainer(),
				ErrorFormatter: NewErrorFormatter(),
			}

			cmd := NewConfigInitCommand(cfg)
			var args []string
			args = append(args, "--output", tt.outputPath)
			if tt.force {
				args = append(args, "--force")
			}
			cmd.SetArgs(args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := cmd.Execute()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.checkFile {
				content, err := os.ReadFile(tt.outputPath)
				if err != nil {
					t.Errorf("Failed to read config file: %v", err)
				}
				if !strings.Contains(string(content), "library = ") {
					t.Error("Config file does not contain expected content")
				}
			}
		})
	}
}

func TestConfigInitCommand_ScaffoldedContent(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "config.toml")

	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewConfigInitCommand(cfg)
	cmd.SetArgs([]string{"--output", outputPath})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Check all expected fields are present
	expectedContent := []string{
		"# Germinator configuration",
		"library = \"~/.config/germinator/library\"",
		"platform = \"\"",
		"[completion]",
		`timeout = "500ms"`,
		`cache_ttl = "5s"`,
	}

	for _, expected := range expectedContent {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Config file does not contain expected line: %q", expected)
		}
	}
}

func TestConfigValidateCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // returns outputPath
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config passes",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "valid.toml")
				content := `library = "~/.config/germinator/library"
platform = "opencode"
[completion]
timeout = "500ms"
cache_ttl = "5s"
`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create valid config: %v", err)
				}
				return path
			},
			expectError: false,
		},
		{
			name: "file not found",
			setup: func(_ *testing.T) string {
				return filepath.Join(tmpDir, "nonexistent.toml")
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "invalid TOML syntax",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "invalid_toml.toml")
				if err := os.WriteFile(path, []byte("invalid [ ["), 0644); err != nil {
					t.Fatalf("Failed to create invalid config: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "failed to parse config file",
		},
		{
			name: "invalid platform value",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "invalid_platform.toml")
				content := `platform = "unknown"
`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create invalid platform config: %v", err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "unknown platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := tt.setup(t)

			cfg := &CommandConfig{
				Services:       NewServiceContainer(),
				ErrorFormatter: NewErrorFormatter(),
			}

			cmd := NewConfigValidateCommand(cfg)
			cmd.SetArgs([]string{"--output", outputPath})

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := cmd.Execute()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				output := buf.String()
				if !strings.Contains(output, "valid") {
					t.Errorf("Expected success message containing 'valid', got %q", output)
				}
			}
		})
	}
}

func TestConfigCommand_Help(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewConfigCommand(cfg)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "config init") {
		t.Error("Help output does not mention 'config init'")
	}
	if !strings.Contains(output, "config validate") {
		t.Error("Help output does not mention 'config validate'")
	}
}

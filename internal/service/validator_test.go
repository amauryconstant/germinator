package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/domain"
)

func TestValidator_Validate(t *testing.T) {
	fixturesDir := filepath.Join("..", "..", "test", "fixtures")

	tests := []struct {
		name        string
		inputPath   string
		platform    string
		wantValid   bool
		wantErr     bool
		errType     any
		errContains string
	}{
		{
			name:      "valid agent - returns empty errors",
			inputPath: filepath.Join(fixturesDir, "agent-valid.md"),
			platform:  PlatformClaudeCode,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "valid command - returns empty errors",
			inputPath: filepath.Join(fixturesDir, "command-valid.md"),
			platform:  PlatformClaudeCode,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "valid skill - returns empty errors",
			inputPath: filepath.Join(fixturesDir, "skill-valid.md"),
			platform:  PlatformClaudeCode,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "valid memory - returns empty errors",
			inputPath: filepath.Join(fixturesDir, "memory-valid.md"),
			platform:  PlatformClaudeCode,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:        "invalid agent - returns validation errors",
			inputPath:   filepath.Join(fixturesDir, "agent-invalid.md"),
			platform:    PlatformClaudeCode,
			wantValid:   false,
			wantErr:     false,
			errContains: "name",
		},
		{
			name:        "invalid command - parse error from invalid YAML",
			inputPath:   filepath.Join(fixturesDir, "command-invalid.md"),
			platform:    PlatformClaudeCode,
			wantValid:   false,
			wantErr:     true,
			errType:     &gerrors.ParseError{},
			errContains: "parse",
		},
		{
			name:      "invalid skill - returns validation errors",
			inputPath: filepath.Join(fixturesDir, "skill-invalid.md"),
			platform:  PlatformClaudeCode,
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "memory with empty frontmatter but content is valid",
			inputPath: filepath.Join(fixturesDir, "memory-invalid.md"),
			platform:  PlatformClaudeCode,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:        "invalid platform - returns config errors",
			inputPath:   filepath.Join(fixturesDir, "agent-valid.md"),
			platform:    "invalid-platform",
			wantValid:   false,
			wantErr:     false,
			errContains: "platform",
		},
		{
			name:        "unrecognized file - returns parse error",
			inputPath:   filepath.Join(fixturesDir, "my-document.md"),
			platform:    PlatformClaudeCode,
			wantValid:   false,
			wantErr:     true,
			errType:     &gerrors.ParseError{},
			errContains: "unrecognizable",
		},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(context.Background(), &application.ValidateRequest{
				InputPath: tt.inputPath,
				Platform:  tt.platform,
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errType != nil {
					switch e := tt.errType.(type) {
					case **gerrors.ParseError:
						if !errors.As(err, e) {
							t.Errorf("expected %T, got %T", e, err)
						}
					case **gerrors.ConfigError:
						if !errors.As(err, e) {
							t.Errorf("expected %T, got %T", e, err)
						}
					}
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.Valid() != tt.wantValid {
				if tt.wantValid {
					t.Errorf("expected valid result, got errors: %v", result.Errors)
				} else {
					t.Errorf("expected invalid result, got valid")
				}
			}

			if tt.errContains != "" && !tt.wantValid {
				found := false
				for _, e := range result.Errors {
					if strings.Contains(e.Error(), tt.errContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got: %v", tt.errContains, result.Errors)
				}
			}
		})
	}
}

func TestValidator_OpenCodeSpecificValidation(t *testing.T) {
	v := NewValidator()

	t.Run("OpenCode mode validation triggers", func(t *testing.T) {
		tests := []struct {
			name        string
			content     string
			filename    string
			platform    string
			wantValid   bool
			errContains string
		}{
			{
				name: "valid mode primary",
				content: `---
name: test-agent
description: Test agent
behavior:
  mode: primary
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformOpenCode,
				wantValid: true,
			},
			{
				name: "valid mode subagent",
				content: `---
name: test-agent
description: Test agent
behavior:
  mode: subagent
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformOpenCode,
				wantValid: true,
			},
			{
				name: "valid mode all",
				content: `---
name: test-agent
description: Test agent
behavior:
  mode: all
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformOpenCode,
				wantValid: true,
			},
			{
				name: "invalid mode value",
				content: `---
name: test-agent
description: Test agent
behavior:
  mode: invalid-mode
---
Content`,
				filename:    "test-agent.md",
				platform:    PlatformOpenCode,
				wantValid:   false,
				errContains: "mode",
			},
			{
				name: "valid temperature in range",
				content: `---
name: test-agent
description: Test agent
behavior:
  temperature: 0.5
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformOpenCode,
				wantValid: true,
			},
			{
				name: "valid temperature at min",
				content: `---
name: test-agent
description: Test agent
behavior:
  temperature: 0.0
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformOpenCode,
				wantValid: true,
			},
			{
				name: "valid temperature at max",
				content: `---
name: test-agent
description: Test agent
behavior:
  temperature: 1.0
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformOpenCode,
				wantValid: true,
			},
			{
				name: "invalid temperature above max",
				content: `---
name: test-agent
description: Test agent
behavior:
  temperature: 1.5
---
Content`,
				filename:    "test-agent.md",
				platform:    PlatformOpenCode,
				wantValid:   false,
				errContains: "temperature",
			},
			{
				name: "invalid temperature negative",
				content: `---
name: test-agent
description: Test agent
behavior:
  temperature: -0.5
---
Content`,
				filename:    "test-agent.md",
				platform:    PlatformOpenCode,
				wantValid:   false,
				errContains: "temperature",
			},
			{
				name: "claude-code ignores invalid mode",
				content: `---
name: test-agent
description: Test agent
behavior:
  mode: invalid-mode
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformClaudeCode,
				wantValid: true,
			},
			{
				name: "claude-code ignores invalid temperature",
				content: `---
name: test-agent
description: Test agent
behavior:
  temperature: 2.5
---
Content`,
				filename:  "test-agent.md",
				platform:  PlatformClaudeCode,
				wantValid: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				inputPath := filepath.Join(tmpDir, tt.filename)

				if err := writeFile(inputPath, tt.content); err != nil {
					t.Fatalf("failed to write file: %v", err)
				}

				result, err := v.Validate(context.Background(), &application.ValidateRequest{
					InputPath: inputPath,
					Platform:  tt.platform,
				})

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result.Valid() != tt.wantValid {
					if tt.wantValid {
						t.Errorf("expected valid result, got errors: %v", result.Errors)
					} else {
						t.Errorf("expected invalid result, got valid")
					}
				}

				if tt.errContains != "" && !tt.wantValid {
					found := false
					for _, e := range result.Errors {
						if strings.Contains(e.Error(), tt.errContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errContains, result.Errors)
					}
				}
			})
		}
	})
}

func TestValidator_InvalidPlatformReturnsConfigError(t *testing.T) {
	v := NewValidator()
	fixturesDir := filepath.Join("..", "..", "test", "fixtures")

	result, err := v.Validate(context.Background(), &application.ValidateRequest{
		InputPath: filepath.Join(fixturesDir, "agent-valid.md"),
		Platform:  "unknown-platform",
	})

	if err != nil {
		t.Fatalf("Validate should not return fatal error: %v", err)
	}

	if result.Valid() {
		t.Fatal("Expected validation errors for invalid platform")
	}

	foundConfigError := false
	for _, e := range result.Errors {
		var configErr *gerrors.ConfigError
		if errors.As(e, &configErr) {
			foundConfigError = true
			if configErr.Field() != "platform" {
				t.Errorf("ConfigError.Field() = %q, want 'platform'", configErr.Field())
			}
			break
		}
	}

	if !foundConfigError {
		t.Error("Expected ConfigError in validation errors")
	}
}

func TestValidator_UnrecognizedFileReturnsParseError(t *testing.T) {
	v := NewValidator()
	fixturesDir := filepath.Join("..", "..", "test", "fixtures")

	_, err := v.Validate(context.Background(), &application.ValidateRequest{
		InputPath: filepath.Join(fixturesDir, "my-document.md"),
		Platform:  PlatformClaudeCode,
	})

	if err == nil {
		t.Fatal("Expected error for unrecognizable filename")
	}

	var parseErr *gerrors.ParseError
	if !errors.As(err, &parseErr) {
		t.Errorf("Expected ParseError, got %T: %v", err, err)
	} else {
		if !strings.Contains(parseErr.Message(), "unrecognizable") {
			t.Errorf("ParseError.Message() should contain 'unrecognizable', got: %q", parseErr.Message())
		}
	}
}

func TestValidator_MissingFileReturnsError(t *testing.T) {
	v := NewValidator()

	_, err := v.Validate(context.Background(), &application.ValidateRequest{
		InputPath: "/nonexistent/file.md",
		Platform:  PlatformClaudeCode,
	})

	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestUnwrapErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantLen  int
		wantMsgs []string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantLen: 0,
		},
		{
			name:     "single error returns slice",
			err:      errors.New("single error"),
			wantLen:  1,
			wantMsgs: []string{"single error"},
		},
		{
			name:     "joined error returns multiple",
			err:      errors.Join(errors.New("error 1"), errors.New("error 2")),
			wantLen:  2,
			wantMsgs: []string{"error 1", "error 2"},
		},
		{
			name:     "three joined errors",
			err:      errors.Join(errors.New("a"), errors.New("b"), errors.New("c")),
			wantLen:  3,
			wantMsgs: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := unwrapErrors(tt.err)

			if len(errs) != tt.wantLen {
				t.Errorf("unwrapErrors() returned %d errors, want %d", len(errs), tt.wantLen)
			}

			for i, msg := range tt.wantMsgs {
				if i >= len(errs) {
					t.Errorf("missing error %d: %q", i, msg)
					continue
				}
				if errs[i].Error() != msg {
					t.Errorf("error[%d] = %q, want %q", i, errs[i].Error(), msg)
				}
			}
		})
	}
}

func TestValidator_AllDocumentTypesWithBothPlatforms(t *testing.T) {
	v := NewValidator()
	fixturesDir := filepath.Join("..", "..", "test", "fixtures")

	tests := []struct {
		name      string
		inputPath string
		platforms []string
		wantValid bool
	}{
		{
			name:      "agent",
			inputPath: filepath.Join(fixturesDir, "agent-valid.md"),
			platforms: []string{PlatformClaudeCode, PlatformOpenCode},
			wantValid: true,
		},
		{
			name:      "command",
			inputPath: filepath.Join(fixturesDir, "command-valid.md"),
			platforms: []string{PlatformClaudeCode, PlatformOpenCode},
			wantValid: true,
		},
		{
			name:      "skill",
			inputPath: filepath.Join(fixturesDir, "skill-valid.md"),
			platforms: []string{PlatformClaudeCode, PlatformOpenCode},
			wantValid: true,
		},
		{
			name:      "memory",
			inputPath: filepath.Join(fixturesDir, "memory-valid.md"),
			platforms: []string{PlatformClaudeCode, PlatformOpenCode},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		for _, platform := range tt.platforms {
			t.Run(tt.name+"_"+platform, func(t *testing.T) {
				result, err := v.Validate(context.Background(), &application.ValidateRequest{
					InputPath: tt.inputPath,
					Platform:  platform,
				})

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result.Valid() != tt.wantValid {
					if tt.wantValid {
						t.Errorf("expected valid result, got errors: %v", result.Errors)
					} else {
						t.Errorf("expected invalid result, got valid")
					}
				}
			})
		}
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

package models

import (
	"testing"
)

func TestAgentValidate(t *testing.T) {
	tests := []struct {
		name        string
		agent       Agent
		expectError bool
		errorCount  int
	}{
		{
			name: "valid agent with all required fields",
			agent: Agent{
				Name:        "code-reviewer",
				Description: "Reviews code for quality and best practices",
			},
			expectError: false,
		},
		{
			name: "valid agent with optional fields",
			agent: Agent{
				Name:            "safe-researcher",
				Description:     "Research agent with restricted capabilities",
				Tools:           []string{"Read", "Grep", "Glob"},
				DisallowedTools: []string{"Write", "Edit"},
				Model:           "sonnet",
				PermissionMode:  "default",
				Skills:          []string{"code-analysis"},
			},
			expectError: false,
		},
		{
			name: "valid agent with opus model",
			agent: Agent{
				Name:           "aggressive-optimizer",
				Description:    "Performance optimizer",
				Model:          "opus",
				PermissionMode: "bypassPermissions",
			},
			expectError: false,
		},
		{
			name:        "missing name",
			agent:       Agent{Description: "A description"},
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "missing description",
			agent:       Agent{Name: "test-agent"},
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "missing both required fields",
			agent:       Agent{},
			expectError: true,
			errorCount:  2,
		},
		{
			name: "invalid name with uppercase",
			agent: Agent{
				Name:        "Code-Reviewer",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid name with spaces",
			agent: Agent{
				Name:        "code reviewer",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid name starts with hyphen",
			agent: Agent{
				Name:        "-test-agent",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "valid model - any value allowed",
			agent: Agent{
				Name:        "test-agent",
				Description: "A description",
				Model:       "anthropic/claude-sonnet-4-20250514",
			},
			expectError: false,
		},
		{
			name: "invalid permission mode",
			agent: Agent{
				Name:           "test-agent",
				Description:    "A description",
				PermissionMode: "invalid-mode",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "valid inherit model",
			agent: Agent{
				Name:        "test-agent",
				Description: "A description",
				Model:       "inherit",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.agent.Validate("claude-code")
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Agent.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Agent.Validate() expected %d errors, got %d", tt.errorCount, len(errs))
			}
		})
	}
}

func TestAgentValidatePlatformRequirement(t *testing.T) {
	tests := []struct {
		name        string
		agent       Agent
		platform    string
		expectError bool
		errorCount  int
	}{
		{
			name:        "empty platform parameter",
			agent:       Agent{Name: "test-agent", Description: "A description"},
			platform:    "",
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "valid claude-code platform",
			agent:       Agent{Name: "test-agent", Description: "A description"},
			platform:    "claude-code",
			expectError: false,
		},
		{
			name:        "valid opencode platform",
			agent:       Agent{Name: "test-agent", Description: "A description"},
			platform:    "opencode",
			expectError: false,
		},
		{
			name:        "unknown platform",
			agent:       Agent{Name: "test-agent", Description: "A description"},
			platform:    "invalid-platform",
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.agent.Validate(tt.platform)
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Agent.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Agent.Validate() expected %d errors, got %d: %v", tt.errorCount, len(errs), errs)
			}
		})
	}
}

func TestCommandValidate(t *testing.T) {
	tests := []struct {
		name        string
		command     Command
		expectError bool
		errorCount  int
	}{
		{
			name: "valid empty command",
			command: Command{
				Name: "test",
			},
			expectError: false,
		},
		{
			name: "valid command with all fields",
			command: Command{
				Name:                   "test",
				AllowedTools:           []string{"Bash", "Read"},
				ArgumentHint:           "[file-path]",
				Context:                "fork",
				Agent:                  "Explore",
				Description:            "Test command",
				Model:                  "claude-sonnet-4",
				DisableModelInvocation: true,
			},
			expectError: false,
		},
		{
			name: "invalid context value",
			command: Command{
				Name:    "test",
				Context: "invalid",
			},
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.command.Validate("claude-code")
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Command.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Command.Validate() expected %d errors, got %d", tt.errorCount, len(errs))
			}
		})
	}
}

func TestCommandValidatePlatformRequirement(t *testing.T) {
	tests := []struct {
		name        string
		command     Command
		platform    string
		expectError bool
		errorCount  int
	}{
		{
			name:        "empty platform parameter",
			command:     Command{Name: "test-command"},
			platform:    "",
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "valid claude-code platform",
			command:     Command{Name: "test-command"},
			platform:    "claude-code",
			expectError: false,
		},
		{
			name:        "valid opencode platform",
			command:     Command{Name: "test-command", Description: "Test command", Content: "echo test"},
			platform:    "opencode",
			expectError: false,
		},
		{
			name:        "unknown platform",
			command:     Command{Name: "test-command"},
			platform:    "invalid-platform",
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.command.Validate(tt.platform)
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Command.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Command.Validate() expected %d errors, got %d: %v", tt.errorCount, len(errs), errs)
			}
		})
	}
}

func TestMemoryValidate(t *testing.T) {
	tests := []struct {
		name        string
		memory      Memory
		expectError bool
		errorCount  int
	}{
		{
			name:        "invalid memory without paths or content",
			memory:      Memory{},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "valid memory with paths",
			memory: Memory{
				Paths: []string{"**/*.go", "src/**/*.ts"},
			},
			expectError: false,
		},
		{
			name: "valid memory with content",
			memory: Memory{
				Content: "# Project Instructions\n\nThis is a memory file.",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.memory.Validate("claude-code")
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Memory.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Memory.Validate() expected %d errors, got %d", tt.errorCount, len(errs))
			}
		})
	}
}

func TestMemoryValidatePlatformRequirement(t *testing.T) {
	tests := []struct {
		name        string
		memory      Memory
		platform    string
		expectError bool
		errorCount  int
	}{
		{
			name:        "empty platform parameter",
			memory:      Memory{Paths: []string{"*.go"}},
			platform:    "",
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "valid claude-code platform",
			memory:      Memory{Paths: []string{"*.go"}},
			platform:    "claude-code",
			expectError: false,
		},
		{
			name:        "valid opencode platform",
			memory:      Memory{Paths: []string{"*.go"}},
			platform:    "opencode",
			expectError: false,
		},
		{
			name:        "unknown platform",
			memory:      Memory{Paths: []string{"*.go"}},
			platform:    "invalid-platform",
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.memory.Validate(tt.platform)
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Memory.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Memory.Validate() expected %d errors, got %d: %v", tt.errorCount, len(errs), errs)
			}
		})
	}
}

func TestSkillValidate(t *testing.T) {
	tests := []struct {
		name        string
		skill       Skill
		expectError bool
		errorCount  int
	}{
		{
			name: "valid skill with required fields",
			skill: Skill{
				Name:        "explaining-code",
				Description: "Explains code with visual diagrams and analogies",
			},
			expectError: false,
		},
		{
			name: "valid skill with all optional fields",
			skill: Skill{
				Name:          "reading-files-safely",
				Description:   "Read files without making changes",
				AllowedTools:  []string{"Read", "Grep", "Glob"},
				Model:         "claude-sonnet-4",
				Context:       "fork",
				Agent:         "Explore",
				UserInvocable: true,
			},
			expectError: false,
		},
		{
			name:        "missing name",
			skill:       Skill{Description: "A description"},
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "missing description",
			skill:       Skill{Name: "test-skill"},
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "missing both required fields",
			skill:       Skill{},
			expectError: true,
			errorCount:  2,
		},
		{
			name: "invalid name with uppercase",
			skill: Skill{
				Name:        "Code-Explainer",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "name valid with hyphens",
			skill: Skill{
				Name:        "skill-with-multiple-hyphens-segments",
				Description: "A description",
			},
			expectError: false,
		},
		{
			name: "name exceeds 64 characters",
			skill: Skill{
				Name:        "this-is-a-very-long-skill-name-that-exceeds-sixty-four-character-limit-which-is-the-maximum-allowed",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "description exceeds 1024 characters",
			skill: Skill{
				Name:        "test-skill",
				Description: "This description is intentionally very long to exceed 1024 character limit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. This description is intentionally very long to exceed 1024 character limit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. This description is intentionally very long to exceed 1024 character limit.",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid context value",
			skill: Skill{
				Name:        "test-skill",
				Description: "A description",
				Context:     "invalid",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "valid user-invocable false",
			skill: Skill{
				Name:          "internal-skill",
				Description:   "A description",
				UserInvocable: false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.skill.Validate("claude-code")
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Skill.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Skill.Validate() expected %d errors, got %d", tt.errorCount, len(errs))
			}
		})
	}
}

func TestSkillValidatePlatformRequirement(t *testing.T) {
	tests := []struct {
		name        string
		skill       Skill
		platform    string
		expectError bool
		errorCount  int
	}{
		{
			name:        "empty platform parameter",
			skill:       Skill{Name: "test-skill", Description: "A description"},
			platform:    "",
			expectError: true,
			errorCount:  1,
		},
		{
			name:        "valid claude-code platform",
			skill:       Skill{Name: "test-skill", Description: "A description"},
			platform:    "claude-code",
			expectError: false,
		},
		{
			name:        "valid opencode platform",
			skill:       Skill{Name: "test-skill", Description: "A description", Content: "Skill content here"},
			platform:    "opencode",
			expectError: false,
		},
		{
			name:        "unknown platform",
			skill:       Skill{Name: "test-skill", Description: "A description"},
			platform:    "invalid-platform",
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.skill.Validate(tt.platform)
			hasError := len(errs) > 0

			if hasError != tt.expectError {
				t.Errorf("Skill.Validate() error = %v, expectError %v", errs, tt.expectError)
			}

			if tt.errorCount > 0 && len(errs) != tt.errorCount {
				t.Errorf("Skill.Validate() expected %d errors, got %d: %v", tt.errorCount, len(errs), errs)
			}
		})
	}
}

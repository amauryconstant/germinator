package models

import (
	"strings"
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
			name: "invalid name with numbers",
			agent: Agent{
				Name:        "123-agent",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid model value",
			agent: Agent{
				Name:        "test-agent",
				Description: "A description",
				Model:       "invalid-model",
			},
			expectError: true,
			errorCount:  1,
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
			errs := tt.agent.Validate()
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
			errs := tt.command.Validate()
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

func TestMemoryValidate(t *testing.T) {
	tests := []struct {
		name        string
		memory      Memory
		expectError bool
		errorCount  int
	}{
		{
			name:        "valid memory without paths",
			memory:      Memory{},
			expectError: false,
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
			errs := tt.memory.Validate()
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
			name: "name exceeds 64 characters",
			skill: Skill{
				Name:        "a-very-long-skill-name-that-exceeds-the-maximum-length-of-sixty-four-characters",
				Description: "A description",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "description exceeds 1024 characters",
			skill: Skill{
				Name:        "test-skill",
				Description: strings.Repeat("a", 1025),
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
			errs := tt.skill.Validate()
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

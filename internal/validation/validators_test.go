package validation

import (
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models/canonical"
)

func TestValidateAgentName(t *testing.T) {
	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
		errorField  string
	}{
		{
			name:        "empty name fails",
			agent:       &canonical.Agent{Name: ""},
			expectError: true,
			errorField:  "name",
		},
		{
			name:        "invalid name with spaces fails",
			agent:       &canonical.Agent{Name: "Invalid Name"},
			expectError: true,
			errorField:  "name",
		},
		{
			name:        "invalid name with uppercase fails",
			agent:       &canonical.Agent{Name: "InvalidName"},
			expectError: true,
			errorField:  "name",
		},
		{
			name:        "valid single word name passes",
			agent:       &canonical.Agent{Name: "valid"},
			expectError: false,
		},
		{
			name:        "valid hyphenated name passes",
			agent:       &canonical.Agent{Name: "valid-name"},
			expectError: false,
		},
		{
			name:        "valid complex hyphenated name passes",
			agent:       &canonical.Agent{Name: "valid-complex-name-123"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentName(tt.agent)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
				if result.Error == nil {
					t.Fatal("expected error but got nil")
				}
				// Check that error message contains field name
				if !strings.Contains(result.Error.Error(), tt.errorField) {
					t.Errorf("error message should contain field %q, got: %s", tt.errorField, result.Error.Error())
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidateAgentDescription(t *testing.T) {
	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
	}{
		{
			name:        "empty description fails",
			agent:       &canonical.Agent{Description: ""},
			expectError: true,
		},
		{
			name:        "valid description passes",
			agent:       &canonical.Agent{Description: "A valid description"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentDescription(tt.agent)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidateAgentPermissionPolicy(t *testing.T) {
	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
	}{
		{
			name:        "empty policy passes",
			agent:       &canonical.Agent{PermissionPolicy: ""},
			expectError: false,
		},
		{
			name:        "valid restrictive policy passes",
			agent:       &canonical.Agent{PermissionPolicy: canonical.PermissionPolicyRestrictive},
			expectError: false,
		},
		{
			name:        "valid balanced policy passes",
			agent:       &canonical.Agent{PermissionPolicy: canonical.PermissionPolicyBalanced},
			expectError: false,
		},
		{
			name:        "valid permissive policy passes",
			agent:       &canonical.Agent{PermissionPolicy: canonical.PermissionPolicyPermissive},
			expectError: false,
		},
		{
			name:        "invalid policy fails",
			agent:       &canonical.Agent{PermissionPolicy: "invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentPermissionPolicy(tt.agent)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
	}{
		{
			name: "valid agent passes",
			agent: &canonical.Agent{
				Name:             "valid-agent",
				Description:      "A valid description",
				PermissionPolicy: canonical.PermissionPolicyBalanced,
			},
			expectError: false,
		},
		{
			name: "missing name fails",
			agent: &canonical.Agent{
				Description: "A valid description",
			},
			expectError: true,
		},
		{
			name: "missing description fails",
			agent: &canonical.Agent{
				Name: "valid-agent",
			},
			expectError: true,
		},
		{
			name: "invalid policy fails",
			agent: &canonical.Agent{
				Name:             "valid-agent",
				Description:      "A valid description",
				PermissionPolicy: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgent(tt.agent)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     *canonical.Command
		expectError bool
	}{
		{
			name: "valid command passes",
			command: &canonical.Command{
				Name:        "valid-command",
				Description: "A valid description",
			},
			expectError: false,
		},
		{
			name: "missing name fails",
			command: &canonical.Command{
				Description: "A valid description",
			},
			expectError: true,
		},
		{
			name: "missing description fails",
			command: &canonical.Command{
				Name: "valid-command",
			},
			expectError: true,
		},
		{
			name: "invalid context fails",
			command: &canonical.Command{
				Name:        "valid-command",
				Description: "A valid description",
				Execution:   canonical.CommandExecution{Context: "invalid"},
			},
			expectError: true,
		},
		{
			name: "valid context passes",
			command: &canonical.Command{
				Name:        "valid-command",
				Description: "A valid description",
				Execution:   canonical.CommandExecution{Context: "fork"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCommand(tt.command)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidateSkill(t *testing.T) {
	tests := []struct {
		name        string
		skill       *canonical.Skill
		expectError bool
	}{
		{
			name: "valid skill passes",
			skill: &canonical.Skill{
				Name:        "valid-skill",
				Description: "A valid description",
			},
			expectError: false,
		},
		{
			name: "missing name fails",
			skill: &canonical.Skill{
				Description: "A valid description",
			},
			expectError: true,
		},
		{
			name: "name too long fails",
			skill: &canonical.Skill{
				Name:        strings.Repeat("a", 65),
				Description: "A valid description",
			},
			expectError: true,
		},
		{
			name: "description too long fails",
			skill: &canonical.Skill{
				Name:        "valid-skill",
				Description: strings.Repeat("a", 1025),
			},
			expectError: true,
		},
		{
			name: "invalid name format fails",
			skill: &canonical.Skill{
				Name:        "Invalid Name",
				Description: "A valid description",
			},
			expectError: true,
		},
		{
			name: "invalid context fails",
			skill: &canonical.Skill{
				Name:        "valid-skill",
				Description: "A valid description",
				Execution:   canonical.SkillExecution{Context: "invalid"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSkill(tt.skill)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidateMemory(t *testing.T) {
	tests := []struct {
		name        string
		memory      *canonical.Memory
		expectError bool
	}{
		{
			name: "memory with paths passes",
			memory: &canonical.Memory{
				Paths: []string{"/path/to/file"},
			},
			expectError: false,
		},
		{
			name: "memory with content passes",
			memory: &canonical.Memory{
				Content: "Some content",
			},
			expectError: false,
		},
		{
			name: "memory with both passes",
			memory: &canonical.Memory{
				Paths:   []string{"/path/to/file"},
				Content: "Some content",
			},
			expectError: false,
		},
		{
			name:        "memory with neither fails",
			memory:      &canonical.Memory{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMemory(tt.memory)
			if tt.expectError {
				if result.IsSuccess() {
					t.Error("expected error but got success")
				}
			} else {
				if result.IsError() {
					t.Errorf("expected success but got error: %v", result.Error)
				}
			}
		})
	}
}

func TestValidatePipelineComposition(t *testing.T) {
	t.Run("agent pipeline stops on first error", func(t *testing.T) {
		agent := &canonical.Agent{
			// Name is empty - should fail on first validator
			Description: "A valid description",
		}

		result := ValidateAgent(agent)
		if result.IsSuccess() {
			t.Error("expected error for missing name")
		}
		// Should not reach description validation
	})

	t.Run("command pipeline validates all fields", func(t *testing.T) {
		command := &canonical.Command{
			Name:        "valid-command",
			Description: "A valid description",
			Execution:   canonical.CommandExecution{Context: "fork"},
		}

		result := ValidateCommand(command)
		if result.IsError() {
			t.Errorf("expected success but got error: %v", result.Error)
		}
	})

	t.Run("skill pipeline validates all fields", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "valid-skill",
			Description: "A valid description",
			Execution:   canonical.SkillExecution{Context: "fork"},
		}

		result := ValidateSkill(skill)
		if result.IsError() {
			t.Errorf("expected success but got error: %v", result.Error)
		}
	})
}

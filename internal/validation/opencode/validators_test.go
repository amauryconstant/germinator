package opencode

import (
	"testing"

	"gitlab.com/amoconst/germinator/internal/models/canonical"
)

func TestValidateAgentMode(t *testing.T) {
	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
	}{
		{
			name:        "empty mode passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Mode: ""}},
			expectError: false,
		},
		{
			name:        "primary mode passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Mode: "primary"}},
			expectError: false,
		},
		{
			name:        "subagent mode passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Mode: "subagent"}},
			expectError: false,
		},
		{
			name:        "all mode passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Mode: "all"}},
			expectError: false,
		},
		{
			name:        "invalid mode fails",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Mode: "invalid"}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentMode(tt.agent)
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

func TestValidateAgentTemperature(t *testing.T) {
	minTemp := 0.0
	maxTemp := 1.0
	midTemp := 0.5
	invalidLow := -0.1
	invalidHigh := 1.1

	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
	}{
		{
			name:        "nil temperature passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Temperature: nil}},
			expectError: false,
		},
		{
			name:        "minimum temperature passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Temperature: &minTemp}},
			expectError: false,
		},
		{
			name:        "maximum temperature passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Temperature: &maxTemp}},
			expectError: false,
		},
		{
			name:        "mid temperature passes",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Temperature: &midTemp}},
			expectError: false,
		},
		{
			name:        "temperature below range fails",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Temperature: &invalidLow}},
			expectError: true,
		},
		{
			name:        "temperature above range fails",
			agent:       &canonical.Agent{Behavior: canonical.AgentBehavior{Temperature: &invalidHigh}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentTemperature(tt.agent)
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

func TestValidateAgentOpenCode(t *testing.T) {
	validMode := "primary"
	validTemp := 0.5
	invalidMode := "invalid"
	invalidTemp := 1.5

	tests := []struct {
		name        string
		agent       *canonical.Agent
		expectError bool
	}{
		{
			name: "valid agent passes",
			agent: &canonical.Agent{
				Behavior: canonical.AgentBehavior{
					Mode:        validMode,
					Temperature: &validTemp,
				},
			},
			expectError: false,
		},
		{
			name: "empty mode and nil temperature passes",
			agent: &canonical.Agent{
				Behavior: canonical.AgentBehavior{
					Mode:        "",
					Temperature: nil,
				},
			},
			expectError: false,
		},
		{
			name: "invalid mode fails",
			agent: &canonical.Agent{
				Behavior: canonical.AgentBehavior{
					Mode: invalidMode,
				},
			},
			expectError: true,
		},
		{
			name: "invalid temperature fails",
			agent: &canonical.Agent{
				Behavior: canonical.AgentBehavior{
					Temperature: &invalidTemp,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentOpenCode(tt.agent)
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

func TestValidateCommandOpenCode(t *testing.T) {
	command := &canonical.Command{
		Name:        "test-command",
		Description: "Test description",
	}

	result := ValidateCommandOpenCode(command)
	if result.IsError() {
		t.Errorf("expected success but got error: %v", result.Error)
	}
}

func TestValidateSkillOpenCode(t *testing.T) {
	skill := &canonical.Skill{
		Name:        "test-skill",
		Description: "Test description",
	}

	result := ValidateSkillOpenCode(skill)
	if result.IsError() {
		t.Errorf("expected success but got error: %v", result.Error)
	}
}

func TestOpenCodePipelineComposition(t *testing.T) {
	t.Run("agent pipeline stops on first error", func(t *testing.T) {
		invalidMode := "invalid"
		invalidTemp := 1.5

		agent := &canonical.Agent{
			Behavior: canonical.AgentBehavior{
				Mode:        invalidMode, // Fails first
				Temperature: &invalidTemp,
			},
		}

		result := ValidateAgentOpenCode(agent)
		if result.IsSuccess() {
			t.Error("expected error for invalid mode")
		}
	})
}

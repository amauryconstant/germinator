// Package opencode provides OpenCode-specific validation functions.
package opencode

import (
	"fmt"

	"gitlab.com/amoconst/germinator/internal/domain"
)

// ValidateAgentMode validates OpenCode-specific agent mode values.
func ValidateAgentMode(a *domain.Agent) domain.Result[bool] {
	if a.Behavior.Mode == "" {
		// Mode is optional, empty is valid
		return domain.NewResult(true)
	}

	validModes := map[string]bool{
		"primary":  true,
		"subagent": true,
		"all":      true,
	}

	if !validModes[a.Behavior.Mode] {
		return domain.NewErrorResult[bool](
			domain.NewValidationError(
				"Agent",
				"behavior.mode",
				a.Behavior.Mode,
				"behavior.mode must be one of: primary, subagent, all",
			),
		)
	}

	return domain.NewResult(true)
}

// ValidateAgentTemperature validates OpenCode-specific agent temperature range.
func ValidateAgentTemperature(a *domain.Agent) domain.Result[bool] {
	if a.Behavior.Temperature == nil {
		// Temperature is optional, nil is valid
		return domain.NewResult(true)
	}

	temp := *a.Behavior.Temperature
	if temp < 0.0 || temp > 1.0 {
		return domain.NewErrorResult[bool](
			domain.NewValidationError(
				"Agent",
				"behavior.temperature",
				fmt.Sprintf("%f", temp),
				"behavior.temperature must be between 0.0 and 1.0",
			),
		)
	}

	return domain.NewResult(true)
}

// ValidateAgentOpenCode composes all OpenCode-specific agent validators.
func ValidateAgentOpenCode(a *domain.Agent) domain.Result[bool] {
	return domain.NewValidationPipeline(
		ValidateAgentMode,
		ValidateAgentTemperature,
	).Validate(a)
}

// ValidateCommandOpenCode validates OpenCode-specific command constraints.
// Currently no OpenCode-specific validation for commands beyond generic rules.
func ValidateCommandOpenCode(c *domain.Command) domain.Result[bool] {
	return domain.NewResult(true)
}

// ValidateSkillOpenCode validates OpenCode-specific skill constraints.
// Currently no OpenCode-specific validation for skills beyond generic rules.
func ValidateSkillOpenCode(s *domain.Skill) domain.Result[bool] {
	return domain.NewResult(true)
}

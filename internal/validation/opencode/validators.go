// Package opencode provides OpenCode-specific validation functions.
package opencode

import (
	"fmt"

	domainerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models/canonical"
	"gitlab.com/amoconst/germinator/internal/validation"
)

// ValidateAgentMode validates OpenCode-specific agent mode values.
func ValidateAgentMode(a *canonical.Agent) validation.Result[bool] {
	if a.Behavior.Mode == "" {
		// Mode is optional, empty is valid
		return validation.NewResult(true)
	}

	validModes := map[string]bool{
		"primary":  true,
		"subagent": true,
		"all":      true,
	}

	if !validModes[a.Behavior.Mode] {
		return validation.NewErrorResult[bool](
			domainerrors.NewValidationError(
				"Agent",
				"behavior.mode",
				a.Behavior.Mode,
				"behavior.mode must be one of: primary, subagent, all",
			),
		)
	}

	return validation.NewResult(true)
}

// ValidateAgentTemperature validates OpenCode-specific agent temperature range.
func ValidateAgentTemperature(a *canonical.Agent) validation.Result[bool] {
	if a.Behavior.Temperature == nil {
		// Temperature is optional, nil is valid
		return validation.NewResult(true)
	}

	temp := *a.Behavior.Temperature
	if temp < 0.0 || temp > 1.0 {
		return validation.NewErrorResult[bool](
			domainerrors.NewValidationError(
				"Agent",
				"behavior.temperature",
				fmt.Sprintf("%f", temp),
				"behavior.temperature must be between 0.0 and 1.0",
			),
		)
	}

	return validation.NewResult(true)
}

// ValidateAgentOpenCode composes all OpenCode-specific agent validators.
func ValidateAgentOpenCode(a *canonical.Agent) validation.Result[bool] {
	return validation.NewValidationPipeline(
		ValidateAgentMode,
		ValidateAgentTemperature,
	).Validate(a)
}

// ValidateCommandOpenCode validates OpenCode-specific command constraints.
// Currently no OpenCode-specific validation for commands beyond generic rules.
func ValidateCommandOpenCode(c *canonical.Command) validation.Result[bool] {
	return validation.NewResult(true)
}

// ValidateSkillOpenCode validates OpenCode-specific skill constraints.
// Currently no OpenCode-specific validation for skills beyond generic rules.
func ValidateSkillOpenCode(s *canonical.Skill) validation.Result[bool] {
	return validation.NewResult(true)
}

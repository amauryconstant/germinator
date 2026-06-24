// Package opencode provides OpenCode-specific validation functions.
package opencode

import (
	"fmt"

	"gitlab.com/amoconst/germinator/internal/core"
)

// ValidateAgentMode validates OpenCode-specific agent mode values.
func ValidateAgentMode(a *core.Agent) core.Result[bool] {
	if a.Behavior.Mode == "" {
		// Mode is optional, empty is valid
		return core.NewResult(true)
	}

	validModes := map[string]bool{
		"primary":  true,
		"subagent": true,
		"all":      true,
	}

	if !validModes[a.Behavior.Mode] {
		return core.NewErrorResult[bool](
			core.NewValidationError(
				"Agent",
				"behavior.mode",
				a.Behavior.Mode,
				"behavior.mode must be one of: primary, subagent, all",
			),
		)
	}

	return core.NewResult(true)
}

// ValidateAgentTemperature validates OpenCode-specific agent temperature range.
func ValidateAgentTemperature(a *core.Agent) core.Result[bool] {
	if a.Behavior.Temperature == nil {
		// Temperature is optional, nil is valid
		return core.NewResult(true)
	}

	temp := *a.Behavior.Temperature
	if temp < 0.0 || temp > 1.0 {
		return core.NewErrorResult[bool](
			core.NewValidationError(
				"Agent",
				"behavior.temperature",
				fmt.Sprintf("%f", temp),
				"behavior.temperature must be between 0.0 and 1.0",
			),
		)
	}

	return core.NewResult(true)
}

// ValidateAgentOpenCode composes all OpenCode-specific agent validators.
func ValidateAgentOpenCode(a *core.Agent) core.Result[bool] {
	return core.NewValidationPipeline(
		ValidateAgentMode,
		ValidateAgentTemperature,
	).Validate(a)
}

// ValidateCommandOpenCode validates OpenCode-specific command constraints.
// Currently no OpenCode-specific validation for commands beyond generic rules.
func ValidateCommandOpenCode(_ *core.Command) core.Result[bool] {
	return core.NewResult(true)
}

// ValidateSkillOpenCode validates OpenCode-specific skill constraints.
// Currently no OpenCode-specific validation for skills beyond generic rules.
func ValidateSkillOpenCode(_ *core.Skill) core.Result[bool] {
	return core.NewResult(true)
}

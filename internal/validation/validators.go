package validation

import (
	"fmt"
	"regexp"

	domainerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models/canonical"
)

// Agent validators

// ValidateAgentName validates that agent name is required and matches the expected pattern.
func ValidateAgentName(a *canonical.Agent) Result[bool] {
	if a.Name == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Agent", "name", a.Name, "name is required"),
		)
	}

	matched, err := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, a.Name)
	if err != nil {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Agent", "name", a.Name, fmt.Sprintf("failed to validate name regex: %v", err)),
		)
	}
	if !matched {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Agent", "name", a.Name, "name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$"),
		)
	}

	return NewResult(true)
}

// ValidateAgentDescription validates that agent description is required.
func ValidateAgentDescription(a *canonical.Agent) Result[bool] {
	if a.Description == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Agent", "description", "", "description is required"),
		)
	}
	return NewResult(true)
}

// ValidateAgentPermissionPolicy validates that permission policy is valid if specified.
func ValidateAgentPermissionPolicy(a *canonical.Agent) Result[bool] {
	if a.PermissionPolicy != "" && !a.PermissionPolicy.IsValid() {
		return NewErrorResult[bool](
			domainerrors.NewValidationError(
				"Agent",
				"permissionPolicy",
				string(a.PermissionPolicy),
				"permissionPolicy must be one of: restrictive, balanced, permissive, analysis, unrestricted",
			),
		)
	}
	return NewResult(true)
}

// ValidateAgent composes all agent validators into a pipeline.
func ValidateAgent(a *canonical.Agent) Result[bool] {
	return NewValidationPipeline(
		ValidateAgentName,
		ValidateAgentDescription,
		ValidateAgentPermissionPolicy,
	).Validate(a)
}

// Command validators

// ValidateCommandName validates that command name is required.
func ValidateCommandName(c *canonical.Command) Result[bool] {
	if c.Name == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Command", "name", "", "name is required"),
		)
	}
	return NewResult(true)
}

// ValidateCommandDescription validates that command description is required.
func ValidateCommandDescription(c *canonical.Command) Result[bool] {
	if c.Description == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Command", "description", "", "description is required"),
		)
	}
	return NewResult(true)
}

// ValidateCommandExecution validates command execution context.
func ValidateCommandExecution(c *canonical.Command) Result[bool] {
	if c.Execution.Context != "" && c.Execution.Context != "fork" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError(
				"Command",
				"execution.context",
				c.Execution.Context,
				"execution.context must be 'fork' if specified",
			),
		)
	}
	return NewResult(true)
}

// ValidateCommand composes all command validators into a pipeline.
func ValidateCommand(c *canonical.Command) Result[bool] {
	return NewValidationPipeline(
		ValidateCommandName,
		ValidateCommandDescription,
		ValidateCommandExecution,
	).Validate(c)
}

// Skill validators

// ValidateSkillName validates that skill name is required and matches the expected pattern.
func ValidateSkillName(s *canonical.Skill) Result[bool] {
	if s.Name == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Skill", "name", "", "name is required"),
		)
	}

	if len(s.Name) < 1 || len(s.Name) > 64 {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Skill", "name", s.Name, fmt.Sprintf("name must be 1-64 characters (got: %d)", len(s.Name))),
		)
	}

	matched, err := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, s.Name)
	if err != nil {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Skill", "name", s.Name, fmt.Sprintf("failed to validate name regex: %v", err)),
		)
	}
	if !matched {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Skill", "name", s.Name, "name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$"),
		)
	}

	return NewResult(true)
}

// ValidateSkillDescription validates that skill description is required and within length limits.
func ValidateSkillDescription(s *canonical.Skill) Result[bool] {
	if s.Description == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Skill", "description", "", "description is required"),
		)
	}

	if len(s.Description) < 1 || len(s.Description) > 1024 {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Skill", "description", s.Description, fmt.Sprintf("description must be 1-1024 characters (got: %d)", len(s.Description))),
		)
	}

	return NewResult(true)
}

// ValidateSkillExecution validates skill execution context.
func ValidateSkillExecution(s *canonical.Skill) Result[bool] {
	if s.Execution.Context != "" && s.Execution.Context != "fork" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError(
				"Skill",
				"execution.context",
				s.Execution.Context,
				"execution.context must be 'fork' if specified",
			),
		)
	}
	return NewResult(true)
}

// ValidateSkill composes all skill validators into a pipeline.
func ValidateSkill(s *canonical.Skill) Result[bool] {
	return NewValidationPipeline(
		ValidateSkillName,
		ValidateSkillDescription,
		ValidateSkillExecution,
	).Validate(s)
}

// Memory validators

// ValidateMemory validates that memory has either paths or content specified.
func ValidateMemory(m *canonical.Memory) Result[bool] {
	if len(m.Paths) == 0 && m.Content == "" {
		return NewErrorResult[bool](
			domainerrors.NewValidationError("Memory", "paths/content", "", "paths or content is required"),
		)
	}
	return NewResult(true)
}

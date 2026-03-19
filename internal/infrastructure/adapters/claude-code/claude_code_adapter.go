// Package claudecode provides an adapter implementation for transforming between
// the Claude Code platform format and canonical domain models.
//
// The adapter handles parsing Claude Code YAML documents (agents, commands, skills, memory)
// into canonical domain structs, and rendering canonical structs back to Claude Code format
// for template rendering.
package claudecode

import (
	"fmt"

	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/adapters"
)

// Adapter implements the Adapter interface for Claude Code platform.
type Adapter struct{}

// New creates and returns a new Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// ToCanonical parses Claude Code YAML format into canonical domain models.
// It identifies the document type from the __type field and returns the appropriate
// canonical struct (Agent, Command, Skill, or Memory) along with nil values for the other types.
func (a *Adapter) ToCanonical(input map[string]interface{}) (*domain.Agent, *domain.Command, *domain.Skill, *domain.Memory, error) {
	docType, ok := input["__type"].(string)
	if !ok {
		return nil, nil, nil, nil, domain.NewParseError("", "missing __type field", nil)
	}

	switch docType {
	case "agent":
		agent, err := a.parseAgent(input)
		return agent, nil, nil, nil, err
	case "command":
		cmd, err := a.parseCommand(input)
		return nil, cmd, nil, nil, err
	case "skill":
		skill, err := a.parseSkill(input)
		return nil, nil, skill, nil, err
	case "memory":
		mem, err := a.parseMemory(input)
		return nil, nil, nil, mem, err
	default:
		return nil, nil, nil, nil, domain.NewParseError("", "unknown document type: "+docType, nil)
	}
}

// FromCanonical converts canonical domain models to Claude Code format maps.
// It accepts a document type string and the corresponding canonical struct,
// returning a map suitable for template rendering.
func (a *Adapter) FromCanonical(docType string, doc interface{}) (map[string]interface{}, error) {
	switch docType {
	case "agent":
		agent, ok := doc.(*domain.Agent)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "claude-code", fmt.Sprintf("expected *domain.Agent, got %T", doc), nil)
		}
		return a.renderAgent(agent)
	case "command":
		cmd, ok := doc.(*domain.Command)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "claude-code", fmt.Sprintf("expected *domain.Command, got %T", doc), nil)
		}
		return a.renderCommand(cmd)
	case "skill":
		skill, ok := doc.(*domain.Skill)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "claude-code", fmt.Sprintf("expected *domain.Skill, got %T", doc), nil)
		}
		return a.renderSkill(skill)
	case "memory":
		mem, ok := doc.(*domain.Memory)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "claude-code", fmt.Sprintf("expected *domain.Memory, got %T", doc), nil)
		}
		return a.renderMemory(mem)
	default:
		return nil, domain.NewTransformError("from-canonical", "claude-code", "unknown document type: "+docType, nil)
	}
}

// PermissionPolicyToPlatform converts a canonical permission policy to Claude Code enum format.
// Returns the corresponding Claude Code permission mode string (e.g., "default", "acceptEdits", "dontAsk").
func (a *Adapter) PermissionPolicyToPlatform(policy domain.PermissionPolicy) (interface{}, error) {
	mapping, ok := adapters.PermissionPolicyMappings[string(policy)]
	if !ok {
		return nil, domain.NewConfigError("permission-policy", string(policy), "unknown permission policy")
	}
	return mapping.ClaudeCode, nil
}

// ConvertToolNameCase converts a tool name to PascalCase format used by Claude Code.
func (a *Adapter) ConvertToolNameCase(name string) string {
	return adapters.ToPascalCase(name)
}

func (a *Adapter) parseAgent(input map[string]interface{}) (*domain.Agent, error) {
	agent := &domain.Agent{}

	if name, ok := input["name"].(string); ok {
		agent.Name = name
	}
	if description, ok := input["description"].(string); ok {
		agent.Description = description
	}

	if tools, ok := input["tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				agent.Tools = append(agent.Tools, adapters.ToLowerCase(toolName))
			}
		}
	} else if tools, ok := input["tools"].([]string); ok {
		for _, toolName := range tools {
			agent.Tools = append(agent.Tools, adapters.ToLowerCase(toolName))
		}
	}

	if disallowedTools, ok := input["disallowedTools"].([]interface{}); ok {
		for _, t := range disallowedTools {
			if toolName, ok := t.(string); ok {
				agent.DisallowedTools = append(agent.DisallowedTools, adapters.ToLowerCase(toolName))
			}
		}
	} else if disallowedTools, ok := input["disallowedTools"].([]string); ok {
		for _, toolName := range disallowedTools {
			agent.DisallowedTools = append(agent.DisallowedTools, adapters.ToLowerCase(toolName))
		}
	}

	if permissionMode, ok := input["permissionMode"].(string); ok {
		agent.PermissionPolicy = a.mapPermissionModeToPolicy(permissionMode)
	}

	agent.Behavior = domain.AgentBehavior{}
	if mode, ok := input["mode"].(string); ok {
		agent.Behavior.Mode = mode
	}
	if temperature, ok := input["temperature"].(float64); ok {
		agent.Behavior.Temperature = &temperature
	}
	if maxSteps, ok := input["maxSteps"].(int); ok {
		agent.Behavior.Steps = maxSteps
	}
	if prompt, ok := input["prompt"].(string); ok {
		agent.Behavior.Prompt = prompt
	}
	if hidden, ok := input["hidden"].(bool); ok {
		agent.Behavior.Hidden = hidden
	}
	if disabled, ok := input["disabled"].(bool); ok {
		agent.Behavior.Disabled = disabled
	}

	if hooks, ok := input["hooks"].(map[string]interface{}); ok {
		agent.Extensions = domain.AgentExtensions{}
		agent.Extensions.Hooks = make(map[string]string)
		for k, v := range hooks {
			if val, ok := v.(string); ok {
				agent.Extensions.Hooks[k] = val
			}
		}
	}

	if model, ok := input["model"].(string); ok {
		agent.Model = model
	}

	agent.Targets = make(domain.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		agent.Targets = a.parseTargets(targets)
	}

	if skills, ok := input["skills"].([]interface{}); ok {
		if agent.Targets == nil {
			agent.Targets = make(domain.PlatformConfig)
		}
		if agent.Targets["claude-code"] == nil {
			agent.Targets["claude-code"] = make(map[string]interface{})
		}
		skillNames := make([]string, 0, len(skills))
		for _, s := range skills {
			if skillName, ok := s.(string); ok {
				skillNames = append(skillNames, skillName)
			}
		}
		agent.Targets["claude-code"]["skills"] = skillNames
	}

	return agent, nil
}

func (a *Adapter) renderAgent(agent *domain.Agent) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "agent"
	output["name"] = agent.Name
	output["description"] = agent.Description

	if len(agent.Tools) > 0 {
		tools := make([]string, len(agent.Tools))
		for i, t := range agent.Tools {
			tools[i] = a.ConvertToolNameCase(t)
		}
		output["tools"] = tools
	}

	if len(agent.DisallowedTools) > 0 {
		disallowedTools := make([]string, len(agent.DisallowedTools))
		for i, t := range agent.DisallowedTools {
			disallowedTools[i] = a.ConvertToolNameCase(t)
		}
		output["disallowedTools"] = disallowedTools
	}

	if agent.PermissionPolicy != "" {
		permissionMode, err := a.PermissionPolicyToPlatform(agent.PermissionPolicy)
		if err != nil {
			return nil, err
		}
		output["permissionMode"] = permissionMode
	}

	if agent.Behavior.Mode != "" {
		output["mode"] = agent.Behavior.Mode
	}
	if agent.Behavior.Temperature != nil {
		output["temperature"] = *agent.Behavior.Temperature
	}
	if agent.Behavior.Steps != 0 {
		output["maxSteps"] = agent.Behavior.Steps
	}
	if agent.Behavior.Prompt != "" {
		output["prompt"] = agent.Behavior.Prompt
	}
	if agent.Behavior.Hidden {
		output["hidden"] = agent.Behavior.Hidden
	}
	if agent.Behavior.Disabled {
		output["disabled"] = agent.Behavior.Disabled
	}

	if len(agent.Extensions.Hooks) > 0 {
		output["hooks"] = agent.Extensions.Hooks
	}

	if agent.Model != "" {
		output["model"] = agent.Model
	}

	if ccTargets, ok := agent.Targets["claude-code"]; ok {
		if skills, ok := ccTargets["skills"].([]string); ok {
			output["skills"] = skills
		}
		if disableModelInvocation, ok := ccTargets["disable-model-invocation"].(bool); ok {
			output["disable-model-invocation"] = disableModelInvocation
		}
	}

	return output, nil
}

func (a *Adapter) parseCommand(input map[string]interface{}) (*domain.Command, error) {
	cmd := &domain.Command{}

	if name, ok := input["name"].(string); ok {
		cmd.Name = name
	}
	if description, ok := input["description"].(string); ok {
		cmd.Description = description
	}

	if tools, ok := input["tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				cmd.Tools = append(cmd.Tools, adapters.ToLowerCase(toolName))
			}
		}
	}

	if execution, ok := input["execution"].(map[string]interface{}); ok {
		cmd.Execution = domain.CommandExecution{}
		if context, ok := execution["context"].(string); ok {
			cmd.Execution.Context = context
		}
		if subtask, ok := execution["subtask"].(bool); ok {
			cmd.Execution.Subtask = subtask
		}
		if agent, ok := execution["agent"].(string); ok {
			cmd.Execution.Agent = agent
		}
	}

	if arguments, ok := input["arguments"].(map[string]interface{}); ok {
		cmd.Arguments = domain.CommandArguments{}
		if hint, ok := arguments["hint"].(string); ok {
			cmd.Arguments.Hint = hint
		}
	}

	if model, ok := input["model"].(string); ok {
		cmd.Model = model
	}

	cmd.Targets = make(domain.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		cmd.Targets = a.parseTargets(targets)
	}

	return cmd, nil
}

func (a *Adapter) renderCommand(cmd *domain.Command) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "command"
	output["name"] = cmd.Name
	output["description"] = cmd.Description

	if len(cmd.Tools) > 0 {
		tools := make([]string, len(cmd.Tools))
		for i, t := range cmd.Tools {
			tools[i] = a.ConvertToolNameCase(t)
		}
		output["tools"] = tools
	}

	execution := make(map[string]interface{})
	if cmd.Execution.Context != "" {
		execution["context"] = cmd.Execution.Context
	}
	if cmd.Execution.Subtask {
		execution["subtask"] = cmd.Execution.Subtask
	}
	if cmd.Execution.Agent != "" {
		execution["agent"] = cmd.Execution.Agent
	}
	if len(execution) > 0 {
		output["execution"] = execution
	}

	if cmd.Arguments.Hint != "" {
		arguments := make(map[string]interface{})
		arguments["hint"] = cmd.Arguments.Hint
		output["arguments"] = arguments
	}

	if cmd.Model != "" {
		output["model"] = cmd.Model
	}

	return output, nil
}

func (a *Adapter) parseSkill(input map[string]interface{}) (*domain.Skill, error) {
	skill := &domain.Skill{}

	if name, ok := input["name"].(string); ok {
		skill.Name = name
	}
	if description, ok := input["description"].(string); ok {
		skill.Description = description
	}

	if tools, ok := input["tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				skill.Tools = append(skill.Tools, adapters.ToLowerCase(toolName))
			}
		}
	}

	skill.Extensions = domain.SkillExtensions{}
	if extensions, ok := input["extensions"].(map[string]interface{}); ok {
		if license, ok := extensions["license"].(string); ok {
			skill.Extensions.License = license
		}
		if compatibility, ok := extensions["compatibility"].([]interface{}); ok {
			for _, c := range compatibility {
				if comp, ok := c.(string); ok {
					skill.Extensions.Compatibility = append(skill.Extensions.Compatibility, comp)
				}
			}
		}
		if metadata, ok := extensions["metadata"].(map[string]interface{}); ok {
			skill.Extensions.Metadata = make(map[string]string)
			for k, v := range metadata {
				if val, ok := v.(string); ok {
					skill.Extensions.Metadata[k] = val
				}
			}
		}
		if hooks, ok := extensions["hooks"].(map[string]interface{}); ok {
			skill.Extensions.Hooks = make(map[string]string)
			for k, v := range hooks {
				if val, ok := v.(string); ok {
					skill.Extensions.Hooks[k] = val
				}
			}
		}
	}

	skill.Execution = domain.SkillExecution{}
	if execution, ok := input["execution"].(map[string]interface{}); ok {
		if context, ok := execution["context"].(string); ok {
			skill.Execution.Context = context
		}
		if agent, ok := execution["agent"].(string); ok {
			skill.Execution.Agent = agent
		}
		if userInvocable, ok := execution["userInvocable"].(bool); ok {
			skill.Execution.UserInvocable = userInvocable
		}
	}

	if model, ok := input["model"].(string); ok {
		skill.Model = model
	}

	skill.Targets = make(domain.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		skill.Targets = a.parseTargets(targets)
	}

	return skill, nil
}

func (a *Adapter) renderSkill(skill *domain.Skill) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "skill"
	output["name"] = skill.Name
	output["description"] = skill.Description

	if len(skill.Tools) > 0 {
		tools := make([]string, len(skill.Tools))
		for i, t := range skill.Tools {
			tools[i] = a.ConvertToolNameCase(t)
		}
		output["tools"] = tools
	}

	if skill.Extensions.License != "" || len(skill.Extensions.Compatibility) > 0 ||
		len(skill.Extensions.Metadata) > 0 || len(skill.Extensions.Hooks) > 0 {
		extensions := make(map[string]interface{})
		if skill.Extensions.License != "" {
			extensions["license"] = skill.Extensions.License
		}
		if len(skill.Extensions.Compatibility) > 0 {
			extensions["compatibility"] = skill.Extensions.Compatibility
		}
		if len(skill.Extensions.Metadata) > 0 {
			extensions["metadata"] = skill.Extensions.Metadata
		}
		if len(skill.Extensions.Hooks) > 0 {
			extensions["hooks"] = skill.Extensions.Hooks
		}
		output["extensions"] = extensions
	}

	if skill.Execution.Context != "" || skill.Execution.Agent != "" || skill.Execution.UserInvocable {
		execution := make(map[string]interface{})
		if skill.Execution.Context != "" {
			execution["context"] = skill.Execution.Context
		}
		if skill.Execution.Agent != "" {
			execution["agent"] = skill.Execution.Agent
		}
		if skill.Execution.UserInvocable {
			execution["userInvocable"] = skill.Execution.UserInvocable
		}
		output["execution"] = execution
	}

	if skill.Model != "" {
		output["model"] = skill.Model
	}

	return output, nil
}

func (a *Adapter) parseMemory(input map[string]interface{}) (*domain.Memory, error) {
	mem := &domain.Memory{}

	if paths, ok := input["paths"].([]interface{}); ok {
		for _, p := range paths {
			if path, ok := p.(string); ok {
				mem.Paths = append(mem.Paths, path)
			}
		}
	}

	if content, ok := input["content"].(string); ok {
		mem.Content = content
	}

	return mem, nil
}

func (a *Adapter) renderMemory(mem *domain.Memory) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "memory"

	if len(mem.Paths) > 0 {
		output["paths"] = mem.Paths
	}

	if mem.Content != "" {
		output["content"] = mem.Content
	}

	return output, nil
}

func (a *Adapter) parseTargets(input map[string]interface{}) domain.PlatformConfig {
	targets := make(domain.PlatformConfig)
	for platform, config := range input {
		if configMap, ok := config.(map[string]interface{}); ok {
			targets[platform] = configMap
		}
	}
	return targets
}

func (a *Adapter) mapPermissionModeToPolicy(mode string) domain.PermissionPolicy {
	switch mode {
	case "default":
		return domain.PermissionPolicyRestrictive
	case "acceptEdits":
		return domain.PermissionPolicyBalanced
	case "dontAsk":
		return domain.PermissionPolicyPermissive
	case "plan":
		return domain.PermissionPolicyAnalysis
	case "bypassPermissions":
		return domain.PermissionPolicyUnrestricted
	default:
		return ""
	}
}

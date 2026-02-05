package opencode

import (
	"fmt"

	"gitlab.com/amoconst/germinator/internal/adapters"
	"gitlab.com/amoconst/germinator/internal/models/canonical"
)

type OpenCodeAdapter struct{}

func New() *OpenCodeAdapter {
	return &OpenCodeAdapter{}
}

func (a *OpenCodeAdapter) ToCanonical(input map[string]interface{}) (*canonical.Agent, *canonical.Command, *canonical.Skill, *canonical.Memory, error) {
	docType, ok := input["__type"].(string)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("missing __type field")
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
		return nil, nil, nil, nil, fmt.Errorf("unknown document type: %s", docType)
	}
}

func (a *OpenCodeAdapter) FromCanonical(docType string, doc interface{}) (map[string]interface{}, error) {
	switch docType {
	case "agent":
		agent, ok := doc.(*canonical.Agent)
		if !ok {
			return nil, fmt.Errorf("expected *canonical.Agent, got %T", doc)
		}
		return a.renderAgent(agent)
	case "command":
		cmd, ok := doc.(*canonical.Command)
		if !ok {
			return nil, fmt.Errorf("expected *canonical.Command, got %T", doc)
		}
		return a.renderCommand(cmd)
	case "skill":
		skill, ok := doc.(*canonical.Skill)
		if !ok {
			return nil, fmt.Errorf("expected *canonical.Skill, got %T", doc)
		}
		return a.renderSkill(skill)
	case "memory":
		mem, ok := doc.(*canonical.Memory)
		if !ok {
			return nil, fmt.Errorf("expected *canonical.Memory, got %T", doc)
		}
		return a.renderMemory(mem)
	default:
		return nil, fmt.Errorf("unknown document type: %s", docType)
	}
}

func (a *OpenCodeAdapter) PermissionPolicyToPlatform(policy canonical.PermissionPolicy) (interface{}, error) {
	mapping, ok := adapters.PermissionPolicyMappings[string(policy)]
	if !ok {
		return nil, fmt.Errorf("unknown permission policy: %s", policy)
	}
	return mapping.OpenCode, nil
}

func (a *OpenCodeAdapter) ConvertToolNameCase(name string) string {
	return adapters.ToLowerCase(name)
}

func (a *OpenCodeAdapter) parseAgent(input map[string]interface{}) (*canonical.Agent, error) {
	agent := &canonical.Agent{}

	if name, ok := input["name"].(string); ok {
		agent.Name = name
	}
	if description, ok := input["description"].(string); ok {
		agent.Description = description
	}

	if tools, ok := input["tools"].(map[string]interface{}); ok {
		for toolName, allowed := range tools {
			if isAllowed, ok := allowed.(bool); ok {
				if isAllowed {
					agent.Tools = append(agent.Tools, adapters.ToLowerCase(toolName))
				} else {
					agent.DisallowedTools = append(agent.DisallowedTools, adapters.ToLowerCase(toolName))
				}
			}
		}
	}

	agent.Behavior = canonical.AgentBehavior{}
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
	if disabled, ok := input["disable"].(bool); ok {
		agent.Behavior.Disabled = disabled
	}

	if permissionMode, ok := input["permissionMode"].(string); ok {
		agent.PermissionPolicy = a.mapPermissionModeToPolicy(permissionMode)
	} else if permission, ok := input["permission"].(map[string]interface{}); ok {
		agent.PermissionPolicy = a.mapPermissionObjectToPolicy(permission)
	}

	if model, ok := input["model"].(string); ok {
		agent.Model = model
	}

	agent.Targets = make(canonical.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		agent.Targets = a.parseTargets(targets)
	}

	if skills, ok := input["skills"].([]interface{}); ok {
		if agent.Targets == nil {
			agent.Targets = make(canonical.PlatformConfig)
		}
		if agent.Targets["opencode"] == nil {
			agent.Targets["opencode"] = make(map[string]interface{})
		}
		skillNames := make([]string, 0, len(skills))
		for _, s := range skills {
			if skillName, ok := s.(string); ok {
				skillNames = append(skillNames, skillName)
			}
		}
		agent.Targets["opencode"]["skills"] = skillNames
	}

	return agent, nil
}

func (a *OpenCodeAdapter) renderAgent(agent *canonical.Agent) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "agent"
	output["name"] = agent.Name
	output["description"] = agent.Description

	tools := make(map[string]bool)
	for _, t := range agent.Tools {
		tools[adapters.ToLowerCase(t)] = true
	}
	for _, t := range agent.DisallowedTools {
		tools[adapters.ToLowerCase(t)] = false
	}
	if len(tools) > 0 {
		output["tools"] = tools
	}

	if agent.Behavior.Mode != "" {
		output["mode"] = agent.Behavior.Mode
	}
	if agent.Behavior.Temperature != nil {
		output["temperature"] = *agent.Behavior.Temperature
	}
	if agent.Behavior.Steps > 0 {
		output["maxSteps"] = agent.Behavior.Steps
	}
	if agent.Behavior.Prompt != "" {
		output["prompt"] = agent.Behavior.Prompt
	}
	if agent.Behavior.Hidden {
		output["hidden"] = agent.Behavior.Hidden
	}
	if agent.Behavior.Disabled {
		output["disable"] = agent.Behavior.Disabled
	}

	if agent.PermissionPolicy != "" {
		permission, err := a.PermissionPolicyToPlatform(agent.PermissionPolicy)
		if err != nil {
			return nil, err
		}
		output["permission"] = permission
	}

	if agent.Model != "" {
		output["model"] = agent.Model
	}

	if ocTargets, ok := agent.Targets["opencode"]; ok {
		if skills, ok := ocTargets["skills"].([]string); ok {
			output["skills"] = skills
		}
	}

	return output, nil
}

func (a *OpenCodeAdapter) parseCommand(input map[string]interface{}) (*canonical.Command, error) {
	cmd := &canonical.Command{}

	if name, ok := input["name"].(string); ok {
		cmd.Name = name
	}
	if description, ok := input["description"].(string); ok {
		cmd.Description = description
	}

	if tools, ok := input["allowed-tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				cmd.Tools = append(cmd.Tools, adapters.ToLowerCase(toolName))
			}
		}
	}

	cmd.Execution = canonical.CommandExecution{}
	if context, ok := input["context"].(string); ok {
		cmd.Execution.Context = context
	}
	if subtask, ok := input["subtask"].(bool); ok {
		cmd.Execution.Subtask = subtask
	}
	if agent, ok := input["agent"].(string); ok {
		cmd.Execution.Agent = agent
	}

	if argumentHint, ok := input["argument-hint"].(string); ok {
		cmd.Arguments.Hint = argumentHint
	}

	if model, ok := input["model"].(string); ok {
		cmd.Model = model
	}

	cmd.Targets = make(canonical.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		cmd.Targets = a.parseTargets(targets)
	}

	return cmd, nil
}

func (a *OpenCodeAdapter) renderCommand(cmd *canonical.Command) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "command"
	output["name"] = cmd.Name
	output["description"] = cmd.Description

	if len(cmd.Tools) > 0 {
		tools := make([]string, len(cmd.Tools))
		for i, t := range cmd.Tools {
			tools[i] = adapters.ToLowerCase(t)
		}
		output["allowed-tools"] = tools
	}

	if cmd.Execution.Agent != "" {
		output["agent"] = cmd.Execution.Agent
	}
	if cmd.Execution.Subtask {
		output["subtask"] = cmd.Execution.Subtask
	}
	if cmd.Execution.Context != "" {
		output["context"] = cmd.Execution.Context
	}

	if cmd.Model != "" {
		output["model"] = cmd.Model
	}

	return output, nil
}

func (a *OpenCodeAdapter) parseSkill(input map[string]interface{}) (*canonical.Skill, error) {
	skill := &canonical.Skill{}

	if name, ok := input["name"].(string); ok {
		skill.Name = name
	}
	if description, ok := input["description"].(string); ok {
		skill.Description = description
	}

	if tools, ok := input["allowed-tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				skill.Tools = append(skill.Tools, adapters.ToLowerCase(toolName))
			}
		}
	}

	skill.Extensions = canonical.SkillExtensions{}
	if license, ok := input["license"].(string); ok {
		skill.Extensions.License = license
	}
	if compatibility, ok := input["compatibility"].([]interface{}); ok {
		for _, c := range compatibility {
			if comp, ok := c.(string); ok {
				skill.Extensions.Compatibility = append(skill.Extensions.Compatibility, comp)
			}
		}
	}
	if metadata, ok := input["metadata"].(map[string]interface{}); ok {
		skill.Extensions.Metadata = make(map[string]string)
		for k, v := range metadata {
			if val, ok := v.(string); ok {
				skill.Extensions.Metadata[k] = val
			}
		}
	}
	if hooks, ok := input["hooks"].(map[string]interface{}); ok {
		skill.Extensions.Hooks = make(map[string]string)
		for k, v := range hooks {
			if val, ok := v.(string); ok {
				skill.Extensions.Hooks[k] = val
			}
		}
	}

	skill.Execution = canonical.SkillExecution{}
	if context, ok := input["context"].(string); ok {
		skill.Execution.Context = context
	}
	if agent, ok := input["agent"].(string); ok {
		skill.Execution.Agent = agent
	}
	if userInvocable, ok := input["user-invocable"].(bool); ok {
		skill.Execution.UserInvocable = userInvocable
	}

	if model, ok := input["model"].(string); ok {
		skill.Model = model
	}

	skill.Targets = make(canonical.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		skill.Targets = a.parseTargets(targets)
	}

	return skill, nil
}

func (a *OpenCodeAdapter) renderSkill(skill *canonical.Skill) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "skill"
	output["name"] = skill.Name
	output["description"] = skill.Description

	if len(skill.Tools) > 0 {
		tools := make([]string, len(skill.Tools))
		for i, t := range skill.Tools {
			tools[i] = adapters.ToLowerCase(t)
		}
		output["allowed-tools"] = tools
	}

	if skill.Extensions.License != "" || len(skill.Extensions.Compatibility) > 0 ||
		len(skill.Extensions.Metadata) > 0 || len(skill.Extensions.Hooks) > 0 {
		if skill.Extensions.License != "" {
			output["license"] = skill.Extensions.License
		}
		if len(skill.Extensions.Compatibility) > 0 {
			output["compatibility"] = skill.Extensions.Compatibility
		}
		if len(skill.Extensions.Metadata) > 0 {
			output["metadata"] = skill.Extensions.Metadata
		}
		if len(skill.Extensions.Hooks) > 0 {
			output["hooks"] = skill.Extensions.Hooks
		}
	}

	if skill.Execution.Agent != "" {
		output["agent"] = skill.Execution.Agent
	}
	if skill.Execution.Context != "" {
		output["context"] = skill.Execution.Context
	}
	if skill.Execution.UserInvocable {
		output["user-invocable"] = skill.Execution.UserInvocable
	}

	if skill.Model != "" {
		output["model"] = skill.Model
	}

	return output, nil
}

func (a *OpenCodeAdapter) parseMemory(input map[string]interface{}) (*canonical.Memory, error) {
	mem := &canonical.Memory{}

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

func (a *OpenCodeAdapter) renderMemory(mem *canonical.Memory) (map[string]interface{}, error) {
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

func (a *OpenCodeAdapter) parseTargets(input map[string]interface{}) canonical.PlatformConfig {
	targets := make(canonical.PlatformConfig)
	for platform, config := range input {
		if configMap, ok := config.(map[string]interface{}); ok {
			targets[platform] = configMap
		}
	}
	return targets
}

func (a *OpenCodeAdapter) mapPermissionModeToPolicy(mode string) canonical.PermissionPolicy {
	switch mode {
	case "default":
		return canonical.PermissionPolicyRestrictive
	case "acceptEdits":
		return canonical.PermissionPolicyBalanced
	case "dontAsk":
		return canonical.PermissionPolicyPermissive
	case "bypassPermissions":
		return canonical.PermissionPolicyUnrestricted
	case "plan":
		return canonical.PermissionPolicyAnalysis
	default:
		return ""
	}
}

func (a *OpenCodeAdapter) mapPermissionObjectToPolicy(permission map[string]interface{}) canonical.PermissionPolicy {
	editDenied := false
	bashDenied := false
	otherDeny := false
	allAllow := true

	for tool, toolPermissions := range permission {
		if toolPerms, ok := toolPermissions.(map[string]interface{}); ok {
			for _, action := range toolPerms {
				if actionStr, ok := action.(string); ok {
					if actionStr != "allow" {
						allAllow = false
					}
					if actionStr == "deny" {
						switch tool {
						case "edit":
							editDenied = true
						case "bash":
							bashDenied = true
						default:
							otherDeny = true
						}
					}
				}
			}
		}
	}

	if editDenied && bashDenied && !otherDeny {
		return canonical.PermissionPolicyAnalysis
	}
	if allAllow {
		return canonical.PermissionPolicyPermissive
	}

	return canonical.PermissionPolicyBalanced
}

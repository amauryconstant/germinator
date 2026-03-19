package opencode

import (
	"fmt"

	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/adapters"
)

// Adapter implements the Adapter interface for OpenCode platform.
type Adapter struct{}

// New creates and returns a new Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// ToCanonical converts OpenCode format to canonical models.
// It parses the input map based on the __type field and returns the appropriate canonical document type.
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

// FromCanonical converts canonical models to OpenCode format.
// It renders the canonical document into a map suitable for YAML serialization.
func (a *Adapter) FromCanonical(docType string, doc interface{}) (map[string]interface{}, error) {
	switch docType {
	case "agent":
		agent, ok := doc.(*domain.Agent)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *domain.Agent, got %T", doc), nil)
		}
		return a.renderAgent(agent)
	case "command":
		cmd, ok := doc.(*domain.Command)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *domain.Command, got %T", doc), nil)
		}
		return a.renderCommand(cmd)
	case "skill":
		skill, ok := doc.(*domain.Skill)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *domain.Skill, got %T", doc), nil)
		}
		return a.renderSkill(skill)
	case "memory":
		mem, ok := doc.(*domain.Memory)
		if !ok {
			return nil, domain.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *domain.Memory, got %T", doc), nil)
		}
		return a.renderMemory(mem)
	default:
		return nil, domain.NewTransformError("from-canonical", "opencode", "unknown document type: "+docType, nil)
	}
}

// PermissionPolicyToPlatform converts a canonical PermissionPolicy to OpenCode permission object format.
// It maps the policy to a PermissionMap with Allow/Ask/Deny actions for each tool.
func (a *Adapter) PermissionPolicyToPlatform(policy domain.PermissionPolicy) (interface{}, error) {
	mapping, ok := adapters.PermissionPolicyMappings[string(policy)]
	if !ok {
		return nil, domain.NewConfigError("permission-policy", string(policy), "unknown permission policy")
	}
	return mapping.OpenCode, nil
}

// ConvertToolNameCase converts a tool name to lowercase for OpenCode.
// OpenCode uses lowercase tool names, so this is an identity operation.
func (a *Adapter) ConvertToolNameCase(name string) string {
	return adapters.ToLowerCase(name)
}

func (a *Adapter) parseAgent(input map[string]interface{}) (*domain.Agent, error) {
	agent := &domain.Agent{}

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

	agent.Targets = make(domain.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		agent.Targets = a.parseTargets(targets)
	}

	if skills, ok := input["skills"].([]interface{}); ok {
		if agent.Targets == nil {
			agent.Targets = make(domain.PlatformConfig)
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

func (a *Adapter) renderAgent(agent *domain.Agent) (map[string]interface{}, error) {
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

func (a *Adapter) parseCommand(input map[string]interface{}) (*domain.Command, error) {
	cmd := &domain.Command{}

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

	cmd.Execution = domain.CommandExecution{}
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

func (a *Adapter) parseSkill(input map[string]interface{}) (*domain.Skill, error) {
	skill := &domain.Skill{}

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

	skill.Extensions = domain.SkillExtensions{}
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

	skill.Execution = domain.SkillExecution{}
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
	case "bypassPermissions":
		return domain.PermissionPolicyUnrestricted
	case "plan":
		return domain.PermissionPolicyAnalysis
	default:
		return ""
	}
}

func (a *Adapter) mapPermissionObjectToPolicy(permission map[string]interface{}) domain.PermissionPolicy {
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
		return domain.PermissionPolicyAnalysis
	}
	if allAllow {
		return domain.PermissionPolicyPermissive
	}

	return domain.PermissionPolicyBalanced
}

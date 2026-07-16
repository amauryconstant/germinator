package opencode

import (
	"fmt"
	"sort"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/permission"
)

// Adapter implements the Adapter interface for OpenCode platform.
type Adapter struct{}

// OpenCode is the package-level singleton for the OpenCode Adapter.
// The adapter is stateless; a single shared instance is safe to use across goroutines.
var OpenCode = &Adapter{}

// ToCanonical converts OpenCode format to canonical models.
// It parses the input map based on the __type field and returns the appropriate canonical document type.
func (a *Adapter) ToCanonical(input map[string]interface{}) (*core.Agent, *core.Command, *core.Skill, *core.Memory, error) {
	docType, ok := input["__type"].(string)
	if !ok {
		return nil, nil, nil, nil, core.NewParseError("", "missing __type field", nil)
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
		return nil, nil, nil, nil, core.NewParseError("", "unknown document type: "+docType, nil)
	}
}

// FromCanonical converts canonical models to OpenCode format.
// It renders the canonical document into a map suitable for YAML serialization.
func (a *Adapter) FromCanonical(docType string, doc interface{}) (map[string]interface{}, error) {
	switch docType {
	case "agent":
		agent, ok := doc.(*core.Agent)
		if !ok {
			return nil, core.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *core.Agent, got %T", doc), nil)
		}
		return a.renderAgent(agent)
	case "command":
		cmd, ok := doc.(*core.Command)
		if !ok {
			return nil, core.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *core.Command, got %T", doc), nil)
		}
		return a.renderCommand(cmd)
	case "skill":
		skill, ok := doc.(*core.Skill)
		if !ok {
			return nil, core.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *core.Skill, got %T", doc), nil)
		}
		return a.renderSkill(skill)
	case "memory":
		mem, ok := doc.(*core.Memory)
		if !ok {
			return nil, core.NewTransformError("from-canonical", "opencode", fmt.Sprintf("expected *core.Memory, got %T", doc), nil)
		}
		return a.renderMemory(mem)
	default:
		return nil, core.NewTransformError("from-canonical", "opencode", "unknown document type: "+docType, nil)
	}
}

// PermissionPolicyToPlatform converts a canonical PermissionPolicy to OpenCode permission object format.
// It maps the policy to a permission.Map with Allow/Ask/Deny actions for each tool.
func (a *Adapter) PermissionPolicyToPlatform(policy core.PermissionPolicy) (interface{}, error) {
	mapping, ok := permission.PermissionPolicyMappings[string(policy)]
	if !ok {
		return nil, core.NewConfigError("permission-policy", string(policy), "unknown permission policy")
	}
	return mapping.OpenCode, nil
}

// ConvertToolNameCase converts a tool name to lowercase for OpenCode.
// OpenCode uses lowercase tool names, so this is an identity operation.
func (a *Adapter) ConvertToolNameCase(name string) string {
	return permission.ToLowerCase(name)
}

//nolint:gocognit,unparam // parseAgent has high cognitive complexity due to nested map structure
func (a *Adapter) parseAgent(input map[string]interface{}) (*core.Agent, error) {
	agent := &core.Agent{}

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
					agent.Tools = append(agent.Tools, permission.ToLowerCase(toolName))
				} else {
					agent.DisallowedTools = append(agent.DisallowedTools, permission.ToLowerCase(toolName))
				}
			}
		}
	}

	agent.Behavior = core.AgentBehavior{}
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

	agent.Targets = make(core.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		agent.Targets = a.parseTargets(targets)
	}

	if skills, ok := input["skills"].([]interface{}); ok {
		if agent.Targets == nil {
			agent.Targets = make(core.PlatformConfig)
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

	sort.Strings(agent.Tools)
	sort.Strings(agent.DisallowedTools)

	return agent, nil
}

func (a *Adapter) renderAgent(agent *core.Agent) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "agent"
	output["name"] = agent.Name
	output["description"] = agent.Description

	tools := make(map[string]bool)
	for _, t := range agent.Tools {
		tools[permission.ToLowerCase(t)] = true
	}
	for _, t := range agent.DisallowedTools {
		tools[permission.ToLowerCase(t)] = false
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

//nolint:unparam // parseCommand always returns nil error - interface requires error return but implementation never fails
func (a *Adapter) parseCommand(input map[string]interface{}) (*core.Command, error) {
	cmd := &core.Command{}

	if name, ok := input["name"].(string); ok {
		cmd.Name = name
	}
	if description, ok := input["description"].(string); ok {
		cmd.Description = description
	}

	if tools, ok := input["allowed-tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				cmd.Tools = append(cmd.Tools, permission.ToLowerCase(toolName))
			}
		}
	}

	cmd.Execution = core.CommandExecution{}
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

	cmd.Targets = make(core.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		cmd.Targets = a.parseTargets(targets)
	}

	return cmd, nil
}

func (a *Adapter) renderCommand(cmd *core.Command) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "command"
	output["name"] = cmd.Name
	output["description"] = cmd.Description

	if len(cmd.Tools) > 0 {
		tools := make([]string, len(cmd.Tools))
		for i, t := range cmd.Tools {
			tools[i] = permission.ToLowerCase(t)
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

//nolint:gocognit,unparam // parseSkill has high cognitive complexity due to nested map structure
func (a *Adapter) parseSkill(input map[string]interface{}) (*core.Skill, error) {
	skill := &core.Skill{}

	if name, ok := input["name"].(string); ok {
		skill.Name = name
	}
	if description, ok := input["description"].(string); ok {
		skill.Description = description
	}

	if tools, ok := input["allowed-tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				skill.Tools = append(skill.Tools, permission.ToLowerCase(toolName))
			}
		}
	}

	skill.Extensions = core.SkillExtensions{}
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

	skill.Execution = core.SkillExecution{}
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

	skill.Targets = make(core.PlatformConfig)
	if targets, ok := input["targets"].(map[string]interface{}); ok {
		skill.Targets = a.parseTargets(targets)
	}

	return skill, nil
}

func (a *Adapter) renderSkill(skill *core.Skill) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	output["__type"] = "skill"
	output["name"] = skill.Name
	output["description"] = skill.Description

	if len(skill.Tools) > 0 {
		tools := make([]string, len(skill.Tools))
		for i, t := range skill.Tools {
			tools[i] = permission.ToLowerCase(t)
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

//nolint:unparam // parseMemory always returns nil error - interface requires error return but implementation never fails
func (a *Adapter) parseMemory(input map[string]interface{}) (*core.Memory, error) {
	mem := &core.Memory{}

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

func (a *Adapter) renderMemory(mem *core.Memory) (map[string]interface{}, error) {
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

func (a *Adapter) parseTargets(input map[string]interface{}) core.PlatformConfig {
	targets := make(core.PlatformConfig)
	for platform, config := range input {
		if configMap, ok := config.(map[string]interface{}); ok {
			targets[platform] = configMap
		}
	}
	return targets
}

func (a *Adapter) mapPermissionModeToPolicy(mode string) core.PermissionPolicy {
	switch mode {
	case "default":
		return core.PermissionPolicyRestrictive
	case "acceptEdits":
		return core.PermissionPolicyBalanced
	case "dontAsk":
		return core.PermissionPolicyPermissive
	case "bypassPermissions":
		return core.PermissionPolicyUnrestricted
	case "plan":
		return core.PermissionPolicyAnalysis
	default:
		return ""
	}
}

func (a *Adapter) mapPermissionObjectToPolicy(perm map[string]interface{}) core.PermissionPolicy {
	editDenied := false
	bashDenied := false
	otherDeny := false
	allAllow := true

	for tool, toolPermissions := range perm {
		if toolPerms, ok := toolPermissions.(map[string]interface{}); ok {
			for _, action := range toolPerms {
				actionStr, ok := action.(string)
				if !ok {
					continue
				}
				actionConst := permission.Action(actionStr)
				if actionConst != permission.Allow {
					allAllow = false
				}
				if actionConst == permission.Deny {
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

	if editDenied && bashDenied && !otherDeny {
		return core.PermissionPolicyAnalysis
	}
	if allAllow {
		return core.PermissionPolicyPermissive
	}

	return core.PermissionPolicyBalanced
}

// Compile-time check: *Adapter satisfies permission.Adapter. Mirror of cmd/canonicalize_test.go:20 precedent.
var _ permission.Adapter = (*Adapter)(nil)

// Package core provides document parsing and serialization functionality.
package core

import (
	"fmt"
	"os"

	claudecode "gitlab.com/amoconst/germinator/internal/adapters/claude-code"
	opencode "gitlab.com/amoconst/germinator/internal/adapters/opencode"
	"gitlab.com/amoconst/germinator/internal/models/canonical"
	yaml "gopkg.in/yaml.v3"
)

// ParsePlatformDocument parses a platform YAML file and converts it to a canonical model.
func ParsePlatformDocument(path string, platform string, docType string) (interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)
	yamlContent, markdownBody, err := extractFrontmatter(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	var input map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &input); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if input == nil {
		input = make(map[string]interface{})
	}

	var adapter interface{}
	switch platform {
	case "claude-code":
		adapter = claudecode.New()
	case "opencode":
		adapter = opencode.New()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	input["__type"] = docType

	agent, command, skill, memory, err := adapter.(interface {
		ToCanonical(map[string]interface{}) (*canonical.Agent, *canonical.Command, *canonical.Skill, *canonical.Memory, error)
	}).ToCanonical(input)

	if err != nil {
		return nil, fmt.Errorf("failed to convert to canonical: %w", err)
	}

	switch docType {
	case "agent":
		if agent == nil {
			return nil, fmt.Errorf("expected agent but got nil")
		}
		return &CanonicalAgent{
			Agent:    *agent,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "command":
		if command == nil {
			return nil, fmt.Errorf("expected command but got nil")
		}
		return &CanonicalCommand{
			Command:  *command,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "skill":
		if skill == nil {
			return nil, fmt.Errorf("expected skill but got nil")
		}
		return &CanonicalSkill{
			Skill:    *skill,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "memory":
		if memory == nil {
			return nil, fmt.Errorf("expected memory but got nil")
		}
		if markdownBody != "" {
			memory.Content = markdownBody
		}
		return &CanonicalMemory{
			Memory:   *memory,
			FilePath: path,
			Content:  memory.Content,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}

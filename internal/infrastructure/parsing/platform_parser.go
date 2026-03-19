// Package parsing provides platform document parsing functionality.
package parsing

import (
	"os"

	"gitlab.com/amoconst/germinator/internal/domain"
	claudecode "gitlab.com/amoconst/germinator/internal/infrastructure/adapters/claude-code"
	opencode "gitlab.com/amoconst/germinator/internal/infrastructure/adapters/opencode"
	yaml "gopkg.in/yaml.v3"
)

// ParsePlatformDocument parses a platform YAML file and converts it to a canonical model.
func ParsePlatformDocument(path string, platform string, docType string) (interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, domain.NewFileError(path, "read", "failed to read file", err)
	}

	fileContent := string(content)
	yamlContent, markdownBody, err := extractFrontmatter(fileContent)
	if err != nil {
		return nil, domain.NewParseError(path, "failed to extract frontmatter", err)
	}

	var input map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &input); err != nil {
		return nil, domain.NewParseError(path, "failed to parse YAML", err)
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
		return nil, domain.NewConfigError("platform", platform, "unsupported platform")
	}

	input["__type"] = docType

	agent, command, skill, memory, err := adapter.(interface {
		ToCanonical(map[string]interface{}) (*domain.Agent, *domain.Command, *domain.Skill, *domain.Memory, error)
	}).ToCanonical(input)

	if err != nil {
		return nil, domain.NewParseError(path, "failed to convert to canonical", err)
	}

	switch docType {
	case "agent":
		if agent == nil {
			return nil, domain.NewParseError(path, "expected agent but got nil", nil)
		}
		return &CanonicalAgent{
			Agent:    *agent,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "command":
		if command == nil {
			return nil, domain.NewParseError(path, "expected command but got nil", nil)
		}
		return &CanonicalCommand{
			Command:  *command,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "skill":
		if skill == nil {
			return nil, domain.NewParseError(path, "expected skill but got nil", nil)
		}
		return &CanonicalSkill{
			Skill:    *skill,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "memory":
		if memory == nil {
			return nil, domain.NewParseError(path, "expected memory but got nil", nil)
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
		return nil, domain.NewParseError(path, "unsupported document type: "+docType, nil)
	}
}

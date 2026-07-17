// Package parser provides platform document parsing functionality.
package parser

import (
	"context"
	"fmt"
	"os"

	claudecode "gitlab.com/amoconst/germinator/internal/claude-code"
	"gitlab.com/amoconst/germinator/internal/core"
	opencode "gitlab.com/amoconst/germinator/internal/opencode"
	yaml "gopkg.in/yaml.v3"
)

// platformAdapter is the narrow contract parser needs from a platform
// adapter. Defined here (the consumer) per the "interfaces where
// consumed" rule (references/01-architecture.md); both
// internal/claude-code and internal/opencode satisfy it via structural
// typing, so no shim or compile-time tag is required.
type platformAdapter interface {
	ToCanonical(input map[string]interface{}) (*core.Agent, *core.Command, *core.Skill, *core.Memory, error)
}

// ParsePlatformDocument parses a platform YAML file and converts it to a canonical model.
// The ctx parameter is checked before the file read so caller cancellation
// propagates before blocking I/O is attempted.
func ParsePlatformDocument(ctx context.Context, path string, platform string, docType string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("parser: platform parse cancelled: %w", err)
	}

	content, err := os.ReadFile(path) //nolint:gosec // G304: User provides file path, tool must read user documents
	if err != nil {
		return nil, core.NewFileError(path, "read", "failed to read file", err)
	}

	fileContent := string(content)
	yamlContent, markdownBody, err := extractFrontmatter(fileContent)
	if err != nil {
		return nil, core.NewParseError(path, "failed to extract frontmatter", err)
	}

	var input map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &input); err != nil {
		return nil, core.NewParseError(path, "failed to parse YAML", err)
	}

	if input == nil {
		input = make(map[string]interface{})
	}

	var adapter platformAdapter
	switch platform {
	case "claude-code":
		adapter = claudecode.ClaudeCode
	case "opencode":
		adapter = opencode.OpenCode
	default:
		return nil, core.NewConfigError("platform", platform, "unsupported platform")
	}

	input["__type"] = docType

	agent, command, skill, memory, err := adapter.ToCanonical(input)

	if err != nil {
		return nil, core.NewParseError(path, "failed to convert to canonical", err)
	}

	switch docType {
	case "agent":
		if agent == nil {
			return nil, core.NewParseError(path, "expected agent but got nil", nil)
		}
		return &CanonicalAgent{
			Agent:    *agent,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "command":
		if command == nil {
			return nil, core.NewParseError(path, "expected command but got nil", nil)
		}
		return &CanonicalCommand{
			Command:  *command,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "skill":
		if skill == nil {
			return nil, core.NewParseError(path, "expected skill but got nil", nil)
		}
		return &CanonicalSkill{
			Skill:    *skill,
			FilePath: path,
			Content:  markdownBody,
		}, nil
	case "memory":
		if memory == nil {
			return nil, core.NewParseError(path, "expected memory but got nil", nil)
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
		return nil, core.NewParseError(path, "unsupported document type: "+docType, nil)
	}
}

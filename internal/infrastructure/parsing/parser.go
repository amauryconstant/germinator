// Package parsing provides document parsing functionality.
package parsing

import (
	"os"
	"strings"

	"gitlab.com/amoconst/germinator/internal/domain"
	yaml "gopkg.in/yaml.v3"
)

// CanonicalAgent extends the Agent domain model with FilePath and Content fields.
type CanonicalAgent struct {
	domain.Agent
	FilePath string
	Content  string
}

// CanonicalCommand extends the Command domain model with FilePath and Content fields.
type CanonicalCommand struct {
	domain.Command
	FilePath string
	Content  string
}

// CanonicalSkill extends the Skill domain model with FilePath and Content fields.
type CanonicalSkill struct {
	domain.Skill
	FilePath string
	Content  string
}

// CanonicalMemory extends the Memory domain model with FilePath and Content fields.
type CanonicalMemory struct {
	domain.Memory
	FilePath string
	Content  string
}

// ParseDocument parses a document file and returns the appropriate struct.
func ParseDocument(filePath string, docType string) (interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, domain.NewFileError(filePath, "read", "failed to read file", err)
	}

	fileContent := string(content)

	switch docType {
	case "memory":
		return parseMemory(filePath, fileContent)

	case "agent", "command", "skill":
		return parseDocumentWithFrontmatter(filePath, fileContent, docType)

	default:
		return nil, domain.NewParseError(filePath, "unsupported document type: "+docType, nil)
	}
}

func parseMemory(filePath string, content string) (interface{}, error) {
	memory := &CanonicalMemory{
		Memory: domain.Memory{
			Content: content,
		},
		FilePath: filePath,
		Content:  "",
	}

	lines := strings.Split(content, "\n")
	if len(lines) >= 2 && lines[0] == "---" {
		var yamlLines []string
		var bodyLines []string
		foundEnd := false

		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				foundEnd = true
				bodyLines = lines[i+1:]
				break
			}
			yamlLines = append(yamlLines, lines[i])
		}

		yamlContent := strings.Join(yamlLines, "\n")
		var frontmatter map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err == nil {
			if paths, ok := frontmatter["paths"].([]interface{}); ok {
				for _, p := range paths {
					if pathStr, ok := p.(string); ok {
						memory.Paths = append(memory.Paths, pathStr)
					}
				}
			}
			if _, ok := frontmatter["content"].(string); ok {
				memory.Content = extractContentFromYamlLines(yamlLines)
				memory.Memory.Content = extractContentFromYamlLines(yamlLines)
			} else if foundEnd {
				memory.Content = strings.Join(bodyLines, "\n")
				memory.Memory.Content = strings.Join(bodyLines, "\n")
			} else {
				memory.Content = ""
				memory.Memory.Content = ""
			}
		} else if foundEnd {
			memory.Content = strings.Join(bodyLines, "\n")
			memory.Memory.Content = strings.Join(bodyLines, "\n")
		}
	} else {
		memory.Content = content
		memory.Memory.Content = content
	}

	return memory, nil
}

func extractContentFromYamlLines(yamlLines []string) string {
	for i, line := range yamlLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "content:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == "|" {
				contentLines := []string{}
				j := i + 1
				for j < len(yamlLines) {
					nextLine := yamlLines[j]
					if nextLine == "" || strings.HasPrefix(nextLine, " ") || strings.HasPrefix(nextLine, "\t") {
						contentLines = append(contentLines, nextLine)
						j++
					} else {
						break
					}
				}
				return strings.Join(contentLines, "\n")
			}
		}
	}
	return ""
}

func parseDocumentWithFrontmatter(filePath string, fileContent string, docType string) (interface{}, error) {
	yamlContent, markdownBody, err := extractFrontmatter(fileContent)
	if err != nil {
		return nil, domain.NewParseError(filePath, "failed to extract frontmatter", err)
	}

	var doc interface{}
	switch docType {
	case "agent":
		var agent CanonicalAgent
		if err := yaml.Unmarshal([]byte(yamlContent), &agent.Agent); err != nil {
			return nil, domain.NewParseError(filePath, "failed to parse agent", err)
		}
		agent.FilePath = filePath
		agent.Content = markdownBody
		doc = &agent

	case "command":
		var command CanonicalCommand
		if err := yaml.Unmarshal([]byte(yamlContent), &command.Command); err != nil {
			return nil, domain.NewParseError(filePath, "failed to parse command", err)
		}
		command.FilePath = filePath
		command.Content = markdownBody
		doc = &command

	case "skill":
		var skill CanonicalSkill
		if err := yaml.Unmarshal([]byte(yamlContent), &skill.Skill); err != nil {
			return nil, domain.NewParseError(filePath, "failed to parse skill", err)
		}
		skill.FilePath = filePath
		skill.Content = markdownBody
		doc = &skill
	}

	return doc, nil
}

//nolint:unparam // extractFrontmatter always returns nil error - function design never fails
func extractFrontmatter(content string) (yamlContent string, markdownBody string, err error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return "", content, nil
	}

	if lines[0] != "---" {
		return "", content, nil
	}

	var yamlLines []string
	var bodyLines []string
	foundEnd := false

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			foundEnd = true
			bodyLines = lines[i+1:]
			break
		}
		yamlLines = append(yamlLines, lines[i])
	}

	if !foundEnd {
		return "", content, nil
	}

	return strings.Join(yamlLines, "\n"), strings.Join(bodyLines, "\n"), nil
}

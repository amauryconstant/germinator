// Package core provides document parsing and serialization functionality.
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/amoconst/germinator/internal/models"
	yaml "gopkg.in/yaml.v3"
)

// ParseDocument parses a document file and returns the appropriate struct.
func ParseDocument(filePath string, docType string) (interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)

	switch docType {
	case "memory":
		return parseMemory(filePath, fileContent)

	case "agent", "command", "skill":
		return parseDocumentWithFrontmatter(filePath, fileContent, docType)

	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}

func parseMemory(filePath string, content string) (interface{}, error) {
	memory := &models.Memory{
		FilePath: filePath,
		Content:  content,
	}

	lines := strings.Split(content, "\n")
	if len(lines) >= 3 && lines[0] == "---" {
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

		if foundEnd {
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
			}
			memory.Content = strings.Join(bodyLines, "\n")
		}
	}

	return memory, nil
}

func parseDocumentWithFrontmatter(filePath string, fileContent string, docType string) (interface{}, error) {
	yamlContent, markdownBody, err := extractFrontmatter(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	var doc interface{}
	switch docType {
	case "agent":
		var agent models.Agent
		if err := yaml.Unmarshal([]byte(yamlContent), &agent); err != nil {
			return nil, fmt.Errorf("failed to parse agent: %w", err)
		}
		agent.FilePath = filePath
		agent.Content = markdownBody
		doc = &agent

	case "command":
		var command models.Command
		if err := yaml.Unmarshal([]byte(yamlContent), &command); err != nil {
			return nil, fmt.Errorf("failed to parse command: %w", err)
		}
		command.FilePath = filePath
		command.Content = markdownBody
		command.Name = extractCommandName(filePath)
		doc = &command

	case "skill":
		var skill models.Skill
		if err := yaml.Unmarshal([]byte(yamlContent), &skill); err != nil {
			return nil, fmt.Errorf("failed to parse skill: %w", err)
		}
		skill.FilePath = filePath
		skill.Content = markdownBody
		doc = &skill
	}

	return doc, nil
}

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

func extractCommandName(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return name
}

// Package parser provides document parsing functionality.
package parser

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gitlab.com/amoconst/germinator/internal/core"
	yaml "gopkg.in/yaml.v3"
)

// CanonicalAgent extends the Agent domain model with FilePath and Content fields.
type CanonicalAgent struct {
	core.Agent
	FilePath string
	Content  string
}

// CanonicalCommand extends the Command domain model with FilePath and Content fields.
type CanonicalCommand struct {
	core.Command
	FilePath string
	Content  string
}

// CanonicalSkill extends the Skill domain model with FilePath and Content fields.
type CanonicalSkill struct {
	core.Skill
	FilePath string
	Content  string
}

// CanonicalMemory extends the Memory domain model with FilePath and Content fields.
type CanonicalMemory struct {
	core.Memory
	FilePath string
	Content  string
}

// ParseDocument parses a document file and returns the appropriate struct.
// The ctx parameter is checked before the file read so caller cancellation
// propagates before blocking I/O is attempted.
func ParseDocument(ctx context.Context, filePath string, docType string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("parser: parse cancelled: %w", err)
	}

	content, err := os.ReadFile(filePath) //nolint:gosec // G304: User provides file path, tool must read user documents
	if err != nil {
		return nil, core.NewFileError(filePath, "read", "failed to read file", err)
	}

	fileContent := string(content)

	switch docType {
	case "memory":
		return parseMemory(ctx, filePath, fileContent)

	case "agent", "command", "skill":
		return parseDocumentWithFrontmatter(ctx, filePath, fileContent, docType)

	default:
		return nil, core.NewParseError(filePath, "unsupported document type: "+docType, nil)
	}
}

func parseMemory(ctx context.Context, filePath string, content string) (interface{}, error) {
	memory := &CanonicalMemory{
		Memory: core.Memory{
			Content: content,
		},
		FilePath: filePath,
		Content:  "",
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 2 || lines[0] != "---" {
		memory.Content = content
		memory.Memory.Content = content
		return memory, nil
	}

	yamlLines, bodyLines, foundEnd, err := scanMemoryFrontmatter(ctx, lines)
	if err != nil {
		return nil, err
	}
	applyMemoryFrontmatter(memory, yamlLines, bodyLines, foundEnd)
	return memory, nil
}

// scanMemoryFrontmatter walks the lines after the opening `---` and returns
// the YAML portion, body portion, and whether a closing `---` was found. The
// ctx check between iterations lets a cancelled caller terminate the scan.
func scanMemoryFrontmatter(ctx context.Context, lines []string) (yamlLines []string, bodyLines []string, foundEnd bool, err error) {
	for i := 1; i < len(lines); i++ {
		if cerr := ctx.Err(); cerr != nil {
			return nil, nil, false, fmt.Errorf("parser: memory scan cancelled: %w", cerr)
		}
		if lines[i] == "---" {
			foundEnd = true
			bodyLines = lines[i+1:]
			break
		}
		yamlLines = append(yamlLines, lines[i])
	}
	return yamlLines, bodyLines, foundEnd, nil
}

// applyMemoryFrontmatter decodes the frontmatter YAML and applies paths /
// content to the memory struct. Body content is used when the YAML has no
// `content:` field but a closing delimiter was found.
func applyMemoryFrontmatter(memory *CanonicalMemory, yamlLines, bodyLines []string, foundEnd bool) {
	yamlContent := strings.Join(yamlLines, "\n")
	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		if foundEnd {
			memory.Content = strings.Join(bodyLines, "\n")
			memory.Memory.Content = strings.Join(bodyLines, "\n")
		}
		return
	}
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
		return
	}
	if foundEnd {
		memory.Content = strings.Join(bodyLines, "\n")
		memory.Memory.Content = strings.Join(bodyLines, "\n")
		return
	}
	memory.Content = ""
	memory.Memory.Content = ""
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

func parseDocumentWithFrontmatter(ctx context.Context, filePath string, fileContent string, docType string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("parser: parse cancelled: %w", err)
	}

	yamlContent, markdownBody, err := extractFrontmatter(fileContent)
	if err != nil {
		return nil, core.NewParseError(filePath, "failed to extract frontmatter", err)
	}

	var doc interface{}
	switch docType {
	case "agent":
		var agent CanonicalAgent
		if err := yaml.Unmarshal([]byte(yamlContent), &agent.Agent); err != nil {
			return nil, core.NewParseError(filePath, "failed to parse agent", err)
		}
		agent.FilePath = filePath
		agent.Content = markdownBody
		doc = &agent

	case "command":
		var command CanonicalCommand
		if err := yaml.Unmarshal([]byte(yamlContent), &command.Command); err != nil {
			return nil, core.NewParseError(filePath, "failed to parse command", err)
		}
		command.FilePath = filePath
		command.Content = markdownBody
		doc = &command

	case "skill":
		var skill CanonicalSkill
		if err := yaml.Unmarshal([]byte(yamlContent), &skill.Skill); err != nil {
			return nil, core.NewParseError(filePath, "failed to parse skill", err)
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

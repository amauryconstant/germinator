package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/models"
)

func TestTransformDocumentSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")
	outputFile := filepath.Join(tmpDir, "output-agent.md")

	content := `---
name: test-agent
description: A test agent
tools:
  - editor
  - bash
---
This is test content
`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err != nil {
		t.Fatalf("TransformDocument failed: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(outputContent) == 0 {
		t.Error("Output file is empty")
	}
}

func TestTransformDocumentParseFailure(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")
	outputFile := filepath.Join(tmpDir, "output-agent.md")

	invalidContent := `---
name: "test-agent" "invalid yaml
---
content`

	if err := os.WriteFile(inputFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}

	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Error("Output file should not exist on parse failure")
	}
}

func TestTransformDocumentWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")
	outputFile := "/nonexistent/directory/output.md"

	content := `---
name: test-agent
description: A test agent
---
content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err == nil {
		t.Error("Expected error for non-existent output directory")
	}
}

func TestValidateDocumentSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")

	content := `---
name: test-agent
description: A test agent
---
content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	errs, err := ValidateDocument(inputFile, "claude-code")
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateDocumentFailure(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")

	content := `---
name: TEST-AGENT
description: ""
---
content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	errs, err := ValidateDocument(inputFile, "claude-code")
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestValidateDocumentMissingFile(t *testing.T) {
	nonExistentFile := "/nonexistent/file.md"

	_, err := ValidateDocument(nonExistentFile, "claude-code")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestTransformAndRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input-agent.md")
	outputFile := filepath.Join(tmpDir, "output-agent.md")

	originalContent := `---
name: test-agent
description: A test agent
tools:
  - editor
  - bash
  - grep
model: sonnet
permissionMode: default
---
This is the agent content
It has multiple lines
And preserves markdown **formatting**
`

	if err := os.WriteFile(inputFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err != nil {
		t.Fatalf("TransformDocument failed: %v", err)
	}

	doc1, err := core.LoadDocument(inputFile, "claude-code")
	if err != nil {
		t.Fatalf("Failed to load original document: %v", err)
	}

	doc2, err := core.LoadDocument(outputFile, "claude-code")
	if err != nil {
		t.Fatalf("Failed to load transformed document: %v", err)
	}

	agent1, ok1 := doc1.(*models.Agent)
	agent2, ok2 := doc2.(*models.Agent)

	if !ok1 || !ok2 {
		t.Fatal("Documents are not of expected type")
	}

	if agent1.Name != agent2.Name {
		t.Errorf("Name mismatch: %q != %q", agent1.Name, agent2.Name)
	}
	if agent1.Description != agent2.Description {
		t.Errorf("Description mismatch: %q != %q", agent1.Description, agent2.Description)
	}
	if len(agent1.Tools) != len(agent2.Tools) {
		t.Errorf("Tools count mismatch: %d != %d", len(agent1.Tools), len(agent2.Tools))
	}
	if agent1.Content != agent2.Content {
		t.Errorf("Content mismatch:\nOriginal: %q\nGot:      %q", agent1.Content, agent2.Content)
	}
}

func TestValidateOpenCodeAgent(t *testing.T) {
	tests := []struct {
		name      string
		agent     *models.Agent
		wantCount int
	}{
		{
			name: "valid mode primary",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "primary",
			},
			wantCount: 0,
		},
		{
			name: "valid mode subagent",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "subagent",
			},
			wantCount: 0,
		},
		{
			name: "valid mode all",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "all",
			},
			wantCount: 0,
		},
		{
			name: "valid mode empty",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "",
			},
			wantCount: 0,
		},
		{
			name: "invalid mode",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "invalid-mode",
			},
			wantCount: 1,
		},
		{
			name: "valid temperature 0.0",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Temperature: 0.0,
			},
			wantCount: 0,
		},
		{
			name: "valid temperature 0.5",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Temperature: 0.5,
			},
			wantCount: 0,
		},
		{
			name: "valid temperature 1.0",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Temperature: 1.0,
			},
			wantCount: 0,
		},
		{
			name: "invalid temperature negative",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Temperature: -0.5,
			},
			wantCount: 1,
		},
		{
			name: "invalid temperature too high",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Temperature: 1.5,
			},
			wantCount: 1,
		},
		{
			name: "valid maxSteps 1",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				MaxSteps:    1,
			},
			wantCount: 0,
		},
		{
			name: "valid maxSteps 50",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				MaxSteps:    50,
			},
			wantCount: 0,
		},
		{
			name: "invalid maxSteps negative",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				MaxSteps:    -5,
			},
			wantCount: 1,
		},
		{
			name: "multiple validation errors",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "invalid",
				Temperature: 1.5,
				MaxSteps:    -1,
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateOpenCodeAgent(tt.agent)
			if len(errs) != tt.wantCount {
				t.Errorf("validateOpenCodeAgent() error count = %d, want %d, errors: %v", len(errs), tt.wantCount, errs)
			}
		})
	}
}

func TestValidateOpenCodeCommand(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *models.Command
		wantCount int
	}{
		{
			name: "template present",
			cmd: &models.Command{
				Name:        "test-command",
				Description: "Test command",
				Content:     "echo $ARGUMENTS",
			},
			wantCount: 0,
		},
		{
			name: "template empty",
			cmd: &models.Command{
				Name:        "test-command",
				Description: "Test command",
				Content:     "",
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateOpenCodeCommand(tt.cmd)
			if len(errs) != tt.wantCount {
				t.Errorf("validateOpenCodeCommand() error count = %d, want %d, errors: %v", len(errs), tt.wantCount, errs)
			}
		})
	}
}

func TestValidateOpenCodeSkill(t *testing.T) {
	tests := []struct {
		name      string
		skill     *models.Skill
		wantCount int
	}{
		{
			name: "valid name git-workflow",
			skill: &models.Skill{
				Name:        "git-workflow",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 0,
		},
		{
			name: "valid name code-review-tool-enhanced",
			skill: &models.Skill{
				Name:        "code-review-tool-enhanced",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 0,
		},
		{
			name: "valid name git2-operations",
			skill: &models.Skill{
				Name:        "git2-operations",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 0,
		},
		{
			name: "invalid name consecutive hyphens",
			skill: &models.Skill{
				Name:        "git--workflow",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 1,
		},
		{
			name: "invalid name leading hyphen",
			skill: &models.Skill{
				Name:        "-git-workflow",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 1,
		},
		{
			name: "invalid name trailing hyphen",
			skill: &models.Skill{
				Name:        "git-workflow-",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 1,
		},
		{
			name: "invalid name uppercase",
			skill: &models.Skill{
				Name:        "Git-Workflow",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 1,
		},
		{
			name: "invalid name underscores",
			skill: &models.Skill{
				Name:        "git_workflow",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 1,
		},
		{
			name: "content present",
			skill: &models.Skill{
				Name:        "test-skill",
				Description: "Test skill",
				Content:     "Skill content",
			},
			wantCount: 0,
		},
		{
			name: "content empty",
			skill: &models.Skill{
				Name:        "test-skill",
				Description: "Test skill",
				Content:     "",
			},
			wantCount: 1,
		},
		{
			name: "multiple validation errors",
			skill: &models.Skill{
				Name:        "Git_Workflow",
				Description: "Test skill",
				Content:     "",
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateOpenCodeSkill(tt.skill)
			if len(errs) != tt.wantCount {
				t.Errorf("validateOpenCodeSkill() error count = %d, want %d, errors: %v", len(errs), tt.wantCount, errs)
			}
		})
	}
}

func TestValidateOpenCodeMemory(t *testing.T) {
	tests := []struct {
		name      string
		mem       *models.Memory
		wantCount int
	}{
		{
			name: "paths only",
			mem: &models.Memory{
				Paths: []string{"README.md", "CONTRIBUTING.md"},
			},
			wantCount: 0,
		},
		{
			name: "content only",
			mem: &models.Memory{
				Content: "Project context and setup instructions",
			},
			wantCount: 0,
		},
		{
			name: "both paths and content",
			mem: &models.Memory{
				Paths:   []string{"README.md"},
				Content: "Additional context",
			},
			wantCount: 0,
		},
		{
			name:      "both empty",
			mem:       &models.Memory{},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateOpenCodeMemory(tt.mem)
			if len(errs) != tt.wantCount {
				t.Errorf("validateOpenCodeMemory() error count = %d, want %d, errors: %v", len(errs), tt.wantCount, errs)
			}
		})
	}
}

func TestTransformOpenCodeAgent(t *testing.T) {
	tests := []struct {
		name         string
		inputContent string
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name: "minimal agent transforms correctly",
			inputContent: `---
name: test-agent
description: Test agent
---
Agent content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "description: Test agent") {
					t.Error("Expected description in output")
				}
				if !strings.Contains(output, "mode: all") {
					t.Error("Expected default mode: all")
				}
			},
		},
		{
			name: "full agent transforms correctly",
			inputContent: `---
name: test-agent
description: Test agent
mode: primary
temperature: 0.5
maxSteps: 50
model: anthropic/claude-sonnet-4-20250514
tools:
  - read
  - write
permissionMode: acceptEdits
---
Agent content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "mode: primary") {
					t.Error("Expected mode: primary")
				}
				if !strings.Contains(output, "temperature: 0.5") {
					t.Error("Expected temperature: 0.5")
				}
				if !strings.Contains(output, "maxSteps: 50") {
					t.Error("Expected maxSteps: 50")
				}
				if !strings.Contains(output, "model: anthropic/claude-sonnet-4-20250514") {
					t.Error("Expected full model ID")
				}
			},
		},
		{
			name: "mixed tools transform to map",
			inputContent: `---
name: test-agent
description: Test agent
tools:
  - read
  - write
disallowedTools:
  - dangerous
---
Agent content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "read: true") {
					t.Error("Expected read: true")
				}
				if !strings.Contains(output, "write: true") {
					t.Error("Expected write: true")
				}
				if !strings.Contains(output, "dangerous: false") {
					t.Error("Expected dangerous: false")
				}
			},
		},
		{
			name: "all permission modes transform correctly",
			inputContent: `---
name: test-agent
description: Test agent
permissionMode: dontAsk
---
Agent content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "permission:") {
					t.Error("Expected permission section")
				}
			},
		},
		{
			name: "agent mode defaults to all when empty",
			inputContent: `---
name: test-agent
description: Test agent
---
Agent content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "mode: all") {
					t.Error("Expected default mode: all")
				}
			},
		},
		{
			name: "OpenCode-specific fields preserved",
			inputContent: `---
name: test-agent
description: Test agent
mode: subagent
temperature: 0.1
maxSteps: 100
hidden: true
prompt: Custom prompt
disable: false
---
Agent content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "mode: subagent") {
					t.Error("Expected mode: subagent")
				}
				if !strings.Contains(output, "temperature: 0.1") {
					t.Error("Expected temperature: 0.1")
				}
				if !strings.Contains(output, "maxSteps: 100") {
					t.Error("Expected maxSteps: 100")
				}
				if !strings.Contains(output, "hidden: true") {
					t.Error("Expected hidden: true")
				}
				if !strings.Contains(output, "prompt: Custom prompt") {
					t.Error("Expected prompt field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test-agent.md")
			outputFile := filepath.Join(tmpDir, "output.md")

			if err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			if err := TransformDocument(inputFile, outputFile, "opencode"); err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			tt.checkOutput(t, string(output))
		})
	}
}

func TestTransformOpenCodeCommand(t *testing.T) {
	tests := []struct {
		name         string
		inputContent string
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name: "minimal command transforms correctly",
			inputContent: `---
name: test-command
description: Test command
---
Command content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "description: Test command") {
					t.Error("Expected description in output")
				}
				if !strings.Contains(output, "Command content") {
					t.Error("Expected content in output")
				}
			},
		},
		{
			name: "command with $ARGUMENTS placeholder preserved",
			inputContent: `---
name: test-command
description: Test command
---
Run with: $ARGUMENTS`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "$ARGUMENTS") {
					t.Error("Expected $ARGUMENTS preserved")
				}
			},
		},
		{
			name: "full command transforms correctly",
			inputContent: `---
name: test-command
description: Test command
agent: build-agent
model: anthropic/claude-sonnet-4-20250514
subtask: true
---
Command content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "agent: build-agent") {
					t.Error("Expected agent field")
				}
				if !strings.Contains(output, "model: anthropic/claude-sonnet-4-20250514") {
					t.Error("Expected model field")
				}
				if !strings.Contains(output, "subtask: true") {
					t.Error("Expected subtask: true")
				}
			},
		},
		{
			name: "subtask field renders correctly",
			inputContent: `---
name: test-command
description: Test command
subtask: false
---
Command content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Command content") {
					t.Error("Expected content")
				}
			},
		},
		{
			name: "content and indentation preserved",
			inputContent: `---
name: test-command
description: Test command
---
Run command with bash output: echo test
ls -la`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Run command") {
					t.Error("Expected content preserved")
				}
				if !strings.Contains(output, `echo test`) {
					t.Error("Expected echo command")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test-command.md")
			outputFile := filepath.Join(tmpDir, "output.md")

			if err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			if err := TransformDocument(inputFile, outputFile, "opencode"); err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			tt.checkOutput(t, string(output))
		})
	}
}

func TestTransformOpenCodeSkill(t *testing.T) {
	tests := []struct {
		name         string
		inputContent string
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name: "minimal skill transforms correctly",
			inputContent: `---
name: test-skill
description: Test skill
---
Skill content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "name: test-skill") {
					t.Error("Expected name in output")
				}
				if !strings.Contains(output, "description: Test skill") {
					t.Error("Expected description in output")
				}
				if !strings.Contains(output, "Skill content") {
					t.Error("Expected content in output")
				}
			},
		},
		{
			name: "full skill with OpenCode fields",
			inputContent: `---
name: test-skill
description: Test skill
license: MIT
compatibility:
  - claude-code
  - opencode
metadata:
  version: 1.0.0
  maintainer: ops
hooks:
  pre-run: validate
  post-run: cleanup
---
Skill content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "license: MIT") {
					t.Error("Expected license field")
				}
				if !strings.Contains(output, "compatibility:") {
					t.Error("Expected compatibility section")
				}
				if !strings.Contains(output, "- claude-code") {
					t.Error("Expected claude-code in compatibility")
				}
				if !strings.Contains(output, "metadata:") {
					t.Error("Expected metadata section")
				}
				if !strings.Contains(output, "version: 1.0.0") {
					t.Error("Expected version in metadata")
				}
				if !strings.Contains(output, "hooks:") {
					t.Error("Expected hooks section")
				}
				if !strings.Contains(output, "pre-run: validate") {
					t.Error("Expected pre-run hook")
				}
			},
		},
		{
			name: "multi-line content preserved",
			inputContent: `---
name: test-skill
description: Test skill
---
## Section 1
Content here

## Section 2
More content`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "## Section 1") {
					t.Error("Expected section headers")
				}
				if !strings.Contains(output, "## Section 2") {
					t.Error("Expected section headers")
				}
			},
		},
		{
			name: "markdown formatting preserved",
			inputContent: `---
name: test-skill
description: Test skill
---
# Header

**Bold** and *italic*

- List item 1
- List item 2`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "# Header") {
					t.Error("Expected markdown header")
				}
				if !strings.Contains(output, "**Bold**") {
					t.Error("Expected bold markdown")
				}
				if !strings.Contains(output, "- List item 1") {
					t.Error("Expected list items")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test-skill.md")
			outputFile := filepath.Join(tmpDir, "output.md")

			if err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			if err := TransformDocument(inputFile, outputFile, "opencode"); err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			tt.checkOutput(t, string(output))
		})
	}
}

func TestTransformOpenCodeMemory(t *testing.T) {
	tests := []struct {
		name         string
		inputContent string
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name: "paths-only converts to @ references",
			inputContent: `---
paths:
  - README.md
  - CONTRIBUTING.md
---
`,
			checkOutput: func(t *testing.T, output string) {
				if strings.Contains(output, "---") {
					t.Error("Expected no YAML frontmatter")
				}
				if !strings.Contains(output, "@README.md") {
					t.Error("Expected @README.md")
				}
				if !strings.Contains(output, "@CONTRIBUTING.md") {
					t.Error("Expected @CONTRIBUTING.md")
				}
			},
		},
		{
			name: "content-only renders as narrative",
			inputContent: `# Project Context

This is the project context.

## More Details

Additional information here.
`,
			checkOutput: func(t *testing.T, output string) {
				if strings.Contains(output, "---") {
					t.Error("Expected no YAML frontmatter")
				}
				if !strings.Contains(output, "# Project Context") {
					t.Error("Expected content")
				}
				if !strings.Contains(output, "This is the project context") {
					t.Error("Expected content")
				}
			},
		},
		{
			name: "both paths and content",
			inputContent: `---
paths:
  - config/mise.toml
---

# Project Rules

Additional context here.
`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "@config/mise.toml") {
					t.Error("Expected @ file reference")
				}
				if !strings.Contains(output, "# Project Rules") {
					t.Error("Expected content")
				}
				if !strings.Contains(output, "Additional context here") {
					t.Error("Expected content")
				}
			},
		},
		{
			name: "multiple paths rendered",
			inputContent: `---
paths:
  - src/*.go
  - test/*.go
  - README.md
---
`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "@src/*.go") {
					t.Error("Expected src pattern")
				}
				if !strings.Contains(output, "@test/*.go") {
					t.Error("Expected test pattern")
				}
				if !strings.Contains(output, "@README.md") {
					t.Error("Expected README")
				}
			},
		},
		{
			name: "nested directory paths",
			inputContent: `---
paths:
  - config/platforms/claude-code/agent.tmpl
  - internal/models/models.go
---
`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "@config/platforms/claude-code/agent.tmpl") {
					t.Error("Expected nested path")
				}
				if !strings.Contains(output, "@internal/models/models.go") {
					t.Error("Expected nested path")
				}
			},
		},
		{
			name: "teaching instructions included",
			inputContent: `---
paths:
  - README.md
---

# Memory

Project info.
`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "@README.md") {
					t.Error("Expected @ reference")
				}
			},
		},
		{
			name: "markdown formatting preserved in content",
			inputContent: `# Project Info

**Important** notes here

## Details

More content.
`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "# Project Info") {
					t.Error("Expected markdown header")
				}
				if !strings.Contains(output, "**Important**") {
					t.Error("Expected bold markdown")
				}
				if !strings.Contains(output, "## Details") {
					t.Error("Expected markdown section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test-memory.md")
			outputFile := filepath.Join(tmpDir, "output.md")

			if err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			if err := TransformDocument(inputFile, outputFile, "opencode"); err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			tt.checkOutput(t, string(output))
		})
	}
}
func TestTransformPermissionModes(t *testing.T) {
	tests := []struct {
		name           string
		permissionMode string
		expectedEdit   string
		expectedBash   string
	}{
		{
			name:           "default mode",
			permissionMode: "default",
			expectedEdit:   "ask",
			expectedBash:   "ask",
		},
		{
			name:           "acceptEdits mode",
			permissionMode: "acceptEdits",
			expectedEdit:   "allow",
			expectedBash:   "ask",
		},
		{
			name:           "dontAsk mode",
			permissionMode: "dontAsk",
			expectedEdit:   "allow",
			expectedBash:   "allow",
		},
		{
			name:           "bypassPermissions mode",
			permissionMode: "bypassPermissions",
			expectedEdit:   "allow",
			expectedBash:   "allow",
		},
		{
			name:           "plan mode",
			permissionMode: "plan",
			expectedEdit:   "deny",
			expectedBash:   "deny",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputContent := fmt.Sprintf(`---
name: test-agent
description: Test agent
permissionMode: %s
---
Agent content`, tt.permissionMode)

			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test-agent.md")
			outputFile := filepath.Join(tmpDir, "output.md")

			if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			if err := TransformDocument(inputFile, outputFile, "opencode"); err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "edit:") {
				t.Error("Expected edit permission in output")
			}
			if !strings.Contains(outputStr, "bash:") {
				t.Error("Expected bash permission in output")
			}
			if !strings.Contains(outputStr, fmt.Sprintf("*: %s", tt.expectedEdit)) {
				t.Errorf("Expected edit permission '*: %s'", tt.expectedEdit)
			}
			if !strings.Contains(outputStr, fmt.Sprintf("*: %s", tt.expectedBash)) {
				t.Errorf("Expected bash permission '*: %s'", tt.expectedBash)
			}
		})
	}
}

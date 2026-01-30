package services

import (
	"os"
	"path/filepath"
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

package library

import (
	"strings"
	"testing"
)

func TestListResources(t *testing.T) {
	lib := &Library{
		Resources: map[string]map[string]Resource{
			"skill": {
				"commit":        {Path: "skills/commit.yaml", Description: "Git commit"},
				"merge-request": {Path: "skills/merge-request.yaml", Description: "MR helper"},
			},
			"agent": {
				"reviewer": {Path: "agents/reviewer.yaml", Description: "Code review"},
			},
		},
	}

	result := ListResources(lib)

	if len(result["skill"]) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(result["skill"]))
	}
	if len(result["agent"]) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(result["agent"]))
	}
}

func TestListResources_SortedByName(t *testing.T) {
	lib := &Library{
		Resources: map[string]map[string]Resource{
			"skill": {
				"zebra":  {Path: "skills/zebra.yaml"},
				"alpha":  {Path: "skills/alpha.yaml"},
				"middle": {Path: "skills/middle.yaml"},
			},
		},
	}

	result := ListResources(lib)
	skills := result["skill"]

	if skills[0].Name != "alpha" {
		t.Errorf("First skill should be 'alpha', got '%s'", skills[0].Name)
	}
	if skills[2].Name != "zebra" {
		t.Errorf("Last skill should be 'zebra', got '%s'", skills[2].Name)
	}
}

func TestListPresets(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {Name: "git-workflow", Description: "Git tools"},
			"code-review":  {Name: "code-review", Description: "Review tools"},
		},
	}

	presets := ListPresets(lib)

	if len(presets) != 2 {
		t.Errorf("Expected 2 presets, got %d", len(presets))
	}

	// Check sorted order
	if presets[0].Name != "code-review" {
		t.Errorf("First preset should be 'code-review', got '%s'", presets[0].Name)
	}
}

func TestFormatResourcesList(t *testing.T) {
	lib := &Library{
		Resources: map[string]map[string]Resource{
			"skill": {
				"commit": {Path: "skills/commit.yaml", Description: "Git commit"},
			},
			"agent": {
				"reviewer": {Path: "agents/reviewer.yaml", Description: "Code review"},
			},
		},
	}

	output := FormatResourcesList(lib)

	if !strings.Contains(output, "Skills:") {
		t.Error("Output should contain 'Skills:' section")
	}
	if !strings.Contains(output, "Agents:") {
		t.Error("Output should contain 'Agents:' section")
	}
	if !strings.Contains(output, "skill/commit") {
		t.Error("Output should contain 'skill/commit'")
	}
	if !strings.Contains(output, "Git commit") {
		t.Error("Output should contain description")
	}
}

func TestFormatResourcesList_Empty(t *testing.T) {
	lib := &Library{
		Resources: map[string]map[string]Resource{},
	}

	output := FormatResourcesList(lib)

	if !strings.Contains(output, "No resources found") {
		t.Errorf("Empty library should show 'No resources found', got: %s", output)
	}
}

func TestFormatPresetsList(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {
				Name:        "git-workflow",
				Description: "Git workflow tools",
				Resources:   []string{"skill/commit", "skill/merge-request"},
			},
		},
	}

	output := FormatPresetsList(lib)

	if !strings.Contains(output, "git-workflow") {
		t.Error("Output should contain preset name")
	}
	if !strings.Contains(output, "Git workflow tools") {
		t.Error("Output should contain description")
	}
	if !strings.Contains(output, "skill/commit") {
		t.Error("Output should contain resources")
	}
}

func TestFormatPresetsList_Empty(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{},
	}

	output := FormatPresetsList(lib)

	if !strings.Contains(output, "No presets found") {
		t.Errorf("Empty library should show 'No presets found', got: %s", output)
	}
}

func TestFormatResourceDetails(t *testing.T) {
	lib := &Library{
		Resources: map[string]map[string]Resource{
			"skill": {
				"commit": {Path: "skills/commit.yaml", Description: "Git commit best practices"},
			},
		},
	}

	output, err := FormatResourceDetails(lib, "skill/commit")
	if err != nil {
		t.Fatalf("FormatResourceDetails() error = %v", err)
	}

	if !strings.Contains(output, "Reference: skill/commit") {
		t.Error("Output should contain reference")
	}
	if !strings.Contains(output, "Path: skills/commit.yaml") {
		t.Error("Output should contain path")
	}
	if !strings.Contains(output, "Description: Git commit best practices") {
		t.Error("Output should contain description")
	}
}

func TestFormatResourceDetails_NotFound(t *testing.T) {
	lib := &Library{
		Resources: map[string]map[string]Resource{},
	}

	_, err := FormatResourceDetails(lib, "skill/nonexistent")
	if err == nil {
		t.Error("FormatResourceDetails() should return error for missing resource")
	}
}

func TestFormatPresetDetails(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {
				Name:        "git-workflow",
				Description: "Git tools",
				Resources:   []string{"skill/commit"},
			},
		},
	}

	output, err := FormatPresetDetails(lib, "git-workflow")
	if err != nil {
		t.Fatalf("FormatPresetDetails() error = %v", err)
	}

	if !strings.Contains(output, "Preset: git-workflow") {
		t.Error("Output should contain preset name")
	}
	if !strings.Contains(output, "Resources:") {
		t.Error("Output should contain resources section")
	}
}

func TestFormatPresetDetails_NotFound(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{},
	}

	_, err := FormatPresetDetails(lib, "nonexistent")
	if err == nil {
		t.Error("FormatPresetDetails() should return error for missing preset")
	}
}

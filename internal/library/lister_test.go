package library

import (
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

	if presets[0].Name != "code-review" {
		t.Errorf("First preset should be 'code-review', got '%s'", presets[0].Name)
	}
}

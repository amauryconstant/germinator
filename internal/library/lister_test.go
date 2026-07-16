package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Len(t, result["skill"], 2, "skills count")
	assert.Len(t, result["agent"], 1, "agents count")
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

	assert.Equal(t, "alpha", skills[0].Name, "first skill")
	assert.Equal(t, "zebra", skills[2].Name, "last skill")
}

func TestListPresets(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {Name: "git-workflow", Description: "Git tools"},
			"code-review":  {Name: "code-review", Description: "Review tools"},
		},
	}

	presets := ListPresets(lib)

	assert.Len(t, presets, 2, "presets count")
	assert.Equal(t, "code-review", presets[0].Name, "first preset")
}

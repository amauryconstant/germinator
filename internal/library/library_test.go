package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		rt       ResourceType
		expected bool
	}{
		{"skill is valid", ResourceTypeSkill, true},
		{"agent is valid", ResourceTypeAgent, true},
		{"command is valid", ResourceTypeCommand, true},
		{"memory is valid", ResourceTypeMemory, true},
		{"invalid type", ResourceType("invalid"), false},
		{"empty type", ResourceType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.rt.IsValid())
		})
	}
}

func TestResource_Validate(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		wantErr  bool
	}{
		{
			name:     "valid resource",
			resource: Resource{Path: "skills/commit.yaml", Description: "Test"},
			wantErr:  false,
		},
		{
			name:     "empty path",
			resource: Resource{Path: "", Description: "Test"},
			wantErr:  true,
		},
		{
			name:     "whitespace only path",
			resource: Resource{Path: "   ", Description: "Test"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resource.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPreset_Validate(t *testing.T) {
	tests := []struct {
		name    string
		preset  Preset
		wantErr bool
	}{
		{
			name: "valid preset",
			preset: Preset{
				Name:        "git-workflow",
				Description: "Git tools",
				Resources:   []string{"skill/commit", "skill/merge-request"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			preset: Preset{
				Name:      "",
				Resources: []string{"skill/commit"},
			},
			wantErr: true,
		},
		{
			name: "whitespace name",
			preset: Preset{
				Name:      "   ",
				Resources: []string{"skill/commit"},
			},
			wantErr: true,
		},
		{
			name: "no resources",
			preset: Preset{
				Name:      "empty-preset",
				Resources: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid resource reference",
			preset: Preset{
				Name:      "bad-preset",
				Resources: []string{"invalid-format"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.preset.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParseRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		wantType string
		wantName string
		wantErr  bool
	}{
		{"valid reference", "skill/commit", "skill", "commit", false},
		{"valid agent reference", "agent/reviewer", "agent", "reviewer", false},
		{"valid command reference", "command/test", "command", "test", false},
		{"valid memory reference", "memory/context", "memory", "context", false},
		{"no slash", "skillcommit", "", "", true},
		{"multiple slashes", "skill/name/extra", "", "", true},
		{"empty type", "/name", "", "", true},
		{"empty name", "skill/", "", "", true},
		{"empty string", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotName, err := ParseRef(tt.ref)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, gotType)
			assert.Equal(t, tt.wantName, gotName)
		})
	}
}

func TestFormatRef(t *testing.T) {
	assert.Equal(t, "skill/commit", FormatRef("skill", "commit"))
}

package library

import (
	"testing"
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
			if got := tt.rt.IsValid(); got != tt.expected {
				t.Errorf("ResourceType.IsValid() = %v, want %v", got, tt.expected)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Resource.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Preset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotType != tt.wantType {
					t.Errorf("ParseRef() type = %v, want %v", gotType, tt.wantType)
				}
				if gotName != tt.wantName {
					t.Errorf("ParseRef() name = %v, want %v", gotName, tt.wantName)
				}
			}
		})
	}
}

func TestFormatRef(t *testing.T) {
	if got := FormatRef("skill", "commit"); got != "skill/commit" {
		t.Errorf("FormatRef() = %v, want skill/commit", got)
	}
}

package library

import (
	"path/filepath"
	"testing"
)

func TestResolveResource(t *testing.T) {
	lib := &Library{
		RootPath: "/test/library",
		Resources: map[string]map[string]Resource{
			"skill": {
				"commit": {Path: "skills/commit.yaml", Description: "Git commit"},
			},
			"agent": {
				"reviewer": {Path: "agents/reviewer.yaml", Description: "Code review"},
			},
		},
	}

	tests := []struct {
		name     string
		ref      string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "resolve skill",
			ref:      "skill/commit",
			wantPath: filepath.Join("/test/library", "skills/commit.yaml"),
			wantErr:  false,
		},
		{
			name:     "resolve agent",
			ref:      "agent/reviewer",
			wantPath: filepath.Join("/test/library", "agents/reviewer.yaml"),
			wantErr:  false,
		},
		{
			name:    "resource not found",
			ref:     "skill/nonexistent",
			wantErr: true,
		},
		{
			name:    "type not found",
			ref:     "nonexistent/test",
			wantErr: true,
		},
		{
			name:    "invalid ref format",
			ref:     "invalidformat",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := ResolveResource(lib, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotPath != tt.wantPath {
				t.Errorf("ResolveResource() path = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestResolveResources(t *testing.T) {
	lib := &Library{
		RootPath: "/test/library",
		Resources: map[string]map[string]Resource{
			"skill": {
				"commit":        {Path: "skills/commit.yaml"},
				"merge-request": {Path: "skills/merge-request.yaml"},
			},
		},
	}

	refs := []string{"skill/commit", "skill/merge-request"}
	paths, err := ResolveResources(lib, refs)
	if err != nil {
		t.Fatalf("ResolveResources() error = %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("ResolveResources() returned %d paths, want 2", len(paths))
	}
}

func TestResolveResources_FailFast(t *testing.T) {
	lib := &Library{
		RootPath: "/test/library",
		Resources: map[string]map[string]Resource{
			"skill": {
				"commit": {Path: "skills/commit.yaml"},
			},
		},
	}

	// Second resource doesn't exist
	refs := []string{"skill/commit", "skill/nonexistent"}
	_, err := ResolveResources(lib, refs)
	if err == nil {
		t.Error("ResolveResources() expected error for missing resource")
	}
}

func TestResolvePreset(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {
				Name:        "git-workflow",
				Description: "Git tools",
				Resources:   []string{"skill/commit", "skill/merge-request"},
			},
		},
	}

	refs, err := ResolvePreset(lib, "git-workflow")
	if err != nil {
		t.Fatalf("ResolvePreset() error = %v", err)
	}

	if len(refs) != 2 {
		t.Errorf("ResolvePreset() returned %d refs, want 2", len(refs))
	}
}

func TestResolvePreset_NotFound(t *testing.T) {
	lib := &Library{Presets: map[string]Preset{}}

	_, err := ResolvePreset(lib, "nonexistent")
	if err == nil {
		t.Error("ResolvePreset() expected error for missing preset")
	}
}

func TestGetOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		typ       string
		resName   string
		platform  string
		outputDir string
		wantPath  string
		wantErr   bool
	}{
		{
			name:      "skill to opencode",
			typ:       "skill",
			resName:   "commit",
			platform:  "opencode",
			outputDir: ".",
			wantPath:  ".opencode/skills/commit/SKILL.md",
			wantErr:   false,
		},
		{
			name:      "skill to claude-code",
			typ:       "skill",
			resName:   "commit",
			platform:  "claude-code",
			outputDir: ".",
			wantPath:  ".claude/skills/commit/SKILL.md",
			wantErr:   false,
		},
		{
			name:      "agent to opencode",
			typ:       "agent",
			resName:   "reviewer",
			platform:  "opencode",
			outputDir: ".",
			wantPath:  ".opencode/agents/reviewer.md",
			wantErr:   false,
		},
		{
			name:      "command to opencode",
			typ:       "command",
			resName:   "test",
			platform:  "opencode",
			outputDir: ".",
			wantPath:  ".opencode/commands/test.md",
			wantErr:   false,
		},
		{
			name:      "memory to opencode",
			typ:       "memory",
			resName:   "context",
			platform:  "opencode",
			outputDir: ".",
			wantPath:  ".opencode/memory/context.md",
			wantErr:   false,
		},
		{
			name:      "agent to claude-code",
			typ:       "agent",
			resName:   "reviewer",
			platform:  "claude-code",
			outputDir: ".",
			wantPath:  ".claude/agents/reviewer.md",
			wantErr:   false,
		},
		{
			name:      "custom output dir",
			typ:       "skill",
			resName:   "commit",
			platform:  "opencode",
			outputDir: "/project",
			wantPath:  "/project/.opencode/skills/commit/SKILL.md",
			wantErr:   false,
		},
		{
			name:      "invalid resource type",
			typ:       "invalid",
			resName:   "test",
			platform:  "opencode",
			outputDir: ".",
			wantErr:   true,
		},
		{
			name:      "invalid platform",
			typ:       "skill",
			resName:   "test",
			platform:  "invalid",
			outputDir: ".",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := GetOutputPath(tt.typ, tt.resName, tt.platform, tt.outputDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotPath != tt.wantPath {
				t.Errorf("GetOutputPath() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestIsValidPlatform(t *testing.T) {
	if !IsValidPlatform("opencode") {
		t.Error("opencode should be valid platform")
	}
	if !IsValidPlatform("claude-code") {
		t.Error("claude-code should be valid platform")
	}
	if IsValidPlatform("invalid") {
		t.Error("invalid should not be valid platform")
	}
}

func TestValidPlatforms(t *testing.T) {
	platforms := ValidPlatforms()
	if len(platforms) != 2 {
		t.Errorf("ValidPlatforms() returned %d platforms, want 2", len(platforms))
	}
}

func TestValidateRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{"valid skill ref", "skill/commit", false},
		{"valid agent ref", "agent/reviewer", false},
		{"valid command ref", "command/test", false},
		{"valid memory ref", "memory/context", false},
		{"invalid format", "invalidformat", true},
		{"invalid type", "invalid/name", true},
		{"empty name", "skill/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRef() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package library

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gerrors "gitlab.com/amoconst/germinator/internal/core"
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

func TestLibraryResolvePreset(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {
				Name:        "git-workflow",
				Description: "Git tools",
				Resources:   []string{"skill/commit", "skill/merge-request"},
			},
		},
	}

	refs, err := lib.ResolvePreset(context.Background(), "git-workflow")
	if err != nil {
		t.Fatalf("ResolvePreset() error = %v", err)
	}
	if len(refs) != 2 {
		t.Errorf("ResolvePreset() returned %d refs, want 2", len(refs))
	}
	if refs[0] != "skill/commit" || refs[1] != "skill/merge-request" {
		t.Errorf("ResolvePreset() refs = %v, want [skill/commit skill/merge-request]", refs)
	}
}

func TestLibraryResolvePreset_NotFound(t *testing.T) {
	lib := &Library{Presets: map[string]Preset{}}

	_, err := lib.ResolvePreset(context.Background(), "ghost")
	if err == nil {
		t.Fatal("ResolvePreset() expected error for missing preset")
	}
	var cfgErr *gerrors.ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("ResolvePreset() expected *core.ConfigError, got %T (%v)", err, err)
	}
}

func TestLibraryResolvePreset_EmptyResources(t *testing.T) {
	// A malformed preset with zero resources would be caught by
	// (*Preset).Validate() at load time; the method itself just
	// returns whatever the Preset.Resources slice contains.
	lib := &Library{
		Presets: map[string]Preset{
			"empty": {Name: "empty", Resources: nil},
		},
	}

	refs, err := lib.ResolvePreset(context.Background(), "empty")
	if err != nil {
		t.Fatalf("ResolvePreset() error = %v", err)
	}
	if len(refs) != 0 {
		t.Errorf("ResolvePreset() returned %d refs, want 0", len(refs))
	}
}

// TestResolvePreset_AcceptAndMayIgnore documents the spec contract from
// library-library-resolution "Cancellation during resolution": ctx is
// forwarded for symmetry with other I/O adapter methods, but the current
// implementation is a pure in-memory map lookup so ctx.Err() is not
// checked. A pre-cancelled ctx MUST NOT cause ResolvePreset to return
// an error today; future I/O additions must respect ctx.
//
// See cli-framework/spec.md "accept-and-may-ignore" pattern.
func TestResolvePreset_AcceptAndMayIgnore(t *testing.T) {
	lib := &Library{Presets: map[string]Preset{
		"alpha": {Name: "alpha", Resources: []string{"skill/a"}},
	}}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	refs, err := lib.ResolvePreset(ctx, "alpha")
	require.NoError(t, err)
	assert.Equal(t, []string{"skill/a"}, refs)
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

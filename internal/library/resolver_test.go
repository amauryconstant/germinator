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
			if tt.wantErr {
				require.Error(t, err)
				if tt.name != "invalid ref format" {
					var nf *gerrors.NotFoundError
					require.True(t, errors.As(err, &nf),
						"not-found path MUST surface as *core.NotFoundError; got %T (%v)", err, err)
					assert.Equal(t, "resource", nf.Entity,
						"resource-lookup misses MUST report entity 'resource'")
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, gotPath)
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
	require.NoError(t, err)
	assert.Len(t, paths, 2)
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
	require.Error(t, err)
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
	require.NoError(t, err)
	assert.Equal(t, []string{"skill/commit", "skill/merge-request"}, refs)
}

func TestLibraryResolvePreset_NotFound(t *testing.T) {
	lib := &Library{Presets: map[string]Preset{}}

	_, err := lib.ResolvePreset(context.Background(), "ghost")
	require.Error(t, err)
	var nf *gerrors.NotFoundError
	require.True(t, errors.As(err, &nf),
		"ResolvePreset() expected *core.NotFoundError; got %T (%v)", err, err)
	assert.Equal(t, "preset", nf.Entity)
	assert.Equal(t, "ghost", nf.Key)
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
	require.NoError(t, err)
	assert.Empty(t, refs)
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
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, gotPath)
		})
	}
}

func TestIsValidPlatform(t *testing.T) {
	assert.True(t, IsValidPlatform("opencode"), "opencode should be valid platform")
	assert.True(t, IsValidPlatform("claude-code"), "claude-code should be valid platform")
	assert.False(t, IsValidPlatform("invalid"), "invalid should not be valid platform")
}

func TestValidPlatforms(t *testing.T) {
	platforms := ValidPlatforms()
	assert.Len(t, platforms, 2)
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
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

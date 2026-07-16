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

// TestResolveResourceEntry covers ResolveResourceEntry (resolver.go:50).
// Returns the *Resource entry on hit, *core.NotFoundError on miss,
// or a ParseRef error on malformed refs.
func TestResolveResourceEntry(t *testing.T) {
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
		name         string
		ref          string
		wantPath     string
		wantDesc     string
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:     "success: skill entry",
			ref:      "skill/commit",
			wantPath: "skills/commit.yaml",
			wantDesc: "Git commit",
		},
		{
			name:     "success: agent entry",
			ref:      "agent/reviewer",
			wantPath: "agents/reviewer.yaml",
			wantDesc: "Code review",
		},
		{
			name:         "error: name not found",
			ref:          "skill/nonexistent",
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "error: type not found",
			ref:          "nonexistent/test",
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:    "error: invalid ref format",
			ref:     "no-slash",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveResourceEntry(lib, tt.ref)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantNotFound {
					var nf *gerrors.NotFoundError
					require.True(t, errors.As(err, &nf),
						"expected *core.NotFoundError, got %T (%v)", err, err)
					assert.Equal(t, "resource", nf.Entity)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantPath, got.Path)
			assert.Equal(t, tt.wantDesc, got.Description)
		})
	}
}

// TestResolvePresetEntry covers ResolvePresetEntry (resolver.go:118).
// Returns the *Preset entry on hit, *core.NotFoundError on miss.
func TestResolvePresetEntry(t *testing.T) {
	lib := &Library{
		Presets: map[string]Preset{
			"git-workflow": {
				Name:        "git-workflow",
				Description: "Git tools",
				Resources:   []string{"skill/commit", "skill/merge-request"},
			},
		},
	}

	t.Run("success: known preset", func(t *testing.T) {
		got, err := ResolvePresetEntry(lib, "git-workflow")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "git-workflow", got.Name)
		assert.Equal(t, "Git tools", got.Description)
		assert.Equal(t, []string{"skill/commit", "skill/merge-request"}, got.Resources)
	})

	t.Run("error: preset not found", func(t *testing.T) {
		got, err := ResolvePresetEntry(lib, "ghost")
		require.Error(t, err)
		assert.Nil(t, got)
		var nf *gerrors.NotFoundError
		require.True(t, errors.As(err, &nf),
			"expected *core.NotFoundError, got %T (%v)", err, err)
		assert.Equal(t, "preset", nf.Entity)
		assert.Equal(t, "ghost", nf.Key)
	})
}

// TestGetOutputPaths covers GetOutputPaths (plural — resolver.go:185),
// the map-returning bulk variant of GetOutputPath (singular). The
// function is fail-fast: it returns nil + error on the first invalid
// ref; subsequent valid refs are not processed.
func TestGetOutputPaths(t *testing.T) {
	tests := []struct {
		name       string
		refs       []string
		platform   string
		outputDir  string
		wantMap    map[string]string
		wantErr    bool
		wantConfig bool
	}{
		{
			name:      "all four resource types to opencode",
			refs:      []string{"skill/commit", "agent/reviewer", "command/build", "memory/context"},
			platform:  "opencode",
			outputDir: ".",
			wantMap: map[string]string{
				"skill/commit":   ".opencode/skills/commit/SKILL.md",
				"agent/reviewer": ".opencode/agents/reviewer.md",
				"command/build":  ".opencode/commands/build.md",
				"memory/context": ".opencode/memory/context.md",
			},
		},
		{
			name:      "use-subdirectory branch (skill)",
			refs:      []string{"skill/commit"},
			platform:  "claude-code",
			outputDir: ".",
			wantMap: map[string]string{
				"skill/commit": ".claude/skills/commit/SKILL.md",
			},
		},
		{
			name:       "invalid platform returns *core.ConfigError",
			refs:       []string{"skill/commit"},
			platform:   "unknown-platform",
			outputDir:  ".",
			wantErr:    true,
			wantConfig: true,
		},
		{
			name:       "invalid resource type returns *core.ConfigError",
			refs:       []string{"unknown/x"},
			platform:   "opencode",
			outputDir:  ".",
			wantErr:    true,
			wantConfig: true,
		},
		{
			name:      "fail-fast on mixed refs (second invalid)",
			refs:      []string{"skill/commit", "unknown/x"},
			platform:  "opencode",
			outputDir: ".",
			wantErr:   true,
			// fail-fast semantics: returns nil map on first failure
		},
		{
			name:      "invalid ref format returns ParseRef error",
			refs:      []string{"no-slash"},
			platform:  "opencode",
			outputDir: ".",
			wantErr:   true,
			// returns the ParseRef error, not *core.ConfigError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOutputPaths(nil, tt.refs, tt.platform, tt.outputDir)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantConfig {
					var ce *gerrors.ConfigError
					require.True(t, errors.As(err, &ce),
						"expected error to wrap *core.ConfigError, got %T (%v)", err, err)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMap, got)
		})
	}
}

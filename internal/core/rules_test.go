package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePlatform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		platform  string
		wantError bool
	}{
		{name: "claude-code", platform: "claude-code", wantError: false},
		{name: "opencode", platform: "opencode", wantError: false},
		{name: "empty", platform: "", wantError: true},
		{name: "unknown", platform: "vscode", wantError: true},
		{name: "uppercase", platform: "Claude-Code", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePlatform(tt.platform)
			if tt.wantError {
				var ve *ValidationError
				require.ErrorAs(t, err, &ve)
				assert.Equal(t, tt.platform, ve.Value())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResolveOutputPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		docType  string
		resName  string
		platform string
		want     string
	}{
		{"skill opencode", "skill", "commit", "opencode", ".opencode/skills/commit/SKILL.md"},
		{"skill claude-code", "skill", "commit", "claude-code", ".claude/skills/commit/SKILL.md"},
		{"agent opencode", "agent", "reviewer", "opencode", ".opencode/agents/reviewer.md"},
		{"agent claude-code", "agent", "reviewer", "claude-code", ".claude/agents/reviewer.md"},
		{"command opencode", "command", "build", "opencode", ".opencode/commands/build.md"},
		{"command claude-code", "command", "build", "claude-code", ".claude/commands/build.md"},
		{"memory opencode", "memory", "context", "opencode", ".opencode/memory/context.md"},
		{"memory claude-code", "memory", "context", "claude-code", ".claude/memory/context.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveOutputPath(tt.docType, tt.resName, tt.platform)
			assert.Equal(t, tt.want, got)
		})
	}
}

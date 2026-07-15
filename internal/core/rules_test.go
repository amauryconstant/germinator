package core

import (
	"slices"
	"strings"
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

func TestCanInstallResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		ref             string
		wantErr         bool
		wantMsgContains string
		wantSuggestions bool
	}{
		{
			name: "valid skill ref",
			ref:  "skill/commit",
		},
		{
			name: "valid agent ref",
			ref:  "agent/reviewer",
		},
		{
			name: "valid command ref",
			ref:  "command/build",
		},
		{
			name: "valid memory ref",
			ref:  "memory/project",
		},
		{
			name: "name with hyphens",
			ref:  "skill/git-commit",
		},
		{
			name:            "wrong type",
			ref:             "skills/commit",
			wantErr:         true,
			wantMsgContains: "ref type must be one of",
			wantSuggestions: true,
		},
		{
			name:            "empty name",
			ref:             "skill/",
			wantErr:         true,
			wantMsgContains: "ref name must be non-empty",
		},
		{
			name:            "no slash",
			ref:             "skill",
			wantErr:         true,
			wantMsgContains: "ref must be type/name",
		},
		{
			name:            "empty ref",
			ref:             "",
			wantErr:         true,
			wantMsgContains: "ref must be type/name",
		},
		{
			name:            "empty type",
			ref:             "/commit",
			wantErr:         true,
			wantMsgContains: "ref must be type/name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CanInstallResource(tt.ref)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			var ve *ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Contains(t, ve.Message(), tt.wantMsgContains)
			assert.Equal(t, "ref", ve.Field())
			assert.Equal(t, tt.ref, ve.Value())
			if tt.wantSuggestions {
				assert.NotEmpty(t, ve.Suggestions())
			}
		})
	}
}

func TestValidResourceTypesContainsAllExpected(t *testing.T) {
	t.Parallel()

	// Regression guard: if a new resource type is added (e.g. "hook"),
	// both validResourceTypes and the AGENTS.md documentation must move
	// in lockstep. Spec at library-library-resource-import/spec.md:23
	// pins the literal list {skill, agent, command, memory}.
	expected := []string{"skill", "agent", "command", "memory"}
	for _, et := range expected {
		assert.True(t, slices.Contains(validResourceTypes, et),
			"validResourceTypes missing %q", et)
	}
	// And specifically: no stale entries like "skills" (plural).
	assert.False(t, slices.Contains(validResourceTypes, "skills"),
		"validResourceTypes must not contain the plural 'skills'")
	assert.False(t, slices.Contains(validResourceTypes, ""),
		"validResourceTypes must not contain the empty string")
	assert.Len(t, validResourceTypes, len(expected),
		"validResourceTypes should contain exactly 4 entries")
}

func TestValidateDocumentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		docType         string
		wantErr         bool
		wantMsgContains string
		wantSuggestions bool
	}{
		{
			name:    "skill",
			docType: "skill",
		},
		{
			name:    "agent",
			docType: "agent",
		},
		{
			name:    "command",
			docType: "command",
		},
		{
			name:    "memory",
			docType: "memory",
		},
		{
			name:            "plural form (skills)",
			docType:         "skills",
			wantErr:         true,
			wantMsgContains: "type must be one of",
			wantSuggestions: true,
		},
		{
			name:            "unknown type",
			docType:         "bot",
			wantErr:         true,
			wantMsgContains: "type must be one of",
			wantSuggestions: true,
		},
		{
			name:            "empty string",
			docType:         "",
			wantErr:         true,
			wantMsgContains: "type must be one of",
			wantSuggestions: true,
		},
		{
			name:            "uppercase",
			docType:         "Agent",
			wantErr:         true,
			wantMsgContains: "type must be one of",
			wantSuggestions: true,
		},
		{
			name:            "type with trailing slash",
			docType:         "skill/",
			wantErr:         true,
			wantMsgContains: "type must be one of",
			wantSuggestions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDocumentType(tt.docType)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			var ve *ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Contains(t, ve.Message(), tt.wantMsgContains)
			assert.Equal(t, "type", ve.Field(),
				"ValidationError.Field() must be 'type' so renderers identify the offending flag")
			assert.Equal(t, tt.docType, ve.Value())
			if tt.wantSuggestions {
				assert.NotEmpty(t, ve.Suggestions(),
					"ValidateDocumentType must attach suggestions listing the canonical types")
				// Spot-check that the suggestion text is actionable.
				suggestions := ve.Suggestions()
				found := false
				for _, s := range suggestions {
					if strings.Contains(s, "skill") &&
						strings.Contains(s, "agent") &&
						strings.Contains(s, "command") &&
						strings.Contains(s, "memory") {
						found = true
						break
					}
				}
				assert.True(t, found,
					"at least one suggestion must enumerate all 4 canonical types, got %v",
					suggestions)
			}
		})
	}
}

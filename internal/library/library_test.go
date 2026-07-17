package library

import (
	"testing"

	"go.uber.org/goleak"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain is the package-level test entry point. goleak.VerifyTestMain
// wraps m.Run() and verifies that no goroutines remain after the test
// suite completes — guarding against leaks from the errgroup.SetLimit
// concurrent orphan scan in adder.go:scanDirectory / scanLevel.
//
// Per `golang-testing` Best Practice 6 (packages with goroutines
// SHOULD use goleak.VerifyTestMain), introduced in harden-tests-and-
// coverage Phase 7 alongside the Phase 5 t.Parallel() additions.
//
// Note: do not append `os.Exit(m.Run())` after goleak.VerifyTestMain;
// VerifyTestMain already wraps m.Run() and exits the process on
// completion. Adding m.Run() a second time would run the test suite
// twice.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

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

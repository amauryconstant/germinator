package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name     string
		variable *string
		desc     string
	}{
		{
			name:     "Version variable exists",
			variable: &Version,
			desc:     "Version should be set at build time via ldflags",
		},
		{
			name:     "Commit variable exists",
			variable: &Commit,
			desc:     "Commit should be set at build time via ldflags",
		},
		{
			name:     "Date variable exists",
			variable: &Date,
			desc:     "Date should be set at build time via ldflags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.variable == nil {
				t.Errorf("%s: variable is nil", tt.desc)
			}
		})
	}
}

func TestVersionDefaultValue(t *testing.T) {
	if Version == "" {
		t.Errorf("Version should have a default value")
	}

	if Version == "dev" || Version == "" {
		t.Log("Version is using default development value")
	}
}

func TestCommitDefaultValue(t *testing.T) {
	if Commit == "" {
		t.Log("Commit is empty - this is expected in development builds")
	}
}

func TestDateDefaultValue(t *testing.T) {
	if Date == "" {
		t.Log("Date is empty - this is expected in development builds")
	}
}

func TestVersionType(t *testing.T) {
	if len(Version) > 0 {
		t.Logf("Version type: string, value: %s", Version)
	}
}

func TestCommitType(t *testing.T) {
	if len(Commit) > 0 {
		t.Logf("Commit type: string, value: %s", Commit)
	}
}

func TestDateType(t *testing.T) {
	if len(Date) > 0 {
		t.Logf("Date type: string, value: %s", Date)
	}
}

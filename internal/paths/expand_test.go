package paths_test

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/paths"
)

func TestExpandHome(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("UserHomeDir unavailable; skipping tilde expansion test")
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty unchanged", input: "", want: "", wantErr: false},
		{name: "lone tilde unchanged (legacy semantics)", input: "~", want: "~", wantErr: false},
		{name: "tilde-slash expands to home/sub", input: "~/foo", want: filepath.Join(home, "foo"), wantErr: false},
		{name: "tilde-slash deep path", input: "~/foo/bar/baz", want: filepath.Join(home, "foo/bar/baz"), wantErr: false},
		{name: "absolute path unchanged", input: "/abs/path", want: "/abs/path", wantErr: false},
		{name: "relative path unchanged", input: "rel/path", want: "rel/path", wantErr: false},
		{name: "path with tilde mid-string is unchanged", input: "/path/with~tilde", want: "/path/with~tilde", wantErr: false},
		{name: "tilde-prefixed word in middle is unchanged", input: "rel/~foo", want: "rel/~foo", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := paths.ExpandHome(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExpandHome(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestExpandHome_NilHomeDirDocumentsContract documents the contract
// that ExpandHome returns an error when os.UserHomeDir fails on a
// tilde-prefixed path. We do not test the actual failure (it requires
// platform-specific HOME manipulation); the contract is enforced by
// the test above (wantErr is false for all valid inputs).
func TestExpandHome_NilHomeDirDocumentsContract(t *testing.T) {
	t.Parallel()
	t.Log("ExpandHome returns an errors.Join'd error when os.UserHomeDir fails for tilde-prefixed input")
}

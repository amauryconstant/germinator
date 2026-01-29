package core

import (
	"strings"
	"testing"
)

func TestTransformPermissionMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{
			name: "default mode",
			mode: "default",
			want: "  edit:\n    *: ask\n  bash:\n    *: ask\n",
		},
		{
			name: "acceptEdits mode",
			mode: "acceptEdits",
			want: "  edit:\n    *: allow\n  bash:\n    *: ask\n",
		},
		{
			name: "dontAsk mode",
			mode: "dontAsk",
			want: "  edit:\n    *: allow\n  bash:\n    *: allow\n",
		},
		{
			name: "bypassPermissions mode",
			mode: "bypassPermissions",
			want: "  edit:\n    *: allow\n  bash:\n    *: allow\n",
		},
		{
			name: "plan mode",
			mode: "plan",
			want: "  edit:\n    *: deny\n  bash:\n    *: deny\n",
		},
		{
			name: "unknown mode",
			mode: "unknown",
			want: "",
		},
		{
			name: "empty mode",
			mode: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformPermissionMode(tt.mode)
			if tt.want == "" {
				if got != "" {
					t.Errorf("transformPermissionMode(%q) = %v, want empty string", tt.mode, got)
				}
				return
			}

			expectedLines := strings.Split(tt.want, "\n")
			for _, line := range expectedLines {
				if line == "" {
					continue
				}
				if !strings.Contains(got, line) {
					t.Errorf("transformPermissionMode(%q) missing expected line %q\nGot:\n%s", tt.mode, line, got)
				}
			}

			expectedTools := []string{"edit", "bash"}
			for _, tool := range expectedTools {
				if !strings.Contains(got, tool+":") {
					t.Errorf("transformPermissionMode(%q) missing tool %q\nGot:\n%s", tt.mode, tool, got)
				}
			}
		})
	}
}

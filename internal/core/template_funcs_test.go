package core

import (
	"reflect"
	"testing"
)

func TestTransformPermissionMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want map[string]interface{}
	}{
		{
			name: "default mode",
			mode: "default",
			want: map[string]interface{}{
				"edit": map[string]string{"*": "ask"},
				"bash": map[string]string{"*": "ask"},
			},
		},
		{
			name: "acceptEdits mode",
			mode: "acceptEdits",
			want: map[string]interface{}{
				"edit": map[string]string{"*": "allow"},
				"bash": map[string]string{"*": "ask"},
			},
		},
		{
			name: "dontAsk mode",
			mode: "dontAsk",
			want: map[string]interface{}{
				"edit": map[string]string{"*": "allow"},
				"bash": map[string]string{"*": "allow"},
			},
		},
		{
			name: "bypassPermissions mode",
			mode: "bypassPermissions",
			want: map[string]interface{}{
				"edit": map[string]string{"*": "allow"},
				"bash": map[string]string{"*": "allow"},
			},
		},
		{
			name: "plan mode",
			mode: "plan",
			want: map[string]interface{}{
				"edit": map[string]string{"*": "deny"},
				"bash": map[string]string{"*": "deny"},
			},
		},
		{
			name: "unknown mode",
			mode: "unknown",
			want: nil,
		},
		{
			name: "empty mode",
			mode: "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformPermissionMode(tt.mode)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transformPermissionMode(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

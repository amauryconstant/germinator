package config

import (
	"errors"
	"os"
	"testing"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Library != "~/.config/germinator/library" {
		t.Errorf("DefaultConfig().Library = %q, want %q", cfg.Library, "~/.config/germinator/library")
	}

	if cfg.Platform != "" {
		t.Errorf("DefaultConfig().Platform = %q, want empty string", cfg.Platform)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		platform    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "empty platform is valid",
			platform: "",
			wantErr:  false,
		},
		{
			name:     "claude-code platform is valid",
			platform: models.PlatformClaudeCode,
			wantErr:  false,
		},
		{
			name:     "opencode platform is valid",
			platform: models.PlatformOpenCode,
			wantErr:  false,
		},
		{
			name:        "invalid platform returns error",
			platform:    "invalid-platform",
			wantErr:     true,
			errContains: "unknown platform",
		},
		{
			name:        "random platform returns error",
			platform:    "foobar",
			wantErr:     true,
			errContains: "available: claude-code, opencode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Platform: tt.platform}
			err := cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Config.Validate() expected error, got nil")
				}

				var configErr *gerrors.ConfigError
				if !errors.As(err, &configErr) {
					t.Fatalf("Config.Validate() error = %T, want *ConfigError", err)
				}

				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Config.Validate() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("Config.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExpandTilde(t *testing.T) {
	home := getHomeDir(t)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "empty path",
			path: "",
			want: "",
		},
		{
			name: "tilde expansion",
			path: "~/some/path",
			want: home + "/some/path",
		},
		{
			name: "absolute path unchanged",
			path: "/absolute/path",
			want: "/absolute/path",
		},
		{
			name: "relative path unchanged",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "lone tilde unchanged",
			path: "~",
			want: "~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandTilde(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expandTilde() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expandTilde() unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestConfigExpandPaths(t *testing.T) {
	home := getHomeDir(t)

	cfg := &Config{
		Library:  "~/my-library",
		Platform: "opencode",
	}

	if err := cfg.ExpandPaths(); err != nil {
		t.Fatalf("Config.ExpandPaths() error: %v", err)
	}

	wantLibrary := home + "/my-library"
	if cfg.Library != wantLibrary {
		t.Errorf("Config.Library after ExpandPaths() = %q, want %q", cfg.Library, wantLibrary)
	}
}

func getHomeDir(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}
	return home
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

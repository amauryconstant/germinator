package config

import (
	"errors"
	"os"
	"testing"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Library != "" {
		t.Errorf("DefaultConfig().Library = %q, want %q", cfg.Library, "")
	}

	if cfg.PlatformDefault != "" {
		t.Errorf("DefaultConfig().PlatformDefault = %q, want empty string", cfg.PlatformDefault)
	}

	if cfg.Debug != false {
		t.Errorf("DefaultConfig().Debug = %v, want false", cfg.Debug)
	}

	// Test completion config defaults
	if cfg.Completion.Timeout != "500ms" {
		t.Errorf("DefaultConfig().Completion.Timeout = %q, want %q", cfg.Completion.Timeout, "500ms")
	}

	if cfg.Completion.CacheTTL != "5s" {
		t.Errorf("DefaultConfig().Completion.CacheTTL = %q, want %q", cfg.Completion.CacheTTL, "5s")
	}
}

func TestCompletionConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "timeout default is 500ms",
			got:      cfg.Completion.Timeout,
			expected: "500ms",
		},
		{
			name:     "cache_ttl default is 5s",
			got:      cfg.Completion.CacheTTL,
			expected: "5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name             string
		platform         string
		completion       CompletionConfig
		wantErr          bool
		errContains      string
		errFieldContains string
	}{
		{
			name:     "empty platform is valid",
			platform: "",
			wantErr:  false,
		},
		{
			name:     "claude-code platform is valid",
			platform: gerrors.PlatformClaudeCode,
			wantErr:  false,
		},
		{
			name:     "opencode platform is valid",
			platform: gerrors.PlatformOpenCode,
			wantErr:  false,
		},
		{
			name:        "invalid platform returns error",
			platform:    "invalid-platform",
			wantErr:     true,
			errContains: "unknown platform",
		},
		{
			name:        "random platform returns error with suggestion",
			platform:    "foobar",
			wantErr:     true,
			errContains: "💡",
		},
		{
			name:       "empty timeout is valid (helpers fall back)",
			platform:   "",
			completion: CompletionConfig{Timeout: "", CacheTTL: ""},
			wantErr:    false,
		},
		{
			name:       "valid timeout is accepted",
			platform:   "",
			completion: CompletionConfig{Timeout: "2s", CacheTTL: ""},
			wantErr:    false,
		},
		{
			name:       "valid cache_ttl is accepted",
			platform:   "",
			completion: CompletionConfig{Timeout: "", CacheTTL: "10s"},
			wantErr:    false,
		},
		{
			name:             "invalid timeout returns ConfigError",
			platform:         "",
			completion:       CompletionConfig{Timeout: "junk", CacheTTL: ""},
			wantErr:          true,
			errContains:      "invalid duration",
			errFieldContains: "completion.timeout",
		},
		{
			name:             "invalid cache_ttl returns ConfigError",
			platform:         "",
			completion:       CompletionConfig{Timeout: "", CacheTTL: "forever"},
			wantErr:          true,
			errContains:      "invalid duration",
			errFieldContains: "completion.cache_ttl",
		},
		{
			name:        "both invalid durations joined into one error",
			platform:    "",
			completion:  CompletionConfig{Timeout: "bad1", CacheTTL: "bad2"},
			wantErr:     true,
			errContains: "completion.timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				PlatformDefault: tt.platform,
				Completion:      tt.completion,
			}
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
				if tt.errFieldContains != "" && !containsString(err.Error(), tt.errFieldContains) {
					t.Errorf("Config.Validate() error = %q, want to contain %q", err.Error(), tt.errFieldContains)
				}
			} else {
				if err != nil {
					t.Fatalf("Config.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestConfigValidate_CollectsAllErrors verifies that Validate uses
// errors.Join semantics so users see every problem at once, not just
// the first failure. Pins the collect-all behavior mandated by the
// application-configuration spec.
func TestConfigValidate_CollectsAllErrors(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		PlatformDefault: "invalid-platform",
		Completion: CompletionConfig{
			Timeout:  "bad-timeout",
			CacheTTL: "bad-ttl",
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}

	// errors.Join walks the chain via repeated errors.As; each ConfigError
	// in the chain reports its own Field(). Verify every expected field
	// is represented by some error in the chain.
	wantFields := map[string]bool{
		"platform":             false,
		"completion.timeout":   false,
		"completion.cache_ttl": false,
	}
	for _, e := range unwrapConfigErrors(err) {
		if _, ok := wantFields[e.Field()]; ok {
			wantFields[e.Field()] = true
		}
	}
	for field, seen := range wantFields {
		if !seen {
			t.Errorf("Validate() error chain missing ConfigError for field %q (chain: %v)", field, err)
		}
	}
}

// unwrapConfigErrors walks an errors.Join chain and collects every
// *ConfigError so a test can assert collect-all semantics.
func unwrapConfigErrors(err error) []*gerrors.ConfigError {
	var collected []*gerrors.ConfigError
	type unwrapper interface{ Unwrap() error }
	type joiner interface{ Unwrap() []error }

	for err != nil {
		if j, ok := err.(joiner); ok {
			for _, sub := range j.Unwrap() {
				if u, ok := sub.(*gerrors.ConfigError); ok {
					collected = append(collected, u)
				} else if u, ok := sub.(unwrapper); ok {
					_ = u
				}
			}
		}
		if cfgErr, ok := err.(*gerrors.ConfigError); ok {
			collected = append(collected, cfgErr)
		}
		u, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return collected
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
		Library:         "~/my-library",
		PlatformDefault: "opencode",
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

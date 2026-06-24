package iostreams

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStylesTTY(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		method    func(Styles, string) string
		input     string
		wantAnsi  bool
		wantValue string
	}{
		{
			name:      "Error renders with ANSI when TTY",
			method:    Styles.Error,
			input:     "boom",
			wantAnsi:  true,
			wantValue: "boom",
		},
		{
			name:      "Success renders with ANSI when TTY",
			method:    Styles.Success,
			input:     "ok",
			wantAnsi:  true,
			wantValue: "ok",
		},
		{
			name:      "Warning renders with ANSI when TTY",
			method:    Styles.Warning,
			input:     "watch out",
			wantAnsi:  true,
			wantValue: "watch out",
		},
		{
			name:      "Dim renders with ANSI when TTY",
			method:    Styles.Dim,
			input:     "muted",
			wantAnsi:  true,
			wantValue: "muted",
		},
		{
			name:      "Bold renders with ANSI when TTY",
			method:    Styles.Bold,
			input:     "strong",
			wantAnsi:  true,
			wantValue: "strong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			styles := NewStyles(true)
			out := tt.method(styles, tt.input)
			if tt.wantAnsi {
				assert.Contains(t, out, "\x1b[", "expected ANSI escape code in output")
				assert.Contains(t, out, tt.wantValue)
			} else {
				assert.Equal(t, tt.wantValue, out)
			}
		})
	}
}

func TestStylesNonTTY(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method func(Styles, string) string
		input  string
	}{
		{"Error", Styles.Error, "boom"},
		{"Success", Styles.Success, "ok"},
		{"Warning", Styles.Warning, "watch out"},
		{"Dim", Styles.Dim, "muted"},
		{"Bold", Styles.Bold, "strong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			styles := NewStyles(false)
			out := tt.method(styles, tt.input)
			assert.Equal(t, tt.input, out)
			assert.False(t, strings.Contains(out, "\x1b["), "expected no ANSI escape codes in non-TTY output")
		})
	}
}

func TestStylesNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	styles := NewStyles(true)
	assert.False(t, styles.Enabled(), "color should be disabled when NO_COLOR is set")

	assert.Equal(t, "boom", styles.Error("boom"))
	assert.Equal(t, "ok", styles.Success("ok"))
	assert.Equal(t, "watch out", styles.Warning("watch out"))
	assert.Equal(t, "muted", styles.Dim("muted"))
	assert.Equal(t, "strong", styles.Bold("strong"))
}

func TestStylesNoColorEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	styles := NewStyles(true)
	assert.True(t, styles.Enabled())
}

func TestStylesNoColorUnset(t *testing.T) {
	styles := NewStyles(true)
	assert.True(t, styles.Enabled())
}

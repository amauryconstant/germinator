package iostreams

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const noColorEnv = "NO_COLOR"

// Styles provides color-rendering helpers backed by lipgloss.
// Color is suppressed when stdout is not a TTY, or when the
// NO_COLOR environment variable is set to any non-empty value.
type Styles struct {
	enabled  bool
	renderer *lipgloss.Renderer
}

// NewStyles returns a Styles instance configured for the given TTY
// state. Color is enabled only when isTTY is true AND NO_COLOR is unset.
func NewStyles(isTTY bool) Styles {
	enabled := isTTY && !noColorSet()
	if !enabled {
		r := lipgloss.NewRenderer(os.Stderr)
		r.SetColorProfile(termenv.Ascii)
		return Styles{enabled: false, renderer: r}
	}
	r := lipgloss.NewRenderer(os.Stderr)
	r.SetColorProfile(termenv.TrueColor)
	return Styles{enabled: true, renderer: r}
}

func (s Styles) Error(s2 string) string {
	if !s.enabled {
		return s2
	}
	style := s.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	return style.Render(s2)
}

// Success renders the input as a green success message when colors are enabled.
func (s Styles) Success(s2 string) string {
	if !s.enabled {
		return s2
	}
	style := s.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	return style.Render(s2)
}

// Warning renders the input as a yellow warning message when colors are enabled.
func (s Styles) Warning(s2 string) string {
	if !s.enabled {
		return s2
	}
	style := s.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	return style.Render(s2)
}

// Dim renders the input with a faint style when colors are enabled.
func (s Styles) Dim(s2 string) string {
	if !s.enabled {
		return s2
	}
	style := s.renderer.NewStyle().Faint(true)
	return style.Render(s2)
}

// Bold renders the input in bold when colors are enabled.
func (s Styles) Bold(s2 string) string {
	if !s.enabled {
		return s2
	}
	style := s.renderer.NewStyle().Bold(true)
	return style.Render(s2)
}

// Enabled reports whether color rendering is currently active.
func (s Styles) Enabled() bool {
	return s.enabled
}

func noColorSet() bool {
	v, ok := os.LookupEnv(noColorEnv)
	return ok && v != ""
}

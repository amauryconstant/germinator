// Package iostreams centralizes terminal I/O behind an IOStreams struct
// that all commands use for stdin, stdout, stderr, verbose output, TTY
// detection, color rendering, and structured logging.
package iostreams

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"
)

// IOStreams is the single terminal I/O boundary for commands.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	Verbose bool
	Logger  *slog.Logger
	Styles  Styles

	stdinTTY     bool
	stdoutTTY    bool
	stderrTTY    bool
	overrideTTY  *bool
	overrideErrT *bool
	debug        bool
}

// System returns an IOStreams wired to the process standard streams.
// TTY state is detected from the actual file descriptors. The Logger
// starts disabled; activate it via SetDebug after configuration has
// been loaded (cfg.Debug in main.go flows from koanf env provider,
// which maps GERMINATOR_DEBUG).
func System() *IOStreams {
	stdoutTTY := isTerminalFile(os.Stdout)
	stdinTTY := isTerminalFile(os.Stdin)
	stderrTTY := isTerminalFile(os.Stderr)

	return &IOStreams{
		In:        os.Stdin,
		Out:       os.Stdout,
		ErrOut:    os.Stderr,
		Logger:    newDebugLogger(os.Stderr, false),
		Styles:    NewStyles(stdoutTTY),
		stdinTTY:  stdinTTY,
		stdoutTTY: stdoutTTY,
		stderrTTY: stderrTTY,
	}
}

// Test returns an IOStreams backed by *bytes.Buffer instances so tests
// can assert on captured output. The concrete type of Out and ErrOut
// is *bytes.Buffer, allowing direct inspection via type assertion.
func Test() *IOStreams {
	f := false
	return &IOStreams{
		In:           &bytes.Buffer{},
		Out:          &bytes.Buffer{},
		ErrOut:       &bytes.Buffer{},
		Logger:       newDebugLogger(io.Discard, false),
		Styles:       NewStyles(false),
		stdinTTY:     false,
		stdoutTTY:    false,
		stderrTTY:    false,
		overrideTTY:  &f,
		overrideErrT: &f,
	}
}

// IsStdoutTTY returns true when stdout is connected to a terminal, or
// when SetStdoutTTY(true) has been called.
func (s *IOStreams) IsStdoutTTY() bool {
	if s.overrideTTY != nil {
		return *s.overrideTTY
	}
	return s.stdoutTTY
}

// IsStdinTTY returns true when stdin is connected to a terminal.
func (s *IOStreams) IsStdinTTY() bool {
	return s.stdinTTY
}

// IsStderrTTY returns true when stderr is connected to a terminal, or
// when SetStderrTTY(true) has been called.
func (s *IOStreams) IsStderrTTY() bool {
	if s.overrideErrT != nil {
		return *s.overrideErrT
	}
	return s.stderrTTY
}

// IsInteractive returns true when both stdin AND stdout are TTYs.
func (s *IOStreams) IsInteractive() bool {
	return s.IsStdoutTTY() && s.IsStdinTTY()
}

// SetStdoutTTY overrides the stdout TTY detection. Intended for tests.
func (s *IOStreams) SetStdoutTTY(v bool) {
	s.overrideTTY = &v
	s.Styles = NewStyles(v)
}

// SetStderrTTY overrides the stderr TTY detection. Intended for tests
// (and for the exit-code canary, which needs to gate on stderr state).
func (s *IOStreams) SetStderrTTY(v bool) {
	s.overrideErrT = &v
}

// SetDebug enables or disables the debug Logger. When enabled, the
// Logger writes at slog.LevelDebug to ErrOut; when disabled, it
// discards all output. Configuration flows through the koanf env
// provider (GERMINATOR_DEBUG) into cfg.Debug, which is then applied
// here at startup. Safe to call on a nil receiver.
func (s *IOStreams) SetDebug(enabled bool) {
	if s == nil {
		return
	}
	s.debug = enabled
	if enabled {
		s.Logger = slog.New(slog.NewTextHandler(s.ErrOut, &slog.HandlerOptions{Level: slog.LevelDebug}))
		return
	}
	s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Verbosef writes a formatted message to ErrOut when Verbose is true.
// A trailing newline is appended.
func (s *IOStreams) Verbosef(format string, args ...any) {
	if !s.Verbose {
		return
	}
	if _, err := fmt.Fprintf(s.ErrOut, format+"\n", args...); err != nil {
		_, _ = fmt.Fprintf(s.ErrOut, "iostreams: Verbosef write failed: %v\n", err)
	}
}

// Warnf writes a yellow-styled warning to ErrOut, independent of the
// Verbose flag and the debug Logger. Used by the exit-code deprecation
// canary (slice 2) and any user-facing warning that should always be
// visible regardless of verbosity or debug-logger gating. A trailing
// newline is appended.
func (s *IOStreams) Warnf(format string, args ...any) {
	if s == nil || s.ErrOut == nil {
		return
	}
	prefix := s.Styles.Warning("Warning: ")
	msg := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprint(s.ErrOut, prefix+msg+"\n"); err != nil {
		_, _ = fmt.Fprintf(s.ErrOut, "iostreams: Warnf write failed: %v\n", err)
	}
}

func isTerminalFile(f *os.File) bool {
	fd := f.Fd()
	if fd > maxFd {
		return false
	}
	return term.IsTerminal(int(fd))
}

const maxFd = 1 << 30

func newDebugLogger(w io.Writer, enabled bool) *slog.Logger {
	if enabled {
		return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

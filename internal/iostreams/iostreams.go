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

	stdinTTY    bool
	stdoutTTY   bool
	overrideTTY *bool
}

// System returns an IOStreams wired to the process standard streams.
// TTY state is detected from the actual file descriptors. The Logger
// is gated on the GERMINATOR_DEBUG environment variable: a no-op
// handler when unset, and a debug-level structured handler writing to
// ErrOut when set to any non-empty value.
func System() *IOStreams {
	stdoutTTY := isTerminalFile(os.Stdout)
	stdinTTY := isTerminalFile(os.Stdin)

	return &IOStreams{
		In:        os.Stdin,
		Out:       os.Stdout,
		ErrOut:    os.Stderr,
		Logger:    newDebugLogger(os.Stderr),
		Styles:    NewStyles(stdoutTTY),
		stdinTTY:  stdinTTY,
		stdoutTTY: stdoutTTY,
	}
}

// Test returns an IOStreams backed by *bytes.Buffer instances so tests
// can assert on captured output. The concrete type of Out and ErrOut
// is *bytes.Buffer, allowing direct inspection via type assertion.
func Test() *IOStreams {
	f := false
	return &IOStreams{
		In:          &bytes.Buffer{},
		Out:         &bytes.Buffer{},
		ErrOut:      &bytes.Buffer{},
		Logger:      newDebugLogger(io.Discard),
		Styles:      NewStyles(false),
		stdinTTY:    false,
		stdoutTTY:   false,
		overrideTTY: &f,
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

// IsInteractive returns true when both stdin AND stdout are TTYs.
func (s *IOStreams) IsInteractive() bool {
	return s.IsStdoutTTY() && s.IsStdinTTY()
}

// SetStdoutTTY overrides the stdout TTY detection. Intended for tests.
func (s *IOStreams) SetStdoutTTY(v bool) {
	s.overrideTTY = &v
	s.Styles = NewStyles(v)
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

func isTerminalFile(f *os.File) bool {
	fd := f.Fd()
	if fd > maxFd {
		return false
	}
	return term.IsTerminal(int(fd))
}

const maxFd = 1 << 30

func newDebugLogger(w io.Writer) *slog.Logger {
	if v, ok := os.LookupEnv("GERMINATOR_DEBUG"); ok && v != "" {
		handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
		return slog.New(handler)
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

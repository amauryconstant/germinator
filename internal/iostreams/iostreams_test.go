package iostreams

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystem(t *testing.T) {
	t.Run("uses real stdio", func(t *testing.T) {
		io2 := System()
		require.NotNil(t, io2)
		assert.Equal(t, os.Stdin, io2.In)
		assert.Equal(t, os.Stdout, io2.Out)
		assert.Equal(t, os.Stderr, io2.ErrOut)
	})

	t.Run("logger uses noop when debug unset", func(t *testing.T) {
		io2 := System()
		io2.SetDebug(false)
		require.NotNil(t, io2.Logger)
		assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("logger uses debug when debug set", func(t *testing.T) {
		io2 := System()
		io2.SetDebug(true)
		require.NotNil(t, io2.Logger)
		assert.True(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
	})
}

func TestSystemStderrTTY(t *testing.T) {
	io2 := System()
	require.NotNil(t, io2)
	_ = io2.IsStderrTTY()
}

func TestTest(t *testing.T) {
	t.Parallel()

	t.Run("buffer-backed streams", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		require.NotNil(t, io2)

		_, ok := io2.In.(*bytes.Buffer)
		assert.True(t, ok, "In should be a *bytes.Buffer")
		_, ok = io2.Out.(*bytes.Buffer)
		assert.True(t, ok, "Out should be a *bytes.Buffer")
		_, ok = io2.ErrOut.(*bytes.Buffer)
		assert.True(t, ok, "ErrOut should be a *bytes.Buffer")
	})

	t.Run("IsStdoutTTY false by default", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		assert.False(t, io2.IsStdoutTTY())
		assert.False(t, io2.IsInteractive())
	})
}

func TestVerbosef(t *testing.T) {
	t.Parallel()

	t.Run("writes to ErrOut when verbose", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Verbose = true
		io2.Verbosef("loading %d files", 5)

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "loading 5 files\n", out.String())
	})

	t.Run("writes nothing when not verbose", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Verbose = false
		io2.Verbosef("loading %d files", 5)

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "", out.String())
	})

	t.Run("respects format args", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Verbose = true
		io2.Verbosef("values: %s = %d", "x", 42)

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "values: x = 42\n", out.String())
	})
}

func TestSetStdoutTTY(t *testing.T) {
	t.Parallel()

	t.Run("override true", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		assert.False(t, io2.IsStdoutTTY())
		io2.SetStdoutTTY(true)
		assert.True(t, io2.IsStdoutTTY())
	})

	t.Run("override false after true", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.SetStdoutTTY(true)
		assert.True(t, io2.IsStdoutTTY())
		io2.SetStdoutTTY(false)
		assert.False(t, io2.IsStdoutTTY())
	})
}

func TestIsInteractive(t *testing.T) {
	t.Parallel()

	t.Run("non-interactive when not tty", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		assert.False(t, io2.IsInteractive())
	})

	t.Run("interactive when both tty", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.SetStdoutTTY(true)
		io2.stdinTTY = true
		assert.True(t, io2.IsInteractive())
	})
}

func TestSystemLoggerWritesToErrOut(t *testing.T) {
	io2 := System()
	io2.SetDebug(true)
	require.True(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))

	io2.Logger.Debug("test message", "key", "value")
}

func TestTestLoggerIsNoop(t *testing.T) {
	io2 := Test()
	require.NotNil(t, io2.Logger)
	assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
}

func TestSetDebug(t *testing.T) {
	t.Parallel()

	t.Run("disabled by default", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("enabled writes to ErrOut", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.SetDebug(true)
		io2.Logger.Debug("hello", "k", "v")

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Contains(t, out.String(), "hello")
		assert.Contains(t, out.String(), "k=v")
	})

	t.Run("toggles back to discard", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.SetDebug(true)
		io2.SetDebug(false)
		assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Empty(t, out.String())
	})

	t.Run("nil receiver is a no-op", func(t *testing.T) {
		t.Parallel()
		var io2 *IOStreams
		assert.NotPanics(t, func() { io2.SetDebug(true) })
	})
}

func TestVerbosefNoopWriter(t *testing.T) {
	t.Parallel()

	io2 := &IOStreams{
		In:        &bytes.Buffer{},
		Out:       io.Discard,
		ErrOut:    io.Discard,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Styles:    NewStyles(false),
		Verbose:   false,
		stdinTTY:  false,
		stdoutTTY: false,
	}
	io2.Verbosef("should not panic")
}

func TestIsStderrTTY(t *testing.T) {
	t.Parallel()

	t.Run("default false", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		assert.False(t, io2.IsStderrTTY())
	})

	t.Run("override true", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.SetStderrTTY(true)
		assert.True(t, io2.IsStderrTTY())
	})

	t.Run("override false after true", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.SetStderrTTY(true)
		io2.SetStderrTTY(false)
		assert.False(t, io2.IsStderrTTY())
	})
}

func TestWarnf(t *testing.T) {
	t.Parallel()

	t.Run("writes to ErrOut", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Warnf("legacy exit code %d will be removed", 5)

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "Warning: legacy exit code 5 will be removed\n", out.String())
	})

	t.Run("does not depend on Verbose", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Verbose = false
		io2.Warnf("always visible")

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "Warning: always visible\n", out.String())
	})

	t.Run("does not depend on Logger", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Logger = nil
		io2.Warnf("no logger needed")

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "Warning: no logger needed\n", out.String())
	})

	t.Run("nil IOStreams is a no-op", func(t *testing.T) {
		t.Parallel()
		var io2 *IOStreams
		assert.NotPanics(t, func() { io2.Warnf("nil receiver") })
	})

	t.Run("respects format args", func(t *testing.T) {
		t.Parallel()
		io2 := Test()
		io2.Warnf("values: %s = %d", "x", 42)

		out, ok := io2.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "Warning: values: x = 42\n", out.String())
	})
}

// TestSystem_NoLongerReadsEnvDebug verifies that iostreams.System() no
// longer inspects GERMINATOR_DEBUG at construction time. After Phase 2
// of the wire-factory-config-pipeline change, debug activation is
// driven by cfg.Debug via IOStreams.SetDebug — System() must return a
// discard Logger regardless of the env var, and SetDebug(true) is the
// only path that enables debug output. Sequential (NOT t.Parallel)
// because t.Setenv is incompatible with parallel subtests per
// golang-testing Rule 4.
func TestSystem_NoLongerReadsEnvDebug(t *testing.T) {
	t.Setenv("GERMINATOR_DEBUG", "1")

	io2 := System()
	require.NotNil(t, io2.Logger)
	assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug),
		"System() MUST return a discard Logger even when GERMINATOR_DEBUG=1 is set; debug activation flows through SetDebug(cfg.Debug)")
}

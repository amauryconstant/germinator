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

	t.Run("logger uses noop when GERMINATOR_DEBUG unset", func(t *testing.T) {
		t.Setenv("GERMINATOR_DEBUG", "")
		io2 := System()
		require.NotNil(t, io2.Logger)
		assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("logger uses debug when GERMINATOR_DEBUG set", func(t *testing.T) {
		t.Setenv("GERMINATOR_DEBUG", "1")
		io2 := System()
		require.NotNil(t, io2.Logger)
		assert.True(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
	})
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
	t.Setenv("GERMINATOR_DEBUG", "1")

	io2 := System()
	require.True(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))

	io2.Logger.Debug("test message", "key", "value")
}

func TestTestLoggerIsNoop(t *testing.T) {
	t.Setenv("GERMINATOR_DEBUG", "")

	io2 := Test()
	require.NotNil(t, io2.Logger)
	assert.False(t, io2.Logger.Enabled(context.TODO(), slog.LevelDebug))
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

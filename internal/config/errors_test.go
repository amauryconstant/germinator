package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteError(t *testing.T) {
	t.Parallel()

	t.Run("constructor stores fields", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("permission denied")
		err := NewWriteError("write", "/etc/germinator/config.toml", cause)
		assert.Equal(t, "write", err.Op())
		assert.Equal(t, "/etc/germinator/config.toml", err.Path())
		assert.Same(t, cause, err.Cause())
	})

	t.Run("Error format with cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("permission denied")
		err := NewWriteError("write", "/etc/germinator/config.toml", cause)
		assert.Equal(t, "write /etc/germinator/config.toml: permission denied", err.Error())
	})

	t.Run("Error format without cause", func(t *testing.T) {
		t.Parallel()
		err := NewWriteError("mkdir", "/etc/germinator", nil)
		assert.Equal(t, "mkdir /etc/germinator", err.Error())
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("disk full")
		err := NewWriteError("write", "/tmp/out.toml", cause)
		assert.Same(t, cause, err.Unwrap())
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		t.Parallel()
		err := NewWriteError("stat", "/tmp/in.toml", nil)
		assert.Nil(t, err.Unwrap())
	})

	t.Run("errors.Is traverses chain", func(t *testing.T) {
		t.Parallel()
		cause := fmt.Errorf("wrapped: %w", errors.New("io failure"))
		err := NewWriteError("write", "/tmp/out.toml", cause)
		var target *WriteError
		require.True(t, errors.As(err, &target))
		assert.Same(t, err, target)
	})
}

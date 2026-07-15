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

	t.Run("NewWriteErrorWithMessage stores fields", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("permission denied")
		err := NewWriteErrorWithMessage(
			"create", "/tmp/cfg.toml",
			"config file already exists (use --force to overwrite)", cause,
		)
		assert.Equal(t, "create", err.Op())
		assert.Equal(t, "/tmp/cfg.toml", err.Path())
		assert.Equal(t, "config file already exists (use --force to overwrite)", err.Message())
		assert.Same(t, cause, err.Cause())
	})

	t.Run("Message returns empty string when not set", func(t *testing.T) {
		t.Parallel()
		err := NewWriteError("write", "/tmp/out.toml", errors.New("disk full"))
		assert.Equal(t, "", err.Message(),
			"plain NewWriteError must leave message empty")
	})

	t.Run("Error format with message and cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("permission denied")
		err := NewWriteErrorWithMessage(
			"create", "/etc/cfg.toml",
			"file already exists", cause,
		)
		assert.Equal(t,
			"create /etc/cfg.toml: file already exists: permission denied",
			err.Error())
	})

	t.Run("Error format with message but no cause", func(t *testing.T) {
		t.Parallel()
		err := NewWriteErrorWithMessage(
			"create", "/tmp/cfg.toml",
			"config file already exists (use --force to overwrite)", nil,
		)
		assert.Equal(t,
			"create /tmp/cfg.toml: config file already exists (use --force to overwrite)",
			err.Error())
	})

	t.Run("Error format with empty message and cause behaves like NewWriteError", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("disk full")
		err := NewWriteErrorWithMessage("write", "/tmp/out.toml", "", cause)
		assert.Equal(t, "write /tmp/out.toml: disk full", err.Error(),
			"empty message must collapse to the NewWriteError shape")
	})
}

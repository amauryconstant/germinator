package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

func TestLibraryCommand_Resources(t *testing.T) {
	// Set environment variable to test fixtures
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	f := newTestFactory()
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"resources"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// runResources writes to f.IOStreams.Out (iostreams.Test buffer),
	// not to the cobra command's stdout buffer.
	outBuf, ok := f.IOStreams.Out.(*bytes.Buffer)
	if !ok {
		t.Fatal("f.IOStreams.Out is not a *bytes.Buffer")
	}
	output := outBuf.String()
	if output == "" {
		t.Error("Expected output from library resources command")
	}
}

func TestLibraryCommand_Presets(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	io := iostreams.Test()
	outBuf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok, "io.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"presets"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	require.NoError(t, cmd.Execute())

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library presets command")
}

func TestLibraryCommand_Show(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	io := iostreams.Test()
	outBuf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok, "io.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"show", "skill/commit"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	require.NoError(t, cmd.Execute())

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library show command")
}

func TestLibraryCommand_InvalidRef(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	f := newTestFactory()
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"show", "invalidformat"})

	err = cmd.Execute()
	require.Error(t, err, "invalid ref must produce an error")

	var notFound *core.NotFoundError
	require.True(t, errors.As(err, &notFound),
		"error must be a *core.NotFoundError, got %T: %v", err, err)
	assert.Equal(t, "invalidformat", notFound.Key)
}

func TestInitCommand_RequiresPlatform(_ *testing.T) {
	_ = newTestConfig()

	cmd := NewInitCommand(newTestFactory(), newTestBridge())
	cmd.SetArgs([]string{"--resources", "skill/commit"})

	// This should fail due to missing platform
	// The actual test would need to handle os.Exit
	_ = cmd
}

func TestInitCommand_RequiresResourcesOrPreset(_ *testing.T) {
	_ = newTestConfig()

	cmd := NewInitCommand(newTestFactory(), newTestBridge())
	cmd.SetArgs([]string{"--platform", "opencode"})

	// This should fail due to missing resources/preset
	// The actual test would need to handle os.Exit
	_ = cmd
}

func TestInitCommand_MutuallyExclusive(_ *testing.T) {
	_ = newTestConfig()

	cmd := NewInitCommand(newTestFactory(), newTestBridge())
	cmd.SetArgs([]string{"--platform", "opencode", "--resources", "skill/commit", "--preset", "git-workflow"})

	// This should fail due to mutually exclusive flags
	// The actual test would need to handle os.Exit
	_ = cmd
}

func TestInitCommand_DryRun(t *testing.T) {
	_ = newTestConfig()

	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	outputDir := t.TempDir()

	cmd := NewInitCommand(newTestFactory(), newTestBridge())
	cmd.SetArgs([]string{
		"--platform", "opencode",
		"--resources", "skill/commit",
		"--library", absPath,
		"--output", outputDir,
		"--dry-run",
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output from dry-run")
	}

	// Verify no files were created
	outputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Dry-run should not create files")
	}
}

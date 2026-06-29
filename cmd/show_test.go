package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

func newShowTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func newShowOpts(t *testing.T, lib *library.Library, ref, output string) (*showOptions, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	io, out, errOut := newShowTestIO()
	opts := &showOptions{
		IO:     io,
		Ref:    ref,
		Output: output,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}
	return opts, out, errOut
}

func TestRunShow(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)

	t.Run("resource ref plain", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "skill/commit", "")

		require.NoError(t, runShow(opts))

		got := out.String()
		assert.Contains(t, got, "Reference: skill/commit")
		assert.Contains(t, got, "Path: skills/skill-commit.md")
		assert.Contains(t, got, "Description: Git commit best practices")
		assert.Empty(t, errOut.String(),
			"plain output must NOT write to stderr")
	})

	t.Run("resource ref JSON", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "skill/commit", "json")

		require.NoError(t, runShow(opts))

		var parsed showResourceRow
		require.NoError(t, json.Unmarshal(out.Bytes(), &parsed),
			"output must be valid JSON: %q", out.String())
		assert.Equal(t, "skill/commit", parsed.Ref)
		assert.Equal(t, "skills/skill-commit.md", parsed.Path)
		assert.Equal(t, "Git commit best practices", parsed.Description)
		assert.Empty(t, errOut.String())
	})

	t.Run("resource ref table", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "skill/commit", "table")

		require.NoError(t, runShow(opts))

		got := out.String()
		assert.Contains(t, got, "REF")
		assert.Contains(t, got, "DESCRIPTION")
		assert.Contains(t, got, "skill/commit")
		assert.NotContains(t, got, "skills/skill-commit.md",
			"Path is `tab:\"-\"` and must be hidden from table output")
		assert.Empty(t, errOut.String())
	})

	t.Run("preset ref plain", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "preset/git-workflow", "")

		require.NoError(t, runShow(opts))

		got := out.String()
		assert.Contains(t, got, "Preset: git-workflow")
		assert.Contains(t, got, "Description: Git workflow tools")
		assert.Contains(t, got, "Resources:")
		assert.Contains(t, got, "skill/commit")
		assert.Contains(t, got, "skill/merge-request")
		assert.Empty(t, errOut.String())
	})

	t.Run("preset ref JSON", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "preset/git-workflow", "json")

		require.NoError(t, runShow(opts))

		var parsed showPresetRow
		require.NoError(t, json.Unmarshal(out.Bytes(), &parsed),
			"output must be valid JSON: %q", out.String())
		assert.Equal(t, "git-workflow", parsed.Name)
		assert.Equal(t, "Git workflow tools", parsed.Description)
		assert.Equal(t, []string{"skill/commit", "skill/merge-request"}, parsed.Resources)
		assert.Empty(t, errOut.String())
	})

	t.Run("preset ref table", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "preset/git-workflow", "table")

		require.NoError(t, runShow(opts))

		got := out.String()
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "DESCRIPTION")
		assert.Contains(t, got, "RESOURCES")
		assert.Contains(t, got, "git-workflow")
		assert.Empty(t, errOut.String())
	})

	t.Run("not-found resource ref", func(t *testing.T) {
		t.Parallel()
		opts, _, _ := newShowOpts(t, lib, "skill/missing", "")

		err := runShow(opts)
		require.Error(t, err)

		var notFound *core.NotFoundError
		require.ErrorAs(t, err, &notFound,
			"error must be a *core.NotFoundError")
		assert.Equal(t, "skill/missing", notFound.Key)
		assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
			"NotFoundError must map to ExitCodeUsage (2)")
	})

	t.Run("not-found preset ref", func(t *testing.T) {
		t.Parallel()
		opts, _, _ := newShowOpts(t, lib, "preset/ghost", "")

		err := runShow(opts)
		require.Error(t, err)

		var notFound *core.NotFoundError
		require.ErrorAs(t, err, &notFound,
			"error must be a *core.NotFoundError")
		assert.Equal(t, "preset/ghost", notFound.Key)
	})

	t.Run("empty ref", func(t *testing.T) {
		t.Parallel()
		opts, _, _ := newShowOpts(t, lib, "", "")

		err := runShow(opts)
		require.Error(t, err)

		var notFound *core.NotFoundError
		require.ErrorAs(t, err, &notFound,
			"empty ref must produce NotFoundError")
		assert.Equal(t, "", notFound.Key)
		assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
			"empty-ref NotFoundError must map to ExitCodeUsage (2)")
	})

	t.Run("no-slash ref", func(t *testing.T) {
		t.Parallel()
		opts, _, _ := newShowOpts(t, lib, "no-slash", "")

		err := runShow(opts)
		require.Error(t, err)

		var notFound *core.NotFoundError
		require.ErrorAs(t, err, &notFound,
			"no-slash ref must produce NotFoundError")
		assert.Equal(t, "no-slash", notFound.Key)
	})

	t.Run("preset prefix with empty name", func(t *testing.T) {
		t.Parallel()
		opts, _, _ := newShowOpts(t, lib, "preset/", "")

		err := runShow(opts)
		require.Error(t, err)

		var notFound *core.NotFoundError
		require.ErrorAs(t, err, &notFound,
			"preset/ with empty name must produce NotFoundError")
		assert.Equal(t, "preset/", notFound.Key)
	})

	t.Run("agent ref plain", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "agent/reviewer", "")

		require.NoError(t, runShow(opts))

		got := out.String()
		assert.Contains(t, got, "Reference: agent/reviewer")
		assert.Contains(t, got, "Path: agents/agent-reviewer.md")
		assert.Contains(t, got, "Description: Code review agent")
		assert.Empty(t, errOut.String())
	})
}

func TestRunShow_StreamDiscipline(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)

	t.Run("missing ref writes to stderr, not stdout", func(t *testing.T) {
		t.Parallel()
		opts, out, errOut := newShowOpts(t, lib, "nonexistent-ref", "")

		err := runShow(opts)
		require.Error(t, err)

		assert.Empty(t, out.String(),
			"stdout must be empty on error path (no data leakage)")
		assert.Empty(t, errOut.String(),
			"runShow must NOT write to stderr; FormatError is the writer")

		output.FormatError(opts.IO, err)
		assert.NotEmpty(t, errOut.String(),
			"FormatError must write the not-found message to stderr")
	})

	t.Run("successful invocation keeps stderr empty", func(t *testing.T) {
		t.Parallel()
		opts, _, errOut := newShowOpts(t, lib, "skill/commit", "")

		require.NoError(t, runShow(opts))
		assert.Empty(t, errOut.String(),
			"stderr must be empty for successful invocation")
	})
}

func TestRunShow_LibraryLoadError(t *testing.T) {
	t.Parallel()

	io, out, errOut := newShowTestIO()
	opts := &showOptions{
		IO:     io,
		Ref:    "skill/commit",
		Output: "json",
		Library: func() (*library.Library, error) {
			return nil, errors.New("library unavailable")
		},
		Ctx: context.Background(),
	}

	err := runShow(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading library")
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestNewCmdShow_RunFInjectionCapturesOpts(t *testing.T) {
	t.Parallel()

	var captured *showOptions
	runF := func(opts *showOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdShow(f, nil, runF)
	cmd.SetArgs([]string{"skill/commit"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	require.NotNil(t, captured.IO)
	assert.Equal(t, io, captured.IO, "opts.IO must be the Factory's IOStreams")
	assert.Equal(t, "skill/commit", captured.Ref, "opts.Ref must be args[0]")
	assert.Equal(t, "plain", captured.Output, "default --output value must be \"plain\"")
	require.NotNil(t, captured.Library, "opts.Library must be a non-nil lazy function")
}

func TestNewCmdShow_HonorsOutputFlagValue(t *testing.T) {
	t.Parallel()

	for _, format := range []string{"json", "table", "plain"} {
		format := format
		t.Run("format="+format, func(t *testing.T) {
			t.Parallel()

			lib := loadFixtureLibrary(t)
			var capturedOutput string
			runF := func(opts *showOptions) error {
				capturedOutput = opts.Output
				opts.Library = func() (*library.Library, error) { return lib, nil }
				return runShow(opts)
			}

			io := iostreams.Test()
			f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
			cmd := NewCmdShow(f, nil, runF)
			cmd.SetArgs([]string{"skill/commit", "--output", format})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			require.NoError(t, cmd.Execute())
			assert.Equal(t, format, capturedOutput,
				"--output flag value must reach opts.Output")
		})
	}
}

func TestNewCmdShow_OldJSONFlagIsRejected(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	cmd := NewCmdShow(f, nil, func(*showOptions) error { return nil })
	cmd.SetArgs([]string{"skill/commit", "--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.Error(t, err, "old --json flag must be rejected as unknown")
	assert.Contains(t, err.Error(), "unknown flag",
		"error must indicate the rejected flag: %v", err)
}

func TestNewCmdShow_PassesLibraryFlagToLoader(t *testing.T) {
	t.Parallel()

	fixtureRel, err := filepath.Abs(filepath.Join("..", "test", "fixtures", "library"))
	require.NoError(t, err)

	var capturedPath string
	var runOpts *showOptions
	runF := func(opts *showOptions) error {
		runOpts = opts
		lib, lerr := opts.Library()
		if lerr != nil {
			return lerr
		}
		capturedPath = lib.RootPath
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	libraryFlag := fixtureRel
	cmd := NewCmdShow(f, &libraryFlag, runF)
	cmd.SetArgs([]string{"skill/commit"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, runOpts)
	assert.Equal(t, fixtureRel, capturedPath,
		"--library flag value must drive the resolved library path")
}

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
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

func newPresetsTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func newPresetsOpts(t *testing.T, lib *library.Library, output string) (*presetsOptions, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	io, out, errOut := newPresetsTestIO()
	opts := &presetsOptions{
		IO:     io,
		Output: output,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}
	return opts, out, errOut
}

func TestRunPresets_Plain(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	opts, out, errOut := newPresetsOpts(t, lib, "")

	require.NoError(t, runPresets(opts))

	assert.Equal(t, formatPresetsList(lib), out.String(),
		"plain output must be byte-identical to formatPresetsList(lib)")
	assert.Empty(t, errOut.String(),
		"plain output must NOT write to stderr (no verbose leakage)")
}

func TestRunPresets_PlainIsDefault(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)

	for _, output := range []string{"", "plain"} {
		output := output
		t.Run("output="+output, func(t *testing.T) {
			t.Parallel()
			opts, out, _ := newPresetsOpts(t, lib, output)
			require.NoError(t, runPresets(opts))
			assert.Equal(t, formatPresetsList(lib), out.String(),
				"plain output (default or explicit) must match formatPresetsList")
		})
	}
}

func TestRunPresets_JSON(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	opts, out, errOut := newPresetsOpts(t, lib, "json")

	require.NoError(t, runPresets(opts))

	var parsed struct {
		Presets []presetsRow `json:"presets"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &parsed),
		"output must be valid JSON: %q", out.String())

	require.Len(t, parsed.Presets, 2,
		"fixture library has 2 presets (git-workflow, code-review)")

	assert.Equal(t, "code-review", parsed.Presets[0].Name)
	assert.Equal(t, "git-workflow", parsed.Presets[1].Name)

	for i, p := range parsed.Presets {
		assert.NotEmpty(t, p.Name, "presets[%d].name must be non-empty", i)
		assert.NotEmpty(t, p.Description, "presets[%d].description must be non-empty", i)
		assert.NotEmpty(t, p.Resources, "presets[%d].resources must be non-empty", i)
	}

	assert.Empty(t, errOut.String(),
		"JSON output must NOT write to stderr (stream discipline)")
}

func TestRunPresets_Table(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	opts, out, errOut := newPresetsOpts(t, lib, "table")

	require.NoError(t, runPresets(opts))

	output := out.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "DESCRIPTION")
	assert.Contains(t, output, "RESOURCES")

	assert.Contains(t, output, "git-workflow")
	assert.Contains(t, output, "code-review")

	assert.Empty(t, errOut.String(),
		"table output must NOT write to stderr (stream discipline)")
}

func TestRunPresets_StreamDiscipline(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)

	for _, format := range []string{"", "plain", "json", "table"} {
		format := format
		t.Run("format="+format, func(t *testing.T) {
			t.Parallel()
			opts, out, errOut := newPresetsOpts(t, lib, format)
			require.NoError(t, runPresets(opts))
			assert.NotEmpty(t, out.String(), "stdout must contain data")
			assert.Empty(t, errOut.String(),
				"stderr must be empty for format %q (no verbose leakage)", format)
		})
	}
}

func TestRunPresets_LibraryLoadError(t *testing.T) {
	t.Parallel()

	io, out, errOut := newPresetsTestIO()
	opts := &presetsOptions{
		IO:     io,
		Output: "json",
		Library: func() (*library.Library, error) {
			return nil, errors.New("library unavailable")
		},
		Ctx: context.Background(),
	}

	err := runPresets(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading library")
	assert.Contains(t, err.Error(), "library unavailable",
		"original cause must be preserved in the chain")
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestNewCmdPresets_RunFInjectionCapturesOpts(t *testing.T) {
	var captured *presetsOptions
	runF := func(opts *presetsOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io)
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdPresets(f, nil, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}))
	require.NotNil(t, captured, "runF must be invoked")
	require.NotNil(t, captured.IO)
	assert.Equal(t, io, captured.IO, "opts.IO must be the Factory's IOStreams")
	assert.Equal(t, "plain", captured.Output, "default --output value must be \"plain\"")
	require.NotNil(t, captured.Library, "opts.Library must be a non-nil lazy function")
}

func TestNewCmdPresets_PassesLibraryFlagToLoader(t *testing.T) {
	fixtureRel, err := filepath.Abs(filepath.Join("..", "test", "fixtures", "library"))
	require.NoError(t, err)

	var capturedPath string
	var runOpts *presetsOptions
	runF := func(opts *presetsOptions) error {
		runOpts = opts
		lib, lerr := opts.Library()
		if lerr != nil {
			return lerr
		}
		capturedPath = lib.RootPath
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io)

	libraryFlag := fixtureRel
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdPresets(f, &libraryFlag, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}))
	require.NotNil(t, runOpts)
	assert.Equal(t, fixtureRel, capturedPath,
		"--library flag value must drive the resolved library path")
}

func TestNewCmdPresets_HonorsOutputFlagValue(t *testing.T) {
	for _, format := range []string{"json", "table", "plain"} {
		format := format
		t.Run("format="+format, func(t *testing.T) {
			lib := loadFixtureLibrary(t)
			var capturedOutput string
			runF := func(opts *presetsOptions) error {
				capturedOutput = opts.Output
				opts.Library = func() (*library.Library, error) { return lib, nil }
				return runPresets(opts)
			}

			io := iostreams.Test()
			f := cmdutil.NewFactory(context.Background(), io)
			require.NoError(t, executeCmd(t, func() any {
				cmd := NewCmdPresets(f, nil, runF)
				cmd.SetOut(&bytes.Buffer{})
				cmd.SetErr(&bytes.Buffer{})
				return cmd
			}, "--output", format))
			assert.Equal(t, format, capturedOutput,
				"--output flag value must reach opts.Output")
		})
	}
}

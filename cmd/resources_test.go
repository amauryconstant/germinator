package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// loadFixtureLibrary loads the test/fixtures/library directory used by the
// E2E golden tests; tests can compare against the same library content
// without invoking the binary.
func loadFixtureLibrary(t *testing.T) *library.Library {
	t.Helper()

	abs, err := filepath.Abs(filepath.Join("..", "test", "fixtures", "library"))
	require.NoError(t, err)

	lib, err := library.LoadLibrary(context.Background(), abs)
	require.NoError(t, err, "fixture library must load: %s", abs)
	return lib
}

func newResourcesTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func newResourcesOpts(t *testing.T, lib *library.Library, output string) (*libraryResourcesOptions, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	io, out, errOut := newResourcesTestIO()
	opts := &libraryResourcesOptions{
		IO:     io,
		Output: output,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}
	return opts, out, errOut
}

func TestRunResources_Plain(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	opts, out, errOut := newResourcesOpts(t, lib, "")

	require.NoError(t, runResources(opts))

	assert.Equal(t, output.FormatResourcesList(lib), out.String(),
		"plain output must be byte-identical to output.FormatResourcesList(lib)")
	assert.Empty(t, errOut.String(),
		"plain output must NOT write to stderr (no verbose leakage)")
}

func TestRunResources_PlainIsDefault(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)

	// Both the zero-value "" and the explicit "plain" must yield the
	// byte-identical plain output (the spec's "Plain is the default"
	// scenario requires this equivalence).
	for _, format := range []string{"", "plain"} {
		format := format
		t.Run("output="+format, func(t *testing.T) {
			t.Parallel()
			opts, out, _ := newResourcesOpts(t, lib, format)
			require.NoError(t, runResources(opts))
			assert.Equal(t, output.FormatResourcesList(lib), out.String(),
				"plain output (default or explicit) must match output.FormatResourcesList")
		})
	}
}

func TestRunResources_JSON(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	opts, out, errOut := newResourcesOpts(t, lib, "json")

	require.NoError(t, runResources(opts))

	// Spec: {"resources": [{"type": "...", "name": "...", "description": "...", "path": "..."}, ...]}
	var parsed struct {
		Resources []resourcesRow `json:"resources"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &parsed),
		"output must be valid JSON: %q", out.String())

	require.Len(t, parsed.Resources, 5,
		"fixture library has 5 resources (2 skills + 1 agent + 1 command + 1 memory)")

	// Validate the stable JSON shape per the library-library-json-output
	// delta spec: each entry must have type, name, description, path.
	for i, r := range parsed.Resources {
		assert.NotEmpty(t, r.Type, "resources[%d].type must be non-empty", i)
		assert.NotEmpty(t, r.Name, "resources[%d].name must be non-empty", i)
		assert.NotEmpty(t, r.Description, "resources[%d].description must be non-empty (no omitempty)", i)
		assert.NotEmpty(t, r.Path, "resources[%d].path must be non-empty", i)
	}

	// Deterministic ordering: skills first, then agents, commands, memory
	assert.Equal(t, "skill", parsed.Resources[0].Type)
	assert.Equal(t, "commit", parsed.Resources[0].Name)
	assert.Equal(t, "skill", parsed.Resources[1].Type)
	assert.Equal(t, "merge-request", parsed.Resources[1].Name)
	assert.Equal(t, "agent", parsed.Resources[2].Type)
	assert.Equal(t, "reviewer", parsed.Resources[2].Name)
	assert.Equal(t, "command", parsed.Resources[3].Type)
	assert.Equal(t, "test", parsed.Resources[3].Name)
	assert.Equal(t, "memory", parsed.Resources[4].Type)
	assert.Equal(t, "context", parsed.Resources[4].Name)

	assert.Empty(t, errOut.String(),
		"JSON output must NOT write to stderr (stream discipline)")
}

func TestRunResources_Table(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	opts, out, errOut := newResourcesOpts(t, lib, "table")

	require.NoError(t, runResources(opts))

	output := out.String()
	// tab:"HEADER" tags drive the table header row.
	assert.Contains(t, output, "TYPE")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "DESCRIPTION")
	assert.NotContains(t, output, "PATH", "Path is `tab:\"-\"` and must be hidden from table output")

	// Table shows each resource's name column.
	assert.Contains(t, output, "commit")
	assert.Contains(t, output, "merge-request")
	assert.Contains(t, output, "reviewer")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "context")

	assert.Empty(t, errOut.String(),
		"table output must NOT write to stderr (stream discipline)")
}

func TestRunResources_StreamDiscipline_AllFormats(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)

	for _, format := range []string{"", "plain", "json", "table"} {
		format := format
		t.Run("format="+format, func(t *testing.T) {
			t.Parallel()
			opts, out, errOut := newResourcesOpts(t, lib, format)
			require.NoError(t, runResources(opts))
			assert.NotEmpty(t, out.String(), "stdout must contain data")
			assert.Empty(t, errOut.String(),
				"stderr must be empty for format %q (no verbose leakage)", format)
		})
	}
}

func TestRunResources_LibraryLoadError(t *testing.T) {
	t.Parallel()

	io, out, errOut := newResourcesTestIO()
	opts := &libraryResourcesOptions{
		IO:     io,
		Output: "json",
		Library: func() (*library.Library, error) {
			return nil, errors.New("library unavailable")
		},
		Ctx: context.Background(),
	}

	err := runResources(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading library")
	assert.Contains(t, err.Error(), "library unavailable",
		"original cause must be preserved in the chain")
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestNewCmdResources_RunFInjectionCapturesOpts(t *testing.T) {
	var captured *libraryResourcesOptions
	runF := func(opts *libraryResourcesOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdResources(f, nil, runF)
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

func TestNewCmdResources_PassesLibraryFlagToLoader(t *testing.T) {
	fixturePath, err := filepath.Abs(filepath.Join("..", "test", "fixtures", "library"))
	require.NoError(t, err)

	var capturedPath string
	var runOpts *libraryResourcesOptions
	runF := func(opts *libraryResourcesOptions) error {
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

	libraryFlag := fixturePath
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdResources(f, &libraryFlag, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}))
	require.NotNil(t, runOpts)
	assert.Equal(t, fixturePath, capturedPath,
		"--library flag value must drive the resolved library path")
}

func TestNewCmdResources_OldJSONFlagIsRejected(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	err := executeCmd(t, func() any {
		cmd := NewCmdResources(f, nil, func(*libraryResourcesOptions) error { return nil })
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		return cmd
	}, "--json")
	require.Error(t, err, "old --json flag must be rejected as unknown")

	// Spec: process exits with code 2 (ExitCodeUsage) when an unknown flag
	// is used. Cobra returns the error from RunE / Flag parsing; the
	// command-level error must surface as a usage error.
	assert.True(t,
		strings.Contains(err.Error(), "unknown flag") ||
			strings.Contains(err.Error(), "json"),
		"error must indicate the rejected flag: %v", err)
}

func TestNewCmdResources_HonorsOutputFlagValue(t *testing.T) {
	for _, format := range []string{"json", "table", "plain"} {
		format := format
		t.Run("format="+format, func(t *testing.T) {
			lib := loadFixtureLibrary(t)
			var capturedOutput string
			runF := func(opts *libraryResourcesOptions) error {
				capturedOutput = opts.Output
				opts.Library = func() (*library.Library, error) { return lib, nil }
				return runResources(opts)
			}

			io := iostreams.Test()
			f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
			require.NoError(t, executeCmd(t, func() any {
				cmd := NewCmdResources(f, nil, runF)
				cmd.SetOut(&bytes.Buffer{})
				cmd.SetErr(&bytes.Buffer{})
				return cmd
			}, "--output", format))
			assert.Equal(t, format, capturedOutput,
				"--output flag value must reach opts.Output")
		})
	}
}

// TestRunResources_DebugLogEmittedWhenEnabled pins the spec scenario
// "GERMINATOR_DEBUG enables debug logging" from the
// application-configuration delta: when cfg.Debug drives
// IOStreams.SetDebug(true), runResources MUST emit a debug log line
// to ErrOut. The ErrOut buffer is the debug channel (verbose goes
// through IOStreams.Verbosef, debug through Logger.Debug; both write
// to ErrOut).
func TestRunResources_DebugLogEmittedWhenEnabled(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	io, _, errOut := newResourcesTestIO()
	io.SetDebug(true)

	opts := &libraryResourcesOptions{
		IO:     io,
		Output: "",
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runResources(opts))
	assert.Contains(t, errOut.String(), "listing library resources",
		"debug logger MUST emit a structured line to ErrOut when SetDebug(true)")
}

// TestRunResources_DebugLogSilentWhenDisabled verifies the negative:
// with debug disabled (default iostreams.Test()), ErrOut stays empty
// after runResources. Pins the "logger uses noop when debug unset"
// contract.
func TestRunResources_DebugLogSilentWhenDisabled(t *testing.T) {
	t.Parallel()

	lib := loadFixtureLibrary(t)
	io, _, errOut := newResourcesTestIO()
	// iostreams.Test() returns a discard Logger by default.

	opts := &libraryResourcesOptions{
		IO:     io,
		Output: "",
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runResources(opts))
	assert.Empty(t, errOut.String(),
		"with debug disabled, runResources MUST NOT emit to ErrOut")
}

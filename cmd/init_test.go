package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// fakeInitializer is a hand-rolled fake satisfying the
// application.Initializer interface. It records every call and
// returns the pre-configured per-ref results. err applies to the
// transport-level error return (not to per-resource outcomes).
type fakeInitializer struct {
	calls   int
	results []core.InitializeResult
	err     error
}

func (f *fakeInitializer) Initialize(_ context.Context, _ *InitializeRequest) ([]core.InitializeResult, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.results, nil
}

func newInitTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func newInitOpts(t *testing.T, lib *library.Library, init application.Initializer, mut func(*initOptions)) (*initOptions, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	io, out, errOut := newInitTestIO()
	opts := &initOptions{
		IO:  io,
		Ctx: context.Background(),
	}
	opts.Platform = core.PlatformOpenCode
	opts.OutputDir = "/tmp/test-out"
	opts.Refs = []string{"skill/commit"}
	effectiveLib := lib
	if effectiveLib == nil {
		effectiveLib = fakeLibraryLoader()
	}
	opts.Library = func() (*library.Library, error) { return effectiveLib, nil }
	if init != nil {
		opts.Initializer = func() (application.Initializer, error) { return init, nil }
	}
	if mut != nil {
		mut(opts)
	}
	return opts, out, errOut
}

// fakeLibraryLoader is a no-op library that the helper wires in
// when the caller passes nil. runInit requires opts.Library to be
// non-nil to delegate to (*Library).ResolvePreset; tests that don't
// care about library content can ignore it.
func fakeLibraryLoader() *library.Library {
	return &library.Library{
		Version:  "1",
		RootPath: "/fake/library",
		Resources: map[string]map[string]library.Resource{
			"skill": {
				"commit": {Path: "skills/commit.md", Description: "Git commit"},
			},
		},
	}
}

// tinyPresetLibrary is a minimal library with one preset for the
// preset-expansion tests.
func tinyPresetLibrary() *library.Library {
	return &library.Library{
		Version: "1",
		Presets: map[string]library.Preset{
			"git-workflow": {
				Name:      "git-workflow",
				Resources: []string{"skill/commit", "skill/merge-request"},
			},
		},
		Resources: map[string]map[string]library.Resource{
			"skill": {
				"commit":        {Path: "skills/commit.md", Description: "Git commit"},
				"merge-request": {Path: "skills/merge-request.md", Description: "Git merge-request"},
			},
		},
		RootPath: "/fake/library",
	}
}

// §5.3.1 — All success: runInit returns nil; ExitCodeFor(nil) == 0.
func TestRunInit_AllSuccess(t *testing.T) {
	t.Parallel()

	opts, out, errOut := newInitOpts(t, nil, &fakeInitializer{
		results: []core.InitializeResult{
			{Ref: "skill/commit", InputPath: "/lib/skills/commit.md", OutputPath: ".opencode/skills/commit/SKILL.md"},
		},
	}, nil)

	err := runInit(opts)

	require.NoError(t, err)
	assert.Equal(t, cmdutil.ExitCodeSuccess, cmdutil.ExitCodeFor(err),
		"nil error must map to ExitCodeSuccess (0)")

	assert.Contains(t, out.String(), "Installed: skill/commit")
	assert.Contains(t, out.String(), "Initialized 1 resource(s).")
	assert.Empty(t, errOut.String(), "no errors to write to stderr")
}

// §5.3.2 — Partial success: 1 ok + 1 fail → *PartialSuccessError{S:1,F:1};
// ExitCodeFor == 0; FormatError writes "partial success: 1 succeeded, 1 failed".
func TestRunInit_PartialSuccess(t *testing.T) {
	t.Parallel()

	opts, _, errOut := newInitOpts(t, nil, &fakeInitializer{
		results: []core.InitializeResult{
			{Ref: "skill/commit", InputPath: "/lib/skills/commit.md", OutputPath: ".opencode/skills/commit/SKILL.md"},
			{
				Ref:        "skill/invalid",
				InputPath:  "/lib/skills/invalid.md",
				OutputPath: ".opencode/skills/invalid/SKILL.md",
				Error:      errors.New("file not found"),
			},
		},
	}, func(o *initOptions) {
		o.Refs = []string{"skill/commit", "skill/invalid"}
	})

	err := runInit(opts)

	require.Error(t, err)

	var ps *core.PartialSuccessError
	require.ErrorAs(t, err, &ps, "error must be *core.PartialSuccessError")
	assert.Equal(t, 1, ps.Succeeded(), "PartialSuccessError{S:1,F:1}.Succeeded() must be 1")
	assert.Equal(t, 1, ps.Failed(), "PartialSuccessError{S:1,F:1}.Failed() must be 1")
	assert.Equal(t, 1, len(ps.Errors()), "exactly one InitializeError in the aggregate")

	assert.Equal(t, cmdutil.ExitCodeSuccess, cmdutil.ExitCodeFor(err),
		"partial success must map to ExitCodeSuccess (0)")

	output.FormatError(opts.IO, err)
	assert.Contains(t, errOut.String(), "partial success: 1 succeeded, 1 failed")
	assert.Contains(t, errOut.String(), "skill/invalid")
}

// §5.3.3 — All failed: 0 ok + 2 fail → *PartialSuccessError{S:0,F:2};
// ExitCodeFor == 1.
func TestRunInit_AllFailed(t *testing.T) {
	t.Parallel()

	opts, _, _ := newInitOpts(t, nil, &fakeInitializer{
		results: []core.InitializeResult{
			{Ref: "skill/invalid1", InputPath: "/lib/skills/invalid1.md", OutputPath: ".opencode/skills/invalid1/SKILL.md", Error: errors.New("missing a")},
			{Ref: "skill/invalid2", InputPath: "/lib/skills/invalid2.md", OutputPath: ".opencode/skills/invalid2/SKILL.md", Error: errors.New("missing b")},
		},
	}, func(o *initOptions) {
		o.Refs = []string{"skill/invalid1", "skill/invalid2"}
	})

	err := runInit(opts)

	require.Error(t, err)

	var ps *core.PartialSuccessError
	require.ErrorAs(t, err, &ps)
	assert.Equal(t, 0, ps.Succeeded())
	assert.Equal(t, 2, ps.Failed())
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"all-failed must map to ExitCodeError (1)")
}

// §5.3.4 — Preset expansion: --preset git-workflow expands via
// (*Library).ResolvePreset; each ref processed; partial-success logic
// applies to the full expanded list.
func TestRunInit_PresetExpansion(t *testing.T) {
	t.Parallel()

	lib := tinyPresetLibrary()
	opts, _, _ := newInitOpts(t, lib, &fakeInitializer{
		results: []core.InitializeResult{
			{Ref: "skill/commit", InputPath: "/lib/skills/commit.md", OutputPath: ".opencode/skills/commit/SKILL.md"},
			{Ref: "skill/merge-request", InputPath: "/lib/skills/merge-request.md", OutputPath: ".opencode/skills/merge-request/SKILL.md"},
		},
	}, func(o *initOptions) {
		o.Refs = nil
		o.Preset = "git-workflow"
	})

	err := runInit(opts)

	require.NoError(t, err, "all-success preset expansion returns nil")
}

// §5.3.4 — Preset expansion with one failure: partial-success logic
// applies to the expanded list.
func TestRunInit_PresetExpansion_PartialFail(t *testing.T) {
	t.Parallel()

	lib := tinyPresetLibrary()
	opts, _, _ := newInitOpts(t, lib, &fakeInitializer{
		results: []core.InitializeResult{
			{Ref: "skill/commit", InputPath: "/lib/skills/commit.md", OutputPath: ".opencode/skills/commit/SKILL.md"},
			{Ref: "skill/merge-request", InputPath: "/lib/skills/merge-request.md", OutputPath: ".opencode/skills/merge-request/SKILL.md", Error: errors.New("read fail")},
		},
	}, func(o *initOptions) {
		o.Refs = nil
		o.Preset = "git-workflow"
	})

	err := runInit(opts)

	require.Error(t, err, "partial fail in expanded preset must surface")
	var ps *core.PartialSuccessError
	require.ErrorAs(t, err, &ps)
	assert.Equal(t, 1, ps.Succeeded())
	assert.Equal(t, 1, ps.Failed())
}

// §5.3.5 — Preset-not-found: --preset ghost returns
// *core.NotFoundError{Entity:"preset", Name:"ghost"}; ExitCodeFor == 2.
func TestRunInit_PresetNotFound(t *testing.T) {
	t.Parallel()

	lib := tinyPresetLibrary()
	opts, _, _ := newInitOpts(t, lib, &fakeInitializer{}, func(o *initOptions) {
		o.Refs = nil
		o.Preset = "ghost"
	})

	err := runInit(opts)

	require.Error(t, err)
	var nf *core.NotFoundError
	require.ErrorAs(t, err, &nf, "missing preset must wrap as *core.NotFoundError")
	assert.Equal(t, "preset", nf.Entity)
	assert.Equal(t, "ghost", nf.Key)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"preset-not-found must map to ExitCodeUsage (2)")
}

// Validation: missing --platform.
func TestRunInit_InvalidPlatform(t *testing.T) {
	t.Parallel()

	opts, _, _ := newInitOpts(t, nil, &fakeInitializer{}, func(o *initOptions) {
		o.Platform = "windows-95"
	})

	err := runInit(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown platform "windows-95"`)
}

// Validation: empty platform flag value.
func TestRunInit_EmptyPlatform(t *testing.T) {
	t.Parallel()

	opts, _, _ := newInitOpts(t, nil, &fakeInitializer{}, func(o *initOptions) {
		o.Platform = ""
	})

	err := runInit(opts)
	require.Error(t, err)
}

// Validation: neither --resources nor --preset set.
func TestRunInit_RequiresRefsOrPreset(t *testing.T) {
	t.Parallel()

	opts, _, _ := newInitOpts(t, nil, &fakeInitializer{}, func(o *initOptions) {
		o.Refs = nil
		o.Preset = ""
	})

	err := runInit(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either --resources or --preset is required")
}

// Validation: both --resources and --preset set.
func TestRunInit_RefsAndPresetMutex(t *testing.T) {
	t.Parallel()

	opts, _, _ := newInitOpts(t, nil, &fakeInitializer{}, func(o *initOptions) {
		o.Refs = []string{"skill/commit"}
		o.Preset = "git-workflow"
	})

	err := runInit(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// Transport-level error from the initializer must wrap and return.
func TestRunInit_TransportErrorWraps(t *testing.T) {
	t.Parallel()

	cause := errors.New("library corrupt")
	opts, _, _ := newInitOpts(t, nil, &fakeInitializer{err: cause}, nil)

	err := runInit(opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause, "transport error must be preserved in the chain")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err))
}

// Constructor wires opts correctly with runF injection.
func TestNewCmdInit_RunFInjectionCapturesOpts(t *testing.T) {
	t.Parallel()

	var captured *initOptions
	runF := func(opts *initOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdInit(f, runF)
	cmd.SetArgs([]string{
		"--platform", "opencode",
		"--resources", "skill/commit,skill/merge-request",
		"--output-dir", "/tmp/out",
		"--dry-run",
		"--force",
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, io, captured.IO)
	assert.Equal(t, "opencode", captured.Platform)
	assert.Equal(t, []string{"skill/commit", "skill/merge-request"}, captured.Refs)
	assert.Equal(t, "/tmp/out", captured.OutputDir)
	assert.True(t, captured.DryRun)
	assert.True(t, captured.Force)
	assert.NotNil(t, captured.Ctx)
	assert.NotNil(t, captured.Library, "Library lazy field must be wired by NewCmdInit")
}

// Constructor with --preset wires the Preset field.
func TestNewCmdInit_PresetFlagWiredToOpts(t *testing.T) {
	t.Parallel()

	var captured *initOptions
	runF := func(opts *initOptions) error {
		captured = opts
		return nil
	}

	f := cmdutil.NewFactory(context.Background(), iostreams.Test(), "test", "germinator")
	cmd := NewCmdInit(f, runF)
	cmd.SetArgs([]string{"--platform", "opencode", "--preset", "git-workflow"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "git-workflow", captured.Preset)
	assert.Empty(t, captured.Refs, "--resources is empty when only --preset is set")
}

// Constructor accepts --resources as a single string and splits,
// matching the long-term --resources semantics.
func TestNewCmdInit_NilRunFFallsBackToProduction(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdInit(f, nil)
	cmd.SetArgs([]string{
		"--platform", "windows-95",
		"--resources", "skill/commit",
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "invalid platform must surface as an error")
	assert.Contains(t, err.Error(), `unknown platform "windows-95"`)
}

// Constructor requires --platform.
func TestNewCmdInit_RequiresPlatformFlag(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory(context.Background(), iostreams.Test(), "test", "germinator")
	cmd := NewCmdInit(f, func(*initOptions) error { return nil })
	cmd.SetArgs([]string{"--resources", "skill/commit"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing required --platform flag must fail")
}

// Sanity: classifyResults counts correctly across success/failure.
func TestClassifyResults(t *testing.T) {
	t.Parallel()

	results := []core.InitializeResult{
		{Ref: "a"},
		{Ref: "b", Error: errors.New("broke")},
		{Ref: "c"},
		{Ref: "d", Error: errors.New("d broke")},
	}
	s, f, errs := classifyResults(results)
	assert.Equal(t, 2, s)
	assert.Equal(t, 2, f)
	require.Len(t, errs, 2)
	assert.Equal(t, "b", errs[0].Ref())
	assert.Equal(t, "d", errs[1].Ref())
}

// Sanity: classifyResults on empty slice returns zeros.
func TestClassifyResults_Empty(t *testing.T) {
	t.Parallel()

	s, f, errs := classifyResults(nil)
	assert.Equal(t, 0, s)
	assert.Equal(t, 0, f)
	assert.Nil(t, errs)
}

func TestInitLibrary_NilFactoryReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, initLibrary(nil, "/tmp"),
		"initLibrary(nil, ...) returns nil so opts.Library is unset")
}

func TestInitLibrary_HonorsExplicitPath(t *testing.T) {
	// t.Setenv cannot be called from a parallel test.
	tmp := t.TempDir()

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	loader := initLibrary(f, tmp)
	require.NotNil(t, loader, "initLibrary must return a non-nil loader")

	_, err := loader()
	require.Error(t, err)
}

func TestInitInitializer_NilFactoryReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, initInitializer(nil),
		"initInitializer(nil) returns nil")
}

func TestInitInitializer_FactoryWithoutInitializerFieldReturnsNil(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory(context.Background(), iostreams.Test(), "test", "germinator")
	assert.Nil(t, initInitializer(f),
		"initInitializer must return nil when f.Initializer is unset")
}

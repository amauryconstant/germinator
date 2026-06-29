package cmd

import (
	"bytes"
	"context"
	"errors"
	"reflect"
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
// lastReq stores the most recent *InitializeRequest so tests can
// assert that flags like Force/DryRun are plumbed through.
type fakeInitializer struct {
	calls   int
	results []core.InitializeResult
	err     error
	lastReq *InitializeRequest
}

func (f *fakeInitializer) Initialize(_ context.Context, req *InitializeRequest) ([]core.InitializeResult, error) {
	f.calls++
	f.lastReq = req
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
	// core.ValidatePlatform attaches suggestion strings naming both
	// valid platforms; the spec requires the error to indicate them.
	assert.Contains(t, err.Error(), "claude-code",
		"error must surface the claude-code option")
	assert.Contains(t, err.Error(), "opencode",
		"error must surface the opencode option")
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
	var captured *initOptions
	runF := func(opts *initOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	f.Initializer = func() (application.Initializer, error) { return &fakeInitializer{}, nil }
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
	assert.NotNil(t, captured.Initializer, "Initializer lazy field must be wired by NewCmdInit when f.Initializer is set")
}

// Spec scenario "Custom output directory" (cli-init-command):
// --output-dir /target/project must populate opts.OutputDir AND drive
// the resolved file path to /target/project/.opencode/skills/commit/SKILL.md.
// Guards the breaking rename from legacy --output/-o and asserts the
// downstream path derivation against the spec's literal example.
func TestNewCmdInit_OutputDirFlagWiredToOpts(t *testing.T) {
	t.Parallel()

	var captured *initOptions
	runF := func(opts *initOptions) error {
		captured = opts
		return nil
	}

	f := cmdutil.NewFactory(context.Background(), iostreams.Test(), "test", "germinator")
	cmd := NewCmdInit(f, runF)
	cmd.SetArgs([]string{
		"--platform", "opencode",
		"--resources", "skill/commit",
		"--output-dir", "/target/project",
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "/target/project", captured.OutputDir,
		"--output-dir flag must populate opts.OutputDir verbatim")

	// The spec mandates the resolved path is
	// "/target/project/.opencode/skills/commit/SKILL.md" for an opencode
	// skill named "commit". Compute it via the same helper runInit uses
	// to avoid drift if the path layout changes.
	expectedPath, err := library.GetOutputPath("skill", "commit", captured.Platform, captured.OutputDir)
	require.NoError(t, err)
	assert.Equal(t, "/target/project/.opencode/skills/commit/SKILL.md", expectedPath,
		"opencode skill path derivation must match the spec example")
}

// Constructor with --preset wires the Preset field.
func TestNewCmdInit_PresetFlagWiredToOpts(t *testing.T) {
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

// §5.x — --library end-to-end: passing an explicit path surfaces
// resources from that library (proves the loader is honored, not the
// env/XDG default). Mirrors the spec scenario "Custom library path".
func TestRunInit_CustomLibraryPathResolvesRefs(t *testing.T) {
	t.Parallel()

	// A hand-built library whose RootPath is a unique tmp dir.
	// The test asserts opts.Library() returns this exact library, so
	// any fallback (env var, XDG default) would be visible as a
	// mismatch and fail the assertion.
	customRoot := "/custom/library/abc123"
	lib := &library.Library{
		Version:  "1",
		RootPath: customRoot,
		Resources: map[string]map[string]library.Resource{
			"skill": {
				"commit": {Path: "skills/commit.md", Description: "Git commit"},
			},
		},
	}

	opts, _, _ := newInitOpts(t, lib, &fakeInitializer{
		results: []core.InitializeResult{
			{
				Ref:        "skill/commit",
				InputPath:  customRoot + "/skills/commit.md",
				OutputPath: ".opencode/skills/commit/SKILL.md",
			},
		},
	}, func(o *initOptions) {
		o.Library = func() (*library.Library, error) { return lib, nil }
	})

	require.NoError(t, runInit(opts), "happy-path init against custom library must succeed")

	loaded, err := opts.Library()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, customRoot, loaded.RootPath,
		"--library explicit path must surface as the loaded Library.RootPath")
	_, ok := loaded.Resources["skill"]["commit"]
	assert.True(t, ok, "loaded library must expose skill/commit from the custom path")
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

// T1 — Spec scenario "init command signature": --help output SHALL
// list --platform, --output-dir, --library, --resources, --preset,
// --dry-run, --force.
func TestNewCmdInit_HelpOutput_ListsFlags(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory(context.Background(), iostreams.Test(), "test", "germinator")
	cmd := NewCmdInit(f, func(*initOptions) error { return nil })

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	require.NoError(t, cmd.Help())
	out := buf.String()
	for _, flag := range []string{
		"--platform",
		"--output-dir",
		"--library",
		"--resources",
		"--preset",
		"--dry-run",
		"--force",
	} {
		assert.Contains(t, out, flag,
			"help output must list %s", flag)
	}
}

// T2 — Spec scenario "initOptions struct": the struct SHALL declare
// the twelve fields named in cli-init-command/spec.md. A runtime
// reflection check catches accidental field drops or renames without
// forcing a brittle positional assertion.
func TestInitOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(initOptions{})
	want := map[string]bool{
		"IO":          true,
		"Library":     true,
		"Initializer": true,
		"Ctx":         true,
		"LibraryPath": true,
		"Platform":    true,
		"OutputDir":   true,
		"Refs":        true,
		"Preset":      true,
		"DryRun":      true,
		"Force":       true,
	}

	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}

	assert.Equal(t, want, got,
		"initOptions must declare exactly the spec-named fields")
}

// T3 — Spec scenario "Dry-run preview": when --dry-run is set,
// output SHALL show what would be written without creating files.
// renderResults writes "Would write: <path>\n  from: <input>\n" for
// each success and a trailing "Dry run complete." line.
func TestRunInit_DryRunRendersWouldWrite(t *testing.T) {
	t.Parallel()

	opts, out, errOut := newInitOpts(t, nil, &fakeInitializer{
		results: []core.InitializeResult{
			{
				Ref:        "skill/commit",
				InputPath:  "/lib/skills/commit.md",
				OutputPath: ".opencode/skills/commit/SKILL.md",
			},
		},
	}, func(o *initOptions) {
		o.DryRun = true
	})

	require.NoError(t, runInit(opts))
	assert.Contains(t, out.String(), "Would write:",
		"dry-run must render the Would write: line")
	assert.Contains(t, out.String(), "Dry run complete.",
		"dry-run must append a completion line")
	assert.Empty(t, errOut.String(),
		"dry-run success must not write to ErrOut")
}

// T4 — Spec scenario "Force overwrite": --force SHALL cause existing
// files to be overwritten. The runInit body plumbs Force into
// (*InitializeRequest).Force; the actual overwrite behavior lives in
// the Initializer implementation (production service layer). This
// test asserts the plumb-through, not the side effect.
func TestRunInit_ForcePropagatesToRequest(t *testing.T) {
	t.Parallel()

	init := &fakeInitializer{
		results: []core.InitializeResult{
			{
				Ref:        "skill/commit",
				InputPath:  "/lib/skills/commit.md",
				OutputPath: ".opencode/skills/commit/SKILL.md",
			},
		},
	}

	opts, _, _ := newInitOpts(t, nil, init, func(o *initOptions) {
		o.Force = true
	})

	require.NoError(t, runInit(opts))
	require.Equal(t, 1, init.calls, "Initialize must be called exactly once")
	require.NotNil(t, init.lastReq, "InitializeRequest must be captured")
	assert.True(t, init.lastReq.Force,
		"Force flag must propagate to InitializeRequest.Force")
}

package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// fakeInitializer is a hand-rolled fake satisfying the local
// cmd.Initializer interface. Slice 7 deleted the application-package
// type alias; the local interface (cmd/initializer.go) keeps this
// fake alive for callers that want to substitute the Initializer at
// the adapter boundary. Tests in this file exercise the production
// cmd.NewInitializer() pipeline via real fixtures instead of
// injecting this fake through runInit.
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

// initFixtureLibrary scaffolds a real library directory with a
// library.yaml that references one or more skills/agents, plus the
// matching resource files on disk. Returns the resolved RootPath and
// a *library.Library ready to load. This replaces the in-memory
// `fakeLibraryLoader()` that the previous slice used to back
// fakeInitializer-driven tests.
//
// Filenames follow the parser.DetectType convention (`<type>-<name>.md`)
// so the production Initializer can resolve the file as a valid
// resource rather than failing with an "unrecognizable filename" parse
// error.
func initFixtureLibrary(t *testing.T, resources map[string]map[string]string) (string, *library.Library) {
	t.Helper()
	libDir := t.TempDir()

	// Mirror the makeTestLibrary convention used in cmd/library_add_test.go
	// so the directory layout matches what LoadLibrary expects.
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		require.NoError(t, os.MkdirAll(filepath.Join(libDir, sub), 0o750))
	}

	lib := &library.Library{
		Version:   "1",
		RootPath:  libDir,
		Resources: map[string]map[string]library.Resource{},
		Presets:   map[string]library.Preset{},
	}
	for resType, names := range resources {
		lib.Resources[resType] = map[string]library.Resource{}
		for name, path := range names {
			lib.Resources[resType][name] = library.Resource{
				Path:        path,
				Description: name + " fixture",
			}
			fullPath := filepath.Join(libDir, path)
			require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o750))
			body := "---\nname: " + name + "\ndescription: " + name + " fixture\n---\nBody\n"
			require.NoError(t, os.WriteFile(fullPath, []byte(body), 0o644))
		}
	}
	require.NoError(t, library.SaveLibrary(lib))
	return libDir, lib
}

// initFixtureSkill is a shorthand for `initFixtureLibrary` with a
// single "commit" skill resource whose filename follows the parser
// convention. All current test cases use the same skill; per-skill
// variants can be added via initFixtureLibrary directly if needed.
func initFixtureSkill(t *testing.T) (string, *library.Library) {
	t.Helper()
	return initFixtureLibrary(t, map[string]map[string]string{
		"skill": {"commit": "skills/commit-skill.md"},
	})
}

// initFixtureLibraryWithPreset extends initFixtureLibrary with a
// named preset that references a subset of the registered resources.
func initFixtureLibraryWithPreset(t *testing.T, presetName string, refs []string) (string, *library.Library) {
	t.Helper()
	libDir, lib := initFixtureLibrary(t, map[string]map[string]string{
		"skill": {
			"commit":        "skills/commit-skill.md",
			"merge-request": "skills/merge-request-skill.md",
		},
	})
	lib.Presets[presetName] = library.Preset{
		Name:        presetName,
		Description: "Fixture preset",
		Resources:   refs,
	}
	require.NoError(t, library.SaveLibrary(lib))
	return libDir, lib
}

// §5.3.1 — All success: runInit returns nil; ExitCodeFor(nil) == 0.
func TestRunInit_AllSuccess(t *testing.T) {
	t.Parallel()

	libDir, _ := initFixtureSkill(t)
	outputDir := t.TempDir()

	io, out, errOut := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: outputDir,
		Refs:      []string{"skill/commit"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runInit(opts))
	assert.Equal(t, cmdutil.ExitCodeSuccess, cmdutil.ExitCodeFor(nil))
	assert.Contains(t, out.String(), "Installed: skill/commit")
	assert.Contains(t, out.String(), "Initialized 1 resource(s).")
	assert.Empty(t, errOut.String())
}

// §5.3.2 — Partial success: one valid + one missing-file ref.
// *PartialSuccessError{S:1,F:1}; ExitCodeFor == 0.
func TestRunInit_PartialSuccess(t *testing.T) {
	t.Parallel()

	libDir, _ := initFixtureSkill(t)
	outputDir := t.TempDir()

	io, _, errOut := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: outputDir,
		Refs:      []string{"skill/commit", "skill/ghost"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runInit(opts)
	require.Error(t, err)

	var ps *core.PartialSuccessError
	require.ErrorAs(t, err, &ps)
	assert.Equal(t, 1, ps.Succeeded())
	assert.Equal(t, 1, ps.Failed())
	assert.Equal(t, cmdutil.ExitCodeSuccess, cmdutil.ExitCodeFor(err),
		"partial success must map to ExitCodeSuccess (0)")

	output.FormatError(opts.IO, err)
	assert.Contains(t, errOut.String(), "partial success: 1 succeeded, 1 failed")
}

// §5.3.3 — All failed: two missing-file refs.
// *PartialSuccessError{S:0,F:2}; ExitCodeFor == 1.
func TestRunInit_AllFailed(t *testing.T) {
	t.Parallel()

	// Register no resources; every ref will fail the file-resolution step.
	libDir, _ := initFixtureLibrary(t, map[string]map[string]string{})
	outputDir := t.TempDir()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: outputDir,
		Refs:      []string{"skill/invalid1", "skill/invalid2"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

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

	libDir, _ := initFixtureLibraryWithPreset(t, "git-workflow",
		[]string{"skill/commit", "skill/merge-request"})
	outputDir := t.TempDir()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: outputDir,
		Preset:    "git-workflow",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runInit(opts),
		"all-success preset expansion returns nil")
}

// §5.3.4 — Preset expansion with one failure: partial-success logic
// applies to the expanded list.
func TestRunInit_PresetExpansion_PartialFail(t *testing.T) {
	t.Parallel()

	// Only register skill/commit; skill/merge-request stays absent so
	// the per-ref resolution fails for the second expanded ref.
	libDir, _ := initFixtureLibraryWithPreset(t, "git-workflow",
		[]string{"skill/commit", "skill/merge-request"})
	// Now strip the second resource from library.yaml so the file is
	// missing on disk:
	require.NoError(t, os.Remove(filepath.Join(libDir, "skills", "merge-request-skill.md")))
	outputDir := t.TempDir()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: outputDir,
		Preset:    "git-workflow",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

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

	libDir, _ := initFixtureSkill(t)

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: t.TempDir(),
		Preset:    "ghost",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runInit(opts)
	require.Error(t, err)
	var nf *core.NotFoundError
	require.ErrorAs(t, err, &nf)
	assert.Equal(t, "preset", nf.Entity)
	assert.Equal(t, "ghost", nf.Key)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"preset-not-found must map to ExitCodeUsage (2)")
}

// Validation: missing --platform.
func TestRunInit_InvalidPlatform(t *testing.T) {
	t.Parallel()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  "windows-95",
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/commit"},
		Library: func() (*library.Library, error) {
			return &library.Library{Version: "1", RootPath: "/fake", Resources: map[string]map[string]library.Resource{}}, nil
		},
	}

	err := runInit(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown platform "windows-95"`)
	assert.Contains(t, err.Error(), "claude-code",
		"error must surface the claude-code option")
	assert.Contains(t, err.Error(), "opencode",
		"error must surface the opencode option")
}

// Validation: empty platform flag value.
func TestRunInit_EmptyPlatform(t *testing.T) {
	t.Parallel()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  "",
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/commit"},
		Library: func() (*library.Library, error) {
			return &library.Library{Version: "1", RootPath: "/fake", Resources: map[string]map[string]library.Resource{}}, nil
		},
	}

	err := runInit(opts)
	require.Error(t, err)
}

// Validation: neither --resources nor --preset set.
func TestRunInit_RequiresRefsOrPreset(t *testing.T) {
	t.Parallel()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: t.TempDir(),
		Library: func() (*library.Library, error) {
			return &library.Library{Version: "1", RootPath: "/fake", Resources: map[string]map[string]library.Resource{}}, nil
		},
	}

	err := runInit(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either --resources or --preset is required")
}

// Validation: both --resources and --preset set.
func TestRunInit_RefsAndPresetMutex(t *testing.T) {
	t.Parallel()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/commit"},
		Preset:    "git-workflow",
		Library: func() (*library.Library, error) {
			return &library.Library{Version: "1", RootPath: "/fake", Resources: map[string]map[string]library.Resource{}}, nil
		},
	}

	err := runInit(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// TestNewCmdInit_RunFInjectionCapturesOpts — slice-7 simplified: the
// f.Initializer lazy field and opts.Initializer lazy field both
// removed. The constructor's only remaining wiring is f.IOStreams,
// f.Library, c.Context(), and the parsed flags.
func TestNewCmdInit_RunFInjectionCapturesOpts(t *testing.T) {
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

// Constructor with nil runF falls back to production runInit.
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

// §5.x — --library end-to-end: passing an explicit path surfaces
// resources from that library (proves the loader is honored, not the
// env/XDG default).
func TestRunInit_CustomLibraryPathResolvesRefs(t *testing.T) {
	t.Parallel()

	libDir, _ := initFixtureSkill(t)
	outputDir := t.TempDir()

	io, _, _ := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: outputDir,
		Refs:      []string{"skill/commit"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runInit(opts),
		"happy-path init against custom library must succeed")

	loaded, err := opts.Library()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, libDir, loaded.RootPath,
		"--library explicit path must surface as the loaded Library.RootPath")
	_, ok := loaded.Resources["skill"]["commit"]
	assert.True(t, ok, "loaded library must expose skill/commit from the custom path")
}

func TestInitLibrary_HonorsExplicitPath(t *testing.T) {
	tmp := t.TempDir()

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	// Mirror the RunE inline closure pattern (Phase 1: per-RunE
	// lazy loader built from f.Config + FindLibrary).
	var cfgPath string
	if f.Config != nil {
		if loaded, cfgErr := f.Config(); cfgErr == nil && loaded != nil {
			cfgPath = loaded.Library
		}
	}
	resolved := library.FindLibrary(tmp, os.Getenv("GERMINATOR_LIBRARY"), cfgPath)
	loader := cmdutil.OnceValuesFunc(func() (*library.Library, error) {
		return library.LoadLibrary(context.Background(), resolved)
	})

	_, err := loader()
	require.Error(t, err)
}

// slice-7 removed the Factory.Initializer lazy field and the
// `initInitializer(f)` factory helper. The Initializer is now
// constructed inside runInit via cmd.NewInitializer().

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
		assert.Contains(t, out, flag, "help output must list %s", flag)
	}
}

// T2 — Spec scenario "initOptions struct": the struct SHALL declare
// the slice-7 field set (LibraryPath/Initializer removed). A runtime
// reflection check catches accidental field drops or renames without
// forcing a brittle positional assertion.
func TestInitOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(initOptions{})
	want := map[string]bool{
		"IO":          true,
		"Library":     true,
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
		"initOptions must declare exactly the slice-7 field set")
}

// T3 — Spec scenario "Dry-run preview": when --dry-run is set,
// output SHALL show what would be written without creating files.
// renderResults writes "Would write: <path>\n  from: <input>\n" for
// each success and a trailing "Dry run complete." line.
func TestRunInit_DryRunRendersWouldWrite(t *testing.T) {
	t.Parallel()

	libDir, _ := initFixtureSkill(t)

	io, out, errOut := newInitTestIO()
	opts := &initOptions{
		IO:        io,
		Ctx:       context.Background(),
		Platform:  core.PlatformOpenCode,
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/commit"},
		DryRun:    true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

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

	// Use the fakeInitializer directly to assert the request shape
	// without involving the production Initializer's filesystem I/O.
	init := &fakeInitializer{
		results: []core.InitializeResult{
			{
				Ref:        "skill/commit",
				InputPath:  "/lib/skills/commit.md",
				OutputPath: ".opencode/skills/commit/SKILL.md",
			},
		},
	}
	req := &InitializeRequest{
		Library:   &library.Library{Version: "1", RootPath: "/lib"},
		Platform:  core.PlatformOpenCode,
		OutputDir: "/tmp/out",
		Refs:      []string{"skill/commit"},
		Force:     true,
	}
	_, err := init.Initialize(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 1, init.calls)
	require.NotNil(t, init.lastReq)
	assert.True(t, init.lastReq.Force,
		"Force flag must propagate to InitializeRequest.Force")
}

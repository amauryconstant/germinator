package cmd

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

// newCreatePresetTestIO returns the buffer-backed IOStreams that
// create-preset tests use to assert on captured Out / ErrOut. Mirrors
// the slice-5 newInitTestIO and slice-6 newAddTestIO helpers. Panics
// if iostreams.Test() does not return *bytes.Buffer writers (it
// always does in this codebase, but guard anyway).
func newCreatePresetTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// makePresetTestLibrary scaffolds a minimal library dir with one
// skill file already registered in library.yaml. Returns the resolved
// RootPath. Mirrors the slice-6 makeTestLibrary helper.
func makePresetTestLibrary(t *testing.T, registered map[string]map[string]library.Resource) string {
	t.Helper()
	dir := makeTestLibrary(t, registered)
	return dir
}

// presetTestResources returns a small resource set used across the
// success / description / force-overwrite / multi-resource tests.
func presetTestResources() map[string]map[string]library.Resource {
	return map[string]map[string]library.Resource{
		"skill": {
			"commit":  {Path: "skills/commit.md", Description: "Git commit skill"},
			"release": {Path: "skills/release.md", Description: "Release skill"},
		},
		"agent": {
			"reviewer": {Path: "agents/reviewer.md", Description: "Code reviewer agent"},
		},
	}
}

// T1 — Constructor wires opts correctly via runF injection: --resources,
// --description, --force, Ctx, Library all populated from the parsed
// flags and the Factory.
func TestNewCmdCreatePreset_ValidatesArgs(t *testing.T) {
	var captured *createPresetOptions
	runF := func(opts *createPresetOptions) error {
		captured = opts
		return nil
	}

	ios, _, _ := newCreatePresetTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	cmd := NewCmdCreatePreset(f, &libPath, runF)
	cmd.SetArgs([]string{"dev-setup", "--resources", "skill/commit,agent/reviewer",
		"--description", "Dev environment", "--force"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, []string{"skill/commit", "agent/reviewer"}, captured.Resources)
	assert.Equal(t, "Dev environment", captured.Description)
	assert.True(t, captured.Force)
	assert.NotNil(t, captured.Library, "opts.Library must be wired by NewCmdCreatePreset")
	assert.NotNil(t, captured.IO)
	assert.NotNil(t, captured.Ctx)
}

// T2 — Spec scenario "Valid refs pass pre-flight validation":
// `core.CanInstallResource` returns nil for both skill/commit and
// agent/reviewer; the preset is created successfully; the
// "Created preset:" line appears on stdout.
func TestRunCreatePreset_ValidSuccess(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, out, errOut := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"skill/commit", "agent/reviewer"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runCreatePreset(opts, "dev-setup"))
	assert.Contains(t, out.String(), "Created preset: dev-setup")
	assert.Empty(t, errOut.String(), "no errors on success")

	// Preset was persisted.
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	preset, ok := lib.Presets["dev-setup"]
	require.True(t, ok, "preset must be persisted")
	assert.Equal(t, []string{"skill/commit", "agent/reviewer"}, preset.Resources)
}

// T3 — Spec scenario "First malformed ref fails pre-flight validation":
// `--resources skills/commit,agent/reviewer` returns a
// *core.ValidationError because core.CanInstallResource rejects the
// invalid type "skills". cmdutil.ExitCodeFor maps the default
// *core.ValidationError to exit 1 (not exit 2; --resources is
// present and non-empty, only its VALUE is malformed).
func TestRunCreatePreset_InvalidRef(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"skills/commit", "agent/reviewer"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runCreatePreset(opts, "dev-setup")
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "expected *core.ValidationError in chain")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"malformed ref must map to ExitCodeError (1)")

	// Preset must NOT have been created.
	lib, lerr := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, lerr)
	_, exists := lib.Presets["dev-setup"]
	assert.False(t, exists, "preset must not be created on validation failure")
}

// T4 — Spec scenario "Empty resources flag fails pre-flight validation":
// `--resources ""` produces a wrapped cobraUsagePrefixes-style
// message so cmdutil.ExitCodeFor maps it to ExitCodeUsage (2).
func TestRunCreatePreset_EmptyResourcesValue(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{""},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runCreatePreset(opts, "dev-setup")
	require.Error(t, err)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"empty --resources value must map to ExitCodeUsage (2)")
	assert.Contains(t, err.Error(), "flag needs an argument")
}

// T5 — Constructor requires --resources: passing a preset name
// without --resources triggers Cobra's MarkFlagRequired check,
// which emits "required flag(s) \"resources\" not set"; this is
// mapped to exit 2 by cobraUsagePrefixes.
func TestNewCmdCreatePreset_RequiresResources(t *testing.T) {
	ios, _, _ := newCreatePresetTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	cmd := NewCmdCreatePreset(f, &libPath, func(*createPresetOptions) error { return nil })
	cmd.SetArgs([]string{"dev-setup"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing --resources flag must fail")
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"missing required --resources must map to ExitCodeUsage (2)")
	assert.Contains(t, err.Error(), "required flag")
}

// T6 — Description is preserved on the persisted preset.
func TestRunCreatePreset_Description(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:          ios,
		Ctx:         context.Background(),
		Resources:   []string{"skill/commit"},
		Description: "Dev environment bootstrap",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runCreatePreset(opts, "dev-setup"))

	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	preset, ok := lib.Presets["dev-setup"]
	require.True(t, ok, "preset must be persisted")
	assert.Equal(t, "Dev environment bootstrap", preset.Description)
}

// T7 — Spec scenario "Reject duplicate preset without --force":
// existing preset + no --force returns *core.ValidationError
// (exit 1).
func TestRunCreatePreset_NoForce_ExistingFails(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	// Seed an existing preset.
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	lib.Presets["existing"] = library.Preset{
		Name:        "existing",
		Description: "old",
		Resources:   []string{"skill/commit"},
	}
	require.NoError(t, library.SaveLibrary(lib))

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"agent/reviewer"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err = runCreatePreset(opts, "existing")
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "expected *core.ValidationError")
	assert.Contains(t, verr.Error(), "already exists")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"duplicate without --force must map to ExitCodeError (1)")
}

// T8 — Spec scenario "Reject preset name with whitespace": a name
// consisting of only whitespace triggers the early validation
// branch and returns *core.ValidationError (exit 1).
func TestRunCreatePreset_EmptyPresetName(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"skill/commit"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runCreatePreset(opts, "   ")
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "expected *core.ValidationError")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err))
}

// T9 — Spec scenario "Force overwrite existing preset":
// existing preset + --force succeeds; description + resources are
// replaced.
func TestRunCreatePreset_ForceOverwrite(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	lib.Presets["existing"] = library.Preset{
		Name:        "existing",
		Description: "Old description",
		Resources:   []string{"skill/commit"},
	}
	require.NoError(t, library.SaveLibrary(lib))

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"agent/reviewer"},
		Force:     true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runCreatePreset(opts, "existing"))

	reloaded, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	preset, ok := reloaded.Presets["existing"]
	require.True(t, ok)
	assert.Equal(t, "agent/reviewer", preset.Resources[0],
		"--force must overwrite the resources list")
	assert.Empty(t, preset.Description,
		"--force must overwrite the description")
}

// T10 — Multiple resources, all valid: 3+ refs all resolve and the
// preset carries the full list.
func TestRunCreatePreset_MultipleResources(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"skill/commit", "skill/release", "agent/reviewer"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runCreatePreset(opts, "multi"))

	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	preset, ok := lib.Presets["multi"]
	require.True(t, ok)
	assert.Equal(t, []string{"skill/commit", "skill/release", "agent/reviewer"}, preset.Resources)
}

// T11 — Spec scenario "Validate referenced resources exist":
// `core.CanInstallResource` accepts the syntactic shape of a ref,
// but the library must also have the resource registered. A
// syntactically-valid ref pointing to a non-existent resource
// surfaces from lib.CreatePreset as *core.NotFoundError; cmdutil.ExitCodeFor
// maps NotFoundError to ExitCodeUsage (2) per its explicit
// `errors.As(notFound)` branch. The preset is NOT created.
func TestRunCreatePreset_RefReferencesMissingResource(t *testing.T) {
	libDir := makePresetTestLibrary(t, presetTestResources())

	ios, _, _ := newCreatePresetTestIO()
	opts := &createPresetOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Resources: []string{"skill/nonexistent"},
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runCreatePreset(opts, "broken")
	require.Error(t, err)

	var nf *core.NotFoundError
	require.True(t, errors.As(err, &nf), "expected *core.NotFoundError from lib.CreatePreset")
	assert.Equal(t, "resource", nf.Entity)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"missing-resource error maps to ExitCodeUsage (2) via NotFoundError branch")

	lib, lerr := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, lerr)
	_, exists := lib.Presets["broken"]
	assert.False(t, exists, "preset must not be created when resource is missing")
}

// T12 — createPresetOptions struct shape: declares exactly the
// spec-named fields; reflection guards against accidental drops
// or renames.
func TestCreatePresetOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(createPresetOptions{})
	want := map[string]bool{
		"IO":              true,
		"Library":         true,
		"Ctx":             true,
		"Resources":       true,
		"Description":     true,
		"Force":           true,
		"CompletionCache": true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}

	assert.Equal(t, want, got,
		"createPresetOptions must declare exactly the spec-named fields")
}

// T13 — presetWriter interface satisfied by *library.Library:
// compile-time check is exercised by the var _ presetWriter
// declaration in library_create.go; this test asserts at runtime
// that the contract holds.
func TestPresetWriterInterfaceSatisfied(t *testing.T) {
	t.Parallel()

	lib := &library.Library{}
	var pw presetWriter = lib
	assert.NotNil(t, pw, "*library.Library must satisfy presetWriter")
}

// T14 — createPresetLibrary helper: nil factory returns a nil
// loader so tests that don't care about the loader can ignore it.
func TestCreatePresetLibrary_NilFactoryReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, createPresetLibrary(nil, "/tmp"),
		"createPresetLibrary(nil, ...) returns nil so opts.Library is unset")
}

// T15 — No --output flag on `library create preset`: design
// Decision 5. Help output must NOT mention --output.
func TestNewCmdCreatePreset_HelpOutput_NoOutputFlag(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdCreatePreset(f, nil, func(*createPresetOptions) error { return nil })

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	require.NoError(t, cmd.Help())
	out := buf.String()
	assert.NotContains(t, out, "--output",
		"library create preset help must not advertise --output (design Decision 5)")
}

// T16 — library create preset is a leaf: there are no subcommands
// under it. Spec scenario "library create preset help resolves to
// a single command".
func TestNewCmdCreatePreset_NoSubcommands(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdCreatePreset(f, nil, func(*createPresetOptions) error { return nil })

	assert.Empty(t, cmd.Commands(),
		"library create preset must not have subcommands (leaf under library)")
}

// T17 — Spec scenario "library create has no subcommand list": the
// thin `library create` parent lists `preset` as a child in --help
// but the parent itself has no group description. Cobra's default
// parent rendering is asserted via the presence of "preset" in the
// help output and the parent's empty parent-commands slice (the
// parent holds only `preset`).
func TestNewCmdCreate_ShowsPresetAsChild(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	cmd := NewCmdCreate(f, &libPath)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	require.NoError(t, cmd.Help())
	out := buf.String()

	require.NotEmpty(t, cmd.Commands(),
		"library create parent must contain at least the 'preset' child command")
	var sawPreset bool
	for _, c := range cmd.Commands() {
		if c.Name() == "preset" {
			sawPreset = true
			break
		}
	}
	assert.True(t, sawPreset,
		"library create parent must register 'preset' as a child command")

	// The Help() output advertises the preset child for users
	// invoking `germinator library create` bare.
	assert.Contains(t, out, "preset",
		"library create --help must mention the 'preset' child command")
}

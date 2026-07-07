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
)

// newLibraryInitTestIO returns the buffer-backed IOStreams that
// library-init tests use to assert on captured Out / ErrOut. Renamed
// from the slice-5/6 newInitTestIO pattern to avoid collision with
// the top-level init command's test helper in init_test.go.
func newLibraryInitTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// newLibraryInitOpts builds a *libraryInitOptions with iostreams.Test()
// buffers and t.Context(). mut may be nil or a function that mutates
// the opts struct after the defaults are set.
func newLibraryInitOpts(t *testing.T, mut func(*libraryInitOptions)) *libraryInitOptions {
	t.Helper()
	ios, _, _ := newLibraryInitTestIO()
	opts := &libraryInitOptions{IO: ios, Ctx: t.Context()}
	if mut != nil {
		mut(opts)
	}
	return opts
}

// makeStubLibrary creates an empty library directory so the
// existing-library error path can be exercised (library.Exists
// returns true and CreateLibrary refuses to overwrite without
// --force).
func makeStubLibrary(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0o750))
}

// T1 — Constructor wires opts correctly via runF injection. Each
// flag is parsed and propagated to the captured opts.
func TestNewCmdLibraryInit_ConstructorWiresOpts(t *testing.T) {
	var captured *libraryInitOptions
	runF := func(opts *libraryInitOptions) error {
		captured = opts
		return nil
	}

	ios, _, _ := newLibraryInitTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdLibraryInit(f, runF)
	cmd.SetArgs([]string{
		"--path", "/tmp/wired-lib",
		"--force",
		"--dry-run",
		"--output", "json",
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/wired-lib", captured.Path)
	assert.True(t, captured.Force, "opts.Force must be true")
	assert.True(t, captured.DryRun, "opts.DryRun must be true")
	assert.Equal(t, "json", captured.Output)
	assert.NotNil(t, captured.IO, "opts.IO must be wired from f.IOStreams")
	assert.NotNil(t, captured.Ctx, "opts.Ctx must be wired from c.Context()")
}

// T2 — Happy path + --path flag preservation (spec scenario
// "--path flag preserved"). Verifies the library tree is created
// at the user-supplied path and is loadable.
func TestRunLibraryInit_HappyPath(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "happy-lib")

	opts := newLibraryInitOpts(t, func(o *libraryInitOptions) {
		o.Path = libPath
	})
	require.NoError(t, runLibraryInit(opts))

	assert.True(t, library.Exists(libPath), "library directory must exist")
	assert.True(t, library.YAMLExists(libPath), "library.yaml must exist")
	for _, dir := range []string{"skills", "agents", "commands", "memory"} {
		info, err := os.Stat(filepath.Join(libPath, dir))
		require.NoError(t, err, "resource directory %s must exist", dir)
		assert.True(t, info.IsDir(), "%s must be a directory", dir)
	}
	lib, err := library.LoadLibrary(t.Context(), libPath)
	require.NoError(t, err, "created library must be loadable")
	assert.Equal(t, "1", lib.Version)
}

// T3 — --force overwrites an existing library (spec scenario
// "--force overwrites existing library"). The pre-existing
// library.yaml is replaced with the default version "1" content.
func TestRunLibraryInit_ForceOverwrite(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "force-lib")
	makeStubLibrary(t, libPath)
	oldYAML := `version: "0"
resources:
  skill: {}
`
	require.NoError(t, os.WriteFile(filepath.Join(libPath, "library.yaml"), []byte(oldYAML), 0o644))

	opts := newLibraryInitOpts(t, func(o *libraryInitOptions) {
		o.Path = libPath
		o.Force = true
	})
	require.NoError(t, runLibraryInit(opts))

	lib, err := library.LoadLibrary(t.Context(), libPath)
	require.NoError(t, err)
	assert.Equal(t, "1", lib.Version, "--force must rewrite the library.yaml version")
}

// T4 — Existing library without --force returns a typed
// *core.FileError (the error CreateLibrary emits when the
// directory already exists). cmdutil.ExitCodeFor maps it to
// ExitCodeError (1).
func TestRunLibraryInit_ExistingWithoutForceFails(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "exists-lib")
	makeStubLibrary(t, libPath)

	err := runLibraryInit(newLibraryInitOpts(t, func(o *libraryInitOptions) {
		o.Path = libPath
	}))
	require.Error(t, err)

	var fileErr *core.FileError
	require.True(t, errors.As(err, &fileErr), "expected *core.FileError in chain")
	assert.Contains(t, fileErr.Message(), "already exists")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"existing-library failure must map to ExitCodeError (1)")
}

// T5 — --dry-run previews without modifying the filesystem. The
// library directory MUST NOT be created (spec scenario "--dry-run
// previews without changes").
func TestRunLibraryInit_DryRun_DoesNotCreate(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "dryrun-lib")

	require.NoError(t, runLibraryInit(newLibraryInitOpts(t, func(o *libraryInitOptions) {
		o.Path = libPath
		o.DryRun = true
	})))

	assert.False(t, library.Exists(libPath),
		"dry-run must not create the library directory")
}

// T6 — Default plain output writes the confirmation line to
// stdout (opts.IO.Out) per the spec scenario "Default plain output".
func TestRunLibraryInit_DefaultPlainOutput(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "plain-lib")
	ios, out, _ := newLibraryInitTestIO()

	require.NoError(t, runLibraryInit(&libraryInitOptions{
		IO: ios, Ctx: t.Context(), Path: libPath,
	}))

	got := out.String()
	assert.Contains(t, got, "Library created successfully at:",
		"plain success line must appear on stdout")
	assert.Contains(t, got, libPath)
}

// T7 — Dry-run plain output writes the "Dry run complete"
// confirmation; the library's own "Would create" preview block
// is printed separately to os.Stdout (legacy behavior preserved).
func TestRunLibraryInit_DryRunPlainOutput(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "dryrun-plain-lib")
	ios, out, _ := newLibraryInitTestIO()

	require.NoError(t, runLibraryInit(&libraryInitOptions{
		IO: ios, Ctx: t.Context(), Path: libPath, DryRun: true,
	}))

	assert.Contains(t, out.String(), "Dry run complete - no changes made")
}

// T8 — --output json produces a JSON payload with the three
// spec-named keys. Created is true for non-dry-run success;
// for dry-run it flips to false (spec scenario "For dry-run in
// json/table: emit a 'would create' preview payload").
func TestRunLibraryInit_JSONOutput(t *testing.T) {
	tests := []struct {
		name       string
		dryRun     bool
		wantTokens []string
	}{
		{"success", false, []string{`"created": true`}},
		{"dry-run preview", true, []string{`"dryRun": true`, `"created": false`}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			libPath := filepath.Join(t.TempDir(), "json-"+tt.name)
			ios, out, _ := newLibraryInitTestIO()

			require.NoError(t, runLibraryInit(&libraryInitOptions{
				IO: ios, Ctx: t.Context(), Path: libPath, DryRun: tt.dryRun, Output: "json",
			}))

			j := out.String()
			assert.Contains(t, j, `"path"`)
			assert.Contains(t, j, `"dryRun"`)
			assert.Contains(t, j, `"created"`)
			assert.Contains(t, j, libPath)
			for _, tok := range tt.wantTokens {
				assert.Contains(t, j, tok)
			}
		})
	}
}

// T9 — --output table renders the row with the three spec columns
// (PATH, DRYRUN, CREATED) on stdout.
func TestRunLibraryInit_TableOutput(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "table-lib")
	ios, out, _ := newLibraryInitTestIO()

	require.NoError(t, runLibraryInit(&libraryInitOptions{
		IO: ios, Ctx: t.Context(), Path: libPath, Output: "table",
	}))

	got := out.String()
	assert.NotEmpty(t, got, "table output must write at least the header row")
	assert.Contains(t, got, "PATH")
	assert.Contains(t, got, "DRYRUN")
	assert.Contains(t, got, "CREATED")
	assert.Contains(t, got, libPath)
}

// T10 — Verbosef output: with opts.IO.Verbose=true the
// "Creating library at: <path>" line appears on stderr (ErrOut).
// This is the canonical slice-2+ verbose pattern via IOStreams.
func TestRunLibraryInit_VerbosefOutput(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "verbose-lib")
	ios, _, errOut := newLibraryInitTestIO()
	ios.Verbose = true

	require.NoError(t, runLibraryInit(&libraryInitOptions{
		IO: ios, Ctx: t.Context(), Path: libPath,
	}))

	got := errOut.String()
	assert.Contains(t, got, "Creating library at:")
	assert.Contains(t, got, libPath)
}

// T11 — Default path resolution: when opts.Path is empty,
// library.DefaultLibraryPath() is used. We pin the resolution via
// XDG_DATA_HOME so the test is hermetic and platform-independent.
func TestRunLibraryInit_DefaultPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	want := filepath.Join(tmp, "germinator", "library")

	require.NoError(t, runLibraryInit(newLibraryInitOpts(t, nil)))
	assert.True(t, library.Exists(want),
		"default-path init must create the XDG-resolved library at %s", want)
}

// T12 — Cancellation: opts.Ctx is checked by library.Init at entry;
// a pre-cancelled context returns a wrapped ctx.Err().
func TestRunLibraryInit_Cancellation(t *testing.T) {
	libPath := filepath.Join(t.TempDir(), "cancel-lib")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := runLibraryInit(newLibraryInitOpts(t, func(o *libraryInitOptions) {
		o.Ctx = ctx
		o.Path = libPath
	}))
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled),
		"cancelled context must surface as wrapped ctx.Err()")
}

// T13 — Spec scenario "Legacy --json flag is rejected": invoking
// --json on the new command returns a Cobra usage error mapped
// to ExitCodeUsage (2).
func TestRunLibraryInit_RejectsLegacyJSONFlag(t *testing.T) {
	ios, _, errOut := newLibraryInitTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdLibraryInit(f, nil)
	cmd.SetArgs([]string{"--json"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(errOut)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"unknown --json flag must map to ExitCodeUsage (2)")
	assert.Contains(t, errOut.String(), "json",
		"stderr must mention the rejected --json flag")
}

// T14 — libraryInitOptions struct shape: declares exactly the
// spec-named fields; reflection guards against accidental drops
// or renames (matches the slice-6 struct-shape tests in
// library_add_test.go / library_create_test.go).
func TestLibraryInitOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(libraryInitOptions{})
	want := map[string]bool{
		"IO":              true,
		"Ctx":             true,
		"Path":            true,
		"Force":           true,
		"DryRun":          true,
		"Output":          true,
		"CompletionCache": true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}
	assert.Equal(t, want, got,
		"libraryInitOptions must declare exactly the spec-named fields")
}

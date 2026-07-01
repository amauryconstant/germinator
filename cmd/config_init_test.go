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
)

// newConfigInitTestIO returns the buffer-backed IOStreams that
// config-init tests use to assert on captured Out / ErrOut. Named
// with a configInit prefix to avoid collision with other init test
// helpers (cmd/init_test.go's newInitTestIO, cmd/library_init_test.go's
// newLibraryInitTestIO).
func newConfigInitTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// newConfigInitOpts builds a *configInitOptions with iostreams.Test()
// buffers and t.Context(). mut may be nil or a function that mutates
// the opts struct after the defaults are set.
func newConfigInitOpts(t *testing.T, mut func(*configInitOptions)) *configInitOptions {
	t.Helper()
	ios, _, _ := newConfigInitTestIO()
	opts := &configInitOptions{IO: ios, Ctx: t.Context()}
	if mut != nil {
		mut(opts)
	}
	return opts
}

// T1 — Constructor wires opts correctly via runF injection. Each
// flag is parsed and propagated to the captured opts (spec scenario
// "config init supports runF injection").
func TestNewCmdConfigInit_ConstructorWiresOpts(t *testing.T) {
	var captured *configInitOptions
	runF := func(opts *configInitOptions) error {
		captured = opts
		return nil
	}

	ios, _, _ := newConfigInitTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdConfigInit(f, runF)
	cmd.SetArgs([]string{
		"--output-path", "/tmp/wired",
		"--force",
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/wired", captured.OutputPath)
	assert.True(t, captured.Force, "opts.Force must be true")
	assert.NotNil(t, captured.IO, "opts.IO must be wired from f.IOStreams")
	assert.NotNil(t, captured.Ctx, "opts.Ctx must be wired from c.Context()")
}

// T2 — Happy path (spec scenario "Init creates config at custom
// location"): the file is created at the user-supplied path and
// contains the expected comment header + completion table.
func TestRunConfigInit_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")

	require.NoError(t, runConfigInit(newConfigInitOpts(t, func(o *configInitOptions) {
		o.OutputPath = path
	})))

	info, err := os.Stat(path)
	require.NoError(t, err, "config file must exist")
	assert.False(t, info.IsDir())

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(got)
	assert.Contains(t, content, "# Germinator configuration")
	assert.Contains(t, content, "[completion]")
}

// T3 — Golden file test (spec scenario "Init produces byte-identical
// output"): the generated file must be byte-identical to the
// pre-change build's output recorded in testdata/config_init_default.golden.
// This pins the scaffolded content against accidental drift.
func TestRunConfigInit_GoldenFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")

	require.NoError(t, runConfigInit(newConfigInitOpts(t, func(o *configInitOptions) {
		o.OutputPath = path
	})))

	got, err := os.ReadFile(path)
	require.NoError(t, err, "generated config must be readable")

	goldenPath := filepath.Join("testdata", "config_init_default.golden")
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file must exist at %s", goldenPath)

	assert.Equal(t, want, got,
		"scaffolded config must be byte-identical to the golden baseline")
}

// T4 — --force overwrites an existing config (spec scenario "Init
// overwrites with force flag"): the pre-existing content is replaced
// with the scaffolded content.
func TestRunConfigInit_ForceOverwrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte("existing"), 0o644))

	require.NoError(t, runConfigInit(newConfigInitOpts(t, func(o *configInitOptions) {
		o.OutputPath = path
		o.Force = true
	})))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(got, []byte("# Germinator configuration")),
		"--force must overwrite with the scaffolded content")
}

// T5 — Existing file without --force returns a typed *core.FileError
// (spec scenario "Init refuses to overwrite without force").
// cmdutil.ExitCodeFor maps it to ExitCodeError (1).
func TestRunConfigInit_ExistingWithoutForceFails(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte("existing"), 0o644))

	err := runConfigInit(newConfigInitOpts(t, func(o *configInitOptions) {
		o.OutputPath = path
	}))
	require.Error(t, err)

	var fileErr *core.FileError
	require.True(t, errors.As(err, &fileErr), "expected *core.FileError in chain")
	assert.Contains(t, fileErr.Message(), "already exists")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"existing-file failure must map to ExitCodeError (1)")
}

// T6 — Parent directories are created (spec scenario "parent
// directories are created with permissions 0750"): a nested path
// that doesn't exist is created on the fly.
func TestRunConfigInit_CreatesParentDirectories(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "subdir", "nested", "config.toml")

	require.NoError(t, runConfigInit(newConfigInitOpts(t, func(o *configInitOptions) {
		o.OutputPath = path
	})))

	info, err := os.Stat(path)
	require.NoError(t, err, "config file must exist at nested path")
	assert.False(t, info.IsDir())
}

// T7 — Default path resolution (spec scenario "Default output path"):
// when opts.OutputPath is empty, the XDG-resolved default is used.
// Pinned via XDG_CONFIG_HOME so the test is hermetic.
func TestRunConfigInit_DefaultPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	want := filepath.Join(tmp, "germinator", "config.toml")

	require.NoError(t, runConfigInit(newConfigInitOpts(t, nil)))

	_, err := os.Stat(want)
	require.NoError(t, err, "default-path init must create config at %s", want)
}

// T8 — Success message goes to STDOUT (spec scenario for config init
// success): the confirmation line is on opts.IO.Out, nothing on
// opts.IO.ErrOut. This keeps stdout scriptable and stderr clean.
func TestRunConfigInit_SuccessOnStdout(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	ios, out, errOut := newConfigInitTestIO()

	require.NoError(t, runConfigInit(&configInitOptions{
		IO: ios, Ctx: t.Context(), OutputPath: path,
	}))

	got := out.String()
	assert.Contains(t, got, "Successfully created config file:")
	assert.Contains(t, got, path,
		"success line must include the resolved path")
	assert.Empty(t, errOut.String(),
		"stderr must stay empty on the success path")
}

// T9 — Spec scenario "Legacy --output returns usage error": the
// BREAKING --output → --output-path rename means invoking --output
// now yields a Cobra usage error mapped to ExitCodeUsage (2).
func TestRunConfigInit_RejectsLegacyOutputFlag(t *testing.T) {
	ios, _, errOut := newConfigInitTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdConfigInit(f, nil)
	cmd.SetArgs([]string{"--output", "/tmp/foo.toml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(errOut)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"unknown --output flag must map to ExitCodeUsage (2)")
	assert.Contains(t, errOut.String(), "output",
		"stderr must mention the rejected --output flag")
}

// T10 — configInitOptions struct shape: declares exactly the
// spec-named fields; reflection guards against accidental drops
// or renames (mirrors TestLibraryInitOptions_StructShape).
func TestConfigInitOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(configInitOptions{})
	want := map[string]bool{
		"IO":         true,
		"Ctx":        true,
		"OutputPath": true,
		"Force":      true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}
	assert.Equal(t, want, got,
		"configInitOptions must declare exactly the spec-named fields")
}

// TestNewConfigCommand_RegistersSubcommands verifies that the slimmed
// parent constructor registers both `init` and `validate` subcommands
// and that --help renders the long description listing both. This
// guards against accidental regressions in cmd/config.go where a
// future change might drop an AddCommand call.
func TestNewConfigCommand_RegistersSubcommands(t *testing.T) {
	ios, _, _ := newConfigInitTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewConfigCommand(f)

	// Both subcommands must be registered
	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	assert.True(t, names["init"], "config parent must register the `init` subcommand")
	assert.True(t, names["validate"], "config parent must register the `validate` subcommand")

	// The long description should mention both subcommands (preserves
	// the public help-text contract from the pre-change build).
	long := cmd.Long
	assert.Contains(t, long, "config init")
	assert.Contains(t, long, "config validate")
}

// T11 — Spec scenario "Init creates config at custom location" (parent
// directory permissions): the spec mandates parent directories are
// created with permissions 0750 and the file with 0600. The file mode
// (0600) is umask-stable; the directory mode is asserted as "no more
// permissive than 0750" because umask may further narrow the bits
// (e.g. umask 0o077 yields 0o700 while umask 0o022 yields 0o750).
func TestRunConfigInit_FileModePermissions(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "subdir", "nested", "config.toml")

	require.NoError(t, runConfigInit(newConfigInitOpts(t, func(o *configInitOptions) {
		o.OutputPath = path
	})))

	// File must be exactly 0600 (no group/other bits to strip, so umask
	// has no effect on this mode).
	fi, err := os.Stat(path)
	require.NoError(t, err, "config file must exist")
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(),
		"config file must be written with 0600 permissions")

	// Newly-created parent directories must be no more permissive than
	// 0750. umask can only narrow the requested 0750, so the resulting
	// perm satisfies `perm <= 0750`.
	di, err := os.Stat(filepath.Join(tmp, "subdir"))
	require.NoError(t, err, "parent directory must exist")
	assert.LessOrEqual(t, di.Mode().Perm(), os.FileMode(0o750),
		"parent directory must be no more permissive than 0750 (umask may narrow it)")
}

// T12 — Spec scenario "Default output path" (fallback branch): when
// opts.OutputPath is empty AND XDG_CONFIG_HOME is unset, the
// ~/.config/germinator/config.toml fallback selected by HOME must be
// created. Pinned via HOME so the test is hermetic and does not touch
// the developer's real ~/.config. Complements T7 which pins the XDG
// branch.
func TestRunConfigInit_DefaultPathFallback(t *testing.T) {
	tmp := t.TempDir()
	// Force the XDG_CONFIG_HOME branch off so GetConfigPath falls back to
	// $HOME/.config/germinator/config.toml.
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmp)

	want := filepath.Join(tmp, ".config", "germinator", "config.toml")

	require.NoError(t, runConfigInit(newConfigInitOpts(t, nil)))

	_, err := os.Stat(want)
	require.NoError(t, err, "fallback-path init must create config at %s", want)
}

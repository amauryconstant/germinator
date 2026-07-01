package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// newConfigValidateTestIO returns the buffer-backed IOStreams that
// config-validate tests use to assert on captured Out / ErrOut.
// Named with a configValidate prefix to avoid collision with other
// validate test helpers (cmd/validate_test.go's newValidateTestIO,
// cmd/library_validate_test.go's newLibraryValidateTestIO).
func newConfigValidateTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// newConfigValidateOpts builds a *configValidateOptions with
// iostreams.Test() buffers and t.Context(). mut may be nil or a
// function that mutates the opts struct after the defaults are set.
func newConfigValidateOpts(t *testing.T, mut func(*configValidateOptions)) *configValidateOptions {
	t.Helper()
	ios, _, _ := newConfigValidateTestIO()
	opts := &configValidateOptions{IO: ios, Ctx: t.Context()}
	if mut != nil {
		mut(opts)
	}
	return opts
}

// writeConfig writes body to path, creating parent directories as
// needed. Centralized so each test doesn't repeat the os.WriteFile
// boilerplate.
func writeConfig(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750),
		"create parent dir for %s", path)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600),
		"write config fixture at %s", path)
}

// T1 — Constructor wires opts correctly via runF injection (spec
// scenario "config validate supports runF injection").
func TestNewCmdConfigValidate_ConstructorWiresOpts(t *testing.T) {
	var captured *configValidateOptions
	runF := func(opts *configValidateOptions) error {
		captured = opts
		return nil
	}

	ios, _, _ := newConfigValidateTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdConfigValidate(f, runF)
	cmd.SetArgs([]string{"--output-path", "/tmp/wired.toml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/wired.toml", captured.OutputPath)
	assert.NotNil(t, captured.IO, "opts.IO must be wired from f.IOStreams")
	assert.NotNil(t, captured.Ctx, "opts.Ctx must be wired from c.Context()")
}

// T2 — Valid config succeeds (spec scenario "Validate succeeds for
// valid config"): no error, single success line on stdout, nothing
// on stderr.
func TestRunConfigValidate_ValidConfig(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	writeConfig(t, path, `library = "~/.config/germinator/library"
platform = "opencode"
[completion]
timeout = "500ms"
cache_ttl = "5s"
`)

	ios, out, errOut := newConfigValidateTestIO()
	require.NoError(t, runConfigValidate(&configValidateOptions{
		IO: ios, Ctx: t.Context(), OutputPath: path,
	}))

	assert.Contains(t, out.String(), "Config file is valid:")
	assert.Contains(t, out.String(), path)
	assert.Empty(t, errOut.String(),
		"stderr must stay empty on the success path")
}

// T3 — File not found (spec scenario "Validate fails when file not
// found"): returns *core.FileError whose IsNotFound() is true and
// maps to ExitCodeError (1).
func TestRunConfigValidate_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.toml")

	err := runConfigValidate(newConfigValidateOpts(t, func(o *configValidateOptions) {
		o.OutputPath = path
	}))
	require.Error(t, err)

	var fileErr *core.FileError
	require.True(t, errors.As(err, &fileErr), "expected *core.FileError in chain")
	assert.True(t, fileErr.IsNotFound(),
		"FileError.IsNotFound must be true for a missing config file")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"file-not-found failure must map to ExitCodeError (1)")
}

// T4 — Malformed TOML (spec scenario "Validate fails on malformed
// TOML"): the parse failure surfaces as a typed error whose message
// references the parse stage.
func TestRunConfigValidate_MalformedTOML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	writeConfig(t, path, "invalid [ [")

	err := runConfigValidate(newConfigValidateOpts(t, func(o *configValidateOptions) {
		o.OutputPath = path
	}))
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "parse",
		"malformed-TOML error must reference the parse stage: %v", err)
}

// T5 — Invalid platform value (spec scenario "Validate fails on
// invalid platform value"): cfgObj.Validate() returns *core.ConfigError
// which propagates through fmt.Errorf("validating config: %w", err);
// errors.As traverses the wrap chain to expose the typed error.
func TestRunConfigValidate_InvalidPlatform(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	writeConfig(t, path, `platform = "unknown"
`)

	err := runConfigValidate(newConfigValidateOpts(t, func(o *configValidateOptions) {
		o.OutputPath = path
	}))
	require.Error(t, err)

	var configErr *core.ConfigError
	require.True(t, errors.As(err, &configErr),
		"expected *core.ConfigError in the wrap chain: %v", err)
	assert.Equal(t, "platform", configErr.Field(),
		"ConfigError.Field must identify the invalid field")
}

// T6 — Default path resolution (spec scenario "Validate uses default
// path"): when opts.OutputPath is empty, the XDG-resolved default is
// validated. Pinned via XDG_CONFIG_HOME so the test is hermetic.
func TestRunConfigValidate_DefaultPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	path := filepath.Join(tmp, "germinator", "config.toml")
	writeConfig(t, path, `platform = "opencode"
`)

	require.NoError(t, runConfigValidate(newConfigValidateOpts(t, nil)))
}

// T7 — Success message on STDOUT (spec requirement: "on success it
// SHALL write a single success line to opts.IO.Out and nothing to
// opts.IO.ErrOut"). Keeps stdout scriptable, stderr clean.
func TestRunConfigValidate_SuccessOnStdout(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	writeConfig(t, path, `platform = "opencode"
`)

	ios, out, errOut := newConfigValidateTestIO()
	require.NoError(t, runConfigValidate(&configValidateOptions{
		IO: ios, Ctx: t.Context(), OutputPath: path,
	}))

	assert.Contains(t, out.String(), "Config file is valid:")
	assert.Contains(t, out.String(), path,
		"success line must include the validated path")
	assert.Empty(t, errOut.String(),
		"stderr must stay empty on the success path")
}

// T8 — Spec scenario "Legacy --output returns usage error": the
// BREAKING --output → --output-path rename means invoking --output
// now yields a Cobra usage error mapped to ExitCodeUsage (2).
func TestRunConfigValidate_RejectsLegacyOutputFlag(t *testing.T) {
	ios, _, errOut := newConfigValidateTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdConfigValidate(f, nil)
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

// T9 — configValidateOptions struct shape: declares exactly the
// spec-named fields (no Force); reflection guards against accidental
// drops or renames.
func TestConfigValidateOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(configValidateOptions{})
	want := map[string]bool{
		"IO":         true,
		"Ctx":        true,
		"OutputPath": true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}
	assert.Equal(t, want, got,
		"configValidateOptions must declare exactly the spec-named fields")
}

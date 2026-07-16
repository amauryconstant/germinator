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
	runF := func(opts *configValidateOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	ios, _, _ := newConfigValidateTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdConfigValidate(f, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "--output-path", "/tmp/wired.toml"))
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
	err := executeCmd(t, func() any {
		cmd := NewCmdConfigValidate(f, nil)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(errOut)
		return cmd
	}, "--output", "/tmp/foo.toml")
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

// T10 — Spec scenario "No double-rendering of validation errors"
// (Requirement: config validate renders errors at the boundary).
//
// The single-handling rule (see the golang-error-handling and
// golang-cli-architecture skills) mandates that errors are EITHER
// rendered OR returned, never both.
// runConfigValidate must NOT write anything to opts.IO.ErrOut; main.go
// renders the returned error exactly once via output.FormatError.
//
// This test exercises every error branch and asserts stderr stays empty
// so a future regression (e.g. someone copies the latent smell in
// cmd/validate.go:120-124) fails loudly rather than double-printing in
// production logs.
func TestRunConfigValidate_DoesNotRenderErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, opts *configValidateOptions)
	}{
		{
			name: "file not found",
			setup: func(t *testing.T, opts *configValidateOptions) {
				opts.OutputPath = filepath.Join(t.TempDir(), "missing.toml")
			},
		},
		{
			name: "malformed TOML",
			setup: func(t *testing.T, opts *configValidateOptions) {
				path := filepath.Join(t.TempDir(), "bad.toml")
				writeConfig(t, path, "invalid [ [")
				opts.OutputPath = path
			},
		},
		{
			name: "invalid platform value",
			setup: func(t *testing.T, opts *configValidateOptions) {
				path := filepath.Join(t.TempDir(), "bad-platform.toml")
				writeConfig(t, path, `platform = "unknown"`+"\n")
				opts.OutputPath = path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, errOut := newConfigValidateTestIO()
			opts := &configValidateOptions{IO: ios, Ctx: t.Context()}
			tt.setup(t, opts)

			err := runConfigValidate(opts)
			require.Error(t, err, "error path must return a non-nil error")
			assert.Empty(t, errOut.String(),
				"command body must NOT render to ErrOut (single-handling rule); "+
					"main.go renders the returned error once via output.FormatError")
		})
	}
}

// T11 — Covers runConfigValidate's koanf.Unmarshal failure branch
// (cmd/config_validate.go:106-108): a TOML file that parses successfully
// but cannot unmarshal into *config.Config surfaces as a *core.ParseError
// carrying the file path. This complements T4 (MalformedTOML) which only
// triggers the earlier koanf.Load branch.
func TestRunConfigValidate_UnmarshalFailure(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "type-mismatch.toml")
	// `platform` is declared as a string in config.Config; an array value
	// parses as valid TOML but fails mapstructure unmarshalling.
	writeConfig(t, path, `platform = ["a", "b"]`+"\n")

	err := runConfigValidate(newConfigValidateOpts(t, func(o *configValidateOptions) {
		o.OutputPath = path
	}))
	require.Error(t, err)

	var parseErr *core.ParseError
	require.True(t, errors.As(err, &parseErr),
		"Unmarshal failure must surface as *core.ParseError: %v", err)
	assert.Equal(t, path, parseErr.Path(),
		"ParseError must carry the offending file path")
}

// T12 — Spec scenario "Validate uses default path" (fallback branch):
// when opts.OutputPath is empty AND XDG_CONFIG_HOME is unset, the
// ~/.config/germinator/config.toml fallback selected by HOME must be
// validated. Pinned via HOME so the test is hermetic and does not touch
// the developer's real ~/.config. Complements T6 which pins the XDG
// branch.
func TestRunConfigValidate_DefaultPathFallback(t *testing.T) {
	tmp := t.TempDir()
	// Force the XDG_CONFIG_HOME branch off so GetConfigPath falls back to
	// $HOME/.config/germinator/config.toml.
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmp)

	path := filepath.Join(tmp, ".config", "germinator", "config.toml")
	writeConfig(t, path, `platform = "opencode"`+"\n")

	require.NoError(t, runConfigValidate(newConfigValidateOpts(t, nil)))
}

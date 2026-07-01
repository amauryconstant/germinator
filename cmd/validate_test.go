package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// fakeValidator is a hand-rolled fake satisfying the local cmd.Validator
// interface (defined in cmd/validate.go). It records the last request
// it received and returns the configured result or error.
type fakeValidator struct {
	calls   int
	lastReq *ValidateRequest
	result  *core.ValidateResult
	err     error
}

// Compile-time interface satisfaction check.
var _ Validator = (*fakeValidator)(nil)

func (f *fakeValidator) Validate(_ context.Context, req *ValidateRequest) (*core.ValidateResult, error) {
	f.calls++
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}
	if f.result != nil {
		return f.result, nil
	}
	return &core.ValidateResult{Errors: nil}, nil
}

func newValidateTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func TestRunValidate_HappyPath(t *testing.T) {
	t.Parallel()

	io, out, errOut := newValidateTestIO()
	fake := &fakeValidator{
		result: &core.ValidateResult{Errors: nil},
	}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/agent.md",
		Platform:  core.PlatformClaudeCode,
	}

	require.NoError(t, runValidate(opts))

	assert.Equal(t, 1, fake.calls, "validator must be called exactly once")
	require.NotNil(t, fake.lastReq)
	assert.Equal(t, "/tmp/agent.md", fake.lastReq.InputPath)
	assert.Equal(t, core.PlatformClaudeCode, fake.lastReq.Platform)

	assert.Equal(t, "Document is valid\n", out.String(),
		"stdout must contain the success line")
	assert.Empty(t, errOut.String(),
		"stderr must be empty when not verbose")
}

func TestRunValidate_SingleError_RendersViaFormatError(t *testing.T) {
	t.Parallel()

	io, out, errOut := newValidateTestIO()
	parseErr := core.NewParseError("/tmp/agent.md", "unrecognizable filename", nil)
	fake := &fakeValidator{
		result: &core.ValidateResult{Errors: []error{parseErr}},
	}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/agent.md",
		Platform:  core.PlatformClaudeCode,
	}

	err := runValidate(opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, parseErr,
		"first error must propagate through the chain")
	assert.Empty(t, out.String(), "stdout must be empty on validation failure")
	assert.Contains(t, errOut.String(), "Error:",
		"FormatError must render the error to stderr")
	assert.Contains(t, errOut.String(), "unrecognizable filename",
		"FormatError must render the parse error message")
}

func TestRunValidate_MultiErrors_RendersAll(t *testing.T) {
	t.Parallel()

	io, out, errOut := newValidateTestIO()
	err1 := core.NewValidationError("Agent", "name", "", "name is required")
	err2 := core.NewValidationError("Agent", "description", "", "description is required")
	fake := &fakeValidator{
		result: &core.ValidateResult{Errors: []error{err1, err2}},
	}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/agent.md",
		Platform:  core.PlatformClaudeCode,
	}

	err := runValidate(opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, err1, "first error must propagate")
	assert.Empty(t, out.String(), "stdout must be empty on validation failure")

	stderr := errOut.String()
	assert.Contains(t, stderr, "name is required",
		"first error must be rendered to stderr")
	assert.Contains(t, stderr, "description is required",
		"second error must be rendered to stderr")

	count := strings.Count(stderr, "Error:")
	assert.Equal(t, 2, count, "each error must be rendered once via FormatError")
}

func TestRunValidate_InvalidPlatform_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	io, _, _ := newValidateTestIO()
	fake := &fakeValidator{}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/agent.md",
		Platform:  "",
	}

	err := runValidate(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr),
		"error must wrap *core.ValidationError")
	assert.Equal(t, 0, fake.calls,
		"validator must NOT be called when platform is invalid")
}

func TestRunValidate_UnknownDocType_ReturnsParseError(t *testing.T) {
	t.Parallel()

	io, _, _ := newValidateTestIO()
	parseErr := core.NewParseError("/tmp/foo.txt", "unrecognizable filename", nil)
	fake := &fakeValidator{err: parseErr}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/foo.txt",
		Platform:  core.PlatformClaudeCode,
	}

	err := runValidate(opts)
	require.Error(t, err)

	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"error must wrap *core.ParseError")
	assert.Contains(t, err.Error(), "validating document",
		"error must be wrapped with 'validating document' context")
}

func TestRunValidate_ValidatorError_Propagates(t *testing.T) {
	t.Parallel()

	io, out, _ := newValidateTestIO()
	cause := errors.New("disk full")
	fake := &fakeValidator{err: cause}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/agent.md",
		Platform:  core.PlatformClaudeCode,
	}

	err := runValidate(opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause,
		"original error must be preserved in the chain")
	assert.Contains(t, err.Error(), "validating document",
		"error must be wrapped with 'validating document' context")
	assert.Empty(t, out.String(),
		"stdout must be empty on error")
}

func TestRunValidate_VerboseProgressToStderr(t *testing.T) {
	t.Parallel()

	io, out, errOut := newValidateTestIO()
	io.Verbose = true
	fake := &fakeValidator{}
	opts := &validateOptions{
		IO:        io,
		Validator: func() (Validator, error) { return fake, nil },
		Ctx:       context.Background(),
		InputPath: "/tmp/agent.md",
		Platform:  core.PlatformClaudeCode,
	}

	require.NoError(t, runValidate(opts))

	assert.Contains(t, errOut.String(), "validating /tmp/agent.md",
		"verbose progress must go to stderr")
	assert.Contains(t, errOut.String(), core.PlatformClaudeCode,
		"verbose progress must mention the platform")
	assert.Equal(t, "Document is valid\n", out.String(),
		"stdout must remain the success line only (no verbose leakage)")
}

func TestNewCmdValidate_RunFCapturesOpts(t *testing.T) {
	var captured *validateOptions
	runF := func(opts *validateOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdValidate(f, runF)
	cmd.SetArgs([]string{"/tmp/agent.md", "--platform", "opencode"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/agent.md", captured.InputPath)
	assert.Equal(t, "opencode", captured.Platform)
	require.NotNil(t, captured.IO)
	assert.Equal(t, io, captured.IO, "opts.IO must be the Factory's IOStreams")
	require.NotNil(t, captured.Ctx, "opts.Ctx must be set from c.Context()")
}

func TestNewCmdValidate_RequiresPlatformFlag(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdValidate(f, func(*validateOptions) error { return nil })
	cmd.SetArgs([]string{"/tmp/agent.md"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing required --platform flag must fail")
}

func TestNewCmdValidate_NilRunFFallsBackToProduction(t *testing.T) {
	io, out, errOut := newValidateTestIO()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	// slice-7: NewCmdValidate's production wiring constructs the
	// Validator lazily inside runValidate (cmd.NewValidator()). The
	// nil-runF path therefore exercises the full runValidate →
	// parse → validate pipeline. With a valid platform but a missing
	// input file, parse fails and the error surfaces through
	// cmdutil.ExitCodeFor → ExitCodeError (1).
	cmd := NewCmdValidate(f, nil)
	cmd.SetArgs([]string{"/nonexistent.md", "--platform", "opencode"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing input file must surface as an error")
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

// slice-7 removed the Factory.Validator lazy field and the
// `validateValidator(f)` factory helper. The Validator is now
// constructed inside runValidate via cmd.NewValidator(); test
// fakes are injected via runValidate directly (see fakeValidator
// at the top of this file).

func TestNewValidator_AdapterSatisfiesInterface(t *testing.T) {
	t.Parallel()

	// Compile-time interface check is already in validate.go
	// (var _ Validator = (*validatorAdapter)(nil)).
	// This test verifies the runtime contract: a value returned by
	// NewValidator() must accept the Validate call shape that the
	// local Validator interface defines.
	result, err := NewValidator().Validate(context.Background(), &ValidateRequest{
		InputPath: "/nonexistent.md",
		Platform:  core.PlatformClaudeCode,
	})
	require.Error(t, err,
		"validateDocument must return a fatal error for unrecognizable file")
	assert.Nil(t, result)
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"fatal error must wrap *core.ParseError")
}

func TestValidateDocument_HappyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := dir + "/agent-valid.md"
	content := `---
name: tester
description: A test agent
---
Body`
	require.NoError(t, os.WriteFile(inputPath, []byte(content), 0o600))

	result, err := validateDocument(context.Background(), &ValidateRequest{
		InputPath: inputPath,
		Platform:  core.PlatformClaudeCode,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid(), "expected valid result")
	assert.Empty(t, result.Errors)
}

func TestUnwrapErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantLen int
	}{
		{name: "nil error returns nil", err: nil, wantLen: 0},
		{name: "single error", err: errors.New("boom"), wantLen: 1},
		{name: "two joined errors", err: errors.Join(errors.New("a"), errors.New("b")), wantLen: 2},
		{name: "three joined errors", err: errors.Join(errors.New("x"), errors.New("y"), errors.New("z")), wantLen: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := unwrapErrors(tt.err)
			assert.Len(t, errs, tt.wantLen)
		})
	}
}

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
	"gitlab.com/amoconst/germinator/internal/output"
	"gitlab.com/amoconst/germinator/internal/validate"
)

// fakeValidator is a hand-rolled fake satisfying the local cmd.Validator
// interface (defined in cmd/validate.go). It records the last request
// it received and returns the configured result or error.
type fakeValidator struct {
	calls   int
	lastReq *validate.Request
	result  *core.ValidateResult
	err     error
}

// Compile-time interface satisfaction check.
var _ Validator = (*fakeValidator)(nil)

func (f *fakeValidator) Validate(ctx context.Context, req *validate.Request) (*core.ValidateResult, error) {
	_ = ctx // accept-and-may-ignore: fake records the request only
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
	assert.Empty(t, out.String(),
		"runValidate must NOT render to stdout (single-handling rule)")
	assert.Empty(t, errOut.String(),
		"runValidate must NOT render to ErrOut (single-handling rule); "+
			"main.go renders the returned error once via output.FormatError")

	output.FormatError(io, err)
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
	assert.Empty(t, out.String(),
		"runValidate must NOT render to stdout (single-handling rule)")
	assert.Empty(t, errOut.String(),
		"runValidate must NOT render to ErrOut (single-handling rule); "+
			"main.go renders the returned error once via output.FormatError")

	output.FormatError(io, err)
	stderr := errOut.String()
	assert.Contains(t, stderr, "name is required",
		"first error must be rendered via the central FormatError handler")
	assert.Equal(t, 1, strings.Count(stderr, "Error:"),
		"only the returned error is rendered (single-handling rule)")
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
	runF := func(opts *validateOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io)
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdValidate(f, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "/tmp/agent.md", "--platform", "opencode"))
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/agent.md", captured.InputPath)
	assert.Equal(t, "opencode", captured.Platform)
	require.NotNil(t, captured.IO)
	assert.Equal(t, io, captured.IO, "opts.IO must be the Factory's IOStreams")
	require.NotNil(t, captured.Ctx, "opts.Ctx must be set from c.Context()")
}

func TestNewCmdValidate_RequiresPlatformFlag(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io)
	err := executeCmd(t, func() any {
		cmd := NewCmdValidate(f, func(*validateOptions) error { return nil })
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "/tmp/agent.md")
	require.Error(t, err, "missing required --platform flag must fail")
}

// slice-7 removed the Factory.Validator lazy field and the
// `validateValidator(f)` factory helper. The Validator is now
// constructed inside runValidate via validate.NewService(); test
// fakes are injected via runValidate directly (see fakeValidator
// at the top of this file).

func TestValidateService_AdapterSatisfiesInterface(t *testing.T) {
	t.Parallel()

	// Compile-time interface check lives in
	// internal/validate/validate.go (var _ Service = (*validateService)(nil)).
	// This test verifies the runtime contract: a value returned by
	// validate.NewService() must accept the Validate call shape that
	// the local Validator interface defines (structural typing).
	result, err := validate.NewService().Validate(context.Background(), &validate.Request{
		InputPath: "/nonexistent.md",
		Platform:  core.PlatformClaudeCode,
	})
	require.Error(t, err,
		"validateService must return a fatal error for unrecognizable file")
	assert.Nil(t, result)
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"fatal error must wrap *core.ParseError")
}

func TestValidateService_HappyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := dir + "/agent-valid.md"
	content := `---
name: tester
description: A test agent
---
Body`
	require.NoError(t, os.WriteFile(inputPath, []byte(content), 0o600))

	result, err := validate.NewService().Validate(context.Background(), &validate.Request{
		InputPath: inputPath,
		Platform:  core.PlatformClaudeCode,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid(), "expected valid result")
	assert.Empty(t, result.Errors)
}

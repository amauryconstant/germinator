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
)

// Compile-time check: fakeCanonicalizer must satisfy the local
// cmd.Canicalizer interface. Mirrors the pattern in cmd/adapt_test.go.
var _ Canonicalizer = (*fakeCanonicalizer)(nil)

// fakeCanonicalizer is a hand-rolled fake satisfying the local
// cmd.Canonicalizer interface (defined in cmd/canonicalize.go). It
// records the last request it received and returns the configured
// result or error. Mirrors fakeTransformer in cmd/adapt_test.go.
type fakeCanonicalizer struct {
	calls   int
	lastReq *CanonicalizeRequest
	result  *core.CanonicalizeResult
	err     error
}

func (f *fakeCanonicalizer) Canonicalize(_ context.Context, req *CanonicalizeRequest) (*core.CanonicalizeResult, error) {
	f.calls++
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}
	if f.result != nil {
		return f.result, nil
	}
	return &core.CanonicalizeResult{OutputPath: req.OutputPath}, nil
}

func newCanonicalizeTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func TestRunCanonicalize_HappyPath(t *testing.T) {
	t.Parallel()

	io, out, errOut := newCanonicalizeTestIO()
	fake := &fakeCanonicalizer{}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      core.PlatformClaudeCode,
		DocType:       "agent",
	}

	require.NoError(t, runCanonicalize(opts))

	assert.Equal(t, 1, fake.calls, "canonicalizer must be called exactly once")
	require.NotNil(t, fake.lastReq)
	assert.Equal(t, "/tmp/in.md", fake.lastReq.InputPath)
	assert.Equal(t, "/tmp/out.yaml", fake.lastReq.OutputPath)
	assert.Equal(t, core.PlatformClaudeCode, fake.lastReq.Platform)
	assert.Equal(t, "agent", fake.lastReq.DocType)

	assert.Equal(t, "Successfully canonicalized document to: /tmp/out.yaml\n", out.String(),
		"stdout must contain the success line")
	assert.Empty(t, errOut.String(), "stderr must be empty when not verbose")
}

func TestRunCanonicalize_InvalidPlatform_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	io, _, _ := newCanonicalizeTestIO()
	fake := &fakeCanonicalizer{}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      "",
		DocType:       "agent",
	}

	err := runCanonicalize(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "error must wrap *core.ValidationError")
	assert.Equal(t, 0, fake.calls, "canonicalizer must NOT be called when platform is invalid")
}

func TestRunCanonicalize_InvalidPlatformValue(t *testing.T) {
	t.Parallel()

	io, _, _ := newCanonicalizeTestIO()
	fake := &fakeCanonicalizer{}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      "windows-95",
		DocType:       "agent",
	}

	err := runCanonicalize(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown platform "windows-95"`)
	assert.Equal(t, 0, fake.calls)
}

func TestRunCanonicalize_ParseError_Propagates(t *testing.T) {
	t.Parallel()

	io, out, _ := newCanonicalizeTestIO()
	cause := core.NewParseError("/tmp/in.md", "bad yaml", errors.New("unclosed bracket"))
	fake := &fakeCanonicalizer{err: cause}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      core.PlatformOpenCode,
		DocType:       "skill",
	}

	err := runCanonicalize(opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause, "original error must be preserved in the chain")

	var parseErr *core.ParseError
	require.True(t, errors.As(err, &parseErr), "error must wrap *core.ParseError")
	assert.Empty(t, out.String(), "stdout must be empty on error")
}

func TestRunCanonicalize_ValidationError_Propagates(t *testing.T) {
	t.Parallel()

	io, _, _ := newCanonicalizeTestIO()
	cause := core.NewValidationError("", "name", "", "missing required field")
	fake := &fakeCanonicalizer{err: cause}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      core.PlatformClaudeCode,
		DocType:       "command",
	}

	err := runCanonicalize(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "error must wrap *core.ValidationError")
}

func TestRunCanonicalize_FileWriteError_Propagates(t *testing.T) {
	t.Parallel()

	io, _, _ := newCanonicalizeTestIO()
	cause := core.NewFileError("/tmp/out.yaml", "write", "permission denied", errors.New("EACCES"))
	fake := &fakeCanonicalizer{err: cause}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      core.PlatformClaudeCode,
		DocType:       "memory",
	}

	err := runCanonicalize(opts)
	require.Error(t, err)

	var ferr *core.FileError
	require.True(t, errors.As(err, &ferr), "error must wrap *core.FileError")
}

func TestRunCanonicalize_VerboseProgressToStderr(t *testing.T) {
	t.Parallel()

	io, out, errOut := newCanonicalizeTestIO()
	io.Verbose = true
	fake := &fakeCanonicalizer{}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      core.PlatformClaudeCode,
		DocType:       "agent",
	}

	require.NoError(t, runCanonicalize(opts))

	assert.Contains(t, errOut.String(), "canonicalizing /tmp/in.md",
		"verbose progress must go to stderr")
	assert.Contains(t, errOut.String(), "/tmp/out.yaml",
		"verbose progress must mention the output path")
	assert.Contains(t, errOut.String(), "claude-code",
		"verbose progress must mention the platform")
	assert.Contains(t, errOut.String(), "agent",
		"verbose progress must mention the doc type")
	assert.Equal(t, "Successfully canonicalized document to: /tmp/out.yaml\n", out.String(),
		"stdout must remain the success line only (no verbose leakage)")
}

func TestRunCanonicalize_StreamDisciplineNonVerbose(t *testing.T) {
	t.Parallel()

	io, out, errOut := newCanonicalizeTestIO()
	fake := &fakeCanonicalizer{}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/tmp/in.md",
		OutputPath:    "/tmp/out.yaml",
		Platform:      core.PlatformClaudeCode,
		DocType:       "agent",
	}

	require.NoError(t, runCanonicalize(opts))

	assert.NotContains(t, out.String(), "canonicalizing",
		"stdout must not contain verbose progress when verbose is off")
	assert.Empty(t, errOut.String(), "stderr must be empty when verbose is off")
}

func TestRunCanonicalize_CanonicalizerReceivesExactRequest(t *testing.T) {
	t.Parallel()

	io, _, _ := newCanonicalizeTestIO()
	fake := &fakeCanonicalizer{}
	opts := &canonicalizeOptions{
		IO:            io,
		Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
		Ctx:           context.Background(),
		InputPath:     "/data/agent.md",
		OutputPath:    "/data/agent.canonical.yaml",
		Platform:      core.PlatformOpenCode,
		DocType:       "skill",
	}

	require.NoError(t, runCanonicalize(opts))
	require.NotNil(t, fake.lastReq)
	assert.Equal(t, "/data/agent.md", fake.lastReq.InputPath)
	assert.Equal(t, "/data/agent.canonical.yaml", fake.lastReq.OutputPath)
	assert.Equal(t, core.PlatformOpenCode, fake.lastReq.Platform)
	assert.Equal(t, "skill", fake.lastReq.DocType)
}

func TestNewCmdCanonicalize_RunFCapturesOpts(t *testing.T) {
	t.Parallel()

	var captured *canonicalizeOptions
	runF := func(opts *canonicalizeOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	f.Canonicalizer = cmdutil.OnceValuesFunc(func() (application.Canonicalizer, error) {
		return &fakeCanonicalizer{}, nil
	})
	cmd := NewCmdCanonicalize(f, runF)
	cmd.SetArgs([]string{"/tmp/in.md", "/tmp/out.yaml", "--platform", "opencode", "--type", "command"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/in.md", captured.InputPath)
	assert.Equal(t, "/tmp/out.yaml", captured.OutputPath)
	assert.Equal(t, "opencode", captured.Platform)
	assert.Equal(t, "command", captured.DocType)
	assert.Equal(t, io, captured.IO, "opts.IO must be the Factory's IOStreams")
	require.NotNil(t, captured.Ctx, "opts.Ctx must be set from c.Context()")
	require.NotNil(t, captured.Canonicalizer,
		"opts.Canonicalizer must be populated by NewCmdCanonicalize (via canonicalizeCanonicalizer)")
}

func TestNewCmdCanonicalize_RequiresPlatformAndTypeFlags(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdCanonicalize(f, func(*canonicalizeOptions) error { return nil })
	cmd.SetArgs([]string{"/tmp/in.md", "/tmp/out.yaml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing required --platform and --type flags must fail")
}

func TestNewCmdCanonicalize_RequiresTypeFlag(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdCanonicalize(f, func(*canonicalizeOptions) error { return nil })
	cmd.SetArgs([]string{"/tmp/in.md", "/tmp/out.yaml", "--platform", "opencode"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing --type flag must fail")
}

func TestCanonicalizeCanonicalizer_NilFactoryReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, canonicalizeCanonicalizer(nil),
		"canonicalizeCanonicalizer(nil) must return nil so opts.Canonicalizer is nil")

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	f.Canonicalizer = nil
	assert.Nil(t, canonicalizeCanonicalizer(f),
		"canonicalizeCanonicalizer with nil f.Canonicalizer must return nil")
}

func TestNewCanonicalizer_AdapterSatisfiesInterface(t *testing.T) {
	t.Parallel()

	// Compile-time interface check is already in canonicalize.go
	// (var _ application.Canonicalizer = (*canonicalizerAdapter)(nil)).
	// This test verifies the runtime contract: a value returned by
	// NewCanonicalizer() must accept the Canonicalize call shape that
	// the local Canonicalizer interface defines, and must propagate
	// the underlying typed error from canonicalizeDocument so callers
	// (legacyBridge consumers) can errors.As their way to the right
	// exit code.
	result, err := NewCanonicalizer().Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  "/nonexistent/agent.md",
		OutputPath: t.TempDir() + "/out.yaml",
		Platform:   core.PlatformClaudeCode,
		DocType:    "agent",
	})
	require.Error(t, err,
		"canonicalizeDocument must return a fatal error for unrecognizable input")
	assert.Nil(t, result,
		"canonicalizeDocument must return a nil result on error")
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"fatal error must wrap *core.ParseError")
}

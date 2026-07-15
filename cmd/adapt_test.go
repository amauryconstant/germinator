package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/transform"
)

// fakeTransformer is a hand-rolled fake satisfying the local cmd.Transformer
// interface (defined in cmd/adapt.go). The Transformer signature matches
// *transform.Service exactly (per cmd/adapt.go), so the fake also satisfies
// *transform.Service by structural typing. It records the last request it
// received and returns the configured result.
type fakeTransformer struct {
	calls   int
	lastReq *transform.Request
	result  *core.TransformResult
	err     error
}

func (f *fakeTransformer) Transform(ctx context.Context, req *transform.Request) (*core.TransformResult, error) {
	_ = ctx // accept-and-may-ignore: fake records the request only
	f.calls++
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}
	if f.result != nil {
		return f.result, nil
	}
	return &core.TransformResult{OutputPath: req.OutputPath}, nil
}

func newAdaptTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

func TestRunAdapt_HappyPath(t *testing.T) {
	t.Parallel()

	io, out, errOut := newAdaptTestIO()
	fake := &fakeTransformer{}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/tmp/in.md",
		OutputPath:  "/tmp/out.md",
		Platform:    core.PlatformOpenCode,
	}

	require.NoError(t, runAdapt(opts))

	assert.Equal(t, 1, fake.calls, "transformer must be called exactly once")
	require.NotNil(t, fake.lastReq)
	assert.Equal(t, "/tmp/in.md", fake.lastReq.InputPath)
	assert.Equal(t, "/tmp/out.md", fake.lastReq.OutputPath)
	assert.Equal(t, core.PlatformOpenCode, fake.lastReq.Platform)

	assert.Equal(t, "wrote /tmp/out.md\n", out.String(), "stdout must contain the success line")
	assert.Empty(t, errOut.String(), "stderr must be empty when not verbose")
}

func TestRunAdapt_InvalidPlatformEmpty(t *testing.T) {
	t.Parallel()

	io, _, _ := newAdaptTestIO()
	fake := &fakeTransformer{}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/tmp/in.md",
		OutputPath:  "/tmp/out.md",
		Platform:    "",
	}

	err := runAdapt(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "error must wrap *core.ValidationError")
	assert.Equal(t, 0, fake.calls, "transformer must NOT be called when platform is invalid")
}

func TestRunAdapt_InvalidPlatformValue(t *testing.T) {
	t.Parallel()

	io, _, _ := newAdaptTestIO()
	fake := &fakeTransformer{}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/tmp/in.md",
		OutputPath:  "/tmp/out.md",
		Platform:    "windows-95",
	}

	err := runAdapt(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown platform "windows-95"`)
	assert.Equal(t, 0, fake.calls)
}

// runAdapt's contract: the production wiring
// (transform.NewService(parser.NewParser(), renderer.NewSerializer()))
// always populates a real Transformer. The opts.Transformer lazy field
// is the per-call injection seam — tests that want to assert the error
// path use a fakeTransformer via runAdapt directly.

func TestRunAdapt_TransformerErrorPropagates(t *testing.T) {
	t.Parallel()

	io, out, _ := newAdaptTestIO()
	cause := errors.New("boom")
	fake := &fakeTransformer{err: cause}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/tmp/in.md",
		OutputPath:  "/tmp/out.md",
		Platform:    core.PlatformClaudeCode,
	}

	err := runAdapt(opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause, "original error must be preserved in the chain")
	assert.True(t, errors.Is(err, cause))
	assert.Empty(t, out.String(), "stdout must be empty on error")
}

func TestRunAdapt_VerboseProgressToStderr(t *testing.T) {
	t.Parallel()

	io, out, errOut := newAdaptTestIO()
	io.Verbose = true
	fake := &fakeTransformer{}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/tmp/in.md",
		OutputPath:  "/tmp/out.md",
		Platform:    core.PlatformClaudeCode,
	}

	require.NoError(t, runAdapt(opts))

	assert.Contains(t, errOut.String(), "transforming /tmp/in.md",
		"verbose progress must go to stderr")
	assert.Contains(t, errOut.String(), "/tmp/out.md",
		"verbose progress must mention the output path")
	assert.Equal(t, "wrote /tmp/out.md\n", out.String(),
		"stdout must remain the success line only (no verbose leakage)")
}

func TestRunAdapt_StreamDisciplineNonVerbose(t *testing.T) {
	t.Parallel()

	io, out, errOut := newAdaptTestIO()
	fake := &fakeTransformer{}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/tmp/in.md",
		OutputPath:  "/tmp/out.md",
		Platform:    core.PlatformClaudeCode,
	}

	require.NoError(t, runAdapt(opts))

	assert.NotContains(t, out.String(), "transforming",
		"stdout must not contain verbose progress when verbose is off")
	assert.Empty(t, errOut.String(), "stderr must be empty when verbose is off")
}

func TestRunAdapt_TransformerReceivesExactRequest(t *testing.T) {
	t.Parallel()

	io, _, _ := newAdaptTestIO()
	fake := &fakeTransformer{}
	opts := &adaptOptions{
		IO:          io,
		Transformer: func() (Transformer, error) { return fake, nil },
		Ctx:         context.Background(),
		InputPath:   "/data/agent.md",
		OutputPath:  "/data/agent.out.md",
		Platform:    core.PlatformClaudeCode,
	}

	require.NoError(t, runAdapt(opts))
	require.NotNil(t, fake.lastReq)
	assert.Equal(t, "/data/agent.md", fake.lastReq.InputPath)
	assert.Equal(t, "/data/agent.out.md", fake.lastReq.OutputPath)
	assert.Equal(t, core.PlatformClaudeCode, fake.lastReq.Platform)
}

func TestNewCmdAdapt_RunFInjectionCapturesOpts(t *testing.T) {
	var captured *adaptOptions
	runF := func(opts *adaptOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdAdapt(f, runF)
	cmd.SetArgs([]string{"/tmp/in.md", "/tmp/out.md", "--platform", "opencode"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured, "runF must be invoked")
	assert.Equal(t, "/tmp/in.md", captured.InputPath)
	assert.Equal(t, "/tmp/out.md", captured.OutputPath)
	assert.Equal(t, "opencode", captured.Platform)
	require.NotNil(t, captured.IO)
	assert.Equal(t, io, captured.IO, "opts.IO must be the Factory's IOStreams")
	require.NotNil(t, captured.Ctx, "opts.Ctx must be set from c.Context()")
}

func TestNewCmdAdapt_NilRunFFallsBackToProduction(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	// NewCmdAdapt's production wiring constructs the Transformer
	// lazily inside runAdapt via
	// transform.NewService(parser.NewParser(), renderer.NewSerializer()).
	// The nil-runF path therefore exercises the full runAdapt →
	// parse → render → write pipeline. With a valid platform but a
	// missing input file, parse fails and the error surfaces through
	// cmdutil.ExitCodeFor → ExitCodeError (1).
	cmd := NewCmdAdapt(f, nil)
	cmd.SetArgs([]string{"/nonexistent.md", "/tmp/out.md", "--platform", "opencode"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing input file must surface as an error")
	assert.Empty(t, io.Out.(*bytes.Buffer).String(),
		"stdout must remain empty on the error path")
}

func TestNewCmdAdapt_RequiresPlatformFlag(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdAdapt(f, func(*adaptOptions) error { return nil })
	cmd.SetArgs([]string{"/tmp/in.md", "/tmp/out.md"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing required --platform flag must fail")
}

// slice-7 removed the Factory.Transformer lazy field and the
// `adaptTransformer(f)` factory helper. The Transformer is now
// constructed inside runAdapt via
// transform.NewService(parser.NewParser(), renderer.NewSerializer());
// test fakes are injected via runAdapt directly (see fakeTransformer
// at the top of this file).

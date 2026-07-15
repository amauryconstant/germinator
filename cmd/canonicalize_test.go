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

func (f *fakeCanonicalizer) Canonicalize(ctx context.Context, req *CanonicalizeRequest) (*core.CanonicalizeResult, error) {
	_ = ctx // accept-and-may-ignore: fake records the request only
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

func TestRunCanonicalize_InvalidDocType_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		docType string
	}{
		{name: "plural form", docType: "skills"},
		{name: "unknown type", docType: "bot"},
		{name: "empty string", docType: ""},
		{name: "uppercase", docType: "Agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			io, _, _ := newCanonicalizeTestIO()
			fake := &fakeCanonicalizer{}
			opts := &canonicalizeOptions{
				IO:            io,
				Canonicalizer: func() (Canonicalizer, error) { return fake, nil },
				Ctx:           context.Background(),
				InputPath:     "/tmp/in.md",
				OutputPath:    "/tmp/out.yaml",
				Platform:      core.PlatformOpenCode,
				DocType:       tt.docType,
			}

			err := runCanonicalize(opts)
			require.Error(t, err,
				"unknown --type must fail before any canonicalizer call")

			var verr *core.ValidationError
			require.True(t, errors.As(err, &verr),
				"error must wrap *core.ValidationError")
			assert.Equal(t, "type", verr.Field(),
				"ValidationError.Field() must be 'type' so the rendered error identifies the offending flag")
			assert.Equal(t, tt.docType, verr.Value())
			assert.NotEmpty(t, verr.Suggestions(),
				"unknown --type must carry suggestions listing the canonical types")
			assert.Equal(t, 0, fake.calls,
				"canonicalizer must NOT be called when --type is invalid")
		})
	}
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
	var captured *canonicalizeOptions
	runF := func(opts *canonicalizeOptions) error {
		captured = opts
		return nil
	}

	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
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
}

func TestNewCmdCanonicalize_RequiresPlatformAndTypeFlags(t *testing.T) {
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
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	cmd := NewCmdCanonicalize(f, func(*canonicalizeOptions) error { return nil })
	cmd.SetArgs([]string{"/tmp/in.md", "/tmp/out.yaml", "--platform", "opencode"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "missing --type flag must fail")
}

// slice-7 removed the Factory.Canonicalizer lazy field and the
// `canonicalizeCanonicalizer(f)` factory helper. The Canonicalizer is
// now constructed inside runCanonicalize via cmd.NewCanonicalizer();
// test fakes are injected via runCanonicalize directly (see
// fakeCanonicalizer at the top of this file).

func TestNewCanonicalizer_AdapterSatisfiesInterface(t *testing.T) {
	t.Parallel()

	// Compile-time interface check is already in canonicalize.go
	// (var _ Canonicalizer = (*canonicalizerAdapter)(nil)).
	// This test verifies the runtime contract: a value returned by
	// NewCanonicalizer() must accept the Canonicalize call shape that
	// the local Canonicalizer interface defines, and must propagate
	// the underlying typed error from canonicalizeDocument so callers
	// can errors.As their way to the right exit code.
	result, err := NewCanonicalizer().Canonicalize(context.Background(), &CanonicalizeRequest{
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

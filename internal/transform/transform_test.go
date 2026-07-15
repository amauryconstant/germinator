package transform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// writeCanonicalAgent writes a minimal canonical agent .md file with
// the given name. The body shape mirrors test/fixtures/agent-test.md
// so the production parser.LoadDocument can resolve it as a valid
// canonical Agent rather than failing the frontmatter check.
func writeCanonicalAgent(t *testing.T, dir, name, description string) string {
	t.Helper()
	path := filepath.Join(dir, "agent-"+name+".md")
	body := "---\nname: " + name + "\ndescription: " + description + "\ntools:\n  - bash\n---\n# Body\n"
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func newTestService() Service {
	return NewService(parser.NewParser(), renderer.NewSerializer())
}

func TestService_Transform_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
	}{
		{"claude-code agent", core.PlatformClaudeCode},
		{"opencode agent", core.PlatformOpenCode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			src := writeCanonicalAgent(t, tmp, "reviewer", "Reviews things")
			out := filepath.Join(tmp, "out.md")

			svc := newTestService()
			result, err := svc.Transform(context.Background(), &Request{
				InputPath:  src,
				OutputPath: out,
				Platform:   tt.platform,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, out, result.OutputPath)

			// Output file exists and is non-empty.
			content, err := os.ReadFile(out)
			require.NoError(t, err)
			assert.NotEmpty(t, content, "output file must contain rendered content")
		})
	}
}

func TestService_Transform_MissingInputFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	missing := filepath.Join(tmp, "no-such-agent.md")
	out := filepath.Join(tmp, "out.md")

	svc := newTestService()
	_, err := svc.Transform(context.Background(), &Request{
		InputPath:  missing,
		OutputPath: out,
		Platform:   core.PlatformOpenCode,
	})
	require.Error(t, err)
	// LoadDocument surfaces the missing-file error wrapped; we
	// don't assert a specific error type because parser error
	// wrapping is the parser's responsibility.
	assert.Contains(t, err.Error(), "no-such-agent")
}

func TestService_Transform_WriteError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	src := writeCanonicalAgent(t, tmp, "x", "x")
	// Output path inside a directory that cannot exist (parent is a
	// regular file, not a directory).
	parentFile := filepath.Join(tmp, "not-a-dir")
	require.NoError(t, os.WriteFile(parentFile, []byte("blocker"), 0o600))
	out := filepath.Join(parentFile, "out.md")

	svc := newTestService()
	_, err := svc.Transform(context.Background(), &Request{
		InputPath:  src,
		OutputPath: out,
		Platform:   core.PlatformOpenCode,
	})
	require.Error(t, err)
	// Write failures surface as *core.FileError so cmdutil can
	// dispatch on type via errors.As.
	var fe *core.FileError
	assert.ErrorAs(t, err, &fe)
	assert.Equal(t, "write", fe.Operation())
}

func TestService_Transform_NilRequest(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	_, err := svc.Transform(context.Background(), nil)
	require.Error(t, err)
	// LoadDocument on a zero-value path surfaces an immediate error;
	// we don't lock the exact message because the parser is the
	// authority on the input path contract.
}

func TestService_Transform_PlatformAssumedValid(t *testing.T) {
	t.Parallel()

	// The Service does not pre-validate Platform; that is the cmd
	// layer's job. We exercise an unsupported platform to confirm
	// the Service forwards the responsibility. The parser's
	// LoadDocument is the first hop and surfaces a *core.ConfigError
	// (typed platform-validation failure) so the cmd layer's
	// errors.As dispatch maps it to ExitCodeError (1) — exactly as
	// if the cmd layer had called core.ValidatePlatform itself.
	tmp := t.TempDir()
	src := writeCanonicalAgent(t, tmp, "x", "x")
	out := filepath.Join(tmp, "out.md")

	svc := newTestService()
	_, err := svc.Transform(context.Background(), &Request{
		InputPath:  src,
		OutputPath: out,
		Platform:   "windows-95",
	})
	require.Error(t, err)
	var ce *core.ConfigError
	assert.ErrorAs(t, err, &ce,
		"unsupported platform must surface as *core.ConfigError via the parser")
}

func TestNewService_ImplementsService(t *testing.T) {
	t.Parallel()

	s := NewService(parser.NewParser(), renderer.NewSerializer())
	assert.NotNil(t, s, "NewService must return a non-nil Service")
}

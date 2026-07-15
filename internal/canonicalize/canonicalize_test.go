package canonicalize_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/canonicalize"
	"gitlab.com/amoconst/germinator/internal/core"
)

// writePlatformDoc writes a Markdown file with YAML frontmatter
// matching the per-platform schema (e.g. claude-code fields like
// `permissionMode`, `model`) into the test temp dir and returns its
// absolute path. Tests scope their fixtures via t.TempDir so
// concurrent runs do not collide.
func writePlatformDoc(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestService_Canonicalize_HappyPath(t *testing.T) {
	t.Parallel()

	svc := canonicalize.NewService()

	tests := []struct {
		name     string
		body     string
		filename string
		platform string
		docType  string
		assertIn []string
	}{
		{
			name: "agent claude-code",
			body: `---
name: reviewer
description: reviews code
tools:
  - Bash
permissionMode: default
model: sonnet
---
Body`,
			filename: "agent-cc.md",
			platform: core.PlatformClaudeCode,
			docType:  "agent",
			assertIn: []string{"name: reviewer", "description: reviews code"},
		},
		{
			name: "skill claude-code",
			body: `---
name: git-release
description: Create releases
---
Body`,
			filename: "skill-cc.md",
			platform: core.PlatformClaudeCode,
			docType:  "skill",
			assertIn: []string{"name: git-release"},
		},
		{
			name: "command claude-code",
			body: `---
name: hello
description: says hello
template: say hello to {{name}}
---
Body`,
			filename: "command-cc.md",
			platform: core.PlatformClaudeCode,
			docType:  "command",
			assertIn: []string{"name: hello"},
		},
		{
			name: "memory claude-code",
			body: `---
paths:
  - src/**/*.go
content: |
  Tracks Go source files.
---
Body`,
			filename: "memory-cc.md",
			platform: core.PlatformClaudeCode,
			docType:  "memory",
			assertIn: []string{"content: |"},
		},
		{
			name: "agent opencode",
			body: `---
name: code-analyzer
description: Analyzes code patterns
tools:
  - bash
model: anthropic/claude-sonnet-4-20250514
behavior:
  mode: primary
  temperature: 0.3
  steps: 25
---
Body`,
			filename: "agent-oc.md",
			platform: core.PlatformOpenCode,
			docType:  "agent",
			assertIn: []string{"name: code-analyzer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			input := writePlatformDoc(t, tt.filename, tt.body)
			output := filepath.Join(dir, "output.yaml")

			result, err := svc.Canonicalize(context.Background(), &canonicalize.Request{
				InputPath:  input,
				OutputPath: output,
				Platform:   tt.platform,
				DocType:    tt.docType,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, output, result.OutputPath,
				"returned OutputPath must match the requested path")

			contents, err := os.ReadFile(output)
			require.NoError(t, err)
			out := string(contents)
			for _, snippet := range tt.assertIn {
				assert.True(t, strings.Contains(out, snippet),
					"output must contain %q (got: %s)", snippet, out)
			}
		})
	}
}

func TestService_Canonicalize_MissingInput(t *testing.T) {
	t.Parallel()

	svc := canonicalize.NewService()
	dir := t.TempDir()
	result, err := svc.Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  filepath.Join(dir, "no-such-doc.md"),
		OutputPath: filepath.Join(dir, "out.yaml"),
		Platform:   core.PlatformClaudeCode,
		DocType:    "agent",
	})
	require.Error(t, err)
	assert.Nil(t, result)
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"missing input must surface as *core.ParseError")
}

func TestService_Canonicalize_ValidationError(t *testing.T) {
	t.Parallel()

	svc := canonicalize.NewService()
	// Agent missing required `name` field → ValidateAgent errors.
	body := `---
description: missing name
tools:
  - Bash
permissionMode: default
model: sonnet
---
Body`
	input := writePlatformDoc(t, "agent-bad.md", body)

	result, err := svc.Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  input,
		OutputPath: filepath.Join(t.TempDir(), "out.yaml"),
		Platform:   core.PlatformClaudeCode,
		DocType:    "agent",
	})
	require.Error(t, err)
	assert.Nil(t, result)
	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr),
		"validation failure must surface as *core.ValidationError")
}

// TestService_Canonicalize_WriteError exercises the I/O write path:
// the output directory is intentionally read-only so os.WriteFile
// fails with EACCES / EROFS. The shell-package must wrap the cause
// in *core.FileError so cmd/cmdutil.ExitCodeFor maps it to exit 1.
func TestService_Canonicalize_WriteError(t *testing.T) {
	t.Parallel()

	if os.Getuid() == 0 {
		t.Skip("running as root — write-permission tests are unreliable")
	}

	svc := canonicalize.NewService()
	body := `---
name: reviewer
description: reviews code
tools:
  - Bash
permissionMode: default
model: sonnet
---
Body`
	input := writePlatformDoc(t, "agent-good.md", body)

	// Parent is a file, not a directory — WriteFile to a path
	// nested under a regular file fails with ENOTDIR on Linux.
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))

	result, err := svc.Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  input,
		OutputPath: filepath.Join(blocker, "out.yaml"),
		Platform:   core.PlatformClaudeCode,
		DocType:    "agent",
	})
	require.Error(t, err)
	assert.Nil(t, result)
	var ferr *core.FileError
	require.True(t, errors.As(err, &ferr),
		"write failure must surface as *core.FileError")
}

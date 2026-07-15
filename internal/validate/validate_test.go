package validate_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/validate"
)

// writeFixture writes a Markdown file with YAML frontmatter into the
// test temp dir and returns its absolute path. Tests scope their
// fixtures via t.TempDir so concurrent runs do not collide.
func writeFixture(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestService_Validate_HappyPath(t *testing.T) {
	t.Parallel()

	svc := validate.NewService()

	tests := []struct {
		name     string
		body     string
		filename string
		platform string
	}{
		{
			name: "valid agent claude-code",
			body: `---
name: reviewer
description: reviews code
tools:
  - Bash
permissionMode: default
model: sonnet
---
Body`,
			filename: "agent-valid.md",
			platform: core.PlatformClaudeCode,
		},
		{
			name: "valid skill claude-code",
			body: `---
name: git-release
description: Create releases
---
Body`,
			filename: "skill-valid.md",
			platform: core.PlatformClaudeCode,
		},
		{
			name: "valid command claude-code",
			body: `---
name: hello
description: says hello
template: say hello to {{name}}
---
Body`,
			filename: "command-valid.md",
			platform: core.PlatformClaudeCode,
		},
		{
			name: "valid memory claude-code",
			body: `---
name: project-context
description: shared project context
---
Body`,
			filename: "memory-valid.md",
			platform: core.PlatformClaudeCode,
		},
		{
			name: "valid agent opencode",
			body: `---
name: reviewer
description: reviews code
mode: primary
model: sonnet
---
Body`,
			filename: "agent-valid-oc.md",
			platform: core.PlatformOpenCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := writeFixture(t, tt.filename, tt.body)
			result, err := svc.Validate(context.Background(), &validate.Request{
				InputPath: path,
				Platform:  tt.platform,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Valid(),
				"expected no validation errors for %s, got %d: %v",
				tt.name, len(result.Errors), result.Errors)
			assert.Empty(t, result.Errors)
		})
	}
}

func TestService_Validate_UnrecognizableFilename(t *testing.T) {
	t.Parallel()

	svc := validate.NewService()
	path := writeFixture(t, "no-pattern-prefix.md", `---
name: x
description: y
---
Body`)

	result, err := svc.Validate(context.Background(), &validate.Request{
		InputPath: path,
		Platform:  core.PlatformClaudeCode,
	})
	require.Error(t, err)
	assert.Nil(t, result, "fatal errors return a nil result")
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"fatal error must wrap *core.ParseError")
	assert.Contains(t, err.Error(), "unrecognizable filename")
}

func TestService_Validate_MissingFile(t *testing.T) {
	t.Parallel()

	svc := validate.NewService()
	result, err := svc.Validate(context.Background(), &validate.Request{
		InputPath: filepath.Join(t.TempDir(), "does-not-exist.md"),
		Platform:  core.PlatformClaudeCode,
	})
	require.Error(t, err)
	assert.Nil(t, result)
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"parse failure must surface as *core.ParseError")
}

func TestService_Validate_InvalidDocType(t *testing.T) {
	t.Parallel()

	svc := validate.NewService()
	// filename has no type-prefix pattern, so parser.DetectType returns ""
	// and Validate surfaces an "unrecognizable filename" *core.ParseError.
	path := writeFixture(t, "totally-untyped-doc.md", "some content")

	result, err := svc.Validate(context.Background(), &validate.Request{
		InputPath: path,
		Platform:  core.PlatformClaudeCode,
	})
	require.Error(t, err)
	assert.Nil(t, result)
	var perr *core.ParseError
	require.True(t, errors.As(err, &perr),
		"unrecognizable input must surface as *core.ParseError")
}

// TestService_Validate_ValidationErrors demonstrates that joined
// validator errors (returned as a *core.ValidationError that wraps
// errors.Join(...)) are flattened into the result.Errors slice so
// callers can render each individually.
func TestService_Validate_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := validate.NewService()
	// Missing required name field → ValidateAgent errors.
	path := writeFixture(t, "agent-missing-name.md", `---
description: missing name
---
Body`)

	result, err := svc.Validate(context.Background(), &validate.Request{
		InputPath: path,
		Platform:  core.PlatformClaudeCode,
	})
	require.NoError(t, err, "validation errors are returned in result, not via err")
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Errors, "missing field must surface as at least one error")
	assert.False(t, result.Valid())
}

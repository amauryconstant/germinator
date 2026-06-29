package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

func TestFormatErrorDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "ParseError",
			err:      core.NewParseError("/tmp/foo.md", "bad yaml", errors.New("yaml: line 1")),
			contains: "parse failed at /tmp/foo.md",
		},
		{
			name:     "ValidationError",
			err:      core.NewValidationError("adapt", "name", "", "name is required"),
			contains: "validation failed: name is required",
		},
		{
			name:     "TransformError",
			err:      core.NewTransformError("render", "opencode", "render failed", nil),
			contains: "transform failed (render for opencode): render failed",
		},
		{
			name:     "FileError",
			err:      core.NewFileError("/tmp/out.md", "write", "permission denied", nil),
			contains: "write /tmp/out.md: permission denied",
		},
		{
			name:     "ConfigError",
			err:      core.NewConfigError("platform", "foo", "invalid platform"),
			contains: "config (platform): invalid platform",
		},
		{
			name:     "NotFoundError",
			err:      core.NewNotFoundError("library ref", "nonexistent-ref"),
			contains: "not found: nonexistent-ref",
		},
		{
			name:     "PartialSuccessError",
			err:      core.NewPartialSuccessError(3, 1, []core.InitializeError{}),
			contains: "partial success: 3 succeeded, 1 failed",
		},
		{
			name:     "GenericError",
			err:      errors.New("something broke"),
			contains: "Error: something broke",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			io := iostreams.Test()
			FormatError(io, tt.err)

			buf, ok := io.ErrOut.(*bytes.Buffer)
			require.True(t, ok)
			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}

func TestFormatErrorNil(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	FormatError(io, nil)

	buf, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.Equal(t, "", buf.String())
}

func TestFormatError_NotFound(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	err := core.NewNotFoundError("library ref", "nonexistent-ref")

	FormatError(io, err)

	stderr, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.Equal(t, "Error: not found: nonexistent-ref\n", stderr.String(),
		"NotFoundError must render canonical message to stderr")

	stdout, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	assert.Empty(t, stdout.String(),
		"NotFoundError must NOT write to stdout (stream discipline)")
}

func TestFormatError_PartialSuccessCrossPackage(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()

	ie := core.NewInitializeError("skill/missing", "/lib/skills/missing.md", "/out/.opencode/skills/missing/SKILL.md", errors.New("file not found"))
	psErr := core.NewPartialSuccessError(3, 1, []core.InitializeError{*ie})

	FormatError(io, psErr)

	buf, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	assert.Contains(t, got, "partial success: 3 succeeded, 1 failed")
	assert.Contains(t, got, "skill/missing")
}

func TestJSONExporter(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	data := payload{Name: "foo", Count: 7}

	io := iostreams.Test()
	exp := NewJSONExporter()
	require.NoError(t, exp.Write(io, data))

	buf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	assert.True(t, strings.HasSuffix(got, "\n"))

	var roundtripped payload
	require.NoError(t, json.Unmarshal([]byte(strings.TrimRight(got, "\n")), &roundtripped))
	assert.Equal(t, data, roundtripped)
}

func TestJSONExporterIndent(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	exp := NewJSONExporter()
	require.NoError(t, exp.Write(io, map[string]any{"a": 1, "b": "x"}))

	buf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	assert.Contains(t, got, "  \"a\": 1")
	assert.Contains(t, got, "  \"b\": \"x\"")
}

func TestTableExporter(t *testing.T) {
	t.Parallel()

	type row struct {
		Name  string `tab:"NAME"`
		Count int    `tab:"COUNT"`
	}

	data := []row{
		{Name: "alpha", Count: 1},
		{Name: "beta", Count: 22},
	}

	io := iostreams.Test()
	exp := NewTableExporter()
	require.NoError(t, exp.Write(io, data))

	buf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	require.GreaterOrEqual(t, len(lines), 3)
	assert.Contains(t, lines[0], "NAME")
	assert.Contains(t, lines[0], "COUNT")
	assert.Contains(t, lines[1], "alpha")
	assert.Contains(t, lines[1], "1")
	assert.Contains(t, lines[2], "beta")
	assert.Contains(t, lines[2], "22")
}

func TestTableExporterEmpty(t *testing.T) {
	t.Parallel()

	type row struct {
		Name string `tab:"NAME"`
	}

	io := iostreams.Test()
	exp := NewTableExporter()
	require.NoError(t, exp.Write(io, []row{}))

	buf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	assert.Equal(t, "", buf.String())
}

func TestTableExporterNonSlice(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	exp := NewTableExporter()
	err := exp.Write(io, "not a slice")
	assert.Error(t, err)
}

func TestAddOutputFlags(t *testing.T) {
	t.Parallel()

	target := ""
	cmd := newTestCmd("test")
	AddOutputFlags(cmd, &target)

	err := cmd.ParseFlags([]string{"--output", "json"})
	require.NoError(t, err)
	assert.Equal(t, "json", target)
}

func TestAddOutputFlagsDefault(t *testing.T) {
	t.Parallel()

	target := ""
	cmd := newTestCmd("test")
	AddOutputFlags(cmd, &target)

	err := cmd.ParseFlags([]string{})
	require.NoError(t, err)
	assert.Equal(t, DefaultOutputFormat, target)
}

func newTestCmd(use string) *cobra.Command {
	cmd := &cobra.Command{Use: use}
	return cmd
}

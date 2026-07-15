package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
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

func TestFormatError_OperationError(t *testing.T) {
	t.Run("basic_render", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := core.NewOperationError("register", "skill/commit", nil)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t, "Error: register: skill/commit\n", stderr.String(),
			"OperationError must render canonical message to stderr")

		stdout, ok := io.Out.(*bytes.Buffer)
		require.True(t, ok)
		assert.Empty(t, stdout.String(),
			"OperationError must NOT write to stdout (stream discipline)")
	})

	t.Run("with_cause", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		cause := errors.New("name taken by skill/x")
		err := core.NewOperationError("register", "skill/commit", cause)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got, "Error: register: skill/commit",
			"stderr must contain the canonical first line")
		assert.Contains(t, got, "name taken by skill/x",
			"stderr must contain the wrapped cause")
		assert.Contains(t, got, "  name taken by skill/x",
			"cause must be rendered on an indented second line")

		stdout, ok := io.Out.(*bytes.Buffer)
		require.True(t, ok)
		assert.Empty(t, stdout.String(),
			"OperationError with cause must NOT write to stdout")
	})

	t.Run("dispatch_precedence", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		base := core.NewOperationError("register", "skill/commit", nil)
		wrapped := fmt.Errorf("registering: %w", base)

		FormatError(io, wrapped)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got, "Error: register: skill/commit",
			"errors.As dispatch must select the OperationError branch")
		assert.NotContains(t, got, "registering:",
			"outer wrapping string must NOT appear in rendered output")
	})
}

func TestFormatError_WriteError(t *testing.T) {
	t.Parallel()

	t.Run("NewWriteError with cause", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := config.NewWriteError("write", "/tmp/out.toml", errors.New("disk full"))

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Equal(t,
			"Error: write /tmp/out.toml: disk full\n",
			got,
			"WriteError must render canonical message to stderr (no duplicated op/path prefix)")
		stdout, okOut := io.Out.(*bytes.Buffer)
		require.True(t, okOut)
		assert.Empty(t, stdout.String(),
			"WriteError must NOT write to stdout (stream discipline)")
	})

	t.Run("NewWriteErrorWithMessage no cause (already-exists path)", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := config.NewWriteErrorWithMessage(
			"create", "/tmp/cfg.toml",
			"config file already exists (use --force to overwrite)", nil,
		)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t,
			"Error: create /tmp/cfg.toml: config file already exists (use --force to overwrite)\n",
			stderr.String(),
			"already-exists path must render the user-friendly message verbatim")
	})
}

// TestFormatError_ObscureBranches covers the lower-frequency dispatch
// arms that Phase 1 added but never tested directly: ValidationError
// with suggestions + field, TransformError with cause, FileError with
// cause. Phase 4 added this to bring internal/output coverage above
// the 90% target in Phase 5 task 5.3.
func TestFormatError_ObscureBranches(t *testing.T) {
	t.Parallel()

	t.Run("ValidationError with suggestions and field", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := core.NewValidationError("adapt", "name", "Agent",
			"name must be unique").
			WithSuggestions([]string{"try a different name"})

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got, "Error: validation failed: name must be unique")
		assert.Contains(t, got, "(field: name)",
			"field must be rendered in parentheses after the message")
		assert.Contains(t, got, "Hint: try a different name",
			"suggestions must be rendered as 'Hint: ...' lines")
	})

	t.Run("TransformError with platform and cause", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		cause := errors.New("template parse failed")
		err := core.NewTransformError("render", "opencode", "render failed", cause)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got,
			"transform failed (render for opencode): render failed: template parse failed",
			"op + platform + message + cause must all appear in order")
	})

	t.Run("FileError with cause", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		cause := errors.New("EACCES")
		err := core.NewFileError("/tmp/out.md", "write", "permission denied", cause)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got,
			"write /tmp/out.md: permission denied: EACCES",
			"op + path + message + cause must all appear in order")
	})
}

func TestFormatError_InitializeError(t *testing.T) {
	t.Parallel()

	t.Run("renders via InitializeError.Error() delegation", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		cause := errors.New("file not found")
		err := core.NewInitializeError("skill/missing", "/lib/in.md", "/lib/out.md", cause)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Equal(t,
			"Error: initialize failed: skill/missing: output: /lib/out.md: file not found\n",
			got,
			"InitializeError must render its own Error() string prefixed by 'Error: '")
		stdout, okOut := io.Out.(*bytes.Buffer)
		require.True(t, okOut)
		assert.Empty(t, stdout.String(),
			"InitializeError must NOT write to stdout (stream discipline)")
	})
}

func TestFormatError_UsageError(t *testing.T) {
	t.Parallel()

	t.Run("clean-break wording via Flag()/Reason()", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := core.NewUsageError("--resources", "must be non-empty list of refs")

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Equal(t,
			"Error: --resources: must be non-empty list of refs\n",
			got,
			"UsageError must render via the clean-break wording (no 'flag needs an argument' prefix)")
		assert.NotContains(t, got, "flag needs an argument",
			"UsageError renderer MUST NOT contain the legacy Cobra-encoded phrasing")
		stdout, okOut := io.Out.(*bytes.Buffer)
		require.True(t, okOut)
		assert.Empty(t, stdout.String(),
			"UsageError must NOT write to stdout (stream discipline)")
	})

	t.Run("wrapped UsageError renders via typed dispatch", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := fmt.Errorf("validating flags: %w",
			core.NewUsageError("--resources", "must be non-empty list of refs"))

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t,
			"Error: --resources: must be non-empty list of refs\n",
			stderr.String(),
			"errors.As must traverse %w and dispatch to the UsageError arm; the outer 'validating flags:' prefix must NOT appear in the rendered output")
	})
}

func TestFormatError_CobraUsageError(t *testing.T) {
	t.Parallel()

	t.Run("renders via wrapped-cause delegation", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		err := core.MustNewCobraUsageError(errors.New("requires at least 1 arg(s), only received 0"))

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t,
			"Error: requires at least 1 arg(s), only received 0\n",
			stderr.String(),
			"CobraUsageError must delegate to the wrapped cause's Error() and prefix 'Error: '")
		stdout, okOut := io.Out.(*bytes.Buffer)
		require.True(t, okOut)
		assert.Empty(t, stdout.String(),
			"CobraUsageError must NOT write to stdout (stream discipline)")
	})
}

func TestFormatError_DispatchOrdering(t *testing.T) {
	t.Parallel()

	t.Run("wrapped InitializeError inside fmt.Errorf %%w", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		cause := errors.New("file not found")
		inner := core.NewInitializeError("skill/missing", "/lib/in.md", "/lib/out.md", cause)
		err := fmt.Errorf("loading library: %w", inner)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		assert.Equal(t,
			"Error: initialize failed: skill/missing: output: /lib/out.md: file not found\n",
			stderr.String(),
			"errors.As must traverse %%w and dispatch on the typed InitializeError")
		assert.NotContains(t, stderr.String(), "loading library",
			"the outer wrapping prefix must NOT appear; the typed case wins over the generic fallback")
	})

	t.Run("OperationError carrying InitializeError (OperationError precedence)", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		cause := errors.New("file not found")
		inner := core.NewInitializeError("skill/missing", "/lib/in.md", "/lib/out.md", cause)
		err := core.NewOperationError("init", "skill/commit", inner)

		FormatError(io, err)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got, "Error: init: skill/commit\n",
			"OperationError case precedes InitializeError; first matching case wins")
		assert.Contains(t, got, "initialize failed: skill/missing: output: /lib/out.md: file not found",
			"the wrapped InitializeError's body must appear as an indented cause line under the OperationError primary line")
	})

	t.Run("PartialSuccessError precedes NotFoundError", func(t *testing.T) {
		t.Parallel()
		io := iostreams.Test()
		nf := core.NewNotFoundError("skill", "missing")
		ie := core.NewInitializeError("skill/missing", "/in", "/out", nf)
		ps := core.NewPartialSuccessError(0, 1, []core.InitializeError{*ie})

		FormatError(io, ps)

		stderr, ok := io.ErrOut.(*bytes.Buffer)
		require.True(t, ok)
		got := stderr.String()
		assert.Contains(t, got, "partial success: 0 succeeded, 1 failed",
			"PartialSuccessError arm must precede NotFoundError so the partial-success renderer wins over the terse not-found renderer")
		assert.Contains(t, got, "skill/missing",
			"the per-resource InitializeError must appear in the partial-success listing")
		assert.NotContains(t, got, "Error: not found:",
			"the NotFoundError arm must NOT win when the chain also contains PartialSuccessError")
	})
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

func TestTableExporterFormatsAllKinds(t *testing.T) {
	t.Parallel()

	// Covers the formatCell switch arms (Bool, Int, Uint, Float,
	// String, struct) and indirectValue's non-pointer path. Phase 4
	// added this test to close the coverage gap introduced by
	// pre-existing table-exporter branches not exercised by
	// TestTableExporter (which only used string + int).
	type inner struct {
		X int `json:"x"`
	}
	type row struct {
		Str    string  `tab:"STR"`
		B      bool    `tab:"B"`
		I      int     `tab:"I"`
		U      uint64  `tab:"U"`
		F      float64 `tab:"F"`
		Nested inner   `tab:"NESTED"`
	}

	data := []row{
		{
			Str:    "hello",
			B:      true,
			I:      -42,
			U:      7,
			F:      3.14,
			Nested: inner{X: 99},
		},
	}

	io := iostreams.Test()
	exp := NewTableExporter()
	require.NoError(t, exp.Write(io, data))

	buf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	// tabwriter pads columns with spaces (default 2-wide); we assert on
	// column content rather than raw tab separators because the writer
	// rewrites the tabbing for human readability.
	assert.Contains(t, got, "STR")
	assert.Contains(t, got, "B")
	assert.Contains(t, got, "I")
	assert.Contains(t, got, "U")
	assert.Contains(t, got, "F")
	assert.Contains(t, got, "NESTED",
		"all headers must be rendered in struct order")
	assert.Contains(t, got, "hello")
	assert.Contains(t, got, "true")
	assert.Contains(t, got, "-42")
	assert.Contains(t, got, "3.14")
	assert.Contains(t, got, `{"x":99}`,
		"nested struct must JSON-compact to its single-line form")
}

func TestTableExporterNilPointerField(t *testing.T) {
	t.Parallel()

	// Covers formatCell's pointer-nil branch (returns empty string).
	type row struct {
		Name  string  `tab:"NAME"`
		Extra *string `tab:"EXTRA"`
	}

	val := "hello"
	data := []row{
		{Name: "alpha", Extra: &val},
		{Name: "beta", Extra: nil},
	}

	io := iostreams.Test()
	exp := NewTableExporter()
	require.NoError(t, exp.Write(io, data))

	buf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	require.GreaterOrEqual(t, len(lines), 3)
	assert.Contains(t, lines[1], "alpha")
	assert.Contains(t, lines[1], "hello",
		"non-nil pointer must dereference to its value")
	assert.Contains(t, lines[2], "beta",
		"nil pointer must render as empty cell; only the name field is non-empty")
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

func TestFormatResourcesList_Empty(t *testing.T) {
	t.Parallel()

	lib := &library.Library{Resources: map[string]map[string]library.Resource{}}

	got := FormatResourcesList(lib)
	assert.Equal(t, "No resources found.\n", got,
		"empty library must render the sentinel with a trailing newline")
}

func TestFormatResourcesList_GroupedWithDescriptions(t *testing.T) {
	t.Parallel()

	lib := &library.Library{
		Resources: map[string]map[string]library.Resource{
			string(library.ResourceTypeSkill): {
				"commit": {Path: "skills/commit.yaml", Description: "Commit helper"},
				"pr":     {Path: "skills/pr.yaml"},
			},
			string(library.ResourceTypeAgent): {
				"reviewer": {Path: "agents/reviewer.md", Description: "Code reviewer"},
			},
		},
	}

	got := FormatResourcesList(lib)

	assert.Contains(t, got, "Skills:\n", "skill group header is rendered")
	assert.Contains(t, got, "Agents:\n", "agent group header is rendered")
	assert.Contains(t, got, "  skill/commit - Commit helper\n",
		"resource with description is rendered with the description")
	assert.Contains(t, got, "  skill/pr\n",
		"resource without description omits the dash")
	assert.Contains(t, got, "  agent/reviewer - Code reviewer\n")

	idxSkill := strings.Index(got, "Skills:")
	idxAgent := strings.Index(got, "Agents:")
	require.Greater(t, idxSkill, -1)
	require.Greater(t, idxAgent, -1)
	assert.Less(t, idxSkill, idxAgent,
		"skills must be grouped before agents (canonical order)")
}

func TestFormatResourcesList_NilLibrarySafe(t *testing.T) {
	t.Parallel()

	got := FormatResourcesList(&library.Library{})
	assert.Equal(t, "No resources found.\n", got,
		"library with nil Resources map must render the empty sentinel")
}

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

// newRefreshTestIO returns the buffer-backed IOStreams that tests use
// to assert on captured Out / ErrOut. Mirrors the slice-5
// newInitTestIO, slice-6 newAddTestIO, and slice-6 newCreatePresetTestIO
// helpers. Panics if iostreams.Test() does not return *bytes.Buffer
// writers (it always does in this codebase, but guard anyway).
func newRefreshTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// makeRefreshTestLibrary scaffolds a minimal library dir with the
// provided resources registered and the corresponding files created
// on disk. Returns the resolved RootPath. Tests use this when they
// want a real library for Refresh to operate on.
//
// The `files` map keys are relative paths under the library root
// (e.g., "skills/commit.md"); the values are the file bodies
// (typically YAML frontmatter + a small body section).
func makeRefreshTestLibrary(
	t *testing.T,
	resources map[string]map[string]library.Resource,
	files map[string]string,
) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
	for rel, content := range files {
		fp := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(fp), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(fp), err)
		}
		if err := os.WriteFile(fp, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", fp, err)
		}
	}
	lib := &library.Library{
		Version:   "1",
		RootPath:  dir,
		Resources: resources,
		Presets:   map[string]library.Preset{},
	}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("save library: %v", err)
	}
	return dir
}

// T1 — Constructor wires opts correctly via runF injection. The
// captured *refreshOptions must carry all parsed flags plus the
// Factory-derived IO / Library / Ctx fields.
func TestNewCmdRefresh_ValidatesArgs(t *testing.T) {
	var captured *refreshOptions
	runF := func(opts *refreshOptions) error {
		captured = opts
		return nil
	}

	ios, _, _ := newRefreshTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	cmd := NewCmdRefresh(f, &libPath, runF)
	cmd.SetArgs([]string{"--dry-run", "--force", "--output", "json"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.True(t, captured.DryRun, "--dry-run must populate opts.DryRun")
	assert.True(t, captured.Force, "--force must populate opts.Force")
	assert.Equal(t, "json", captured.Output, "--output must populate opts.Output")
	assert.NotNil(t, captured.Library, "opts.Library must be wired by NewCmdRefresh")
	assert.NotNil(t, captured.IO, "opts.IO must be wired by NewCmdRefresh")
	assert.NotNil(t, captured.Ctx, "opts.Ctx must be wired by NewCmdRefresh")
}

// T2 — refreshLibrary helper: nil factory returns a nil loader so
// tests that don't care about the loader can ignore it (mirrors the
// slice-6 addLibrary / createPresetLibrary helpers).
func TestRefreshLibrary_NilFactoryReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, refreshLibrary(nil, ""),
		"refreshLibrary(nil, ...) returns nil so opts.Library is unset")
}

// T2b — refreshLibrary closure honors cfg.Library when f.Config is wired.
// Pins task 4.4's nil-safe closure pattern: the cfgPath inside the
// closure must come from f.Config().Library, falling through silently
// when f.Config is unset. Sequential (NOT t.Parallel) because
// t.Setenv is incompatible with parallel subtests per golang-testing
// Rule 4.
func TestRefreshLibrary_HonorsConfigLibrary(t *testing.T) {
	cfg := &config.Config{Library: "/from/cfg/path"}

	f := &cmdutil.Factory{
		RootContext:     context.Background(),
		CompletionCache: cmdutil.NewCompletionCache(),
	}
	f.Config = func() (*config.Config, error) { return cfg, nil }
	t.Setenv("GERMINATOR_LIBRARY", "")

	loader := refreshLibrary(f, "")
	require.NotNil(t, loader, "refreshLibrary must return a non-nil loader when f is non-nil")

	// Invoke the closure; we don't care if LoadLibrary succeeds —
	// we only need to prove that the resolved path (which FindLibrary
	// derives from cfgPath) reflects cfg.Library. We assert via the
	// library.LoadLibrary error message, which embeds the resolved
	// path. Use a non-existent path that surfaces a recognizable
	// error containing the cfg.Library path.
	_, err := loader()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "/from/cfg/path",
		"resolved library path MUST reflect cfg.Library when flag and env are unset")
}

// T2c — refreshLibrary closure survives f.Config == nil without panicking.
// Defense-in-depth for test paths that build a bare cmdutil.NewFactory(...)
// without setting f.Config. The closure must gracefully fall through
// to env-only resolution. Sequential (NOT t.Parallel) because
// t.Setenv is incompatible with parallel subtests per golang-testing
// Rule 4.
func TestRefreshLibrary_FConfigIsNilFallsBack(t *testing.T) {
	f := &cmdutil.Factory{
		RootContext:     context.Background(),
		CompletionCache: cmdutil.NewCompletionCache(),
		// f.Config intentionally left nil.
	}
	t.Setenv("GERMINATOR_LIBRARY", "/from/env/path")

	loader := refreshLibrary(f, "")
	require.NotNil(t, loader, "loader must be non-nil even when f.Config is nil")

	_, err := loader()
	require.Error(t, err, "library load is expected to fail on the bogus env path")
	assert.Contains(t, err.Error(), "/from/env/path",
		"resolved path MUST fall through to env when f.Config is nil")
}

// T3 — refreshOptions struct shape: declares exactly the spec-named
// fields; reflection guards against accidental drops or renames.
func TestRefreshOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(refreshOptions{})
	want := map[string]bool{
		"IO":              true,
		"Library":         true,
		"Ctx":             true,
		"DryRun":          true,
		"Force":           true,
		"Output":          true,
		"CompletionCache": true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}
	assert.Equal(t, want, got,
		"refreshOptions must declare exactly the spec-named fields")
}

// T4 — refresherLibrary interface satisfied by *library.Library:
// the var _ refresherLibrary = (*library.Library)(nil) compile-time
// check in library_refresh.go is exercised at runtime by this test.
func TestRefresherLibraryInterfaceSatisfied(t *testing.T) {
	t.Parallel()

	lib := &library.Library{}
	var rl refresherLibrary = lib
	assert.NotNil(t, rl, "*library.Library must satisfy refresherLibrary")
}

// T5 — Spec scenario "Unchanged resources reported": a registered
// resource matching library.yaml exactly produces an Unchanged
// section in plain output (per design Decision 7).
func TestRunRefresh_HappyPathAllUnchanged(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"commit": {Path: "skills/commit.md", Description: "Commit skill"},
			},
		},
		map[string]string{
			"skills/commit.md": "---\nname: commit\ndescription: Commit skill\n---\n# Commit\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Output: "plain",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.Contains(t, got, "Unchanged:",
		"all-matching refresh must render an Unchanged section")
	assert.Contains(t, got, "[unchanged] skill/commit",
		"Unchanged section must list the matching resource")
	assert.NotContains(t, got, "Refreshed:",
		"all-matching refresh must not render a Refreshed section")
	assert.NotContains(t, got, "Errors:",
		"all-matching refresh must not render an Errors section")
}

// T6 — Description drift triggers a Refreshed entry in plain output.
func TestRunRefresh_Refreshed_DescriptionDrift(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"commit": {Path: "skills/commit.md", Description: "old description"},
			},
		},
		map[string]string{
			"skills/commit.md": "---\nname: commit\ndescription: new description\n---\n# Commit\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Output: "plain",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.Contains(t, got, "Refreshed:",
		"description drift must render a Refreshed section")
	assert.Contains(t, got, "[refreshed] skill/commit: description",
		"Refreshed entry must carry the (ref, field) pair")
}

// T7 — Missing registered file: silently skipped (no section
// rendered). This matches the refresher.go behavior that returns
// from processResource without populating any result slice when
// the file is neither at the registered path nor anywhere in the
// type directory.
func TestRunRefresh_Skipped_MissingFile(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"ghost": {Path: "skills/ghost.md", Description: "Ghost skill"},
			},
		},
		map[string]string{},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Output: "plain",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.NotContains(t, got, "Refreshed:",
		"missing file must not render a Refreshed section")
	assert.NotContains(t, got, "Errors:",
		"missing file must not render an Errors section")
	assert.NotContains(t, got, "Unchanged:",
		"missing file must not render an Unchanged section")
	assert.NotContains(t, got, "Skipped:",
		"missing file is silent (refresh.go records no entry)")
}

// T8 — Name mismatch surfaces an Errors entry plus a Skipped entry
// (per recordNameMismatch in refresher.go).
func TestRunRefresh_Error_NameMismatch(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"commit": {Path: "skills/commit.md", Description: "Commit"},
			},
		},
		map[string]string{
			// Frontmatter name != registered name => mismatch.
			"skills/commit.md": "---\nname: different\ndescription: Commit\n---\n# Commit\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Output: "plain",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.Contains(t, got, "Errors:",
		"name mismatch must render an Errors section")
	assert.Contains(t, got, "[error] skill/commit",
		"Errors section must carry the offending ref")
	assert.Contains(t, got, "name_mismatch",
		"Errors section must report the mismatch reason")
	assert.Contains(t, got, "Skipped:",
		"name mismatch records both Errors and Skipped (per refresher.go)")
}

// T9 — Dry-run: prepend "Dry-run: no changes made" in plain output,
// even when Refreshed entries would have been written in non-dry-
// run mode. The dry-run prefix is a literal header, not a literal
// claim about whether the scan produced changes.
func TestRunRefresh_DryRun_PlainPrefix(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"commit": {Path: "skills/commit.md", Description: "old"},
			},
		},
		map[string]string{
			"skills/commit.md": "---\nname: commit\ndescription: new\n---\n# C\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:      ios,
		Ctx:     context.Background(),
		Output:  "plain",
		DryRun:  true,
		Library: func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.True(t,
		bytes.HasPrefix([]byte(got), []byte("Dry-run: no changes made\n")),
		"dry-run plain output must start with 'Dry-run: no changes made'; got: %q", got)
	assert.Contains(t, got, "Refreshed:",
		"dry-run refresh with drift still renders the Refreshed section")

	// Dry-run must not mutate library.yaml on disk.
	reloaded, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	assert.Equal(t, "old", reloaded.Resources["skill"]["commit"].Description,
		"dry-run must not persist the description update")
}

// T10 — JSON output: payload includes refreshed/unchanged/skipped/
// errors keys per design Decision 7. The payload uses stable field
// names (no omitempty) so consumers can rely on the shape.
func TestRunRefresh_JSON(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"a": {Path: "skills/a.md", Description: "A"},
				"b": {Path: "skills/b.md", Description: "old B"},
			},
		},
		map[string]string{
			"skills/a.md": "---\nname: a\ndescription: A\n---\n# A\n",
			"skills/b.md": "---\nname: b\ndescription: new B\n---\n# B\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:      ios,
		Ctx:     context.Background(),
		Output:  "json",
		DryRun:  true,
		Library: func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runRefresh(opts))

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(out.Bytes(), &got),
		"JSON output must parse; got: %q", out.String())

	assert.Contains(t, got, "refreshed", "JSON payload must include 'refreshed'")
	assert.Contains(t, got, "unchanged", "JSON payload must include 'unchanged'")
	assert.Contains(t, got, "skipped", "JSON payload must include 'skipped'")
	assert.Contains(t, got, "errors", "JSON payload must include 'errors'")

	// Refreshed section must carry b (description drift), Unchanged
	// must carry a (no drift). Validate the inner arrays.
	var refreshed []map[string]any
	require.NoError(t, json.Unmarshal(got["refreshed"], &refreshed))
	assert.Len(t, refreshed, 1, "Refreshed must carry exactly one entry (skill/b)")
	assert.Equal(t, "skill/b", refreshed[0]["ref"])

	var unchanged []map[string]any
	require.NoError(t, json.Unmarshal(got["unchanged"], &unchanged))
	assert.Len(t, unchanged, 1, "Unchanged must carry exactly one entry (skill/a)")
	assert.Equal(t, "skill/a", unchanged[0]["ref"])
}

// T11 — Table output: renders rows for the Refreshed entries with
// the (REF, FIELD, OLD, NEW) column headers.
func TestRunRefresh_Table(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"a": {Path: "skills/a.md", Description: "old A"},
			},
		},
		map[string]string{
			"skills/a.md": "---\nname: a\ndescription: new A\n---\n# A\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:      ios,
		Ctx:     context.Background(),
		Output:  "table",
		DryRun:  true,
		Library: func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.NotEmpty(t, got, "table output must write to stdout")
	// TableExporter renders tab:"HEADER" struct tags as the header row.
	assert.Contains(t, got, "REF", "table must render REF column header")
	assert.Contains(t, got, "FIELD", "table must render FIELD column header")
	assert.Contains(t, got, "OLD", "table must render OLD column header")
	assert.Contains(t, got, "NEW", "table must render NEW column header")
	assert.Contains(t, got, "skill/a", "table row must carry the resource ref")
}

// T12 — Verbosef: when IOStreams.Verbose is true, runRefresh logs
// a progress line to ErrOut. The IOStreams.Verbosef check short-
// circuits when Verbose is false (the default), so this test must
// flip the flag explicitly.
func TestRunRefresh_Verbosef(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"a": {Path: "skills/a.md", Description: "A"},
			},
		},
		map[string]string{
			"skills/a.md": "---\nname: a\ndescription: A\n---\n# A\n",
		},
	)

	ios, _, errOut := newRefreshTestIO()
	ios.Verbose = true
	opts := &refreshOptions{
		IO:      ios,
		Ctx:     context.Background(),
		Output:  "plain",
		Library: func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runRefresh(opts))
	assert.Contains(t, errOut.String(), "refreshing library at",
		"verbose mode must log the refresh progress to ErrOut")
}

// T13 — Library load failure: opts.Library returns an error →
// runRefresh wraps it with "loading library" and returns. main.go's
// centralized error handler renders the chain via output.FormatError.
func TestRunRefresh_LibraryLoadError(t *testing.T) {
	ios, _, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Output: "plain",
		Library: func() (*library.Library, error) {
			return nil, errors.New("boom")
		},
	}

	err := runRefresh(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading library",
		"load failure must be wrapped with 'loading library'")
	assert.Contains(t, err.Error(), "boom",
		"wrapped error must preserve the underlying cause")
}

// T14 — Unchanged section: when the LastSynced field is empty (no
// mtime), the rendered entry shows "(no mtime)" instead of an empty
// timestamp. Defends against a stale frontend that prints "( )".
func TestRunRefresh_Plain_UnchangedSectionNoMtime(t *testing.T) {
	libDir := makeRefreshTestLibrary(t,
		map[string]map[string]library.Resource{
			"skill": {
				"a": {Path: "skills/a.md", Description: "A"},
			},
		},
		map[string]string{
			"skills/a.md": "---\nname: a\ndescription: A\n---\n# A\n",
		},
	)

	ios, out, _ := newRefreshTestIO()
	opts := &refreshOptions{
		IO:      ios,
		Ctx:     context.Background(),
		Output:  "plain",
		Library: func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runRefresh(opts))
	got := out.String()
	assert.Contains(t, got, "Unchanged:")
	assert.Regexp(t, `\[unchanged\] skill/a \([^)]+\)`, got,
		"Unchanged entry must render in the '  [unchanged] ref (...)' form")
}

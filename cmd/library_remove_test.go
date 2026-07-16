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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

// newRemoveTestIO returns the buffer-backed IOStreams that remove
// tests use to assert on captured Out / ErrOut. Mirrors the slice-5
// newInitTestIO / slice-6 newAddTestIO helpers. Panics if
// iostreams.Test() does not return *bytes.Buffer writers.
func newRemoveTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// removeResourceFixture scaffolds a minimal library dir with a real
// skill file already registered in library.yaml. Returns the
// resolved RootPath. The physical file is at
// `<dir>/skills/commit.md` so a successful removal must delete it
// (os.Remove path).
func removeResourceFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
	const name = "commit"
	const description = "Git commit best practices"
	body := "---\nname: " + name + "\ndescription: " + description + "\n---\n# body\n"
	path := filepath.Join(dir, "skills", name+".md")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write resource file: %v", err)
	}
	lib := &library.Library{
		Version:  "1",
		RootPath: dir,
		Resources: map[string]map[string]library.Resource{
			"skill": {
				name: {Path: "skills/" + name + ".md", Description: description},
			},
		},
		Presets: map[string]library.Preset{},
	}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("save library: %v", err)
	}
	return dir
}

// removePresetFixture scaffolds a minimal library dir with a
// preset already registered. Returns the resolved RootPath.
func removePresetFixture(t *testing.T, refs []string) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
	const presetName = "wp"
	const description = "workflow"
	lib := &library.Library{
		Version:   "1",
		RootPath:  dir,
		Resources: map[string]map[string]library.Resource{},
		Presets: map[string]library.Preset{
			presetName: {
				Name:        presetName,
				Description: description,
				Resources:   refs,
			},
		},
	}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("save library: %v", err)
	}
	return dir
}

// removeConflictFixture scaffolds a library where a preset
// references the resource the test will try to remove. The
// physical file is also created so the test can assert that the
// conflict path does not partially delete it.
func removeConflictFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
	const name = "commit"
	body := "---\nname: " + name + "\ndescription: referenced\n---\n# body\n"
	if err := os.WriteFile(filepath.Join(dir, "skills", name+".md"), []byte(body), 0o600); err != nil {
		t.Fatalf("write resource file: %v", err)
	}
	lib := &library.Library{
		Version:  "1",
		RootPath: dir,
		Resources: map[string]map[string]library.Resource{
			"skill": {
				name: {Path: "skills/" + name + ".md", Description: "referenced"},
			},
		},
		Presets: map[string]library.Preset{
			"wp": {
				Name:        "wp",
				Description: "uses the resource",
				Resources:   []string{"skill/" + name},
			},
		},
	}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("save library: %v", err)
	}
	return dir
}

// T1 — Constructor wires opts correctly for the `resource` sub-command
// via runF injection. Verifies the legacy positional `<ref>` is
// preserved (no --type / --name flag substitution).
func TestNewCmdRemove_Resource_ValidatesArgs(t *testing.T) {
	var captured *removeOptions
	runF := func(opts *removeOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	ios, _, _ := newRemoveTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	f.Library = func() (*library.Library, error) {
		return library.LoadLibrary(context.Background(), t.TempDir())
	}
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdRemove(f, nil, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "resource", "skill/commit", "--force", "--output", "json"))
	require.NotNil(t, captured)
	assert.Equal(t, "skill/commit", captured.Ref)
	assert.Empty(t, captured.PresetName,
		"resource sub-command must NOT populate PresetName")
	assert.True(t, captured.Force)
	assert.Equal(t, "json", captured.Output)
	assert.NotNil(t, captured.Library, "opts.Library must be wired by NewCmdRemove")
	assert.NotNil(t, captured.IO)
	assert.NotNil(t, captured.Ctx)
}

// T2 — Constructor wires opts correctly for the `preset` sub-command
// via runF injection. Verifies the legacy positional `<name>` is
// preserved.
func TestNewCmdRemove_Preset_ValidatesArgs(t *testing.T) {
	var captured *removeOptions
	runF := func(opts *removeOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	ios, _, _ := newRemoveTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	f.Library = func() (*library.Library, error) {
		return library.LoadLibrary(context.Background(), t.TempDir())
	}
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdRemove(f, nil, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "preset", "git-workflow", "--force", "--output", "json"))
	require.NotNil(t, captured)
	assert.Equal(t, "git-workflow", captured.PresetName)
	assert.Empty(t, captured.Ref,
		"preset sub-command must NOT populate Ref")
	assert.True(t, captured.Force)
	assert.Equal(t, "json", captured.Output)
	assert.NotNil(t, captured.Library, "opts.Library must be wired by NewCmdRemove")
	assert.NotNil(t, captured.IO)
	assert.NotNil(t, captured.Ctx)
}

// T3 — Resource sub-command requires exactly 1 positional arg
// (cobra.ExactArgs(1) validator). cobra returns "accepts 1 arg(s),
// received 0" which cmdutil.ExitCodeFor currently maps to the
// default-error case (exit 1) because the prefix list does not
// include "accepts". We assert the error is returned; the exit code
// mapping is a separate concern tracked outside this task.
func TestNewCmdRemove_Resource_RequiresArg(t *testing.T) {
	ios, _, _ := newRemoveTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	err := executeCmd(t, func() any {
		cmd := NewCmdRemove(f, nil, func(*removeOptions) error { return nil })
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "resource")
	require.Error(t, err, "missing ref must fail Cobra ExactArgs(1)")
	assert.Contains(t, err.Error(), "accepts",
		"cobra ExactArgs(1) error must mention the arg count")
}

// T4 — Preset sub-command requires exactly 1 positional arg.
func TestNewCmdRemove_Preset_RequiresArg(t *testing.T) {
	ios, _, _ := newRemoveTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	err := executeCmd(t, func() any {
		cmd := NewCmdRemove(f, nil, func(*removeOptions) error { return nil })
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "preset")
	require.Error(t, err, "missing preset name must fail Cobra ExactArgs(1)")
	assert.Contains(t, err.Error(), "accepts",
		"cobra ExactArgs(1) error must mention the arg count")
}

// T5 — Resource removal happy path: the file is deleted, the
// library.yaml is updated, and the plain "Removed resource:" line
// appears on stdout.
func TestRunRemove_Resource_Success(t *testing.T) {
	libDir := removeResourceFixture(t)

	ios, out, errOut := newRemoveTestIO()
	opts := &removeOptions{
		IO:  ios,
		Ctx: context.Background(),
		Ref: "skill/commit",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	assert.Contains(t, out.String(), "Removed resource: skill/commit",
		"plain output must contain the byte-identical 'Removed resource: <ref>' line")
	assert.Empty(t, errOut.String(), "no errors on success")

	// The physical file is gone.
	physical := filepath.Join(libDir, "skills", "commit.md")
	if _, err := os.Stat(physical); !os.IsNotExist(err) {
		t.Errorf("physical file should be deleted; stat err = %v", err)
	}

	// library.yaml no longer references the resource.
	lib, lerr := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, lerr)
	if _, exists := lib.Resources["skill"]["commit"]; exists {
		t.Errorf("library.yaml must drop the resource entry after remove")
	}
}

// T6 — Resource removal surfaces a missing physical file as a
// typed NotFoundError (Phase 3.10, Design Decision #6: idempotent
// removal becomes non-idempotent). The library.yaml is NOT
// mutated; the error path emits no output.
func TestRunRemove_Resource_MissingFile(t *testing.T) {
	libDir := removeResourceFixture(t)

	// Pre-delete the physical file (but keep library.yaml pointing
	// at it) to simulate a partial prior state.
	physical := filepath.Join(libDir, "skills", "commit.md")
	if err := os.Remove(physical); err != nil {
		t.Fatalf("pre-delete physical file: %v", err)
	}

	ios, out, errOut := newRemoveTestIO()
	opts := &removeOptions{
		IO:  ios,
		Ctx: context.Background(),
		Ref: "skill/commit",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	// Phase 3.10 (Design Decision #6): idempotent removal becomes
	// non-idempotent — a missing physical file now surfaces as
	// *core.NotFoundError (entity "library file", key = path) so the
	// caller can distinguish "already gone" from "I removed it". The
	// prior silent os.IsNotExist swallow was the B-014 review finding.
	err := runRemove(opts)
	require.Error(t, err, "missing physical file must surface *core.NotFoundError (Phase 3.10)")

	var nf *core.NotFoundError
	require.True(t, errors.As(err, &nf),
		"error must be *core.NotFoundError, got %T: %v", err, err)
	assert.Equal(t, "library file", nf.Entity)
	assert.Equal(t, physical, nf.Key)
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"missing physical file must map to ExitCodeError (1) via NotFoundError branch")

	// The library.yaml entry is still present — the missing-file
	// surface happens before any state mutation.
	lib, lerr := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, lerr)
	_, exists := lib.Resources["skill"]["commit"]
	assert.True(t, exists, "library.yaml entry must NOT be removed on missing-physical-file")
	assert.Empty(t, out.String(), "stdout must be empty on error path")
	assert.Empty(t, errOut.String(), "stderr must be empty — single-handling rule")
}

// T7 — Resource removal where a preset references the resource
// returns an error (not a NotFoundError; the underlying
// gerrors.NewFileError wraps the message). The file is NOT deleted.
func TestRunRemove_Resource_PresetReferenceConflict(t *testing.T) {
	libDir := removeConflictFixture(t)

	ios, out, errOut := newRemoveTestIO()
	opts := &removeOptions{
		IO:    ios,
		Ctx:   context.Background(),
		Ref:   "skill/commit",
		Force: true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runRemove(opts)
	require.Error(t, err,
		"--force does not override the preset-reference safety check")
	assert.Empty(t, out.String(),
		"stdout must be empty on the error path")
	assert.Empty(t, errOut.String(),
		"stderr must be empty on the error path; FormatError is the writer")

	// The error message must mention the blocking preset name.
	assert.Contains(t, err.Error(), "wp",
		"error must mention the blocking preset name")

	// File is still present (no partial deletion).
	if _, err := os.Stat(filepath.Join(libDir, "skills", "commit.md")); err != nil {
		t.Errorf("physical file must NOT be deleted on conflict; stat err = %v", err)
	}
}

// T8 — Preset removal happy path: the preset is dropped from
// library.yaml, the resource it referenced is untouched, and the
// "Removed preset:" line appears on stdout.
func TestRunRemove_Preset_Success(t *testing.T) {
	libDir := removePresetFixture(t, []string{"skill/commit"})

	ios, out, errOut := newRemoveTestIO()
	opts := &removeOptions{
		IO:         ios,
		Ctx:        context.Background(),
		PresetName: "wp",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	assert.Contains(t, out.String(), "Removed preset: wp",
		"plain output must contain the byte-identical 'Removed preset: <name>' line")
	assert.Empty(t, errOut.String())

	lib, lerr := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, lerr)
	if _, exists := lib.Presets["wp"]; exists {
		t.Errorf("library.yaml must drop the preset entry after remove")
	}
}

// T9 — Preset removal of a non-existent preset returns a
// *core.NotFoundError. cmdutil.ExitCodeFor maps NotFoundError to
// ExitCodeUsage (2) per the existing branch in
// internal/cmdutil/exit.go.
func TestRunRemove_Preset_NotFound(t *testing.T) {
	libDir := removePresetFixture(t, []string{"skill/commit"})

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:         ios,
		Ctx:        context.Background(),
		PresetName: "ghost",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runRemove(opts)
	require.Error(t, err)
	assert.Empty(t, out.String(),
		"stdout must be empty on the error path")

	var notFound *core.NotFoundError
	require.True(t, errors.As(err, &notFound),
		"expected *core.NotFoundError, got %T: %v", err, err)
	assert.Equal(t, "preset", notFound.Entity)
	assert.Equal(t, "ghost", notFound.Key)
	// Per enforce-error-discipline (Phase 1.1): *core.NotFoundError now
	// maps to ExitCodeError (1) — a runtime lookup miss is an
	// operational error, not a user-input validation error.
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"missing preset must map to ExitCodeError (1) via NotFoundError branch")
}

// T10 — Resource removal --output json produces a JSON payload with
// the spec's required fields (type, resourceType, name, fileDeleted,
// libraryPath). Use RemoveResourceOutput for the shape.
func TestRunRemove_Resource_OutputJSON(t *testing.T) {
	libDir := removeResourceFixture(t)

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Ref:    "skill/commit",
		Output: "json",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))

	var parsed library.RemoveResourceOutput
	require.NoError(t, json.Unmarshal(out.Bytes(), &parsed),
		"output must be valid JSON: %q", out.String())
	assert.Equal(t, "resource", parsed.Type)
	assert.Equal(t, "skill", parsed.ResourceType)
	assert.Equal(t, "commit", parsed.Name)
	assert.Equal(t, filepath.Join(libDir, "skills", "commit.md"), parsed.FileDeleted)
	assert.Equal(t, libDir, parsed.LibraryPath)
}

// T11 — Preset removal --output json produces a JSON payload with
// the spec's required fields (type, name, resourcesRemoved).
func TestRunRemove_Preset_OutputJSON(t *testing.T) {
	libDir := removePresetFixture(t, []string{"skill/commit", "skill/pr"})

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:         ios,
		Ctx:        context.Background(),
		PresetName: "wp",
		Output:     "json",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))

	var parsed library.RemovePresetOutput
	require.NoError(t, json.Unmarshal(out.Bytes(), &parsed),
		"output must be valid JSON: %q", out.String())
	assert.Equal(t, "preset", parsed.Type)
	assert.Equal(t, "wp", parsed.Name)
	assert.Equal(t, []string{"skill/commit", "skill/pr"}, parsed.ResourcesRemoved)
}

// T12 — Resource removal --output table renders a single row with
// columns "REF" and "ACTION" (per the spec scenario "Table output").
func TestRunRemove_Resource_OutputTable(t *testing.T) {
	libDir := removeResourceFixture(t)

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:     ios,
		Ctx:    context.Background(),
		Ref:    "skill/commit",
		Output: "table",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	got := out.String()
	assert.Contains(t, got, "REF")
	assert.Contains(t, got, "ACTION")
	assert.Contains(t, got, "skill/commit")
	assert.Contains(t, got, "removed")
}

// T13 — Preset removal --output table renders a single row with
// columns "NAME" and "ACTION" (per the spec scenario "Table output"
// for presets).
func TestRunRemove_Preset_OutputTable(t *testing.T) {
	libDir := removePresetFixture(t, []string{"skill/commit"})

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:         ios,
		Ctx:        context.Background(),
		PresetName: "wp",
		Output:     "table",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	got := out.String()
	assert.Contains(t, got, "NAME")
	assert.Contains(t, got, "ACTION")
	assert.Contains(t, got, "wp")
	assert.Contains(t, got, "removed")
}

// T14 — Plain output for resource removal: "Removed resource: <ref>"
// line is byte-identical to the legacy pre-change build.
func TestRunRemove_Resource_OutputPlain(t *testing.T) {
	libDir := removeResourceFixture(t)

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:  ios,
		Ctx: context.Background(),
		Ref: "skill/commit",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	assert.Equal(t, "Removed resource: skill/commit\n", out.String(),
		"plain output must be byte-identical to legacy build")
}

// T15 — Plain output for preset removal: "Removed preset: <name>"
// line is byte-identical to the legacy pre-change build.
func TestRunRemove_Preset_OutputPlain(t *testing.T) {
	libDir := removePresetFixture(t, []string{"skill/commit"})

	ios, out, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:         ios,
		Ctx:        context.Background(),
		PresetName: "wp",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	assert.Equal(t, "Removed preset: wp\n", out.String(),
		"plain output must be byte-identical to legacy build")
}

// T16 — removeOptions struct shape: declares exactly the
// spec-named fields; reflection guards against accidental drops
// or renames.
func TestRemoveOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(removeOptions{})
	want := map[string]bool{
		"IO":              true,
		"Library":         true,
		"Ctx":             true,
		"Ref":             true,
		"PresetName":      true,
		"Force":           true,
		"Output":          true,
		"CompletionCache": true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}

	assert.Equal(t, want, got,
		"removeOptions must declare exactly the spec-named fields")
}

// T17 — removerLibrary interface satisfied by *library.Library:
// compile-time check is exercised by the var _ removerLibrary
// declaration in library_remove.go; this test asserts at runtime
// that the contract holds.
func TestRemoverLibraryInterfaceSatisfied(t *testing.T) {
	t.Parallel()

	lib := &library.Library{}
	var rl removerLibrary = lib
	assert.NotNil(t, rl, "*library.Library must satisfy removerLibrary")
}

// T18 — Constructor returns a parent with both `resource` and
// `preset` sub-commands (per the sub-command dispatch structure
// required by the task spec).
func TestNewCmdRemove_HasSubcommands(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdRemove(f, nil, func(*removeOptions) error { return nil })

	var sawResource, sawPreset bool
	for _, c := range cmd.Commands() {
		switch c.Name() {
		case "resource":
			sawResource = true
		case "preset":
			sawPreset = true
		}
	}
	assert.True(t, sawResource,
		"library remove parent must register 'resource' as a child command")
	assert.True(t, sawPreset,
		"library remove parent must register 'preset' as a child command")
}

// T19 — Parent command has no Run/RunE (the dispatch happens via
// sub-commands). Cobra's default help-rendering kicks in for a
// bare `library remove` invocation.
func TestNewCmdRemove_ParentHasNoRun(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	cmd := NewCmdRemove(f, nil, func(*removeOptions) error { return nil })
	assert.Nil(t, cmd.RunE,
		"parent library remove must not have a RunE (sub-commands handle dispatch)")
}

// T20 — `--force` and `--output` are inherited by both sub-commands
// (PersistentFlags on the parent). Invoking
// `library remove resource <ref> --output json` works (the
// `library remove resource skill/commit --output json` scenario
// in the spec).
func TestNewCmdRemove_OutputFlagOnSubcommand(t *testing.T) {
	var captured *removeOptions
	runF := func(opts *removeOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	ios, _, _ := newRemoveTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	// Put --output AFTER the sub-command to exercise PersistentFlags
	// inheritance (the spec's "JSON output" scenario).
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdRemove(f, nil, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "resource", "skill/commit", "--output", "json"))
	require.NotNil(t, captured)
	assert.Equal(t, "json", captured.Output,
		"--output after the sub-command must be inherited from the parent")
}

// T21 — Invalid resource ref (no slash) returns a
// *core.ConfigError (via library.ParseRef). cmdutil.ExitCodeFor
// maps ConfigError to ExitCodeError (1) via the typed-error
// branch.
func TestRunRemove_Resource_InvalidRef(t *testing.T) {
	libDir := removeResourceFixture(t)

	ios, _, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:  ios,
		Ctx: context.Background(),
		Ref: "no-slash",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runRemove(opts)
	require.Error(t, err)
	var cfgErr *core.ConfigError
	require.True(t, errors.As(err, &cfgErr),
		"expected *core.ConfigError from library.ParseRef, got %T: %v", err, err)
}

// T22 — Invalid resource type (e.g., "fake/commit") returns a
// *core.NotFoundError (cmd layer can't find a registered type).
func TestRunRemove_Resource_InvalidType(t *testing.T) {
	libDir := removeResourceFixture(t)

	ios, _, _ := newRemoveTestIO()
	opts := &removeOptions{
		IO:  ios,
		Ctx: context.Background(),
		Ref: "fake/commit",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runRemove(opts)
	require.Error(t, err)
	var notFound *core.NotFoundError
	require.True(t, errors.As(err, &notFound),
		"expected *core.NotFoundError, got %T: %v", err, err)
}

// T23b — inline closure honors cfg.Library when f.Config is wired.
// Pins the per-RunE path resolution pattern. Sequential (NOT
// t.Parallel) because t.Setenv is incompatible with parallel
// subtests per golang-testing Rule 4.
func TestRemoveLibrary_HonorsConfigLibrary(t *testing.T) {
	cfg := &config.Config{Library: "/from/cfg/path"}

	f := &cmdutil.Factory{
		RootContext:     context.Background(),
		CompletionCache: cmdutil.NewCompletionCache(),
	}
	f.Config = func() (*config.Config, error) { return cfg, nil }
	t.Setenv("GERMINATOR_LIBRARY", "")

	// Mirror the RunE inline closure pattern (Phase 1).
	var cfgPath string
	if f.Config != nil {
		if loaded, cfgErr := f.Config(); cfgErr == nil && loaded != nil {
			cfgPath = loaded.Library
		}
	}
	resolved := library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"), cfgPath)
	loader := cmdutil.OnceValuesFunc(func() (*library.Library, error) {
		return library.LoadLibrary(context.Background(), resolved)
	})

	_, err := loader()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "/from/cfg/path",
		"resolved library path MUST reflect cfg.Library when flag and env are unset")
}

// T23c — inline closure survives f.Config == nil without panicking.
// Sequential (NOT t.Parallel) because t.Setenv is incompatible with
// parallel subtests per golang-testing Rule 4.
func TestRemoveLibrary_FConfigIsNilFallsBack(t *testing.T) {
	f := &cmdutil.Factory{
		RootContext:     context.Background(),
		CompletionCache: cmdutil.NewCompletionCache(),
		// f.Config intentionally left nil.
	}
	t.Setenv("GERMINATOR_LIBRARY", "/from/env/path")

	// Mirror the RunE inline closure pattern (Phase 1).
	var cfgPath string
	if f.Config != nil {
		if loaded, cfgErr := f.Config(); cfgErr == nil && loaded != nil {
			cfgPath = loaded.Library
		}
	}
	resolved := library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"), cfgPath)
	loader := cmdutil.OnceValuesFunc(func() (*library.Library, error) {
		return library.LoadLibrary(context.Background(), resolved)
	})

	_, err := loader()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "/from/env/path",
		"resolved path MUST fall through to env when f.Config is nil")
}

// T24 — Spec scenario "Invalidate after runRemove": a successful
// resource removal MUST invalidate the completion cache so the next
// shell completion does not surface the deleted resource. Mirrors
// TestRunAdd_InvalidatesCompletionCache (cmd/library_add_test.go).
func TestRunRemove_Resource_InvalidatesCompletionCache(t *testing.T) {
	t.Parallel()

	libDir := removeResourceFixture(t)

	ios, _, _ := newRemoveTestIO()
	cache := cmdutil.NewCompletionCache()

	// Pre-populate the cache with a stale library snapshot so we can
	// observe it being cleared by the post-mutation Invalidate call.
	staleLib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	cache.Set(libDir, staleLib, 5*time.Second)
	require.NotNil(t, cache.Get(libDir), "precondition: cache must hold the stale entry")

	opts := &removeOptions{
		IO:              ios,
		Ctx:             context.Background(),
		Ref:             "skill/commit",
		CompletionCache: cache,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	assert.Nil(t, cache.Get(libDir),
		"cache entry MUST be cleared after a successful mutation")
}

// T25 — Preset removal also invalidates the completion cache so the
// next completion does not surface the removed preset. Mirrors
// TestRunAdd_InvalidatesCompletionCache.
func TestRunRemove_Preset_InvalidatesCompletionCache(t *testing.T) {
	t.Parallel()

	libDir := removePresetFixture(t, []string{"skill/commit"})

	ios, _, _ := newRemoveTestIO()
	cache := cmdutil.NewCompletionCache()

	staleLib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	cache.Set(libDir, staleLib, 5*time.Second)
	require.NotNil(t, cache.Get(libDir), "precondition: cache must hold the stale entry")

	opts := &removeOptions{
		IO:              ios,
		Ctx:             context.Background(),
		PresetName:      "wp",
		CompletionCache: cache,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runRemove(opts))
	assert.Nil(t, cache.Get(libDir),
		"cache entry MUST be cleared after a successful mutation")
}

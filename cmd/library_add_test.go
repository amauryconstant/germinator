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

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// newAddTestIO returns the buffer-backed IOStreams that tests use to
// assert on captured Out / ErrOut. Mirrors the slice-5 initTestIO
// helper. Panics if iostreams.Test() does not return *bytes.Buffer
// writers (it always does in this codebase, but guard anyway).
func newAddTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios := iostreams.Test()
	out, okOut := ios.Out.(*bytes.Buffer)
	errOut, okErr := ios.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return ios, out, errOut
}

// makeTestLibrary scaffolds a minimal library dir with one skill
// file already registered in library.yaml. Returns the resolved
// RootPath. Tests use this when they want a real library for
// AddResource to operate on.
func makeTestLibrary(t *testing.T, registered map[string]map[string]library.Resource) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
	lib := &library.Library{
		Version:   "1",
		RootPath:  dir,
		Resources: registered,
		Presets:   map[string]library.Preset{},
	}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("save library: %v", err)
	}
	return dir
}

// makeTestSkillFile writes a minimal skill .md file with the given
// frontmatter name and returns the path. Useful for explicit-mode
// tests that need a canonical source file.
func makeTestSkillFile(t *testing.T, dir, name, description string) string {
	t.Helper()
	path := filepath.Join(dir, "skill-"+name+".md")
	body := "---\nname: " + name + "\ndescription: " + description + "\ntools:\n  - bash\n---\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return path
}

// T1 — Constructor wires opts correctly via runF injection.
func TestNewCmdAdd_ValidatesArgs(t *testing.T) {
	var captured *addOptions
	runF := func(opts *addOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	ios, _, _ := newAddTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdAdd(f, &libPath, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	},
		"/tmp/skill-test.md", "--type", "skill", "--name", "test",
		"--platform", "opencode", "--description", "desc",
		"--force", "--dry-run", "--output", "json",
	))
	require.NotNil(t, captured)
	assert.Equal(t, []string{"/tmp/skill-test.md"}, captured.InputPaths)
	assert.Equal(t, "skill", captured.Type)
	assert.Equal(t, "test", captured.Name)
	assert.Equal(t, "opencode", captured.Platform)
	assert.Equal(t, "desc", captured.Description)
	assert.True(t, captured.Force)
	assert.True(t, captured.DryRun)
	assert.Equal(t, "json", captured.Output)
	assert.NotNil(t, captured.Library, "opts.Library must be wired by NewCmdAdd")
	assert.NotNil(t, captured.IO)
	assert.NotNil(t, captured.Ctx)
}

// T2 — Empty input without --discover fails Cobra Args validation;
// cmdutil.ExitCodeFor maps the "requires at least" message to exit 2.
func TestNewCmdAdd_RequiresDiscoverOrInput(t *testing.T) {
	ios, _, _ := newAddTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	err := executeCmd(t, func() any {
		cmd := NewCmdAdd(f, &libPath, func(*addOptions) error { return nil })
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	})
	require.Error(t, err)
	// Per enforce-error-discipline (Phase 1.2): the cobra substring-prefix
	// dispatch fallback was dropped. The cobra `accepts at most 1 arg(s)`
	// string is not a typed pflag error and is not yet wrapped in
	// *core.CobraUsageError (zero current call sites per task 3.14), so
	// it falls through to ExitCodeError (1).
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"empty input without --discover falls through to ExitCodeError (1) after Phase 1.2")
}

// T3 — --discover with no positional args is accepted by Cobra's Args closure.
func TestNewCmdAdd_AcceptsDiscoverFlag(t *testing.T) {
	var captured *addOptions
	runF := func(opts *addOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured = opts
		return nil
	}

	ios, _, _ := newAddTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdAdd(f, &libPath, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "--discover"))
	require.NotNil(t, captured)
	assert.True(t, captured.Discover, "--discover must populate opts.Discover")
}

// T4 — Explicit mode success: one skill file, valid prefix, success
// line on stdout and no stderr; runAdd returns nil (no partial-error
// since all succeeded).
func TestRunAdd_ExplicitMode_Success(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "newskill", "New skill")

	ios, out, errOut := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "skill",
		Name:       "newskill",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runAdd(opts))

	assert.Contains(t, out.String(), "Added resource: skill/newskill",
		"stdout must contain the byte-identical 'Added resource: X/Y' line")
	assert.Empty(t, errOut.String(), "no errors expected on full success")

	// Library now registers the new skill.
	lib, err := opts.Library()
	require.NoError(t, err)
	if _, exists := lib.Resources["skill"]["newskill"]; !exists {
		t.Errorf("library must register skill/newskill after add")
	}
}

// T5 — Invalid type/name combo returns *core.ValidationError from
// core.CanInstallResource, which output.FormatError renders to stderr.
func TestRunAdd_ExplicitMode_InvalidRef(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	src := filepath.Join(t.TempDir(), "x.md")

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "skills", // plural — invalid
		Name:       "commit",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runAdd(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr), "expected *core.ValidationError in chain")
	assert.Empty(t, out.String(), "no success lines on validation failure")
}

// T6 — Discover mode plain output: a fixture with one orphan and
// one registered skill triggers a successful registration in batch
// mode, producing an "Added:" line on stdout.
func TestRunAdd_DiscoverMode_PlainOutput(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{
		"skill": {
			"existing": {Path: "skills/existing.md", Description: "Existing"},
		},
	})
	// Drop an orphan into skills/ that isn't in library.yaml.
	srcDir := t.TempDir()
	orphan := makeTestSkillFile(t, srcDir, "orphan", "Orphan skill")
	if err := os.Rename(orphan, filepath.Join(libDir, "skills", "orphan.md")); err != nil {
		t.Fatalf("rename: %v", err)
	}

	ios, out, errOut := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "plain",
		Discover: true,
		Batch:    true,
		Force:    true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runAdd(opts))
	assert.Contains(t, out.String(), "Added resource: skill/orphan",
		"plain output must contain 'Added resource: skill/orphan' for the registered orphan")
	assert.Empty(t, errOut.String(),
		"plain output for successful registration must not write to stderr")
}

// T7 — Discover mode JSON output: the JSON payload matches the
// net-new discoverJSONPayload schema (added/conflicts/failed/summary).
func TestRunAdd_DiscoverMode_JSONOutput(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	orphan := makeTestSkillFile(t, srcDir, "jsonorphan", "Orphan skill")
	if err := os.Rename(orphan, filepath.Join(libDir, "skills", "jsonorphan.md")); err != nil {
		t.Fatalf("rename: %v", err)
	}

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "json",
		Discover: true,
		Batch:    true,
		Force:    true,
		Library:  func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runAdd(opts))
	j := out.String()
	// Net-new JSON shape — must mention the orphaned key.
	assert.Contains(t, j, "\"added\"", "JSON payload must contain an 'added' key")
	assert.Contains(t, j, "\"summary\"", "JSON payload must contain a 'summary' key")
}

// T8 — Discover mode table output: the table payload produces a
// tab-aligned row. We assert that output.Write() put SOMETHING on
// stdout and no errors.
func TestRunAdd_DiscoverMode_TableOutput(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	orphan := makeTestSkillFile(t, srcDir, "tableorphan", "Table orphan")
	if err := os.Rename(orphan, filepath.Join(libDir, "skills", "tableorphan.md")); err != nil {
		t.Fatalf("rename: %v", err)
	}

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "table",
		Discover: true,
		Batch:    true,
		Force:    true,
		Library:  func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runAdd(opts))
	assert.NotEmpty(t, out.String(), "table output must write at least one row to stdout")
}

// T9 — Batch mode + name conflict: orphan name collides with
// registered resource under a different type → *core.OperationError
// per file rendered to stderr + partial success returned.
func TestRunAdd_BatchMode_NameConflict(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{
		// Existing skill with name "clash". The orphan is an agent
		// with the same name — that triggers a cross-type conflict.
		"skill": {
			"clash": {Path: "skills/clash.md", Description: "Existing skill"},
		},
	})
	srcDir := t.TempDir()
	orphan := makeTestSkillFile(t, srcDir, "clash", "Conflicting agent")
	if err := os.Rename(orphan, filepath.Join(libDir, "agents", "clash.md")); err != nil {
		t.Fatalf("rename: %v", err)
	}

	ios, _, errOut := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "plain",
		Discover: true,
		Batch:    true,
		Force:    false, // Force is required to bypass, so conflict must surface
		Library:  func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	err := runAdd(opts)
	require.Error(t, err, "name conflict must surface as a partial-success error")

	var ps *core.PartialSuccessError
	require.True(t, errors.As(err, &ps), "expected *core.PartialSuccessError")

	// runAdd must NOT render to ErrOut (single-handling rule per
	// cmd/AGENTS.md) — main.go renders the returned error once via
	// output.FormatError.
	assert.Empty(t, errOut.String(),
		"runAdd must NOT render per-file errors (single-handling rule)")

	output.FormatError(ios, err)
	// Central handler renders *core.PartialSuccessError as
	// `partial success: N succeeded, M failed\n  - initialize failed: <ref>: <opErr>\n`.
	// The OperationError.Error() returns "<op>: <resource>" (the Op tag
	// "register" is what flows through the chain here; the underlying
	// ErrNameConflict is reachable via errors.Is for programmatic checks).
	assert.Contains(t, errOut.String(), "register: agent/clash",
		"per-file error must surface the OperationError.Op+Resource through the typed-error chain")
}

// T10 — Cancellation test: cancel opts.Ctx mid-batch, verify that
// the function returns a wrapped ctx.Err() and partial successes are
// accumulated.
func TestRunAdd_BatchMode_Cancellation(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	// Create multiple orphan skills so cancellation has something to interrupt.
	srcDir := t.TempDir()
	for _, name := range []string{"cancel1", "cancel2", "cancel3"} {
		p := makeTestSkillFile(t, srcDir, name, "Cancel test")
		if err := os.Rename(p, filepath.Join(libDir, "skills", name+".md")); err != nil {
			t.Fatalf("rename: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately — first orphan processing may finish, but
	// cancellation surfaces by the second iteration.
	cancel()

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      ctx,
		Output:   "plain",
		Discover: true,
		Batch:    true,
		Force:    true,
		Library:  func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	err := runAdd(opts)
	require.Error(t, err, "cancelled batch must return non-nil error")

	// err must wrap context.Canceled (or DeadlineExceeded).
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error must wrap ctx.Err(); got %v", err)
	}
	_ = out // output is best-effort here
}

// T11 — Library field shape: addOptions must declare the fields
// named in the task spec; reflection check guards against accidental
// drops or renames.
func TestAddOptions_StructShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(addOptions{})
	want := map[string]bool{
		"IO":              true,
		"Library":         true,
		"Adder":           true,
		"Ctx":             true,
		"InputPaths":      true,
		"Name":            true,
		"Description":     true,
		"Type":            true,
		"Platform":        true,
		"Discover":        true,
		"Batch":           true,
		"Force":           true,
		"DryRun":          true,
		"Output":          true,
		"CompletionCache": true,
	}
	got := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[typ.Field(i).Name] = true
	}
	assert.Equal(t, want, got,
		"addOptions must declare exactly the spec-named fields")
}

// T18b — Cmd layer populates Stdout on AddRequest. Verifies that
// runAdd fills AddRequest.Stdout with opts.IO.Out (the cmd-side
// writer discipline) instead of leaving it nil or, worse, writing
// dry-run output to os.Stdout directly. Closes the
// library-library-batch-add spec scenario "Cmd layer populates the
// writer field" by injecting a thin test-local stub (testAdder) via
// opts.Adder and capturing the *library.AddRequest passed to Add.
//
// Also covers library-library-resource-import by exercising the
// same Stdout field on the AddRequest path (not just the
// BatchAddOptions path covered below).
//
// Per internal/AGENTS.md "When to Mock vs Use Real Implementations":
// production paths use real *library.Library; tests that need to
// assert wire-shape without mutating library.yaml use a thin
// test-local stub that records the call and short-circuits the
// real body.
func TestRunAdd_PopulatesStdoutOnAddRequest(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "newskill", "New skill")

	ios, _, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "skill",
		Name:       "newskill",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	var captured *library.AddRequest
	opts.Adder = testAdder{
		addFn: func(_ context.Context, req *library.AddRequest) error {
			captured = req
			return nil
		},
	}

	require.NoError(t, runAdd(opts))

	if captured == nil {
		t.Fatal("opts.Adder.Add was not called; runAdd did not reach runAddExplicit")
	}
	if captured.Stdout == nil {
		t.Fatal("captured AddRequest.Stdout is nil; cmd layer must populate it from opts.IO.Out")
	}
	if captured.Stdout != ios.Out {
		t.Errorf("captured AddRequest.Stdout = %p, want opts.IO.Out = %p "+
			"(cmd/library_add.go AddRequest literal must include Stdout: opts.IO.Out)",
			captured.Stdout, ios.Out)
	}
}

// T18c — Cmd layer populates Stdout on BatchAddOptions (file
// batch path). Verifies the runAddBatchFiles site
// (cmd/library_add.go:534) populates Stdout: opts.IO.Out.
func TestRunAddBatchFiles_PopulatesStdoutOnBatchAddOptions(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "batchskill", "Batch skill")

	ios, _, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Batch:      true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	var captured library.BatchAddOptions
	opts.Adder = testAdder{
		batchFn: func(_ context.Context, bo *library.BatchAddOptions) (*library.BatchAddResult, error) {
			captured = *bo
			return &library.BatchAddResult{Summary: library.BatchSummary{Added: 1}}, nil
		},
	}

	require.NoError(t, runAdd(opts))

	if captured.Stdout == nil {
		t.Fatal("captured BatchAddOptions.Stdout is nil; cmd layer must populate it from opts.IO.Out")
	}
	if captured.Stdout != ios.Out {
		t.Errorf("captured BatchAddOptions.Stdout = %p, want opts.IO.Out = %p "+
			"(cmd/library_add.go BatchAddOptions literal must include Stdout: opts.IO.Out)",
			captured.Stdout, ios.Out)
	}
}

// testAdder is a thin test-local stub satisfying the adderLibrary
// interface. Each method delegates to a function field (if set) so
// tests can capture inputs and short-circuit the real (*library.Library)
// body (which would mutate library.yaml on disk). Per
// internal/AGENTS.md, production paths use real *library.Library
// instances; this stub is reserved for tests that need to assert
// wire-shape (e.g., writer discipline) without on-disk side effects.
type testAdder struct {
	addFn      func(ctx context.Context, req *library.AddRequest) error
	batchFn    func(ctx context.Context, opts *library.BatchAddOptions) (*library.BatchAddResult, error)
	discoverFn func(ctx context.Context, opts *library.DiscoverOptions) (*library.DiscoverResult, error)
}

func (t testAdder) Add(ctx context.Context, req *library.AddRequest) error {
	if t.addFn == nil {
		return nil
	}
	return t.addFn(ctx, req)
}

func (t testAdder) BatchAddResources(ctx context.Context, opts *library.BatchAddOptions) (*library.BatchAddResult, error) {
	if t.batchFn == nil {
		return &library.BatchAddResult{}, nil
	}
	return t.batchFn(ctx, opts)
}

func (t testAdder) DiscoverOrphans(ctx context.Context, opts *library.DiscoverOptions) (*library.DiscoverResult, error) {
	if t.discoverFn == nil {
		return &library.DiscoverResult{}, nil
	}
	return t.discoverFn(ctx, opts)
}

// T12 — Compile-time assertion `var _ adderLibrary = (*library.Library)(nil)`
// runs at package init and is the same check exercised at runtime via
// this test. The check is now satisfied by *library.Library directly
// (the slice-8 method-form additions on *Library), so no adapter
// shim or interface-local type is required.
func TestAdderLibraryInterfaceSatisfied(t *testing.T) {
	t.Parallel()

	var al adderLibrary = &library.Library{}
	assert.NotNil(t, al, "*library.Library must satisfy adderLibrary")
}

// T14 — Golden file: explicit-mode plain output for a single
// successful skill add. Regenerated alongside the implementation
// (per design Decision 9: "Plain-output byte-identical guarantee is
// relaxed in slice 6 because the implementation is also new"); the
// baseline is captured from this test's own output rather than from
// the pre-change library_add.go.
//
// Determinism: this test rewrites the captured output to a stable
// path before comparison so the random t.TempDir() suffix does not
// invalidate the baseline on each run. Tests run with the unmodified
// file; set GOLDEN_UPDATE=1 to regenerate.
func TestRunAdd_PlainOutput_Golden(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "goldenskill", "Golden")

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "skill",
		Name:       "goldenskill",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runAdd(opts))

	// Normalize the captured output: replace the volatile t.TempDir()
	// prefix with a stable sentinel so the golden file remains
	// comparable across runs.
	got := normalizeGolden(out.String())
	goldenPath := filepath.Join("testdata", "library_add_plain.golden")
	if os.Getenv("GOLDEN_UPDATE") != "" {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
		t.Logf("updated golden file at %s", goldenPath)
		return
	}
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file missing; set GOLDEN_UPDATE=1 to generate")
	assert.Equal(t, string(want), got, "plain output must match the recorded golden baseline")
}

// normalizeGolden rewrites a captured plain-output string so volatile
// runtime values (e.g., t.TempDir() paths) become stable sentinels.
// This keeps golden files comparable across runs without forcing the
// implementation to use synthetic paths.
//
// Strategy: locate the first temp-prefix (e.g., "/tmp/...") after a
// known output marker ("Added: "), then replace the prefix through
// the next slash with "<TMPDIR>". Leaves the relative-path suffix
// intact for the golden comparison. There is at most one Added line
// per output, so a single pass is sufficient.
func normalizeGolden(s string) string {
	const marker = "Added: "
	i := findIndex(s, marker)
	if i < 0 {
		return s
	}
	pathStart := i + len(marker)
	replaced, handled := replaceFirstTempPrefix(s, pathStart)
	if !handled {
		return s
	}
	return replaced
}

// replaceFirstTempPrefix rewrites the temp prefix in s starting at
// pathStart. Returns the rewritten string and handled=true when a
// /tmp/ or /var/folders/ segment was replaced. The replacement spans
// from the absolute root up to (and including) the per-test
// randomized subdir, leaving the trailing "/<num>/<file>" path
// intact for stability.
func replaceFirstTempPrefix(s string, pathStart int) (string, bool) {
	rest := s[pathStart:]
	for _, prefix := range []string{"/tmp/", "/var/folders/"} {
		j := findIndex(rest, prefix)
		if j < 0 {
			continue
		}
		// Find the next "/<digit>/..." boundary (the per-test index).
		// Empirically the structure is: <tempBase>/Test<name>XXXX/<num>/file
		// so we walk from j+len(prefix) to find the first "/<digit>/" segment.
		afterBase := j + len(prefix)
		// Skip past "<TestName>XXXX" — walk until we hit a '/' that is
		// followed by digits and then '/'.
		k := afterBase
		for k < len(rest) && rest[k] != '/' {
			k++
		}
		if k >= len(rest) {
			continue
		}
		// Confirm the segment after k starts with digits and a slash.
		seg := rest[k+1:]
		if allDigitsUntilSlash(seg) {
			// Replace <prefix>...<k> with "<TMPDIR>".
			return s[:pathStart+j] + "<TMPDIR>" + rest[k:], true
		}
	}
	return s, false
}

func allDigitsUntilSlash(s string) bool {
	i := 0
	for i < len(s) && s[i] != '/' {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		i++
	}
	return i > 0 && i < len(s) && s[i] == '/'
}

func findIndex(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// T15 — Batch-with-files (Mode 4, --batch without --discover) happy
// path: two valid skill source files are registered; plain output
// shows "Added resource:" per file plus the summary line.
func TestRunAdd_BatchFiles_Success(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src1 := makeTestSkillFile(t, srcDir, "batch1", "Batch one")
	src2 := makeTestSkillFile(t, srcDir, "batch2", "Batch two")

	ios, out, errOut := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src1, src2},
		Type:       "skill",
		Batch:      true,
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runAdd(opts))
	assert.Contains(t, out.String(), "Added resource: skill/batch1")
	assert.Contains(t, out.String(), "Added resource: skill/batch2")
	assert.Contains(t, out.String(), "Added 2,")
	assert.Empty(t, errOut.String())
}

// T16 — Batch-with-files with a failed entry: one valid source plus
// a missing source file produces a *core.PartialSuccessError with
// one success and one failure.
func TestRunAdd_BatchFiles_PartialFailure(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	good := makeTestSkillFile(t, srcDir, "good", "Good")
	bad := filepath.Join(t.TempDir(), "does-not-exist.md")

	ios, _, errOut := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{good, bad},
		Type:       "skill",
		Batch:      true,
		Force:      true,
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	err := runAdd(opts)
	require.Error(t, err)

	var ps *core.PartialSuccessError
	require.True(t, errors.As(err, &ps), "expected *core.PartialSuccessError")
	assert.Equal(t, 1, ps.Succeeded(), "one file succeeded")
	assert.Equal(t, 1, ps.Failed(), "one file failed")
	// runAdd must NOT render to ErrOut (single-handling rule) — main.go
	// renders the returned *core.PartialSuccessError once via FormatError.
	assert.Empty(t, errOut.String(),
		"runAdd must NOT render per-file errors (single-handling rule)")

	output.FormatError(ios, err)
	assert.Contains(t, errOut.String(), "partial success: 1 succeeded, 1 failed",
		"central handler must render the partial-success aggregate")
}

// T17 — Batch-with-files rejects an invalid --platform value up
// front before any I/O is performed.
func TestRunAdd_BatchFiles_InvalidPlatform(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	src := makeTestSkillFile(t, t.TempDir(), "x", "x")

	ios, _, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		InputPaths: []string{src},
		Platform:   "bogus-platform",
		Batch:      true,
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	err := runAdd(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "platform",
		"invalid --platform must surface a platform-related error")
}

// T18 — Batch-with-files in dry-run mode succeeds without writing
// any files and produces a successful batch output line.
func TestRunAdd_BatchFiles_DryRun(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	src := makeTestSkillFile(t, t.TempDir(), "dry", "Dry")

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "skill",
		Batch:      true,
		DryRun:     true,
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runAdd(opts))
	assert.Contains(t, out.String(), "Added",
		"dry-run batch produces the standard Added line(s)")
}

// T19 — Explicit-mode JSON output renders the discover-style payload
// via NewJSONExporter; the net-new shape must include "summary".
func TestRunAdd_ExplicitMode_JSONOutput(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	src := makeTestSkillFile(t, t.TempDir(), "json", "JSON")

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "json",
		InputPaths: []string{src},
		Type:       "skill",
		Name:       "json",
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runAdd(opts))
	j := out.String()
	assert.Contains(t, j, "\"summary\"", "JSON payload must carry a summary block")
	assert.Contains(t, j, "\"succeeded\"")
	assert.Contains(t, j, "\"total\"")
}

// T20 — Explicit-mode table output renders a single-row table via
// NewTableExporter; stdout must be non-empty.
func TestRunAdd_ExplicitMode_TableOutput(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	src := makeTestSkillFile(t, t.TempDir(), "tbl", "Table")

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "table",
		InputPaths: []string{src},
		Type:       "skill",
		Name:       "tbl",
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	require.NoError(t, runAdd(opts))
	assert.NotEmpty(t, out.String(), "table output writes a row to stdout")
	assert.Contains(t, out.String(), "explicit",
		"explicit-mode table row identifies itself as 'explicit'")
}

// T21 — Explicit-mode plain output with a partial-success aggregate
// (one valid, one missing source) returns *core.PartialSuccessError
// and writes the success line on stdout.
func TestRunAdd_ExplicitMode_PartialSuccess(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	good := makeTestSkillFile(t, t.TempDir(), "good", "Good")
	bad := filepath.Join(t.TempDir(), "missing.md")

	ios, out, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{good, bad},
		Name:       "good",
		Type:       "skill",
		Library:    func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
	}

	err := runAdd(opts)
	require.Error(t, err)

	var ps *core.PartialSuccessError
	require.True(t, errors.As(err, &ps), "expected *core.PartialSuccessError")
	assert.Equal(t, 1, ps.Succeeded())
	// explicit --name "good" with --type "skill" → resolved ref is "skill/good".
	assert.Contains(t, out.String(), "Added resource: skill/good")
}

// T22 — cmdLayerDetect is a pure function that maps a source path
// to a (docType, name) pair via the legacy filename prefix patterns.
func TestCmdLayerDetect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantType string
		wantName string
	}{
		{"prefix agent", "agent-reviewer.md", "agent", "reviewer"},
		{"prefix skill", "skill-commit.md", "skill", "commit"},
		{"prefix command", "command-build.md", "command", "build"},
		{"prefix memory", "memory-notes.md", "memory", "notes"},
		{"suffix agent", "reviewer-agent.md", "agent", "reviewer"},
		{"suffix skill", "commit-skill.md", "skill", "commit"},
		{"no match", "orphan.md", "", ""},
		{"no extension", "skill-commit", "skill", "commit"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			typ, name := cmdLayerDetect(tt.path)
			assert.Equal(t, tt.wantType, typ)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

// T23 — resolveAddedRef combines explicit opts.Type/Name with
// cmdLayerDetect to render the canonical "<type>/<name>" ref.
func TestResolveAddedRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts *addOptions
		path string
		want string
	}{
		{"both flags set", &addOptions{Type: "skill", Name: "x"}, "ignored.md", "skill/x"},
		{"only type set", &addOptions{Type: "skill"}, "anything.md", "skill/anything"},
		{"only name set, prefix detected", &addOptions{Name: "reviewer"}, "/dir/agent-reviewer.md", "agent/reviewer"},
		{"neither, prefix detected", &addOptions{}, "/dir/skill-commit.md", "skill/commit"},
		{"no pattern match", &addOptions{}, "/dir/orphan.md", "orphan.md"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, resolveAddedRef(tt.opts, tt.path))
		})
	}
}

// T24 — typeFromRef / nameFromRef split a "type/name" string.
func TestTypeAndNameFromRef(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "skill", typeFromRef("skill/commit"))
	assert.Equal(t, "commit", nameFromRef("skill/commit"))
	assert.Equal(t, "", typeFromRef("no-slash"))
	assert.Equal(t, "", nameFromRef("no-slash"))
	assert.Equal(t, "", typeFromRef(""))
	assert.Equal(t, "", nameFromRef(""))
}

// T25 — deriveLibraryPath returns lib.RootPath when opts.Library
// succeeds, and "" when the loader returns an error or is nil.
func TestDeriveLibraryPath(t *testing.T) {
	t.Parallel()

	t.Run("loader returns library", func(t *testing.T) {
		t.Parallel()
		libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
		opts := &addOptions{
			Library: func() (*library.Library, error) { return library.LoadLibrary(context.Background(), libDir) },
		}
		assert.Equal(t, libDir, deriveLibraryPath(opts))
	})

	t.Run("loader returns error", func(t *testing.T) {
		t.Parallel()
		opts := &addOptions{
			Library: func() (*library.Library, error) { return nil, errors.New("nope") },
		}
		assert.Equal(t, "", deriveLibraryPath(opts))
	})

	t.Run("loader is nil", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", deriveLibraryPath(&addOptions{}))
	})
}

// T23 — Spec scenario "Legacy --json flag is rejected": invoking
// `library add --json` returns a Cobra usage error and exit code 2
// (`ExitCodeUsage`). Cobra reports the unknown flag to stderr; the
// run function never executes.
func TestRunAdd_RejectsLegacyJSONFlag(t *testing.T) {
	t.Parallel()

	ios, _, errOut := newAddTestIO()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	libPath := ""
	err := executeCmd(t, func() any {
		cmd := NewCmdAdd(f, &libPath, nil)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(errOut)
		return cmd
	}, "./missing.md", "--type", "skill", "--name", "x", "--json")
	require.Error(t, err)
	assert.Equal(t, cmdutil.ExitCodeUsage, cmdutil.ExitCodeFor(err),
		"unknown --json flag must map to ExitCodeUsage (2)")
	assert.Contains(t, errOut.String(), "json",
		"stderr must mention the rejected --json flag")
}

// T24 — Spec scenario "Empty name fails pre-flight validation":
// `library add <file> --type skill --name ""` resolves the ref to
// `skill/`, which `core.CanInstallResource` rejects as empty-name
// before any I/O is performed. The returned error is a
// `*core.ValidationError` (exit 1 via default-error case).
func TestRunAdd_EmptyNameAbortsBeforeIO(t *testing.T) {
	t.Parallel()

	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "any", "Any")

	ios, _, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "skill",
		Name:       "",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runAdd(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr),
		"expected *core.ValidationError in chain")
	assert.Contains(t, verr.Message(), "ref name must be non-empty")
	assert.Equal(t, "skill/", verr.Value(),
		"resolved ref must be 'skill/' (empty-name slot)")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"empty-name ValidationError must map to ExitCodeError (1)")

	// No file is copied: skills/ contains the directory only.
	entries, readErr := os.ReadDir(filepath.Join(libDir, "skills"))
	require.NoError(t, readErr)
	for _, e := range entries {
		if !e.IsDir() {
			t.Errorf("skills/ must not contain new files after pre-flight rejection; got %q", e.Name())
		}
	}
}

// T25 — Spec scenario "Malformed ref (no slash) fails pre-flight
// validation": type="" with name="commit" yields a resolved ref of
// `commit` (no slash). `core.CanInstallResource("commit")` returns a
// ValidationError and the import aborts before I/O.
func TestRunAdd_NoSlashAbortsBeforeIO(t *testing.T) {
	t.Parallel()

	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "commit", "Commit")

	ios, _, _ := newAddTestIO()
	opts := &addOptions{
		IO:         ios,
		Ctx:        context.Background(),
		Output:     "plain",
		InputPaths: []string{src},
		Type:       "", // not detected: library falls back to filename
		Name:       "commit",
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runAdd(opts)
	require.Error(t, err)

	var verr *core.ValidationError
	require.True(t, errors.As(err, &verr),
		"expected *core.ValidationError in chain")
	assert.Equal(t, "commit", verr.Value(),
		"resolved ref must be 'commit' (no slash)")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"no-slash ValidationError must map to ExitCodeError (1)")
}

// T26 — Spec scenario "All conflicts returns exit 1": when every
// orphan collides, the partial-success aggregate has Succeeded==0
// and Failed==N. `cmdutil.ExitCodeFor(err)` returns
// `ExitCodeError` (1); the central handler (main.go) renders the
// returned *core.PartialSuccessError once via output.FormatError;
// stdout is empty (no data leakage on error paths).
func TestRunAdd_DiscoverMode_AllConflicts(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{
		"skill": {
			"clash1": {Path: "skills/clash1.md", Description: "Existing 1"},
			"clash2": {Path: "skills/clash2.md", Description: "Existing 2"},
		},
	})
	// Two orphans under agents/, each with a name that collides with
	// an existing skill — cross-type name conflict per orphan.
	for _, name := range []string{"clash1", "clash2"} {
		srcDir := t.TempDir()
		orphan := makeTestSkillFile(t, srcDir, name, name)
		if err := os.Rename(orphan, filepath.Join(libDir, "agents", name+".md")); err != nil {
			t.Fatalf("rename %s: %v", name, err)
		}
	}

	ios, out, errOut := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "plain",
		Discover: true,
		Batch:    true,
		Force:    true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runAdd(opts)
	require.Error(t, err, "all-conflicts batch must return non-nil error")

	var ps *core.PartialSuccessError
	require.True(t, errors.As(err, &ps), "expected *core.PartialSuccessError")
	assert.Equal(t, 0, ps.Succeeded(), "Succeeded must be 0")
	assert.Equal(t, 2, ps.Failed(), "Failed must be 2")
	assert.Equal(t, cmdutil.ExitCodeError, cmdutil.ExitCodeFor(err),
		"all-failures aggregate must map to ExitCodeError (1)")
	assert.Empty(t, out.String(), "stdout must be empty on all-failure paths")
	// runAdd must NOT render per-file errors directly to ErrOut
	// (single-handling rule) — main.go renders the returned error once.
	assert.Empty(t, errOut.String(),
		"runAdd must NOT render per-file errors (single-handling rule)")

	output.FormatError(ios, err)
	gotErr := errOut.String()
	assert.Contains(t, gotErr, "initialize failed: agent/clash1",
		"central handler must enumerate per-resource InitializeError for first conflict")
	assert.Contains(t, gotErr, "initialize failed: agent/clash2",
		"central handler must enumerate per-resource InitializeError for second conflict")
}

// T27 — Spec scenario "Name conflict counts as failure, not success":
// mixed discovery (one valid orphan + one conflicting orphan) yields
// Succeeded==1 / Failed==1, exit 0 (partial success), stdout carries
// the success line, stderr carries the per-file OperationError.
func TestRunAdd_DiscoverMode_MixedSuccessAndConflict(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{
		"skill": {
			"clash": {Path: "skills/clash.md", Description: "Existing skill"},
		},
	})
	// One valid orphan (no conflict) under skills/, plus one
	// cross-type orphan under agents/ that conflicts with skill/clash.
	srcDir := t.TempDir()
	valid := makeTestSkillFile(t, srcDir, "fresh", "Fresh orphan")
	if err := os.Rename(valid, filepath.Join(libDir, "skills", "fresh.md")); err != nil {
		t.Fatalf("rename fresh: %v", err)
	}
	srcConflict := makeTestSkillFile(t, t.TempDir(), "clash", "Conflicting orphan")
	if err := os.Rename(srcConflict, filepath.Join(libDir, "agents", "clash.md")); err != nil {
		t.Fatalf("rename clash: %v", err)
	}

	ios, out, errOut := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "plain",
		Discover: true,
		Batch:    true,
		Force:    true,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runAdd(opts)
	require.Error(t, err, "partial-success must surface as a non-nil error")

	var ps *core.PartialSuccessError
	require.True(t, errors.As(err, &ps), "expected *core.PartialSuccessError")
	assert.Equal(t, 1, ps.Succeeded(), "Succeeded must be 1")
	assert.Equal(t, 1, ps.Failed(), "Failed must be 1")
	assert.Equal(t, cmdutil.ExitCodeSuccess, cmdutil.ExitCodeFor(err),
		"partial-success with Succeeded>0 must map to ExitCodeSuccess (0)")
	assert.Contains(t, out.String(), "Added resource: skill/fresh",
		"stdout must carry the success line for the valid orphan")
	// runAdd must NOT render per-file errors directly to ErrOut
	// (single-handling rule) — main.go renders the returned
	// *core.PartialSuccessError once via output.FormatError.
	assert.Empty(t, errOut.String(),
		"runAdd must NOT render per-file errors (single-handling rule)")

	output.FormatError(ios, err)
	assert.Contains(t, errOut.String(), "initialize failed: agent/clash",
		"central handler must enumerate the per-resource InitializeError for the conflict")
}

// T28 — Spec scenario "OperationError preserves wrapped cause":
// `errors.Is(err, library.ErrNameConflict)` traverses through
// `*core.OperationError.Unwrap()` to the typed sentinel. The cmd-layer
// aggregator now wraps `library.ErrNameConflict` (sourced from
// `ConflictInfo.Cause`) so the chain is preserved end-to-end.
func TestRunAdd_NameConflictCauseIsErrNameConflict(t *testing.T) {
	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{
		"skill": {
			"clash": {Path: "skills/clash.md", Description: "Existing skill"},
		},
	})
	srcDir := t.TempDir()
	orphan := makeTestSkillFile(t, srcDir, "clash", "Conflicting agent")
	if err := os.Rename(orphan, filepath.Join(libDir, "agents", "clash.md")); err != nil {
		t.Fatalf("rename: %v", err)
	}

	ios, _, _ := newAddTestIO()
	opts := &addOptions{
		IO:       ios,
		Ctx:      context.Background(),
		Output:   "plain",
		Discover: true,
		Batch:    true,
		Force:    false,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	err := runAdd(opts)
	require.Error(t, err)

	var ps *core.PartialSuccessError
	require.True(t, errors.As(err, &ps), "expected *core.PartialSuccessError")

	saw := false
	for _, ie := range ps.Errors() {
		if errors.Is(&ie, library.ErrNameConflict) {
			saw = true
			break
		}
	}
	assert.True(t, saw,
		"errors.Is must traverse *core.InitializeError → *core.OperationError → library.ErrNameConflict")
}

// T29 — Spec scenario "Invalidate after runAdd": a successful runAdd
// that registers a resource MUST invalidate the completion cache so
// the next shell completion reflects the new state without waiting
// for the TTL. This test pre-populates the cache with a stale library
// (lacking the newly-added resource) and asserts the entry is cleared
// after runAdd returns. Verifying the full carapace invocation would
// require a real shell environment; the cache-cleared assertion is
// the authoritative signal here.
func TestRunAdd_InvalidatesCompletionCache(t *testing.T) {
	t.Parallel()

	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "cachedskill", "Cached skill")

	ios, _, _ := newAddTestIO()
	cache := cmdutil.NewCompletionCache()

	// Pre-populate the cache with a stale library snapshot so we can
	// observe it being cleared by the post-mutation Invalidate call.
	staleLib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	cache.Set(libDir, staleLib, 5*time.Second)
	require.NotNil(t, cache.Get(libDir), "precondition: cache must hold the stale entry")

	opts := &addOptions{
		IO:              ios,
		Ctx:             context.Background(),
		Output:          "plain",
		InputPaths:      []string{src},
		Type:            "skill",
		Name:            "cachedskill",
		CompletionCache: cache,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runAdd(opts))
	assert.Nil(t, cache.Get(libDir),
		"cache entry MUST be cleared after a successful mutation")
}

// T30 — Dry-run invocations MUST NOT invalidate the cache (no
// mutation occurred).
func TestRunAdd_DryRunDoesNotInvalidateCompletionCache(t *testing.T) {
	t.Parallel()

	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "dryrunskill", "Dry-run skill")

	ios, _, _ := newAddTestIO()
	cache := cmdutil.NewCompletionCache()

	staleLib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)
	cache.Set(libDir, staleLib, 5*time.Second)

	opts := &addOptions{
		IO:              ios,
		Ctx:             context.Background(),
		Output:          "plain",
		InputPaths:      []string{src},
		Type:            "skill",
		Name:            "dryrunskill",
		DryRun:          true,
		CompletionCache: cache,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}

	require.NoError(t, runAdd(opts))
	assert.NotNil(t, cache.Get(libDir),
		"cache entry MUST survive a dry-run (no mutation occurred)")
}

// T31 — Spec scenario "Fresh resource appears in completion": after a
// successful runAdd, the next completion lookup (actionResources) MUST
// surface the new resource without waiting for the cache TTL. This is
// the end-to-end form of T29 — instead of only asserting that the
// cache entry was cleared, we drive the carapace Action callback and
// inspect the rendered values.
func TestRunAdd_FreshResourceAppearsInCompletion(t *testing.T) {
	t.Parallel()

	libDir := makeTestLibrary(t, map[string]map[string]library.Resource{})
	srcDir := t.TempDir()
	src := makeTestSkillFile(t, srcDir, "freshskill", "Fresh skill")

	ios, _, _ := newAddTestIO()
	cache := cmdutil.NewCompletionCache()

	opts := &addOptions{
		IO:              ios,
		Ctx:             context.Background(),
		Output:          "plain",
		InputPaths:      []string{src},
		Type:            "skill",
		Name:            "freshskill",
		CompletionCache: cache,
		Library: func() (*library.Library, error) {
			return library.LoadLibrary(context.Background(), libDir)
		},
	}
	require.NoError(t, runAdd(opts))

	// Build a minimal cobra command so actionResources can resolve the
	// library path via resolveLibraryPath. The command does not need to
	// be executed; actionResources only consults cmd.Flags() and
	// cmd.Flag("library").Changed.
	cmd := &cobra.Command{Use: "show"}
	cmd.Flags().String("library", libDir, "library path")
	cmd.Flag("library").Changed = true

	f := &cmdutil.Factory{
		RootContext:     context.Background(),
		CompletionCache: cache,
	}

	values := addTestActionValues(t, actionResources(f, cmd).Invoke(carapace.Context{}))
	assert.Contains(t, values, "skill/freshskill",
		"actionResources MUST surface the freshly-added resource without waiting for the TTL")
}

// addTestActionValues mirrors actionValuesAsStrings in completions_test.go
// but lives here so library_add_test.go stays self-contained.
func addTestActionValues(t *testing.T, ia carapace.InvokedAction) []string {
	t.Helper()
	data, err := json.Marshal(ia)
	require.NoError(t, err, "json.Marshal of InvokedAction must succeed")
	var decoded struct {
		Values []struct {
			Value string `json:"value"`
		} `json:"values"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded),
		"json.Unmarshal of InvokedAction must succeed")
	out := make([]string, 0, len(decoded.Values))
	for _, v := range decoded.Values {
		out = append(out, v.Value)
	}
	return out
}

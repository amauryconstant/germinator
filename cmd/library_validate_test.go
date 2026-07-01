package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

// newLibraryValidateTestIO returns an iostreams.IOStreams backed by
// *bytes.Buffer instances, matching the convention used elsewhere
// in the migrated command tests (cmd/library_add_test.go,
// cmd/resources_test.go). Named with a libraryValidate prefix to
// avoid collision with cmd/validate_test.go's newValidateTestIO.
func newLibraryValidateTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	io := iostreams.Test()
	out, okOut := io.Out.(*bytes.Buffer)
	errOut, okErr := io.ErrOut.(*bytes.Buffer)
	if !okOut || !okErr {
		panic("iostreams.Test did not return *bytes.Buffer-backed streams")
	}
	return io, out, errOut
}

// writeLibraryFile scaffolds a temp library directory containing
// the supplied library.yaml body and an optional skills/<file>
// entry. Returns the temp directory; the caller is responsible for
// closing over it (t.TempDir handles cleanup via test cleanup).
func writeLibraryFile(t *testing.T, yamlBody string, skills map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(yamlBody), 0644); err != nil {
		t.Fatalf("write library.yaml: %v", err)
	}

	if len(skills) > 0 {
		skillsDir := filepath.Join(tmpDir, "skills")
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("mkdir skills: %v", err)
		}
		for name, body := range skills {
			if err := os.WriteFile(filepath.Join(skillsDir, name), []byte(body), 0644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
	}

	return tmpDir
}

// captureOptsViaRunF builds the validate command with runF that
// captures the resolved *libraryValidateOptions. Tests inspect the
// captured options to assert wiring without depending on
// runLibraryValidate.
func captureOptsViaRunF(t *testing.T, args []string) *libraryValidateOptions {
	t.Helper()
	io, _, _ := newLibraryValidateTestIO()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	captured := &libraryValidateOptions{}
	cmd := NewCmdLibraryValidate(f, func(opts *libraryValidateOptions) error {
		captured = opts
		return nil
	})
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute(), "runF must succeed")
	return captured
}

// loadTempLibrary loads the library at path, asserting success so
// the validate tests don't paper over loader regressions.
func loadTempLibrary(t *testing.T, path string) *library.Library {
	t.Helper()
	lib, err := library.LoadLibrary(context.Background(), path)
	require.NoError(t, err, "fixture library must load: %s", path)
	return lib
}

// TestNewCmdLibraryValidate_ConstructorInjectOpts verifies the
// NewCmdXxx(f, runF) shape: the constructor receives the Factory,
// runF is invoked with a fully-populated *libraryValidateOptions,
// and the resolved env path flows into opts.Library resolution.
//
// The constructor does NOT register its own --library flag — the
// parent's persistent --library is read inside RunE via
// c.Flags().Lookup("library") so children transparently inherit it
// at parse time. Therefore this test exercises the constructor
// without --library; the flag-inheritance path is covered by the
// TestLibraryCommand_Validate tests below.
//
// Not t.Parallel: t.Setenv cannot be combined with t.Parallel.
func TestNewCmdLibraryValidate_ConstructorInjectOpts(t *testing.T) {
	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	t.Setenv("GERMINATOR_LIBRARY", tmpDir)

	captured := captureOptsViaRunF(t, []string{})

	require.NotNil(t, captured)
	assert.NotNil(t, captured.IO, "opts.IO must come from the Factory")
	assert.Equal(t, "plain", captured.Output, "default --output value must be \"plain\"")
	assert.False(t, captured.Fix, "default --fix value must be false")
	require.NotNil(t, captured.Library, "opts.Library must be a non-nil lazy function")

	// opts.Library() must load the env-resolved library.
	lib, err := captured.Library()
	require.NoError(t, err)
	assert.Equal(t, tmpDir, lib.RootPath,
		"opts.Library() must load the env-resolved path")

	// Ctx must come from cobra's context (not nil).
	assert.NotNil(t, captured.Ctx, "opts.Ctx must be populated from c.Context()")
}

// TestNewCmdLibraryValidate_AcceptsNoLibraryFlag verifies the
// command parses clean when no --library flag is supplied; the
// resolution falls back to env (set per-test) or XDG default.
// Equivalent to opts.Library being non-nil (factory wires a loader).
//
// Not t.Parallel: t.Setenv cannot be combined with t.Parallel.
func TestNewCmdLibraryValidate_AcceptsNoLibraryFlag(t *testing.T) {
	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	t.Setenv("GERMINATOR_LIBRARY", tmpDir)

	captured := captureOptsViaRunF(t, []string{})

	require.NotNil(t, captured.Library, "opts.Library must be wired (factory/env fallback)")
	lib, err := captured.Library()
	require.NoError(t, err)
	assert.Equal(t, tmpDir, lib.RootPath,
		"opts.Library() must resolve to the env-set path")
}

// TestRunLibraryValidate_Clean verifies the happy path: a clean
// library yields "valid" with error/warning counts of 0 in plain
// output and exit 0.
func TestRunLibraryValidate_Clean(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "valid",
		"clean library must be reported valid")
	assert.Contains(t, output, "errors: 0",
		"clean library must report errors: 0")
}

// TestRunLibraryValidate_ErrorSeverityMissingFile verifies that
// validation issues at error severity (missing file) render in
// plain output and the error count surfaces.
func TestRunLibraryValidate_ErrorSeverityMissingFile(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	// runLibraryValidate returns nil even with errors — errors are
	// rendered as data; only an actual command failure returns
	// non-nil. cmdutil.ExitCodeFor on nil returns 0 (clean exit).
	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "missing",
		"missing-file issue must be rendered in plain output")
	assert.Contains(t, output, "errors: 1",
		"error count must surface in plain output")
}

// TestRunLibraryValidate_OrphanWarning verifies the warning path:
// an orphan file produces a warning in plain output but the library
// stays valid (warnings don't invalidate the library).
func TestRunLibraryValidate_OrphanWarning(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
		"orphan.md": "---\nname: orphan\n---\nOrphan",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "orphan",
		"orphan warning must be rendered in plain output")
	assert.Contains(t, output, "warnings: 1",
		"warning count must surface in plain output")
	assert.Contains(t, output, "valid",
		"library with warnings-only stays valid")
}

// TestRunLibraryValidate_FixRemovesMissingAndGhost verifies the
// --fix path: missing entries are removed and ghost preset refs are
// stripped; the plain output reports the fix was applied.
func TestRunLibraryValidate_FixRemovesMissingAndGhost(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		Fix:    true,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	assert.Contains(t, out.String(), "Fix applied",
		"plain output must indicate that --fix cleaned the library")

	// Verify the disk-level mutation: skill/missing must be gone
	// from library.yaml AND skill/ghost must be stripped from the
	// preset's resources.
	modified, err := os.ReadFile(filepath.Join(tmpDir, "library.yaml"))
	require.NoError(t, err, "library.yaml must be readable post-fix")
	body := string(modified)
	assert.NotContains(t, body, "skill/missing",
		"missing entry must be removed from library.yaml")
	// The preset must still exist (commit is valid), but ghost must
	// be stripped from the resources list. Use a substring check
	// against the bare ref string to avoid false positives.
	assert.NotContains(t, body, "skill/ghost",
		"ghost preset ref must be stripped")
}

// TestRunLibraryValidate_FixJSONPayloadIncludesFixSection verifies
// task 7.4.6: --fix --output json produces a payload whose `fix`
// field enumerates RemovedEntries and StrippedRefs. The non-fix
// case (validated below) must omit the field entirely.
func TestRunLibraryValidate_FixJSONPayloadIncludesFixSection(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "json",
		Fix:    true,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	var payload struct {
		Valid        bool `json:"valid"`
		ErrorCount   int  `json:"errorCount"`
		WarningCount int  `json:"warningCount"`
		Issues       []struct {
			Type     string `json:"type"`
			Severity string `json:"severity"`
			Ref      string `json:"ref"`
		} `json:"issues"`
		Fix *struct {
			RemovedEntries []string `json:"removedEntries"`
			StrippedRefs   []string `json:"strippedRefs"`
		} `json:"fix"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &payload),
		"--output json must produce valid JSON: %q", out.String())

	assert.NotNil(t, payload.Fix,
		"--fix --output json must include a non-nil fix section")
	if payload.Fix != nil {
		assert.Contains(t, payload.Fix.RemovedEntries, "skill/missing",
			"fix.RemovedEntries must include skill/missing")
		assert.Contains(t, payload.Fix.StrippedRefs, "skill/ghost",
			"fix.StrippedRefs must include skill/ghost")
	}
}

// TestRunLibraryValidate_JSONOmitsFixSectionWhenNoFix verifies the
// conditional omitempty behavior: without --fix, the JSON payload
// must not contain a `fix` field at all (so downstream parsers can
// distinguish fresh runs from fix-applied runs).
func TestRunLibraryValidate_JSONOmitsFixSectionWhenNoFix(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "json",
		// Fix is the zero value (false).
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(out.Bytes(), &raw),
		"--output json must produce valid JSON: %q", out.String())
	_, hasFix := raw["fix"]
	assert.False(t, hasFix,
		"non-fix --output json MUST omit the fix key (got: %s)", out.String())
}

// TestRunLibraryValidate_TableRenders verifies --output table
// produces a tab-aligned table with the canonical columns
// (severity, type, ref, message). The spec scenario
// "Table output" requires the four-column shape.
func TestRunLibraryValidate_TableRenders(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "table",
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "SEVERITY",
		"table output must include the SEVERITY column header")
	assert.Contains(t, output, "TYPE",
		"table output must include the TYPE column header")
	assert.Contains(t, output, "REF",
		"table output must include the REF column header")
	assert.Contains(t, output, "MESSAGE",
		"table output must include the MESSAGE column header")
	assert.Contains(t, output, "missing",
		"the missing-file issue must appear in the table rows")
}

// TestRunLibraryValidate_TableFixRenders verifies the --fix --output
// table path renders the (action, ref) shape per the spec scenario
// "--fix with --output table renders action/ref table". Each
// removed entry / stripped ref appears as a row.
func TestRunLibraryValidate_TableFixRenders(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "table",
		Fix:    true,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "ACTION",
		"--fix --output table must include the ACTION column header")
	assert.Contains(t, output, "REF",
		"--fix --output table must include the REF column header")
	assert.Contains(t, strings.ToLower(output), "removed",
		"action column must label removed entries")
	assert.Contains(t, strings.ToLower(output), "stripped",
		"action column must label stripped refs")
	assert.Contains(t, output, "skill/missing",
		"removed entry must appear as a row")
	assert.Contains(t, output, "skill/ghost",
		"stripped ref must appear as a row")
}

// TestRunLibraryValidate_LibraryLoadError verifies the error path:
// opts.Library returning an error surfaces via output.FormatError
// (stderr) AND is wrapped and returned (so cmdutil.ExitCodeFor
// returns 1). The stdout buffer stays empty (scriptability).
func TestRunLibraryValidate_LibraryLoadError(t *testing.T) {
	t.Parallel()

	io, out, errOut := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		Library: func() (*library.Library, error) {
			return nil, assert.AnError
		},
		Ctx: context.Background(),
	}

	err := runLibraryValidate(opts)
	require.Error(t, err, "load failure must propagate")
	assert.Contains(t, err.Error(), "loading library")
	assert.Contains(t, err.Error(), "assert.AnError",
		"original cause must be preserved in the chain")
	assert.NotEmpty(t, errOut.String(),
		"load failure must render to stderr via output.FormatError")
	assert.Empty(t, out.String(),
		"stdout must stay empty on the error path")
}

// TestNewCmdLibraryValidate_OldJSONFlagIsRejected verifies the
// legacy `--json` flag (slice 2 deleted it) is no longer
// recognized: a usage error returns from Execute so callers can map
// to exit code 2 via cmdutil.ExitCodeFor's cobraUsagePrefixes
// branch.
func TestNewCmdLibraryValidate_OldJSONFlagIsRejected(t *testing.T) {
	t.Parallel()

	io, _, _ := newLibraryValidateTestIO()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	cmd := NewCmdLibraryValidate(f, func(*libraryValidateOptions) error { return nil })
	cmd.SetArgs([]string{"--json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.Error(t, err, "legacy --json flag must be rejected")
	assert.True(t,
		strings.Contains(err.Error(), "unknown flag") ||
			strings.Contains(err.Error(), "json"),
		"error must indicate the rejected flag: %v", err)
}

// TestValidatorLibraryInterfaceSatisfied mirrors the runtime
// checks in library_refresh_test.go (TestRefresherLibraryInterfaceSatisfied)
// and library_remove_test.go (TestRemoverLibraryInterfaceSatisfied):
// confirms *library.Library satisfies the validatorLibrary contract
// declared at cmd/library_validate.go:39. The compile-time check
// `var _ validatorLibrary = (*library.Library)(nil)` is exercised
// here at runtime so an accidental interface drift shows up in
// `go test`, not only at build time.
func TestValidatorLibraryInterfaceSatisfied(t *testing.T) {
	t.Parallel()

	lib := &library.Library{}
	var vl validatorLibrary = lib
	assert.NotNil(t, vl, "*library.Library must satisfy validatorLibrary")
}

// TestRunLibraryValidate_FixNoOpOnClean verifies the spec scenario
// "--fix with no issues is a no-op" (library-library-validation):
// running validate with --fix on a clean library (a) prints
// "no fixes needed" in plain output and (b) leaves library.yaml
// byte-identical on disk. This defends the maintenance-feature
// contract: --fix is opt-in and never destructive when there is
// nothing to clean up.
func TestRunLibraryValidate_FixNoOpOnClean(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	pre, err := os.ReadFile(filepath.Join(tmpDir, "library.yaml"))
	require.NoError(t, err, "pre-state library.yaml must be readable")

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		Fix:    true,
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "no fixes needed",
		"--fix on clean library must indicate no fixes were applied")
	assert.NotContains(t, output, "Fix applied",
		"--fix on clean library must not print 'Fix applied' (no mutation occurred)")

	post, err := os.ReadFile(filepath.Join(tmpDir, "library.yaml"))
	require.NoError(t, err, "post-state library.yaml must be readable")
	assert.True(t, bytes.Equal(pre, post),
		"--fix on clean library must leave library.yaml byte-identical")
}

// TestRunLibraryValidate_ReadOnlyWithoutFix verifies the spec
// scenario "validate without --fix is read-only": running validate
// (without --fix) on a library with issues surfaces the issues for
// inspection but does NOT mutate library.yaml. This is the
// safety-critical constraint: `germinator library validate` is a
// read-only operation unless --fix is passed.
func TestRunLibraryValidate_ReadOnlyWithoutFix(t *testing.T) {
	t.Parallel()

	libYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets: {}
`
	tmpDir := writeLibraryFile(t, libYAML, map[string]string{
		"commit.md": "---\nname: commit\n---\nContent",
	})
	lib := loadTempLibrary(t, tmpDir)

	pre, err := os.ReadFile(filepath.Join(tmpDir, "library.yaml"))
	require.NoError(t, err, "pre-state library.yaml must be readable")

	io, out, _ := newLibraryValidateTestIO()
	opts := &libraryValidateOptions{
		IO:     io,
		Output: "plain",
		// Fix is the zero value (false) — the spec scenario under test.
		Library: func() (*library.Library, error) {
			return lib, nil
		},
		Ctx: context.Background(),
	}

	require.NoError(t, runLibraryValidate(opts))

	output := out.String()
	assert.Contains(t, output, "missing",
		"validate must surface the missing-file issue for inspection")
	assert.Contains(t, output, "Hint: Run with --fix to auto-clean",
		"validate without --fix must hint at the --fix flag")
	assert.NotContains(t, output, "Fix applied",
		"validate without --fix must not emit a 'Fix applied' line")

	post, err := os.ReadFile(filepath.Join(tmpDir, "library.yaml"))
	require.NoError(t, err, "post-state library.yaml must be readable")
	assert.True(t, bytes.Equal(pre, post),
		"validate without --fix must leave library.yaml byte-identical")
}

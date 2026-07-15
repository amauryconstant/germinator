package install

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// installFixtureLibrary scaffolds a real library directory with one
// registered skill resource whose filename follows the parser's
// DetectType convention. Returns the resolved RootPath.
func installFixtureLibrary(t *testing.T, skillName string) string {
	t.Helper()
	libDir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		require.NoError(t, os.MkdirAll(filepath.Join(libDir, sub), 0o750))
	}
	src := filepath.Join(libDir, "skills", "skill-"+skillName+".md")
	body := "---\nname: " + skillName + "\ndescription: " + skillName + " fixture\n---\nBody\n"
	require.NoError(t, os.WriteFile(src, []byte(body), 0o600))
	lib := &library.Library{
		Version:   "1",
		RootPath:  libDir,
		Resources: map[string]map[string]library.Resource{},
		Presets:   map[string]library.Preset{},
	}
	lib.Resources["skill"] = map[string]library.Resource{
		skillName: {Path: "skills/skill-" + skillName + ".md", Description: skillName + " fixture"},
	}
	require.NoError(t, library.SaveLibrary(lib))
	return libDir
}

func newInstallTestService() Service {
	return NewService(parser.NewParser(), renderer.NewSerializer())
}

func TestService_Initialize_HappyPath(t *testing.T) {
	libDir := installFixtureLibrary(t, "commit")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	svc := newInstallTestService()
	results, err := svc.Initialize(context.Background(), &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/commit"},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, "skill/commit", r.Ref)
	assert.Empty(t, r.Error, "happy-path ref must not record an error")

	// Output file exists at the resolved path.
	expectedPath, perr := library.GetOutputPath("skill", "commit", core.PlatformOpenCode, outDir)
	require.NoError(t, perr)
	_, statErr := os.Stat(expectedPath)
	assert.NoError(t, statErr, "install must write the output file at the resolved path")
}

func TestService_Initialize_DryRunSkipsWrites(t *testing.T) {
	libDir := installFixtureLibrary(t, "dryrunskill")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	svc := newInstallTestService()
	results, err := svc.Initialize(context.Background(), &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/dryrunskill"},
		DryRun:    true,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Empty(t, results[0].Error)

	// Output file must NOT exist under dry-run.
	expectedPath, perr := library.GetOutputPath("skill", "dryrunskill", core.PlatformOpenCode, outDir)
	require.NoError(t, perr)
	_, statErr := os.Stat(expectedPath)
	assert.True(t, os.IsNotExist(statErr), "dry-run must NOT create the output file")
}

func TestService_Initialize_NoForceRejectsExisting(t *testing.T) {
	libDir := installFixtureLibrary(t, "existing")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	// Pre-create the output file to force the existing-file branch.
	expectedPath, perr := library.GetOutputPath("skill", "existing", core.PlatformOpenCode, outDir)
	require.NoError(t, perr)
	require.NoError(t, os.MkdirAll(filepath.Dir(expectedPath), 0o750))
	require.NoError(t, os.WriteFile(expectedPath, []byte("blocker"), 0o600))

	svc := newInstallTestService()
	results, err := svc.Initialize(context.Background(), &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/existing"},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	var fe *core.FileError
	require.ErrorAs(t, results[0].Error, &fe,
		"existing file must surface as *core.FileError in result.Error")
	assert.Equal(t, "write", fe.Operation())
}

func TestService_Initialize_ForceOverwrites(t *testing.T) {
	libDir := installFixtureLibrary(t, "forcey")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	expectedPath, perr := library.GetOutputPath("skill", "forcey", core.PlatformOpenCode, outDir)
	require.NoError(t, perr)
	require.NoError(t, os.MkdirAll(filepath.Dir(expectedPath), 0o750))
	require.NoError(t, os.WriteFile(expectedPath, []byte("blocker"), 0o600))

	svc := newInstallTestService()
	results, err := svc.Initialize(context.Background(), &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/forcey"},
		Force:     true,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Empty(t, results[0].Error, "force=true must overwrite without error")

	// File is replaced (not necessarily with the original "blocker" content).
	content, rerr := os.ReadFile(expectedPath)
	require.NoError(t, rerr)
	assert.NotEqual(t, "blocker", string(content), "force=true must overwrite the file")
}

func TestService_Initialize_MissingRef(t *testing.T) {
	libDir := installFixtureLibrary(t, "lonely")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	svc := newInstallTestService()
	results, err := svc.Initialize(context.Background(), &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/ghost"},
	})
	require.NoError(t, err, "missing-ref failure lives in result.Error, not the error return")
	require.Len(t, results, 1)
	require.Error(t, results[0].Error)
}

func TestService_Initialize_PartialSuccess(t *testing.T) {
	libDir := installFixtureLibrary(t, "present")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	outDir := t.TempDir()
	svc := newInstallTestService()
	results, err := svc.Initialize(context.Background(), &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/present", "skill/ghost"},
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Empty(t, results[0].Error, "first ref must succeed")
	assert.Error(t, results[1].Error, "second ref must fail (missing)")
}

func TestService_Initialize_CtxCancelled(t *testing.T) {
	libDir := installFixtureLibrary(t, "cancel")
	lib, err := library.LoadLibrary(context.Background(), libDir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	outDir := t.TempDir()
	svc := newInstallTestService()
	results, err := svc.Initialize(ctx, &Request{
		Library:   lib,
		Platform:  core.PlatformOpenCode,
		OutputDir: outDir,
		Refs:      []string{"skill/cancel"},
	})
	// LoadDocument is ctx-aware: a cancelled context surfaces as a
	// per-resource failure in result.Error so callers can distinguish
	// it from transport-level errors. The error return stays nil
	// (consistent with the per-ref failure surface for missing refs).
	require.NoError(t, err, "transport-level error return stays nil; per-ref failure is in result.Error")
	require.Len(t, results, 1)
	require.Error(t, results[0].Error, "cancelled LoadDocument surfaces as result.Error")
}

func TestService_Initialize_NilRequest(t *testing.T) {
	svc := newInstallTestService()
	_, err := svc.Initialize(context.Background(), nil)
	require.Error(t, err)
	var ve *core.ValidationError
	assert.ErrorAs(t, err, &ve)
}

func TestService_Initialize_NilLibrary(t *testing.T) {
	svc := newInstallTestService()
	_, err := svc.Initialize(context.Background(), &Request{
		Platform:  core.PlatformOpenCode,
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/x"},
	})
	require.Error(t, err)
	var ve *core.ValidationError
	assert.ErrorAs(t, err, &ve)
}

func TestNewService_ImplementsService(t *testing.T) {
	t.Parallel()

	s := NewService(parser.NewParser(), renderer.NewSerializer())
	assert.NotNil(t, s, "NewService must return a non-nil Service")
}

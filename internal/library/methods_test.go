package library

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests cover the (*Library) X method forms (Refresh,
// RemoveResource, RemovePreset, Validate, Fix) introduced in
// slice 7.0 to let *library.Library satisfy the cmd-side interfaces
// declared in tasks 7.1-7.4 without an adapter shim.
//
// The pattern mirrors the slice-6 CreatePreset method coverage
// (creator.go:145) and follows the AGENTS.md "tests alongside code"
// convention: package-internal black-box tests, no mocks, table-driven
// where it makes sense.

func TestLibrary_Refresh(t *testing.T) {
	tests := []struct {
		name              string
		libraryYAML       string
		files             map[string]string
		dryRun            bool
		force             bool
		wantErr           bool
		wantUnchangedRefs []string
	}{
		{
			name: "success: matching files recorded as unchanged",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: same description
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\ndescription: same description\n---\n# Commit\n",
			},
			dryRun:            false,
			force:             false,
			wantErr:           false,
			wantUnchangedRefs: []string{"skill/commit"},
		},
		{
			name: "success: dry-run does not modify library",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: old description
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\ndescription: new description\n---\n# Commit\n",
			},
			dryRun:  true,
			force:   false,
			wantErr: false,
		},
		{
			name: "error: nil lib",
			libraryYAML: `version: "1"
resources: {}
`,
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name: "error: empty RootPath",
			libraryYAML: `version: "1"
resources: {}
`,
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.name == "error: nil lib" {
				_, err := ((*Library)(nil)).Refresh(context.Background(), &RefreshRequest{DryRun: tt.dryRun, Force: tt.force})
				require.Error(t, err)
				return
			}

			if tt.name == "error: empty RootPath" {
				_, err := (&Library{}).Refresh(context.Background(), &RefreshRequest{DryRun: tt.dryRun, Force: tt.force})
				require.Error(t, err)
				return
			}

			libraryPath := filepath.Join(tmpDir, "library.yaml")
			require.NoError(t, os.WriteFile(libraryPath, []byte(tt.libraryYAML), 0o600))
			for rel, content := range tt.files {
				fp := filepath.Join(tmpDir, rel)
				require.NoError(t, os.MkdirAll(filepath.Dir(fp), 0o750))
				require.NoError(t, os.WriteFile(fp, []byte(content), 0o600))
			}

			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			result, err := lib.Refresh(context.Background(), &RefreshRequest{DryRun: tt.dryRun, Force: tt.force})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			gotUnchanged := make([]string, 0, len(result.Unchanged))
			for _, u := range result.Unchanged {
				gotUnchanged = append(gotUnchanged, u.Ref)
			}
			if !stringSlicesEqualUnordered(gotUnchanged, tt.wantUnchangedRefs) {
				assert.Equal(t, tt.wantUnchangedRefs, gotUnchanged, "Unchanged refs = %v, want %v")
			}
		})
	}
}

func TestLibrary_Refresh_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	_, err := lib.Refresh(ctx, &RefreshRequest{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_RemoveResource(t *testing.T) {
	tests := []struct {
		name        string
		setupLib    func(t *testing.T, libDir string)
		req         *RemoveResourceRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "success: removes existing resource",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
				srcPath := filepath.Join(t.TempDir(), "skill.md")
				content := "---\nname: target\ndescription: target skill\ntype: skill\ntools:\n  - bash\n---\n# Target\n"
				require.NoError(t, os.WriteFile(srcPath, []byte(content), 0o600))
				require.NoError(t, AddResource(context.Background(), AddRequest{Source: srcPath, LibraryPath: libDir, Type: "skill"}))
			},
			req:     &RemoveResourceRequest{Ref: "skill/target"},
			wantErr: false,
		},
		{
			name: "error: invalid ref format",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			req:         &RemoveResourceRequest{Ref: "no-slash"},
			wantErr:     true,
			errContains: "invalid resource reference format",
		},
		{
			name: "error: resource not found",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			req:         &RemoveResourceRequest{Ref: "skill/ghost"},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:     "error: empty RootPath",
			setupLib: func(_ *testing.T, _ string) {},
			req:      &RemoveResourceRequest{Ref: "skill/anything"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			tt.setupLib(t, libDir)

			if tt.name == "error: empty RootPath" {
				err := (&Library{}).RemoveResource(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)

			err = lib.RemoveResource(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				assert.Contains(t, err.Error(), tt.errContains, "error must contain")
			}
		})
	}
}

func TestLibrary_RemoveResource_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	err := lib.RemoveResource(ctx, &RemoveResourceRequest{Ref: "skill/x"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_RemovePreset(t *testing.T) {
	tests := []struct {
		name        string
		setupLib    func(t *testing.T, libDir string)
		req         *RemovePresetRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "success: removes existing preset",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
				lib, err := LoadLibrary(context.Background(), libDir)
				require.NoError(t, err)
				require.NoError(t, AddPreset(lib, Preset{Name: "wp", Description: "d", Resources: []string{"skill/any"}}))
				require.NoError(t, SaveLibrary(lib))
			},
			req:     &RemovePresetRequest{Name: "wp"},
			wantErr: false,
		},
		{
			name: "error: empty name",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			req:         &RemovePresetRequest{Name: ""},
			wantErr:     true,
			errContains: "preset name is required",
		},
		{
			name: "error: preset not found",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			req:         &RemovePresetRequest{Name: "missing-preset"},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:     "error: empty RootPath",
			setupLib: func(_ *testing.T, _ string) {},
			req:      &RemovePresetRequest{Name: "x"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			tt.setupLib(t, libDir)

			if tt.name == "error: empty RootPath" {
				err := (&Library{}).RemovePreset(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)

			err = lib.RemovePreset(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				assert.Contains(t, err.Error(), tt.errContains, "error must contain")
			}
		})
	}
}

func TestLibrary_RemovePreset_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	err := lib.RemovePreset(ctx, &RemovePresetRequest{Name: "x"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_Validate(t *testing.T) {
	tests := []struct {
		name             string
		setupLib         func(t *testing.T, libDir string)
		req              *ValidateRequest
		wantErr          bool
		wantValid        bool
		wantFixApplied   bool
		wantFixHasFields bool
	}{
		{
			name: "success: clean library validates",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
				// Add a resource whose file exists
				srcPath := filepath.Join(t.TempDir(), "skill.md")
				content := "---\nname: ok\ndescription: ok\ntype: skill\ntools:\n  - bash\n---\n# Ok\n"
				require.NoError(t, os.WriteFile(srcPath, []byte(content), 0o600))
				require.NoError(t, AddResource(context.Background(), AddRequest{Source: srcPath, LibraryPath: libDir, Type: "skill"}))
			},
			req:            &ValidateRequest{Fix: false},
			wantErr:        false,
			wantValid:      true,
			wantFixApplied: false,
		},
		{
			name: "success: --fix triggers FixLibrary and populates FixResult",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
				// Add a resource that references a missing file
				lib, err := LoadLibrary(context.Background(), libDir)
				require.NoError(t, err)
				lib.Resources["skill"] = map[string]Resource{
					"missing": {Path: "skills/missing.md", Description: "missing skill"},
				}
				require.NoError(t, SaveLibrary(lib))
			},
			req:              &ValidateRequest{Fix: true},
			wantErr:          false,
			wantValid:        false,
			wantFixApplied:   true,
			wantFixHasFields: true,
		},
		{
			name:     "error: empty RootPath",
			setupLib: func(_ *testing.T, _ string) {},
			req:      &ValidateRequest{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			tt.setupLib(t, libDir)

			if tt.name == "error: empty RootPath" {
				_, err := (&Library{}).Validate(context.Background(), tt.req)
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)

			result, err := lib.Validate(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if result.Valid != tt.wantValid {
				assert.Equal(t, tt.wantValid, result.Valid, "Valid = %v, want %v")
			}
			if result.FixApplied != tt.wantFixApplied {
				assert.Equal(t, tt.wantFixApplied, result.FixApplied, "FixApplied = %v, want %v")
			}
			if tt.wantFixHasFields && result.FixResult == nil {
				assert.Fail(t, "expected FixResult to be populated, got nil")
			}
		})
	}
}

func TestLibrary_Validate_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	_, err := lib.Validate(ctx, &ValidateRequest{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_Fix(t *testing.T) {
	tests := []struct {
		name         string
		setupLib     func(t *testing.T, libDir string)
		wantErr      bool
		wantRemovals int
	}{
		{
			name: "success: removes missing entries and strips ghost refs",
			setupLib: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
				lib, err := LoadLibrary(context.Background(), libDir)
				require.NoError(t, err)
				lib.Resources["skill"] = map[string]Resource{
					"missing": {Path: "skills/missing.md", Description: "missing skill"},
				}
				lib.Presets["workflow"] = Preset{
					Name:      "workflow",
					Resources: []string{"skill/missing", "skill/ghost"},
				}
				require.NoError(t, SaveLibrary(lib))
			},
			wantErr:      false,
			wantRemovals: 1,
		},
		{
			name:     "error: empty RootPath",
			setupLib: func(_ *testing.T, _ string) {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			tt.setupLib(t, libDir)

			if tt.name == "error: empty RootPath" {
				_, err := (&Library{}).Fix(context.Background(), &FixRequest{})
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)

			result, err := lib.Fix(context.Background(), &FixRequest{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if len(result.RemovedEntries) != tt.wantRemovals {
				assert.Len(t, result.RemovedEntries, tt.wantRemovals, "RemovedEntries count count")
			}
		})
	}
}

func TestLibrary_Fix_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	_, err := lib.Fix(ctx, &FixRequest{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		req     *InitRequest
		wantErr bool
	}{
		{
			name: "success: dry-run on non-existing path",
			req: &InitRequest{
				Path:   filepath.Join(t.TempDir(), "new-lib"),
				DryRun: true,
			},
			wantErr: false,
		},
		{
			name:    "error: nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(context.Background(), tt.req, io.Discard)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestInit_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := Init(ctx, &InitRequest{Path: filepath.Join(t.TempDir(), "x")}, io.Discard)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_Add(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(t *testing.T, libDir string)
		req      *AddRequest
		wantErr  bool
		errMatch string
	}{
		{
			name: "success: writes file and updates lib.Resources",
			prepare: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			req:     &AddRequest{Name: "added", Type: "skill"},
			wantErr: false,
		},
		{
			name:    "error: nil request",
			prepare: func(_ *testing.T, _ string) {},
			req:     nil,
			wantErr: true,
		},
		{
			name:    "error: empty RootPath",
			prepare: func(_ *testing.T, _ string) {},
			req:     &AddRequest{Name: "x", Type: "skill"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			tt.prepare(t, libDir)

			// Build a real source file for the happy path.
			if tt.name == "success: writes file and updates lib.Resources" {
				srcDir := t.TempDir()
				src := filepath.Join(srcDir, "skill-added.md")
				body := "---\nname: added\ndescription: added skill\ntype: skill\ntools:\n  - bash\n---\n# Added\n"
				require.NoError(t, os.WriteFile(src, []byte(body), 0o600))
				tt.req.Source = src
				tt.req.LibraryPath = libDir
			}

			if tt.name == "error: empty RootPath" {
				err := (&Library{}).Add(context.Background(), tt.req)
				require.Error(t, err)
				return
			}
			if tt.name == "error: nil request" {
				lib := &Library{RootPath: t.TempDir()}
				err := lib.Add(context.Background(), nil)
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)
			addErr := lib.Add(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, addErr)
				return
			}
			require.NoError(t, addErr)
			if _, exists := lib.Resources["skill"]["added"]; !exists {
				assert.Fail(t, "lib.Resources must contain skill/added after Add")
			}
		})
	}
}

func TestLibrary_Add_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	err := lib.Add(ctx, &AddRequest{Name: "x", Type: "skill"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_BatchAddResources(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T, libDir string) []string
		opts      func(libDir string, sources []string) *BatchAddOptions
		wantErr   bool
		errMatch  string
		wantAdded int
	}{
		{
			name: "success: registers valid sources",
			prepare: func(t *testing.T, libDir string) []string {
				createTestLibrary(t, libDir)
				srcDir := t.TempDir()
				var sources []string
				for _, name := range []string{"ba1", "ba2"} {
					p := filepath.Join(srcDir, "skill-"+name+".md")
					body := "---\nname: " + name + "\ndescription: " + name + "\ntype: skill\ntools:\n  - bash\n---\n# " + name + "\n"
					require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
					sources = append(sources, p)
				}
				return sources
			},
			opts: func(libDir string, sources []string) *BatchAddOptions {
				return &BatchAddOptions{Sources: sources, LibraryPath: libDir}
			},
			wantAdded: 2,
		},
		{
			name:    "error: nil options",
			prepare: func(_ *testing.T, _ string) []string { return nil },
			opts:    func(_ string, _ []string) *BatchAddOptions { return nil },
			wantErr: true,
		},
		{
			name:    "error: empty RootPath",
			prepare: func(_ *testing.T, _ string) []string { return nil },
			opts:    func(_ string, _ []string) *BatchAddOptions { return &BatchAddOptions{} },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			sources := tt.prepare(t, libDir)
			opts := tt.opts(libDir, sources)

			if tt.name == "error: empty RootPath" {
				_, err := (&Library{}).BatchAddResources(context.Background(), opts)
				require.Error(t, err)
				return
			}
			if tt.name == "error: nil options" {
				lib := &Library{RootPath: t.TempDir()}
				_, err := lib.BatchAddResources(context.Background(), nil)
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)
			result, batchErr := lib.BatchAddResources(context.Background(), opts)
			if tt.wantErr {
				require.Error(t, batchErr)
				return
			}
			require.NoError(t, batchErr)
			require.NotNil(t, result)
			if result.Summary.Added != tt.wantAdded {
				assert.Equal(t, tt.wantAdded, result.Summary.Added, "Summary.Added mismatch")
			}
		})
	}
}

func TestLibrary_BatchAddResources_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	_, err := lib.BatchAddResources(ctx, &BatchAddOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestLibrary_DiscoverOrphans(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(t *testing.T, libDir string)
		opts     *DiscoverOptions
		wantErr  bool
		wantOrph int
	}{
		{
			name: "success: empty library, no orphans",
			prepare: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			opts:     &DiscoverOptions{},
			wantOrph: 0,
		},
		{
			name:    "error: nil options",
			prepare: func(_ *testing.T, _ string) {},
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "error: empty RootPath",
			prepare: func(_ *testing.T, _ string) {},
			opts:    &DiscoverOptions{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libDir := t.TempDir()
			tt.prepare(t, libDir)

			if tt.name == "error: empty RootPath" {
				_, err := (&Library{}).DiscoverOrphans(context.Background(), tt.opts)
				require.Error(t, err)
				return
			}
			if tt.name == "error: nil options" {
				lib := &Library{RootPath: t.TempDir()}
				_, err := lib.DiscoverOrphans(context.Background(), nil)
				require.Error(t, err)
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			require.NoError(t, err)
			tt.opts.LibraryPath = libDir
			result, discErr := lib.DiscoverOrphans(context.Background(), tt.opts)
			if tt.wantErr {
				require.Error(t, discErr)
				return
			}
			require.NoError(t, discErr)
			if tt.wantErr {
				return
			}
			if len(result.Orphans) != tt.wantOrph {
				assert.Len(t, result.Orphans, tt.wantOrph, "len(Orphans) count")
			}
		})
	}
}

func TestLibrary_DiscoverOrphans_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	lib := &Library{RootPath: t.TempDir()}
	_, err := lib.DiscoverOrphans(ctx, &DiscoverOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

// stringSlicesEqualUnordered reports whether two string slices contain
// the same elements regardless of order. Used by the Unchanged test
// to avoid being order-sensitive (lib.Resources iteration order is
// map-dependent).
func stringSlicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
	}
	for _, c := range seen {
		if c != 0 {
			return false
		}
	}
	return true
}

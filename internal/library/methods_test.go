package library

import (
	"context"
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
				if err == nil {
					t.Fatal("expected error from nil lib, got nil")
				}
				return
			}

			if tt.name == "error: empty RootPath" {
				_, err := (&Library{}).Refresh(context.Background(), &RefreshRequest{DryRun: tt.dryRun, Force: tt.force})
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}

			libraryPath := filepath.Join(tmpDir, "library.yaml")
			if err := os.WriteFile(libraryPath, []byte(tt.libraryYAML), 0o600); err != nil {
				t.Fatalf("write library.yaml: %v", err)
			}
			for rel, content := range tt.files {
				fp := filepath.Join(tmpDir, rel)
				if err := os.MkdirAll(filepath.Dir(fp), 0o750); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(fp, []byte(content), 0o600); err != nil {
					t.Fatalf("write file: %v", err)
				}
			}

			lib, err := LoadLibrary(context.Background(), tmpDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}

			result, err := lib.Refresh(context.Background(), &RefreshRequest{DryRun: tt.dryRun, Force: tt.force})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Refresh() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			gotUnchanged := make([]string, 0, len(result.Unchanged))
			for _, u := range result.Unchanged {
				gotUnchanged = append(gotUnchanged, u.Ref)
			}
			if !stringSlicesEqualUnordered(gotUnchanged, tt.wantUnchangedRefs) {
				t.Errorf("Unchanged refs = %v, want %v", gotUnchanged, tt.wantUnchangedRefs)
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
	assert.ErrorIs(t, err, context.Canceled)
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
				if err := os.WriteFile(srcPath, []byte(content), 0o600); err != nil {
					t.Fatalf("write src: %v", err)
				}
				if err := AddResource(context.Background(), AddRequest{Source: srcPath, LibraryPath: libDir, Type: "skill"}); err != nil {
					t.Fatalf("AddResource: %v", err)
				}
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
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}

			err = lib.RemoveResource(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RemoveResource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
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
	assert.ErrorIs(t, err, context.Canceled)
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
				if err != nil {
					t.Fatalf("LoadLibrary: %v", err)
				}
				if err := AddPreset(lib, Preset{Name: "wp", Description: "d", Resources: []string{"skill/any"}}); err != nil {
					t.Fatalf("AddPreset: %v", err)
				}
				if err := SaveLibrary(lib); err != nil {
					t.Fatalf("SaveLibrary: %v", err)
				}
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
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}

			err = lib.RemovePreset(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RemovePreset() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
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
	assert.ErrorIs(t, err, context.Canceled)
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
				if err := os.WriteFile(srcPath, []byte(content), 0o600); err != nil {
					t.Fatalf("write src: %v", err)
				}
				if err := AddResource(context.Background(), AddRequest{Source: srcPath, LibraryPath: libDir, Type: "skill"}); err != nil {
					t.Fatalf("AddResource: %v", err)
				}
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
				if err != nil {
					t.Fatalf("LoadLibrary: %v", err)
				}
				lib.Resources["skill"] = map[string]Resource{
					"missing": {Path: "skills/missing.md", Description: "missing skill"},
				}
				if err := SaveLibrary(lib); err != nil {
					t.Fatalf("SaveLibrary: %v", err)
				}
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
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}

			result, err := lib.Validate(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.FixApplied != tt.wantFixApplied {
				t.Errorf("FixApplied = %v, want %v", result.FixApplied, tt.wantFixApplied)
			}
			if tt.wantFixHasFields && result.FixResult == nil {
				t.Error("expected FixResult to be populated, got nil")
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
	assert.ErrorIs(t, err, context.Canceled)
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
				if err != nil {
					t.Fatalf("LoadLibrary: %v", err)
				}
				lib.Resources["skill"] = map[string]Resource{
					"missing": {Path: "skills/missing.md", Description: "missing skill"},
				}
				lib.Presets["workflow"] = Preset{
					Name:      "workflow",
					Resources: []string{"skill/missing", "skill/ghost"},
				}
				if err := SaveLibrary(lib); err != nil {
					t.Fatalf("SaveLibrary: %v", err)
				}
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
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}

			result, err := lib.Fix(context.Background(), &FixRequest{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(result.RemovedEntries) != tt.wantRemovals {
				t.Errorf("RemovedEntries count = %d, want %d", len(result.RemovedEntries), tt.wantRemovals)
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
	assert.ErrorIs(t, err, context.Canceled)
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
			if (err != nil) != tt.wantErr {
				t.Fatalf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInit_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := Init(ctx, &InitRequest{Path: filepath.Join(t.TempDir(), "x")}, io.Discard)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
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
				if err := os.WriteFile(src, []byte(body), 0o600); err != nil {
					t.Fatalf("write src: %v", err)
				}
				tt.req.Source = src
				tt.req.LibraryPath = libDir
			}

			if tt.name == "error: empty RootPath" {
				err := (&Library{}).Add(context.Background(), tt.req)
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}
			if tt.name == "error: nil request" {
				lib := &Library{RootPath: t.TempDir()}
				err := lib.Add(context.Background(), nil)
				if err == nil {
					t.Fatal("expected error from nil request, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}
			if addErr := lib.Add(context.Background(), tt.req); (addErr != nil) != tt.wantErr {
				t.Fatalf("Add() error = %v, wantErr %v", addErr, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if _, exists := lib.Resources["skill"]["added"]; !exists {
				t.Error("lib.Resources must contain skill/added after Add")
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
	assert.ErrorIs(t, err, context.Canceled)
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
					if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
						t.Fatalf("write src: %v", err)
					}
					sources = append(sources, p)
				}
				return sources
			},
			opts:      func(libDir string, sources []string) *BatchAddOptions { return &BatchAddOptions{Sources: sources, LibraryPath: libDir} },
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
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}
			if tt.name == "error: nil options" {
				lib := &Library{RootPath: t.TempDir()}
				_, err := lib.BatchAddResources(context.Background(), nil)
				if err == nil {
					t.Fatal("expected error from nil options, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}
			result, batchErr := lib.BatchAddResources(context.Background(), opts)
			if (batchErr != nil) != tt.wantErr {
				t.Fatalf("BatchAddResources() error = %v, wantErr %v", batchErr, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Summary.Added != tt.wantAdded {
				t.Errorf("Summary.Added = %d, want %d", result.Summary.Added, tt.wantAdded)
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
	assert.ErrorIs(t, err, context.Canceled)
}

func TestLibrary_DiscoverOrphans(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T, libDir string)
		opts      *DiscoverOptions
		wantErr   bool
		wantOrph  int
	}{
		{
			name: "success: empty library, no orphans",
			prepare: func(t *testing.T, libDir string) {
				createTestLibrary(t, libDir)
			},
			opts:    &DiscoverOptions{},
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
				if err == nil {
					t.Fatal("expected error from empty RootPath, got nil")
				}
				return
			}
			if tt.name == "error: nil options" {
				lib := &Library{RootPath: t.TempDir()}
				_, err := lib.DiscoverOrphans(context.Background(), nil)
				if err == nil {
					t.Fatal("expected error from nil options, got nil")
				}
				return
			}

			lib, err := LoadLibrary(context.Background(), libDir)
			if err != nil {
				t.Fatalf("LoadLibrary: %v", err)
			}
			tt.opts.LibraryPath = libDir
			result, discErr := lib.DiscoverOrphans(context.Background(), tt.opts)
			if (discErr != nil) != tt.wantErr {
				t.Fatalf("DiscoverOrphans() error = %v, wantErr %v", discErr, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(result.Orphans) != tt.wantOrph {
				t.Errorf("len(Orphans) = %d, want %d", len(result.Orphans), tt.wantOrph)
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
	assert.ErrorIs(t, err, context.Canceled)
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

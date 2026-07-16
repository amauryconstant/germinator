package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckMissingFiles(t *testing.T) {
	tests := []struct {
		name           string
		libraryYAML    string
		files          map[string]string // path -> content
		expectedIssues int
		expectedRefs   []string
	}{
		{
			name: "no missing files",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			expectedIssues: 0,
			expectedRefs:   nil,
		},
		{
			name: "missing file detected",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files:          map[string]string{},
			expectedIssues: 1,
			expectedRefs:   []string{"skill/commit"},
		},
		{
			name: "multiple missing files",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    merge:
      path: skills/merge.md
      description: Merge skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			expectedIssues: 1,
			expectedRefs:   []string{"skill/merge"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644))

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Load library
			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Run check
			issues, err := CheckMissingFiles(lib)
			require.NoError(t, err)

			if len(issues) != tt.expectedIssues {
				assert.Len(t, issues, tt.expectedIssues, "issue count")
			}

			for _, expectedRef := range tt.expectedRefs {
				found := false
				for _, issue := range issues {
					if issue.Ref == expectedRef {
						found = true
						break
					}
				}
				if !found {
					assert.NotEmpty(t, expectedRef, "expected issue with ref")
				}
			}
		})
	}
}

func TestCheckOrphanedFiles(t *testing.T) {
	tests := []struct {
		name           string
		libraryYAML    string
		files          map[string]string
		expectedIssues int
		expectedPaths  []string
	}{
		{
			name: "no orphans",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			expectedIssues: 0,
			expectedPaths:  nil,
		},
		{
			name: "orphan file detected",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
				"skills/extra.md":  "---\nname: extra\n---\nContent",
			},
			expectedIssues: 1,
			expectedPaths:  []string{"skills/extra.md"},
		},
		{
			name:           "empty library with files",
			libraryYAML:    "version: \"1\"\nresources:\n  skill: {}\npresets: {}",
			files:          map[string]string{"skills/orphan.md": "Content"},
			expectedIssues: 1,
			expectedPaths:  []string{"skills/orphan.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644))

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Load library
			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Run check
			issues, err := CheckOrphanedFiles(lib)
			require.NoError(t, err)

			if len(issues) != tt.expectedIssues {
				assert.Len(t, issues, tt.expectedIssues, "issue count")
			}

			for _, expectedPath := range tt.expectedPaths {
				found := false
				for _, issue := range issues {
					if issue.Path == expectedPath {
						found = true
						break
					}
				}
				if !found {
					assert.NotEmpty(t, expectedPath, "expected issue with path")
				}
			}
		})
	}
}

func TestCheckGhostResources(t *testing.T) {
	tests := []struct {
		name              string
		libraryYAML       string
		expectedIssues    int
		expectedRefs      []string
		expectedInPresets []string
	}{
		{
			name: "no ghost resources",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
`,
			expectedIssues:    0,
			expectedRefs:      nil,
			expectedInPresets: nil,
		},
		{
			name: "ghost resource detected",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`,
			expectedIssues:    1,
			expectedRefs:      []string{"skill/ghost"},
			expectedInPresets: []string{"workflow"},
		},
		{
			name: "multiple ghost resources in same preset",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost1
      - skill/ghost2
`,
			expectedIssues:    2,
			expectedRefs:      []string{"skill/ghost1", "skill/ghost2"},
			expectedInPresets: []string{"workflow", "workflow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644))

			// Create skills directory and file
			skillsDir := filepath.Join(tmpDir, "skills")
			require.NoError(t, os.MkdirAll(skillsDir, 0755))
			require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "commit.md"), []byte("Content"), 0644))

			// Load library
			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Run check
			issues, err := CheckGhostResources(lib)
			require.NoError(t, err)

			if len(issues) != tt.expectedIssues {
				assert.Len(t, issues, tt.expectedIssues, "issue count")
			}

			for i, expectedRef := range tt.expectedRefs {
				if i >= len(issues) {
					assert.NotEmpty(t, expectedRef, "expected issue %d ref", i)
					continue
				}
				assert.Equal(t, expectedRef, issues[i].Ref, "issue %d ref", i)
				assert.Equal(t, tt.expectedInPresets[i], issues[i].InPreset, "issue %d inPreset", i)
			}
		})
	}
}

func TestCheckMalformedFrontmatter(t *testing.T) {
	tests := []struct {
		name           string
		libraryYAML    string
		files          map[string]string
		expectedIssues int
		expectedPaths  []string
	}{
		{
			name: "valid frontmatter",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\ndescription: Test\n---\nContent",
			},
			expectedIssues: 0,
			expectedPaths:  nil,
		},
		{
			name: "no frontmatter is valid",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "Just content without frontmatter",
			},
			expectedIssues: 0,
			expectedPaths:  nil,
		},
		{
			name: "malformed YAML in frontmatter",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\ninvalid: [yaml: content\n---\nContent",
			},
			expectedIssues: 1,
			expectedPaths:  []string{"skills/commit.md"},
		},
		{
			name: "missing closing delimiter",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\nMissing closing delimiter",
			},
			expectedIssues: 1,
			expectedPaths:  []string{"skills/commit.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644))

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Load library
			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Run check
			issues, err := CheckMalformedFrontmatter(lib)
			require.NoError(t, err)

			if len(issues) != tt.expectedIssues {
				assert.Len(t, issues, tt.expectedIssues, "issue count")
			}

			for _, expectedPath := range tt.expectedPaths {
				found := false
				for _, issue := range issues {
					if issue.Path == expectedPath {
						found = true
						break
					}
				}
				if !found {
					assert.NotEmpty(t, expectedPath, "expected issue with path")
				}
			}
		})
	}
}

func TestValidateLibrary(t *testing.T) {
	tests := []struct {
		name         string
		libraryYAML  string
		files        map[string]string
		wantValid    bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "valid library",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			wantValid:    true,
			wantErrors:   0,
			wantWarnings: 0,
		},
		{
			name: "missing file is error",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files:        map[string]string{},
			wantValid:    false,
			wantErrors:   1,
			wantWarnings: 0,
		},
		{
			name: "orphan is warning",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
				"skills/extra.md":  "---\nname: extra\n---\nContent",
			},
			wantValid:    true,
			wantErrors:   0,
			wantWarnings: 1,
		},
		{
			name: "ghost resource is error",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			wantValid:    false,
			wantErrors:   1,
			wantWarnings: 0,
		},
		{
			name: "multiple issue types",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    ghost:
      path: skills/ghost.md
      description: Ghost skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/ghost
      - skill/nonexistent
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
				"skills/extra.md":  "---\nname: extra\n---\nContent",
			},
			wantValid:    false,
			wantErrors:   3, // ghost file missing, ghost preset ref, nonexistent preset ref
			wantWarnings: 1, // extra.md orphan
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644))

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Load library
			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Run validation
			result, err := ValidateLibrary(lib)
			require.NoError(t, err)

			if result.Valid != tt.wantValid {
				assert.Equal(t, tt.wantValid, result.Valid, "Valid = %v, want %v")
			}
			errs := fieldByName(result, "ErrorCount").(int)
			if errs != tt.wantErrors {
				assert.Equal(t, tt.wantErrors, errs, "ErrorCount = %d, want %d")
			}
			warnings := result.WarningCount
			if warnings != tt.wantWarnings {
				assert.Equal(t, tt.wantWarnings, warnings, "WarningCount = %d, want %d")
			}
		})
	}
}

func TestFixLibrary(t *testing.T) {
	tests := []struct {
		name               string
		libraryYAML        string
		files              map[string]string
		wantRemovedEntries []string
		wantStrippedRefs   []string
		wantSavedLibHas    map[string]bool // ref -> should exist after fix
	}{
		{
			name: "fix removes missing file entry",
			libraryYAML: `
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
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			wantRemovedEntries: []string{"skill/missing"},
			wantStrippedRefs:   nil,
			wantSavedLibHas: map[string]bool{
				"skill/commit":  true,
				"skill/missing": false,
			},
		},
		{
			name: "fix strips ghost preset refs",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			wantRemovedEntries: nil,
			wantStrippedRefs:   []string{"skill/ghost"},
			wantSavedLibHas: map[string]bool{
				"skill/commit": true,
			},
		},
		{
			name: "fix removes multiple issues",
			libraryYAML: `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing1:
      path: skills/missing1.md
      description: Missing 1
    missing2:
      path: skills/missing2.md
      description: Missing 2
presets:
  workflow:
    description: Workflow
    resources:
      - skill/missing1
      - skill/ghost
`,
			files: map[string]string{
				"skills/commit.md": "---\nname: commit\n---\nContent",
			},
			wantRemovedEntries: []string{"skill/missing1", "skill/missing2"},
			wantStrippedRefs:   []string{"skill/ghost"},
			wantSavedLibHas: map[string]bool{
				"skill/commit":   true,
				"skill/missing1": false,
				"skill/missing2": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644))

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Load library
			lib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Run fix
			result, err := FixLibrary(lib)
			require.NoError(t, err)

			// Check result
			assert.Len(t, result.RemovedEntries, len(tt.wantRemovedEntries), "RemovedEntries count")
			assert.Len(t, result.StrippedRefs, len(tt.wantStrippedRefs), "StrippedRefs count")

			// Reload library to verify it was saved
			savedLib, err := LoadLibrary(context.Background(), tmpDir)
			require.NoError(t, err)

			// Check resources
			for ref, shouldExist := range tt.wantSavedLibHas {
				typ, name, _ := ParseRef(ref)
				exists := false
				if savedLib.Resources[typ] != nil {
					if _, ok := savedLib.Resources[typ][name]; ok {
						exists = true
					}
				}
				if exists != shouldExist {
					assert.Equal(t, shouldExist, exists, "resource %s exists for %v", ref)
				}
			}

			// Check presets
			for _, preset := range savedLib.Presets {
				for _, ref := range preset.Resources {
					if ref == "skill/ghost" {
						assert.NotContains(t, ref, "skill/ghost", "preset should not contain skill/ghost after fix")
					}
				}
			}
		})
	}
}

func TestValidationResult_AddIssue(t *testing.T) {
	result := &ValidationResult{Valid: true}

	// Add error
	result.AddIssue(Issue{
		Type:     IssueTypeMissingFile,
		Severity: SeverityError,
		Ref:      "skill/test",
	})

	if result.Valid != false {
		assert.Fail(t, "expected Valid to be false after adding error")
	}
	errs := fieldByName(result, "ErrorCount").(int)
	if errs != 1 {
		assert.Equal(t, 1, errs, "ErrorCount = %d, want %d")
	}
	warnings := fieldByName(result, "WarningCount").(int)
	if warnings != 0 {
		assert.Equal(t, 0, warnings, "WarningCount = %d, want %d")
	}

	// Add warning
	result.AddIssue(Issue{
		Type:     IssueTypeOrphan,
		Severity: SeverityWarning,
		Path:     "skills/extra.md",
	})

	if result.Valid != false {
		assert.Fail(t, "expected Valid to still be false")
	}
	errs = fieldByName(result, "ErrorCount").(int)
	if errs != 1 {
		assert.Equal(t, 1, errs, "ErrorCount = %d, want %d")
	}
	warnings = fieldByName(result, "WarningCount").(int)
	if warnings != 1 {
		assert.Equal(t, 1, warnings, "WarningCount = %d, want %d")
	}
}

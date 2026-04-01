package library

import (
	"os"
	"path/filepath"
	"testing"
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
			if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", path, err)
				}
			}

			// Load library
			lib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to load library: %v", err)
			}

			// Run check
			issues, err := CheckMissingFiles(lib)
			if err != nil {
				t.Fatalf("CheckMissingFiles() error = %v", err)
			}

			if len(issues) != tt.expectedIssues {
				t.Errorf("expected %d issues, got %d", tt.expectedIssues, len(issues))
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
					t.Errorf("expected issue with ref %q", expectedRef)
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
			if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", path, err)
				}
			}

			// Load library
			lib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to load library: %v", err)
			}

			// Run check
			issues, err := CheckOrphanedFiles(lib)
			if err != nil {
				t.Fatalf("CheckOrphanedFiles() error = %v", err)
			}

			if len(issues) != tt.expectedIssues {
				t.Errorf("expected %d issues, got %d", tt.expectedIssues, len(issues))
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
					t.Errorf("expected issue with path %q", expectedPath)
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
			if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Create skills directory and file
			skillsDir := filepath.Join(tmpDir, "skills")
			if err := os.MkdirAll(skillsDir, 0755); err != nil {
				t.Fatalf("Failed to create skills directory: %v", err)
			}
			if err := os.WriteFile(filepath.Join(skillsDir, "commit.md"), []byte("Content"), 0644); err != nil {
				t.Fatalf("Failed to write commit.md: %v", err)
			}

			// Load library
			lib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to load library: %v", err)
			}

			// Run check
			issues, err := CheckGhostResources(lib)
			if err != nil {
				t.Fatalf("CheckGhostResources() error = %v", err)
			}

			if len(issues) != tt.expectedIssues {
				t.Errorf("expected %d issues, got %d", tt.expectedIssues, len(issues))
			}

			for i, expectedRef := range tt.expectedRefs {
				if i >= len(issues) {
					t.Errorf("expected issue %d with ref %q", i, expectedRef)
					continue
				}
				if issues[i].Ref != expectedRef {
					t.Errorf("issue %d ref = %q, want %q", i, issues[i].Ref, expectedRef)
				}
				if issues[i].InPreset != tt.expectedInPresets[i] {
					t.Errorf("issue %d inPreset = %q, want %q", i, issues[i].InPreset, tt.expectedInPresets[i])
				}
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
			if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", path, err)
				}
			}

			// Load library
			lib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to load library: %v", err)
			}

			// Run check
			issues, err := CheckMalformedFrontmatter(lib)
			if err != nil {
				t.Fatalf("CheckMalformedFrontmatter() error = %v", err)
			}

			if len(issues) != tt.expectedIssues {
				t.Errorf("expected %d issues, got %d", tt.expectedIssues, len(issues))
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
					t.Errorf("expected issue with path %q", expectedPath)
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
			if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", path, err)
				}
			}

			// Load library
			lib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to load library: %v", err)
			}

			// Run validation
			result, err := ValidateLibrary(lib)
			if err != nil {
				t.Fatalf("ValidateLibrary() error = %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.ErrorCount != tt.wantErrors {
				t.Errorf("ErrorCount = %d, want %d", result.ErrorCount, tt.wantErrors)
			}
			if result.WarningCount != tt.wantWarnings {
				t.Errorf("WarningCount = %d, want %d", result.WarningCount, tt.wantWarnings)
			}
		})
	}
}

func TestFixLibrary(t *testing.T) {
	tests := []struct {
		name                  string
		libraryYAML           string
		files                 map[string]string
		wantMissingFileRefs   []string
		wantGhostResourceRefs []string
		wantSavedLibHas       map[string]bool // ref -> should exist after fix
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
			wantMissingFileRefs:   []string{"skill/missing"},
			wantGhostResourceRefs: nil,
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
			wantMissingFileRefs:   nil,
			wantGhostResourceRefs: []string{"skill/ghost"},
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
			wantMissingFileRefs:   []string{"skill/missing1", "skill/missing2"},
			wantGhostResourceRefs: []string{"skill/ghost"},
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
			if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Create directories and files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", path, err)
				}
			}

			// Load library
			lib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to load library: %v", err)
			}

			// Run fix
			result, err := FixLibrary(lib)
			if err != nil {
				t.Fatalf("FixLibrary() error = %v", err)
			}

			// Check result
			if len(result.MissingFileRefs) != len(tt.wantMissingFileRefs) {
				t.Errorf("MissingFileRefs count = %d, want %d", len(result.MissingFileRefs), len(tt.wantMissingFileRefs))
			}
			if len(result.GhostResourceRefs) != len(tt.wantGhostResourceRefs) {
				t.Errorf("GhostResourceRefs count = %d, want %d", len(result.GhostResourceRefs), len(tt.wantGhostResourceRefs))
			}

			// Reload library to verify it was saved
			savedLib, err := LoadLibrary(tmpDir)
			if err != nil {
				t.Fatalf("Failed to reload library: %v", err)
			}

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
					t.Errorf("resource %s exists = %v, want %v", ref, exists, shouldExist)
				}
			}

			// Check presets
			for presetName, preset := range savedLib.Presets {
				for _, ref := range preset.Resources {
					if ref == "skill/ghost" {
						t.Errorf("preset %s should not contain skill/ghost after fix", presetName)
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
		t.Error("expected Valid to be false after adding error")
	}
	if result.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", result.ErrorCount)
	}
	if result.WarningCount != 0 {
		t.Errorf("WarningCount = %d, want 0", result.WarningCount)
	}

	// Add warning
	result.AddIssue(Issue{
		Type:     IssueTypeOrphan,
		Severity: SeverityWarning,
		Path:     "skills/extra.md",
	})

	if result.Valid != false {
		t.Error("expected Valid to still be false")
	}
	if result.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", result.ErrorCount)
	}
	if result.WarningCount != 1 {
		t.Errorf("WarningCount = %d, want 1", result.WarningCount)
	}
}

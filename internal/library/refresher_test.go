package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRefreshLibrary(t *testing.T) {
	tests := []struct {
		name        string
		libraryYAML string
		files       map[string]string
		opts        RefreshOptions
		wantRefresh int
		wantSkipped int
		wantErrors  int
	}{
		{
			name: "refresh updates stale description",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: old description
`,
			files: map[string]string{
				"skills/commit.md": `---
name: commit
description: new description
---
# Commit

Best practices for commit messages.
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: false},
			wantRefresh: 1,
			wantSkipped: 0,
			wantErrors:  0,
		},
		{
			name: "refresh skips unchanged description",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: same description
`,
			files: map[string]string{
				"skills/commit.md": `---
name: commit
description: same description
---
# Commit
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: false},
			wantRefresh: 0,
			wantSkipped: 0,
			wantErrors:  0,
		},
		{
			name: "refresh discovers renamed file",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/old.md
      description: old description
`,
			files: map[string]string{
				"skills/commit.md": `---
name: commit
description: new description
---
# Commit
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: false},
			wantRefresh: 2,
			wantSkipped: 0,
			wantErrors:  0,
		},
		{
			name: "refresh skips missing file",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: description
`,
			files:       map[string]string{},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: false},
			wantRefresh: 0,
			wantSkipped: 0,
			wantErrors:  0,
		},
		{
			name: "refresh errors on name mismatch when file renamed to different name",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/old.md
      description: description
`,
			files: map[string]string{
				"skills/new.md": `---
name: new
description: description
---
# New
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: false},
			wantRefresh: 0,
			wantSkipped: 1,
			wantErrors:  1,
		},
		{
			name: "refresh skips silently on uncertain rename with force",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/old.md
      description: description
`,
			files: map[string]string{
				"skills/new.md": `---
name: new
description: description
---
# New
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: true},
			wantRefresh: 0,
			wantSkipped: 1,
			wantErrors:  0,
		},
		{
			name: "refresh errors on malformed frontmatter",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: description
`,
			files: map[string]string{
				"skills/commit.md": `---
invalid: yaml: content: that: is: malformed
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: false, Force: false},
			wantRefresh: 0,
			wantSkipped: 1,
			wantErrors:  1,
		},
		{
			name: "dry-run shows changes without modifying",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: old description
`,
			files: map[string]string{
				"skills/commit.md": `---
name: commit
description: new description
---
# Commit
`,
			},
			opts:        RefreshOptions{LibraryPath: "", DryRun: true, Force: false},
			wantRefresh: 1,
			wantSkipped: 0,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			libraryPath := filepath.Join(tmpDir, "library.yaml")
			if err := os.WriteFile(libraryPath, []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Write resource files
			for relPath, content := range tt.files {
				filePath := filepath.Join(tmpDir, relPath)
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", relPath, err)
				}
			}

			opts := tt.opts
			opts.LibraryPath = tmpDir

			result, err := RefreshLibrary(opts)
			if err != nil {
				t.Fatalf("RefreshLibrary() error = %v", err)
			}

			if len(result.Refreshed) != tt.wantRefresh {
				t.Errorf("RefreshLibrary() refreshed = %d, want %d", len(result.Refreshed), tt.wantRefresh)
			}
			if len(result.Skipped) != tt.wantSkipped {
				t.Errorf("RefreshLibrary() skipped = %d, want %d", len(result.Skipped), tt.wantSkipped)
			}
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("RefreshLibrary() errors = %d, want %d", len(result.Errors), tt.wantErrors)
			}
		})
	}
}

func TestDiscoverOrphans(t *testing.T) {
	tests := []struct {
		name          string
		libraryYAML   string
		files         map[string]string
		opts          DiscoverOptions
		wantOrphans   int
		wantAdded     int
		wantConflicts int
	}{
		{
			name: "discovers orphans not in library",
			libraryYAML: `version: "1"
resources:
  skill:
    existing:
      path: skills/existing.md
      description: existing skill
`,
			files: map[string]string{
				"skills/existing.md": `---
name: existing
description: existing skill
---
# Existing
`,
				"skills/new-skill.md": `---
name: new-skill
description: new skill description
---
# New Skill
`,
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:   1,
			wantAdded:     0,
			wantConflicts: 0,
		},
		{
			name: "discovers orphans in all directories",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/skill.md": "---\nname: skill\ndescription: s\n---\n",
				"agents/agent.md": "---\nname: agent\ndescription: a\n---\n",
				"commands/cmd.md": "---\nname: cmd\ndescription: c\n---\n",
				"memory/mem.md":   "---\nname: mem\ndescription: m\n---\n",
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:   4,
			wantAdded:     0,
			wantConflicts: 0,
		},
		{
			name: "no orphans when all registered",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: commit skill
`,
			files: map[string]string{
				"skills/commit.md": `---
name: commit
description: commit skill
---
# Commit
`,
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:   0,
			wantAdded:     0,
			wantConflicts: 0,
		},
		{
			name: "force registers orphans",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/new-skill.md": `---
name: new-skill
description: new skill
---
# New Skill
`,
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: true},
			wantOrphans:   1,
			wantAdded:     1,
			wantConflicts: 0,
		},
		{
			name: "name conflict with existing resource",
			libraryYAML: `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: existing commit
`,
			files: map[string]string{
				"skills/commit.md": `---
name: commit
description: existing commit
---
# Commit
`,
				"commands/commit.md": `---
name: commit
description: command commit
---
# Command Commit
`,
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:   0,
			wantAdded:     0,
			wantConflicts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			libraryPath := filepath.Join(tmpDir, "library.yaml")
			if err := os.WriteFile(libraryPath, []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Write resource files
			for relPath, content := range tt.files {
				filePath := filepath.Join(tmpDir, relPath)
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", relPath, err)
				}
			}

			opts := tt.opts
			opts.LibraryPath = tmpDir

			result, err := DiscoverOrphans(opts)
			if err != nil {
				t.Fatalf("DiscoverOrphans() error = %v", err)
			}

			if len(result.Orphans) != tt.wantOrphans {
				t.Errorf("DiscoverOrphans() orphans = %d, want %d", len(result.Orphans), tt.wantOrphans)
			}
			if len(result.Added) != tt.wantAdded {
				t.Errorf("DiscoverOrphans() added = %d, want %d", len(result.Added), tt.wantAdded)
			}
			if len(result.Conflicts) != tt.wantConflicts {
				t.Errorf("DiscoverOrphans() conflicts = %d, want %d", len(result.Conflicts), tt.wantConflicts)
			}
		})
	}
}

func TestDiscoverOrphans_RecursiveScanning(t *testing.T) {
	tests := []struct {
		name           string
		libraryYAML    string
		files          map[string]string
		opts           DiscoverOptions
		wantOrphans    int
		wantScanned    int
		wantTotalAdded int
	}{
		{
			name: "discovers orphans in nested directories",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/skill.md":             "---\nname: skill\ndescription: s\n---\n",
				"skills/nested/deep/skill.md": "---\nname: nested-skill\ndescription: nested\n---\n",
				"skills/another/skill.md":     "---\nname: another-skill\ndescription: another\n---\n",
			},
			opts:           DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:    3,
			wantScanned:    3,
			wantTotalAdded: 0,
		},
		{
			name: "tracks total scanned count correctly",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/skill1.md": "---\nname: skill1\ndescription: s1\n---\n",
				"skills/skill2.md": "---\nname: skill2\ndescription: s2\n---\n",
				"agents/agent1.md": "---\nname: agent1\ndescription: a1\n---\n",
			},
			opts:           DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:    3,
			wantScanned:    3,
			wantTotalAdded: 0,
		},
		{
			name: "recursive with force adds all orphans",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/skill.md":             "---\nname: skill\ndescription: s\n---\n",
				"skills/nested/deep/skill.md": "---\nname: nested-skill\ndescription: nested\n---\n",
			},
			opts:           DiscoverOptions{LibraryPath: "", DryRun: false, Force: true},
			wantOrphans:    2,
			wantScanned:    2,
			wantTotalAdded: 2,
		},
		{
			name: "skips non-md files",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/skill.md":   "---\nname: skill\ndescription: s\n---\n",
				"skills/readme.txt": "README content",
				"skills/data.json":  "{}",
			},
			opts:           DiscoverOptions{LibraryPath: "", DryRun: false, Force: false},
			wantOrphans:    1,
			wantScanned:    1, // Only .md files are counted
			wantTotalAdded: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			libraryPath := filepath.Join(tmpDir, "library.yaml")
			if err := os.WriteFile(libraryPath, []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Write resource files
			for relPath, content := range tt.files {
				filePath := filepath.Join(tmpDir, relPath)
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", relPath, err)
				}
			}

			opts := tt.opts
			opts.LibraryPath = tmpDir

			result, err := DiscoverOrphans(opts)
			if err != nil {
				t.Fatalf("DiscoverOrphans() error = %v", err)
			}

			if len(result.Orphans) != tt.wantOrphans {
				t.Errorf("DiscoverOrphans() orphans = %d, want %d", len(result.Orphans), tt.wantOrphans)
			}
			if result.Summary.TotalScanned != tt.wantScanned {
				t.Errorf("DiscoverOrphans() TotalScanned = %d, want %d", result.Summary.TotalScanned, tt.wantScanned)
			}
			if result.Summary.TotalAdded != tt.wantTotalAdded {
				t.Errorf("DiscoverOrphans() TotalAdded = %d, want %d", result.Summary.TotalAdded, tt.wantTotalAdded)
			}
		})
	}
}

func TestDiscoverOrphans_BatchMode(t *testing.T) {
	tests := []struct {
		name          string
		libraryYAML   string
		files         map[string]string
		opts          DiscoverOptions
		wantOrphans   int
		wantAdded     int
		wantFailed    int
		wantConflicts int
	}{
		{
			name: "batch mode reports orphans without auto-registering",
			libraryYAML: `version: "1"
resources: {}
`,
			files: map[string]string{
				"skills/skill1.md": "---\nname: skill1\ndescription: s1\n---\n",
				"skills/skill2.md": "---\nname: skill2\ndescription: s2\n---\n",
				"skills/skill3.md": "---\nname: skill3\ndescription: s3\n---\n",
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: true, Batch: true},
			wantOrphans:   3,
			wantAdded:     0, // Batch mode no longer auto-registers; CLI uses BatchAddResources
			wantFailed:    0,
			wantConflicts: 0,
		},
		{
			name: "batch mode with conflicts reports conflicting",
			libraryYAML: `version: "1"
resources:
  skill:
    skill1:
      path: skills/skill1.md
      description: existing skill1
`,
			files: map[string]string{
				"skills/skill1.md": "---\nname: skill1\ndescription: existing\n---\n",
				"skills/skill2.md": "---\nname: skill2\ndescription: s2\n---\n",
				"skills/skill3.md": "---\nname: skill3\ndescription: s3\n---\n",
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: true, Batch: true},
			wantOrphans:   2, // skill2 and skill3 (skill1 is registered)
			wantAdded:     0, // Batch mode no longer auto-registers
			wantFailed:    0,
			wantConflicts: 0,
		},
		{
			name: "batch mode reports orphans without registration",
			libraryYAML: `version: "1"
resources:
  skill:
    skill1:
      path: skills/skill1.md
      description: existing
    skill2:
      path: skills/skill2.md
      description: existing
`,
			files: map[string]string{
				"skills/skill1.md": "---\nname: skill1\ndescription: existing\n---\n",
				"skills/skill2.md": "---\nname: skill2\ndescription: existing\n---\n",
				"skills/skill3.md": "---\nname: skill3\ndescription: s3\n---\n",
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: true, Batch: true},
			wantOrphans:   1, // Only skill3 (skill1 and skill2 are registered)
			wantAdded:     0, // Batch mode no longer auto-registers
			wantFailed:    0,
			wantConflicts: 0,
		},
		{
			name: "non-batch force mode fails on conflict",
			libraryYAML: `version: "1"
resources:
  skill:
    skill:
      path: skills/skill.md
      description: existing skill
`,
			files: map[string]string{
				"skills/skill.md":   "---\nname: skill\ndescription: s\n---\n",
				"commands/skill.md": "---\nname: skill\ndescription: command skill\n---\n",
			},
			opts:          DiscoverOptions{LibraryPath: "", DryRun: false, Force: true, Batch: false},
			wantOrphans:   0, // skill in commands has conflict with existing skill in library
			wantAdded:     0,
			wantFailed:    0,
			wantConflicts: 1, // Name conflict blocks registration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write library.yaml
			libraryPath := filepath.Join(tmpDir, "library.yaml")
			if err := os.WriteFile(libraryPath, []byte(tt.libraryYAML), 0644); err != nil {
				t.Fatalf("Failed to write library.yaml: %v", err)
			}

			// Write resource files
			for relPath, content := range tt.files {
				filePath := filepath.Join(tmpDir, relPath)
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", relPath, err)
				}
			}

			opts := tt.opts
			opts.LibraryPath = tmpDir

			result, err := DiscoverOrphans(opts)
			if err != nil {
				t.Fatalf("DiscoverOrphans() error = %v", err)
			}

			if len(result.Orphans) != tt.wantOrphans {
				t.Errorf("DiscoverOrphans() orphans = %d, want %d", len(result.Orphans), tt.wantOrphans)
			}
			if len(result.Added) != tt.wantAdded {
				t.Errorf("DiscoverOrphans() added = %d, want %d", len(result.Added), tt.wantAdded)
			}
			if result.Summary.TotalFailed != tt.wantFailed {
				t.Errorf("DiscoverOrphans() TotalFailed = %d, want %d", result.Summary.TotalFailed, tt.wantFailed)
			}
			if len(result.Conflicts) != tt.wantConflicts {
				t.Errorf("DiscoverOrphans() conflicts = %d, want %d", len(result.Conflicts), tt.wantConflicts)
			}
		})
	}
}

func TestDiscoverOrphans_Summary(t *testing.T) {
	tmpDir := t.TempDir()

	// Write library.yaml
	libraryYAML := `version: "1"
resources: {}
`
	libraryPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(libraryPath, []byte(libraryYAML), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	// Write resource files
	files := map[string]string{
		"skills/skill1.md": "---\nname: skill1\ndescription: s1\n---\n",
		"skills/skill2.md": "---\nname: skill2\ndescription: s2\n---\n",
		"agents/agent1.md": "---\nname: agent1\ndescription: a1\n---\n",
	}
	for relPath, content := range files {
		filePath := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", relPath, err)
		}
	}

	result, err := DiscoverOrphans(DiscoverOptions{
		LibraryPath: tmpDir,
		DryRun:      false,
		Force:       false,
		Batch:       false,
	})
	if err != nil {
		t.Fatalf("DiscoverOrphans() error = %v", err)
	}

	// Verify summary fields
	if result.Summary.TotalScanned != 3 {
		t.Errorf("Summary.TotalScanned = %d, want 3", result.Summary.TotalScanned)
	}
	if result.Summary.TotalOrphans != 3 {
		t.Errorf("Summary.TotalOrphans = %d, want 3", result.Summary.TotalOrphans)
	}
	if result.Summary.TotalAdded != 0 {
		t.Errorf("Summary.TotalAdded = %d, want 0", result.Summary.TotalAdded)
	}
	if result.Summary.TotalSkipped != 0 {
		t.Errorf("Summary.TotalSkipped = %d, want 0", result.Summary.TotalSkipped)
	}
	if result.Summary.TotalFailed != 0 {
		t.Errorf("Summary.TotalFailed = %d, want 0", result.Summary.TotalFailed)
	}
}

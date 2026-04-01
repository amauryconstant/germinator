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

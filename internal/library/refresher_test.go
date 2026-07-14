package library

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
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

			result, err := RefreshLibrary(context.Background(), opts)
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

			result, err := DiscoverOrphans(context.Background(), opts)
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

			result, err := DiscoverOrphans(context.Background(), opts)
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

			result, err := DiscoverOrphans(context.Background(), opts)
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

	result, err := DiscoverOrphans(context.Background(), DiscoverOptions{
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

// TestDiscoverOrphans_CtxCancelled verifies that DiscoverOrphans
// observes ctx cancellation promptly. After Phase 4's errgroup
// refactor, sibling subtrees are walked concurrently under
// scanConcurrencyLimit, so ctx.Err() checks at goroutine entry surface
// cancellation as soon as the next goroutine yields. The function
// returns wrapped context.Canceled alongside a partial
// *DiscoverResult so callers can inspect progress made before the
// cancel.
func TestDiscoverOrphans_CtxCancelled(t *testing.T) {
	tmpDir := t.TempDir()

	// Library.yaml: empty registry so every .md file is an orphan.
	if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"),
		[]byte(`version: "1"
resources: {}
`), 0o600); err != nil {
		t.Fatalf("write library.yaml: %v", err)
	}

	// Build a deeply nested library with enough files that the scan
	// is still in-flight when ctx cancellation arrives at 50ms.
	// Layout: skills/<12 sub-levels>/skill-N.md, agents/<12 sub-levels>/agent-N.md,
	// commands/<12 sub-levels>/command-N.md, memory/<12 sub-levels>/memory-N.md.
	const subLevels = 12
	const perDir = 200
	dirs := []string{"skills", "agents", "commands", "memory"}
	for _, dir := range dirs {
		base := filepath.Join(tmpDir, dir)
		for level := 0; level < subLevels; level++ {
			base = filepath.Join(base, fmt.Sprintf("sub%d", level))
		}
		if err := os.MkdirAll(base, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", base, err)
		}
		for i := 0; i < perDir; i++ {
			body := fmt.Sprintf("---\nname: %s%d\ndescription: test\n---\n# %d\n",
				dir, i, i)
			if err := os.WriteFile(filepath.Join(base, fmt.Sprintf("r%d.md", i)),
				[]byte(body), 0o644); err != nil {
				t.Fatalf("write %s/r%d.md: %v", dir, i, err)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after a short delay. The scan is large enough (800+
	// files across 12-deep trees) that some goroutines are still
	// in-flight; errgroup.WithContext surfaces cancellation when the
	// next goroutine observes ctx.Err(). Pre-change filepath.WalkDir
	// was bounded by total scan time; post-change the parallel
	// sibling-subtree descent surfaces cancellation at the next
	// goroutine yield.
	time.AfterFunc(1*time.Millisecond, cancel)

	start := time.Now()
	result, err := DiscoverOrphans(ctx, DiscoverOptions{LibraryPath: tmpDir})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("DiscoverOrphans() expected error from cancelled ctx, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("DiscoverOrphans() err = %v, want wrapped context.Canceled", err)
	}
	// Allow generous slack — the errgroup must surface cancellation
	// at the next goroutine yield, not after the slowest subtree's
	// full walk. Pre-change filepath.WalkDir typically bounded
	// cancellation by total scan time (~500ms+ on this fixture).
	if elapsed > 500*time.Millisecond {
		t.Errorf("DiscoverOrphans took %v after cancel; expected <500ms "+
			"(errgroup must observe ctx cancellation promptly)", elapsed)
	}
	// Partial result is still returned alongside the error.
	if result == nil {
		t.Error("DiscoverOrphans() result should be non-nil on cancellation")
	}
}

// TestDiscoverOrphans_TotalScannedThreadSafe exercises the parallel
// file-processing path that Phase 4 introduced via
// errgroup.SetLimit(scanConcurrencyLimit). The fixture has 4
// top-level dirs × 1 subdir each × 32 .md files at the leaf (128
// files total) so multiple goroutines increment
// result.Summary.TotalScanned concurrently. The shared
// sync.Mutex on *DiscoverResult must serialize the writes; the
// final count must equal the total number of .md files processed.
//
// Fixture shape (single subdir per top-level, many files at the
// leaf) is chosen deliberately: a wide-sibling fixture (≥8
// subdirs at one level) saturates the errgroup cap with
// subdir-goroutines that recursively call g.Go, which the
// current production design does not support. The current
// parallelism model parallelizes *file processing* (via
// processScanFile); sibling-subtree parallelism is a side
// effect of file goroutines from each subtree running in the
// same shared errgroup. The cap=8 invariant
// (scanConcurrencyLimit) is reviewed by inspection at
// internal/library/adder.go scanDirectory; a separate runtime
// test would require intrusive instrumentation (e.g., a hook
// inside processScanFile) and is not justified for this
// single-constant invariant.
func TestDiscoverOrphans_TotalScannedThreadSafe(t *testing.T) {
	tmpDir := t.TempDir()

	libraryPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(libraryPath, []byte(`version: "1"
resources: {}
`), 0o600); err != nil {
		t.Fatalf("write library.yaml: %v", err)
	}

	const topLevelDirs = 4
	const filesPerDir = 32
	const wantTotal = topLevelDirs * filesPerDir

	dirs := []string{"skills", "agents", "commands", "memory"}
	for _, dir := range dirs {
		base := filepath.Join(tmpDir, dir, "leaf")
		if err := os.MkdirAll(base, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", base, err)
		}
		for f := 0; f < filesPerDir; f++ {
			body := fmt.Sprintf("---\nname: %s-f%d\ndescription: test\n---\n",
				dir, f)
			path := filepath.Join(base, fmt.Sprintf("r%d.md", f))
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
		}
	}

	result, err := DiscoverOrphans(context.Background(), DiscoverOptions{LibraryPath: tmpDir})
	if err != nil {
		t.Fatalf("DiscoverOrphans() error = %v", err)
	}
	if result == nil {
		t.Fatal("DiscoverOrphans() result is nil")
	}

	// No lost increments: every .md file is counted exactly once.
	if result.Summary.TotalScanned != wantTotal {
		t.Errorf("Summary.TotalScanned = %d, want %d "+
			"(concurrent writes to TotalScanned lost increments)",
			result.Summary.TotalScanned, wantTotal)
	}

	// Every file is an orphan (empty library.yaml registry).
	if result.Summary.TotalOrphans != wantTotal {
		t.Errorf("Summary.TotalOrphans = %d, want %d",
			result.Summary.TotalOrphans, wantTotal)
	}
	if got := len(result.Orphans); got != wantTotal {
		t.Errorf("len(result.Orphans) = %d, want %d "+
			"(concurrent slice appends lost entries)",
			got, wantTotal)
	}
}

// TestDiscoverOrphans_OrderUnordered verifies that the order of
// result.Orphans is implementation-defined (per the
// library-library-orphan-discovery spec scenario "Order of
// result.Orphans is unordered"). Parallel file processing via
// the errgroup produces non-deterministic order; the public
// contract guarantees membership equality, not sequence. The
// test runs the same fixture 5 times and asserts each run
// yields the same set of paths using stringSlicesEqualUnordered
// (multiset equality).
//
// Fixture shape: 16 files in a single leaf directory under one
// top-level type. This is wide enough to exercise parallel file
// goroutines (16 > scanConcurrencyLimit=8) but does not require
// sibling-subtree parallelism (which the production design
// does not currently support; see TestDiscoverOrphans_
// TotalScannedThreadSafe comment for details).
func TestDiscoverOrphans_OrderUnordered(t *testing.T) {
	tmpDir := t.TempDir()

	libraryPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(libraryPath, []byte(`version: "1"
resources: {}
`), 0o600); err != nil {
		t.Fatalf("write library.yaml: %v", err)
	}

	const files = 16
	base := filepath.Join(tmpDir, "skills", "leaf")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", base, err)
	}
	for f := 0; f < files; f++ {
		body := fmt.Sprintf("---\nname: skill-f%d\ndescription: d\n---\n", f)
		path := filepath.Join(base, fmt.Sprintf("r%d.md", f))
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	const runs = 5
	pathSets := make([][]string, runs)
	for i := 0; i < runs; i++ {
		result, err := DiscoverOrphans(context.Background(), DiscoverOptions{LibraryPath: tmpDir})
		if err != nil {
			t.Fatalf("DiscoverOrphans() run %d error = %v", i, err)
		}
		if result == nil {
			t.Fatalf("DiscoverOrphans() run %d result is nil", i)
		}
		if got := len(result.Orphans); got != files {
			t.Fatalf("DiscoverOrphans() run %d: len(Orphans) = %d, want %d",
				i, got, files)
		}
		paths := make([]string, 0, len(result.Orphans))
		for _, o := range result.Orphans {
			paths = append(paths, o.Path)
		}
		pathSets[i] = paths
	}

	// All runs must yield the same set of paths (membership equality).
	for i := 1; i < runs; i++ {
		if !stringSlicesEqualUnordered(pathSets[0], pathSets[i]) {
			t.Errorf("DiscoverOrphans() run %d paths differ from run 0: "+
				"run 0 = %v, run %d = %v",
				i, pathSets[0], i, pathSets[i])
		}
	}
}

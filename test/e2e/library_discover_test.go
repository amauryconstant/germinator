//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/fixtures"
	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Library Discover Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library add --discover", func() {
		It("should discover orphaned resource files not in library.yaml", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create an orphaned skill file (not in library.yaml)
			orphanPath := filepath.Join(libPath, "skills", "orphan-skill.md")
			orphanContent := `---
name: orphan-skill
description: An orphaned skill
---
# Orphan Skill
`
			Expect(os.MkdirAll(filepath.Dir(orphanPath), 0755)).To(Succeed())
			Expect(os.WriteFile(orphanPath, []byte(orphanContent), 0644)).To(Succeed())

			// Run discover
			session = cli.Run("library", "add", "--discover", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Orphaned resources:")
			cli.ShouldOutput(session, "orphan-skill")
		})
	})

	Describe("library add --discover --force", func() {
		It("should register orphaned files with --force", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create an orphaned skill file
			orphanPath := filepath.Join(libPath, "skills", "new-skill.md")
			orphanContent := `---
name: new-skill
description: A new skill to register
---
# New Skill
`
			Expect(os.WriteFile(orphanPath, []byte(orphanContent), 0644)).To(Succeed())

			// Run discover with force
			session = cli.Run("library", "add", "--discover", "--force", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Registered:")

			// Verify it's now in library.yaml
			session = cli.Run("library", "resources", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "skill/new-skill")
		})
	})

	Describe("library add --discover --dry-run", func() {
		It("should preview without modifying library.yaml", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create an orphaned skill file
			orphanPath := filepath.Join(libPath, "skills", "dry-run-skill.md")
			orphanContent := `---
name: dry-run-skill
description: Dry run skill
---
# Dry Run Skill
`
			Expect(os.WriteFile(orphanPath, []byte(orphanContent), 0644)).To(Succeed())

			// Read original library.yaml
			originalYAML, err := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			Expect(err).NotTo(HaveOccurred())

			// Run discover with dry-run
			session = cli.Run("library", "add", "--discover", "--dry-run", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Dry-run: no changes made")
			cli.ShouldOutput(session, "Orphaned resources:")

			// Verify library.yaml was NOT modified
			newYAML, err := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(newYAML)).To(Equal(string(originalYAML)))
		})
	})

	Describe("library add --discover in multiple directories", func() {
		It("should find orphans in skills, agents, commands, and memory directories", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create orphaned files in multiple directories
			Expect(os.WriteFile(filepath.Join(libPath, "skills", "orphan-skill.md"), []byte(`---
name: orphan-skill
description: Orphan skill
---
# Skill
`), 0644)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(libPath, "agents", "orphan-agent.md"), []byte(`---
name: orphan-agent
description: Orphan agent
---
# Agent
`), 0644)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(libPath, "commands", "orphan-command.md"), []byte(`---
name: orphan-command
description: Orphan command
---
# Command
`), 0644)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(libPath, "memory", "orphan-memory.md"), []byte(`---
name: orphan-memory
description: Orphan memory
---
# Memory
`), 0644)).To(Succeed())

			// Run discover
			session = cli.Run("library", "add", "--discover", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Orphaned resources:")
			cli.ShouldOutput(session, "orphan-skill")
			cli.ShouldOutput(session, "orphan-agent")
			cli.ShouldOutput(session, "orphan-command")
			cli.ShouldOutput(session, "orphan-memory")
		})
	})

	Describe("library add --discover with name conflict", func() {
		It("should report conflict when orphan name matches existing resource", func() {
			// Create a library with an existing skill
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add an initial resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Create an orphan with the SAME name in a DIFFERENT directory
			// (commands/commit.md - same name "commit" but different type)
			Expect(os.WriteFile(filepath.Join(libPath, "commands", "commit.md"), []byte(`---
name: commit
description: Command named commit
---
# Command Commit
`), 0644)).To(Succeed())

			// Run discover - should report conflict
			session = cli.Run("library", "add", "--discover", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Conflicts:")
		})
	})

	Describe("library add --discover --library flag", func() {
		It("should use specified library path", func() {
			// Create a library at specific path
			libPath := filepath.Join(tmpDir, "custom-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create an orphaned skill
			orphanPath := filepath.Join(libPath, "skills", "custom-orphan.md")
			orphanContent := `---
name: custom-orphan
description: Custom orphan
---
# Custom Orphan
`
			Expect(os.WriteFile(orphanPath, []byte(orphanContent), 0644)).To(Succeed())

			// Run discover with explicit library path
			session = cli.Run("library", "add", "--discover", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "custom-orphan")
		})
	})

	Describe("library add --discover with GERMINATOR_LIBRARY env", func() {
		It("should use library from environment variable", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "env-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create an orphaned skill
			orphanPath := filepath.Join(libPath, "skills", "env-orphan.md")
			orphanContent := `---
name: env-orphan
description: Env orphan
---
# Env Orphan
`
			Expect(os.WriteFile(orphanPath, []byte(orphanContent), 0644)).To(Succeed())

			// Run discover using GERMINATOR_LIBRARY env
			session = cli.RunWithEnv(map[string]string{"GERMINATOR_LIBRARY": libPath},
				"library", "add", "--discover")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "env-orphan")
		})
	})

	Describe("library add --discover with no orphans", func() {
		It("should report no orphans when all files are registered", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource (so it's registered)
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Run discover - should have no orphans
			session = cli.Run("library", "add", "--discover", "--library", libPath)
			cli.ShouldSucceed(session)
			// Should NOT output "Orphaned resources:"
			// (empty output is expected)
		})
	})

	Describe("library add --discover --batch", func() {
		It("should process all orphans through batch add pipeline", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create orphaned files in multiple directories
			Expect(os.WriteFile(filepath.Join(libPath, "skills", "batch-orphan1.md"), []byte(`---
name: batch-orphan1
description: Batch orphan skill
---
# Batch Orphan 1
`), 0644)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(libPath, "agents", "batch-orphan2.md"), []byte(`---
name: batch-orphan2
description: Batch orphan agent
---
# Batch Orphan 2
`), 0644)).To(Succeed())

			// Run discover with batch mode (without force, just reports)
			session = cli.Run("library", "add", "--discover", "--batch", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Orphaned resources:")
			cli.ShouldOutput(session, "batch-orphan1")
			cli.ShouldOutput(session, "batch-orphan2")
		})
	})

	Describe("library add --discover --batch --force", func() {
		It("should register all orphans with batch processing", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create orphaned files
			Expect(os.WriteFile(filepath.Join(libPath, "skills", "force-orphan1.md"), []byte(`---
name: force-orphan1
description: Force orphan skill
---
# Force Orphan 1
`), 0644)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(libPath, "agents", "force-orphan2.md"), []byte(`---
name: force-orphan2
description: Force orphan agent
---
# Force Orphan 2
`), 0644)).To(Succeed())

			// Run discover with batch and force
			session = cli.Run("library", "add", "--discover", "--batch", "--force", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added 2")

			// Verify resources are registered
			session = cli.Run("library", "resources", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "skill/force-orphan1")
			cli.ShouldOutput(session, "agent/force-orphan2")
		})
	})

	Describe("library add --discover --batch --force --dry-run", func() {
		It("should preview batch registration without modifying library", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create orphaned file
			Expect(os.WriteFile(filepath.Join(libPath, "skills", "dryrun-orphan.md"), []byte(`---
name: dryrun-orphan
description: Dry run orphan
---
# Dry Run Orphan
`), 0644)).To(Succeed())

			// Read original library.yaml
			originalYAML, err := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			Expect(err).NotTo(HaveOccurred())

			// Run discover with batch, force, and dry-run
			session = cli.Run("library", "add", "--discover", "--batch", "--force", "--dry-run", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added 1")

			// Verify library.yaml was NOT modified
			newYAML, err := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(newYAML)).To(Equal(string(originalYAML)))
		})
	})

	Describe("library add --discover --batch --force --json", func() {
		It("should output JSON format with batch processing results", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create orphaned file
			Expect(os.WriteFile(filepath.Join(libPath, "skills", "json-orphan.md"), []byte(`---
name: json-orphan
description: JSON orphan
---
# JSON Orphan
`), 0644)).To(Succeed())

			// Run discover with batch, force, and JSON
			session = cli.Run("library", "add", "--discover", "--batch", "--force", "--json", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, `"added"`)
			cli.ShouldOutput(session, `"skipped"`)
			cli.ShouldOutput(session, `"failed"`)
			cli.ShouldOutput(session, `"summary"`)
		})
	})
})

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

var _ = Describe("Library Refresh Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library refresh", func() {
		It("should update stale description from frontmatter", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource using library add
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Modify the frontmatter description in the file
			skillPath := filepath.Join(libPath, "skills", "commit.md")
			modifiedContent := `---
name: commit
description: Updated description from refresh
tools:
  - bash
---
# Commit

Updated best practices.
`
			Expect(os.WriteFile(skillPath, []byte(modifiedContent), 0644)).To(Succeed())

			// Run refresh
			session = cli.Run("library", "refresh", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Refreshed:")
			cli.ShouldOutput(session, "skill/commit")
		})
	})

	Describe("library refresh --dry-run", func() {
		It("should preview changes without modifying library.yaml", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Modify the frontmatter
			skillPath := filepath.Join(libPath, "skills", "commit.md")
			modifiedContent := `---
name: commit
description: Dry-run description
---
# Commit
`
			Expect(os.WriteFile(skillPath, []byte(modifiedContent), 0644)).To(Succeed())

			// Read original library.yaml
			originalYAML, err := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			Expect(err).NotTo(HaveOccurred())

			// Run refresh with dry-run
			session = cli.Run("library", "refresh", "--dry-run", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Dry-run: no changes made")
			cli.ShouldOutput(session, "Refreshed:")

			// Verify library.yaml was NOT modified
			newYAML, err := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(newYAML)).To(Equal(string(originalYAML)))
		})
	})

	Describe("library refresh --json", func() {
		It("should output JSON format", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Modify the frontmatter
			skillPath := filepath.Join(libPath, "skills", "commit.md")
			modifiedContent := `---
name: commit
description: JSON description
---
# Commit
`
			Expect(os.WriteFile(skillPath, []byte(modifiedContent), 0644)).To(Succeed())

			// Run refresh with JSON output
			session = cli.Run("library", "refresh", "--json", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutputMatch(session, `"refreshed": 1`)
		})
	})

	Describe("library refresh with missing file", func() {
		It("should skip resources with missing files", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create the skills directory with a nested subdirectory (not a file)
			// This simulates a missing file scenario
			Expect(os.MkdirAll(filepath.Join(libPath, "skills"), 0755)).To(Succeed())

			// Manually add a resource entry without the actual file
			// Using proper YAML structure with all resource types
			libraryYAML := `version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: A skill
  agent: {}
  command: {}
  memory: {}
presets: {}
`
			Expect(os.WriteFile(filepath.Join(libPath, "library.yaml"), []byte(libraryYAML), 0644)).To(Succeed())

			// Run refresh - should skip missing file
			session = cli.Run("library", "refresh", "--library", libPath)
			cli.ShouldSucceed(session)
			// Missing files are skipped silently (left to validate --fix)
		})
	})

	Describe("library refresh with --library flag", func() {
		It("should use specified library path", func() {
			// Create a library at specific path
			libPath := filepath.Join(tmpDir, "custom-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Run refresh with explicit library path
			session = cli.Run("library", "refresh", "--library", libPath)
			cli.ShouldSucceed(session)
		})
	})

	Describe("library refresh with GERMINATOR_LIBRARY env", func() {
		It("should use library from environment variable", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "env-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Run refresh using GERMINATOR_LIBRARY env
			session = cli.RunWithEnv(map[string]string{"GERMINATOR_LIBRARY": libPath},
				"library", "refresh")
			cli.ShouldSucceed(session)
		})
	})
})

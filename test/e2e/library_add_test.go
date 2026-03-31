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

var _ = Describe("Library Add Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library add <source>", func() {
		It("should add a canonical skill to the library", func() {
			// Create a library first
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add the skill
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added resource: skill/commit")

			// Verify file was copied
			Expect(filepath.Join(libPath, "skills", "commit.md")).To(BeAnExistingFile())

			// Verify library.yaml was updated
			session = cli.Run("library", "resources", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "skill/commit")
		})
	})

	Describe("library add with --type flag", func() {
		It("should use explicit type over filename detection", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create a source file without type in frontmatter but with skill prefix
			sourceDir := filepath.Join(tmpDir, "source")
			Expect(os.MkdirAll(sourceDir, 0o755)).To(Succeed())
			sourcePath := filepath.Join(sourceDir, "my-resource.md")
			content := `---
name: my-resource
description: My custom resource
---
# My Resource`
			Expect(os.WriteFile(sourcePath, []byte(content), 0o644)).To(Succeed())

			// Add with explicit type
			session = cli.Run("library", "add", sourcePath, "--type", "skill", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added resource: skill/my-resource")

			// Verify file was copied to skills directory
			Expect(filepath.Join(libPath, "skills", "my-resource.md")).To(BeAnExistingFile())
		})
	})

	Describe("library add with --name flag", func() {
		It("should use explicit name over frontmatter", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create a source file with name in frontmatter
			sourceDir := filepath.Join(tmpDir, "source")
			Expect(os.MkdirAll(sourceDir, 0o755)).To(Succeed())
			sourcePath := filepath.Join(sourceDir, "skill-test.md")
			content := `---
name: original-name
description: Test description
---
# Test Skill`
			Expect(os.WriteFile(sourcePath, []byte(content), 0o644)).To(Succeed())

			// Add with explicit name
			session = cli.Run("library", "add", sourcePath, "--name", "custom-name", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added resource: skill/custom-name")

			// Verify file was copied with proper name
			Expect(filepath.Join(libPath, "skills", "custom-name.md")).To(BeAnExistingFile())
		})
	})

	Describe("library add with --description flag", func() {
		It("should use explicit description over frontmatter", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create a source file with description in frontmatter
			sourceDir := filepath.Join(tmpDir, "source")
			Expect(os.MkdirAll(sourceDir, 0o755)).To(Succeed())
			sourcePath := filepath.Join(sourceDir, "skill-test.md")
			content := `---
name: test-skill
description: Original description
---
# Test Skill`
			Expect(os.WriteFile(sourcePath, []byte(content), 0o644)).To(Succeed())

			// Add with explicit description
			session = cli.Run("library", "add", sourcePath, "--description", "New description", "--library", libPath)
			cli.ShouldSucceed(session)

			// Verify library resources shows the new description
			session = cli.Run("library", "resources", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "New description")
		})
	})

	Describe("library add error on existing resource", func() {
		It("should error without --force when resource exists", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Try to add again without --force
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldFailWithExit(session, 1)
			cli.ShouldErrorOutput(session, "already exists")
		})
	})

	Describe("library add with --force", func() {
		It("should replace existing resource with --force", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Modify the source file
			modifiedContent := `---
name: commit
description: Modified description
tools:
  - bash
---
# Modified Commit Skill

Modified content`
			Expect(os.WriteFile(sourcePath, []byte(modifiedContent), 0o644)).To(Succeed())

			// Add again with --force
			session = cli.Run("library", "add", sourcePath, "--force", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added resource: skill/commit")

			// Verify content was updated
			content, err := os.ReadFile(filepath.Join(libPath, "skills", "commit.md"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("Modified Commit Skill"))
		})
	})

	Describe("library add --dry-run", func() {
		It("should preview without creating files", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add with dry-run
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--dry-run", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Would add resource: skill/commit")

			// Verify nothing was created
			Expect(filepath.Join(libPath, "skills", "commit.md")).NotTo(BeAnExistingFile())
		})
	})

	Describe("library add with --library flag", func() {
		It("should use specified library path", func() {
			// Create a library at specific path
			libPath := filepath.Join(tmpDir, "custom-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add resource using --library flag
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added resource: skill/commit")

			// Verify it was added to the correct library
			Expect(filepath.Join(libPath, "skills", "commit.md")).To(BeAnExistingFile())
		})
	})

	Describe("library add with GERMINATOR_LIBRARY env", func() {
		It("should use library from environment variable", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "env-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add resource using GERMINATOR_LIBRARY env
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.RunWithEnv(map[string]string{"GERMINATOR_LIBRARY": libPath},
				"library", "add", sourcePath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Added resource: skill/commit")

			// Verify it was added
			Expect(filepath.Join(libPath, "skills", "commit.md")).To(BeAnExistingFile())
		})
	})

	Describe("library add with nonexistent source", func() {
		It("should error with helpful message", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Try to add nonexistent file - exit code 6 is NotFound
			session = cli.Run("library", "add", "/nonexistent/file.md", "--library", libPath)
			cli.ShouldFailWithExit(session, 6)
			cli.ShouldErrorOutput(session, "source file not found")
		})
	})

	Describe("library add with invalid type", func() {
		It("should error with valid type suggestions", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Create a source file
			sourceDir := filepath.Join(tmpDir, "source")
			Expect(os.MkdirAll(sourceDir, 0o755)).To(Succeed())
			sourcePath := filepath.Join(sourceDir, "test.md")
			content := `---
name: test
description: Test
---
# Test`
			Expect(os.WriteFile(sourcePath, []byte(content), 0o644)).To(Succeed())

			// Try to add with invalid type - exit code 3 is Config error
			session = cli.Run("library", "add", sourcePath, "--type", "invalid", "--library", libPath)
			cli.ShouldFailWithExit(session, 3)
			cli.ShouldErrorOutput(session, "invalid resource type")
		})
	})

	// Note: Missing name scenario is not easily testable because filename
	// extraction always produces a non-empty name from files with extensions.
	// The implementation uses filename as fallback when frontmatter name is missing.
})

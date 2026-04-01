//go:build e2e

package e2e_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/fixtures"
	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Library Remove Resource Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library remove resource <ref>", func() {
		It("should remove a resource from the library", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource first
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Verify resource was added
			Expect(filepath.Join(libPath, "skills", "commit.md")).To(BeAnExistingFile())

			// Remove the resource
			session = cli.Run("library", "remove", "resource", "skill/commit", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Removed resource: skill/commit")

			// Verify file was deleted
			Expect(filepath.Join(libPath, "skills", "commit.md")).NotTo(BeAnExistingFile())

			// Verify library.yaml no longer has the resource
			session = cli.Run("library", "resources", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "No resources")
		})
	})

	Describe("library remove resource with --json flag", func() {
		It("should output JSON on success", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource first
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Remove with JSON flag
			session = cli.Run("library", "remove", "resource", "skill/commit", "--library", libPath, "--json")
			cli.ShouldSucceed(session)

			// Verify JSON output contains expected fields
			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring(`"type":"resource"`))
			Expect(output).To(ContainSubstring(`"resourceType":"skill"`))
			Expect(output).To(ContainSubstring(`"name":"commit"`))
			Expect(output).To(ContainSubstring(`"fileDeleted"`))
			Expect(output).To(ContainSubstring(`"libraryPath"`))
		})
	})

	Describe("library remove resource error on nonexistent resource", func() {
		It("should error with helpful message", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Try to remove nonexistent resource - exit code 6 is NotFound
			session = cli.Run("library", "remove", "resource", "skill/nonexistent", "--library", libPath)
			cli.ShouldFailWithExit(session, 6)
			cli.ShouldErrorOutput(session, "not found")
		})
	})

	Describe("library remove resource error when referenced by preset", func() {
		It("should error with helpful message", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource first
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Create a preset that references the resource
			session = cli.Run("library", "create", "preset", "test-preset",
				"--resources", "skill/commit", "--library", libPath)
			cli.ShouldSucceed(session)

			// Try to remove the resource - should fail
			session = cli.Run("library", "remove", "resource", "skill/commit", "--library", libPath)
			cli.ShouldFailWithExit(session, 1)
			cli.ShouldErrorOutput(session, "referenced by preset")
		})
	})

	Describe("library remove resource with --library flag", func() {
		It("should use specified library path", func() {
			// Create a library at specific path
			libPath := filepath.Join(tmpDir, "custom-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Remove using --library flag
			session = cli.Run("library", "remove", "resource", "skill/commit", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Removed resource: skill/commit")

			// Verify file was deleted
			Expect(filepath.Join(libPath, "skills", "commit.md")).NotTo(BeAnExistingFile())
		})
	})
})

var _ = Describe("Library Remove Preset Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library remove preset <name>", func() {
		It("should remove a preset from the library", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource first
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Create a preset
			session = cli.Run("library", "create", "preset", "workflow",
				"--resources", "skill/commit", "--description", "Git workflow", "--library", libPath)
			cli.ShouldSucceed(session)

			// Verify preset was created
			session = cli.Run("library", "presets", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "workflow")

			// Remove the preset
			session = cli.Run("library", "remove", "preset", "workflow", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Removed preset: workflow")

			// Verify preset no longer exists
			session = cli.Run("library", "presets", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "No presets")

			// Verify resource still exists (preset removal doesn't delete resources)
			Expect(filepath.Join(libPath, "skills", "commit.md")).To(BeAnExistingFile())
		})
	})

	Describe("library remove preset with --json flag", func() {
		It("should output JSON on success", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource first
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Create a preset
			session = cli.Run("library", "create", "preset", "workflow",
				"--resources", "skill/commit", "--library", libPath)
			cli.ShouldSucceed(session)

			// Remove with JSON flag
			session = cli.Run("library", "remove", "preset", "workflow", "--library", libPath, "--json")
			cli.ShouldSucceed(session)

			// Verify JSON output contains expected fields
			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring(`"type":"preset"`))
			Expect(output).To(ContainSubstring(`"name":"workflow"`))
			Expect(output).To(ContainSubstring(`"resourcesRemoved"`))
			Expect(output).To(ContainSubstring(`"skill/commit"`))
		})
	})

	Describe("library remove preset error on nonexistent preset", func() {
		It("should error with helpful message", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "test-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Try to remove nonexistent preset - exit code 6 is NotFound
			session = cli.Run("library", "remove", "preset", "nonexistent", "--library", libPath)
			cli.ShouldFailWithExit(session, 6)
			cli.ShouldErrorOutput(session, "not found")
		})
	})

	Describe("library remove preset with --library flag", func() {
		It("should use specified library path", func() {
			// Create a library at specific path
			libPath := filepath.Join(tmpDir, "custom-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Create a preset
			session = cli.Run("library", "create", "preset", "workflow",
				"--resources", "skill/commit", "--library", libPath)
			cli.ShouldSucceed(session)

			// Remove using --library flag
			session = cli.Run("library", "remove", "preset", "workflow", "--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Removed preset: workflow")
		})
	})

	Describe("library remove resource with GERMINATOR_LIBRARY env", func() {
		It("should use library from environment variable", func() {
			// Create a library
			libPath := filepath.Join(tmpDir, "env-library")
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Add a resource using --library
			sourcePath := filepath.Join(fixtures.LibraryDir(), "skills", "skill-commit.md")
			session = cli.Run("library", "add", sourcePath, "--library", libPath)
			cli.ShouldSucceed(session)

			// Remove using GERMINATOR_LIBRARY env
			session = cli.RunWithEnv(map[string]string{"GERMINATOR_LIBRARY": libPath},
				"library", "remove", "resource", "skill/commit")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Removed resource: skill/commit")
		})
	})
})

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

var _ = Describe("Library Init Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library init --path", func() {
		It("should create library at specified path", func() {
			libPath := filepath.Join(tmpDir, "my-library")

			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Verify library structure was created
			Expect(libPath).To(BeAnExistingFile())
			Expect(filepath.Join(libPath, "library.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(libPath, "skills")).To(BeAnExistingFile())
			Expect(filepath.Join(libPath, "agents")).To(BeAnExistingFile())
			Expect(filepath.Join(libPath, "commands")).To(BeAnExistingFile())
			Expect(filepath.Join(libPath, "memory")).To(BeAnExistingFile())
		})
	})

	Describe("library init --dry-run", func() {
		It("should preview without creating files", func() {
			libPath := filepath.Join(tmpDir, "dry-run-library")

			session := cli.Run("library", "init", "--path", libPath, "--dry-run")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Would create library at:")
			cli.ShouldOutput(session, "library.yaml")

			// Verify nothing was created
			Expect(libPath).NotTo(BeAnExistingFile())
		})
	})

	Describe("library init --force", func() {
		It("should overwrite existing library", func() {
			libPath := filepath.Join(tmpDir, "force-library")

			// Create initial library
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Force overwrite
			session = cli.Run("library", "init", "--path", libPath, "--force")
			cli.ShouldSucceed(session)

			// Verify library still exists
			Expect(libPath).To(BeAnExistingFile())
			Expect(filepath.Join(libPath, "library.yaml")).To(BeAnExistingFile())
		})
	})

	Describe("library init with invalid path", func() {
		It("should return appropriate error", func() {
			// Use a path that should fail (protected directory)
			// On Unix, /protected is typically not writable by normal users
			libPath := filepath.Join("/protected", "library")

			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("library init without permissions", func() {
		It("should return appropriate error", func() {
			// Create a file where a directory should be
			fileAsDir := filepath.Join(tmpDir, "file")
			f, err := os.Create(fileAsDir)
			Expect(err).NotTo(HaveOccurred())
			f.Close()

			session := cli.Run("library", "init", "--path", fileAsDir)
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("created library is loadable", func() {
		It("should be loadable by library resources command", func() {
			libPath := filepath.Join(tmpDir, "loadable-library")

			// Create library
			session := cli.Run("library", "init", "--path", libPath)
			cli.ShouldSucceed(session)

			// Try to load it using library resources command
			session = cli.Run("library", "resources", "--library", libPath)
			cli.ShouldSucceed(session)
		})
	})
})

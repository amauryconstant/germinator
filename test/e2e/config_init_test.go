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

var _ = Describe("Config Init Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("config init at custom path", func() {
		It("should create config file with documented fields", func() {
			dest := filepath.Join(tmpDir, "config.toml")
			session := cli.Run("config", "init", "--output-path", dest)
			cli.ShouldSucceed(session)

			Expect(dest).To(BeAnExistingFile())
			content, err := os.ReadFile(dest)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("# Germinator configuration"))
			Expect(string(content)).To(ContainSubstring("[completion]"))
		})
	})

	Describe("config init default path", func() {
		It("should write to $XDG_CONFIG_HOME when set", func() {
			// Use RunWithEnv so only the subprocess sees XDG_CONFIG_HOME;
			// the developer's real home directory is never touched.
			expectedPath := filepath.Join(tmpDir, "germinator", "config.toml")

			session := cli.RunWithEnv(
				map[string]string{"XDG_CONFIG_HOME": tmpDir},
				"config", "init",
			)
			cli.ShouldSucceed(session)

			Expect(expectedPath).To(BeAnExistingFile())
			content, err := os.ReadFile(expectedPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("# Germinator configuration"))
		})
	})

	Describe("config init --force", func() {
		It("should overwrite an existing config file", func() {
			dest := filepath.Join(tmpDir, "config.toml")

			// First write succeeds
			session := cli.Run("config", "init", "--output-path", dest)
			cli.ShouldSucceed(session)

			// Pre-corrupt the file
			Expect(os.WriteFile(dest, []byte("stale content"), 0o644)).To(Succeed())

			// Second write without --force fails
			session = cli.Run("config", "init", "--output-path", dest)
			cli.ShouldFailWithExit(session, 1)

			// Third write with --force succeeds and overwrites
			session = cli.Run("config", "init", "--output-path", dest, "--force")
			cli.ShouldSucceed(session)

			content, err := os.ReadFile(dest)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("# Germinator configuration"))
			Expect(string(content)).NotTo(ContainSubstring("stale content"))
		})
	})

	Describe("config init creates parent directories", func() {
		It("should create nested parent directories", func() {
			dest := filepath.Join(tmpDir, "nested", "deep", "config.toml")
			session := cli.Run("config", "init", "--output-path", dest)
			cli.ShouldSucceed(session)
			Expect(dest).To(BeAnExistingFile())
		})
	})

	Describe("config init rejects legacy --output flag", func() {
		It("should fail with usage error (exit 2)", func() {
			dest := filepath.Join(tmpDir, "config.toml")
			session := cli.Run("config", "init", "--output", dest)
			cli.ShouldFailWithExit(session, 2)
			// Cobra's unknown-flag message should mention the flag name in stderr.
			cli.ShouldErrorOutput(session, "--output")
		})
	})
})

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

var _ = Describe("Init Command", func() {
	var cli *helpers.GerminatorCLI
	var tmpDir string

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = GinkgoT().TempDir()
	})

	libraryEnv := func() map[string]string {
		return map[string]string{"GERMINATOR_LIBRARY": fixtures.LibraryDir()}
	}

	Describe("init with dry-run", func() {
		It("should preview changes without writing files", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--resources", "skill/commit",
				"--output", tmpDir, "--dry-run",
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Dry run complete")
			cli.ShouldOutput(session, "skill-commit.md")
			Expect(filepath.Join(tmpDir, ".opencode", "skills", "commit", "SKILL.md")).NotTo(BeAnExistingFile())
		})
	})

	Describe("init with force overwrite", func() {
		It("should overwrite existing files", func() {
			outputPath := filepath.Join(tmpDir, ".opencode", "skills", "commit", "SKILL.md")
			Expect(os.MkdirAll(filepath.Dir(outputPath), 0755)).To(Succeed())
			Expect(os.WriteFile(outputPath, []byte("old content"), 0644)).To(Succeed())

			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--resources", "skill/commit",
				"--output", tmpDir, "--force",
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Successfully installed")

			content, err := os.ReadFile(outputPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).NotTo(ContainSubstring("old content"))
		})
	})

	Describe("init fails without force when file exists", func() {
		It("should return error when file exists and force not set", func() {
			outputPath := filepath.Join(tmpDir, ".opencode", "skills", "commit", "SKILL.md")
			Expect(os.MkdirAll(filepath.Dir(outputPath), 0755)).To(Succeed())
			Expect(os.WriteFile(outputPath, []byte("existing content"), 0644)).To(Succeed())

			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--resources", "skill/commit",
				"--output", tmpDir,
			)
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "exists")
		})
	})

	Describe("init with preset", func() {
		It("should expand preset and install all resources", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--preset", "git-workflow",
				"--output", tmpDir,
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Successfully installed 2 resource(s)")

			Expect(filepath.Join(tmpDir, ".opencode", "skills", "commit", "SKILL.md")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, ".opencode", "skills", "merge-request", "SKILL.md")).To(BeAnExistingFile())
		})
	})

	Describe("init fails for nonexistent resource", func() {
		It("should return error for invalid resource reference", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--resources", "skill/nonexistent",
				"--output", tmpDir,
			)
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "nonexistent")
		})
	})

	Describe("init fails for nonexistent preset", func() {
		It("should return error for invalid preset name", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--preset", "nonexistent-preset",
				"--output", tmpDir,
			)
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "nonexistent-preset")
		})
	})

	Describe("init requires platform flag", func() {
		It("should fail with exit code 1 when platform not specified", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--resources", "skill/commit",
				"--output", tmpDir,
			)
			cli.ShouldFailWithExit(session, 1)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("required"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("init requires resources or preset", func() {
		It("should fail with exit code 3 when neither specified", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode",
				"--output", tmpDir,
			)
			cli.ShouldFailWithExit(session, 3)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("resources"),
				ContainSubstring("preset"),
				ContainSubstring("required"),
			))
		})
	})

	Describe("init rejects mutually exclusive flags", func() {
		It("should fail with exit code 3 when both resources and preset specified", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode",
				"--resources", "skill/commit",
				"--preset", "git-workflow",
				"--output", tmpDir,
			)
			cli.ShouldFailWithExit(session, 3)
			cli.ShouldErrorOutput(session, "mutually exclusive")
		})
	})

	Describe("init fails for invalid platform", func() {
		It("should fail with exit code 3 for unknown platform", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "invalid-platform",
				"--resources", "skill/commit",
				"--output", tmpDir,
			)
			cli.ShouldFailWithExit(session, 3)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("unknown"),
				ContainSubstring("invalid"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("init succeeds with claude-code platform", func() {
		It("should install resources to .claude directory", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "claude-code",
				"--resources", "skill/commit",
				"--output", tmpDir,
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Successfully installed")
			Expect(filepath.Join(tmpDir, ".claude", "skills", "commit", "SKILL.md")).To(BeAnExistingFile())
		})
	})
})

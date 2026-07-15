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
				"--output-dir", tmpDir, "--dry-run",
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
				"--output-dir", tmpDir, "--force",
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Initialized 1 resource(s)")

			content, err := os.ReadFile(outputPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).NotTo(ContainSubstring("old content"))
		})
	})

	Describe("init fails without force when file exists", func() {
		It("should fail with exit code 1 when file exists and force not set", func() {
			outputPath := filepath.Join(tmpDir, ".opencode", "skills", "commit", "SKILL.md")
			Expect(os.MkdirAll(filepath.Dir(outputPath), 0755)).To(Succeed())
			Expect(os.WriteFile(outputPath, []byte("existing content"), 0644)).To(Succeed())

			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--resources", "skill/commit",
				"--output-dir", tmpDir,
			)
			// Per the slice-5 contract, the per-resource failure aggregates
			// as *core.PartialSuccessError{Succeeded: 0, Failed: 1}; the
			// 0/1 bucket maps to ExitCodeError (1) via cmdutil.ExitCodeFor.
			cli.ShouldFailWithExit(session, 1)
			cli.ShouldErrorOutput(session, "exists")
		})
	})

	Describe("init with preset", func() {
		It("should expand preset and install all resources", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--preset", "git-workflow",
				"--output-dir", tmpDir,
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Initialized 2 resource(s)")

			Expect(filepath.Join(tmpDir, ".opencode", "skills", "commit", "SKILL.md")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, ".opencode", "skills", "merge-request", "SKILL.md")).To(BeAnExistingFile())
		})
	})

	Describe("init fails for nonexistent resource", func() {
		It("should fail with exit code 1 for invalid resource reference", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--resources", "skill/nonexistent",
				"--output-dir", tmpDir,
			)
			// Per the slice-5 contract, the per-resource failure aggregates
			// as *core.PartialSuccessError{Succeeded: 0, Failed: 1} which
			// maps to ExitCodeError (1) via cmdutil.ExitCodeFor.
			cli.ShouldFailWithExit(session, 1)
			cli.ShouldErrorOutput(session, "nonexistent")
		})
	})

	Describe("init fails for nonexistent preset", func() {
		It("should fail with exit code 1 for invalid preset name (NotFoundError → ExitCodeError)", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode", "--preset", "nonexistent-preset",
				"--output-dir", tmpDir,
			)
			// Per enforce-error-discipline (Phase 1.1): --preset <missing>
			// returns *core.NotFoundError; cmdutil.ExitCodeFor maps this
			// to ExitCodeError (1) — a runtime lookup miss is an
			// operational error, not a user-input validation error.
			cli.ShouldFailWithExit(session, 1)
			cli.ShouldErrorOutput(session, "nonexistent-preset")
		})
	})

	Describe("init requires platform flag", func() {
		It("should fail with exit code 1 when platform not specified (Cobra required-flag falls through to default)", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--resources", "skill/commit",
				"--output-dir", tmpDir,
			)
			// Per enforce-error-discipline (Phase 1.2): the cobra
			// substring-prefix dispatch fallback was dropped. Cobra's
			// `MarkFlagRequired`-emitted `required flag(s) "..." not set`
			// string is not a typed pflag error and not yet wrapped in
			// *core.CobraUsageError (zero current call sites per task
			// 3.14), so it falls through to ExitCodeError (1).
			cli.ShouldFailWithExit(session, 1)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("required"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("init requires resources or preset", func() {
		It("should fail with exit code 1 when neither specified", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode",
				"--output-dir", tmpDir,
			)
			cli.ShouldFailWithExit(session, 1)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("resources"),
				ContainSubstring("preset"),
				ContainSubstring("required"),
			))
		})
	})

	Describe("init rejects mutually exclusive flags", func() {
		It("should fail with exit code 1 when both resources and preset specified", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "opencode",
				"--resources", "skill/commit",
				"--preset", "git-workflow",
				"--output-dir", tmpDir,
			)
			cli.ShouldFailWithExit(session, 1)
			cli.ShouldErrorOutput(session, "mutually exclusive")
		})
	})

	Describe("init fails for invalid platform", func() {
		It("should fail with exit code 1 for unknown platform", func() {
			session := cli.RunWithEnv(libraryEnv(),
				"init", "--platform", "invalid-platform",
				"--resources", "skill/commit",
				"--output-dir", tmpDir,
			)
			cli.ShouldFailWithExit(session, 1)
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
				"--output-dir", tmpDir,
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Initialized 1 resource(s)")
			Expect(filepath.Join(tmpDir, ".claude", "skills", "commit", "SKILL.md")).To(BeAnExistingFile())
		})
	})
})

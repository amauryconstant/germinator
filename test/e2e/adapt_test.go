//go:build e2e

package e2e_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/fixtures"
	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Adapt Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("adapting a valid document", func() {
		var outputPath string

		BeforeEach(func() {
			var err error
			outputPath, err = fixtures.TempOutputFile("adapt-test")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			// Cleanup output file
			os.Remove(outputPath)
		})

		It("should succeed with exit code 0, create output file, and display success message", func() {
			session := cli.Run("adapt", fixtures.ValidDocument(), outputPath, "--platform", "opencode")
			cli.ShouldSucceed(session)
			// Slice-2 changed runAdapt's success message from
			// "transformed successfully" to "wrote <output path>"
			// (see tasks.md task 2.2.4).
			cli.ShouldOutput(session, "wrote ")
			Expect(fixtures.FileExists(outputPath)).To(BeTrue(), "Output file should be created")
		})
	})

	Describe("adapting without platform flag", func() {
		It("should fail with exit code 2 and show required flag error", func() {
			outputPath, err := fixtures.TempOutputFile("adapt-no-platform")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("adapt", fixtures.ValidDocument(), outputPath)
			// Cobra's MarkFlagRequired enforcement maps to ExitCodeUsage (2)
			// via internal/cmdutil/exit.go cobraUsagePrefixes.
			cli.ShouldFailWithExit(session, 2)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("required"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("adapting a nonexistent file", func() {
		It("should fail with exit code > 0 and show file error", func() {
			outputPath, err := fixtures.TempOutputFile("adapt-nonexistent")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("adapt", fixtures.NonexistentFile(), outputPath, "--platform", "opencode")
			// CLI returns non-zero for file errors. Slice-2 changed the
			// error prefix to capital "Error:" via output.FormatError.
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "Error")
		})
	})

	Describe("adapting a valid document with claude-code platform", func() {
		var outputPath string

		BeforeEach(func() {
			var err error
			outputPath, err = fixtures.TempOutputFile("adapt-claude-code")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(outputPath)
		})

		It("should succeed with exit code 0, create output file, and display success message", func() {
			session := cli.Run("adapt", fixtures.ValidDocument(), outputPath, "--platform", "claude-code")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "wrote ")
			Expect(fixtures.FileExists(outputPath)).To(BeTrue(), "Output file should be created")
		})
	})

	Describe("adapting a nonexistent file with claude-code platform", func() {
		It("should fail with exit code > 0 and show file error", func() {
			outputPath, err := fixtures.TempOutputFile("adapt-nonexistent-claude-code")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("adapt", fixtures.NonexistentFile(), outputPath, "--platform", "claude-code")
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "Error")
		})
	})
})

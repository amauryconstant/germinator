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
			cli.ShouldOutput(session, "transformed successfully")
			Expect(fixtures.FileExists(outputPath)).To(BeTrue(), "Output file should be created")
		})
	})

	Describe("adapting without platform flag", func() {
		It("should fail with exit code 1 and show required flag error", func() {
			outputPath, err := fixtures.TempOutputFile("adapt-no-platform")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("adapt", fixtures.ValidDocument(), outputPath)
			cli.ShouldFailWithExit(session, 1)
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
			// CLI returns exit code 3 for file/parse errors
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "error")
		})
	})
})

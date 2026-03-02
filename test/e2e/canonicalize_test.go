//go:build e2e

package e2e_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/fixtures"
	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Canonicalize Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("canonicalizing a valid document", func() {
		var outputPath string

		BeforeEach(func() {
			var err error
			outputPath, err = fixtures.TempOutputFile("canonicalize-test")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(outputPath)
		})

		It("should succeed with exit code 0, create output file, and display success message for opencode platform", func() {
			session := cli.Run("canonicalize", fixtures.ValidDocument(), outputPath, "--platform", "opencode", "--type", "agent")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Successfully canonicalized")
			Expect(fixtures.FileExists(outputPath)).To(BeTrue(), "Output file should be created")
		})

		It("should succeed with exit code 0, create output file, and display success message for claude-code platform", func() {
			session := cli.Run("canonicalize", fixtures.ValidDocument(), outputPath, "--platform", "claude-code", "--type", "agent")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Successfully canonicalized")
			Expect(fixtures.FileExists(outputPath)).To(BeTrue(), "Output file should be created")
		})
	})

	Describe("canonicalizing without platform flag", func() {
		It("should fail with exit code 1 and show required flag error", func() {
			outputPath, err := fixtures.TempOutputFile("canonicalize-no-platform")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("canonicalize", fixtures.ValidDocument(), outputPath, "--type", "agent")
			cli.ShouldFailWithExit(session, 1)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("required"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("canonicalizing without type flag", func() {
		It("should fail with exit code 1 and show required flag error", func() {
			outputPath, err := fixtures.TempOutputFile("canonicalize-no-type")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("canonicalize", fixtures.ValidDocument(), outputPath, "--platform", "opencode")
			cli.ShouldFailWithExit(session, 1)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("required"),
				ContainSubstring("type"),
			))
		})
	})

	Describe("canonicalizing with an invalid platform", func() {
		It("should fail with exit code > 0 and show invalid platform error", func() {
			outputPath, err := fixtures.TempOutputFile("canonicalize-invalid-platform")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("canonicalize", fixtures.ValidDocument(), outputPath, "--platform", "invalid-platform", "--type", "agent")
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("unknown"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("canonicalizing with an invalid type", func() {
		It("should fail with exit code > 0 and show invalid type error", func() {
			outputPath, err := fixtures.TempOutputFile("canonicalize-invalid-type")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("canonicalize", fixtures.ValidDocument(), outputPath, "--platform", "opencode", "--type", "invalid-type")
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("unknown"),
				ContainSubstring("type"),
			))
		})
	})

	Describe("canonicalizing a nonexistent file", func() {
		It("should fail with exit code > 0 and show file error", func() {
			outputPath, err := fixtures.TempOutputFile("canonicalize-nonexistent")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(outputPath)

			session := cli.Run("canonicalize", fixtures.NonexistentFile(), outputPath, "--platform", "opencode", "--type", "agent")
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "error")
		})
	})
})

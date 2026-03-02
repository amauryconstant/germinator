//go:build e2e

package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/fixtures"
	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Validate Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("validating a valid document", func() {
		It("should succeed with exit code 0 and display success message", func() {
			session := cli.Run("validate", fixtures.ValidDocument(), "--platform", "opencode")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Document is valid")
		})
	})

	Describe("validating without platform flag", func() {
		It("should fail with exit code 1 and show required flag error", func() {
			session := cli.Run("validate", fixtures.ValidDocument())
			cli.ShouldFailWithExit(session, 1)
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("required"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("validating a nonexistent file", func() {
		It("should fail with exit code > 0 and show file error", func() {
			session := cli.Run("validate", fixtures.NonexistentFile(), "--platform", "opencode")
			// CLI returns exit code 3 for file/parse errors
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "error")
		})
	})

	Describe("validating with an invalid platform", func() {
		It("should fail with exit code > 0 and show invalid platform error", func() {
			session := cli.Run("validate", fixtures.ValidDocument(), "--platform", "invalid-platform")
			// CLI returns exit code 2 for config errors
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			output := cli.GetErrorOutput(session)
			Expect(output).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("unknown"),
				ContainSubstring("platform"),
			))
		})
	})

	Describe("validating an invalid document", func() {
		It("should fail with exit code > 0 and show validation errors", func() {
			session := cli.Run("validate", fixtures.InvalidDocument(), "--platform", "opencode")
			// CLI returns exit code 2 for validation errors
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "Error")
		})
	})
})

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
		It("should fail with exit code 2 and show required flag error", func() {
			session := cli.Run("validate", fixtures.ValidDocument())
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

	Describe("validating a nonexistent file", func() {
		It("should fail with exit code > 0 and show file error", func() {
			session := cli.Run("validate", fixtures.NonexistentFile(), "--platform", "opencode")
			// Parse/file errors emit an ExitCodeError (1) via cmdutil.ExitCodeFor.
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			// Stderr now contains only output.FormatError output (Cobra's
			// usage block is suppressed via cmd/root.go SilenceUsage).
			cli.ShouldErrorOutput(session, "Error")
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
			// CLI returns non-zero exit for validation errors.
			// Slice-2 changed the error prefix to capital "Error:" via
			// output.FormatError; previously the message used lowercase
			// "validation error" which the test no longer matches.
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "Error")
		})
	})

	Describe("validating a valid document with claude-code platform", func() {
		It("should succeed with exit code 0 and display success message", func() {
			session := cli.Run("validate", fixtures.ValidDocument(), "--platform", "claude-code")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Document is valid")
		})
	})

	Describe("validating a nonexistent file with claude-code platform", func() {
		It("should fail with exit code > 0 and show file error", func() {
			session := cli.Run("validate", fixtures.NonexistentFile(), "--platform", "claude-code")
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "Error")
		})
	})

	Describe("validating an invalid document with claude-code platform", func() {
		It("should fail with exit code > 0 and show validation errors", func() {
			session := cli.Run("validate", fixtures.InvalidDocument(), "--platform", "claude-code")
			Expect(session.ExitCode()).To(BeNumerically(">", 0))
			cli.ShouldErrorOutput(session, "Error")
		})
	})
})

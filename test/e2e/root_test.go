//go:build e2e

package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Root Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("running root command without arguments", func() {
		It("should succeed with exit code 0 and display help/usage information", func() {
			session := cli.Run()
			cli.ShouldSucceed(session)
			output := cli.GetOutput(session)
			// Should show usage information
			Expect(output).To(Or(
				ContainSubstring("Usage:"),
				ContainSubstring("germinator"),
				ContainSubstring("Available Commands"),
			))
		})
	})

	Describe("running with --help flag", func() {
		It("should succeed with exit code 0 and display help information", func() {
			session := cli.Run("--help")
			cli.ShouldSucceed(session)
			output := cli.GetOutput(session)
			// Should show usage information
			Expect(output).To(Or(
				ContainSubstring("Usage:"),
				ContainSubstring("germinator"),
				ContainSubstring("Available Commands"),
			))
		})
	})
})

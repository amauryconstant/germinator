//go:build e2e

package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

// supportedShells mirrors cmd.completionShells and is duplicated here
// so the E2E test is self-describing. If a shell is added to the
// completion command, add it here too.
var supportedShells = []string{
	"bash", "zsh", "fish", "powershell", "elvish",
	"nushell", "oil", "tcsh", "xonsh", "cmd-clink",
}

var _ = Describe("Completion Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("listing supported shells", func() {
		It("should list every supported shell in --help", func() {
			session := cli.Run("completion", "--help")
			cli.ShouldSucceed(session)
			out := string(session.Out.Contents())
			for _, shell := range supportedShells {
				Expect(out).To(ContainSubstring(shell),
					"completion --help MUST list the %q subcommand", shell)
			}
		})
	})

	Describe("generating shell snippets", func() {
		It("should produce a non-empty script for each major shell", func() {
			for _, shell := range []string{"bash", "zsh", "fish"} {
				session := cli.Run("completion", shell)
				cli.ShouldSucceed(session)
				Expect(session.Out.Contents()).ToNot(BeEmpty(),
					"completion %s MUST produce a non-empty snippet", shell)
			}
		})
	})
})

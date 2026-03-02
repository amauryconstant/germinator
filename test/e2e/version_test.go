//go:build e2e

package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"

	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Version Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("displaying version info", func() {
		It("should succeed with exit code 0 and display version info", func() {
			session := cli.Run("version")
			cli.ShouldSucceed(session)
			// Version format: germinator <version> (<commit>) <date>
			// When built with gexec (no ldflags): germinator dev ()
			// When built with mise: germinator v0.5.0-20-g463ae72-dirty (463ae72...) 2026-03-02
			// We just check that it starts with "germinator" and contains some version info
			cli.ShouldOutput(session, "germinator")
		})
	})
})

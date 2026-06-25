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

var _ = Describe("library resources --output", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	readFixture := func(format string) string {
		path := filepath.Join("fixtures", "library-resources", format, "output."+formatExt(format))
		data, err := os.ReadFile(path) //nolint:gosec // G304: test reads fixture
		Expect(err).NotTo(HaveOccurred(), "read fixture %s", path)
		return string(data)
	}

	Describe("plain format (default)", func() {
		It("matches the pre-change plain output byte-for-byte", func() {
			session := cli.Run("library", "resources", "--library", fixtures.LibraryDir())
			cli.ShouldSucceed(session)
			Expect(string(session.Out.Contents())).To(Equal(readFixture("plain")),
				"plain output must be byte-identical to the pre-change build")
		})
	})

	Describe("--output json", func() {
		It("emits JSON with the documented structure", func() {
			session := cli.Run("library", "resources", "--library", fixtures.LibraryDir(), "--output", "json")
			cli.ShouldSucceed(session)
			Expect(string(session.Out.Contents())).To(Equal(readFixture("json")))
		})
	})

	Describe("--output table", func() {
		It("emits a tab-aligned table with the documented columns", func() {
			session := cli.Run("library", "resources", "--library", fixtures.LibraryDir(), "--output", "table")
			cli.ShouldSucceed(session)
			Expect(string(session.Out.Contents())).To(Equal(readFixture("table")))
		})
	})

	Describe("--json flag is rejected", func() {
		It("returns exit code 2 (ExitCodeUsage)", func() {
			session := cli.Run("library", "resources", "--library", fixtures.LibraryDir(), "--json")
			cli.ShouldFailWithExit(session, 2)
		})
	})
})

func formatExt(format string) string {
	switch format {
	case "plain", "table":
		return "txt"
	case "json":
		return "json"
	}
	return "txt"
}

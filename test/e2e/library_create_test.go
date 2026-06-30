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

// seedLibrary initialises a fresh library under tmpDir/test-library and
// imports the canonical skill-commit, agent-reviewer, and command-build
// fixtures so subsequent library create preset invocations have valid
// references to point at.
func seedLibrary(cli *helpers.GerminatorCLI, tmpDir string) string {
	libPath := filepath.Join(tmpDir, "test-library")
	session := cli.Run("library", "init", "--path", libPath)
	cli.ShouldSucceed(session)

	add := func(relFixture string) {
		s := cli.Run("library", "add",
			filepath.Join(fixtures.LibraryDir(), relFixture),
			"--library", libPath)
		cli.ShouldSucceed(s)
	}
	add(filepath.Join("skills", "skill-commit.md"))
	add(filepath.Join("agents", "agent-reviewer.md"))
	add(filepath.Join("commands", "command-test.md"))
	return libPath
}

var _ = Describe("Library Create Preset Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("library create preset <name> --resources <refs>", func() {
		It("creates a preset with a single resource reference", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "single",
				"--resources", "skill/commit",
				"--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Created preset: single")

			presetsSession := cli.Run("library", "presets", "--library", libPath)
			cli.ShouldSucceed(presetsSession)
			cli.ShouldOutput(presetsSession, "single")
		})

		It("creates a preset with multiple resource references", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "multi",
				"--resources", "skill/commit,agent/reviewer,command/test",
				"--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Created preset: multi")
		})

		It("creates a preset with --description", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "described",
				"--resources", "skill/commit,agent/reviewer",
				"--description", "Development setup",
				"--library", libPath)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Created preset: described")
		})

		It("overwrites an existing preset with --force", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "overwrite-target",
				"--resources", "skill/commit",
				"--library", libPath)
			cli.ShouldSucceed(session)

			session = cli.Run("library", "create", "preset", "overwrite-target",
				"--resources", "agent/reviewer",
				"--library", libPath)
			cli.ShouldFailWithExit(session, 1)

			session = cli.Run("library", "create", "preset", "overwrite-target",
				"--resources", "agent/reviewer",
				"--force",
				"--library", libPath)
			cli.ShouldSucceed(session)
		})
	})

	Describe("library create preset failure paths", func() {
		It("returns exit 2 when --resources is missing (Cobra required-flag)", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "missing-flags",
				"--library", libPath)
			cli.ShouldFailWithExit(session, 2)
		})

		It("returns exit 2 when --resources is an empty string", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "empty-resources",
				"--resources", "",
				"--library", libPath)
			cli.ShouldFailWithExit(session, 2)
		})

		It("returns exit 1 when a ref has an invalid type", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "bad-ref",
				"--resources", "skills/commit",
				"--library", libPath)
			cli.ShouldFailWithExit(session, 1)
		})

		It("returns exit 1 when a ref has an empty name", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "empty-name",
				"--resources", "skill/",
				"--library", libPath)
			cli.ShouldFailWithExit(session, 1)
		})

		It("returns exit 1 when the preset name argument is missing", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset",
				"--resources", "skill/commit",
				"--library", libPath)
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("library create preset help", func() {
		It("does not advertise an --output flag (Decision 5)", func() {
			session := cli.Run("library", "create", "preset", "--help")
			cli.ShouldSucceed(session)
			stdout := string(session.Out.Contents())
			Expect(stdout).NotTo(ContainSubstring("--output"),
				"library create preset must NOT expose --output (Decision 5)")
		})

		It("lists --resources as a flag of the preset command", func() {
			session := cli.Run("library", "create", "preset", "--help")
			cli.ShouldSucceed(session)
			stdout := string(session.Out.Contents())
			Expect(stdout).To(ContainSubstring("--resources"),
				"help must list --resources as a flag of the preset command")
		})
	})

	Describe("library create preset writes to a real library.yaml", func() {
		It("persists the preset so library presets shows it", func() {
			libPath := seedLibrary(cli, tmpDir)

			session := cli.Run("library", "create", "preset", "persisted",
				"--resources", "skill/commit,agent/reviewer",
				"--library", libPath)
			cli.ShouldSucceed(session)

			libraryYAML := filepath.Join(libPath, "library.yaml")
			contents, err := os.ReadFile(libraryYAML)
			Expect(err).NotTo(HaveOccurred(), "library.yaml must exist")
			Expect(string(contents)).To(ContainSubstring("persisted"))
		})
	})
})

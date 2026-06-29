//go:build e2e

package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"

	"gitlab.com/amoconst/germinator/test/e2e/fixtures"
	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

var _ = Describe("Library Command", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	Describe("listing library resources", func() {
		It("should display resources grouped by type", func() {
			session := cli.Run("library", "resources", "--library", fixtures.LibraryDir())
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Skills:")
			cli.ShouldOutput(session, "skill/commit")
			cli.ShouldOutput(session, "skill/merge-request")
			cli.ShouldOutput(session, "Agents:")
			cli.ShouldOutput(session, "agent/reviewer")
			cli.ShouldOutput(session, "Commands:")
			cli.ShouldOutput(session, "command/test")
			cli.ShouldOutput(session, "Memorys:")
			cli.ShouldOutput(session, "memory/context")
		})
	})

	Describe("listing library presets", func() {
		It("should display presets with descriptions", func() {
			session := cli.Run("library", "presets", "--library", fixtures.LibraryDir())
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "git-workflow")
			cli.ShouldOutput(session, "Git workflow tools")
			cli.ShouldOutput(session, "code-review")
			cli.ShouldOutput(session, "Code review tools")
		})
	})

	Describe("showing a resource", func() {
		It("should display resource details", func() {
			session := cli.Run("library", "show", "skill/commit", "--library", fixtures.LibraryDir())
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Reference: skill/commit")
			cli.ShouldOutput(session, "Path: skills/skill-commit.md")
			cli.ShouldOutput(session, "Description: Git commit best practices")
		})
	})

	Describe("showing a preset", func() {
		It("should display preset details with resources", func() {
			session := cli.Run("library", "show", "preset/git-workflow", "--library", fixtures.LibraryDir())
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Preset: git-workflow")
			cli.ShouldOutput(session, "Description: Git workflow tools")
			cli.ShouldOutput(session, "Resources:")
			cli.ShouldOutput(session, "skill/commit")
			cli.ShouldOutput(session, "skill/merge-request")
		})
	})

	Describe("showing a missing reference", func() {
		It("should fail with not-found error (ExitCodeUsage=2 per slice-5 §5.0.1)", func() {
			session := cli.Run("library", "show", "invalid-format", "--library", fixtures.LibraryDir())
			cli.ShouldFailWithExit(session, 2)
			cli.ShouldErrorOutput(session, "not found: invalid-format")
		})
	})

	Describe("using --library flag", func() {
		It("should use the specified library path", func() {
			session := cli.Run("library", "resources", "--library", fixtures.LibraryDir())
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "skill/commit")
		})
	})

	Describe("using GERMINATOR_LIBRARY environment variable", func() {
		It("should use the library path from environment", func() {
			session := cli.RunWithEnv(map[string]string{"GERMINATOR_LIBRARY": fixtures.LibraryDir()}, "library", "resources")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "skill/commit")
		})
	})
})

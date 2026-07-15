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
		It("should fail with not-found error (ExitCodeError=1 per enforce-error-discipline Phase 1.1)", func() {
			session := cli.Run("library", "show", "invalid-format", "--library", fixtures.LibraryDir())
			cli.ShouldFailWithExit(session, 1)
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

	// Task 7.3 E2E: Config.Library precedence. Writes a config file at
	// $XDG_CONFIG_HOME/germinator/config.toml with `library = "..."` and
	// asserts the resolved library matches the config-file value when
	// neither flag nor env var is set.
	Describe("using config file library setting", func() {
		It("should resolve library from XDG config when no flag or env is set", func() {
			xdgConfigHome := GinkgoT().TempDir()
			configDir := filepath.Join(xdgConfigHome, "germinator")
			Expect(os.MkdirAll(configDir, 0o750)).To(Succeed())
			configPath := filepath.Join(configDir, "config.toml")
			Expect(os.WriteFile(configPath, []byte(`library = "`+fixtures.LibraryDir()+`"`+"\n"), 0o600)).To(Succeed())

			session := cli.RunWithEnv(
				map[string]string{"XDG_CONFIG_HOME": xdgConfigHome},
				"library", "resources",
			)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "skill/commit")
		})
	})
})

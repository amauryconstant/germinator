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

var _ = Describe("Config Validate Command", func() {
	var (
		cli    *helpers.GerminatorCLI
		tmpDir string
	)

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
		tmpDir = fixtures.TempDir()
	})

	Describe("valid config", func() {
		It("should succeed and print valid message", func() {
			path := filepath.Join(tmpDir, "valid.toml")
			content := `library = "~/.config/germinator/library"
platform = "opencode"
[completion]
timeout = "500ms"
cache_ttl = "5s"
`
			Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())

			session := cli.Run("config", "validate", "--output-path", path)
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Config file is valid:")
		})
	})

	Describe("file not found", func() {
		It("should fail with exit 1", func() {
			path := filepath.Join(tmpDir, "nonexistent.toml")
			session := cli.Run("config", "validate", "--output-path", path)
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("invalid platform value", func() {
		It("should fail with exit 1", func() {
			path := filepath.Join(tmpDir, "invalid.toml")
			content := `platform = "unknown"
`
			Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())

			session := cli.Run("config", "validate", "--output-path", path)
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("malformed TOML", func() {
		It("should fail with exit 1", func() {
			path := filepath.Join(tmpDir, "malformed.toml")
			Expect(os.WriteFile(path, []byte("invalid [ ["), 0o644)).To(Succeed())

			session := cli.Run("config", "validate", "--output-path", path)
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("rejects legacy --output flag", func() {
		It("should fail with usage error (exit 2)", func() {
			path := filepath.Join(tmpDir, "any.toml")
			session := cli.Run("config", "validate", "--output", path)
			cli.ShouldFailWithExit(session, 2)
		})
	})
})

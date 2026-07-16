//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/amoconst/germinator/test/e2e/helpers"
)

// E2E byte-equality golden tests for the Claude Code adapter rendering
// pipeline. Mirrors the OpenCode golden tests at
// test/e2e/opencode_adapter_golden_test.go: same fixture path shape,
// same byte-equality comparison semantic, same `//go:build e2e` tag.
//
// Per design Decision 6, byte-equality tests live in the E2E tier
// because the renderer output is sensitive to dependency drift.
// Semantic-equality round-trip tests live in the default suite at
// internal/renderer/serializer_test.go (TestPlatformRoundTrip).

func repoRootFromCallerCC() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

var _ = Describe("claude-code adapter byte-equality rendering", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	It("renders the permission-balanced agent byte-equally", func() {
		repoRoot := repoRootFromCallerCC()
		inputPath := filepath.Join(repoRoot, "test", "fixtures", "canonical", "agent-permission-balanced.md")
		goldenPath := filepath.Join(repoRoot, "test", "e2e", "fixtures", "claude-code", "agent-balanced.md")
		outDir := GinkgoT().TempDir()
		outPath := filepath.Join(outDir, "agent-balanced.md")

		session := cli.Run("adapt", inputPath, outPath, "--platform", "claude-code")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "wrote "+outPath)

		gotBytes, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred(), "read rendered output %s", outPath)

		wantBytes, err := os.ReadFile(goldenPath)
		Expect(err).NotTo(HaveOccurred(), "read golden fixture %s", goldenPath)

		Expect(string(gotBytes)).To(Equal(string(wantBytes)),
			"claude-code render must be byte-identical to %s", goldenPath)
	})
})

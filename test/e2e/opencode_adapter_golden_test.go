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

// E2E byte-equality golden tests for the adapter rendering pipeline.
//
// These tests live under `//go:build e2e` and assert against the
// renderer output via the germinator CLI. They are byte-equality
// comparisons against golden fixtures (per design Decision 6): the
// output is sensitive to dependency drift, so they live in a separate
// CI stage from the default suite. Round-trip semantic-equality
// tests live in the default suite at
// internal/renderer/serializer_test.go (TestParseRenderRoundTrip,
// TestPlatformRoundTrip).
//
// The fixtures (test/e2e/fixtures/<platform>/agent-balanced.md) are
// the byte-identical output that `germinator adapt` produces for the
// canonical agent-permission-balanced fixture, captured per-platform.
//
// Path resolution uses runtime.Caller(0) so the test works regardless
// of the working directory `go test` is invoked from (the binary
// inherits the test process's CWD, which is the package directory
// `test/e2e/` — not the project root).

// repoRootFromCaller resolves the repository root via runtime.Caller(0)
// on the caller's source file. Each test file passes its own source
// file so the resolution is local to the file declaring the test.
func repoRootFromCaller() string {
	_, file, _, _ := runtime.Caller(1)
	// file is e.g. test/e2e/opencode_adapter_golden_test.go
	return filepath.Join(filepath.Dir(file), "..", "..")
}

var _ = Describe("opencode adapter byte-equality rendering", func() {
	var cli *helpers.GerminatorCLI

	BeforeEach(func() {
		cli = helpers.NewGerminatorCLI(BinaryPath())
	})

	It("renders the permission-balanced agent byte-equally", func() {
		repoRoot := repoRootFromCaller()
		inputPath := filepath.Join(repoRoot, "test", "fixtures", "canonical", "agent-permission-balanced.md")
		goldenPath := filepath.Join(repoRoot, "test", "e2e", "fixtures", "opencode", "agent-balanced.md")
		outDir := GinkgoT().TempDir()
		outPath := filepath.Join(outDir, "agent-balanced.md")

		session := cli.Run("adapt", inputPath, outPath, "--platform", "opencode")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "wrote "+outPath)

		gotBytes, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred(), "read rendered output %s", outPath)

		wantBytes, err := os.ReadFile(goldenPath)
		Expect(err).NotTo(HaveOccurred(), "read golden fixture %s", goldenPath)

		Expect(string(gotBytes)).To(Equal(string(wantBytes)),
			"opencode render must be byte-identical to %s", goldenPath)
	})
})

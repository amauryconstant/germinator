//go:build e2e

package e2e_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var (
	germinatorPath string
)

var _ = BeforeSuite(func() {
	var err error

	// Build the germinator binary for E2E tests
	By("Building germinator-e2e binary")
	germinatorPath, err = gexec.Build("gitlab.com/amoconst/germinator/cmd")
	Expect(err).NotTo(HaveOccurred(), "Failed to build germinator binary")
	Expect(germinatorPath).To(BeAnExistingFile())

	By("Germinator binary built successfully", func() {
		GinkgoLogr.Info("Binary path", "path", germinatorPath)
	})
})

var _ = AfterSuite(func() {
	By("Cleaning up build artifacts")
	gexec.CleanupBuildArtifacts()
})

// BinaryPath returns the path to the built germinator binary
func BinaryPath() string {
	return germinatorPath
}

// DefaultTimeout is the default timeout for CLI operations
const DefaultTimeout = 30 * time.Second

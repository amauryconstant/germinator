//go:build e2e

package helpers

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// GerminatorCLI provides utilities for running the germinator binary in tests
type GerminatorCLI struct {
	binaryPath string
	timeout    time.Duration
}

// NewGerminatorCLI creates a new CLI helper for the given binary path
func NewGerminatorCLI(binaryPath string) *GerminatorCLI {
	return &GerminatorCLI{
		binaryPath: binaryPath,
		timeout:    30 * time.Second,
	}
}

// SetTimeout configures the timeout for CLI operations
func (c *GerminatorCLI) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Run executes the germinator binary with the given arguments and returns the session
func (c *GerminatorCLI) Run(args ...string) *gexec.Session {
	By("Running germinator command", func() {
		GinkgoLogr.Info("Executing", "binary", c.binaryPath, "args", args)
	})

	command := exec.Command(c.binaryPath, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred(), "Failed to start germinator command")

	// Wait for the command to complete
	Eventually(session, c.timeout).Should(gexec.Exit(), "Command did not complete within timeout")

	return session
}

// ShouldSucceed asserts that the session exited with code 0
func (c *GerminatorCLI) ShouldSucceed(session *gexec.Session) {
	Expect(session.ExitCode()).To(Equal(0), "Expected exit code 0, got %d\nStdout: %s\nStderr: %s",
		session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
}

// ShouldFailWithExit asserts that the session exited with the specified code
func (c *GerminatorCLI) ShouldFailWithExit(session *gexec.Session, code int) {
	Expect(session.ExitCode()).To(Equal(code), "Expected exit code %d, got %d\nStdout: %s\nStderr: %s",
		code, session.ExitCode(), string(session.Out.Contents()), string(session.Err.Contents()))
}

// ShouldOutput asserts that stdout contains the expected string
func (c *GerminatorCLI) ShouldOutput(session *gexec.Session, expected string) {
	Expect(string(session.Out.Contents())).To(ContainSubstring(expected),
		"Expected stdout to contain %q\nActual stdout: %s\nStderr: %s",
		expected, string(session.Out.Contents()), string(session.Err.Contents()))
}

// ShouldOutputMatch asserts that stdout matches the expected regex pattern
func (c *GerminatorCLI) ShouldOutputMatch(session *gexec.Session, pattern string) {
	Expect(string(session.Out.Contents())).To(MatchRegexp(pattern),
		"Expected stdout to match pattern %q\nActual stdout: %s",
		pattern, string(session.Out.Contents()))
}

// ShouldErrorOutput asserts that stderr contains the expected string
func (c *GerminatorCLI) ShouldErrorOutput(session *gexec.Session, expected string) {
	Expect(string(session.Err.Contents())).To(ContainSubstring(expected),
		"Expected stderr to contain %q\nActual stderr: %s\nStdout: %s",
		expected, string(session.Err.Contents()), string(session.Out.Contents()))
}

// GetOutput returns the stdout content
func (c *GerminatorCLI) GetOutput(session *gexec.Session) string {
	return string(session.Out.Contents())
}

// GetErrorOutput returns the stderr content
func (c *GerminatorCLI) GetErrorOutput(session *gexec.Session) string {
	return string(session.Err.Contents())
}

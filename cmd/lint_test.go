package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var lintIssueRE = regexp.MustCompile(`^([a-zA-Z0-9_/\-]+\.go):(\d+):(\d+):\s*(.+)$`)

func TestLintBaseline(t *testing.T) {
	t.Parallel()

	baselinePath := filepath.Join("testdata", "lint_baseline.txt")
	baseline, err := readFile(baselinePath)
	require.NoError(t, err, "lint baseline file is required at %s", baselinePath)

	current := collectLintIssues(t, 8)
	baselineIssues := extractLintIssues(baseline)
	currentIssues := extractLintIssues(current)

	if !bytes.Equal(baselineIssues, currentIssues) {
		t.Logf("lint output differs from baseline. Run `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` to update the baseline if the new violations are intentional.")
		t.Logf("Baseline issues:\n%s\n\nCurrent issues:\n%s", baselineIssues, currentIssues)
		t.Fail()
	}
}

// collectLintIssues runs mise run lint several times and returns the
// union of all issue lines seen across runs. This compensates for the
// non-deterministic reporting order of golangci-lint's parallel
// linters: the union of stable violations is what we want to gate on.
func collectLintIssues(t *testing.T, runs int) []byte {
	t.Helper()
	seen := map[string]struct{}{}
	for i := 0; i < runs; i++ {
		cmd := exec.Command("mise", "run", "lint")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		_ = cmd.Run()
		for _, issue := range extractIssueLines(buf.Bytes()) {
			seen[issue] = struct{}{}
		}
	}
	var lines []string
	for line := range seen {
		lines = append(lines, line)
	}
	sort.Strings(lines)
	return []byte(strings.Join(lines, "\n"))
}

// extractLintIssues parses the lint output and returns a sorted list
// of issue lines (one per violation) suitable for stable comparison.
func extractLintIssues(in []byte) []byte {
	lines := extractIssueLines(in)
	sort.Strings(lines)
	return []byte(strings.Join(lines, "\n"))
}

func extractIssueLines(in []byte) []string {
	var issues []string
	for _, line := range strings.Split(string(in), "\n") {
		trimmed := strings.TrimPrefix(line, "[lint] ")
		if !lintIssueRE.MatchString(trimmed) {
			continue
		}
		issues = append(issues, trimmed)
	}
	return issues
}

func readFile(p string) ([]byte, error) {
	return readFileImpl(p)
}

// TestNoNewForbidigoPatterns is a regression smoke test for the slice-2
// migration: ensure the migrated pilot commands (cmd/adapt.go and
// cmd/resources.go) do not re-introduce forbidden patterns like
// `fmt.Fprintf(os.Stdout|Stderr)` or `os.Exit(`. The check is
// grep-based so it does not require a full golangci-lint run.
//
// If a new intentional pattern is added (e.g., a debug print during
// refactoring), update this test alongside the change. It is NOT a
// replacement for the lint baseline test above; it complements it by
// catching the specific patterns that the forbidigo linter enforces
// for the pilot commands without depending on golangci-lint's binary.
func TestNoNewForbidigoPatterns(t *testing.T) {
	t.Parallel()

	patterns := []struct {
		path    string
		pattern string
	}{
		{path: "adapt.go", pattern: `fmt\.Fprintf\(os\.`},
		{path: "adapt.go", pattern: `os\.Exit\(`},
		{path: "resources.go", pattern: `fmt\.Fprintf\(os\.`},
		{path: "resources.go", pattern: `os\.Exit\(`},
	}

	for _, p := range patterns {
		data, err := os.ReadFile(p.path)
		if err != nil {
			t.Fatalf("read %s: %v", p.path, err)
		}
		if regexp.MustCompile(p.pattern).Match(data) {
			t.Errorf("%s contains forbidden pattern %q; use opts.IO.Out/ErrOut instead of os.Stdout/Stderr, and use cmdutil.ExitCodeFor instead of os.Exit", p.path, p.pattern)
		}
	}
}

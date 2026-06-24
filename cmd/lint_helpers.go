package cmd

import (
	"fmt"
	"os"
)

//nolint:gosec // G304: readFileImpl reads caller-supplied paths (test-injection seam for lint_test.go).
func readFileImpl(p string) ([]byte, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", p, err)
	}
	return data, nil
}

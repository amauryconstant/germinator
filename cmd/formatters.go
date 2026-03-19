package cmd

import (
	"fmt"
	"strings"

	"gitlab.com/amoconst/germinator/internal/domain"
)

func formatDryRunOutput(results []domain.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		fmt.Fprintf(&output, "Would write: %s\n  from: %s\n", result.OutputPath, result.InputPath)
	}
	return output.String()
}

func formatSuccessOutput(results []domain.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		fmt.Fprintf(&output, "Installed: %s -> %s\n", result.Ref, result.OutputPath)
	}
	return output.String()
}

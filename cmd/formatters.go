package cmd

import (
	"fmt"
	"strings"

	"gitlab.com/amoconst/germinator/internal/application"
)

func formatDryRunOutput(results []application.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		fmt.Fprintf(&output, "Would write: %s\n  from: %s\n", result.OutputPath, result.InputPath)
	}
	return output.String()
}

func formatSuccessOutput(results []application.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		fmt.Fprintf(&output, "Installed: %s -> %s\n", result.Ref, result.OutputPath)
	}
	return output.String()
}

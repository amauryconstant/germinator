package cmd

import (
	"fmt"
	"strings"

	"gitlab.com/amoconst/germinator/internal/domain"
)

func formatDryRunOutput(results []domain.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		if result.Error == nil {
			fmt.Fprintf(&output, "Would write: %s\n  from: %s\n", result.OutputPath, result.InputPath)
		} else {
			fmt.Fprintf(&output, "Would skip: %s (error: %v)\n", result.Ref, result.Error)
		}
	}
	return output.String()
}

func formatSuccessOutput(results []domain.InitializeResult) string {
	var output strings.Builder
	successCount := 0
	failCount := 0
	for _, result := range results {
		if result.Error == nil {
			fmt.Fprintf(&output, "Installed: %s -> %s\n", result.Ref, result.OutputPath)
			successCount++
		} else {
			fmt.Fprintf(&output, "Failed: %s (%v)\n", result.Ref, result.Error)
			failCount++
		}
	}
	return output.String()
}

func formatInitializeSummary(successCount, failCount int) string {
	if failCount == 0 {
		return fmt.Sprintf("Initialized %d resource(s).\n", successCount)
	}
	return fmt.Sprintf("Initialized %d resource(s), %d failed.\n", successCount, failCount)
}

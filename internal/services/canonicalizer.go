// Package services provides document transformation and validation services.
package services

import (
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
)

// CanonicalizeDocument converts a platform document to canonical YAML format.
func CanonicalizeDocument(inputPath string, outputPath string, platform string, docType string) error {
	doc, err := core.ParsePlatformDocument(inputPath, platform, docType)
	if err != nil {
		return fmt.Errorf("failed to parse platform document: %w", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return fmt.Errorf("validation failed: %v", errs)
	}

	yamlBytes, err := core.MarshalCanonical(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal canonical document: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(yamlBytes), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// validateCanonicalDoc validates a canonical document and returns any validation errors.
func validateCanonicalDoc(doc interface{}) []error {
	switch d := doc.(type) {
	case *core.CanonicalAgent:
		return d.Validate()
	case *core.CanonicalCommand:
		return d.Validate()
	case *core.CanonicalSkill:
		return d.Validate()
	case *core.CanonicalMemory:
		return d.Validate()
	default:
		return []error{fmt.Errorf("unknown document type: %T", doc)}
	}
}

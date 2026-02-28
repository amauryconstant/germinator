// Package services provides document transformation and validation services.
package services

import (
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

// CanonicalizeDocument converts a platform document to canonical YAML format.
func CanonicalizeDocument(inputPath string, outputPath string, platform string, docType string) error {
	doc, err := core.ParsePlatformDocument(inputPath, platform, docType)
	if err != nil {
		return gerrors.NewParseError(inputPath, "failed to parse platform document", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return gerrors.NewValidationError(errs[0].Error(), "", nil)
	}

	yamlBytes, err := core.MarshalCanonical(doc)
	if err != nil {
		return gerrors.NewTransformError("marshal", platform, "failed to marshal canonical document", err)
	}

	if err := os.WriteFile(outputPath, []byte(yamlBytes), 0644); err != nil {
		return gerrors.NewFileError(outputPath, "write", "failed to write output file", err)
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
		return []error{gerrors.NewParseError("", "unknown document type", nil)}
	}
}

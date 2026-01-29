// Package services provides business logic for document transformation and validation.
package services

import (
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/models"
)

// TransformDocument transforms a document to the target platform format.
func TransformDocument(inputPath, outputPath, platform string) error {
	doc, err := core.LoadDocument(inputPath, platform)
	if err != nil {
		return fmt.Errorf("failed to load document: %w", err)
	}

	rendered, err := core.RenderDocument(doc, platform)
	if err != nil {
		return fmt.Errorf("failed to render document: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ValidateDocument validates a document and returns any validation errors.
func ValidateDocument(inputPath, platform string) ([]error, error) {
	docType := core.DetectType(inputPath)
	if docType == "" {
		return nil, fmt.Errorf("unrecognizable filename: %s", inputPath)
	}

	doc, parseErr := core.ParseDocument(inputPath, docType)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse document: %w", parseErr)
	}

	switch d := doc.(type) {
	case *models.Agent:
		return d.Validate(platform), nil
	case *models.Command:
		return d.Validate(platform), nil
	case *models.Memory:
		return d.Validate(platform), nil
	case *models.Skill:
		return d.Validate(platform), nil
	default:
		return nil, fmt.Errorf("unknown document type: %T", d)
	}
}

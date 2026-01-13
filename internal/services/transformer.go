package services

import (
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/pkg/models"
)

func TransformDocument(inputPath, outputPath, platform string) error {
	doc, err := core.LoadDocument(inputPath)
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

func ValidateDocument(inputPath string) ([]error, error) {
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
		return d.Validate(), nil
	case *models.Command:
		return d.Validate(), nil
	case *models.Memory:
		return d.Validate(), nil
	case *models.Skill:
		return d.Validate(), nil
	default:
		return nil, fmt.Errorf("unknown document type: %T", d)
	}
}

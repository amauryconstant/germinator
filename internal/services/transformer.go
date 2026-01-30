// Package services provides business logic for document transformation and validation.
package services

import (
	"fmt"
	"os"
	"regexp"

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

// validateOpenCodeAgent validates OpenCode-specific Agent constraints.
func validateOpenCodeAgent(agent *models.Agent) []error {
	var errs []error

	if agent.Mode != "" && agent.Mode != "primary" && agent.Mode != "subagent" && agent.Mode != "all" {
		errs = append(errs, fmt.Errorf("invalid mode: %s (valid values: primary, subagent, all)", agent.Mode))
	}

	if agent.Temperature < 0.0 || agent.Temperature > 1.0 {
		errs = append(errs, fmt.Errorf("temperature must be between 0.0 and 1.0, got %f", agent.Temperature))
	}

	if agent.MaxSteps != 0 && agent.MaxSteps < 1 {
		errs = append(errs, fmt.Errorf("maxSteps must be >= 1, got %d", agent.MaxSteps))
	}

	return errs
}

// validateOpenCodeCommand validates OpenCode-specific Command constraints.
func validateOpenCodeCommand(cmd *models.Command) []error {
	var errs []error

	if cmd.Content == "" {
		errs = append(errs, fmt.Errorf("template (content) is required"))
	}

	return errs
}

// validateOpenCodeSkill validates OpenCode-specific Skill constraints.
func validateOpenCodeSkill(skill *models.Skill) []error {
	var errs []error

	skillNameRegex := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !skillNameRegex.MatchString(skill.Name) {
		errs = append(errs, fmt.Errorf("skill name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$, got %s", skill.Name))
	}

	if skill.Content == "" {
		errs = append(errs, fmt.Errorf("content is required"))
	}

	return errs
}

// validateOpenCodeMemory validates OpenCode-specific Memory constraints.
func validateOpenCodeMemory(mem *models.Memory) []error {
	var errs []error

	if len(mem.Paths) == 0 && mem.Content == "" {
		errs = append(errs, fmt.Errorf("paths or content is required"))
	}

	return errs
}

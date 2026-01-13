package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/amoconst/germinator/pkg/models"
)

func RenderDocument(doc interface{}, platform string) (string, error) {
	docType, err := getDocType(doc)
	if err != nil {
		return "", fmt.Errorf("failed to determine document type: %w", err)
	}

	tmplPath, err := getTemplatePath(platform, docType+".tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to get template path: %w", err)
	}

	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", tmplPath, err)
	}

	tmpl, err := template.New(docType).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, doc); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return sb.String(), nil
}

func getTemplatePath(platform string, filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	relPath := filepath.Join("config", "templates", platform, filename)
	tmplPath := filepath.Join(cwd, relPath)

	if _, err := os.Stat(tmplPath); err != nil {
		altPath := filepath.Join(filepath.Join(cwd, "..", ".."), relPath)
		if _, err := os.Stat(altPath); err == nil {
			return filepath.Abs(altPath)
		}
		return "", fmt.Errorf("template file not found: %s", relPath)
	}

	return filepath.Abs(tmplPath)
}

func getDocType(doc interface{}) (string, error) {
	switch d := doc.(type) {
	case *models.Agent:
		return "agent", nil
	case *models.Command:
		return "command", nil
	case *models.Memory:
		return "memory", nil
	case *models.Skill:
		return "skill", nil
	default:
		return "", fmt.Errorf("unknown document type: %T", d)
	}
}

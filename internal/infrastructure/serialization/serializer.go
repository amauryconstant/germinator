// Package serialization provides document serialization and rendering functionality.
package serialization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	gerrors "gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/adapters"
	claudecode "gitlab.com/amoconst/germinator/internal/infrastructure/adapters/claude-code"
	opencode "gitlab.com/amoconst/germinator/internal/infrastructure/adapters/opencode"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
)

type templateContext struct {
	Doc     interface{}
	Adapter interface{}
}

type canonicalTemplateContext struct {
	Doc interface{}
}

// RenderDocument renders a document using platform-specific template.
func RenderDocument(doc interface{}, platform string) (string, error) {
	docType, err := getDocType(doc)
	if err != nil {
		return "", gerrors.NewTransformError("render", platform, "failed to determine document type", err)
	}

	tmplPath, err := getTemplatePath(platform, docType+".tmpl")
	if err != nil {
		return "", gerrors.NewTransformError("render", platform, "failed to get template path", err)
	}

	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", gerrors.NewFileError(tmplPath, "read", "failed to read template file", err)
	}

	var adapter interface{}
	switch platform {
	case "claude-code":
		adapter = claudecode.New()
	case "opencode":
		adapter = opencode.New()
	}

	ctx := templateContext{
		Doc:     doc,
		Adapter: adapter,
	}

	tmpl, err := template.New(docType).Funcs(createTemplateFuncMap()).Parse(string(tmplContent))
	if err != nil {
		return "", gerrors.NewTransformError("render", platform, "failed to parse template", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, ctx); err != nil {
		return "", gerrors.NewTransformError("render", platform, "failed to execute template", err)
	}

	return sb.String(), nil
}

// createTemplateFuncMap creates and returns a FuncMap with all template functions.
// This map is passed to template.Funcs() to make custom functions available in templates.
//
// It combines Sprig's built-in functions with our custom functions:
//   - Sprig provides string functions like lower, upper, trim, etc.
//   - permissionPolicyToClaudeCode: converts canonical permission policy to Claude Code enum
//   - permissionPolicyToOpenCode: converts canonical permission policy to OpenCode permission map as YAML string
//   - convertToolNameCase: converts tool name to platform-specific case
//
// Returns:
//   - map[string]interface{}: Template function map containing Sprig and custom functions
func createTemplateFuncMap() map[string]interface{} {
	funcMap := sprig.FuncMap()

	funcMap["permissionPolicyToClaudeCode"] = func(policy gerrors.PermissionPolicy) string {
		if policy == "" {
			return ""
		}
		adapter := claudecode.New()
		result, err := adapter.PermissionPolicyToPlatform(policy)
		if err != nil {
			return ""
		}
		if s, ok := result.(string); ok {
			return s
		}
		return ""
	}

	funcMap["permissionPolicyToOpenCode"] = func(policy gerrors.PermissionPolicy) string {
		if policy == "" {
			return ""
		}
		adapter := opencode.New()
		result, err := adapter.PermissionPolicyToPlatform(policy)
		if err != nil {
			return ""
		}
		if permMap, ok := result.(adapters.PermissionMap); ok {
			var sb strings.Builder
			sb.WriteString("  edit:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.Edit))
			sb.WriteString("\n")
			sb.WriteString("  bash:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.Bash))
			sb.WriteString("\n")
			sb.WriteString("  read:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.Read))
			sb.WriteString("\n")
			sb.WriteString("  grep:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.Grep))
			sb.WriteString("\n")
			sb.WriteString("  glob:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.Glob))
			sb.WriteString("\n")
			sb.WriteString("  list:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.List))
			sb.WriteString("\n")
			sb.WriteString("  webfetch:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.WebFetch))
			sb.WriteString("\n")
			sb.WriteString("  websearch:\n")
			sb.WriteString("    \"*\": ")
			sb.WriteString(string(permMap.WebSearch))
			return sb.String()
		}
		return ""
	}

	funcMap["convertToolNameCase"] = func(name string, platform string) string {
		switch platform {
		case "claude-code":
			adapter := claudecode.New()
			return adapter.ConvertToolNameCase(name)
		case "opencode":
			adapter := opencode.New()
			return adapter.ConvertToolNameCase(name)
		default:
			return name
		}
	}

	return funcMap
}

// findProjectRoot walks up the directory tree from cwd looking for go.mod.
// Returns the project root directory and an error if not found.
func findProjectRoot(cwd string) (string, error) {
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", gerrors.NewFileError(cwd, "read", "project root not found (no go.mod)", nil)
}

func getTemplatePath(platform string, filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}

	relPath := filepath.Join("config", "templates", platform, filename)

	// First try CWD (works when running from project root)
	tmplPath := filepath.Join(cwd, relPath)
	if _, err := os.Stat(tmplPath); err == nil {
		absPath, err := filepath.Abs(tmplPath)
		if err != nil {
			return "", fmt.Errorf("resolving absolute path: %w", err)
		}
		return absPath, nil
	}

	// Try finding project root (works when running from package directory in tests)
	projectRoot, err := findProjectRoot(cwd)
	if err != nil {
		return "", gerrors.NewFileError(relPath, "read", "template file not found", nil)
	}
	tmplPath = filepath.Join(projectRoot, relPath)
	if _, err := os.Stat(tmplPath); err == nil {
		absPath, err := filepath.Abs(tmplPath)
		if err != nil {
			return "", fmt.Errorf("resolving absolute path: %w", err)
		}
		return absPath, nil
	}

	return "", gerrors.NewFileError(relPath, "read", "template file not found", nil)
}

func getDocType(doc interface{}) (string, error) {
	switch d := doc.(type) {
	case *parsing.CanonicalAgent:
		return "agent", nil
	case *parsing.CanonicalCommand:
		return "command", nil
	case *parsing.CanonicalMemory:
		return "memory", nil
	case *parsing.CanonicalSkill:
		return "skill", nil
	default:
		return "", gerrors.NewTransformError("marshal", "canonical", fmt.Sprintf("unknown document type: %T", d), nil)
	}
}

// MarshalCanonical serializes a canonical model to YAML string using canonical templates.
func MarshalCanonical(doc interface{}) (string, error) {
	docType, err := getDocType(doc)
	if err != nil {
		return "", gerrors.NewTransformError("marshal", "canonical", "failed to determine document type", err)
	}

	tmplPath, err := getCanonicalTemplatePath(docType + ".tmpl")
	if err != nil {
		return "", gerrors.NewTransformError("marshal", "canonical", "failed to get template path", err)
	}

	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", gerrors.NewFileError(tmplPath, "read", "failed to read template file", err)
	}

	ctx := canonicalTemplateContext{
		Doc: doc,
	}

	tmpl, err := template.New(docType).Funcs(createCanonicalTemplateFuncMap()).Parse(string(tmplContent))
	if err != nil {
		return "", gerrors.NewTransformError("marshal", "canonical", "failed to parse template", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, ctx); err != nil {
		return "", gerrors.NewTransformError("marshal", "canonical", "failed to execute template", err)
	}

	return sb.String(), nil
}

// getCanonicalTemplatePath returns the absolute path to a canonical template file.
func getCanonicalTemplatePath(filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}

	relPath := filepath.Join("config", "templates", "canonical", filename)

	// First try CWD (works when running from project root)
	tmplPath := filepath.Join(cwd, relPath)
	if _, err := os.Stat(tmplPath); err == nil {
		absPath, err := filepath.Abs(tmplPath)
		if err != nil {
			return "", fmt.Errorf("resolving absolute path: %w", err)
		}
		return absPath, nil
	}

	// Try finding project root (works when running from package directory in tests)
	projectRoot, err := findProjectRoot(cwd)
	if err != nil {
		return "", gerrors.NewFileError(relPath, "read", "canonical template file not found", nil)
	}
	tmplPath = filepath.Join(projectRoot, relPath)
	if _, err := os.Stat(tmplPath); err == nil {
		absPath, err := filepath.Abs(tmplPath)
		if err != nil {
			return "", fmt.Errorf("resolving absolute path: %w", err)
		}
		return absPath, nil
	}

	return "", gerrors.NewFileError(relPath, "read", "canonical template file not found", nil)
}

// createCanonicalTemplateFuncMap creates and returns a FuncMap with minimal Sprig functions for canonical templates.
func createCanonicalTemplateFuncMap() map[string]interface{} {
	return sprig.FuncMap()
}

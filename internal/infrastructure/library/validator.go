// Package library provides library management for canonical resources.
package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/domain"
	yaml "gopkg.in/yaml.v3"
)

// IssueType represents the type of validation issue.
type IssueType string

// IssueType constants define the different validation issue types.
const (
	IssueTypeMissingFile          IssueType = "missing-file"
	IssueTypeGhostResource        IssueType = "ghost-resource"
	IssueTypeOrphan               IssueType = "orphan"
	IssueTypeMalformedFrontmatter IssueType = "malformed-frontmatter"
)

// Severity represents the severity level of an issue.
type Severity string

// Severity constants.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Issue represents a single validation issue found in the library.
type Issue struct {
	// Type is the type of issue.
	Type IssueType `yaml:"type"`
	// Severity indicates whether this is an error or warning.
	Severity Severity `yaml:"severity"`
	// Ref is the resource reference (e.g., "skill/commit") for missing-file and ghost-resource issues.
	Ref string `yaml:"ref,omitempty"`
	// Path is the file path relevant to the issue.
	Path string `yaml:"path,omitempty"`
	// InPreset is the preset name that references a ghost resource.
	InPreset string `yaml:"inPreset,omitempty"`
	// Message provides additional context about the issue.
	Message string `yaml:"message,omitempty"`
}

// ValidationResult holds the result of library validation.
type ValidationResult struct {
	// Valid is true if no issues were found.
	Valid bool `yaml:"valid"`
	// ErrorCount is the number of error-level issues.
	ErrorCount int `yaml:"errorCount"`
	// WarningCount is the number of warning-level issues.
	WarningCount int `yaml:"warningCount"`
	// Issues contains all detected issues.
	Issues []Issue `yaml:"issues"`
	// FixApplied is true if --fix was applied.
	FixApplied bool `yaml:"fixApplied,omitempty"`
	// FixResult contains information about fixes applied (if any).
	FixResult *FixResult `yaml:"fixResult,omitempty"`
}

// AddIssue adds an issue and updates counts accordingly.
func (vr *ValidationResult) AddIssue(issue Issue) {
	vr.Issues = append(vr.Issues, issue)
	if issue.Severity == SeverityError {
		vr.ErrorCount++
	} else {
		vr.WarningCount++
	}
	vr.Valid = vr.ErrorCount == 0
}

// ValidateLibrary validates the library for various issues.
// It runs all four checks: missing files, orphaned files, ghost resources, and malformed frontmatter.
func ValidateLibrary(lib *Library) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	// Run all checks
	missingFileIssues, err := CheckMissingFiles(lib)
	if err != nil {
		return nil, fmt.Errorf("checking missing files: %w", err)
	}
	for _, issue := range missingFileIssues {
		result.AddIssue(issue)
	}

	orphanIssues, err := CheckOrphanedFiles(lib)
	if err != nil {
		return nil, fmt.Errorf("checking orphaned files: %w", err)
	}
	for _, issue := range orphanIssues {
		result.AddIssue(issue)
	}

	ghostIssues, err := CheckGhostResources(lib)
	if err != nil {
		return nil, fmt.Errorf("checking ghost resources: %w", err)
	}
	for _, issue := range ghostIssues {
		result.AddIssue(issue)
	}

	malformedIssues, err := CheckMalformedFrontmatter(lib)
	if err != nil {
		return nil, fmt.Errorf("checking malformed frontmatter: %w", err)
	}
	for _, issue := range malformedIssues {
		result.AddIssue(issue)
	}

	return result, nil
}

// CheckMissingFiles verifies that entries in library.yaml have corresponding files on disk.
func CheckMissingFiles(lib *Library) ([]Issue, error) {
	var issues []Issue

	for typ, resources := range lib.Resources {
		for name, res := range resources {
			if res.Path == "" {
				continue
			}
			fullPath := filepath.Join(lib.RootPath, res.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				ref := FormatRef(typ, name)
				issues = append(issues, Issue{
					Type:     IssueTypeMissingFile,
					Severity: SeverityError,
					Ref:      ref,
					Path:     res.Path,
					Message:  fmt.Sprintf("resource %q references file %q which does not exist", ref, res.Path),
				})
			}
		}
	}

	return issues, nil
}

// CheckOrphanedFiles finds files on disk that are not registered in library.yaml.
func CheckOrphanedFiles(lib *Library) ([]Issue, error) {
	var issues []Issue

	// Collect all registered paths
	registeredPaths := make(map[string]bool)
	for _, resources := range lib.Resources {
		for _, res := range resources {
			if res.Path != "" {
				registeredPaths[res.Path] = true
			}
		}
	}

	// Check each resource type directory
	for _, resType := range ValidResourceTypes {
		dirName := fmt.Sprintf("%ss", resType) // skills, agents, commands, memory
		dirPath := filepath.Join(lib.RootPath, dirName)

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Directory doesn't exist, skip
			}
			return nil, fmt.Errorf("reading directory %s: %w", dirPath, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			relPath := filepath.Join(dirName, entry.Name())
			if !registeredPaths[relPath] {
				issues = append(issues, Issue{
					Type:     IssueTypeOrphan,
					Severity: SeverityWarning,
					Path:     relPath,
					Message:  fmt.Sprintf("file %q exists but is not registered in library.yaml", relPath),
				})
			}
		}
	}

	return issues, nil
}

// CheckGhostResources verifies that preset references exist in the library.
// A preset reference is considered ghost if the resource either:
// 1. Doesn't exist in the library Resources map
// 2. Exists but the referenced file is missing
func CheckGhostResources(lib *Library) ([]Issue, error) {
	var issues []Issue

	// Build a set of all valid resource refs (resources with existing files)
	validRefs := make(map[string]bool)
	for typ, resources := range lib.Resources {
		for name, res := range resources {
			if res.Path == "" {
				continue
			}
			fullPath := filepath.Join(lib.RootPath, res.Path)
			if _, err := os.Stat(fullPath); err == nil {
				validRefs[FormatRef(typ, name)] = true
			}
		}
	}

	// Check each preset's resource references
	for presetName, preset := range lib.Presets {
		for _, ref := range preset.Resources {
			if !validRefs[ref] {
				issues = append(issues, Issue{
					Type:     IssueTypeGhostResource,
					Severity: SeverityError,
					Ref:      ref,
					InPreset: presetName,
					Message:  fmt.Sprintf("preset %q references non-existent resource %q", presetName, ref),
				})
			}
		}
	}

	return issues, nil
}

// CheckMalformedFrontmatter parses frontmatter from each resource file and checks for validity.
func CheckMalformedFrontmatter(lib *Library) ([]Issue, error) {
	var issues []Issue

	for typ, resources := range lib.Resources {
		for name, res := range resources {
			if res.Path == "" {
				continue
			}
			fullPath := filepath.Join(lib.RootPath, res.Path)

			// Check if file exists first
			info, err := os.Stat(fullPath)
			if err != nil {
				continue // Missing file is handled by CheckMissingFiles
			}
			if info.IsDir() {
				continue
			}

			// Read the file
			content, err := os.ReadFile(fullPath) //nolint:gosec // G304: User provides library path, reading library files
			if err != nil {
				continue
			}

			// Check for frontmatter
			if !hasFrontmatter(string(content)) {
				continue // No frontmatter is valid
			}

			// Try to parse the frontmatter
			_, err = parseFrontmatter(string(content))
			if err != nil {
				ref := FormatRef(typ, name)
				issues = append(issues, Issue{
					Type:     IssueTypeMalformedFrontmatter,
					Severity: SeverityError,
					Ref:      ref,
					Path:     res.Path,
					Message:  fmt.Sprintf("resource %q has malformed frontmatter: %v", ref, err),
				})
			}
		}
	}

	return issues, nil
}

// hasFrontmatter checks if content has YAML frontmatter delimiters.
func hasFrontmatter(content string) bool {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return false
	}
	return strings.TrimSpace(lines[0]) == "---"
}

// parseFrontmatter extracts and parses YAML frontmatter from content.
func parseFrontmatter(content string) (map[string]interface{}, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, gerrors.NewParseError("<content>", "no frontmatter delimiters found", nil)
	}

	var yamlLines []string
	foundEnd := false

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			foundEnd = true
			break
		}
		yamlLines = append(yamlLines, lines[i])
	}

	if !foundEnd {
		return nil, gerrors.NewParseError("<content>", "frontmatter missing closing delimiter", nil)
	}

	yamlContent := strings.Join(yamlLines, "\n")
	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return nil, gerrors.NewParseError("<content>", fmt.Sprintf("invalid YAML in frontmatter: %v", err), nil)
	}

	return frontmatter, nil
}

// FixLibrary removes missing entries and ghost preset refs from library.yaml.
// It only modifies library.yaml, never deletes actual files.
func FixLibrary(lib *Library) (*FixResult, error) {
	result := &FixResult{}

	// First pass: identify missing file entries (but don't remove yet)
	// We need to track these to avoid double-counting as ghost refs
	missingRefs := make(map[string]bool)
	for typ, resources := range lib.Resources {
		for name, res := range resources {
			if res.Path == "" {
				continue
			}
			fullPath := filepath.Join(lib.RootPath, res.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				ref := FormatRef(typ, name)
				missingRefs[ref] = true
				result.MissingFileRefs = append(result.MissingFileRefs, ref)
			}
		}
	}

	// Remove missing file entries from Resources
	for typ, resources := range lib.Resources {
		for name := range resources {
			ref := FormatRef(typ, name)
			if missingRefs[ref] {
				delete(lib.Resources[typ], name)
			}
		}
	}

	// Remove empty resource type maps
	for typ := range lib.Resources {
		if len(lib.Resources[typ]) == 0 {
			delete(lib.Resources, typ)
		}
	}

	// Identify and strip ghost resources from presets
	stripGhostResources(lib, missingRefs, result)

	// Save the fixed library
	if err := SaveLibrary(lib); err != nil {
		return nil, fmt.Errorf("saving fixed library: %w", err)
	}

	return result, nil
}

// stripGhostResources removes ghost preset refs and empty presets.
func stripGhostResources(lib *Library, missingRefs map[string]bool, result *FixResult) {
	for presetName, preset := range lib.Presets {
		validResources := filterPresetRefs(lib, missingRefs, preset.Resources, result)
		preset.Resources = validResources
		if len(preset.Resources) == 0 {
			delete(lib.Presets, presetName)
			continue
		}
		lib.Presets[presetName] = preset
	}
}

// filterPresetRefs filters out ghost refs from a preset's resources.
func filterPresetRefs(lib *Library, missingRefs map[string]bool, refs []string, result *FixResult) []string {
	var valid []string
	for _, ref := range refs {
		if missingRefs[ref] {
			continue // Already counted as missing file
		}
		if isGhostRef(lib, ref) {
			result.GhostResourceRefs = append(result.GhostResourceRefs, ref)
			continue
		}
		valid = append(valid, ref)
	}
	return valid
}

// isGhostRef checks if a preset ref points to a non-existent resource.
func isGhostRef(lib *Library, ref string) bool {
	typ, name, err := ParseRef(ref)
	if err != nil {
		return false
	}
	if _, exists := lib.Resources[typ]; !exists {
		return true
	}
	if _, exists := lib.Resources[typ][name]; !exists {
		return true
	}
	return false
}

// FixResult contains information about fixes applied to the library.
type FixResult struct {
	// MissingFileRefs are refs whose files were missing and entries were removed.
	MissingFileRefs []string `yaml:"missingFileRefs"`
	// GhostResourceRefs are resource refs that were stripped from presets.
	GhostResourceRefs []string `yaml:"ghostResourceRefs"`
}

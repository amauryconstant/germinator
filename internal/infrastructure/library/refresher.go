package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// RefreshOptions contains options for refreshing the library.
type RefreshOptions struct {
	LibraryPath string
	DryRun      bool
	Force       bool
}

// RefreshResult contains the result of a refresh operation.
type RefreshResult struct {
	Refreshed []RefreshChange
	Skipped   []SkipInfo
	Errors    []RefreshError
}

// RefreshChange represents a change made during refresh.
type RefreshChange struct {
	Ref   string
	Field string
	Old   string
	New   string
}

// SkipInfo represents a skipped resource during refresh.
type SkipInfo struct {
	Ref    string
	Reason string
}

// RefreshError represents an error that occurred during refresh.
type RefreshError struct {
	Ref   string
	Field string
	Type  string
}

// RefreshLibrary syncs metadata from registered resource files into library.yaml.
// It updates descriptions when they differ from frontmatter, updates paths when files
// are renamed (if frontmatter name matches), and detects conflicts.
func RefreshLibrary(opts RefreshOptions) (*RefreshResult, error) {
	// Load the library
	lib, err := LoadLibrary(opts.LibraryPath)
	if err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	result := &RefreshResult{}

	// Process each resource type
	for resType, resources := range lib.Resources {
		for name, res := range resources {
			ref := FormatRef(resType, name)
			processResource(opts, lib, ref, resType, name, res, result)
		}
	}

	// Save if not dry-run and we have changes
	if !opts.DryRun && len(result.Refreshed) > 0 {
		if err := SaveLibrary(lib); err != nil {
			return nil, fmt.Errorf("saving library: %w", err)
		}
	}

	return result, nil
}

// processResource processes a single resource for refresh.
func processResource(opts RefreshOptions, lib *Library, ref, resType, name string, res Resource, result *RefreshResult) {
	filePath := filepath.Join(lib.RootPath, res.Path)
	originalPath := filePath
	fileRenamed := false

	// Check if file exists - if not, search directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dirPath := filepath.Join(lib.RootPath, resType+"s")
		foundPath, foundName := searchForFile(dirPath, name)

		if foundPath != "" && foundName == name {
			// Found renamed file, update path
			filePath = foundPath
			if !opts.DryRun {
				res.Path = resType + "s/" + filepath.Base(foundPath)
				lib.Resources[resType][name] = res
			}
			result.Refreshed = append(result.Refreshed, RefreshChange{
				Ref:   ref,
				Field: "path",
				Old:   originalPath,
				New:   foundPath,
			})
			fileRenamed = true
		} else if foundPath != "" {
			// Found a file but name doesn't match
			recordNameMismatch(opts, ref, result)
			return
		} else {
			// File not found at registered path and not found in directory
			// Skip silently (left to validate --fix)
			return
		}
	}

	// Check for name mismatch conflict (only if file wasn't renamed)
	if !fileRenamed {
		frontmatterName := extractFrontmatterField(filePath, "name")
		if frontmatterName != "" && frontmatterName != name {
			recordNameMismatch(opts, ref, result)
			return
		}
	}

	// Check for malformed frontmatter
	if isMalformedFrontmatter(filePath) {
		result.Errors = append(result.Errors, RefreshError{
			Ref:   ref,
			Type:  "malformed_frontmatter",
			Field: "frontmatter",
		})
		result.Skipped = append(result.Skipped, SkipInfo{
			Ref:    ref,
			Reason: "malformed_frontmatter",
		})
		return
	}

	// Update description if different
	frontmatterDesc := extractFrontmatterField(filePath, "description")
	if frontmatterDesc != "" && frontmatterDesc != res.Description {
		if !opts.DryRun {
			res.Description = frontmatterDesc
			lib.Resources[resType][name] = res
		}
		result.Refreshed = append(result.Refreshed, RefreshChange{
			Ref:   ref,
			Field: "description",
			Old:   res.Description,
			New:   frontmatterDesc,
		})
	}
}

// recordNameMismatch records a name mismatch error and skip.
func recordNameMismatch(opts RefreshOptions, ref string, result *RefreshResult) {
	if !opts.Force {
		result.Errors = append(result.Errors, RefreshError{
			Ref:   ref,
			Type:  "name_mismatch",
			Field: "name",
		})
	}
	result.Skipped = append(result.Skipped, SkipInfo{
		Ref:    ref,
		Reason: "name_mismatch",
	})
}

// isMalformedFrontmatter checks if the file has malformed frontmatter.
func isMalformedFrontmatter(filePath string) bool {
	yamlContent, yamlErr := extractFrontmatter(filePath)

	// Check for malformed frontmatter: file starts with --- but extractFrontmatter returned empty
	if yamlErr == nil && yamlContent == "" {
		// Check if raw file content starts with --- (malformed if so)
		content, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is derived from library resource path, not user input
		if err != nil {
			return false
		}
		lines := strings.Split(string(content), "\n")
		return len(lines) > 0 && strings.TrimSpace(lines[0]) == "---"
	}

	// Check if YAML parsing fails
	if yamlErr == nil && yamlContent != "" {
		var frontmatter map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
			return true
		}
	}

	return false
}

// searchForFile searches a directory for a file with matching frontmatter name.
// Returns (path, name, found) where found is true if a file was located,
// regardless of whether the name matched.
func searchForFile(dirPath, targetName string) (string, string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(dirPath, entry.Name())
		name := extractFrontmatterField(filePath, "name")
		if name == targetName {
			return filePath, name
		}
		// Found a file but name doesn't match - return it anyway
		// so the caller can distinguish "found with mismatch" from "not found"
		if name != "" {
			return filePath, name
		}
	}

	return "", ""
}

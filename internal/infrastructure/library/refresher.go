package library

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Skip missing files silently (left to validate --fix)
		return
	}

	// Extract frontmatter name for conflict checking
	frontmatterName := extractFrontmatterField(filePath, "name")

	// Check for name mismatch conflict
	if frontmatterName != "" && frontmatterName != name {
		// Error only if not force mode
		if !opts.Force {
			result.Errors = append(result.Errors, RefreshError{
				Ref:   ref,
				Type:  "name_mismatch",
				Field: "name",
			})
		}
		// Skip regardless of force flag when there's a name mismatch
		result.Skipped = append(result.Skipped, SkipInfo{
			Ref:    ref,
			Reason: "name_mismatch",
		})
		return
	}

	// Extract frontmatter description
	frontmatterDesc := extractFrontmatterField(filePath, "description")

	// Determine if frontmatter is malformed (invalid YAML syntax)
	yamlContent, yamlErr := extractFrontmatter(filePath)
	if yamlErr == nil && yamlContent != "" {
		// Try to parse the YAML
		var frontmatter map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
			// Malformed YAML
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
	}

	// Check if we need to search for the file in the directory (renamed file)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File not at registered path, search directory
		dirPath := filepath.Join(lib.RootPath, resType+"s")
		foundPath, foundName := searchForFile(dirPath, name)
		if foundPath != "" && foundName == name {
			// Verify frontmatter name matches
			if !opts.DryRun {
				res.Path = resType + "s/" + filepath.Base(foundPath)
				lib.Resources[resType][name] = res
			}
			result.Refreshed = append(result.Refreshed, RefreshChange{
				Ref:   ref,
				Field: "path",
				Old:   res.Path,
				New:   resType + "s/" + filepath.Base(foundPath),
			})
		}
	}

	// Update description if different
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

// searchForFile searches a directory for a file with matching frontmatter name.
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
	}

	return "", ""
}

package library

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ResourceInfo contains display information about a resource.
type ResourceInfo struct {
	Type        string
	Name        string
	Path        string
	Description string
}

// ListResources returns all resources grouped by type.
// Returns a map where keys are resource types and values are lists of ResourceInfo.
func ListResources(lib *Library) map[string][]ResourceInfo {
	result := make(map[string][]ResourceInfo)

	for typ, resources := range lib.Resources {
		var infos []ResourceInfo
		for name, res := range resources {
			infos = append(infos, ResourceInfo{
				Type:        typ,
				Name:        name,
				Path:        res.Path,
				Description: res.Description,
			})
		}
		// Sort by name
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Name < infos[j].Name
		})
		result[typ] = infos
	}

	return result
}

// ListPresets returns all presets sorted by name.
func ListPresets(lib *Library) []Preset {
	presets := make([]Preset, 0, len(lib.Presets))
	for _, preset := range lib.Presets {
		presets = append(presets, preset)
	}

	// Sort by name
	sort.Slice(presets, func(i, j int) bool {
		return presets[i].Name < presets[j].Name
	})

	return presets
}

// FormatResourcesList formats resources for CLI output, grouped by type.
func FormatResourcesList(lib *Library) string {
	var sb strings.Builder

	resources := ListResources(lib)

	// Define order of types
	typeOrder := []string{
		string(ResourceTypeSkill),
		string(ResourceTypeAgent),
		string(ResourceTypeCommand),
		string(ResourceTypeMemory),
	}

	hasContent := false
	for _, typ := range typeOrder {
		infos, ok := resources[typ]
		if !ok || len(infos) == 0 {
			continue
		}

		if hasContent {
			sb.WriteString("\n")
		}
		hasContent = true

		// Capitalize type for header
		header := cases.Title(language.English).String(typ) + "s"
		sb.WriteString(header + ":\n")

		for _, info := range infos {
			ref := FormatRef(info.Type, info.Name)
			if info.Description != "" {
				sb.WriteString(fmt.Sprintf("  %s - %s\n", ref, info.Description))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", ref))
			}
		}
	}

	if !hasContent {
		return "No resources found.\n"
	}

	return sb.String()
}

// FormatPresetsList formats presets for CLI output.
func FormatPresetsList(lib *Library) string {
	var sb strings.Builder

	presets := ListPresets(lib)

	if len(presets) == 0 {
		return "No presets found.\n"
	}

	for _, preset := range presets {
		if preset.Description != "" {
			sb.WriteString(fmt.Sprintf("%s - %s\n", preset.Name, preset.Description))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", preset.Name))
		}

		// List resources in preset
		for _, ref := range preset.Resources {
			sb.WriteString(fmt.Sprintf("  - %s\n", ref))
		}
	}

	return sb.String()
}

// FormatResourceDetails formats a single resource for detailed display.
func FormatResourceDetails(lib *Library, ref string) (string, error) {
	typ, name, err := ParseRef(ref)
	if err != nil {
		return "", err
	}

	resources, ok := lib.Resources[typ]
	if !ok {
		return "", fmt.Errorf("resource not found: %s", ref)
	}

	res, ok := resources[name]
	if !ok {
		return "", fmt.Errorf("resource not found: %s", ref)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Reference: %s\n", ref))
	sb.WriteString(fmt.Sprintf("Path: %s\n", res.Path))
	if res.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", res.Description))
	}

	return sb.String(), nil
}

// FormatPresetDetails formats a single preset for detailed display.
func FormatPresetDetails(lib *Library, name string) (string, error) {
	preset, ok := lib.Presets[name]
	if !ok {
		return "", fmt.Errorf("preset not found: %s", name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Preset: %s\n", name))
	if preset.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", preset.Description))
	}
	sb.WriteString("Resources:\n")
	for _, ref := range preset.Resources {
		sb.WriteString(fmt.Sprintf("  - %s\n", ref))
	}

	return sb.String(), nil
}

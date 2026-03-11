package library

import (
	"sort"
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

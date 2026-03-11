package cmd

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gitlab.com/amoconst/germinator/internal/library"
)

func formatResourcesList(lib *library.Library) string {
	var sb strings.Builder

	resources := library.ListResources(lib)

	typeOrder := []string{
		string(library.ResourceTypeSkill),
		string(library.ResourceTypeAgent),
		string(library.ResourceTypeCommand),
		string(library.ResourceTypeMemory),
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

		header := cases.Title(language.English).String(typ) + "s"
		sb.WriteString(header + ":\n")

		for _, info := range infos {
			ref := library.FormatRef(info.Type, info.Name)
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

func formatPresetsList(lib *library.Library) string {
	var sb strings.Builder

	presets := library.ListPresets(lib)

	if len(presets) == 0 {
		return "No presets found.\n"
	}

	for _, preset := range presets {
		if preset.Description != "" {
			sb.WriteString(fmt.Sprintf("%s - %s\n", preset.Name, preset.Description))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", preset.Name))
		}

		for _, ref := range preset.Resources {
			sb.WriteString(fmt.Sprintf("  - %s\n", ref))
		}
	}

	return sb.String()
}

func formatResourceDetails(lib *library.Library, ref string) (string, error) {
	typ, name, err := library.ParseRef(ref)
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

func formatPresetDetails(lib *library.Library, name string) (string, error) {
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

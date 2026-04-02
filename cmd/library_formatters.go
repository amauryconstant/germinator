package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
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
				fmt.Fprintf(&sb, "  %s - %s\n", ref, info.Description)
			} else {
				fmt.Fprintf(&sb, "  %s\n", ref)
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
			fmt.Fprintf(&sb, "%s - %s\n", preset.Name, preset.Description)
		} else {
			fmt.Fprintf(&sb, "%s\n", preset.Name)
		}

		for _, ref := range preset.Resources {
			fmt.Fprintf(&sb, "  - %s\n", ref)
		}
	}

	return sb.String()
}

func formatResourceDetails(lib *library.Library, ref string) (string, error) {
	typ, name, err := library.ParseRef(ref)
	if err != nil {
		return "", fmt.Errorf("parsing ref %q: %w", ref, err)
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
	fmt.Fprintf(&sb, "Reference: %s\n", ref)
	fmt.Fprintf(&sb, "Path: %s\n", res.Path)
	if res.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", res.Description)
	}

	return sb.String(), nil
}

func formatPresetDetails(lib *library.Library, name string) (string, error) {
	preset, ok := lib.Presets[name]
	if !ok {
		return "", fmt.Errorf("preset not found: %s", name)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Preset: %s\n", name)
	if preset.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", preset.Description)
	}
	sb.WriteString("Resources:\n")
	for _, ref := range preset.Resources {
		fmt.Fprintf(&sb, "  - %s\n", ref)
	}

	return sb.String(), nil
}

// formatPresetOutput formats a preset for command output (e.g., after creation).
func formatPresetOutput(lib *library.Library, name string) string {
	preset, ok := lib.Presets[name]
	if !ok {
		return fmt.Sprintf("Preset '%s' created successfully.\n", name)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Preset '%s' created successfully.\n", name)
	if preset.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", preset.Description)
	}
	sb.WriteString("Resources:\n")
	for _, ref := range preset.Resources {
		fmt.Fprintf(&sb, "  - %s\n", ref)
	}

	return sb.String()
}

// ResourcesJSONOutput represents the JSON output for library resources.
type ResourcesJSONOutput struct {
	Resources []ResourceInfoJSON `json:"resources"`
}

// ResourceInfoJSON represents a resource in JSON output.
type ResourceInfoJSON struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// outputResourcesJSON outputs resources in JSON format.
func outputResourcesJSON(c *cobra.Command, lib *library.Library) error {
	resources := library.ListResources(lib)

	typeOrder := []string{
		string(library.ResourceTypeSkill),
		string(library.ResourceTypeAgent),
		string(library.ResourceTypeCommand),
		string(library.ResourceTypeMemory),
	}

	var allResources []ResourceInfoJSON
	for _, typ := range typeOrder {
		infos, ok := resources[typ]
		if !ok || len(infos) == 0 {
			continue
		}
		for _, info := range infos {
			allResources = append(allResources, ResourceInfoJSON{
				Type:        info.Type,
				Name:        info.Name,
				Path:        info.Path,
				Description: info.Description,
			})
		}
	}

	output := ResourcesJSONOutput{Resources: allResources}
	encoder := json.NewEncoder(c.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}

// PresetsJSONOutput represents the JSON output for library presets.
type PresetsJSONOutput struct {
	Presets []PresetInfoJSON `json:"presets"`
}

// PresetInfoJSON represents a preset in JSON output.
type PresetInfoJSON struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Resources   []string `json:"resources"`
}

// outputPresetsJSON outputs presets in JSON format.
func outputPresetsJSON(c *cobra.Command, lib *library.Library) error {
	presets := library.ListPresets(lib)

	allPresets := make([]PresetInfoJSON, 0, len(presets))
	for _, preset := range presets {
		allPresets = append(allPresets, PresetInfoJSON{
			Name:        preset.Name,
			Description: preset.Description,
			Resources:   preset.Resources,
		})
	}

	output := PresetsJSONOutput{Presets: allPresets}
	encoder := json.NewEncoder(c.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}

// ShowResourceJSONOutput represents the JSON output for a single resource.
type ShowResourceJSONOutput struct {
	Ref         string `json:"ref"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// outputShowResourceJSON outputs a resource in JSON format.
func outputShowResourceJSON(c *cobra.Command, lib *library.Library, ref string) error {
	typ, name, err := library.ParseRef(ref)
	if err != nil {
		return fmt.Errorf("parsing ref %q: %w", ref, err)
	}

	resources, ok := lib.Resources[typ]
	if !ok {
		return fmt.Errorf("resource not found: %s", ref)
	}

	res, ok := resources[name]
	if !ok {
		return fmt.Errorf("resource not found: %s", ref)
	}

	output := ShowResourceJSONOutput{
		Ref:         ref,
		Path:        res.Path,
		Description: res.Description,
	}

	encoder := json.NewEncoder(c.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}

// ShowPresetJSONOutput represents the JSON output for a single preset.
type ShowPresetJSONOutput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Resources   []string `json:"resources"`
}

// outputShowPresetJSON outputs a preset in JSON format.
func outputShowPresetJSON(c *cobra.Command, lib *library.Library, name string) error {
	preset, ok := lib.Presets[name]
	if !ok {
		return fmt.Errorf("preset not found: %s", name)
	}

	output := ShowPresetJSONOutput{
		Name:        preset.Name,
		Description: preset.Description,
		Resources:   preset.Resources,
	}

	encoder := json.NewEncoder(c.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}

// FormatBatchAddSummary formats and outputs the batch add summary.
func FormatBatchAddSummary(c *cobra.Command, result *library.BatchAddResult) {
	if result == nil {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "No resources processed.")
		return
	}

	// Output each category if non-empty
	if len(result.Added) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nAdded:")
		for _, added := range result.Added {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s\n", added.Ref)
		}
	}

	if len(result.Skipped) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nSkipped:")
		for _, skipped := range result.Skipped {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s (%s)\n", skipped.Source, skipped.Issue)
		}
	}

	if len(result.Failed) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nFailed:")
		for _, failed := range result.Failed {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s: %s\n", failed.Source, failed.Error)
		}
	}

	// Output summary
	_, _ = fmt.Fprintf(c.OutOrStdout(), "\nAdded %d, skipped %d, failed %d\n",
		result.Summary.Added, result.Summary.Skipped, result.Summary.Failed)
}

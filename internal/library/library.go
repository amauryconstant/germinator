// Package library provides library management for canonical resources.
package library

import (
	"fmt"
	"strings"
)

// ResourceType represents the type of a library resource.
type ResourceType string

const (
	ResourceTypeSkill   ResourceType = "skill"
	ResourceTypeAgent   ResourceType = "agent"
	ResourceTypeCommand ResourceType = "command"
	ResourceTypeMemory  ResourceType = "memory"
)

// ValidResourceTypes contains all valid resource types.
var ValidResourceTypes = []ResourceType{
	ResourceTypeSkill,
	ResourceTypeAgent,
	ResourceTypeCommand,
	ResourceTypeMemory,
}

// IsValid checks if the resource type is valid.
func (rt ResourceType) IsValid() bool {
	for _, t := range ValidResourceTypes {
		if rt == t {
			return true
		}
	}
	return false
}

// String returns the string representation of the resource type.
func (rt ResourceType) String() string {
	return string(rt)
}

// Resource represents a single library resource entry.
type Resource struct {
	// Path is the relative path to the resource file from the library root.
	Path string `yaml:"path"`
	// Description is a human-readable description of the resource.
	Description string `yaml:"description"`
}

// Validate checks if the resource has valid fields.
func (r *Resource) Validate() error {
	if r.Path == "" {
		return fmt.Errorf("resource path is required")
	}
	if strings.TrimSpace(r.Path) == "" {
		return fmt.Errorf("resource path cannot be whitespace only")
	}
	return nil
}

// Preset represents a named collection of resource references.
type Preset struct {
	// Name is the preset identifier.
	Name string `yaml:"name"`
	// Description is a human-readable description of the preset.
	Description string `yaml:"description"`
	// Resources is a list of resource references in "type/name" format.
	Resources []string `yaml:"resources"`
}

// Validate checks if the preset has valid fields.
func (p *Preset) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("preset name is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("preset name cannot be whitespace only")
	}
	if len(p.Resources) == 0 {
		return fmt.Errorf("preset %q must have at least one resource", p.Name)
	}
	for _, ref := range p.Resources {
		if _, _, err := ParseRef(ref); err != nil {
			return fmt.Errorf("preset %q has invalid resource reference %q: %w", p.Name, ref, err)
		}
	}
	return nil
}

// Library represents the library index with resources and presets.
type Library struct {
	// Version is the library format version.
	Version string `yaml:"version"`
	// RootPath is the absolute path to the library directory.
	RootPath string `yaml:"-"`
	// Resources maps resource type to name to resource entry.
	// Structure: Resources["skill"]["commit"] = Resource{Path: "skills/commit.yaml", ...}
	Resources map[string]map[string]Resource `yaml:"resources"`
	// Presets maps preset name to preset definition.
	Presets map[string]Preset `yaml:"presets"`
}

// ParseRef parses a resource reference in "type/name" format.
func ParseRef(ref string) (typ, name string, err error) {
	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource reference format: %q (expected type/name)", ref)
	}
	typ, name = parts[0], parts[1]
	if typ == "" || name == "" {
		return "", "", fmt.Errorf("invalid resource reference format: %q (type and name cannot be empty)", ref)
	}
	return typ, name, nil
}

// FormatRef creates a resource reference from type and name.
func FormatRef(typ, name string) string {
	return fmt.Sprintf("%s/%s", typ, name)
}

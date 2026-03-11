package library

import (
	"fmt"
	"path/filepath"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

// ResolveResource resolves a resource reference to an absolute file path.
// The ref must be in "type/name" format (e.g., "skill/commit").
func ResolveResource(lib *Library, ref string) (string, error) {
	typ, name, err := ParseRef(ref)
	if err != nil {
		return "", err
	}

	resources, ok := lib.Resources[typ]
	if !ok {
		return "", gerrors.NewFileError(ref, "resolve", "resource not found", nil)
	}

	res, ok := resources[name]
	if !ok {
		return "", gerrors.NewFileError(ref, "resolve", "resource not found", nil)
	}

	return filepath.Join(lib.RootPath, res.Path), nil
}

// ResolveResources resolves multiple resource references to absolute file paths.
// Returns an error on the first failure (fail-fast).
func ResolveResources(lib *Library, refs []string) ([]string, error) {
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		path, err := ResolveResource(lib, ref)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// ResolvePreset resolves a preset name to a list of resource references.
// Returns references in "type/name" format.
func ResolvePreset(lib *Library, name string) ([]string, error) {
	preset, ok := lib.Presets[name]
	if !ok {
		return nil, gerrors.NewConfigError("preset", name, "preset not found")
	}
	return preset.Resources, nil
}

// OutputPathConfig holds configuration for output path derivation.
type OutputPathConfig struct {
	// Directory is the base directory name (e.g., ".opencode")
	Directory string
	// Subdirectory is the resource type subdirectory (e.g., "skills", "agents")
	Subdirectory string
	// FileSuffix is the suffix for the output file (e.g., "/SKILL.md" or ".md")
	FileSuffix string
	// UseSubdirectory indicates if the resource should be in a subdirectory
	UseSubdirectory bool
}

// PlatformOutputPaths maps platforms to their output path configurations by resource type.
var PlatformOutputPaths = map[string]map[ResourceType]OutputPathConfig{
	"opencode": {
		ResourceTypeSkill:   {Directory: ".opencode", Subdirectory: "skills", FileSuffix: "/SKILL.md", UseSubdirectory: true},
		ResourceTypeAgent:   {Directory: ".opencode", Subdirectory: "agents", FileSuffix: ".md", UseSubdirectory: false},
		ResourceTypeCommand: {Directory: ".opencode", Subdirectory: "commands", FileSuffix: ".md", UseSubdirectory: false},
		ResourceTypeMemory:  {Directory: ".opencode", Subdirectory: "memory", FileSuffix: ".md", UseSubdirectory: false},
	},
	"claude-code": {
		ResourceTypeSkill:   {Directory: ".claude", Subdirectory: "skills", FileSuffix: "/SKILL.md", UseSubdirectory: true},
		ResourceTypeAgent:   {Directory: ".claude", Subdirectory: "agents", FileSuffix: ".md", UseSubdirectory: false},
		ResourceTypeCommand: {Directory: ".claude", Subdirectory: "commands", FileSuffix: ".md", UseSubdirectory: false},
		ResourceTypeMemory:  {Directory: ".claude", Subdirectory: "memory", FileSuffix: ".md", UseSubdirectory: false},
	},
}

// GetOutputPath returns the platform-specific output path for a resource.
// The outputDir is the base directory (e.g., "." for current directory).
func GetOutputPath(typ, name, platform, outputDir string) (string, error) {
	resourceType := ResourceType(typ)
	if !resourceType.IsValid() {
		return "", gerrors.NewConfigError("resource-type", typ, "invalid resource type")
	}

	platformConfigs, ok := PlatformOutputPaths[platform]
	if !ok {
		return "", gerrors.NewConfigError("platform", platform, "unknown platform")
	}

	config, ok := platformConfigs[resourceType]
	if !ok {
		return "", fmt.Errorf("unsupported resource type for platform: %s/%s", typ, platform)
	}

	var path string
	if config.UseSubdirectory {
		// e.g., .opencode/skills/commit/SKILL.md
		path = filepath.Join(outputDir, config.Directory, config.Subdirectory, name, config.FileSuffix)
	} else {
		// e.g., .opencode/agents/commit.md
		path = filepath.Join(outputDir, config.Directory, config.Subdirectory, name+config.FileSuffix)
	}

	return path, nil
}

// GetOutputPaths returns all output paths for a list of resource references.
func GetOutputPaths(lib *Library, refs []string, platform, outputDir string) (map[string]string, error) {
	paths := make(map[string]string, len(refs))
	for _, ref := range refs {
		typ, name, err := ParseRef(ref)
		if err != nil {
			return nil, err
		}
		path, err := GetOutputPath(typ, name, platform, outputDir)
		if err != nil {
			return nil, err
		}
		paths[ref] = path
	}
	return paths, nil
}

// IsValidPlatform checks if the platform is supported.
func IsValidPlatform(platform string) bool {
	_, ok := PlatformOutputPaths[platform]
	return ok
}

// ValidPlatforms returns the list of valid platforms.
func ValidPlatforms() []string {
	platforms := make([]string, 0, len(PlatformOutputPaths))
	for p := range PlatformOutputPaths {
		platforms = append(platforms, p)
	}
	return platforms
}

// ValidateRef validates a resource reference format and checks if the type is valid.
func ValidateRef(ref string) error {
	typ, name, err := ParseRef(ref)
	if err != nil {
		return err
	}

	resourceType := ResourceType(typ)
	if !resourceType.IsValid() {
		validTypes := make([]string, len(ValidResourceTypes))
		for i, t := range ValidResourceTypes {
			validTypes[i] = string(t)
		}
		return fmt.Errorf("invalid resource type: %s (valid types: %s)", typ, strings.Join(validTypes, ", "))
	}

	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("resource name cannot be empty or whitespace")
	}

	return nil
}

package library

import (
	"context"
	"path/filepath"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

// ResolveResource resolves a resource reference to an absolute file path.
// The ref must be in "type/name" format (e.g., "skill/commit"). On a
// miss (unknown type or unknown name), returns *core.NotFoundError so
// cmdutil.ExitCodeFor maps the failure to ExitCodeError (1) — a
// runtime lookup miss is an operational error, not a user-input
// validation error.
func ResolveResource(lib *Library, ref string) (string, error) {
	typ, name, err := ParseRef(ref)
	if err != nil {
		return "", err
	}

	resources, ok := lib.Resources[typ]
	if !ok {
		return "", gerrors.NewNotFoundError("resource", ref)
	}

	res, ok := resources[name]
	if !ok {
		return "", gerrors.NewNotFoundError("resource", ref)
	}

	return filepath.Join(lib.RootPath, res.Path), nil
}

// ResolveResourceEntry resolves a resource reference to the canonical
// *Resource entry, returning *core.NotFoundError on miss (entity
// "resource", key = ref). This is the entry-point shape used by
// callers that need the full Resource struct (description, path)
// rather than the joined filesystem path returned by ResolveResource.
// The ref must be in "type/name" format (e.g., "skill/commit").
//
// This helper coexists with ResolveResource by design: the path-only
// form is the right tool for callers that only need a filesystem
// destination, while the entry form is the right tool for callers
// that need to render the resource's metadata. Returning the
// *Resource directly avoids leaking the dual map-lookup into every
// cmd-layer consumer (the resolver owns the lookup; the cmd layer
// renders).
func ResolveResourceEntry(lib *Library, ref string) (*Resource, error) {
	typ, name, err := ParseRef(ref)
	if err != nil {
		return nil, err
	}

	resources, ok := lib.Resources[typ]
	if !ok {
		return nil, gerrors.NewNotFoundError("resource", ref)
	}

	res, ok := resources[name]
	if !ok {
		return nil, gerrors.NewNotFoundError("resource", ref)
	}

	return &res, nil
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

// ResolvePreset resolves a preset name to a list of resource
// references using the receiver library. Returns references in
// "type/name" format. The ctx parameter is accepted for parity with
// future context-aware preset loading (e.g., for cancellation) but is
// unused in the current implementation. Resolution is a pure in-memory
// map lookup; if the resolution path is extended to perform I/O in the
// future, ctx SHALL be forwarded to that I/O (per cli-framework/spec.md
// accept-and-may-ignore pattern).
//
// On miss, returns *core.NotFoundError{Entity: "preset", Key: name} so
// cmdutil.ExitCodeFor maps the failure to ExitCodeError (1) — a
// runtime lookup miss is an operational error, not a user-input
// validation error. cmd/init's runInit pass-through uses this typed
// error directly without re-wrap.
func (lib *Library) ResolvePreset(ctx context.Context, name string) ([]string, error) {
	_ = ctx // accept-and-may-ignore: pure in-memory lookup, no I/O to forward to today
	preset, ok := lib.Presets[name]
	if !ok {
		return nil, gerrors.NewNotFoundError("preset", name)
	}
	return preset.Resources, nil
}

// ResolvePresetEntry resolves a preset name to the canonical *Preset
// entry, returning *core.NotFoundError on miss (entity "preset",
// key = name). This is the entry-point shape used by callers that
// need the full Preset struct (description, resource list) rather
// than the bare resource-ref slice returned by ResolvePreset.
//
// This helper coexists with (*Library).ResolvePreset by design: the
// ref-slice form is the right tool for callers that only need to
// iterate refs (init), while the entry form is the right tool for
// callers that need to render the preset's metadata (show). Returning
// the *Preset directly avoids leaking the dual map-lookup into every
// cmd-layer consumer.
func ResolvePresetEntry(lib *Library, name string) (*Preset, error) {
	preset, ok := lib.Presets[name]
	if !ok {
		return nil, gerrors.NewNotFoundError("preset", name)
	}
	return &preset, nil
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
		return "", gerrors.NewConfigError("resource-type", typ, "unsupported resource type for platform")
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
func GetOutputPaths(_ *Library, refs []string, platform, outputDir string) (map[string]string, error) {
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
		return gerrors.NewConfigError("resource-type", typ, "invalid resource type").WithSuggestions(validTypes)
	}

	if strings.TrimSpace(name) == "" {
		return gerrors.NewValidationError(ref, "name", name, "resource name cannot be empty or whitespace")
	}

	return nil
}

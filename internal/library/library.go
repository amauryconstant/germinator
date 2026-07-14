package library

// Package library provides library management for canonical resources.

import (
	"context"
	"fmt"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

// ResourceType represents the type of a library resource.
type ResourceType string

// ResourceType constants define the type of resource in the library.
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
		return gerrors.NewValidationError("", "path", "", "resource path is required")
	}
	if strings.TrimSpace(r.Path) == "" {
		return gerrors.NewValidationError("", "path", "", "resource path cannot be whitespace only")
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
		return gerrors.NewValidationError("", "name", "", "preset name is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return gerrors.NewValidationError("", "name", "", "preset name cannot be whitespace only")
	}
	if len(p.Resources) == 0 {
		return gerrors.NewValidationError("", "resources", "", "preset must have at least one resource")
	}
	for _, ref := range p.Resources {
		if _, _, err := ParseRef(ref); err != nil {
			return gerrors.NewValidationError("", "resources", ref, "invalid resource reference in preset").WithContext(p.Name)
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
		return "", "", gerrors.NewConfigError("reference", ref, "invalid resource reference format (expected type/name)")
	}
	typ, name = parts[0], parts[1]
	if typ == "" || name == "" {
		return "", "", gerrors.NewConfigError("reference", ref, "invalid resource reference format (type and name cannot be empty)")
	}
	return typ, name, nil
}

// FormatRef creates a resource reference from type and name.
func FormatRef(typ, name string) string {
	return fmt.Sprintf("%s/%s", typ, name)
}

// Refresh is the method form of the package-level RefreshLibrary
// function. It mirrors the package-level function so *library.Library
// can satisfy the cmd-side refresherLibrary interface without an
// adapter shim, matching the slice-6 (*Library).CreatePreset precedent
// at internal/library/creator.go:145.
//
// The method receiver is required because the in-memory library
// state (lib.Resources) is mutated during refresh; subsequent
// SaveLibrary persists the mutations to library.yaml. The method
// checks ctx.Err() at entry to honor caller-supplied cancellation
// before any I/O, and asserts that lib is non-nil with a
// non-empty RootPath so the loader step inside RefreshLibrary
// targets a real on-disk library.
//
// On success returns a *RefreshResult with the Refreshed,
// Unchanged, Skipped, and Errors slices populated per design
// Decision 7.
func (lib *Library) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("refresh library: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return nil, gerrors.NewValidationError("library refresh", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	return RefreshLibrary(ctx, RefreshOptions{
		LibraryPath: lib.RootPath,
		DryRun:      req.DryRun,
		Force:       req.Force,
	})
}

// RemoveResource is the method form of the package-level RemoveResource
// function. It mirrors the package-level function so *library.Library
// can satisfy the cmd-side removerLibrary interface without an
// adapter shim, matching the slice-6 (*Library).CreatePreset precedent
// at internal/library/creator.go:145.
//
// The method receiver carries the loaded library (and its RootPath).
// ctx.Err() is checked at entry to honor caller-supplied cancellation
// before any I/O; the receiver is asserted non-nil with a non-empty
// RootPath so the loader step inside RemoveResource targets a real
// on-disk library.
func (lib *Library) RemoveResource(ctx context.Context, req *RemoveResourceRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("remove resource: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return gerrors.NewValidationError("library remove resource", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	if _, err := RemoveResource(ctx, RemoveResourceOptions{
		Ref:         req.Ref,
		LibraryPath: lib.RootPath,
	}); err != nil {
		return fmt.Errorf("removing resource: %w", err)
	}
	return nil
}

// RemovePreset is the method form of the package-level RemovePreset
// function. It mirrors the package-level function so *library.Library
// can satisfy the cmd-side removerLibrary interface without an
// adapter shim, matching the slice-6 (*Library).CreatePreset precedent
// at internal/library/creator.go:145.
//
// The method receiver carries the loaded library (and its RootPath).
// ctx.Err() is checked at entry to honor caller-supplied cancellation
// before any I/O; the receiver is asserted non-nil with a non-empty
// RootPath so the loader step inside RemovePreset targets a real
// on-disk library.
func (lib *Library) RemovePreset(ctx context.Context, req *RemovePresetRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("remove preset: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return gerrors.NewValidationError("library remove preset", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	if _, err := RemovePreset(ctx, RemovePresetOptions{
		Name:        req.Name,
		LibraryPath: lib.RootPath,
	}); err != nil {
		return fmt.Errorf("removing preset: %w", err)
	}
	return nil
}

// Validate is the method form of the package-level ValidateLibrary
// function. It mirrors the package-level function so *library.Library
// can satisfy the cmd-side validatorLibrary interface without an
// adapter shim, matching the slice-6 (*Library).CreatePreset precedent
// at internal/library/creator.go:145.
//
// When req.Fix is true and the validation scan finds error-level
// issues, Validate internally calls lib.Fix(ctx, &FixRequest{}) and
// merges the resulting *FixResult into the returned *ValidationResult
// (FixApplied = true, FixResult populated) so the cmd layer can
// surface the fix report via --output json / --output table without a
// second call. ctx.Err() is checked at entry to honor caller-supplied
// cancellation before any I/O; the receiver is asserted non-nil with
// a non-empty RootPath so the loader step inside ValidateLibrary
// targets a real on-disk library.
func (lib *Library) Validate(ctx context.Context, req *ValidateRequest) (*ValidationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("validate library: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return nil, gerrors.NewValidationError("library validate", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}

	result, err := ValidateLibrary(lib)
	if err != nil {
		return nil, fmt.Errorf("validating library: %w", err)
	}

	if req != nil && req.Fix && !result.Valid {
		fixResult, fixErr := lib.Fix(ctx, &FixRequest{})
		if fixErr != nil {
			return nil, fmt.Errorf("fixing library: %w", fixErr)
		}
		result.FixApplied = true
		result.FixResult = fixResult
	}

	return result, nil
}

// Fix is the method form of the package-level FixLibrary function. It
// mirrors the package-level function so *library.Library can satisfy
// the cmd-side validatorLibrary interface (via the embedded
// (*Library).Validate -> (*Library).Fix chain) without an adapter
// shim, matching the slice-6 (*Library).CreatePreset precedent at
// internal/library/creator.go:145.
//
// FixRequest is currently empty; the method operates on the live
// library carried by the *Library receiver. ctx.Err() is checked at
// entry to honor caller-supplied cancellation before any I/O; the
// receiver is asserted non-nil with a non-empty RootPath so
// FixLibrary targets a real on-disk library.
func (lib *Library) Fix(ctx context.Context, _ *FixRequest) (*FixResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("fix library: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return nil, gerrors.NewValidationError("library fix", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	return FixLibrary(lib)
}

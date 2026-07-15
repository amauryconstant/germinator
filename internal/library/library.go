package library

// Package library provides library management for canonical resources.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

// Add is the method form of the package-level AddResource function.
// It mirrors the package-level function so *library.Library can
// satisfy the cmd-side adderLibrary interface without an adapter
// shim, matching the slice-7 (*Library).CreatePreset /
// (*Library).Refresh precedent at internal/library/creator.go:145
// and internal/library/library.go:143.
//
// The method receiver carries the loaded library (and its RootPath),
// so the existing-resource check reads lib.Resources directly instead
// of re-loading from disk. On a successful mutation the receiver's
// in-memory Resources map is updated alongside the on-disk YAML so
// callers see the new resource via subsequent lib.Resources lookups
// without an extra LoadLibrary call. ctx.Err() is checked at entry
// and between file-I/O steps so caller-supplied cancellation is
// honored before any work.
func (lib *Library) Add(ctx context.Context, req *AddRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return gerrors.NewValidationError("library add", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	if req == nil {
		return gerrors.NewValidationError("library add", "request", "",
			"add request must not be nil")
	}

	if err := validateSourceExists(req.Source); err != nil {
		return err
	}

	docType, err := detectType(req.Source, req.Type)
	if err != nil {
		return err
	}

	name, err := detectName(req.Source, req.Name)
	if err != nil {
		return err
	}

	description := detectDescription(req.Source, req.Description)

	resourceKey := FormatRef(docType, name)
	existingTypeMap, typeExists := lib.Resources[docType]
	_, nameExists := existingTypeMap[name]
	if typeExists && nameExists && !req.Force {
		return gerrors.NewFileError(lib.RootPath, "add",
			fmt.Sprintf("resource %s already exists (use --force to overwrite)", resourceKey), nil)
	}

	if req.DryRun {
		if req.Stdout != nil {
			_, _ = fmt.Fprintln(req.Stdout, "Would add resource:", resourceKey)
			_, _ = fmt.Fprintln(req.Stdout, "  Type:", docType)
			_, _ = fmt.Fprintln(req.Stdout, "  Name:", name)
			_, _ = fmt.Fprintln(req.Stdout, "  Description:", description)
			_, _ = fmt.Fprintln(req.Stdout, "  Source:", req.Source)
		}
		return nil
	}

	targetDir := filepath.Join(lib.RootPath, docType+"s")
	targetFile := filepath.Join(targetDir, name+".md")

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return gerrors.NewFileError(targetDir, "create", "failed to create resource directory", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}

	content, err := os.ReadFile(req.Source)
	if err != nil {
		return gerrors.NewFileError(req.Source, "read", "failed to read source file", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("adding resource: %w", err)
	}

	if err := os.WriteFile(targetFile, content, 0o644); err != nil { //nolint:gosec // G302: standard file permissions
		return gerrors.NewFileError(targetFile, "write", "failed to write resource file", err)
	}

	if err := addResourceToLibrary(lib.RootPath, docType, name, targetFile, description); err != nil {
		return fmt.Errorf("updating library.yaml: %w", err)
	}

	// Update the receiver's in-memory map so callers see the new
	// resource via subsequent lib.Resources lookups without an extra
	// LoadLibrary call. The on-disk mutation is authoritative; this
	// keeps the in-memory view consistent.
	if lib.Resources == nil {
		lib.Resources = make(map[string]map[string]Resource)
	}
	if lib.Resources[docType] == nil {
		lib.Resources[docType] = make(map[string]Resource)
	}
	relPath, relErr := filepath.Rel(lib.RootPath, targetFile)
	if relErr != nil {
		return fmt.Errorf("computing relative path: %w", relErr)
	}
	lib.Resources[docType][name] = Resource{
		Path:        relPath,
		Description: description,
	}

	if _, err := LoadLibrary(ctx, lib.RootPath); err != nil {
		return fmt.Errorf("validating updated library: %w", err)
	}

	return nil
}

// BatchAddResources is the method form of the package-level
// BatchAddResources function. It mirrors the package-level function
// so *library.Library can satisfy the cmd-side adderLibrary
// interface without an adapter shim, matching the slice-7
// (*Library).CreatePreset precedent at internal/library/creator.go:145.
//
// The method receiver carries the loaded library; per-file
// registration delegates to lib.Add so the receiver's RootPath and
// in-memory Resources map are reused instead of re-loading for each
// source. ctx.Err() is checked at entry and between files so caller-
// supplied cancellation is honored before the next iteration.
func (lib *Library) BatchAddResources(ctx context.Context, opts *BatchAddOptions) (*BatchAddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("batch add: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return nil, gerrors.NewValidationError("library batch add", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	if opts == nil {
		return nil, gerrors.NewValidationError("library batch add", "request", "",
			"batch add options must not be nil")
	}

	result := &BatchAddResult{}

	orphanMap := make(map[string]Orphan, len(opts.Orphans))
	for _, orphan := range opts.Orphans {
		orphanMap[orphan.Path] = orphan
	}

	files, err := collectSourceFiles(opts.Sources)
	if err != nil {
		return nil, fmt.Errorf("collecting source files: %w", err)
	}
	result.Summary.Total = len(files)

	for _, source := range files {
		if cerr := ctx.Err(); cerr != nil {
			return result, fmt.Errorf("batch add: %w", cerr)
		}
		var orphan Orphan
		if o, ok := orphanMap[source]; ok {
			orphan = o
		}
		_ = processBatchAddFile(ctx, lib, source, opts, result, orphan)
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("batch add: %w", err)
	}

	return result, nil
}

// DiscoverOrphans is the method form of the package-level
// DiscoverOrphans function. It mirrors the package-level function
// so *library.Library can satisfy the cmd-side adderLibrary
// interface without an adapter shim, matching the slice-7
// (*Library).CreatePreset precedent at internal/library/creator.go:145.
//
// The method receiver carries the loaded library; the scanner uses
// lib.Resources directly instead of re-loading from disk. ctx.Err()
// is checked at entry and between scanned directories so caller-
// supplied cancellation is honored before the next subtree walk.
func (lib *Library) DiscoverOrphans(ctx context.Context, opts *DiscoverOptions) (*DiscoverResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("discovering orphans: %w", err)
	}
	if lib == nil || lib.RootPath == "" {
		return nil, gerrors.NewValidationError("library discover orphans", "rootPath", "",
			"library is not loaded (RootPath is empty)")
	}
	if opts == nil {
		return nil, gerrors.NewValidationError("library discover orphans", "request", "",
			"discover options must not be nil")
	}

	result := &DiscoverResult{}

	directories := map[string]string{
		"skills":   "skill",
		"agents":   "agent",
		"commands": "command",
		"memory":   "memory",
	}

	for dir, resType := range directories {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("discovering orphans: %w", err)
		}
		dirPath := filepath.Join(lib.RootPath, dir)
		if err := scanDirectory(ctx, dirPath, resType, lib, *opts, result); err != nil {
			return result, err
		}
	}

	result.Summary.TotalOrphans = len(result.Orphans)

	if !opts.Batch && opts.Force && len(result.Conflicts) == 0 && !opts.DryRun {
		for _, orphan := range result.Orphans {
			if err := ctx.Err(); err != nil {
				return result, fmt.Errorf("discovering orphans: %w", err)
			}
			if err := registerOrphan(lib.RootPath, orphan); err != nil {
				return nil, err
			}
			result.Added = append(result.Added, AddSuccess{
				Type: orphan.Type,
				Name: orphan.Name,
				Path: orphan.Path,
			})
			result.Summary.TotalAdded++
		}
	}

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("discovering orphans: %w", err)
	}

	return result, nil
}

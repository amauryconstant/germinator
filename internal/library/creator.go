package library

// Package library provides library management for canonical resources.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

// CreateOptions contains options for library creation.
type CreateOptions struct {
	// Path is the directory path where the library will be created.
	Path string
	// DryRun preview changes without creating files or directories.
	DryRun bool
	// Force overwrites an existing library at the target path.
	Force bool
	// Stdout receives dry-run output (typically opts.IO.Out from the
	// cmd layer). Optional: nil means "no dry-run output" so tests
	// can construct CreateOptions{} without a writer.
	Stdout io.Writer
}

// CreateLibrary creates a new library directory structure at the specified path.
// It creates library.yaml and empty resource directories (skills, agents, commands, memory).
// If DryRun is true, it prints what would be created without making changes.
// If Force is false and a library already exists at Path, an error is returned.
func CreateLibrary(opts CreateOptions) error {
	// Check if library already exists
	exists := Exists(opts.Path)
	if exists && !opts.Force {
		return gerrors.NewFileError(opts.Path, "create", "library already exists at path (use --force to overwrite)", nil)
	}

	// Dry run mode - print what would be created
	if opts.DryRun {
		if opts.Stdout != nil {
			_, _ = fmt.Fprintln(opts.Stdout, "Would create library at:", opts.Path)
			_, _ = fmt.Fprintln(opts.Stdout, "  -", filepath.Join(opts.Path, "library.yaml"))
			_, _ = fmt.Fprintln(opts.Stdout, "  -", filepath.Join(opts.Path, "skills")+"/")
			_, _ = fmt.Fprintln(opts.Stdout, "  -", filepath.Join(opts.Path, "agents")+"/")
			_, _ = fmt.Fprintln(opts.Stdout, "  -", filepath.Join(opts.Path, "commands")+"/")
			_, _ = fmt.Fprintln(opts.Stdout, "  -", filepath.Join(opts.Path, "memory")+"/")
		}
		return nil
	}

	// Create directory structure
	dirs := []string{
		opts.Path,
		filepath.Join(opts.Path, "skills"),
		filepath.Join(opts.Path, "agents"),
		filepath.Join(opts.Path, "commands"),
		filepath.Join(opts.Path, "memory"),
	}

	for _, dir := range dirs {
		// Unix permission bits (0o750) are no-ops on Windows;
		// Windows support is out of scope.
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return gerrors.NewFileError(dir, "create", "failed to create directory", err)
		}
	}

	// Create library.yaml
	yamlContent := defaultLibraryYAML()
	yamlPath := filepath.Join(opts.Path, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644); err != nil { //nolint:gosec // G302: Creating new file with standard permissions
		return gerrors.NewFileError(yamlPath, "write", "failed to write library.yaml", err)
	}

	// Validate created library by loading it
	// TODO(slice-7): replace with caller context (c.Context() in runF wiring).
	ctx := context.Background()
	if _, err := LoadLibrary(ctx, opts.Path); err != nil {
		// Validation failed - leave partial structure for debugging
		return fmt.Errorf("library created but validation failed: %w (partial structure left for debugging)", err)
	}

	return nil
}

// Init is the package-level entry point for `library init` (slice 7
// forward path). It is a thin adapter that maps *InitRequest to
// CreateOptions and delegates to CreateLibrary.
//
// Init stays a package function (not a method on *Library) per
// design Decision 6: init creates a fresh library, so there is no
// pre-existing *Library to receive a method. The cmd layer's
// runLibraryInit calls Init directly without an interface or adapter
// shim, matching the CreatePreset / (*Library).CreatePreset dual
// form at internal/library/creator.go:127.
//
// ctx is checked at entry to honor caller-supplied cancellation
// before any I/O.
func Init(ctx context.Context, req *InitRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("init library: %w", err)
	}
	if req == nil {
		return gerrors.NewValidationError("library init", "request", "",
			"init request must not be nil")
	}
	if err := CreateLibrary(CreateOptions{
		Path:   req.Path,
		DryRun: req.DryRun,
		Force:  req.Force,
		Stdout: req.Stdout,
	}); err != nil {
		return fmt.Errorf("creating library: %w", err)
	}
	return nil
}

// defaultLibraryYAML returns the default library.yaml content.
func defaultLibraryYAML() string {
	return `version: "1"
resources:
  skill: {}
  agent: {}
  command: {}
  memory: {}
presets: {}
`
}

// CreatePresetRequest contains the parameters for CreatePreset.
//
// Name is the preset identifier (required). Description is optional
// human-readable text. Resources is the slice of "type/name" refs the
// preset bundles; each ref must parse via ParseRef and resolve to a
// registered resource in lib. Force bypasses the existence check so an
// existing preset can be overwritten.
type CreatePresetRequest struct {
	Name        string
	Description string
	Resources   []string
	Force       bool
}

// CreatePreset adds or replaces a preset in lib. It centralizes the
// validation + mutation steps the cmd layer used to perform inline.
//
// Pre-flight:
//   - ctx.Err() guard (caller-supplied cancellation is honored before
//     any I/O).
//   - Name is required and non-empty after trimming.
//   - Resources must contain at least one ref; each ref must parse
//     via ParseRef and resolve to a registered resource under
//     lib.Resources.
//
// Mutation:
//   - PresetExists + !Force returns *gerrors.ValidationError so the
//     cmd layer maps it to exit 1 via the default-error case in
//     cmdutil.ExitCodeFor.
//   - AddPreset applies lib.Preset.Validate which enforces the same
//     non-empty Resources constraint as a defensive check.
//   - SaveLibrary persists the in-memory mutation to disk.
//
// Returns wrapped errors using fmt.Errorf("%w", err) so callers can
// errors.Is / errors.As the typed core errors without losing context.
func CreatePreset(ctx context.Context, lib *Library, req *CreatePresetRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("create preset: %w", err)
	}
	if req == nil {
		return gerrors.NewValidationError("library", "request", "", "create preset request must not be nil")
	}
	return lib.CreatePreset(ctx, req)
}

// CreatePreset is the method form of the preset creation routine.
// It mirrors the package-level CreatePreset so *library.Library can
// satisfy the cmd-side presetWriter interface without an adapter
// shim (matching the slice-5 ResolvePreset dual form).
//
// The method receiver is required because preset creation is a
// mutating operation against the live in-memory library state;
// subsequent SaveLibrary persists the mutation to library.yaml.
func (lib *Library) CreatePreset(ctx context.Context, req *CreatePresetRequest) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("create preset: %w", err)
	}
	if req == nil {
		return gerrors.NewValidationError("library", "request", "", "create preset request must not be nil")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return gerrors.NewValidationError("library create preset", "name", req.Name, "preset name cannot be empty or whitespace")
	}

	if len(req.Resources) == 0 {
		return gerrors.NewValidationError("library create preset", "resources", "", "preset must reference at least one resource")
	}

	// Authoritative validation: each ref must parse and resolve against
	// the live library. core.CanInstallResource is the cmd-layer
	// pre-flight (string-only) check; this is the library-layer
	// authoritative check (must exist on disk + in lib.Resources).
	for _, ref := range req.Resources {
		typ, resName, parseErr := ParseRef(ref)
		if parseErr != nil {
			return gerrors.NewValidationError("library create preset", "resources", ref, "invalid resource reference").WithContext(parseErr.Error())
		}

		typeResources, ok := lib.Resources[typ]
		if !ok {
			return gerrors.NewNotFoundError("resource type", typ)
		}
		if _, exists := typeResources[resName]; !exists {
			return gerrors.NewNotFoundError("resource", ref)
		}
	}

	if PresetExists(lib, name) && !req.Force {
		return gerrors.NewValidationError("library create preset", "name", name, fmt.Sprintf("preset %q already exists (use --force to overwrite)", name))
	}

	preset := Preset{
		Name:        name,
		Description: req.Description,
		Resources:   append([]string(nil), req.Resources...),
	}

	if err := AddPreset(lib, preset); err != nil {
		return fmt.Errorf("adding preset: %w", err)
	}

	if err := SaveLibrary(lib); err != nil {
		return fmt.Errorf("saving library: %w", err)
	}

	return nil
}

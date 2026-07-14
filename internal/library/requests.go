package library

// Package library provides library management for canonical resources.
//
// This file defines the *Request / *Result types used by the
// (*Library) X method forms (slice 7 forward path) so *Library can
// satisfy the cmd-side interfaces declared in tasks 7.1-7.4 without an
// adapter shim. The package-level function forms in
// {refresher,remover,validator}.go preserve their existing public
// signatures and delegate to these methods internally.

// InitRequest contains the parameters for Init.
//
// Init creates a fresh library directory; there is no pre-existing
// *Library to receive a method, so Init stays a package function
// (design Decision 6) and uses this request type to match the
// CreatePreset / (*Library).CreatePreset dual form at
// internal/library/creator.go.
//
// The dry-run writer is forwarded separately via the
// library.Init(ctx, req, stdout) parameter per the package's
// io.Writer-parameter convention (see CreateLibrary at
// internal/library/creator.go); no writer field lives on
// InitRequest.
type InitRequest struct {
	// Path is the directory where the library will be created.
	Path string
	// DryRun previews the creation without writing any files.
	DryRun bool
	// Force overwrites an existing library at Path.
	Force bool
}

// RefreshRequest contains the parameters for (*Library).Refresh.
//
// Mirror of library.RefreshOptions (the package-level function form);
// the method form consumes this type so cmd-side code can pass
// pre-parsed flag values without rebuilding an options struct.
type RefreshRequest struct {
	// DryRun previews the refresh without writing library.yaml.
	DryRun bool
	// Force skips resources that would otherwise error (name
	// mismatch, malformed frontmatter) so the scan can continue.
	Force bool
}

// RemoveResourceRequest contains the parameters for
// (*Library).RemoveResource.
//
// Ref is the "type/name" reference of the resource to remove
// (parsed internally via ParseRef). Force bypasses the
// preset-reference safety check (the existing public RemoveResource
// function still calls os.Remove on the file).
type RemoveResourceRequest struct {
	// Ref is the resource reference in "type/name" format.
	Ref string
	// Force bypasses interactive confirmation prompts.
	Force bool
}

// RemovePresetRequest contains the parameters for
// (*Library).RemovePreset.
//
// Name is the preset identifier (required). Force bypasses
// interactive confirmation prompts.
type RemovePresetRequest struct {
	// Name is the preset identifier.
	Name string
	// Force bypasses interactive confirmation prompts.
	Force bool
}

// ValidateRequest contains the parameters for (*Library).Validate.
//
// Fix triggers the in-place auto-cleanup of library.yaml (removes
// missing file entries, strips ghost preset refs); when true, the
// returned *ValidationResult has FixApplied = true and FixResult
// populated with the *FixResult produced by (*Library).Fix.
type ValidateRequest struct {
	// Fix runs (*Library).Fix after the validation scan when true.
	Fix bool
}

// FixRequest contains the parameters for (*Library).Fix.
//
// Currently empty: fix operates on the live library carried by the
// *Library receiver and does not take caller-supplied inputs. The
// type is kept for forward-compatibility (e.g. a future --dry-run or
// --scope flag) and to match the dual-form convention used by
// CreatePreset / (*Library).CreatePreset.
type FixRequest struct{}

// RefreshUnchanged represents a resource that was scanned during a
// refresh and matched library.yaml exactly (no description drift, no
// path change, no conflict). LastSynced carries the file's mtime as
// an RFC3339 string when available, or the empty string otherwise
// (e.g. file is missing).
//
// Added per design Decision 7: the plain output renders an
// "Unchanged:" section listing these resources; the spec scenario
// "Unchanged resources reported" requires the section.
type RefreshUnchanged struct {
	// Ref is the resource reference in "type/name" format.
	Ref string
	// LastSynced is the file's modification time (RFC3339) when
	// available, or empty string if the file's mtime cannot be
	// determined.
	LastSynced string
}

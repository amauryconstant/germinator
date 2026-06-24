// Package application defines service interfaces and data transfer objects
// for the Germinator application layer.
package application

import "gitlab.com/amoconst/germinator/internal/library"

// TransformRequest contains the input parameters for document transformation.
type TransformRequest struct {
	// InputPath is the path to the source document.
	InputPath string
	// OutputPath is the path where the transformed document will be written.
	OutputPath string
	// Platform is the target platform (opencode or claude-code).
	Platform string
}

// ValidateRequest contains the input parameters for document validation.
type ValidateRequest struct {
	// InputPath is the path to the document to validate.
	InputPath string
	// Platform is the target platform for validation context.
	Platform string
}

// CanonicalizeRequest contains the input parameters for document canonicalization.
type CanonicalizeRequest struct {
	// InputPath is the path to the platform-specific document.
	InputPath string
	// OutputPath is the path where the canonical YAML will be written.
	OutputPath string
	// Platform is the source platform (opencode or claude-code).
	Platform string
	// DocType is the document type (agent, command, skill, or memory).
	DocType string
}

// InitializeRequest contains the input parameters for resource initialization.
type InitializeRequest struct {
	// Library is the loaded library containing resources.
	Library *library.Library
	// Platform is the target platform (opencode or claude-code).
	Platform string
	// OutputDir is the base output directory.
	OutputDir string
	// Refs are the resource references to initialize (e.g., "skill/commit").
	Refs []string
	// DryRun indicates whether to preview changes without writing.
	DryRun bool
	// Force indicates whether to overwrite existing files.
	Force bool
}

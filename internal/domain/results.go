package domain

// TransformResult contains the result of a document transformation.
type TransformResult struct {
	// OutputPath is the path where the transformed document was written.
	OutputPath string
}

// ValidateResult contains the result of document validation.
type ValidateResult struct {
	// Errors contains any validation errors found.
	// These are business-level validation issues, not fatal errors.
	Errors []error
}

// Valid returns true if no validation errors were found.
func (r *ValidateResult) Valid() bool {
	return len(r.Errors) == 0
}

// CanonicalizeResult contains the result of document canonicalization.
type CanonicalizeResult struct {
	// OutputPath is the path where the canonical YAML was written.
	OutputPath string
}

// InitializeResult contains the result of initializing a single resource.
type InitializeResult struct {
	// Ref is the resource reference (e.g., "skill/commit").
	Ref string
	// InputPath is the source file path.
	InputPath string
	// OutputPath is the destination file path.
	OutputPath string
	// Error is any error that occurred during initialization.
	Error error
}

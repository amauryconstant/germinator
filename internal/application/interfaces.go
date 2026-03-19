package application

import (
	"context"

	"gitlab.com/amoconst/germinator/internal/domain"
)

// Transformer handles document transformation from canonical to platform-specific format.
type Transformer interface {
	// Transform transforms a document to the target platform format.
	Transform(ctx context.Context, req *TransformRequest) (*domain.TransformResult, error)
}

// Validator handles document validation against platform-specific rules.
type Validator interface {
	// Validate validates a document and returns any validation errors.
	// The error return indicates a fatal error (couldn't validate).
	// Validation issues are returned in the result's Errors field.
	Validate(ctx context.Context, req *ValidateRequest) (*domain.ValidateResult, error)
}

// Canonicalizer handles conversion of platform documents to canonical format.
type Canonicalizer interface {
	// Canonicalize converts a platform document to canonical YAML format.
	Canonicalize(ctx context.Context, req *CanonicalizeRequest) (*domain.CanonicalizeResult, error)
}

// Initializer handles resource installation from the library.
type Initializer interface {
	// Initialize installs resources from the library to the target directory.
	Initialize(ctx context.Context, req *InitializeRequest) ([]domain.InitializeResult, error)
}

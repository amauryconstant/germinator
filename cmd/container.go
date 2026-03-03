package cmd

import (
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/services"
)

// ServiceContainer holds service instances for dependency injection.
// It is passed through the command tree via CommandConfig.
type ServiceContainer struct {
	// Transformer handles document transformation.
	Transformer application.Transformer
	// Validator handles document validation.
	Validator application.Validator
	// Canonicalizer handles document canonicalization.
	Canonicalizer application.Canonicalizer
	// Initializer handles resource initialization.
	Initializer application.Initializer
}

// NewServiceContainer creates a new ServiceContainer with all services initialized.
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		Transformer:   services.NewTransformer(),
		Validator:     services.NewValidator(),
		Canonicalizer: services.NewCanonicalizer(),
		Initializer:   services.NewInitializer(),
	}
}

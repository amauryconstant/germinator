package cmd

// ServiceContainer holds service instances for dependency injection.
// It is passed through the command tree via CommandConfig.
type ServiceContainer struct {
	// Future services will be added here as the application grows
	// Examples: Validator, Transformer, Canonicalizer, etc.
}

// NewServiceContainer creates a new ServiceContainer with all services initialized.
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{}
}

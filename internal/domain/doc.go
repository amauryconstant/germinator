// Package domain provides core business types and domain logic for the germinator application.
//
// This package consolidates domain types that were previously scattered across
// multiple packages (models/canonical, errors, validation, application).
//
// Domain Purity:
// This package has no external dependencies other than the Go standard library.
// This is enforced via depguard linting rules to prevent architectural drift.
//
// Contents:
//   - Agent, Command, Skill, Memory types (canonical document models)
//   - Domain error types with immutable builder pattern
//   - Result[T] functional error handling
//   - Validation pipeline and validators
//   - Service result types (TransformResult, ValidateResult, etc.)
//
// The domain layer is the foundation of the application and must remain
// independent of infrastructure concerns.
package domain

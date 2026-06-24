// Package core provides core business types and domain logic for the germinator application.
//
// This package consolidates domain types that were previously scattered across
// multiple packages (models/canonical, errors, validation, application).
//
// Core Purity:
// This package has no external dependencies other than the Go standard library
// and github.com/samber/lo.
// This is enforced via depguard linting rules to prevent architectural drift.
//
// Contents:
//   - Agent, Command, Skill, Memory types (canonical document models)
//   - Domain error types with immutable builder pattern
//   - Result[T] functional error handling
//   - Validation pipeline and validators
//   - Service result types (TransformResult, ValidateResult, etc.)
//   - Business rule functions (ValidatePlatform, ResolveOutputPath)
//
// The core layer is the foundation of the application and must remain
// independent of infrastructure concerns.
package core

// Note: design Decision 2 calls for a `type Domain = core` alias as a
// backward-compat shim for external consumers of the old
// internal/domain import path. That alias is not representable in Go
// (type aliases require a type on the right-hand side, not a package).
// The package is under internal/ visibility, so there are no external
// consumers in the Go sense, and the only in-tree references have been
// updated. The alias is therefore omitted; the design should be revised
// to drop this requirement.

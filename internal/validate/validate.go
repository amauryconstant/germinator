// Package validate provides document validation as an I/O shell-package
// service. The package exists to satisfy the imperative-shell boundary
// defined by the golang-cli-architecture skill: any code that performs
// filesystem reads and dispatches to functional-core validators lives
// here at the package edge, not in cmd/.
//
// The Service interface, Request type, and NewService constructor are
// the canonical contract. cmd/validate.go declares a local Validator
// interface that is structurally identical to validate.Service (same
// parameter / return types); *validateService satisfies both via
// structural typing so cmd/ does not have to import this package's
// adapter directly.
package validate

import (
	"context"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/core/opencode"
	"gitlab.com/amoconst/germinator/internal/parser"
)

// Request carries the inputs for document validation. Lifted from
// cmd/validate.go so the shell-package owns its types per the
// golang-cli-architecture "interfaces where consumed" rule (the
// consumer, cmd/validate.go, declares its Validator interface using
// this type via the shared import).
type Request struct {
	InputPath string
	Platform  string
}

// Service is the per-call contract for document validation. Returns
// a *core.ValidateResult; callers inspect result.Valid() to dispatch
// success vs. error rendering. Platform is assumed pre-validated by
// the caller (cmd/validate.go validates via core.ValidatePlatform
// before resolving the Service).
type Service interface {
	Validate(ctx context.Context, req *Request) (*core.ValidateResult, error)
}

// validateService is the production implementation. Zero-size because
// the orchestrating state (a parser + validators) is held by the
// functional core / opencode adapters — the service is a thin
// dispatch across document types.
type validateService struct{}

// Compile-time confirmation that *validateService satisfies the
// Service interface declared in this package.
var _ Service = (*validateService)(nil)

// NewService returns the production wiring for document validation.
// Mirrors canonicalize.NewService at internal/canonicalize/canonicalize.go.
func NewService() Service {
	return &validateService{}
}

// Validate implements Service. Orchestrates parse → validate across
// each canonical document type (agent / command / skill / memory) and,
// when the platform is opencode, adds the platform-specific validators
// from internal/core/opencode on top of the shared core validators.
//
// The errors returned from each validator are joined; we unwrap with
// unwrapJoinedErrors so the slice lives flat in *core.ValidateResult.Errors.
// Fatal errors (unrecognized doc type / parse failure) short-circuit
// and are returned as *core.ParseError so cmd/cmdutil.ExitCodeFor maps
// them to exit 1 via errors.As.
func (validateService) Validate(ctx context.Context, req *Request) (*core.ValidateResult, error) {
	docType := parser.DetectType(ctx, req.InputPath)
	if docType == "" {
		return nil, core.NewParseError(req.InputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := parser.ParseDocument(ctx, req.InputPath, docType)
	if parseErr != nil {
		return nil, core.NewParseError(req.InputPath, "failed to parse document", parseErr)
	}

	var errs []error

	switch d := doc.(type) {
	case *parser.CanonicalAgent:
		if result := core.ValidateAgent(&d.Agent); result.IsError() {
			errs = append(errs, unwrapJoinedErrors(result.Error)...)
		}
		if req.Platform == core.PlatformOpenCode {
			if result := opencode.ValidateAgentOpenCode(&d.Agent); result.IsError() {
				errs = append(errs, unwrapJoinedErrors(result.Error)...)
			}
		}
	case *parser.CanonicalCommand:
		if result := core.ValidateCommand(&d.Command); result.IsError() {
			errs = append(errs, unwrapJoinedErrors(result.Error)...)
		}
		if req.Platform == core.PlatformOpenCode {
			if result := opencode.ValidateCommandOpenCode(&d.Command); result.IsError() {
				errs = append(errs, unwrapJoinedErrors(result.Error)...)
			}
		}
	case *parser.CanonicalMemory:
		if result := core.ValidateMemory(&d.Memory); result.IsError() {
			errs = append(errs, unwrapJoinedErrors(result.Error)...)
		}
	case *parser.CanonicalSkill:
		if result := core.ValidateSkill(&d.Skill); result.IsError() {
			errs = append(errs, unwrapJoinedErrors(result.Error)...)
		}
		if req.Platform == core.PlatformOpenCode {
			if result := opencode.ValidateSkillOpenCode(&d.Skill); result.IsError() {
				errs = append(errs, unwrapJoinedErrors(result.Error)...)
			}
		}
	default:
		return nil, core.NewParseError(req.InputPath, "unknown document type", nil)
	}

	return &core.ValidateResult{Errors: errs}, nil
}

// unwrapJoinedErrors unwraps a joined error into individual errors.
// If the error is not a joined error, returns a slice with just that
// error. Renamed from cmd/validate.go's unwrapErrors to avoid
// potential collisions with future shell-package siblings and to
// reflect the actual semantics (we only know how to flatten joined
// errors; the type is implicit on the receiving side).
func unwrapJoinedErrors(err error) []error {
	if err == nil {
		return nil
	}

	type multipleUnwrapper interface {
		Unwrap() []error
	}

	if unwrapper, ok := err.(multipleUnwrapper); ok {
		return unwrapper.Unwrap()
	}

	return []error{err}
}
